package projects

import (
	"strconv"
	"time"

	"github.com/JorgeSaicoski/go-project-manager/internal/db"
	"github.com/JorgeSaicoski/go-project-manager/internal/services/projects"
	"github.com/JorgeSaicoski/microservice-commons/responses"
	"github.com/JorgeSaicoski/microservice-commons/types"
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

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req InternalCreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.BadRequest(c, err.Error())
		return
	}

	project, err := h.projectService.CreateProject(req.ToProject())
	if err != nil {
		responses.InternalError(c, err.Error())
		return
	}

	response := ProjectToResponse(project)
	responses.Created(c, "Project created successfully", response)
}

func (h *ProjectHandler) GetProject(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		responses.BadRequest(c, "Invalid project ID")
		return
	}

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
		responses.BadRequest(c, "User ID required")
		return
	}

	project, err := h.projectService.GetProject(uint(id), userID)
	if err != nil {
		if err.Error() == "user cannot access this project" {
			responses.Forbidden(c, err.Error())
			return
		}
		responses.NotFound(c, err.Error())
		return
	}

	response := ProjectToResponse(project)
	responses.Success(c, "Project retrieved successfully", response)
}

func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		responses.BadRequest(c, "Invalid project ID")
		return
	}

	var req struct {
		UpdateProjectRequest
		UserID string `json:"userId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.BadRequest(c, err.Error())
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
			responses.Forbidden(c, err.Error())
			return
		}
		responses.InternalError(c, err.Error())
		return
	}

	response := ProjectToResponse(project)
	responses.Success(c, "Project updated successfully", response)
}

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		responses.BadRequest(c, "Invalid project ID")
		return
	}

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
		responses.BadRequest(c, "User ID required")
		return
	}

	err = h.projectService.DeleteProject(uint(id), userID)
	if err != nil {
		if err.Error() == "only project owner can delete project" {
			responses.Forbidden(c, err.Error())
			return
		}
		responses.InternalError(c, err.Error())
		return
	}

	responses.Success(c, "Project deleted successfully", nil)
}

func (h *ProjectHandler) GetUserProjects(c *gin.Context) {
	userID := c.Query("userId")
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}

	if userID == "" {
		responses.BadRequest(c, "User ID required")
		return
	}

	projects, err := h.projectService.GetUserProjects(userID)
	if err != nil {
		responses.InternalError(c, err.Error())
		return
	}

	projectResponses := ProjectsToResponse(projects)
	response := types.ListResponse[ProjectResponse]{
		Data: projectResponses,
		Meta: types.ResponseMetadata{
			Count:     len(projectResponses),
			Timestamp: time.Now(),
		},
	}

	responses.Success(c, "Projects retrieved successfully", response)
}

func (h *ProjectHandler) AddProjectMember(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		responses.BadRequest(c, "Invalid project ID")
		return
	}

	var req struct {
		AddMemberRequest
		RequestingUserID string `json:"requestingUserId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.BadRequest(c, err.Error())
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
			responses.Forbidden(c, err.Error())
			return
		}
		if err.Error() == "user is already a member of this project" {
			responses.Conflict(c, err.Error())
			return
		}
		responses.InternalError(c, err.Error())
		return
	}

	response := MemberToResponse(member)
	responses.Created(c, "Member added successfully", response)
}

func (h *ProjectHandler) GetProjectMembers(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		responses.BadRequest(c, "Invalid project ID")
		return
	}

	userID := c.Query("userId")
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}

	if userID == "" {
		responses.BadRequest(c, "User ID required")
		return
	}

	members, err := h.projectService.GetProjectMembers(uint(id), userID)
	if err != nil {
		if err.Error() == "user cannot access this project" {
			responses.Forbidden(c, err.Error())
			return
		}
		responses.InternalError(c, err.Error())
		return
	}

	memberResponses := MembersToResponse(members)
	responses.Success(c, "Members retrieved successfully", gin.H{
		"members": memberResponses,
		"total":   len(memberResponses),
	})
}
