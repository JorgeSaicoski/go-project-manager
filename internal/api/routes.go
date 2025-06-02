package api

import (
	"strings"
	"time"

	"github.com/JorgeSaicoski/pgconnect"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// ProjectRouter handles routing for project-related endpoints
type ProjectRouter struct {
	handler *ProjectHandler
	router  *gin.Engine
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
		AllowedHeaders:   []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization"},
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

	// Create the project handler
	projectHandler := NewProjectHandler(database)

	return &ProjectRouter{
		handler: projectHandler,
		router:  router,
	}
}

// RegisterRoutes sets up all project-related routes
func (tr *ProjectRouter) RegisterRoutes() {
	// Projects endpoints
	projectsGroup := tr.router.Group("/projects")
	projectsGroup.Use(AuthMiddleware())
	{

	}

	// Home route
	tr.router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.tmpl", gin.H{
			"title": "Projects",
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
