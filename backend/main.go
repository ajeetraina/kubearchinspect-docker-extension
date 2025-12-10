package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ImageResult represents the inspection result for a single image
type ImageResult struct {
	Image           string   `json:"image"`
	Namespace       string   `json:"namespace"`
	PodName         string   `json:"podName"`
	Kind            string   `json:"kind"`
	Status          string   `json:"status"`
	Message         string   `json:"message,omitempty"`
	SupportedArch   []string `json:"supportedArch,omitempty"`
	IsArmCompatible bool     `json:"isArmCompatible"`
	HasUpdateAvail  bool     `json:"hasUpdateAvailable"`
}

// InspectResponse represents the API response
type InspectResponse struct {
	Results  []ImageResult `json:"results"`
	Summary  Summary       `json:"summary"`
	ScanTime string        `json:"scanTime"`
}

// Summary represents the scan summary
type Summary struct {
	Total         int `json:"total"`
	ArmCompatible int `json:"armCompatible"`
	NotCompatible int `json:"notCompatible"`
	Errors        int `json:"errors"`
	CanUpgrade    int `json:"canUpgrade"`
}

// KubeContext represents a Kubernetes context
type KubeContext struct {
	Name      string `json:"name"`
	Cluster   string `json:"cluster"`
	IsCurrent bool   `json:"isCurrent"`
}

var logger = logrus.New()

func main() {
	var socketPath string
	flag.StringVar(&socketPath, "socket", "/run/guest-services/kubearchinspect.sock", "Unix domain socket to listen on")
	flag.Parse()

	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Remove existing socket file
	os.RemoveAll(socketPath)

	logger.Infof("Starting KubeArchInspect backend on %s", socketPath)

	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Routes
	e.GET("/health", healthHandler)
	e.GET("/contexts", getContextsHandler)
	e.GET("/inspect", inspectHandler)
	e.POST("/inspect", inspectHandler)

	// Create Unix socket listener
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		logger.Fatalf("Failed to listen on socket: %v", err)
	}

	e.Listener = listener
	logger.Fatal(e.Start(""))
}

func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
}

func getContextsHandler(c echo.Context) error {
	contexts, err := getKubeContexts()
	if err != nil {
		logger.Errorf("Failed to get contexts: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, contexts)
}

func getKubeContexts() ([]KubeContext, error) {
	kubeconfig := getKubeconfigPath()
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	var contexts []KubeContext
	for name, ctx := range config.Contexts {
		contexts = append(contexts, KubeContext{
			Name:      name,
			Cluster:   ctx.Cluster,
			IsCurrent: name == config.CurrentContext,
		})
	}
	return contexts, nil
}

func inspectHandler(c echo.Context) error {
	ctxName := c.QueryParam("context")
	namespace := c.QueryParam("namespace")

	logger.Infof("Inspecting cluster with context: %s, namespace: %s", ctxName, namespace)

	client, err := getKubernetesClient(ctxName)
	if err != nil {
		logger.Errorf("Failed to get Kubernetes client: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to connect to cluster: %v", err))
	}

	results, err := inspectCluster(c.Request().Context(), client, namespace)
	if err != nil {
		logger.Errorf("Failed to inspect cluster: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Calculate summary
	summary := Summary{Total: len(results)}
	for _, r := range results {
		switch r.Status {
		case "compatible":
			summary.ArmCompatible++
		case "not-compatible":
			summary.NotCompatible++
			if r.HasUpdateAvail {
				summary.CanUpgrade++
			}
		case "error":
			summary.Errors++
		}
	}

	response := InspectResponse{
		Results:  results,
		Summary:  summary,
		ScanTime: time.Now().Format(time.RFC3339),
	}

	return c.JSON(http.StatusOK, response)
}

func getKubeconfigPath() string {
	// Check for Docker Desktop's kubeconfig location
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE")
	}

	// Check standard locations
	paths := []string{
		filepath.Join(home, ".kube", "config"),
		"/root/.kube/config",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return filepath.Join(home, ".kube", "config")
}

func getKubernetesClient(contextName string) (*kubernetes.Clientset, error) {
	kubeconfig := getKubeconfigPath()

	var config *rest.Config
	var err error

	if contextName != "" {
		config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
			&clientcmd.ConfigOverrides{CurrentContext: contextName},
		).ClientConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func inspectCluster(ctx context.Context, client *kubernetes.Clientset, namespace string) ([]ImageResult, error) {
	// Collect unique images from pods
	imageMap := make(map[string]*ImageResult)
	var mu sync.Mutex

	// Get pods
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			key := container.Image
			if _, exists := imageMap[key]; !exists {
				imageMap[key] = &ImageResult{
					Image:     container.Image,
					Namespace: pod.Namespace,
					PodName:   pod.Name,
					Kind:      "Pod",
				}
			}
		}
	}

	// Get deployments
	deployments, err := client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.Warnf("Failed to list deployments: %v", err)
	} else {
		for _, deploy := range deployments.Items {
			for _, container := range deploy.Spec.Template.Spec.Containers {
				key := container.Image
				if _, exists := imageMap[key]; !exists {
					imageMap[key] = &ImageResult{
						Image:     container.Image,
						Namespace: deploy.Namespace,
						PodName:   deploy.Name,
						Kind:      "Deployment",
					}
				}
			}
		}
	}

	// Get daemonsets
	daemonsets, err := client.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.Warnf("Failed to list daemonsets: %v", err)
	} else {
		for _, ds := range daemonsets.Items {
			for _, container := range ds.Spec.Template.Spec.Containers {
				key := container.Image
				if _, exists := imageMap[key]; !exists {
					imageMap[key] = &ImageResult{
						Image:     container.Image,
						Namespace: ds.Namespace,
						PodName:   ds.Name,
						Kind:      "DaemonSet",
					}
				}
			}
		}
	}

	// Get statefulsets
	statefulsets, err := client.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.Warnf("Failed to list statefulsets: %v", err)
	} else {
		for _, sts := range statefulsets.Items {
			for _, container := range sts.Spec.Template.Spec.Containers {
				key := container.Image
				if _, exists := imageMap[key]; !exists {
					imageMap[key] = &ImageResult{
						Image:     container.Image,
						Namespace: sts.Namespace,
						PodName:   sts.Name,
						Kind:      "StatefulSet",
					}
				}
			}
		}
	}

	// Check each image for ARM64 support concurrently
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // Limit concurrent registry calls

	for _, result := range imageMap {
		wg.Add(1)
		go func(r *ImageResult) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			checkImageArm64Support(r)
			mu.Lock()
			mu.Unlock()
		}(result)
	}

	wg.Wait()

	// Convert map to slice
	var results []ImageResult
	for _, r := range imageMap {
		results = append(results, *r)
	}

	return results, nil
}

func checkImageArm64Support(result *ImageResult) {
	ref, err := name.ParseReference(result.Image)
	if err != nil {
		result.Status = "error"
		result.Message = fmt.Sprintf("Failed to parse image reference: %v", err)
		return
	}

	// Try to get the image index/manifest
	idx, err := remote.Index(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		// Try as a single-platform image
		img, imgErr := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
		if imgErr != nil {
			result.Status = "error"
			result.Message = fmt.Sprintf("Failed to fetch image: %v", err)
			return
		}

		// Check config for architecture
		cfg, cfgErr := img.ConfigFile()
		if cfgErr != nil {
			result.Status = "error"
			result.Message = fmt.Sprintf("Failed to get image config: %v", cfgErr)
			return
		}

		result.SupportedArch = []string{cfg.Architecture}
		if cfg.Architecture == "arm64" || cfg.Architecture == "aarch64" {
			result.Status = "compatible"
			result.IsArmCompatible = true
		} else {
			result.Status = "not-compatible"
			result.IsArmCompatible = false
			result.Message = fmt.Sprintf("Only supports %s", cfg.Architecture)
		}
		return
	}

	// It's a multi-arch image, check the manifest
	idxManifest, err := idx.IndexManifest()
	if err != nil {
		result.Status = "error"
		result.Message = fmt.Sprintf("Failed to get index manifest: %v", err)
		return
	}

	var architectures []string
	hasArm64 := false

	for _, desc := range idxManifest.Manifests {
		if desc.Platform != nil {
			arch := desc.Platform.Architecture
			if desc.Platform.Variant != "" {
				arch = fmt.Sprintf("%s/%s", arch, desc.Platform.Variant)
			}
			architectures = append(architectures, arch)
			if desc.Platform.Architecture == "arm64" || desc.Platform.Architecture == "aarch64" {
				hasArm64 = true
			}
		}
	}

	result.SupportedArch = architectures
	if hasArm64 {
		result.Status = "compatible"
		result.IsArmCompatible = true
	} else {
		result.Status = "not-compatible"
		result.IsArmCompatible = false
		result.Message = fmt.Sprintf("Image only supports: %s", strings.Join(architectures, ", "))
	}
}
