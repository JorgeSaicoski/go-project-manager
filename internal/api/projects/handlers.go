package projects

import (
	"net/http"
	"strconv"

	"github.com/JorgeSaicoski/go-project-manager/internal/db"
	"github.com/JorgeSaicoski/go-project-manager/internal/services/projects"
	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	projectService *projects.ProjectService
}

func NewProjectHandler(projectService *projects.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

// CreateProject handles internal project creation requests
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req InternalCreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.CreateProject(req.ToProject())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := ProjectToResponse(project)
	c.JSON(http.StatusCreated, response)
}

// GetProject retrieves a project by ID
func (h *ProjectHandler) GetProject(c *gin.Context) {
	// Get project ID from URL parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get requesting user ID from request body or header
	// Since this is internal service, the calling service should provide this
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		// Fallback to request body
		var req struct {
			UserID string `json:"userId"`
		}
		if err := c.ShouldBindJSON(&req); err == nil {
			userID = req.UserID
		}
	}

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	project, err := h.projectService.GetProject(uint(id), userID)
	if err != nil {
		if err.Error() == "user cannot access this project" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	response := ProjectToResponse(project)
	c.JSON(http.StatusOK, response)
}

// UpdateProject updates a project
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	// Get project ID from URL parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req struct {
		UpdateProjectRequest
		UserID string `json:"userId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := &req.UpdateProjectRequest
	projectUpdates := &db.BaseProject{
		Title:       updates.Title,
		Description: updates.Description,
		Status:      updates.Status,
		StartDate:   updates.StartDate,
		EndDate:     updates.EndDate,
	}

	project, err := h.projectService.UpdateProject(uint(id), projectUpdates, req.UserID)
	if err != nil {
		if err.Error() == "user cannot update this project" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := ProjectToResponse(project)
	c.JSON(http.StatusOK, response)
}

// DeleteProject deletes a project
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	// Get project ID from URL parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get user ID from request
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		var req struct {
			UserID string `json:"userId"`
		}
		if err := c.ShouldBindJSON(&req); err == nil {
			userID = req.UserID
		}
	}

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	err = h.projectService.DeleteProject(uint(id), userID)
	if err != nil {
		if err.Error() == "only project owner can delete project" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

// GetUserProjects retrieves all projects for a user
func (h *ProjectHandler) GetUserProjects(c *gin.Context) {
	// Get user ID from query parameter for internal calls
	userID := c.Query("userId")
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	projects, err := h.projectService.GetUserProjects(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses := ProjectsToResponse(projects)
	response := ProjectListResponse{
		Projects: responses,
		Total:    len(responses),
	}

	c.JSON(http.StatusOK, response)
}

// AddProjectMember adds a member to a project
func (h *ProjectHandler) AddProjectMember(c *gin.Context) {
	// Get project ID from URL parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req struct {
		AddMemberRequest
		RequestingUserID string `json:"requestingUserId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := h.projectService.AddProjectMember(
		uint(id),
		req.UserID,
		req.Role,
		req.Permissions,
		req.RequestingUserID,
	)
	if err != nil {
		if err.Error() == "user cannot add members to this project" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "user is already a member of this project" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := MemberToResponse(member)
	c.JSON(http.StatusCreated, response)
}

// GetProjectMembers retrieves all members of a project
func (h *ProjectHandler) GetProjectMembers(c *gin.Context) {
	// Get project ID from URL parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get requesting user ID
	userID := c.Query("userId")
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	members, err := h.projectService.GetProjectMembers(uint(id), userID)
	if err != nil {
		if err.Error() == "user cannot access this project" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses := MembersToResponse(members)
	c.JSON(http.StatusOK, gin.H{
		"members": responses,
		"total":   len(responses),
	})
}
