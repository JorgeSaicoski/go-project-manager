package companies

import (
	"errors"
	"time"

	"github.com/JorgeSaicoski/go-project-manager/internal/db"
	"github.com/JorgeSaicoski/pgconnect"
)

type CompanyService struct {
	companyRepo       *pgconnect.Repository[db.Company]
	companyMemberRepo *pgconnect.Repository[db.CompanyMember]
}

func NewCompanyService(database *pgconnect.DB) *CompanyService {
	return &CompanyService{
		companyRepo:       pgconnect.NewRepository[db.Company](database),
		companyMemberRepo: pgconnect.NewRepository[db.CompanyMember](database),
	}
}

func (s *CompanyService) CreateCompany(company *db.Company) (*db.Company, error) {
	// Set the owner as the ID
	if company.ID == "" {
		return nil, errors.New("company ID is required")
	}

	// Save to database
	if err := s.companyRepo.Create(company); err != nil {
		return nil, err
	}

	// Add the owner as a company member with admin role
	ownerMember := &db.CompanyMember{
		CompanyID: company.ID,
		UserID:    company.OwnerID,
		Role:      "owner",
		Status:    "active",
		JoinedAt:  &time.Time{},
		InvitedAt: time.Now(),
		InvitedBy: company.OwnerID,
	}
	now := time.Now()
	ownerMember.JoinedAt = &now

	if err := s.companyMemberRepo.Create(ownerMember); err != nil {
		// Rollback company creation if member creation fails
		s.companyRepo.Delete(company)
		return nil, err
	}

	return company, nil
}

func (s *CompanyService) GetCompany(id string, userID string) (*db.Company, error) {
	var company db.Company
	if err := s.companyRepo.FindByID(id, &company); err != nil {
		return nil, err
	}

	// Check if user can access this company
	canAccess, err := s.userCanAccessCompany(userID, id)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, errors.New("user cannot access this company")
	}

	return &company, nil
}

func (s *CompanyService) UpdateCompany(id string, updates *db.Company, userID string) (*db.Company, error) {
	// Get existing company
	var company db.Company
	if err := s.companyRepo.FindByID(id, &company); err != nil {
		return nil, err
	}

	// Check permissions - only owner or admin can update
	canUpdate, err := s.userCanUpdateCompany(userID, id)
	if err != nil {
		return nil, err
	}
	if !canUpdate {
		return nil, errors.New("user cannot update this company")
	}

	// Update fields
	if updates.Name != "" {
		company.Name = updates.Name
	}
	if updates.Type != "" {
		company.Type = updates.Type
	}

	if err := s.companyRepo.Update(&company); err != nil {
		return nil, err
	}

	return &company, nil
}

func (s *CompanyService) DeleteCompany(id string, userID string) error {
	var company db.Company
	if err := s.companyRepo.FindByID(id, &company); err != nil {
		return err
	}

	// Only owner can delete company
	if company.OwnerID != userID {
		return errors.New("only company owner can delete company")
	}

	// Delete all company members first
	if err := s.companyMemberRepo.DeleteWhere("company_id = ?", id); err != nil {
		return err
	}

	// Delete company
	return s.companyRepo.Delete(&company)
}

func (s *CompanyService) GetUserCompanies(userID string) ([]db.Company, error) {
	// Get companies where user is a member
	var members []db.CompanyMember
	if err := s.companyMemberRepo.FindWhere(&members, "user_id = ? AND status = ?", userID, "active"); err != nil {
		return nil, err
	}

	var companies []db.Company
	for _, member := range members {
		var company db.Company
		if err := s.companyRepo.FindByID(member.CompanyID, &company); err == nil {
			companies = append(companies, company)
		}
	}

	return companies, nil
}

func (s *CompanyService) GetCompanyMembers(companyID string, requestingUserID string) ([]db.CompanyMember, error) {
	// Check if user can view members
	canAccess, err := s.userCanAccessCompany(requestingUserID, companyID)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, errors.New("user cannot access this company")
	}

	var members []db.CompanyMember
	if err := s.companyMemberRepo.FindWhere(&members, "company_id = ?", companyID); err != nil {
		return nil, err
	}

	return members, nil
}

func (s *CompanyService) AddCompanyMember(companyID, userID, role string, requestingUserID string) (*db.CompanyMember, error) {
	// Check if requesting user can add members
	canManage, err := s.userCanManageCompanyMembers(requestingUserID, companyID)
	if err != nil {
		return nil, err
	}
	if !canManage {
		return nil, errors.New("user cannot add members to this company")
	}

	// Check if user is already a member
	var existing db.CompanyMember
	err = s.companyMemberRepo.FindOne(&existing, "company_id = ? AND user_id = ?", companyID, userID)
	if err == nil {
		return nil, errors.New("user is already a member of this company")
	}

	member := &db.CompanyMember{
		CompanyID: companyID,
		UserID:    userID,
		Role:      role,
		Status:    "active",
		InvitedAt: time.Now(),
		InvitedBy: requestingUserID,
	}
	now := time.Now()
	member.JoinedAt = &now

	if err := s.companyMemberRepo.Create(member); err != nil {
		return nil, err
	}

	return member, nil
}

func (s *CompanyService) RemoveCompanyMember(companyID, userID string, requestingUserID string) error {
	// Check permissions
	canManage, err := s.userCanManageCompanyMembers(requestingUserID, companyID)
	if err != nil {
		return err
	}

	// Users can remove themselves
	isSelfRemoval := userID == requestingUserID
	if !canManage && !isSelfRemoval {
		return errors.New("user cannot remove members from this company")
	}

	// Cannot remove company owner
	var company db.Company
	if err := s.companyRepo.FindByID(companyID, &company); err != nil {
		return err
	}
	if company.OwnerID == userID {
		return errors.New("cannot remove company owner")
	}

	// Remove member
	return s.companyMemberRepo.DeleteWhere("company_id = ? AND user_id = ?", companyID, userID)
}

// Private helper methods

func (s *CompanyService) userCanAccessCompany(userID, companyID string) (bool, error) {
	var member db.CompanyMember
	err := s.companyMemberRepo.FindOne(&member, "company_id = ? AND user_id = ? AND status = ?", companyID, userID, "active")
	return err == nil, nil
}

func (s *CompanyService) userCanUpdateCompany(userID, companyID string) (bool, error) {
	var company db.Company
	if err := s.companyRepo.FindByID(companyID, &company); err != nil {
		return false, err
	}

	// Owner can always update
	if company.OwnerID == userID {
		return true, nil
	}

	// Check if user is admin
	var member db.CompanyMember
	err := s.companyMemberRepo.FindOne(&member, "company_id = ? AND user_id = ? AND role = ? AND status = ?", companyID, userID, "admin", "active")
	return err == nil, nil
}

func (s *CompanyService) userCanManageCompanyMembers(userID, companyID string) (bool, error) {
	var company db.Company
	if err := s.companyRepo.FindByID(companyID, &company); err != nil {
		return false, err
	}

	// Owner can always manage
	if company.OwnerID == userID {
		return true, nil
	}

	// Check if user is admin or manager
	var member db.CompanyMember
	err := s.companyMemberRepo.FindOne(&member, "company_id = ? AND user_id = ? AND status = ?", companyID, userID, "active")
	if err != nil {
		return false, nil
	}

	return member.Role == "admin" || member.Role == "manager", nil
}
