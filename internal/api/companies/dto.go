package companies

import (
	"time"

	"github.com/JorgeSaicoski/go-project-manager/internal/db"
)

// Request DTOs

type CreateCompanyRequest struct {
	ID      string `json:"id" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Type    string `json:"type" binding:"required"` // enterprise, school, personal
	OwnerID string `json:"ownerId" binding:"required"`
}

type UpdateCompanyRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type AddMemberRequest struct {
	UserID     string   `json:"userId" binding:"required"`
	Role       string   `json:"role" binding:"required"`
	Salary     *float64 `json:"salary,omitempty"`
	HourlyRate *float64 `json:"hourlyRate,omitempty"`
}

// Response DTOs

type CompanyResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	OwnerID string `json:"ownerId"`
}

type CompanyMemberResponse struct {
	ID         uint       `json:"id"`
	CompanyID  string     `json:"companyId"`
	UserID     string     `json:"userId"`
	Role       string     `json:"role"`
	Status     string     `json:"status"`
	JoinedAt   *time.Time `json:"joinedAt"`
	InvitedAt  time.Time  `json:"invitedAt"`
	InvitedBy  string     `json:"invitedBy"`
	Salary     *float64   `json:"salary,omitempty"`
	HourlyRate *float64   `json:"hourlyRate,omitempty"`
}

type CompanyListResponse struct {
	Companies []CompanyResponse `json:"companies"`
	Total     int               `json:"total"`
}

type MemberListResponse struct {
	Members []CompanyMemberResponse `json:"members"`
	Total   int                     `json:"total"`
}

// Internal service request (for service-to-service calls)
type InternalCreateCompanyRequest struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	OwnerID string `json:"ownerId"`
}

// Conversion methods

func (r *CreateCompanyRequest) ToCompany() *db.Company {
	return &db.Company{
		ID:      r.ID,
		Name:    r.Name,
		Type:    r.Type,
		OwnerID: r.OwnerID,
	}
}

func (r *InternalCreateCompanyRequest) ToCompany() *db.Company {
	return &db.Company{
		ID:      r.ID,
		Name:    r.Name,
		Type:    r.Type,
		OwnerID: r.OwnerID,
	}
}

func CompanyToResponse(company *db.Company) CompanyResponse {
	return CompanyResponse{
		ID:      company.ID,
		Name:    company.Name,
		Type:    company.Type,
		OwnerID: company.OwnerID,
	}
}

func CompaniesToResponse(companies []db.Company) []CompanyResponse {
	responses := make([]CompanyResponse, len(companies))
	for i, company := range companies {
		responses[i] = CompanyToResponse(&company)
	}
	return responses
}

func MemberToResponse(member *db.CompanyMember) CompanyMemberResponse {
	return CompanyMemberResponse{
		ID:         member.ID,
		CompanyID:  member.CompanyID,
		UserID:     member.UserID,
		Role:       member.Role,
		Status:     member.Status,
		JoinedAt:   member.JoinedAt,
		InvitedAt:  member.InvitedAt,
		InvitedBy:  member.InvitedBy,
		Salary:     member.Salary,
		HourlyRate: member.HourlyRate,
	}
}

func MembersToResponse(members []db.CompanyMember) []CompanyMemberResponse {
	responses := make([]CompanyMemberResponse, len(members))
	for i, member := range members {
		responses[i] = MemberToResponse(&member)
	}
	return responses
}
