package models

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// User represents a user in the system
type User struct {
	ID                     int64      `json:"id" db:"id"`
	Name                   string     `json:"name" db:"name"`
	Email                  string     `json:"email" db:"email"`
	Password               string     `json:"-" db:"password"` // Password is not exposed in JSON
	Phone                  string     `json:"phone" db:"phone"`
	Role                   string     `json:"role" db:"role"`
	IsActive               bool       `json:"is_active" db:"is_active"`
	IsAdmin                bool       `json:"is_admin" db:"is_admin"`
	IsSeller               bool       `json:"is_seller" db:"is_seller"`
	EmailVerified          bool       `json:"email_verified" db:"email_verified"`
	EmailVerificationToken string     `json:"-" db:"email_verification_token"`
	EmailVerifiedAt        *time.Time `json:"email_verified_at" db:"email_verified_at"`
	PasswordResetToken     string     `json:"-" db:"password_reset_token"`
	PasswordResetExpiry    *time.Time `json:"-" db:"password_reset_expiry"`
	LastLoginAt            *time.Time `json:"last_login_at" db:"last_login_at"`
	LastLoginIP            string     `json:"-" db:"last_login_ip"`
	LoginAttempts          int        `json:"-" db:"login_attempts"`
	LockedUntil            *time.Time `json:"-" db:"locked_until"`
	TwoFactorEnabled       bool       `json:"two_factor_enabled" db:"two_factor_enabled"`
	TwoFactorSecret        string     `json:"-" db:"two_factor_secret"`
	RememberToken          string     `json:"-" db:"remember_token"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
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

// IsLocked checks if the user account is locked
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return u.LockedUntil.After(time.Now())
}

// IncrementLoginAttempts increments failed login attempts
func (u *User) IncrementLoginAttempts() {
	u.LoginAttempts++
	if u.LoginAttempts >= 5 {
		lockDuration := 15 * time.Minute
		lockedUntil := time.Now().Add(lockDuration)
		u.LockedUntil = &lockedUntil
	}
}

// ResetLoginAttempts resets login attempts on successful login
func (u *User) ResetLoginAttempts() {
	u.LoginAttempts = 0
	u.LockedUntil = nil
}

// ValidatePassword checks password strength
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("şifre en az 8 karakter olmalıdır")
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)

	if !hasUpper {
		return errors.New("şifre en az bir büyük harf içermelidir")
	}
	if !hasLower {
		return errors.New("şifre en az bir küçük harf içermelidir")
	}
	if !hasNumber {
		return errors.New("şifre en az bir rakam içermelidir")
	}
	if !hasSpecial {
		return errors.New("şifre en az bir özel karakter içermelidir")
	}

	return nil
}
