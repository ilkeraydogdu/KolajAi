package models

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// User represents a user in the system
type User struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"` // Password is not exposed in JSON
	Phone     string    `json:"phone" db:"phone"`
	Role      string    `json:"role" db:"role"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	IsAdmin   bool      `json:"is_admin" db:"is_admin"`
	IsSeller  bool      `json:"is_seller" db:"is_seller"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
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

	return nil
}
