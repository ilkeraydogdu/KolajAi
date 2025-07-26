package models

import "time"

// Category represents a product category
type Category struct {
    ID          int       `json:"id" db:"id"`
    Name        string    `json:"name" db:"name"`
    Description string    `json:"description" db:"description"`
    ParentID    *int      `json:"parent_id" db:"parent_id"`
    Image       string    `json:"image" db:"image"`
    IsActive    bool      `json:"is_active" db:"is_active"`
    SortOrder   int       `json:"sort_order" db:"sort_order"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Product represents a product in the marketplace
type Product struct {
    ID              int       `json:"id" db:"id"`
    VendorID        int       `json:"vendor_id" db:"vendor_id"`
    CategoryID      int       `json:"category_id" db:"category_id"`
    Name            string    `json:"name" db:"name"`
    Description     string    `json:"description" db:"description"`
    ShortDesc       string    `json:"short_desc" db:"short_desc"`
    SKU             string    `json:"sku" db:"sku"`
    Price           float64   `json:"price" db:"price"`
    ComparePrice    float64   `json:"compare_price" db:"compare_price"`
    CostPrice       float64   `json:"cost_price" db:"cost_price"`
    WholesalePrice  float64   `json:"wholesale_price" db:"wholesale_price"`
    MinWholesaleQty int       `json:"min_wholesale_qty" db:"min_wholesale_qty"`
    Stock           int       `json:"stock" db:"stock"`
    MinStock        int       `json:"min_stock" db:"min_stock"`
    Weight          float64   `json:"weight" db:"weight"`
    Dimensions      string    `json:"dimensions" db:"dimensions"`
    Status          string    `json:"status" db:"status"` // draft, active, inactive, out_of_stock
    IsDigital       bool      `json:"is_digital" db:"is_digital"`
    IsFeatured      bool      `json:"is_featured" db:"is_featured"`
    AllowReviews    bool      `json:"allow_reviews" db:"allow_reviews"`
    MetaTitle       string    `json:"meta_title" db:"meta_title"`
    MetaDesc        string    `json:"meta_desc" db:"meta_desc"`
    Tags            string    `json:"tags" db:"tags"`
    ViewCount       int       `json:"view_count" db:"view_count"`
    SalesCount      int       `json:"sales_count" db:"sales_count"`
    Rating          float64   `json:"rating" db:"rating"`
    ReviewCount     int       `json:"review_count" db:"review_count"`
    CreatedAt       time.Time `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
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
    ID        int       `json:"id" db:"id"`
    ProductID int       `json:"product_id" db:"product_id"`
    UserID    int       `json:"user_id" db:"user_id"`
    OrderID   int       `json:"order_id" db:"order_id"`
    Rating    int       `json:"rating" db:"rating"`
    Title     string    `json:"title" db:"title"`
    Comment   string    `json:"comment" db:"comment"`
    Images    string    `json:"images" db:"images"`
    IsVerified bool     `json:"is_verified" db:"is_verified"`
    Status    string    `json:"status" db:"status"` // pending, approved, rejected
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}