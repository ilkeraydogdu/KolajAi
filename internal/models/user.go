package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID             int64     `json:"id" db:"id"`
	Email          string    `json:"email" db:"email"`
	Name           string    `json:"name" db:"name"`
	Password       string    `json:"-" db:"password"` // Backward compatibility
	PasswordHash   string    `json:"-" db:"password_hash"`
	Role           string    `json:"role" db:"role"` // admin, user, vendor
	IsActive       bool      `json:"is_active" db:"is_active"`
	IsAdmin        bool      `json:"is_admin" db:"is_admin"` // Backward compatibility
	EmailVerified  bool      `json:"email_verified" db:"email_verified"`
	TwoFactorEnabled bool   `json:"two_factor_enabled" db:"two_factor_enabled"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	LastLoginAt    *time.Time `json:"last_login_at" db:"last_login_at"`
	
	// Additional fields
	Phone          string    `json:"phone" db:"phone"`
	Address        string    `json:"address" db:"address"`
	City           string    `json:"city" db:"city"`
	Country        string    `json:"country" db:"country"`
	PostalCode     string    `json:"postal_code" db:"postal_code"`
	ProfilePicture string    `json:"profile_picture" db:"profile_picture"`
	
	// Vendor specific fields
	CompanyName    string    `json:"company_name,omitempty" db:"company_name"`
	TaxNumber      string    `json:"tax_number,omitempty" db:"tax_number"`
	VendorVerified bool      `json:"vendor_verified,omitempty" db:"vendor_verified"`
}

// GetPermissions returns user permissions based on role
func (u *User) GetPermissions() []string {
	switch u.Role {
	case "admin":
		return []string{"admin", "vendor", "user"}
	case "vendor":
		return []string{"vendor", "user"}
	case "user":
		return []string{"user"}
	default:
		return []string{}
	}
}

// HasAdminRole checks if user has admin role
func (u *User) HasAdminRole() bool {
	return u.Role == "admin" || u.IsAdmin // Check both for backward compatibility
}

// IsVendor checks if user has vendor role
func (u *User) IsVendor() bool {
	return u.Role == "vendor" || u.Role == "admin"
}
