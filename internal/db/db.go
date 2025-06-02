package db

import (
	"fmt"
	"os"
	"time"

	"github.com/JorgeSaicoski/pgconnect"
	"gorm.io/gorm"
)

var DB *pgconnect.DB

type Project struct {
	gorm.Model
	Title       string  `json:"title" gorm:"not null"`
	Description *string `json:"description,omitempty"`
	Status      string  `json:"status" gorm:"default:'active'"` // active, completed, paused, cancelled
	Type        string  `json:"type" gorm:"not null"`           // personal, freelance, study, work

	// Owner and company relationship
	OwnerID   string  `json:"ownerId" gorm:"not null;index"`    // User who created the project
	CompanyID *string `json:"companyId,omitempty" gorm:"index"` // Optional company association

	// Time tracking
	EstimatedHours *float64 `json:"estimatedHours,omitempty"`
	ActualHours    float64  `json:"actualHours" gorm:"default:0"`

	// Dates
	StartDate *time.Time `json:"startDate,omitempty"`
	DueDate   *time.Time `json:"dueDate,omitempty"`
	EndDate   *time.Time `json:"endDate,omitempty"`

	// Metadata
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// ProjectMember represents users assigned to a project
type ProjectMember struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ProjectID uint      `json:"projectId" gorm:"not null;index"`
	UserID    string    `json:"userId" gorm:"not null;index"`
	Role      string    `json:"role" gorm:"default:'member'"` // owner, manager, member, viewer
	JoinedAt  time.Time `json:"joinedAt" gorm:"autoCreateTime"`

	// Relationships
	Project Project `json:"-" gorm:"foreignKey:ProjectID"`
}

// ProjectPermission defines what actions a user can perform on a project
type ProjectPermission struct {
	ID               uint `json:"id" gorm:"primaryKey"`
	ProjectMemberID  uint `json:"projectMemberId" gorm:"not null;index"`
	CanEdit          bool `json:"canEdit" gorm:"default:false"`
	CanDelete        bool `json:"canDelete" gorm:"default:false"`
	CanManageMembers bool `json:"canManageMembers" gorm:"default:false"`
	CanViewTime      bool `json:"canViewTime" gorm:"default:true"`
	CanEditTime      bool `json:"canEditTime" gorm:"default:false"`

	// Relationships
	ProjectMember ProjectMember `json:"-" gorm:"foreignKey:ProjectMemberID"`
}

func ConnectDatabase() {
	// Create config using environment variables
	config := pgconnect.DefaultConfig()

	// Override with environment variables if they exist
	config.Host = getEnv("POSTGRES_HOST", config.Host)
	config.Port = getEnv("POSTGRES_PORT", config.Port)
	config.User = getEnv("POSTGRES_USER", config.User)
	config.Password = getEnv("POSTGRES_PASSWORD", config.Password)
	config.DatabaseName = getEnv("POSTGRES_DB", "taskdb")
	config.SSLMode = getEnv("POSTGRES_SSLMODE", config.SSLMode)
	config.TimeZone = getEnv("POSTGRES_TIMEZONE", config.TimeZone)

	// Retry loop for database connection
	var err error
	maxRetries := 3
	retryDelay := 30 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Printf("Attempting to connect to database (attempt %d of %d)\n", attempt, maxRetries)

		DB, err = pgconnect.New(config)
		if err == nil {
			fmt.Println("Successfully connected to database")
			break
		}

		fmt.Printf("Failed to connect to database: %v\n", err)

		if attempt < maxRetries {
			fmt.Printf("Retrying in %v...\n", retryDelay)
			time.Sleep(retryDelay)
		} else {
			panic("Failed to connect to database after maximum retry attempts")
		}
	}

	// Auto migrate the Task model
	DB.AutoMigrate(&Project{}, &ProjectMember{}, &ProjectPermission{})
}

// Helper function to get environment variable with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
