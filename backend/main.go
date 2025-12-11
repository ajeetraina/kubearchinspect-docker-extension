package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
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

// Image to inspect
type ImageRequest struct {
	Image        string `json:"image"`
	ResourceType string `json:"resourceType"`
	ResourceName string `json:"resourceName"`
	Namespace    string `json:"namespace"`
}

// Request body for batch inspection
type InspectRequest struct {
	Images []ImageRequest `json:"images"`
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
	Results  []InspectResult `json:"results"`
	Summary  Summary         `json:"summary"`
	ScanTime string          `json:"scanTime"`
}

type Summary struct {
	Total        int `json:"total"`
	Compatible   int `json:"compatible"`
	Incompatible int `json:"incompatible"`
	Errors       int `json:"errors"`
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

	// Main inspection endpoint - accepts list of images
	e.POST("/inspect", handleInspect)

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
			"error": "Invalid request body",
		})
	}

	if len(req.Images) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "No images provided",
		})
	}

	log.Infof("Received request to inspect %d images", len(req.Images))

	startTime := time.Now()

	// Inspect images concurrently
	results := inspectImages(req.Images)

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
		Results:  results,
		Summary:  summary,
		ScanTime: scanTime,
	}

	log.WithFields(logrus.Fields{
		"total":        summary.Total,
		"compatible":   summary.Compatible,
		"incompatible": summary.Incompatible,
		"errors":       summary.Errors,
		"scan_time":    scanTime,
	}).Info("Inspection completed")

	return c.JSON(http.StatusOK, response)
}

func inspectImages(images []ImageRequest) []InspectResult {
	results := make([]InspectResult, len(images))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // Limit concurrent requests

	for i, img := range images {
		wg.Add(1)
		go func(idx int, imgReq ImageRequest) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result := InspectResult{
				Image:        imgReq.Image,
				ResourceType: imgReq.ResourceType,
				ResourceName: imgReq.ResourceName,
				Namespace:    imgReq.Namespace,
			}

			archs, err := getImageArchitectures(imgReq.Image)
			if err != nil {
				result.Error = err.Error()
				log.WithError(err).WithField("image", imgReq.Image).Warn("Failed to inspect image")
			} else {
				result.Architectures = archs
				result.IsArmCompatible = contains(archs, "arm64")
				log.WithFields(logrus.Fields{
					"image":          imgReq.Image,
					"architectures":  strings.Join(archs, ", "),
					"arm_compatible": result.IsArmCompatible,
				}).Debug("Image inspected")
			}

			results[idx] = result
		}(i, img)
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
