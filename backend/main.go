package main

import (
    "context"
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "github.com/sirupsen/logrus"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
    "strings"
    "path/filepath"
    "os"
)

type Resource struct {
    Name           string `json:"name"`
    Namespace      string `json:"namespace"`
    Kind           string `json:"kind"`
    IsArmCompatible bool   `json:"isArmCompatible"`
    Image          string `json:"image,omitempty"`
}

func main() {
    // Echo instance
    e := echo.New()

    // Middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(middleware.CORS())

    // Routes
    e.GET("/inspect", handleInspect)

    // Start server
    port := ":8080"
    logrus.Infof("Starting server on port %s", port)
    e.Logger.Fatal(e.Start(port))
}

func handleInspect(c echo.Context) error {
    client, err := getKubernetesClient()
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
    }

    resources, err := inspectResources(c.Request().Context(), client)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
    }

    return c.JSON(http.StatusOK, resources)
}

func getKubernetesClient() (*kubernetes.Clientset, error) {
    // Try in-cluster config first
    config, err := rest.InClusterConfig()
    if err != nil {
        // Fallback to kubeconfig
        kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
        config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
        if err != nil {
            return nil, err
        }
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, err
    }

    return clientset, nil
}

func inspectResources(ctx context.Context, client *kubernetes.Clientset) ([]Resource, error) {
    var resources []Resource

    // Get deployments from all namespaces
    deployments, err := client.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, err
    }

    for _, deployment := range deployments.Items {
        containers := deployment.Spec.Template.Spec.Containers
        for _, container := range containers {
            resources = append(resources, Resource{
                Name:           deployment.Name,
                Namespace:      deployment.Namespace,
                Kind:           "Deployment",
                IsArmCompatible: isArmImage(container.Image),
                Image:          container.Image,
            })
        }
    }

    return resources, nil
}

func isArmImage(image string) bool {
    armIndicators := []string{
        "arm64",
        "arm/v7",
        "arm/v8",
        "aarch64",
    }

    imageLower := strings.ToLower(image)
    for _, indicator := range armIndicators {
        if strings.Contains(imageLower, indicator) {
            return true
        }
    }

    return false
}