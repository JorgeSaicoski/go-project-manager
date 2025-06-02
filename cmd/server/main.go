package main

import (
	"os"

	"github.com/JorgeSaicoski/go-project-manager/internal/api"
	"github.com/JorgeSaicoski/go-project-manager/internal/db"
)

func main() {
	// Connect to the database
	db.ConnectDatabase()

	// Get router config, possibly from environment variables
	config := api.DefaultRouterConfig()

	// Override with environment variables if needed
	if origins := getEnv("ALLOWED_ORIGINS", ""); origins != "" {
		config.AllowedOrigins = origins
	}

	// Create router with full configuration
	projectRouter := api.NewProjectRouter(db.DB, config)

	// Register all routes
	projectRouter.RegisterRoutes()

	// Start the server
	port := getEnv("PORT", "8000")
	projectRouter.Run(":" + port)
}

// Helper function to get environment variables with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback

}
