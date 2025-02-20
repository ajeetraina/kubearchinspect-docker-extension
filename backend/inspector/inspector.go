package inspector

import (
	"context"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Resource struct {
	Name           string `json:"name"`
	Namespace      string `json:"namespace"`
	Kind           string `json:"kind"`
	IsArmCompatible bool   `json:"isArmCompatible"`
	Image          string `json:"image,omitempty"`
}

type Inspector struct {
	client *kubernetes.Clientset
}

func NewInspector(client *kubernetes.Clientset) *Inspector {
	return &Inspector{client: client}
}

func (i *Inspector) InspectResources(ctx context.Context) ([]Resource, error) {
	var resources []Resource

	// Get all namespaces
	namespaces, err := i.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Inspect each namespace
	for _, ns := range namespaces.Items {
		// Get deployments
		deployments, err := i.client.AppsV1().Deployments(ns.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			continue
		}

		for _, deployment := range deployments.Items {
			for _, container := range deployment.Spec.Template.Spec.Containers {
				resources = append(resources, Resource{
					Name:           deployment.Name,
					Namespace:      deployment.Namespace,
					Kind:           "Deployment",
					IsArmCompatible: isArmImage(container.Image),
					Image:          container.Image,
				})
			}
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