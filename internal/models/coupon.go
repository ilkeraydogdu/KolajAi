package models

import (
	"fmt"
	"time"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
)

// Coupon represents a discount coupon
type Coupon struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	
	// Basic Information
	Code        string    `json:"code" gorm:"size:50;unique;not null" validate:"required,min=3,max=50"`
	Name        string    `json:"name" gorm:"size:200;not null" validate:"required,min=5,max=200"`
	Description string    `json:"description" gorm:"type:text"`
	
	// Coupon Type and Value
	Type        CouponType `json:"type" gorm:"not null"`
	Value       float64    `json:"value" gorm:"type:decimal(15,2);not null" validate:"required,gt=0"`
	Currency    string     `json:"currency" gorm:"size:3;default:'TRY'"`
	MaxDiscount *float64   `json:"max_discount" gorm:"type:decimal(15,2)"` // for percentage coupons
	
	// Usage Limits
	UsageLimit      *int  `json:"usage_limit"`              // total usage limit
	UsageLimitPerUser *int `json:"usage_limit_per_user"`    // per user limit
	UsedCount       int   `json:"used_count" gorm:"default:0"`
	
	// Minimum Requirements
	MinOrderAmount  *float64 `json:"min_order_amount" gorm:"type:decimal(15,2)"`
	MinItemCount    *int     `json:"min_item_count"`
	
	// Validity Period
	ValidFrom   time.Time  `json:"valid_from" validate:"required"`
	ValidUntil  *time.Time `json:"valid_until"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	
	// Target Restrictions
	ApplicableProducts   ProductRestrictions `json:"applicable_products" gorm:"type:json"`
	ApplicableCategories CategoryRestrictions `json:"applicable_categories" gorm:"type:json"`
	ApplicableVendors    VendorRestrictions  `json:"applicable_vendors" gorm:"type:json"`
	ApplicableUsers      UserRestrictions    `json:"applicable_users" gorm:"type:json"`
	
	// Geographic Restrictions
	ApplicableCountries []string `json:"applicable_countries" gorm:"type:json"`
	ExcludedCountries   []string `json:"excluded_countries" gorm:"type:json"`
	
	// Combination Rules
	CanCombineWithOthers bool     `json:"can_combine_with_others" gorm:"default:false"`
	ExcludedCoupons      []string `json:"excluded_coupons" gorm:"type:json"`
	
	// Priority and Stacking
	Priority    int  `json:"priority" gorm:"default:0"`
	IsStackable bool `json:"is_stackable" gorm:"default:false"`
	
	// Marketing
	IsPublic        bool   `json:"is_public" gorm:"default:true"`
	IsPromotional   bool   `json:"is_promotional" gorm:"default:false"`
	PromotionalText string `json:"promotional_text" gorm:"size:500"`
	
	// Vendor/Admin
	VendorID  *uint `json:"vendor_id" gorm:"index"`
	Vendor    *Vendor `json:"vendor,omitempty" gorm:"foreignKey:VendorID"`
	CreatedBy uint  `json:"created_by" gorm:"index;not null"`
	
	// Analytics
	ViewCount int `json:"view_count" gorm:"default:0"`
	
	// Metadata
	Metadata  CouponMetadata `json:"metadata" gorm:"type:json"`
	
	// Relationships
	CouponUsages []CouponUsage `json:"coupon_usages,omitempty" gorm:"foreignKey:CouponID"`
	
	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// CouponType represents the type of coupon
type CouponType string

const (
	CouponTypeFixedAmount      CouponType = "fixed_amount"      // Fixed amount discount
	CouponTypePercentage       CouponType = "percentage"        // Percentage discount
	CouponTypeFreeShipping     CouponType = "free_shipping"     // Free shipping
	CouponTypeBuyXGetY         CouponType = "buy_x_get_y"       // Buy X get Y free
	CouponTypeFirstOrder       CouponType = "first_order"       // First order discount
	CouponTypeReferral         CouponType = "referral"          // Referral discount
	CouponTypeLoyalty          CouponType = "loyalty"           // Loyalty points discount
	CouponTypeVolumeDiscount   CouponType = "volume_discount"   // Volume-based discount
	CouponTypeTimeLimit        CouponType = "time_limit"        // Flash sale discount
)

// ProductRestrictions defines product-based restrictions
type ProductRestrictions struct {
	IncludedProducts []uint   `json:"included_products,omitempty"`
	ExcludedProducts []uint   `json:"excluded_products,omitempty"`
	IncludedBrands   []string `json:"included_brands,omitempty"`
	ExcludedBrands   []string `json:"excluded_brands,omitempty"`
	IncludedTags     []string `json:"included_tags,omitempty"`
	ExcludedTags     []string `json:"excluded_tags,omitempty"`
}

// CategoryRestrictions defines category-based restrictions
type CategoryRestrictions struct {
	IncludedCategories []uint `json:"included_categories,omitempty"`
	ExcludedCategories []uint `json:"excluded_categories,omitempty"`
}

// VendorRestrictions defines vendor-based restrictions
type VendorRestrictions struct {
	IncludedVendors []uint `json:"included_vendors,omitempty"`
	ExcludedVendors []uint `json:"excluded_vendors,omitempty"`
}

// UserRestrictions defines user-based restrictions
type UserRestrictions struct {
	IncludedUsers    []uint         `json:"included_users,omitempty"`
	ExcludedUsers    []uint         `json:"excluded_users,omitempty"`
	CustomerTiers    []CustomerTier `json:"customer_tiers,omitempty"`
	NewCustomersOnly bool           `json:"new_customers_only,omitempty"`
	MinRegistrationDays int         `json:"min_registration_days,omitempty"`
}

// CouponMetadata holds additional coupon data
type CouponMetadata struct {
	Campaign        string                 `json:"campaign,omitempty"`
	Source          string                 `json:"source,omitempty"`
	Medium          string                 `json:"medium,omitempty"`
	UtmParams       map[string]string      `json:"utm_params,omitempty"`
	CustomFields    map[string]interface{} `json:"custom_fields,omitempty"`
	InternalNotes   string                 `json:"internal_notes,omitempty"`
}

// CouponUsage represents a coupon usage record
type CouponUsage struct {
	ID         uint     `json:"id" gorm:"primaryKey"`
	CouponID   uint     `json:"coupon_id" gorm:"index;not null"`
	Coupon     Coupon   `json:"coupon" gorm:"foreignKey:CouponID"`
	CustomerID uint     `json:"customer_id" gorm:"index;not null"`
	Customer   Customer `json:"customer" gorm:"foreignKey:CustomerID"`
	OrderID    uint     `json:"order_id" gorm:"index;not null"`
	Order      Order    `json:"order" gorm:"foreignKey:OrderID"`
	
	// Usage Details
	DiscountAmount float64 `json:"discount_amount" gorm:"type:decimal(15,2);not null"`
	Currency       string  `json:"currency" gorm:"size:3;not null"`
	OrderAmount    float64 `json:"order_amount" gorm:"type:decimal(15,2);not null"`
	
	// Context
	IPAddress  string `json:"ip_address" gorm:"size:45"`
	UserAgent  string `json:"user_agent" gorm:"size:500"`
	
	// Status
	Status     CouponUsageStatus `json:"status" gorm:"default:'used'"`
	RefundedAt *time.Time        `json:"refunded_at"`
	
	// Timestamps
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CouponUsageStatus represents coupon usage status
type CouponUsageStatus string

const (
	CouponUsageStatusUsed     CouponUsageStatus = "used"
	CouponUsageStatusRefunded CouponUsageStatus = "refunded"
	CouponUsageStatusCancelled CouponUsageStatus = "cancelled"
)

// Discount represents automatic discounts and promotions
type Discount struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	
	// Basic Information
	Name        string    `json:"name" gorm:"size:200;not null" validate:"required,min=5,max=200"`
	Description string    `json:"description" gorm:"type:text"`
	
	// Discount Type and Value
	Type        DiscountType `json:"type" gorm:"not null"`
	Value       float64      `json:"value" gorm:"type:decimal(15,2);not null" validate:"required,gt=0"`
	Currency    string       `json:"currency" gorm:"size:3;default:'TRY'"`
	MaxDiscount *float64     `json:"max_discount" gorm:"type:decimal(15,2)"`
	
	// Trigger Conditions
	TriggerType      DiscountTrigger `json:"trigger_type" gorm:"not null"`
	TriggerValue     float64         `json:"trigger_value" gorm:"type:decimal(15,2)"`
	TriggerCondition string          `json:"trigger_condition" gorm:"size:500"`
	
	// Target Restrictions (same as coupon)
	ApplicableProducts   ProductRestrictions  `json:"applicable_products" gorm:"type:json"`
	ApplicableCategories CategoryRestrictions `json:"applicable_categories" gorm:"type:json"`
	ApplicableVendors    VendorRestrictions   `json:"applicable_vendors" gorm:"type:json"`
	ApplicableUsers      UserRestrictions     `json:"applicable_users" gorm:"type:json"`
	
	// Validity Period
	ValidFrom   time.Time  `json:"valid_from" validate:"required"`
	ValidUntil  *time.Time `json:"valid_until"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	
	// Usage Limits
	UsageLimit *int `json:"usage_limit"`
	UsedCount  int  `json:"used_count" gorm:"default:0"`
	
	// Priority and Combination
	Priority            int  `json:"priority" gorm:"default:0"`
	CanCombineWithCoupons bool `json:"can_combine_with_coupons" gorm:"default:true"`
	
	// Vendor/Admin
	VendorID  *uint `json:"vendor_id" gorm:"index"`
	Vendor    *Vendor `json:"vendor,omitempty" gorm:"foreignKey:VendorID"`
	CreatedBy uint  `json:"created_by" gorm:"index;not null"`
	
	// Metadata
	Metadata DiscountMetadata `json:"metadata" gorm:"type:json"`
	
	// Relationships
	DiscountUsages []DiscountUsage `json:"discount_usages,omitempty" gorm:"foreignKey:DiscountID"`
	
	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// DiscountType represents the type of discount
type DiscountType string

const (
	DiscountTypeFixedAmount    DiscountType = "fixed_amount"
	DiscountTypePercentage     DiscountType = "percentage"
	DiscountTypeFreeShipping   DiscountType = "free_shipping"
	DiscountTypeBuyXGetY       DiscountType = "buy_x_get_y"
	DiscountTypeVolumeDiscount DiscountType = "volume_discount"
	DiscountTypeBundleDiscount DiscountType = "bundle_discount"
)

// DiscountTrigger represents what triggers the discount
type DiscountTrigger string

const (
	DiscountTriggerOrderAmount     DiscountTrigger = "order_amount"
	DiscountTriggerItemQuantity    DiscountTrigger = "item_quantity"
	DiscountTriggerFirstOrder      DiscountTrigger = "first_order"
	DiscountTriggerCustomerTier    DiscountTrigger = "customer_tier"
	DiscountTriggerProductCombo    DiscountTrigger = "product_combo"
	DiscountTriggerTimeBasedFlash  DiscountTrigger = "time_based_flash"
	DiscountTriggerInventoryLevel  DiscountTrigger = "inventory_level"
)

// DiscountMetadata holds additional discount data
type DiscountMetadata struct {
	Campaign      string                 `json:"campaign,omitempty"`
	AutoApply     bool                   `json:"auto_apply,omitempty"`
	DisplayBadge  bool                   `json:"display_badge,omitempty"`
	BadgeText     string                 `json:"badge_text,omitempty"`
	CustomFields  map[string]interface{} `json:"custom_fields,omitempty"`
	InternalNotes string                 `json:"internal_notes,omitempty"`
}

// DiscountUsage represents a discount usage record
type DiscountUsage struct {
	ID         uint     `json:"id" gorm:"primaryKey"`
	DiscountID uint     `json:"discount_id" gorm:"index;not null"`
	Discount   Discount `json:"discount" gorm:"foreignKey:DiscountID"`
	CustomerID uint     `json:"customer_id" gorm:"index;not null"`
	Customer   Customer `json:"customer" gorm:"foreignKey:CustomerID"`
	OrderID    uint     `json:"order_id" gorm:"index;not null"`
	Order      Order    `json:"order" gorm:"foreignKey:OrderID"`
	
	// Usage Details
	DiscountAmount float64 `json:"discount_amount" gorm:"type:decimal(15,2);not null"`
	Currency       string  `json:"currency" gorm:"size:3;not null"`
	OrderAmount    float64 `json:"order_amount" gorm:"type:decimal(15,2);not null"`
	
	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Implement driver.Valuer interfaces
func (pr ProductRestrictions) Value() (driver.Value, error) {
	return json.Marshal(pr)
}

func (pr *ProductRestrictions) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, pr)
}

func (cr CategoryRestrictions) Value() (driver.Value, error) {
	return json.Marshal(cr)
}

func (cr *CategoryRestrictions) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, cr)
}

func (vr VendorRestrictions) Value() (driver.Value, error) {
	return json.Marshal(vr)
}

func (vr *VendorRestrictions) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, vr)
}

func (ur UserRestrictions) Value() (driver.Value, error) {
	return json.Marshal(ur)
}

func (ur *UserRestrictions) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, ur)
}

func (cm CouponMetadata) Value() (driver.Value, error) {
	return json.Marshal(cm)
}

func (cm *CouponMetadata) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, cm)
}

func (dm DiscountMetadata) Value() (driver.Value, error) {
	return json.Marshal(dm)
}

func (dm *DiscountMetadata) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, dm)
}

// TableName methods
func (Coupon) TableName() string {
	return "coupons"
}

func (CouponUsage) TableName() string {
	return "coupon_usages"
}

func (Discount) TableName() string {
	return "discounts"
}

func (DiscountUsage) TableName() string {
	return "discount_usages"
}

// Coupon methods
func (c *Coupon) IsValid() bool {
	now := time.Now()
	return c.IsActive && 
		   c.ValidFrom.Before(now) && 
		   (c.ValidUntil == nil || c.ValidUntil.After(now)) &&
		   (c.UsageLimit == nil || c.UsedCount < *c.UsageLimit)
}

func (c *Coupon) IsExpired() bool {
	return c.ValidUntil != nil && time.Now().After(*c.ValidUntil)
}

func (c *Coupon) IsUsageLimitReached() bool {
	return c.UsageLimit != nil && c.UsedCount >= *c.UsageLimit
}

func (c *Coupon) CanBeUsedByCustomer(customerID uint, currentUsageCount int) bool {
	if c.UsageLimitPerUser == nil {
		return true
	}
	return currentUsageCount < *c.UsageLimitPerUser
}

func (c *Coupon) CalculateDiscount(orderAmount float64) float64 {
	switch c.Type {
	case CouponTypeFixedAmount:
		if orderAmount < c.Value {
			return orderAmount
		}
		return c.Value
		
	case CouponTypePercentage:
		discount := orderAmount * (c.Value / 100)
		if c.MaxDiscount != nil && discount > *c.MaxDiscount {
			return *c.MaxDiscount
		}
		return discount
		
	case CouponTypeFreeShipping:
		// This would be handled in shipping calculation
		return 0
		
	default:
		return 0
	}
}

func (c *Coupon) GetDisplayValue() string {
	switch c.Type {
	case CouponTypeFixedAmount:
		return fmt.Sprintf("%.2f %s", c.Value, c.Currency)
	case CouponTypePercentage:
		return fmt.Sprintf("%%%.0f", c.Value)
	case CouponTypeFreeShipping:
		return "Free Shipping"
	default:
		return "Discount"
	}
}

func (c *Coupon) NormalizeCode() {
	c.Code = strings.ToUpper(strings.TrimSpace(c.Code))
}

// Discount methods
func (d *Discount) IsValid() bool {
	now := time.Now()
	return d.IsActive && 
		   d.ValidFrom.Before(now) && 
		   (d.ValidUntil == nil || d.ValidUntil.After(now)) &&
		   (d.UsageLimit == nil || d.UsedCount < *d.UsageLimit)
}

func (d *Discount) CalculateDiscount(orderAmount float64) float64 {
	switch d.Type {
	case DiscountTypeFixedAmount:
		if orderAmount < d.Value {
			return orderAmount
		}
		return d.Value
		
	case DiscountTypePercentage:
		discount := orderAmount * (d.Value / 100)
		if d.MaxDiscount != nil && discount > *d.MaxDiscount {
			return *d.MaxDiscount
		}
		return discount
		
	default:
		return 0
	}
}

func (d *Discount) ShouldTrigger(orderAmount float64, itemCount int, customer *Customer) bool {
	switch d.TriggerType {
	case DiscountTriggerOrderAmount:
		return orderAmount >= d.TriggerValue
		
	case DiscountTriggerItemQuantity:
		return float64(itemCount) >= d.TriggerValue
		
	case DiscountTriggerFirstOrder:
		return customer != nil && customer.TotalOrders == 0
		
	case DiscountTriggerCustomerTier:
		// This would need additional logic based on customer tier
		return true
		
	default:
		return false
	}
}