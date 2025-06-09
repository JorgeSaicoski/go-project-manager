package main

import (
	"os"
	"strings"
	"time"

	"github.com/JorgeSaicoski/go-project-manager/internal/api/companies"
	"github.com/JorgeSaicoski/go-project-manager/internal/api/projects"
	"github.com/JorgeSaicoski/go-project-manager/internal/db"
	companiesService "github.com/JorgeSaicoski/go-project-manager/internal/services/companies"
	projectsService "github.com/JorgeSaicoski/go-project-manager/internal/services/projects"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Connect to the database
	db.ConnectDatabase()

	// Initialize services
	projectService := projectsService.NewProjectService(db.DB)
	companyService := companiesService.NewCompanyService(db.DB)

	// Create gin router
	router := gin.Default()

	// Configure CORS
	setupCORS(router)

	// Setup routes
	setupRoutes(router, projectService, companyService)

	// Start the server
	port := getEnv("PORT", "8000")
	router.Run(":" + port)
}

func setupCORS(router *gin.Engine) {
	allowedOrigins := getEnv("ALLOWED_ORIGINS", "http://localhost:3000")
	origins := strings.Split(allowedOrigins, ",")

	router.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "X-User-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
}

func setupRoutes(router *gin.Engine, projectService *projectsService.ProjectService, companyService *companiesService.CompanyService) {
	// API group for all internal endpoints
	api := router.Group("/api")

	// Register domain routes
	projects.RegisterRoutes(api, projectService)
	companies.RegisterRoutes(api, companyService)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "project-manager",
		})
	})
}

// Helper function to get environment variables with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
