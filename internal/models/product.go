package models

import (
	"errors"
	"strings"
	"time"
)

// Note: Category struct moved to category.go model

// Product represents a product in the marketplace
type Product struct {
	ID              int       `json:"id" db:"id"`
	VendorID        int       `json:"vendor_id" db:"vendor_id"`
	CategoryID      int       `json:"category_id" db:"category_id"` // Fixed: uint -> int for consistency
	Name            string    `json:"name" db:"name" validate:"required,min=1,max=255"`
	Description     string    `json:"description" db:"description" validate:"max=5000"`
	ShortDesc       string    `json:"short_desc" db:"short_desc" validate:"max=500"`
	SKU             string    `json:"sku" db:"sku" validate:"required,min=3,max=100"`
	Price           float64   `json:"price" db:"price" validate:"required,min=0"`
	ComparePrice    float64   `json:"compare_price" db:"compare_price" validate:"min=0"`
	CostPrice       float64   `json:"cost_price" db:"cost_price" validate:"min=0"`
	WholesalePrice  float64   `json:"wholesale_price" db:"wholesale_price" validate:"min=0"`
	MinWholesaleQty int       `json:"min_wholesale_qty" db:"min_wholesale_qty" validate:"min=1"`
	Stock           int       `json:"stock" db:"stock" validate:"min=0"`
	MinStock        int       `json:"min_stock" db:"min_stock" validate:"min=0"`
	Weight          float64   `json:"weight" db:"weight" validate:"min=0"`
	Dimensions      string    `json:"dimensions" db:"dimensions" validate:"max=100"`
	Status          string    `json:"status" db:"status" validate:"oneof=draft active inactive out_of_stock"`
	IsDigital       bool      `json:"is_digital" db:"is_digital"`
	IsFeatured      bool      `json:"is_featured" db:"is_featured"`
	AllowReviews    bool      `json:"allow_reviews" db:"allow_reviews"`
	MetaTitle       string    `json:"meta_title" db:"meta_title" validate:"max=255"`
	MetaDesc        string    `json:"meta_desc" db:"meta_desc" validate:"max=500"`
	Tags            string    `json:"tags" db:"tags" validate:"max=1000"`
	ViewCount       int       `json:"view_count" db:"view_count" validate:"min=0"`
	SalesCount      int       `json:"sales_count" db:"sales_count" validate:"min=0"`
	Rating          float64   `json:"rating" db:"rating" validate:"min=0,max=5"`
	ReviewCount     int       `json:"review_count" db:"review_count" validate:"min=0"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	
	// Computed fields for templates (not stored in DB)
	DiscountPrice      float64 `json:"discount_price,omitempty" db:"-"`
	DiscountPercentage int     `json:"discount_percentage,omitempty" db:"-"`
}

// Validate checks if the product data is valid
func (p *Product) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return errors.New("product name cannot be empty")
	}

	if len(p.Name) > 255 {
		return errors.New("product name cannot exceed 255 characters")
	}

	if len(p.Description) > 5000 {
		return errors.New("product description cannot exceed 5000 characters")
	}

	if len(p.ShortDesc) > 500 {
		return errors.New("product short description cannot exceed 500 characters")
	}

	if strings.TrimSpace(p.SKU) == "" {
		return errors.New("product SKU cannot be empty")
	}

	if len(p.SKU) < 3 || len(p.SKU) > 100 {
		return errors.New("product SKU must be between 3 and 100 characters")
	}

	if p.Price < 0 {
		return errors.New("product price cannot be negative")
	}

	if p.ComparePrice < 0 {
		return errors.New("product compare price cannot be negative")
	}

	if p.CostPrice < 0 {
		return errors.New("product cost price cannot be negative")
	}

	if p.WholesalePrice < 0 {
		return errors.New("product wholesale price cannot be negative")
	}

	if p.Stock < 0 {
		return errors.New("product stock cannot be negative")
	}

	if p.MinStock < 0 {
		return errors.New("product minimum stock cannot be negative")
	}

	if p.Weight < 0 {
		return errors.New("product weight cannot be negative")
	}

	if p.Rating < 0 || p.Rating > 5 {
		return errors.New("product rating must be between 0 and 5")
	}

	if p.ReviewCount < 0 {
		return errors.New("product review count cannot be negative")
	}

	if p.ViewCount < 0 {
		return errors.New("product view count cannot be negative")
	}

	if p.SalesCount < 0 {
		return errors.New("product sales count cannot be negative")
	}

	if p.VendorID <= 0 {
		return errors.New("valid vendor ID is required")
	}

	if p.CategoryID <= 0 {
		return errors.New("valid category ID is required")
	}

	// Validate status
	validStatuses := []string{"draft", "active", "inactive", "out_of_stock"}
	statusValid := false
	for _, status := range validStatuses {
		if p.Status == status {
			statusValid = true
			break
		}
	}
	if !statusValid {
		return errors.New("invalid product status")
	}

	return nil
}

// IsAvailable checks if the product is available for purchase
func (p *Product) IsAvailable() bool {
	return p.Status == "active" && p.Stock > 0
}

// CalculateDiscountPrice calculates discount price if compare price is set
func (p *Product) CalculateDiscountPrice() {
	if p.ComparePrice > p.Price && p.ComparePrice > 0 {
		p.DiscountPrice = p.Price
		p.DiscountPercentage = int(((p.ComparePrice - p.Price) / p.ComparePrice) * 100)
	}
}

// HasDiscount checks if product has discount
func (p *Product) HasDiscount() bool {
	return p.ComparePrice > p.Price && p.ComparePrice > 0
}

// ProductImage represents product images
type ProductImage struct {
	ID        int       `json:"id" db:"id"`
	ProductID int       `json:"product_id" db:"product_id"`
	ImageURL  string    `json:"image_url" db:"image_url"`
	AltText   string    `json:"alt_text" db:"alt_text"`
	SortOrder int       `json:"sort_order" db:"sort_order"`
	IsPrimary bool      `json:"is_primary" db:"is_primary"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ProductVariant represents product variants (size, color, etc.)
type ProductVariant struct {
	ID        int     `json:"id" db:"id"`
	ProductID int     `json:"product_id" db:"product_id"`
	Name      string  `json:"name" db:"name"`
	Value     string  `json:"value" db:"value"`
	Price     float64 `json:"price" db:"price"`
	Stock     int     `json:"stock" db:"stock"`
	SKU       string  `json:"sku" db:"sku"`
	IsActive  bool    `json:"is_active" db:"is_active"`
}

// ProductAttribute represents product attributes
type ProductAttribute struct {
	ID        int    `json:"id" db:"id"`
	ProductID int    `json:"product_id" db:"product_id"`
	Name      string `json:"name" db:"name"`
	Value     string `json:"value" db:"value"`
}

// ProductReview represents product reviews
type ProductReview struct {
	ID         int       `json:"id" db:"id"`
	ProductID  int       `json:"product_id" db:"product_id"`
	UserID     int       `json:"user_id" db:"user_id"`
	OrderID    int       `json:"order_id" db:"order_id"`
	Rating     int       `json:"rating" db:"rating"`
	Title      string    `json:"title" db:"title"`
	Comment    string    `json:"comment" db:"comment"`
	Images     string    `json:"images" db:"images"`
	IsVerified bool      `json:"is_verified" db:"is_verified"`
	Status     string    `json:"status" db:"status"` // pending, approved, rejected
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}
