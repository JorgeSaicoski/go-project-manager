# Project-Core Module

The foundational module for the Personal Manager system, providing shared project management functionality and infrastructure for specialized project types.

## 🎯 Purpose

Project-Core serves as the base layer for all project types in the Personal Manager ecosystem. It handles common functionality that all project types need: basic project information, company management, user membership, and permissions.

## 🏗️ Architecture Role

This module is part of a **microservices architecture** where specialized modules extend the base functionality:

```
Personal Manager System
├── project-core (this module) ← Base functionality
├── professional-tracker       ← Time tracking & freelance work  
├── education-manager          ← Courses & student management
└── finance-tracker           ← Financial goals & account tracking
```

## 📦 What This Module Provides

### Core Models
- **BaseProject**: Common project attributes (title, description, dates, status)
- **Company**: Organization management (enterprises, schools, personal companies)  
- **CompanyMember**: User-company relationships with roles and permissions
- **ProjectMember**: User-project relationships with custom roles

### Key Features
- ✅ **Multi-Company Support**: Users can belong to multiple organizations
- ✅ **Invitation System**: Companies can invite users with specific roles
- ✅ **Custom Permissions**: Flexible role-based access control
- ✅ **Project Ownership**: Clear ownership and collaboration rules
- ✅ **Company Types**: Support for enterprises, schools, and personal organizations

## 🔗 Integration with Specialized Modules

Specialized modules reference projects created in Project-Core:

```go
// In professional-tracker module
type ProfessionalProject struct {
    BaseProjectID string `json:"baseProjectId"` // Links to project-core
    // Professional-specific fields...
    ClientName    string `json:"clientName"`
    SalaryPerHour float64 `json:"salaryPerHour"`
}

// In education-manager module  
type EducationProject struct {
    BaseProjectID string `json:"baseProjectId"` // Links to project-core
    // Education-specific fields...
    CourseLevel   string `json:"courseLevel"`
    TeacherID     string `json:"teacherId"`
}
```

## 🚀 Getting Started

### Prerequisites
- Go 1.23+
- PostgreSQL
- Keycloak (for authentication)

### Installation
```bash
git clone https://github.com/JorgeSaicoski/go-project-manager.git
cd go-project-manager
go mod download
```

### Configuration
Set up your environment variables:
```bash
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=yourpassword
export POSTGRES_DB=project_core_db
export KEYCLOAK_PUBLIC_KEY=your_keycloak_public_key
```

### Run
```bash
go run cmd/server/main.go
```

## 📚 API Endpoints

### Projects
- `GET /projects` - List user's projects
- `POST /projects` - Create new project
- `GET /projects/{id}` - Get project details
- `PUT /projects/{id}` - Update project
- `DELETE /projects/{id}` - Delete project

### Companies
- `GET /companies` - List user's companies
- `POST /companies` - Create company
- `GET /companies/{id}/members` - List company members
- `POST /companies/{id}/invite` - Invite user to company

### Members & Permissions
- `GET /projects/{id}/members` - List project members
- `POST /projects/{id}/members` - Add member to project
- `PUT /projects/{id}/members/{userId}/permissions` - Update member permissions

## 🔧 Development

### Database Migrations
```bash
# Models auto-migrate on startup
# BaseProject, Company, CompanyMember, ProjectMember tables will be created
```

### Adding New Project Types
1. Create your specialized module (e.g., `my-new-tracker`)
2. Reference `BaseProject.ID` in your specialized model
3. Use Project-Core APIs for basic project operations
4. Implement your domain-specific logic separately

## 🏢 Company Types & Use Cases

### Enterprise Companies
- Multi-user organizations
- Employee project assignment
- Cost tracking and reporting
- Role-based project access

### Educational Institutions
- Student and teacher management
- Course-based project organization
- Payment and enrollment tracking
- Level-based access control

### Personal Companies
- Individual project organization
- Freelance work tracking  
- Personal goal management
- Private project spaces

## 🔐 Security & Permissions

### Role-Based Access Control
- Company owners create custom roles
- Roles have custom permissions
- Project-level permission inheritance
- Invitation-based membership

### Authentication
- Keycloak JWT integration
- User context in all operations
- Company-scoped data access
- Project ownership validation

## 🌟 Future Roadmap

- [ ] **Advanced Permissions**: Hierarchical role inheritance
- [ ] **Company Templates**: Pre-configured company types
- [ ] **Bulk Operations**: Multi-project management
- [ ] **Audit Logging**: Track all changes and access
- [ ] **Company Analytics**: Member activity and project statistics

## 🤝 Contributing

This module is part of the larger Personal Manager system. See the [main project documentation](https://github.com/JorgeSaicoski/personal-manager) for:
- Overall system architecture
- Planned specialized modules
- Development guidelines
- Contribution process

## 📄 License

MIT License - See [LICENSE](LICENSE) for details.

---

**Note**: This is a foundational module. For complete functionality, you'll need to integrate with specialized modules for your specific use case (professional tracking, education management, or finance tracking).