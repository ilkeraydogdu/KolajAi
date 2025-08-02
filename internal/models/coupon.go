package models

import (
	"errors"
	"strings"
	"time"
)

// Coupon represents a discount coupon
type Coupon struct {
	ID                int64     `json:"id" db:"id"`
	Code              string    `json:"code" db:"code"`
	Name              string    `json:"name" db:"name"`
	Description       string    `json:"description" db:"description"`
	Type              string    `json:"type" db:"type"` // percentage, fixed_amount, free_shipping
	Value             float64   `json:"value" db:"value"`
	MinOrderAmount    float64   `json:"min_order_amount" db:"min_order_amount"`
	MaxDiscountAmount float64   `json:"max_discount_amount" db:"max_discount_amount"`
	UsageLimit        int       `json:"usage_limit" db:"usage_limit"`
	UsageCount        int       `json:"usage_count" db:"usage_count"`
	UserUsageLimit    int       `json:"user_usage_limit" db:"user_usage_limit"`
	ValidFrom         time.Time `json:"valid_from" db:"valid_from"`
	ValidUntil        time.Time `json:"valid_until" db:"valid_until"`
	IsActive          bool      `json:"is_active" db:"is_active"`
	ApplicableToAll   bool      `json:"applicable_to_all" db:"applicable_to_all"`
	FirstTimeOnly     bool      `json:"first_time_only" db:"first_time_only"`
	CreatedBy         int64     `json:"created_by" db:"created_by"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
	
	// Related data
	Categories        []int64   `json:"categories,omitempty"`
	Products          []int64   `json:"products,omitempty"`
	Users             []int64   `json:"users,omitempty"`
}

// CouponUsage represents coupon usage history
type CouponUsage struct {
	ID        int64     `json:"id" db:"id"`
	CouponID  int64     `json:"coupon_id" db:"coupon_id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	OrderID   int64     `json:"order_id" db:"order_id"`
	DiscountAmount float64 `json:"discount_amount" db:"discount_amount"`
	UsedAt    time.Time `json:"used_at" db:"used_at"`
	
	// Related data
	Coupon    *Coupon   `json:"coupon,omitempty"`
	User      *User     `json:"user,omitempty"`
	Order     *Order    `json:"order,omitempty"`
}

// Discount represents a general discount rule
type Discount struct {
	ID                int64     `json:"id" db:"id"`
	Name              string    `json:"name" db:"name"`
	Description       string    `json:"description" db:"description"`
	Type              string    `json:"type" db:"type"` // bulk_discount, category_discount, seasonal_discount
	DiscountType      string    `json:"discount_type" db:"discount_type"` // percentage, fixed_amount
	Value             float64   `json:"value" db:"value"`
	MinQuantity       int       `json:"min_quantity" db:"min_quantity"`
	MaxQuantity       int       `json:"max_quantity" db:"max_quantity"`
	MinOrderAmount    float64   `json:"min_order_amount" db:"min_order_amount"`
	MaxDiscountAmount float64   `json:"max_discount_amount" db:"max_discount_amount"`
	Priority          int       `json:"priority" db:"priority"`
	IsActive          bool      `json:"is_active" db:"is_active"`
	IsStackable       bool      `json:"is_stackable" db:"is_stackable"`
	ValidFrom         time.Time `json:"valid_from" db:"valid_from"`
	ValidUntil        time.Time `json:"valid_until" db:"valid_until"`
	CreatedBy         int64     `json:"created_by" db:"created_by"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
	
	// Related data
	Categories        []int64   `json:"categories,omitempty"`
	Products          []int64   `json:"products,omitempty"`
	CustomerGroups    []int64   `json:"customer_groups,omitempty"`
}

// CustomerGroup represents a customer group for targeted discounts
type CustomerGroup struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Criteria    string    `json:"criteria" db:"criteria"` // JSON string with criteria
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedBy   int64     `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	
	// Related data
	Customers   []int64   `json:"customers,omitempty"`
}

// Validate checks if the coupon data is valid
func (c *Coupon) Validate() error {
	if strings.TrimSpace(c.Code) == "" {
		return errors.New("coupon code cannot be empty")
	}
	
	if strings.TrimSpace(c.Name) == "" {
		return errors.New("coupon name cannot be empty")
	}
	
	if c.Type == "" {
		return errors.New("coupon type cannot be empty")
	}
	
	validTypes := []string{"percentage", "fixed_amount", "free_shipping"}
	isValidType := false
	for _, validType := range validTypes {
		if c.Type == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return errors.New("invalid coupon type")
	}
	
	if c.Value <= 0 {
		return errors.New("coupon value must be greater than zero")
	}
	
	if c.Type == "percentage" && c.Value > 100 {
		return errors.New("percentage discount cannot be greater than 100")
	}
	
	if c.ValidFrom.After(c.ValidUntil) {
		return errors.New("valid from date cannot be after valid until date")
	}
	
	if c.UsageLimit < 0 {
		return errors.New("usage limit cannot be negative")
	}
	
	if c.UserUsageLimit < 0 {
		return errors.New("user usage limit cannot be negative")
	}
	
	return nil
}

// IsValid checks if the coupon is currently valid
func (c *Coupon) IsValid() bool {
	now := time.Now()
	return c.IsActive && 
		   now.After(c.ValidFrom) && 
		   now.Before(c.ValidUntil) &&
		   (c.UsageLimit == 0 || c.UsageCount < c.UsageLimit)
}

// CanBeUsedBy checks if the coupon can be used by a specific user
func (c *Coupon) CanBeUsedBy(userID int64, userUsageCount int) bool {
	if !c.IsValid() {
		return false
	}
	
	if c.UserUsageLimit > 0 && userUsageCount >= c.UserUsageLimit {
		return false
	}
	
	return true
}

// CalculateDiscount calculates the discount amount for a given order amount
func (c *Coupon) CalculateDiscount(orderAmount float64) float64 {
	if orderAmount < c.MinOrderAmount {
		return 0
	}
	
	var discount float64
	
	switch c.Type {
	case "percentage":
		discount = orderAmount * (c.Value / 100)
	case "fixed_amount":
		discount = c.Value
	case "free_shipping":
		// Free shipping discount would be calculated based on shipping cost
		// For now, return 0 as shipping cost calculation is not implemented
		return 0
	default:
		return 0
	}
	
	// Apply maximum discount limit
	if c.MaxDiscountAmount > 0 && discount > c.MaxDiscountAmount {
		discount = c.MaxDiscountAmount
	}
	
	// Ensure discount doesn't exceed order amount
	if discount > orderAmount {
		discount = orderAmount
	}
	
	return discount
}

// Validate checks if the discount data is valid
func (d *Discount) Validate() error {
	if strings.TrimSpace(d.Name) == "" {
		return errors.New("discount name cannot be empty")
	}
	
	if d.Type == "" {
		return errors.New("discount type cannot be empty")
	}
	
	validTypes := []string{"bulk_discount", "category_discount", "seasonal_discount"}
	isValidType := false
	for _, validType := range validTypes {
		if d.Type == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return errors.New("invalid discount type")
	}
	
	if d.DiscountType == "" {
		return errors.New("discount type cannot be empty")
	}
	
	validDiscountTypes := []string{"percentage", "fixed_amount"}
	isValidDiscountType := false
	for _, validType := range validDiscountTypes {
		if d.DiscountType == validType {
			isValidDiscountType = true
			break
		}
	}
	if !isValidDiscountType {
		return errors.New("invalid discount type")
	}
	
	if d.Value <= 0 {
		return errors.New("discount value must be greater than zero")
	}
	
	if d.DiscountType == "percentage" && d.Value > 100 {
		return errors.New("percentage discount cannot be greater than 100")
	}
	
	if d.ValidFrom.After(d.ValidUntil) {
		return errors.New("valid from date cannot be after valid until date")
	}
	
	if d.MinQuantity < 0 {
		return errors.New("minimum quantity cannot be negative")
	}
	
	if d.MaxQuantity > 0 && d.MaxQuantity < d.MinQuantity {
		return errors.New("maximum quantity cannot be less than minimum quantity")
	}
	
	return nil
}

// IsValid checks if the discount is currently valid
func (d *Discount) IsValid() bool {
	now := time.Now()
	return d.IsActive && 
		   now.After(d.ValidFrom) && 
		   now.Before(d.ValidUntil)
}

// CalculateDiscount calculates the discount amount
func (d *Discount) CalculateDiscount(orderAmount float64, quantity int) float64 {
	if !d.IsValid() {
		return 0
	}
	
	if orderAmount < d.MinOrderAmount {
		return 0
	}
	
	if quantity < d.MinQuantity {
		return 0
	}
	
	if d.MaxQuantity > 0 && quantity > d.MaxQuantity {
		return 0
	}
	
	var discount float64
	
	switch d.DiscountType {
	case "percentage":
		discount = orderAmount * (d.Value / 100)
	case "fixed_amount":
		discount = d.Value
	default:
		return 0
	}
	
	// Apply maximum discount limit
	if d.MaxDiscountAmount > 0 && discount > d.MaxDiscountAmount {
		discount = d.MaxDiscountAmount
	}
	
	// Ensure discount doesn't exceed order amount
	if discount > orderAmount {
		discount = orderAmount
	}
	
	return discount
}

// Validate checks if the customer group data is valid
func (cg *CustomerGroup) Validate() error {
	if strings.TrimSpace(cg.Name) == "" {
		return errors.New("customer group name cannot be empty")
	}
	
	if cg.CreatedBy <= 0 {
		return errors.New("valid creator ID is required")
	}
	
	return nil
}