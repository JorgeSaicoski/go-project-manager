package companies

import (
	"net/http"

	"github.com/JorgeSaicoski/go-project-manager/internal/db"
	"github.com/JorgeSaicoski/go-project-manager/internal/services/companies"
	"github.com/gin-gonic/gin"
)

type CompanyHandler struct {
	companyService *companies.CompanyService
}

func NewCompanyHandler(companyService *companies.CompanyService) *CompanyHandler {
	return &CompanyHandler{
		companyService: companyService,
	}
}

// CreateCompany handles internal company creation requests
func (h *CompanyHandler) CreateCompany(c *gin.Context) {
	var req InternalCreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	company, err := h.companyService.CreateCompany(req.ToCompany())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := CompanyToResponse(company)
	c.JSON(http.StatusCreated, response)
}

// GetCompany retrieves a company by ID
func (h *CompanyHandler) GetCompany(c *gin.Context) {
	companyID := c.Param("id")

	// Get requesting user ID from request body or header
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

	company, err := h.companyService.GetCompany(companyID, userID)
	if err != nil {
		if err.Error() == "user cannot access this company" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	response := CompanyToResponse(company)
	c.JSON(http.StatusOK, response)
}

// UpdateCompany updates a company
func (h *CompanyHandler) UpdateCompany(c *gin.Context) {
	companyID := c.Param("id")

	var req struct {
		UpdateCompanyRequest
		UserID string `json:"userId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := &req.UpdateCompanyRequest
	companyUpdates := &db.Company{
		Name: updates.Name,
		Type: updates.Type,
	}

	company, err := h.companyService.UpdateCompany(companyID, companyUpdates, req.UserID)
	if err != nil {
		if err.Error() == "user cannot update this company" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := CompanyToResponse(company)
	c.JSON(http.StatusOK, response)
}

// DeleteCompany deletes a company
func (h *CompanyHandler) DeleteCompany(c *gin.Context) {
	companyID := c.Param("id")

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

	err := h.companyService.DeleteCompany(companyID, userID)
	if err != nil {
		if err.Error() == "only company owner can delete company" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Company deleted successfully"})
}

// GetUserCompanies retrieves all companies for a user
func (h *CompanyHandler) GetUserCompanies(c *gin.Context) {
	// Get user ID from query parameter for internal calls
	userID := c.Query("userId")
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	companies, err := h.companyService.GetUserCompanies(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses := CompaniesToResponse(companies)
	response := CompanyListResponse{
		Companies: responses,
		Total:     len(responses),
	}

	c.JSON(http.StatusOK, response)
}

// GetCompanyMembers retrieves all members of a company
func (h *CompanyHandler) GetCompanyMembers(c *gin.Context) {
	companyID := c.Param("id")

	// Get requesting user ID
	userID := c.Query("userId")
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	members, err := h.companyService.GetCompanyMembers(companyID, userID)
	if err != nil {
		if err.Error() == "user cannot access this company" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses := MembersToResponse(members)
	response := MemberListResponse{
		Members: responses,
		Total:   len(responses),
	}
	c.JSON(http.StatusOK, response)
}

// AddCompanyMember adds a member to a company
func (h *CompanyHandler) AddCompanyMember(c *gin.Context) {
	companyID := c.Param("id")

	var req struct {
		AddMemberRequest
		RequestingUserID string `json:"requestingUserId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := h.companyService.AddCompanyMember(
		companyID,
		req.UserID,
		req.Role,
		req.RequestingUserID,
	)
	if err != nil {
		if err.Error() == "user cannot add members to this company" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "user is already a member of this company" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update salary/hourly rate if provided
	if req.Salary != nil {
		member.Salary = req.Salary
	}
	if req.HourlyRate != nil {
		member.HourlyRate = req.HourlyRate
	}

	response := MemberToResponse(member)
	c.JSON(http.StatusCreated, response)
}

// RemoveCompanyMember removes a member from a company
func (h *CompanyHandler) RemoveCompanyMember(c *gin.Context) {
	companyID := c.Param("id")
	userID := c.Param("userId")

	// Get requesting user ID
	requestingUserID := c.GetHeader("X-User-ID")
	if requestingUserID == "" {
		var req struct {
			RequestingUserID string `json:"requestingUserId"`
		}
		if err := c.ShouldBindJSON(&req); err == nil {
			requestingUserID = req.RequestingUserID
		}
	}

	if requestingUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Requesting User ID required"})
		return
	}

	err := h.companyService.RemoveCompanyMember(companyID, userID, requestingUserID)
	if err != nil {
		if err.Error() == "user cannot remove members from this company" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "cannot remove company owner" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
}
