package companies

import (
	"time"

	"github.com/JorgeSaicoski/go-project-manager/internal/db"
	"github.com/JorgeSaicoski/go-project-manager/internal/services/companies"
	"github.com/JorgeSaicoski/microservice-commons/responses"
	"github.com/JorgeSaicoski/microservice-commons/types"
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

func (h *CompanyHandler) CreateCompany(c *gin.Context) {
	var req InternalCreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.BadRequest(c, err.Error())
		return
	}

	company, err := h.companyService.CreateCompany(req.ToCompany())
	if err != nil {
		responses.InternalError(c, err.Error())
		return
	}

	response := CompanyToResponse(company)
	responses.Created(c, "Company created successfully", response)
}

func (h *CompanyHandler) GetCompany(c *gin.Context) {
	companyID := c.Param("id")

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

	company, err := h.companyService.GetCompany(companyID, userID)
	if err != nil {
		if err.Error() == "user cannot access this company" {
			responses.Forbidden(c, err.Error())
			return
		}
		responses.NotFound(c, err.Error())
		return
	}

	response := CompanyToResponse(company)
	responses.Success(c, "Company retrieved successfully", response)
}

func (h *CompanyHandler) UpdateCompany(c *gin.Context) {
	companyID := c.Param("id")

	var req struct {
		UpdateCompanyRequest
		UserID string `json:"userId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.BadRequest(c, err.Error())
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
			responses.Forbidden(c, err.Error())
			return
		}
		responses.InternalError(c, err.Error())
		return
	}

	response := CompanyToResponse(company)
	responses.Success(c, "Company updated successfully", response)
}

func (h *CompanyHandler) DeleteCompany(c *gin.Context) {
	companyID := c.Param("id")

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

	err := h.companyService.DeleteCompany(companyID, userID)
	if err != nil {
		if err.Error() == "only company owner can delete company" {
			responses.Forbidden(c, err.Error())
			return
		}
		responses.InternalError(c, err.Error())
		return
	}

	responses.Success(c, "Company deleted successfully", nil)
}

func (h *CompanyHandler) GetUserCompanies(c *gin.Context) {
	userID := c.Query("userId")
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}

	if userID == "" {
		responses.BadRequest(c, "User ID required")
		return
	}

	companies, err := h.companyService.GetUserCompanies(userID)
	if err != nil {
		responses.InternalError(c, err.Error())
		return
	}

	companyResponses := CompaniesToResponse(companies)
	response := types.ListResponse[CompanyResponse]{
		Data: companyResponses,
		Meta: types.ResponseMetadata{
			Count:     len(companyResponses),
			Timestamp: time.Now(),
		},
	}

	responses.Success(c, "Companies retrieved successfully", response)
}

func (h *CompanyHandler) GetCompanyMembers(c *gin.Context) {
	companyID := c.Param("id")

	userID := c.Query("userId")
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}

	if userID == "" {
		responses.BadRequest(c, "User ID required")
		return
	}

	members, err := h.companyService.GetCompanyMembers(companyID, userID)
	if err != nil {
		if err.Error() == "user cannot access this company" {
			responses.Forbidden(c, err.Error())
			return
		}
		responses.InternalError(c, err.Error())
		return
	}

	memberResponses := MembersToResponse(members)
	response := types.ListResponse[CompanyMemberResponse]{
		Data: memberResponses,
		Meta: types.ResponseMetadata{
			Count:     len(memberResponses),
			Timestamp: time.Now(),
		},
	}
	responses.Success(c, "Members retrieved successfully", response)
}

func (h *CompanyHandler) AddCompanyMember(c *gin.Context) {
	companyID := c.Param("id")

	var req struct {
		AddMemberRequest
		RequestingUserID string `json:"requestingUserId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		responses.BadRequest(c, err.Error())
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
			responses.Forbidden(c, err.Error())
			return
		}
		if err.Error() == "user is already a member of this company" {
			responses.Conflict(c, err.Error())
			return
		}
		responses.InternalError(c, err.Error())
		return
	}

	if req.Salary != nil {
		member.Salary = req.Salary
	}
	if req.HourlyRate != nil {
		member.HourlyRate = req.HourlyRate
	}

	response := MemberToResponse(member)
	responses.Created(c, "Member added successfully", response)
}

func (h *CompanyHandler) RemoveCompanyMember(c *gin.Context) {
	companyID := c.Param("id")
	userID := c.Param("userId")

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
		responses.BadRequest(c, "Requesting User ID required")
		return
	}

	err := h.companyService.RemoveCompanyMember(companyID, userID, requestingUserID)
	if err != nil {
		if err.Error() == "user cannot remove members from this company" {
			responses.Forbidden(c, err.Error())
			return
		}
		if err.Error() == "cannot remove company owner" {
			responses.Forbidden(c, err.Error())
			return
		}
		responses.InternalError(c, err.Error())
		return
	}

	responses.Success(c, "Member removed successfully", nil)
}
