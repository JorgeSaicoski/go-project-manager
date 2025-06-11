package db

import (
	"time"
)

type BaseProject struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Status      string     `json:"status"` // active, completed, paused, cancelled
	OwnerID     string     `json:"ownerId"`
	CompanyID   *string    `json:"companyId,omitempty"`
	StartDate   *time.Time `json:"startDate"`
	EndDate     *time.Time `json:"endDate"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type ProjectMember struct {
	ProjectID   string    `json:"projectId"`   // External project ID
	ProjectType string    `json:"projectType"` // professional, education, finance
	UserID      string    `json:"userId"`
	Role        string    `json:"role"`
	Permissions []string  `json:"permissions" gorm:"type:text[]"`
	JoinedAt    time.Time `json:"joinedAt"`
}

type Company struct {
	ID      string          `json:"id" gorm:"primaryKey"`
	Name    string          `json:"name"`
	Type    string          `json:"type"` // enterprise, school, personal
	OwnerID string          `json:"ownerId"`
	Members []CompanyMember `json:"members" gorm:"foreignKey:CompanyID"`
}

type CompanyMember struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	CompanyID string     `json:"companyId"`
	UserID    string     `json:"userId"`
	Role      string     `json:"role"`     // admin, manager, employee, student, teacher
	Status    string     `json:"status"`   // active, invited, suspended
	JoinedAt  *time.Time `json:"joinedAt"` // nil if still invited
	InvitedAt time.Time  `json:"invitedAt"`
	InvitedBy string     `json:"invitedBy"` // UserID of who sent invitation

	// Company-specific data
	Salary     *float64 `json:"salary,omitempty"`     // For employees
	HourlyRate *float64 `json:"hourlyRate,omitempty"` // For freelancers/contractors
}
