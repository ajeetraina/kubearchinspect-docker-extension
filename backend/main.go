package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
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
	ResourceName    string   `json:"resourceName"`
	ResourceType    string   `json:"resourceType"`
	Error           string   `json:"error,omitempty"`
	SupportedArch   []string `json:"supportedArch,omitempty"`
	IsArmCompatible bool     `json:"isArmCompatible"`
}

// InspectResponse represents the API response
type InspectResponse struct {
	Results   []ImageResult `json:"results"`
	Summary   Summary       `json:"summary"`
	ScanTime  string        `json:"scanTime"`
	Context   string        `json:"context"`
	Namespace string        `json:"namespace"`
}

// Summary represents the scan summary
type Summary struct {
	Total         int `json:"total"`
	ArmCompatible int `json:"armCompatible"`
	NotCompatible int `json:"notCompatible"`
	Errors        int `json:"errors"`
}

// KubeContext represents a Kubernetes context
type KubeContext struct {
	Name      string `json:"name"`
	Cluster   string `json:"cluster"`
	IsCurrent bool   `json:"isCurrent"`
}

// ContextsResponse represents the contexts API response
type ContextsResponse struct {
	Contexts []KubeContext `json:"contexts"`
	Current  string        `json:"current"`
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
	logger.Info("=== getContextsHandler called ===")
	contexts, current, err := getKubeContexts()
	if err != nil {
		logger.Errorf("Failed to get contexts: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	logger.Infof("Returning %d contexts, current: %s", len(contexts), current)
	return c.JSON(http.StatusOK, ContextsResponse{
		Contexts: contexts,
		Current:  current,
	})
}

func getKubeContexts() ([]KubeContext, string, error) {
	kubeconfig := getKubeconfigPath()
	
	// Add debugging
	logger.Infof("Looking for kubeconfig at: %s", kubeconfig)
	
	// Check if file exists
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		logger.Errorf("Kubeconfig file does not exist at: %s", kubeconfig)
		
		// Try to list directory contents
		if dir := filepath.Dir(kubeconfig); dir != "" {
			logger.Infof("Attempting to list contents of directory: %s", dir)
			if entries, err := os.ReadDir(dir); err == nil {
				logger.Infof("Contents of %s:", dir)
				for _, entry := range entries {
					logger.Infof("  - %s (isDir: %v)", entry.Name(), entry.IsDir())
				}
			} else {
				logger.Errorf("Cannot read directory %s: %v", dir, err)
			}
		}
		
		return nil, "", fmt.Errorf("kubeconfig file not found at %s", kubeconfig)
	}
	
	logger.Infof("✓ Found kubeconfig at: %s", kubeconfig)
	
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		logger.Errorf("Failed to load kubeconfig: %v", err)
		return nil, "", fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	logger.Infof("✓ Successfully loaded kubeconfig with %d contexts", len(config.Contexts))
	
	if len(config.Contexts) == 0 {
		logger.Warn("Kubeconfig loaded but contains no contexts")
		return []KubeContext{}, "", nil
	}
	
	var contexts []KubeContext
	for name, ctx := range config.Contexts {
		logger.Infof("Found context: %s (cluster: %s, current: %v)", name, ctx.Cluster, name == config.CurrentContext)
		contexts = append(contexts, KubeContext{
			Name:      name,
			Cluster:   ctx.Cluster,
			IsCurrent: name == config.CurrentContext,
		})
	}
	
	logger.Infof("✓ Returning %d contexts, current context: %s", len(contexts), config.CurrentContext)
	return contexts, config.CurrentContext, nil
}

func inspectHandler(c echo.Context) error {
	ctxName := c.QueryParam("context")
	namespace := c.QueryParam("namespace")

	if namespace == "" || namespace == "all" {
		namespace = "" // Empty string means all namespaces in K8s API
	}

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
		if r.Error != "" {
			summary.Errors++
		} else if r.IsArmCompatible {
			summary.ArmCompatible++
		} else {
			summary.NotCompatible++
		}
	}

	response := InspectResponse{
		Results:   results,
		Summary:   summary,
		ScanTime:  time.Now().Format(time.RFC3339),
		Context:   ctxName,
		Namespace: namespace,
	}

	return c.JSON(http.StatusOK, response)
}

func getKubeconfigPath() string {
	// Check KUBECONFIG environment variable first
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		logger.Infof("Using KUBECONFIG from environment: %s", kubeconfig)
		return kubeconfig
	}
	
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

	logger.Infof("Checking kubeconfig in standard locations (HOME=%s)", home)
	for _, p := range paths {
		logger.Infof("  Checking: %s", p)
		if _, err := os.Stat(p); err == nil {
			logger.Infof("  ✓ Found at: %s", p)
			return p
		}
	}

	logger.Warnf("No kubeconfig found, defaulting to: %s", filepath.Join(home, ".kube", "config"))
	return filepath.Join(home, ".kube", "config")
}

func getKubernetesClient(contextName string) (*kubernetes.Clientset, error) {
	kubeconfig := getKubeconfigPath()
	logger.Infof("Creating Kubernetes client with context: %s, kubeconfig: %s", contextName, kubeconfig)

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
		logger.Errorf("Failed to build Kubernetes config: %v", err)
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Errorf("Failed to create Kubernetes client: %v", err)
		return nil, err
	}
	
	logger.Info("✓ Successfully created Kubernetes client")
	return client, nil
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

	logger.Infof("Found %d pods", len(pods.Items))
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			key := container.Image
			if _, exists := imageMap[key]; !exists {
				imageMap[key] = &ImageResult{
					Image:        container.Image,
					Namespace:    pod.Namespace,
					ResourceName: pod.Name,
					ResourceType: "Pod",
				}
			}
		}
	}

	// Get deployments
	deployments, err := client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.Warnf("Failed to list deployments: %v", err)
	} else {
		logger.Infof("Found %d deployments", len(deployments.Items))
		for _, deploy := range deployments.Items {
			for _, container := range deploy.Spec.Template.Spec.Containers {
				key := container.Image
				if _, exists := imageMap[key]; !exists {
					imageMap[key] = &ImageResult{
						Image:        container.Image,
						Namespace:    deploy.Namespace,
						ResourceName: deploy.Name,
						ResourceType: "Deployment",
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
		logger.Infof("Found %d daemonsets", len(daemonsets.Items))
		for _, ds := range daemonsets.Items {
			for _, container := range ds.Spec.Template.Spec.Containers {
				key := container.Image
				if _, exists := imageMap[key]; !exists {
					imageMap[key] = &ImageResult{
						Image:        container.Image,
						Namespace:    ds.Namespace,
						ResourceName: ds.Name,
						ResourceType: "DaemonSet",
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
		logger.Infof("Found %d statefulsets", len(statefulsets.Items))
		for _, sts := range statefulsets.Items {
			for _, container := range sts.Spec.Template.Spec.Containers {
				key := container.Image
				if _, exists := imageMap[key]; !exists {
					imageMap[key] = &ImageResult{
						Image:        container.Image,
						Namespace:    sts.Namespace,
						ResourceName: sts.Name,
						ResourceType: "StatefulSet",
					}
				}
			}
		}
	}

	logger.Infof("Found %d unique images to inspect", len(imageMap))

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

	logger.Infof("Completed inspection of %d images", len(results))
	return results, nil
}

func checkImageArm64Support(result *ImageResult) {
	ref, err := name.ParseReference(result.Image)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to parse image reference: %v", err)
		return
	}

	// Try to get the image index/manifest
	idx, err := remote.Index(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		// Try as a single-platform image
		img, imgErr := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
		if imgErr != nil {
			result.Error = fmt.Sprintf("Failed to fetch image: %v", err)
			return
		}

		// Check config for architecture
		cfg, cfgErr := img.ConfigFile()
		if cfgErr != nil {
			result.Error = fmt.Sprintf("Failed to get image config: %v", cfgErr)
			return
		}

		result.SupportedArch = []string{cfg.Architecture}
		if cfg.Architecture == "arm64" || cfg.Architecture == "aarch64" {
			result.IsArmCompatible = true
		} else {
			result.IsArmCompatible = false
		}
		return
	}

	// It's a multi-arch image, check the manifest
	idxManifest, err := idx.IndexManifest()
	if err != nil {
		result.Error = fmt.Sprintf("Failed to get index manifest: %v", err)
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
	result.IsArmCompatible = hasArm64
}
