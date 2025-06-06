package api

import (
	"github.com/JorgeSaicoski/go-project-manager/internal/db"
	"github.com/JorgeSaicoski/pgconnect"
)

// Handler encapsulates all the task-related API handlers
type ProjectHandler struct {
	repo *pgconnect.Repository[db.BaseProject]
}

// NewProjectHandler creates and returns a new ProjectHandler instance
func NewProjectHandler(database *pgconnect.DB) *ProjectHandler {
	return &ProjectHandler{
		repo: pgconnect.NewRepository[db.BaseProject](database),
	}
}
