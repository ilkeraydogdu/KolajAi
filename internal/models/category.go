package models

import (
	"time"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Category represents a product category
type Category struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	
	// Basic Information
	Name        string    `json:"name" gorm:"size:200;not null" validate:"required,min=2,max=200"`
	Slug        string    `json:"slug" gorm:"size:250;unique;not null" validate:"required"`
	Description string    `json:"description" gorm:"type:text"`
	
	// Hierarchy
	ParentID    *uint     `json:"parent_id" gorm:"index"`
	Parent      *Category `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children    []Category `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Level       int       `json:"level" gorm:"default:0"`
	Path        string    `json:"path" gorm:"size:500"` // e.g., "1/2/3" for breadcrumb
	
	// Display
	DisplayName string    `json:"display_name" gorm:"size:200"`
	ShortName   string    `json:"short_name" gorm:"size:100"`
	Icon        string    `json:"icon" gorm:"size:500"`
	Image       string    `json:"image" gorm:"size:500"`
	Color       string    `json:"color" gorm:"size:7"` // hex color
	
	// Status and Visibility
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	IsVisible   bool      `json:"is_visible" gorm:"default:true"`
	IsFeatured  bool      `json:"is_featured" gorm:"default:false"`
	
	// Ordering and Priority
	SortOrder   int       `json:"sort_order" gorm:"default:0"`
	Priority    int       `json:"priority" gorm:"default:0"`
	
	// SEO
	MetaTitle       string `json:"meta_title" gorm:"size:200"`
	MetaDescription string `json:"meta_description" gorm:"size:500"`
	MetaKeywords    string `json:"meta_keywords" gorm:"size:500"`
	
	// Product Count (cached)
	ProductCount int `json:"product_count" gorm:"default:0"`
	
	// Commission and Vendor
	Commission   float64 `json:"commission" gorm:"type:decimal(5,4);default:0"` // percentage
	VendorID     *uint   `json:"vendor_id" gorm:"index"`
	Vendor       *Vendor `json:"vendor,omitempty" gorm:"foreignKey:VendorID"`
	
	// Category Attributes and Filters
	Attributes   CategoryAttributes `json:"attributes" gorm:"type:json"`
	Filters      CategoryFilters    `json:"filters" gorm:"type:json"`
	
	// Marketplace Integration
	MarketplaceCategories []MarketplaceCategory `json:"marketplace_categories,omitempty" gorm:"foreignKey:CategoryID"`
	
	// Relationships
	Products     []Product           `json:"products,omitempty" gorm:"many2many:product_categories;"`
	CategoryTags []CategoryTag       `json:"category_tags,omitempty" gorm:"foreignKey:CategoryID"`
	
	// Analytics
	ViewCount    int       `json:"view_count" gorm:"default:0"`
	ClickCount   int       `json:"click_count" gorm:"default:0"`
	
	// Timestamps
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// CategoryAttributes holds category-specific attributes
type CategoryAttributes struct {
	RequiredAttributes []CategoryAttribute `json:"required_attributes,omitempty"`
	OptionalAttributes []CategoryAttribute `json:"optional_attributes,omitempty"`
	CustomFields       []CategoryField     `json:"custom_fields,omitempty"`
}

// CategoryAttribute represents a product attribute for this category
type CategoryAttribute struct {
	Name        string                 `json:"name"`
	Type        AttributeType          `json:"type"`
	Required    bool                   `json:"required"`
	Options     []string               `json:"options,omitempty"`
	DefaultValue string                `json:"default_value,omitempty"`
	Validation  map[string]interface{} `json:"validation,omitempty"`
}

// AttributeType represents attribute data types
type AttributeType string

const (
	AttributeTypeText     AttributeType = "text"
	AttributeTypeNumber   AttributeType = "number"
	AttributeTypeBoolean  AttributeType = "boolean"
	AttributeTypeSelect   AttributeType = "select"
	AttributeTypeMultiSelect AttributeType = "multi_select"
	AttributeTypeDate     AttributeType = "date"
	AttributeTypeColor    AttributeType = "color"
	AttributeTypeFile     AttributeType = "file"
	AttributeTypeURL      AttributeType = "url"
)

// CategoryField represents custom fields for categories
type CategoryField struct {
	Name        string      `json:"name"`
	Label       string      `json:"label"`
	Type        FieldType   `json:"type"`
	Required    bool        `json:"required"`
	Options     []string    `json:"options,omitempty"`
	Placeholder string      `json:"placeholder,omitempty"`
	HelpText    string      `json:"help_text,omitempty"`
}

// FieldType represents field types
type FieldType string

const (
	FieldTypeText     FieldType = "text"
	FieldTypeTextarea FieldType = "textarea"
	FieldTypeNumber   FieldType = "number"
	FieldTypeEmail    FieldType = "email"
	FieldTypeURL      FieldType = "url"
	FieldTypeSelect   FieldType = "select"
	FieldTypeCheckbox FieldType = "checkbox"
	FieldTypeRadio    FieldType = "radio"
	FieldTypeDate     FieldType = "date"
	FieldTypeFile     FieldType = "file"
)

// CategoryFilters holds category-specific filters
type CategoryFilters struct {
	PriceRanges    []PriceRange    `json:"price_ranges,omitempty"`
	BrandFilters   []BrandFilter   `json:"brand_filters,omitempty"`
	AttributeFilters []AttributeFilter `json:"attribute_filters,omitempty"`
	CustomFilters  []CustomFilter  `json:"custom_filters,omitempty"`
}

// PriceRange represents price filter ranges
type PriceRange struct {
	Label    string  `json:"label"`
	MinPrice float64 `json:"min_price"`
	MaxPrice float64 `json:"max_price"`
}

// BrandFilter represents brand filtering options
type BrandFilter struct {
	BrandName string `json:"brand_name"`
	Enabled   bool   `json:"enabled"`
}

// AttributeFilter represents attribute-based filters
type AttributeFilter struct {
	AttributeName string   `json:"attribute_name"`
	FilterType    string   `json:"filter_type"` // range, checkbox, radio
	Options       []string `json:"options,omitempty"`
}

// CustomFilter represents custom filtering options
type CustomFilter struct {
	Name     string      `json:"name"`
	Type     FilterType  `json:"type"`
	Options  []string    `json:"options,omitempty"`
	Enabled  bool        `json:"enabled"`
}

// FilterType represents filter types
type FilterType string

const (
	FilterTypeRange    FilterType = "range"
	FilterTypeCheckbox FilterType = "checkbox"
	FilterTypeRadio    FilterType = "radio"
	FilterTypeSelect   FilterType = "select"
	FilterTypeToggle   FilterType = "toggle"
)

// MarketplaceCategory represents category mapping to external marketplaces
type MarketplaceCategory struct {
	ID                uint        `json:"id" gorm:"primaryKey"`
	CategoryID        uint        `json:"category_id" gorm:"index;not null"`
	Category          Category    `json:"category" gorm:"foreignKey:CategoryID"`
	
	MarketplaceName   string      `json:"marketplace_name" gorm:"size:100;not null"`
	ExternalCategoryID string     `json:"external_category_id" gorm:"size:255;not null"`
	ExternalCategoryName string   `json:"external_category_name" gorm:"size:500"`
	ExternalPath      string      `json:"external_path" gorm:"size:1000"`
	
	// Mapping Configuration
	IsActive          bool        `json:"is_active" gorm:"default:true"`
	CommissionRate    float64     `json:"commission_rate" gorm:"type:decimal(5,4);default:0"`
	
	// Sync Information
	LastSyncAt        *time.Time  `json:"last_sync_at"`
	SyncStatus        SyncStatus  `json:"sync_status" gorm:"default:'pending'"`
	SyncError         string      `json:"sync_error" gorm:"type:text"`
	
	// Timestamps
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}

// SyncStatus represents synchronization status
type SyncStatus string

const (
	SyncStatusPending SyncStatus = "pending"
	SyncStatusSynced  SyncStatus = "synced"
	SyncStatusFailed  SyncStatus = "failed"
	SyncStatusPartial SyncStatus = "partial"
)

// CategoryTag represents tags associated with categories
type CategoryTag struct {
	ID         uint     `json:"id" gorm:"primaryKey"`
	CategoryID uint     `json:"category_id" gorm:"index;not null"`
	Category   Category `json:"category" gorm:"foreignKey:CategoryID"`
	
	TagName    string   `json:"tag_name" gorm:"size:100;not null"`
	TagType    TagType  `json:"tag_type" gorm:"default:'general'"`
	Color      string   `json:"color" gorm:"size:7"`
	
	CreatedAt  time.Time `json:"created_at"`
}

// TagType represents tag types
type TagType string

const (
	TagTypeGeneral    TagType = "general"
	TagTypeSeasonal   TagType = "seasonal"
	TagTypePromotion  TagType = "promotion"
	TagTypeTrending   TagType = "trending"
	TagTypeNew        TagType = "new"
	TagTypeFeatured   TagType = "featured"
)

// CategoryTree represents the hierarchical category structure
type CategoryTree struct {
	Category Category       `json:"category"`
	Children []CategoryTree `json:"children,omitempty"`
}

// Implement driver.Valuer interfaces
func (ca CategoryAttributes) Value() (driver.Value, error) {
	return json.Marshal(ca)
}

func (ca *CategoryAttributes) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, ca)
}

func (cf CategoryFilters) Value() (driver.Value, error) {
	return json.Marshal(cf)
}

func (cf *CategoryFilters) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, cf)
}

// TableName methods
func (Category) TableName() string {
	return "categories"
}

func (MarketplaceCategory) TableName() string {
	return "marketplace_categories"
}

func (CategoryTag) TableName() string {
	return "category_tags"
}

// Category methods
func (c *Category) IsRoot() bool {
	return c.ParentID == nil
}

func (c *Category) HasChildren() bool {
	return len(c.Children) > 0
}

func (c *Category) GetFullPath() string {
	if c.Parent == nil {
		return c.Name
	}
	return c.Parent.GetFullPath() + " > " + c.Name
}

func (c *Category) GetBreadcrumbs() []Category {
	var breadcrumbs []Category
	current := c
	
	for current != nil {
		breadcrumbs = append([]Category{*current}, breadcrumbs...)
		current = current.Parent
	}
	
	return breadcrumbs
}

func (c *Category) UpdateProductCount(count int) {
	c.ProductCount = count
}

func (c *Category) IncrementViewCount() {
	c.ViewCount++
}

func (c *Category) IncrementClickCount() {
	c.ClickCount++
}

func (c *Category) IsVisibleToPublic() bool {
	return c.IsActive && c.IsVisible
}

func (c *Category) GetDisplayName() string {
	if c.DisplayName != "" {
		return c.DisplayName
	}
	return c.Name
}

func (c *Category) HasRequiredAttributes() bool {
	return len(c.Attributes.RequiredAttributes) > 0
}

func (c *Category) GetRequiredAttributeNames() []string {
	var names []string
	for _, attr := range c.Attributes.RequiredAttributes {
		names = append(names, attr.Name)
	}
	return names
}

func (c *Category) HasFilters() bool {
	return len(c.Filters.PriceRanges) > 0 ||
		   len(c.Filters.BrandFilters) > 0 ||
		   len(c.Filters.AttributeFilters) > 0 ||
		   len(c.Filters.CustomFilters) > 0
}

// MarketplaceCategory methods
func (mc *MarketplaceCategory) NeedsSync() bool {
	return mc.SyncStatus == SyncStatusPending || mc.SyncStatus == SyncStatusFailed
}

func (mc *MarketplaceCategory) MarkSynced() {
	now := time.Now()
	mc.LastSyncAt = &now
	mc.SyncStatus = SyncStatusSynced
	mc.SyncError = ""
}

func (mc *MarketplaceCategory) MarkSyncFailed(error string) {
	now := time.Now()
	mc.LastSyncAt = &now
	mc.SyncStatus = SyncStatusFailed
	mc.SyncError = error
}

// Helper functions for building category trees
func BuildCategoryTree(categories []Category) []CategoryTree {
	categoryMap := make(map[uint]*Category)
	var rootCategories []CategoryTree
	
	// Create a map for quick lookup
	for i := range categories {
		categoryMap[categories[i].ID] = &categories[i]
	}
	
	// Build the tree structure
	for _, category := range categories {
		if category.ParentID == nil {
			// Root category
			tree := CategoryTree{
				Category: category,
				Children: buildChildren(category.ID, categoryMap),
			}
			rootCategories = append(rootCategories, tree)
		}
	}
	
	return rootCategories
}

func buildChildren(parentID uint, categoryMap map[uint]*Category) []CategoryTree {
	var children []CategoryTree
	
	for _, category := range categoryMap {
		if category.ParentID != nil && *category.ParentID == parentID {
			tree := CategoryTree{
				Category: *category,
				Children: buildChildren(category.ID, categoryMap),
			}
			children = append(children, tree)
		}
	}
	
	return children
}

func FlattenCategoryTree(tree []CategoryTree) []Category {
	var categories []Category
	
	for _, node := range tree {
		categories = append(categories, node.Category)
		categories = append(categories, FlattenCategoryTree(node.Children)...)
	}
	
	return categories
}