package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"net/http"
)

func main() {
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
	// This is a placeholder response
	result := map[string]interface{}{
		"status": "success",
		"message": "Inspection service is running",
	}

	return c.JSON(http.StatusOK, result)
}