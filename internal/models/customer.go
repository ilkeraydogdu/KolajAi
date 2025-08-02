package models

import (
	"errors"
	"strings"
	"time"
)

// Customer represents a customer in the system
type Customer struct {
	ID                int64     `json:"id" db:"id"`
	UserID            int64     `json:"user_id" db:"user_id"`
	FirstName         string    `json:"first_name" db:"first_name"`
	LastName          string    `json:"last_name" db:"last_name"`
	Email             string    `json:"email" db:"email"`
	Phone             string    `json:"phone" db:"phone"`
	DateOfBirth       *time.Time `json:"date_of_birth" db:"date_of_birth"`
	Gender            string    `json:"gender" db:"gender"` // male, female, other
	CustomerType      string    `json:"customer_type" db:"customer_type"` // individual, corporate
	CompanyName       string    `json:"company_name" db:"company_name"`
	TaxNumber         string    `json:"tax_number" db:"tax_number"`
	PreferredLanguage string    `json:"preferred_language" db:"preferred_language"`
	Newsletter        bool      `json:"newsletter" db:"newsletter"`
	SMSNotifications  bool      `json:"sms_notifications" db:"sms_notifications"`
	EmailNotifications bool     `json:"email_notifications" db:"email_notifications"`
	LoyaltyPoints     int       `json:"loyalty_points" db:"loyalty_points"`
	TotalSpent        float64   `json:"total_spent" db:"total_spent"`
	OrderCount        int       `json:"order_count" db:"order_count"`
	LastOrderDate     *time.Time `json:"last_order_date" db:"last_order_date"`
	Status            string    `json:"status" db:"status"` // active, inactive, blocked
	Notes             string    `json:"notes" db:"notes"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
	
	// Related data
	Addresses         []Address `json:"addresses,omitempty"`
	User              *User     `json:"user,omitempty"`
}

// Address represents a customer address
type Address struct {
	ID           int64     `json:"id" db:"id"`
	CustomerID   int64     `json:"customer_id" db:"customer_id"`
	Type         string    `json:"type" db:"type"` // billing, shipping, both
	Title        string    `json:"title" db:"title"` // Home, Work, etc.
	FirstName    string    `json:"first_name" db:"first_name"`
	LastName     string    `json:"last_name" db:"last_name"`
	Company      string    `json:"company" db:"company"`
	AddressLine1 string    `json:"address_line1" db:"address_line1"`
	AddressLine2 string    `json:"address_line2" db:"address_line2"`
	City         string    `json:"city" db:"city"`
	State        string    `json:"state" db:"state"`
	PostalCode   string    `json:"postal_code" db:"postal_code"`
	Country      string    `json:"country" db:"country"`
	Phone        string    `json:"phone" db:"phone"`
	IsDefault    bool      `json:"is_default" db:"is_default"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Validate checks if the customer data is valid
func (c *Customer) Validate() error {
	if strings.TrimSpace(c.FirstName) == "" {
		return errors.New("first name cannot be empty")
	}
	
	if strings.TrimSpace(c.LastName) == "" {
		return errors.New("last name cannot be empty")
	}
	
	if strings.TrimSpace(c.Email) == "" {
		return errors.New("email cannot be empty")
	}
	
	if c.UserID <= 0 {
		return errors.New("valid user ID is required")
	}
	
	if c.CustomerType != "" && c.CustomerType != "individual" && c.CustomerType != "corporate" {
		return errors.New("customer type must be individual or corporate")
	}
	
	return nil
}

// GetFullName returns the customer's full name
func (c *Customer) GetFullName() string {
	return strings.TrimSpace(c.FirstName + " " + c.LastName)
}

// Validate checks if the address data is valid
func (a *Address) Validate() error {
	if strings.TrimSpace(a.FirstName) == "" {
		return errors.New("first name cannot be empty")
	}
	
	if strings.TrimSpace(a.LastName) == "" {
		return errors.New("last name cannot be empty")
	}
	
	if strings.TrimSpace(a.AddressLine1) == "" {
		return errors.New("address line 1 cannot be empty")
	}
	
	if strings.TrimSpace(a.City) == "" {
		return errors.New("city cannot be empty")
	}
	
	if strings.TrimSpace(a.Country) == "" {
		return errors.New("country cannot be empty")
	}
	
	if a.CustomerID <= 0 {
		return errors.New("valid customer ID is required")
	}
	
	if a.Type != "" && a.Type != "billing" && a.Type != "shipping" && a.Type != "both" {
		return errors.New("address type must be billing, shipping, or both")
	}
	
	return nil
}

// GetFullAddress returns the formatted full address
func (a *Address) GetFullAddress() string {
	parts := []string{}
	
	if a.AddressLine1 != "" {
		parts = append(parts, a.AddressLine1)
	}
	if a.AddressLine2 != "" {
		parts = append(parts, a.AddressLine2)
	}
	if a.City != "" {
		parts = append(parts, a.City)
	}
	if a.State != "" {
		parts = append(parts, a.State)
	}
	if a.PostalCode != "" {
		parts = append(parts, a.PostalCode)
	}
	if a.Country != "" {
		parts = append(parts, a.Country)
	}
	
	return strings.Join(parts, ", ")
}