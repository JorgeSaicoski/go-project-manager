package projects

import (
	"github.com/JorgeSaicoski/go-project-manager/internal/services/projects"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all project-related routes
func RegisterRoutes(router *gin.RouterGroup, projectService *projects.ProjectService) {
	handler := NewProjectHandler(projectService)

	// Internal API routes for service-to-service communication
	internal := router.Group("/internal/projects")
	{
		// Project CRUD
		internal.POST("", handler.CreateProject)       // Create project
		internal.GET("/:id", handler.GetProject)       // Get project by ID
		internal.PUT("/:id", handler.UpdateProject)    // Update project
		internal.DELETE("/:id", handler.DeleteProject) // Delete project

		// User projects
		internal.GET("", handler.GetUserProjects) // Get user's projects (query: userId)

		// Project members
		internal.GET("/:id/members", handler.GetProjectMembers) // Get project members
		internal.POST("/:id/members", handler.AddProjectMember) // Add member to project
	}
}
