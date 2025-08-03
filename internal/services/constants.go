package services

// Product related constants
const (
	// Pagination defaults
	DefaultProductLimit  = 20
	MaxProductLimit      = 100
	DefaultPageSize      = 20
	
	// Product limits
	MaxProductNameLength        = 255
	MaxProductDescriptionLength = 5000
	MaxProductShortDescLength   = 500
	MinProductSKULength         = 3
	MaxProductSKULength         = 100
	MaxProductTagsLength        = 1000
	MaxProductMetaTitleLength   = 255
	MaxProductMetaDescLength    = 500
	MaxProductDimensionsLength  = 100
	
	// Rating limits
	MinProductRating = 0.0
	MaxProductRating = 5.0
	
	// Review limits
	MaxReviewsForRatingCalc = 1000
	
	// Stock limits
	MinStock = 0
	MinPrice = 0.0
	
	// SKU generation
	SKUPrefixLength = 3
	
	// Status values
	ProductStatusDraft       = "draft"
	ProductStatusActive      = "active"
	ProductStatusInactive    = "inactive"
	ProductStatusOutOfStock  = "out_of_stock"
	
	ReviewStatusPending  = "pending"
	ReviewStatusApproved = "approved"
	ReviewStatusRejected = "rejected"
)

// Valid product statuses
var ValidProductStatuses = []string{
	ProductStatusDraft,
	ProductStatusActive,
	ProductStatusInactive,
	ProductStatusOutOfStock,
}

// Valid review statuses
var ValidReviewStatuses = []string{
	ReviewStatusPending,
	ReviewStatusApproved,
	ReviewStatusRejected,
}

// Allowed sort columns for products
var AllowedProductSortColumns = []string{
	"id",
	"name", 
	"price",
	"created_at",
	"updated_at",
	"rating",
	"sales_count",
	"view_count",
}

// Valid sort orders
var ValidSortOrders = []string{
	"ASC",
	"DESC",
}