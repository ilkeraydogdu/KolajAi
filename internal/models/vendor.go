package models

import "time"

// Vendor represents a seller in the marketplace
type Vendor struct {
	ID          int       `json:"id" db:"id"`
	UserID      int       `json:"user_id" db:"user_id"`
	CompanyName string    `json:"company_name" db:"company_name"`
	BusinessID  string    `json:"business_id" db:"business_id"`
	Description string    `json:"description" db:"description"`
	Logo        string    `json:"logo" db:"logo"`
	Phone       string    `json:"phone" db:"phone"`
	Address     string    `json:"address" db:"address"`
	City        string    `json:"city" db:"city"`
	Country     string    `json:"country" db:"country"`
	Website     string    `json:"website" db:"website"`
	Status      string    `json:"status" db:"status"` // pending, approved, suspended, rejected
	Rating      float64   `json:"rating" db:"rating"`
	TotalSales  float64   `json:"total_sales" db:"total_sales"`
	Commission  float64   `json:"commission" db:"commission"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// VendorDocument represents vendor verification documents
type VendorDocument struct {
	ID         int       `json:"id" db:"id"`
	VendorID   int       `json:"vendor_id" db:"vendor_id"`
	Type       string    `json:"type" db:"type"` // business_license, tax_certificate, etc.
	FileName   string    `json:"file_name" db:"file_name"`
	FilePath   string    `json:"file_path" db:"file_path"`
	Status     string    `json:"status" db:"status"` // pending, approved, rejected
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
}
