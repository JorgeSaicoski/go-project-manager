package api

import (
	"strings"
	"time"

	"github.com/JorgeSaicoski/go-project-manager/internal/api/projects"
	projectsService "github.com/JorgeSaicoski/go-project-manager/internal/services/projects"
	"github.com/JorgeSaicoski/pgconnect"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// ProjectRouter handles routing for project-related endpoints
type ProjectRouter struct {
	router *gin.Engine
}

// RouterConfig holds configuration for the router
type RouterConfig struct {
	AllowedOrigins   string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
	TemplatesGlob    string
}

// DefaultRouterConfig returns a default router configuration
func DefaultRouterConfig() RouterConfig {
	return RouterConfig{
		AllowedOrigins:   "http://localhost:3000",
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "X-User-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60, // 12 hours
		TemplatesGlob:    "public/*",
	}
}

// NewProjectRouter creates a new ProjectRouter instance with full configuration
func NewProjectRouter(database *pgconnect.DB, config RouterConfig) *ProjectRouter {
	router := gin.Default()

	// Configure CORS middleware
	origins := strings.Split(config.AllowedOrigins, ",")
	router.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     config.AllowedMethods,
		AllowHeaders:     config.AllowedHeaders,
		ExposeHeaders:    config.ExposeHeaders,
		AllowCredentials: config.AllowCredentials,
		MaxAge:           time.Duration(config.MaxAge) * time.Second,
	}))

	// Load HTML templates if specified
	if config.TemplatesGlob != "" {
		router.LoadHTMLGlob(config.TemplatesGlob)
	}

	return &ProjectRouter{
		router: router,
	}
}

// RegisterRoutes sets up all routes for the internal service
func (tr *ProjectRouter) RegisterRoutes(projectService *projectsService.ProjectService) {
	// API group for all internal endpoints
	api := tr.router.Group("/api/v1")

	// Register projects routes
	projects.RegisterRoutes(api, projectService)

	// Health check endpoint (no auth needed for internal services)
	tr.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "project-manager",
		})
	})

	// Home route (optional, for testing)
	tr.router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "Project Manager",
			"version": "1.0.0",
			"endpoints": []string{
				"POST /api/v1/internal/projects",
				"GET /api/v1/internal/projects/:id",
				"PUT /api/v1/internal/projects/:id",
				"DELETE /api/v1/internal/projects/:id",
				"GET /api/v1/internal/projects",
				"GET /api/v1/internal/projects/:id/members",
				"POST /api/v1/internal/projects/:id/members",
			},
		})
	})
}

// GetRouter returns the configured gin router
func (tr *ProjectRouter) GetRouter() *gin.Engine {
	return tr.router
}

// Run starts the HTTP server
func (tr *ProjectRouter) Run(addr string) error {
	return tr.router.Run(addr)
}
