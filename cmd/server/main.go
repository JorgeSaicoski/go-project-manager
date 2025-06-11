package main

import (
	"github.com/JorgeSaicoski/go-project-manager/internal/api/companies"
	"github.com/JorgeSaicoski/go-project-manager/internal/api/projects"
	"github.com/JorgeSaicoski/go-project-manager/internal/db"
	companiesService "github.com/JorgeSaicoski/go-project-manager/internal/services/companies"
	projectsService "github.com/JorgeSaicoski/go-project-manager/internal/services/projects"
	"github.com/JorgeSaicoski/microservice-commons/config"
	"github.com/JorgeSaicoski/microservice-commons/database"
	"github.com/JorgeSaicoski/microservice-commons/server"
	"github.com/gin-gonic/gin"
)

func main() {
	server := server.NewServer(server.ServerOptions{
		ServiceName:    "project-core",
		ServiceVersion: "1.0.0",
		SetupRoutes:    setupRoutes,
	})
	server.Start()
}

func setupRoutes(router *gin.Engine, cfg *config.Config) {
	// Connect to database using microservice-commons
	dbConnection, err := database.ConnectWithConfig(cfg.DatabaseConfig)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	// Auto-migrate models
	if err := database.QuickMigrate(dbConnection, &db.BaseProject{}, &db.ProjectMember{}, &db.Company{}, &db.CompanyMember{}); err != nil {
		panic("Failed to migrate database: " + err.Error())
	}

	// Initialize services
	projectService := projectsService.NewProjectService(dbConnection)
	companyService := companiesService.NewCompanyService(dbConnection)

	// Setup routes
	api := router.Group("/api")
	projects.RegisterRoutes(api, projectService)
	companies.RegisterRoutes(api, companyService)
}
