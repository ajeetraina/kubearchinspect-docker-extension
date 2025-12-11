package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)
}

type InspectRequest struct {
	Context   string `query:"context"`
	Namespace string `query:"namespace"`
}

type InspectResult struct {
	Image           string   `json:"image"`
	ResourceType    string   `json:"resourceType"`
	ResourceName    string   `json:"resourceName"`
	Namespace       string   `json:"namespace"`
	IsArmCompatible bool     `json:"isArmCompatible"`
	Architectures   []string `json:"architectures"`
	Error           string   `json:"error,omitempty"`
}

type InspectResponse struct {
	Context  string          `json:"context"`
	Results  []InspectResult `json:"results"`
	Summary  Summary         `json:"summary"`
	ScanTime string          `json:"scanTime"`
}

type Summary struct {
	Total       int `json:"total"`
	Compatible  int `json:"compatible"`
	Incompatible int `json:"incompatible"`
	Errors      int `json:"errors"`
}

type Resource struct {
	Type      string
	Name      string
	Namespace string
	Image     string
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Health check endpoint
	e.GET("/hello", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "KubeArchInspect backend is running",
		})
	})

	// Main inspection endpoint
	e.GET("/inspect", handleInspect)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Infof("Starting server on port %s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatal(err)
	}
}

func handleInspect(c echo.Context) error {
	var req InspectRequest
	if err := c.Bind(&req); err != nil {
		log.WithError(err).Error("Failed to bind request")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request parameters",
		})
	}

	if req.Context == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "context parameter is required",
		})
	}

	if req.Namespace == "" {
		req.Namespace = "all"
	}

	log.WithFields(logrus.Fields{
		"context":   req.Context,
		"namespace": req.Namespace,
	}).Info("Starting inspection")

	startTime := time.Now()

	// Get resources from cluster
	resources, err := getResources(req.Context, req.Namespace)
	if err != nil {
		log.WithError(err).Error("Failed to get resources")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to get resources: %v", err),
		})
	}

	log.Infof("Found %d resources to inspect", len(resources))

	// Inspect images concurrently
	results := inspectImages(resources)

	// Calculate summary
	summary := Summary{Total: len(results)}
	for _, result := range results {
		if result.Error != "" {
			summary.Errors++
		} else if result.IsArmCompatible {
			summary.Compatible++
		} else {
			summary.Incompatible++
		}
	}

	scanTime := time.Since(startTime).String()

	response := InspectResponse{
		Context:  req.Context,
		Results:  results,
		Summary:  summary,
		ScanTime: scanTime,
	}

	log.WithFields(logrus.Fields{
		"total":       summary.Total,
		"compatible":  summary.Compatible,
		"incompatible": summary.Incompatible,
		"errors":      summary.Errors,
		"scan_time":   scanTime,
	}).Info("Inspection completed")

	return c.JSON(http.StatusOK, response)
}

func getResources(contextName, namespace string) ([]Resource, error) {
	args := []string{
		"--context", contextName,
		"get", "pods,deployments,statefulsets,daemonsets,jobs,cronjobs",
		"-o", "json",
	}

	if namespace != "all" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "--all-namespaces")
	}

	cmd := exec.Command("kubectl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("kubectl command failed: %v, output: %s", err, string(output))
	}

	var result struct {
		Items []map[string]interface{} `json:"items"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse kubectl output: %v", err)
	}

	var resources []Resource
	seen := make(map[string]bool)

	for _, item := range result.Items {
		kind := item["kind"].(string)
		metadata := item["metadata"].(map[string]interface{})
		resourceName := metadata["name"].(string)
		resourceNamespace := metadata["namespace"].(string)

		var containers []map[string]interface{}

		// Extract containers based on resource type
		spec := item["spec"].(map[string]interface{})
		switch kind {
		case "Pod":
			if cs, ok := spec["containers"].([]interface{}); ok {
				for _, c := range cs {
					containers = append(containers, c.(map[string]interface{}))
				}
			}
		case "Deployment", "StatefulSet", "DaemonSet":
			if template, ok := spec["template"].(map[string]interface{}); ok {
				if templateSpec, ok := template["spec"].(map[string]interface{}); ok {
					if cs, ok := templateSpec["containers"].([]interface{}); ok {
						for _, c := range cs {
							containers = append(containers, c.(map[string]interface{}))
						}
					}
				}
			}
		case "Job", "CronJob":
			var jobSpec map[string]interface{}
			if kind == "CronJob" {
				if jobTemplate, ok := spec["jobTemplate"].(map[string]interface{}); ok {
					jobSpec = jobTemplate["spec"].(map[string]interface{})
				}
			} else {
				jobSpec = spec
			}
			if template, ok := jobSpec["template"].(map[string]interface{}); ok {
				if templateSpec, ok := template["spec"].(map[string]interface{}); ok {
					if cs, ok := templateSpec["containers"].([]interface{}); ok {
						for _, c := range cs {
							containers = append(containers, c.(map[string]interface{}))
						}
					}
				}
			}
		}

		// Extract images
		for _, container := range containers {
			if image, ok := container["image"].(string); ok && image != "" {
				key := fmt.Sprintf("%s/%s/%s/%s", kind, resourceNamespace, resourceName, image)
				if !seen[key] {
					seen[key] = true
					resources = append(resources, Resource{
						Type:      kind,
						Name:      resourceName,
						Namespace: resourceNamespace,
						Image:     image,
					})
				}
			}
		}
	}

	return resources, nil
}

func inspectImages(resources []Resource) []InspectResult {
	results := make([]InspectResult, len(resources))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // Limit concurrent requests

	for i, resource := range resources {
		wg.Add(1)
		go func(idx int, res Resource) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result := InspectResult{
				Image:        res.Image,
				ResourceType: res.Type,
				ResourceName: res.Name,
				Namespace:    res.Namespace,
			}

			archs, err := getImageArchitectures(res.Image)
			if err != nil {
				result.Error = err.Error()
				log.WithError(err).WithField("image", res.Image).Warn("Failed to inspect image")
			} else {
				result.Architectures = archs
				result.IsArmCompatible = contains(archs, "arm64")
				log.WithFields(logrus.Fields{
					"image":         res.Image,
					"architectures": strings.Join(archs, ", "),
					"arm_compatible": result.IsArmCompatible,
				}).Debug("Image inspected")
			}

			results[idx] = result
		}(i, resource)
	}

	wg.Wait()
	return results
}

func getImageArchitectures(imageName string) ([]string, error) {
	ref, err := name.ParseReference(imageName)
	if err != nil {
		return nil, fmt.Errorf("invalid image reference: %v", err)
	}

	desc, err := remote.Get(ref, remote.WithContext(context.Background()))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %v", err)
	}

	// Check if it's a manifest list (multi-arch)
	if desc.MediaType.IsIndex() {
		idx, err := desc.ImageIndex()
		if err != nil {
			return nil, fmt.Errorf("failed to get image index: %v", err)
		}

		manifest, err := idx.IndexManifest()
		if err != nil {
			return nil, fmt.Errorf("failed to get index manifest: %v", err)
		}

		archs := make(map[string]bool)
		for _, m := range manifest.Manifests {
			if m.Platform != nil {
				archs[m.Platform.Architecture] = true
			}
		}

		var result []string
		for arch := range archs {
			result = append(result, arch)
		}
		return result, nil
	}

	// Single architecture image
	img, err := desc.Image()
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %v", err)
	}

	config, err := img.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %v", err)
	}

	return []string{config.Architecture}, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}