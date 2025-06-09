package companies

import (
	"github.com/JorgeSaicoski/go-project-manager/internal/services/companies"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all company-related routes
func RegisterRoutes(router *gin.RouterGroup, companyService *companies.CompanyService) {
	handler := NewCompanyHandler(companyService)

	// Internal API routes for service-to-service communication
	internal := router.Group("/internal/companies")
	{
		// Company CRUD
		internal.POST("", handler.CreateCompany)       // Create company
		internal.GET("/:id", handler.GetCompany)       // Get company by ID
		internal.PUT("/:id", handler.UpdateCompany)    // Update company
		internal.DELETE("/:id", handler.DeleteCompany) // Delete company

		// User companies
		internal.GET("", handler.GetUserCompanies) // Get user's companies (query: userId)

		// Company members
		internal.GET("/:id/members", handler.GetCompanyMembers)              // Get company members
		internal.POST("/:id/members", handler.AddCompanyMember)              // Add member to company
		internal.DELETE("/:id/members/:userId", handler.RemoveCompanyMember) // Remove member from company
	}
}
