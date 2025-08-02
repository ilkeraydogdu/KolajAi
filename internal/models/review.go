package models

import (
	"time"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Review represents a product review
type Review struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	
	// Relationships
	ProductID   uint      `json:"product_id" gorm:"index;not null"`
	Product     Product   `json:"product" gorm:"foreignKey:ProductID"`
	CustomerID  uint      `json:"customer_id" gorm:"index;not null"`
	Customer    Customer  `json:"customer" gorm:"foreignKey:CustomerID"`
	OrderID     *uint     `json:"order_id" gorm:"index"`
	Order       *Order    `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	VendorID    *uint     `json:"vendor_id" gorm:"index"`
	Vendor      *Vendor   `json:"vendor,omitempty" gorm:"foreignKey:VendorID"`
	
	// Review Content
	Title       string    `json:"title" gorm:"size:200;not null" validate:"required,min=5,max=200"`
	Content     string    `json:"content" gorm:"type:text;not null" validate:"required,min=10,max=2000"`
	Rating      int       `json:"rating" gorm:"not null" validate:"required,min=1,max=5"`
	
	// Review Categories (specific ratings)
	QualityRating    *int  `json:"quality_rating" validate:"omitempty,min=1,max=5"`
	ValueRating      *int  `json:"value_rating" validate:"omitempty,min=1,max=5"`
	ServiceRating    *int  `json:"service_rating" validate:"omitempty,min=1,max=5"`
	DeliveryRating   *int  `json:"delivery_rating" validate:"omitempty,min=1,max=5"`
	
	// Review Status and Moderation
	Status      ReviewStatus `json:"status" gorm:"default:'pending'"`
	IsVerified  bool         `json:"is_verified" gorm:"default:false"`
	IsFeatured  bool         `json:"is_featured" gorm:"default:false"`
	
	// Engagement
	HelpfulCount    int    `json:"helpful_count" gorm:"default:0"`
	NotHelpfulCount int    `json:"not_helpful_count" gorm:"default:0"`
	ReportCount     int    `json:"report_count" gorm:"default:0"`
	
	// Media Attachments
	Images          ReviewImages `json:"images" gorm:"type:json"`
	Videos          ReviewVideos `json:"videos" gorm:"type:json"`
	
	// Purchase Verification
	PurchaseVerified bool      `json:"purchase_verified" gorm:"default:false"`
	PurchaseDate     *time.Time `json:"purchase_date"`
	
	// Moderation
	ModeratedBy     *uint     `json:"moderated_by" gorm:"index"`
	ModeratedAt     *time.Time `json:"moderated_at"`
	ModerationNotes string    `json:"moderation_notes" gorm:"type:text"`
	
	// Response from Vendor
	VendorResponse  *ReviewResponse `json:"vendor_response,omitempty" gorm:"foreignKey:ReviewID"`
	
	// Metadata
	Metadata        ReviewMetadata `json:"metadata" gorm:"type:json"`
	IPAddress       string         `json:"ip_address" gorm:"size:45"`
	UserAgent       string         `json:"user_agent" gorm:"size:500"`
	
	// Timestamps
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// ReviewStatus represents the status of a review
type ReviewStatus string

const (
	ReviewStatusPending   ReviewStatus = "pending"
	ReviewStatusApproved  ReviewStatus = "approved"
	ReviewStatusRejected  ReviewStatus = "rejected"
	ReviewStatusFlagged   ReviewStatus = "flagged"
	ReviewStatusHidden    ReviewStatus = "hidden"
	ReviewStatusSpam      ReviewStatus = "spam"
)

// ReviewImages holds image attachments
type ReviewImages struct {
	Images []ReviewImage `json:"images,omitempty"`
}

// ReviewImage represents an image attachment
type ReviewImage struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url"`
	Caption     string `json:"caption,omitempty"`
	Size        int64  `json:"size"`
	MimeType    string `json:"mime_type"`
}

// ReviewVideos holds video attachments
type ReviewVideos struct {
	Videos []ReviewVideo `json:"videos,omitempty"`
}

// ReviewVideo represents a video attachment
type ReviewVideo struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url"`
	Caption     string `json:"caption,omitempty"`
	Duration    int    `json:"duration"` // seconds
	Size        int64  `json:"size"`
	MimeType    string `json:"mime_type"`
}

// ReviewMetadata holds additional review data
type ReviewMetadata struct {
	DeviceType     string                 `json:"device_type,omitempty"`
	Platform       string                 `json:"platform,omitempty"`
	AppVersion     string                 `json:"app_version,omitempty"`
	Location       string                 `json:"location,omitempty"`
	LanguageCode   string                 `json:"language_code,omitempty"`
	CustomFields   map[string]interface{} `json:"custom_fields,omitempty"`
}

// ReviewResponse represents vendor response to a review
type ReviewResponse struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ReviewID    uint      `json:"review_id" gorm:"index;not null;unique"`
	Review      Review    `json:"review" gorm:"foreignKey:ReviewID"`
	VendorID    uint      `json:"vendor_id" gorm:"index;not null"`
	Vendor      Vendor    `json:"vendor" gorm:"foreignKey:VendorID"`
	
	Content     string    `json:"content" gorm:"type:text;not null" validate:"required,min=10,max=1000"`
	Status      ResponseStatus `json:"status" gorm:"default:'active'"`
	
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ResponseStatus represents vendor response status
type ResponseStatus string

const (
	ResponseStatusActive ResponseStatus = "active"
	ResponseStatusHidden ResponseStatus = "hidden"
)

// ReviewHelpful represents helpful votes for reviews
type ReviewHelpful struct {
	ID         uint     `json:"id" gorm:"primaryKey"`
	ReviewID   uint     `json:"review_id" gorm:"index;not null"`
	Review     Review   `json:"review" gorm:"foreignKey:ReviewID"`
	CustomerID uint     `json:"customer_id" gorm:"index;not null"`
	Customer   Customer `json:"customer" gorm:"foreignKey:CustomerID"`
	
	IsHelpful  bool      `json:"is_helpful"`
	CreatedAt  time.Time `json:"created_at"`
	
	// Composite unique index
	// gorm:"uniqueIndex:idx_review_customer"
}

// ReviewReport represents reports for inappropriate reviews
type ReviewReport struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	ReviewID   uint       `json:"review_id" gorm:"index;not null"`
	Review     Review     `json:"review" gorm:"foreignKey:ReviewID"`
	CustomerID uint       `json:"customer_id" gorm:"index;not null"`
	Customer   Customer   `json:"customer" gorm:"foreignKey:CustomerID"`
	
	Reason     ReportReason `json:"reason" gorm:"not null"`
	Comment    string       `json:"comment" gorm:"type:text"`
	Status     ReportStatus `json:"status" gorm:"default:'pending'"`
	
	ProcessedBy *uint      `json:"processed_by" gorm:"index"`
	ProcessedAt *time.Time `json:"processed_at"`
	
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ReportReason represents reason for reporting a review
type ReportReason string

const (
	ReportReasonSpam           ReportReason = "spam"
	ReportReasonInappropriate  ReportReason = "inappropriate"
	ReportReasonFakeReview     ReportReason = "fake_review"
	ReportReasonOffensive      ReportReason = "offensive"
	ReportReasonMisleading     ReportReason = "misleading"
	ReportReasonOther          ReportReason = "other"
)

// ReportStatus represents status of a report
type ReportStatus string

const (
	ReportStatusPending   ReportStatus = "pending"
	ReportStatusProcessed ReportStatus = "processed"
	ReportStatusDismissed ReportStatus = "dismissed"
)

// ProductRating represents aggregated ratings for a product
type ProductRating struct {
	ID              uint    `json:"id" gorm:"primaryKey"`
	ProductID       uint    `json:"product_id" gorm:"index;not null;unique"`
	Product         Product `json:"product" gorm:"foreignKey:ProductID"`
	
	// Overall Rating
	AverageRating   float64 `json:"average_rating" gorm:"type:decimal(3,2);default:0"`
	TotalReviews    int     `json:"total_reviews" gorm:"default:0"`
	
	// Rating Distribution
	Rating5Count    int     `json:"rating_5_count" gorm:"default:0"`
	Rating4Count    int     `json:"rating_4_count" gorm:"default:0"`
	Rating3Count    int     `json:"rating_3_count" gorm:"default:0"`
	Rating2Count    int     `json:"rating_2_count" gorm:"default:0"`
	Rating1Count    int     `json:"rating_1_count" gorm:"default:0"`
	
	// Category Ratings
	AverageQuality  float64 `json:"average_quality" gorm:"type:decimal(3,2);default:0"`
	AverageValue    float64 `json:"average_value" gorm:"type:decimal(3,2);default:0"`
	AverageService  float64 `json:"average_service" gorm:"type:decimal(3,2);default:0"`
	AverageDelivery float64 `json:"average_delivery" gorm:"type:decimal(3,2);default:0"`
	
	// Review Metrics
	VerifiedReviews int     `json:"verified_reviews" gorm:"default:0"`
	WithPhotos      int     `json:"with_photos" gorm:"default:0"`
	WithVideos      int     `json:"with_videos" gorm:"default:0"`
	
	// Last Update
	UpdatedAt       time.Time `json:"updated_at"`
}

// Implement driver.Valuer interface for ReviewImages
func (ri ReviewImages) Value() (driver.Value, error) {
	return json.Marshal(ri)
}

// Implement sql.Scanner interface for ReviewImages
func (ri *ReviewImages) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	
	return json.Unmarshal(bytes, ri)
}

// Implement driver.Valuer interface for ReviewVideos
func (rv ReviewVideos) Value() (driver.Value, error) {
	return json.Marshal(rv)
}

// Implement sql.Scanner interface for ReviewVideos
func (rv *ReviewVideos) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	
	return json.Unmarshal(bytes, rv)
}

// Implement driver.Valuer interface for ReviewMetadata
func (rm ReviewMetadata) Value() (driver.Value, error) {
	return json.Marshal(rm)
}

// Implement sql.Scanner interface for ReviewMetadata
func (rm *ReviewMetadata) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	
	return json.Unmarshal(bytes, rm)
}

// TableName returns the table name for Review
func (Review) TableName() string {
	return "reviews"
}

// TableName returns the table name for ReviewResponse
func (ReviewResponse) TableName() string {
	return "review_responses"
}

// TableName returns the table name for ReviewHelpful
func (ReviewHelpful) TableName() string {
	return "review_helpful"
}

// TableName returns the table name for ReviewReport
func (ReviewReport) TableName() string {
	return "review_reports"
}

// TableName returns the table name for ProductRating
func (ProductRating) TableName() string {
	return "product_ratings"
}

// IsApproved checks if review is approved
func (r *Review) IsApproved() bool {
	return r.Status == ReviewStatusApproved
}

// CanBeModerated checks if review can be moderated
func (r *Review) CanBeModerated() bool {
	return r.Status == ReviewStatusPending || r.Status == ReviewStatusFlagged
}

// HasMedia checks if review has media attachments
func (r *Review) HasMedia() bool {
	return len(r.Images.Images) > 0 || len(r.Videos.Videos) > 0
}

// GetHelpfulnessRatio returns helpfulness ratio
func (r *Review) GetHelpfulnessRatio() float64 {
	total := r.HelpfulCount + r.NotHelpfulCount
	if total == 0 {
		return 0
	}
	return float64(r.HelpfulCount) / float64(total)
}

// IsHighQuality checks if review is considered high quality
func (r *Review) IsHighQuality() bool {
	return len(r.Content) > 100 && r.HasMedia() && r.IsVerified
}

// GetOverallRating returns overall rating with category weights
func (r *Review) GetOverallRating() float64 {
	ratings := []int{r.Rating}
	
	if r.QualityRating != nil {
		ratings = append(ratings, *r.QualityRating)
	}
	if r.ValueRating != nil {
		ratings = append(ratings, *r.ValueRating)
	}
	if r.ServiceRating != nil {
		ratings = append(ratings, *r.ServiceRating)
	}
	if r.DeliveryRating != nil {
		ratings = append(ratings, *r.DeliveryRating)
	}
	
	total := 0
	for _, rating := range ratings {
		total += rating
	}
	
	return float64(total) / float64(len(ratings))
}

// GetRatingPercentages returns rating distribution as percentages
func (pr *ProductRating) GetRatingPercentages() map[int]float64 {
	if pr.TotalReviews == 0 {
		return map[int]float64{5: 0, 4: 0, 3: 0, 2: 0, 1: 0}
	}
	
	return map[int]float64{
		5: float64(pr.Rating5Count) / float64(pr.TotalReviews) * 100,
		4: float64(pr.Rating4Count) / float64(pr.TotalReviews) * 100,
		3: float64(pr.Rating3Count) / float64(pr.TotalReviews) * 100,
		2: float64(pr.Rating2Count) / float64(pr.TotalReviews) * 100,
		1: float64(pr.Rating1Count) / float64(pr.TotalReviews) * 100,
	}
}

// GetVerificationRate returns percentage of verified reviews
func (pr *ProductRating) GetVerificationRate() float64 {
	if pr.TotalReviews == 0 {
		return 0
	}
	return float64(pr.VerifiedReviews) / float64(pr.TotalReviews) * 100
}

// GetMediaRate returns percentage of reviews with media
func (pr *ProductRating) GetMediaRate() float64 {
	if pr.TotalReviews == 0 {
		return 0
	}
	mediaReviews := pr.WithPhotos + pr.WithVideos
	return float64(mediaReviews) / float64(pr.TotalReviews) * 100
}