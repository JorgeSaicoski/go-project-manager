package projects

import (
	"errors"
	"strconv"
	"time"

	"github.com/JorgeSaicoski/go-project-manager/internal/db"
	"github.com/JorgeSaicoski/pgconnect"
)

type ProjectService struct {
	projectRepo       *pgconnect.Repository[db.BaseProject]
	memberRepo        *pgconnect.Repository[db.ProjectMember]
	companyMemberRepo *pgconnect.Repository[db.CompanyMember]
}

func NewProjectService(database *pgconnect.DB) *ProjectService {
	return &ProjectService{
		projectRepo:       pgconnect.NewRepository[db.BaseProject](database),
		memberRepo:        pgconnect.NewRepository[db.ProjectMember](database),
		companyMemberRepo: pgconnect.NewRepository[db.CompanyMember](database),
	}
}

func (s *ProjectService) CreateProject(project *db.BaseProject) (*db.BaseProject, error) {
	// Business logic: validate company ownership if company is specified
	if project.CompanyID != nil {
		canCreate, err := s.userCanCreateInCompany(project.OwnerID, *project.CompanyID)
		if err != nil {
			return nil, err
		}
		if !canCreate {
			return nil, errors.New("user cannot create projects in this company")
		}
	}

	// Set defaults
	if project.Status == "" {
		project.Status = "active"
	}
	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now

	// Save to database
	if err := s.projectRepo.Create(project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) GetProject(id uint, userID string) (*db.BaseProject, error) {
	var project db.BaseProject
	if err := s.projectRepo.FindByID(id, &project); err != nil {
		return nil, err
	}

	// Business logic: check if user can access this project
	canAccess, err := s.userCanAccessProject(userID, &project)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, errors.New("user cannot access this project")
	}

	return &project, nil
}

func (s *ProjectService) UpdateProject(id uint, updates *db.BaseProject, userID string) (*db.BaseProject, error) {
	// Get existing project
	var project db.BaseProject
	if err := s.projectRepo.FindByID(id, &project); err != nil {
		return nil, err
	}

	// Business logic: check permissions
	canUpdate, err := s.userCanUpdateProject(userID, &project)
	if err != nil {
		return nil, err
	}
	if !canUpdate {
		return nil, errors.New("user cannot update this project")
	}

	// Update fields
	if updates.Title != "" {
		project.Title = updates.Title
	}
	if updates.Description != nil {
		project.Description = updates.Description
	}
	if updates.Status != "" {
		project.Status = updates.Status
	}
	if updates.StartDate != nil {
		project.StartDate = updates.StartDate
	}
	if updates.EndDate != nil {
		project.EndDate = updates.EndDate
	}
	project.UpdatedAt = time.Now()

	if err := s.projectRepo.Update(&project); err != nil {
		return nil, err
	}

	return &project, nil
}

func (s *ProjectService) DeleteProject(id uint, userID string) error {
	var project db.BaseProject
	if err := s.projectRepo.FindByID(id, &project); err != nil {
		return err
	}

	// Business logic: only owner can delete
	if project.OwnerID != userID {
		return errors.New("only project owner can delete project")
	}

	return s.projectRepo.Delete(&project)
}

func (s *ProjectService) GetUserProjects(userID string) ([]db.BaseProject, error) {
	// Get projects where user is owner
	var ownedProjects []db.BaseProject
	if err := s.projectRepo.FindWhere(&ownedProjects, "owner_id = ?", userID); err != nil {
		return nil, err
	}

	// Get projects where user is a member
	var members []db.ProjectMember
	if err := s.memberRepo.FindWhere(&members, "user_id = ?", userID); err != nil {
		return nil, err
	}

	var memberProjects []db.BaseProject
	for _, member := range members {
		// Convert string project ID to uint
		projectID, err := strconv.ParseUint(member.ProjectID, 10, 32)
		if err != nil {
			continue // Skip invalid project IDs
		}

		var project db.BaseProject
		if err := s.projectRepo.FindByID(uint(projectID), &project); err == nil {
			memberProjects = append(memberProjects, project)
		}
	}

	// Combine and deduplicate
	allProjects := append(ownedProjects, memberProjects...)
	return s.deduplicateProjects(allProjects), nil
}

func (s *ProjectService) AddProjectMember(projectID uint, userID, role string, permissions []string, requestingUserID string) (*db.ProjectMember, error) {
	var project db.BaseProject
	if err := s.projectRepo.FindByID(projectID, &project); err != nil {
		return nil, err
	}

	// Business logic: check if requesting user can add members
	canAddMembers, err := s.userCanManageProjectMembers(requestingUserID, &project)
	if err != nil {
		return nil, err
	}
	if !canAddMembers {
		return nil, errors.New("user cannot add members to this project")
	}

	// Check if user is already a member
	var existing db.ProjectMember
	err = s.memberRepo.FindOne(&existing, "project_id = ? AND user_id = ?", strconv.Itoa(int(projectID)), userID)
	if err == nil {
		// User already exists as member
		return nil, errors.New("user is already a member of this project")
	}

	member := &db.ProjectMember{
		ProjectID:   strconv.Itoa(int(projectID)),
		ProjectType: "core", // This is the core project manager
		UserID:      userID,
		Role:        role,
		Permissions: permissions,
		JoinedAt:    time.Now(),
	}

	if err := s.memberRepo.Create(member); err != nil {
		return nil, err
	}

	return member, nil
}

func (s *ProjectService) GetProjectMembers(projectID uint, requestingUserID string) ([]db.ProjectMember, error) {
	var project db.BaseProject
	if err := s.projectRepo.FindByID(projectID, &project); err != nil {
		return nil, err
	}

	// Check if user can view members
	canAccess, err := s.userCanAccessProject(requestingUserID, &project)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, errors.New("user cannot access this project")
	}

	var members []db.ProjectMember
	if err := s.memberRepo.FindWhere(&members, "project_id = ?", strconv.Itoa(int(projectID))); err != nil {
		return nil, err
	}

	return members, nil
}

// Private helper methods for business logic

func (s *ProjectService) userCanCreateInCompany(userID, companyID string) (bool, error) {
	var member db.CompanyMember
	err := s.companyMemberRepo.FindOne(&member, "company_id = ? AND user_id = ? AND status = ?", companyID, userID, "active")
	if err != nil {
		return false, nil // User not found in company
	}

	// Business rule: active members with admin, manager roles can create projects
	return member.Role == "admin" || member.Role == "manager" || member.Role == "owner", nil
}

func (s *ProjectService) userCanAccessProject(userID string, project *db.BaseProject) (bool, error) {
	// Owner can always access
	if project.OwnerID == userID {
		return true, nil
	}

	// Check if user is a project member
	var member db.ProjectMember
	err := s.memberRepo.FindOne(&member, "project_id = ? AND user_id = ?", strconv.Itoa(int(project.ID)), userID)
	if err == nil {
		return true, nil
	}

	// Check if user is in the same company (if project belongs to a company)
	if project.CompanyID != nil {
		var companyMember db.CompanyMember
		err := s.companyMemberRepo.FindOne(&companyMember, "company_id = ? AND user_id = ? AND status = ?", *project.CompanyID, userID, "active")
		if err == nil {
			return true, nil
		}
	}

	return false, nil
}

func (s *ProjectService) userCanUpdateProject(userID string, project *db.BaseProject) (bool, error) {
	// Owner can always update
	if project.OwnerID == userID {
		return true, nil
	}

	// Check if user is a project member with update permissions
	var member db.ProjectMember
	err := s.memberRepo.FindOne(&member, "project_id = ? AND user_id = ?", strconv.Itoa(int(project.ID)), userID)
	if err == nil {
		// Check if member has update permission
		for _, permission := range member.Permissions {
			if permission == "update" || permission == "admin" {
				return true, nil
			}
		}
	}

	return false, nil
}

func (s *ProjectService) userCanManageProjectMembers(userID string, project *db.BaseProject) (bool, error) {
	// Owner can always manage members
	if project.OwnerID == userID {
		return true, nil
	}

	// Check if user has member management permissions
	var member db.ProjectMember
	err := s.memberRepo.FindOne(&member, "project_id = ? AND user_id = ?", strconv.Itoa(int(project.ID)), userID)
	if err == nil {
		for _, permission := range member.Permissions {
			if permission == "manage_members" || permission == "admin" {
				return true, nil
			}
		}
	}

	return false, nil
}

func (s *ProjectService) deduplicateProjects(projects []db.BaseProject) []db.BaseProject {
	seen := make(map[uint]bool)
	var result []db.BaseProject

	for _, project := range projects {
		if !seen[project.ID] {
			seen[project.ID] = true
			result = append(result, project)
		}
	}

	return result
}
