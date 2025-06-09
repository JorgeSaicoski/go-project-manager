package db

import (
	"fmt"
	"os"
	"time"

	"github.com/JorgeSaicoski/pgconnect"
)

var DB *pgconnect.DB

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

	// Auto migrate all models
	DB.AutoMigrate(&BaseProject{}, &ProjectMember{}, &Company{}, &CompanyMember{})
}

// Helper function to get environment variable with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
