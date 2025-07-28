package models

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// UserRole represents different user roles in the system
type UserRole string

const (
	RoleUser     UserRole = "user"
	RoleVendor   UserRole = "vendor"
	RoleAdmin    UserRole = "admin"
	RoleModerator UserRole = "moderator"
	RoleSupport  UserRole = "support"
)

// UserPermission represents specific permissions
type UserPermission struct {
	ID          int64     `json:"id" db:"id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	Permission  string    `json:"permission" db:"permission"`
	Resource    string    `json:"resource" db:"resource"`
	GrantedAt   time.Time `json:"granted_at" db:"granted_at"`
	GrantedBy   int64     `json:"granted_by" db:"granted_by"`
}

// User represents a user in the system
type User struct {
	ID              int64            `json:"id" db:"id"`
	Name            string           `json:"name" db:"name"`
	Email           string           `json:"email" db:"email"`
	Password        string           `json:"-" db:"password"` // Password is not exposed in JSON
	Phone           string           `json:"phone" db:"phone"`
	IsActive        bool             `json:"is_active" db:"is_active"`
	IsAdmin         bool             `json:"is_admin" db:"is_admin"`
	Role            UserRole         `json:"role" db:"role"`
	Permissions     []UserPermission `json:"permissions,omitempty"`
	AIAccess        bool             `json:"ai_access" db:"ai_access"`
	AIEditAccess    bool             `json:"ai_edit_access" db:"ai_edit_access"`
	AITemplateAccess bool            `json:"ai_template_access" db:"ai_template_access"`
	CreatedAt       time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at" db:"updated_at"`
}

// HasPermission checks if user has specific permission
func (u *User) HasPermission(permission, resource string) bool {
	// Admin has all permissions
	if u.IsAdmin || u.Role == RoleAdmin {
		return true
	}

	// Check specific permissions
	for _, perm := range u.Permissions {
		if perm.Permission == permission && (perm.Resource == resource || perm.Resource == "*") {
			return true
		}
	}

	return false
}

// CanUseAI checks if user can use AI features
func (u *User) CanUseAI() bool {
	return u.AIAccess || u.IsAdmin || u.Role == RoleAdmin
}

// CanEditWithAI checks if user can edit products with AI
func (u *User) CanEditWithAI() bool {
	return u.AIEditAccess || u.IsAdmin || u.Role == RoleAdmin
}

// CanUseAITemplates checks if user can use AI template features
func (u *User) CanUseAITemplates() bool {
	return u.AITemplateAccess || u.IsAdmin || u.Role == RoleAdmin
}

// Validate checks if the user data is valid
func (u *User) Validate() error {
	if strings.TrimSpace(u.Name) == "" {
		return errors.New("name cannot be empty")
	}

	if strings.TrimSpace(u.Email) == "" {
		return errors.New("email cannot be empty")
	}

	// Basic email validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(u.Email) {
		return errors.New("invalid email format")
	}

	if strings.TrimSpace(u.Password) == "" {
		return errors.New("password cannot be empty")
	}

	// Validate role
	validRoles := []UserRole{RoleUser, RoleVendor, RoleAdmin, RoleModerator, RoleSupport}
	roleValid := false
	for _, validRole := range validRoles {
		if u.Role == validRole {
			roleValid = true
			break
		}
	}
	if !roleValid {
		u.Role = RoleUser // Default to user role
	}

	return nil
}
