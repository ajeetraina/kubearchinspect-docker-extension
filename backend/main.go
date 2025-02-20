package main

import (
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "github.com/sirupsen/logrus"
)

func main() {
    e := echo.New()
    e.Use(middleware.CORS())
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())

    // Initialize Kubernetes client
    client, err := kubernetes.GetKubernetesClient()
    if err != nil {
        logrus.Fatalf("Failed to create Kubernetes client: %v", err)
    }

    // Create inspector
    inspector := inspector.NewKubernetesInspector(client)

    // Routes
    e.GET("/inspect", inspector.HandleInspect)
    e.GET("/statistics", inspector.HandleStatistics)
    e.POST("/export", inspector.HandleExport)

    // Start server
    e.Logger.Fatal(e.Start(":8080"))
}