package models

import (
	"time"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Customer represents a customer in the system
type Customer struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"index;not null"`
	User        User      `json:"user" gorm:"foreignKey:UserID"`
	
	// Personal Information
	FirstName   string    `json:"first_name" gorm:"size:100;not null" validate:"required,min=2,max=100"`
	LastName    string    `json:"last_name" gorm:"size:100;not null" validate:"required,min=2,max=100"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	Gender      string    `json:"gender" gorm:"size:10" validate:"omitempty,oneof=male female other"`
	
	// Contact Information
	Phone       string    `json:"phone" gorm:"size:20" validate:"omitempty,e164"`
	AlternatePhone string `json:"alternate_phone" gorm:"size:20" validate:"omitempty,e164"`
	
	// Preferences
	Language    string    `json:"language" gorm:"size:5;default:'tr'" validate:"omitempty,len=2"`
	Currency    string    `json:"currency" gorm:"size:3;default:'TRY'" validate:"omitempty,len=3"`
	Newsletter  bool      `json:"newsletter" gorm:"default:false"`
	SMSNotifications bool `json:"sms_notifications" gorm:"default:false"`
	
	// Customer Status
	Status      CustomerStatus `json:"status" gorm:"default:'active'"`
	Tier        CustomerTier   `json:"tier" gorm:"default:'bronze'"`
	
	// Business Information (for B2B customers)
	IsBusinessCustomer bool   `json:"is_business_customer" gorm:"default:false"`
	CompanyName       string  `json:"company_name" gorm:"size:200"`
	TaxNumber         string  `json:"tax_number" gorm:"size:50"`
	TaxOffice         string  `json:"tax_office" gorm:"size:100"`
	
	// Tracking
	LastLoginAt   *time.Time `json:"last_login_at"`
	LastOrderAt   *time.Time `json:"last_order_at"`
	TotalOrders   int        `json:"total_orders" gorm:"default:0"`
	TotalSpent    float64    `json:"total_spent" gorm:"type:decimal(15,2);default:0"`
	
	// Relationships
	Addresses     []Address  `json:"addresses" gorm:"foreignKey:CustomerID"`
	Orders        []Order    `json:"orders" gorm:"foreignKey:CustomerID"`
	Reviews       []Review   `json:"reviews" gorm:"foreignKey:CustomerID"`
	
	// Metadata
	Metadata      CustomerMetadata `json:"metadata" gorm:"type:json"`
	
	// Timestamps
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// CustomerStatus represents customer account status
type CustomerStatus string

const (
	CustomerStatusActive    CustomerStatus = "active"
	CustomerStatusInactive  CustomerStatus = "inactive"
	CustomerStatusSuspended CustomerStatus = "suspended"
	CustomerStatusBlocked   CustomerStatus = "blocked"
)

// CustomerTier represents customer loyalty tier
type CustomerTier string

const (
	CustomerTierBronze   CustomerTier = "bronze"
	CustomerTierSilver   CustomerTier = "silver"
	CustomerTierGold     CustomerTier = "gold"
	CustomerTierPlatinum CustomerTier = "platinum"
	CustomerTierVIP      CustomerTier = "vip"
)

// CustomerMetadata holds additional customer data
type CustomerMetadata struct {
	Source           string                 `json:"source,omitempty"`           // registration source
	ReferralCode     string                 `json:"referral_code,omitempty"`    // how they were referred
	PreferredBrands  []string               `json:"preferred_brands,omitempty"`  // favorite brands
	Interests        []string               `json:"interests,omitempty"`         // customer interests
	CustomFields     map[string]interface{} `json:"custom_fields,omitempty"`     // extensible fields
}

// Address represents a customer address
type Address struct {
	ID          uint        `json:"id" gorm:"primaryKey"`
	CustomerID  uint        `json:"customer_id" gorm:"index;not null"`
	Customer    Customer    `json:"customer" gorm:"foreignKey:CustomerID"`
	
	// Address Details
	Title       string      `json:"title" gorm:"size:100;not null" validate:"required,min=2,max=100"`
	FirstName   string      `json:"first_name" gorm:"size:100;not null" validate:"required,min=2,max=100"`
	LastName    string      `json:"last_name" gorm:"size:100;not null" validate:"required,min=2,max=100"`
	CompanyName string      `json:"company_name" gorm:"size:200"`
	
	// Location
	AddressLine1 string     `json:"address_line1" gorm:"size:255;not null" validate:"required,max=255"`
	AddressLine2 string     `json:"address_line2" gorm:"size:255"`
	City         string     `json:"city" gorm:"size:100;not null" validate:"required,max=100"`
	State        string     `json:"state" gorm:"size:100"`
	PostalCode   string     `json:"postal_code" gorm:"size:20;not null" validate:"required,max=20"`
	Country      string     `json:"country" gorm:"size:2;not null;default:'TR'" validate:"required,len=2"`
	
	// Contact
	Phone        string     `json:"phone" gorm:"size:20" validate:"omitempty,e164"`
	
	// Address Type and Preferences
	Type         AddressType `json:"type" gorm:"default:'shipping'"`
	IsDefault    bool        `json:"is_default" gorm:"default:false"`
	IsBilling    bool        `json:"is_billing" gorm:"default:false"`
	IsShipping   bool        `json:"is_shipping" gorm:"default:true"`
	
	// Geolocation (for delivery optimization)
	Latitude     *float64   `json:"latitude"`
	Longitude    *float64   `json:"longitude"`
	
	// Delivery Instructions
	DeliveryInstructions string `json:"delivery_instructions" gorm:"type:text"`
	
	// Timestamps
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// AddressType represents the type of address
type AddressType string

const (
	AddressTypeShipping AddressType = "shipping"
	AddressTypeBilling  AddressType = "billing"
	AddressTypeBoth     AddressType = "both"
)

// Implement driver.Valuer interface for CustomerMetadata
func (cm CustomerMetadata) Value() (driver.Value, error) {
	return json.Marshal(cm)
}

// Implement sql.Scanner interface for CustomerMetadata
func (cm *CustomerMetadata) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	
	return json.Unmarshal(bytes, cm)
}

// TableName returns the table name for Customer
func (Customer) TableName() string {
	return "customers"
}

// TableName returns the table name for Address
func (Address) TableName() string {
	return "addresses"
}

// GetFullName returns the customer's full name
func (c *Customer) GetFullName() string {
	return c.FirstName + " " + c.LastName
}

// GetDefaultAddress returns the default address for the customer
func (c *Customer) GetDefaultAddress() *Address {
	for _, addr := range c.Addresses {
		if addr.IsDefault {
			return &addr
		}
	}
	return nil
}

// GetBillingAddress returns the billing address for the customer
func (c *Customer) GetBillingAddress() *Address {
	for _, addr := range c.Addresses {
		if addr.IsBilling {
			return &addr
		}
	}
	return c.GetDefaultAddress()
}

// GetShippingAddress returns the shipping address for the customer
func (c *Customer) GetShippingAddress() *Address {
	for _, addr := range c.Addresses {
		if addr.IsShipping && addr.IsDefault {
			return &addr
		}
	}
	
	for _, addr := range c.Addresses {
		if addr.IsShipping {
			return &addr
		}
	}
	return nil
}

// IsVIP checks if customer is VIP tier
func (c *Customer) IsVIP() bool {
	return c.Tier == CustomerTierVIP || c.Tier == CustomerTierPlatinum
}

// CanReceiveDiscount checks if customer is eligible for discounts
func (c *Customer) CanReceiveDiscount() bool {
	return c.Status == CustomerStatusActive && c.TotalOrders > 0
}

// GetFullAddress returns formatted full address
func (a *Address) GetFullAddress() string {
	address := a.AddressLine1
	if a.AddressLine2 != "" {
		address += ", " + a.AddressLine2
	}
	address += ", " + a.City
	if a.State != "" {
		address += ", " + a.State
	}
	address += " " + a.PostalCode
	return address
}

// GetContactName returns the contact person name for the address
func (a *Address) GetContactName() string {
	return a.FirstName + " " + a.LastName
}