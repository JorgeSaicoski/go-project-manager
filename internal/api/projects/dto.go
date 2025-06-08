package projects

import (
	"time"

	"github.com/JorgeSaicoski/go-project-manager/internal/db"
)

// Request DTOs

type CreateProjectRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description *string    `json:"description"`
	Status      string     `json:"status"`
	CompanyID   *string    `json:"companyId"`
	StartDate   *time.Time `json:"startDate"`
	EndDate     *time.Time `json:"endDate"`
	OwnerID     string     `json:"ownerId" binding:"required"` // Set by calling service
}

type UpdateProjectRequest struct {
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Status      string     `json:"status"`
	StartDate   *time.Time `json:"startDate"`
	EndDate     *time.Time `json:"endDate"`
}

type AddMemberRequest struct {
	UserID      string   `json:"userId" binding:"required"`
	Role        string   `json:"role" binding:"required"`
	Permissions []string `json:"permissions"`
}

// Response DTOs

type ProjectResponse struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Status      string     `json:"status"`
	OwnerID     string     `json:"ownerId"`
	CompanyID   *string    `json:"companyId"`
	StartDate   *time.Time `json:"startDate"`
	EndDate     *time.Time `json:"endDate"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type ProjectMemberResponse struct {
	ProjectID   string    `json:"projectId"`
	ProjectType string    `json:"projectType"`
	UserID      string    `json:"userId"`
	Role        string    `json:"role"`
	Permissions []string  `json:"permissions"`
	JoinedAt    time.Time `json:"joinedAt"`
}

type ProjectListResponse struct {
	Projects []ProjectResponse `json:"projects"`
	Total    int               `json:"total"`
}

// Internal service request (for service-to-service calls)
type InternalCreateProjectRequest struct {
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Status      string     `json:"status"`
	OwnerID     string     `json:"ownerId"`
	CompanyID   *string    `json:"companyId"`
	StartDate   *time.Time `json:"startDate"`
	EndDate     *time.Time `json:"endDate"`
}

// Conversion methods

func (r *CreateProjectRequest) ToProject() *db.BaseProject {
	return &db.BaseProject{
		Title:       r.Title,
		Description: r.Description,
		Status:      r.Status,
		OwnerID:     r.OwnerID,
		CompanyID:   r.CompanyID,
		StartDate:   r.StartDate,
		EndDate:     r.EndDate,
	}
}

func (r *InternalCreateProjectRequest) ToProject() *db.BaseProject {
	return &db.BaseProject{
		Title:       r.Title,
		Description: r.Description,
		Status:      r.Status,
		OwnerID:     r.OwnerID,
		CompanyID:   r.CompanyID,
		StartDate:   r.StartDate,
		EndDate:     r.EndDate,
	}
}

func ProjectToResponse(project *db.BaseProject) ProjectResponse {
	return ProjectResponse{
		ID:          project.ID,
		Title:       project.Title,
		Description: project.Description,
		Status:      project.Status,
		OwnerID:     project.OwnerID,
		CompanyID:   project.CompanyID,
		StartDate:   project.StartDate,
		EndDate:     project.EndDate,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}
}

func ProjectsToResponse(projects []db.BaseProject) []ProjectResponse {
	responses := make([]ProjectResponse, len(projects))
	for i, project := range projects {
		responses[i] = ProjectToResponse(&project)
	}
	return responses
}

func MemberToResponse(member *db.ProjectMember) ProjectMemberResponse {
	return ProjectMemberResponse{
		ProjectID:   member.ProjectID,
		ProjectType: member.ProjectType,
		UserID:      member.UserID,
		Role:        member.Role,
		Permissions: member.Permissions,
		JoinedAt:    member.JoinedAt,
	}
}

func MembersToResponse(members []db.ProjectMember) []ProjectMemberResponse {
	responses := make([]ProjectMemberResponse, len(members))
	for i, member := range members {
		responses[i] = MemberToResponse(&member)
	}
	return responses
}
