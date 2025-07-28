package models

import (
	"encoding/json"
	"time"
)

// ImageDimensions represents image dimensions
type ImageDimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// AITemplate represents an AI-generated template
type AITemplate struct {
	ID          int64                  `json:"id" db:"id"`
	UserID      int64                  `json:"user_id" db:"user_id"`
	Name        string                 `json:"name" db:"name"`
	Type        AITemplateType         `json:"type" db:"type"`
	Category    string                 `json:"category" db:"category"`
	Content     json.RawMessage        `json:"content" db:"content"`
	Metadata    json.RawMessage        `json:"metadata" db:"metadata"`
	IsPublic    bool                   `json:"is_public" db:"is_public"`
	IsActive    bool                   `json:"is_active" db:"is_active"`
	UsageCount  int64                  `json:"usage_count" db:"usage_count"`
	Rating      float64                `json:"rating" db:"rating"`
	Tags        []string               `json:"tags"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// AITemplateType represents different template types
type AITemplateType string

const (
	TemplateTypeSocialMedia    AITemplateType = "social_media"
	TemplateTypeProductImage   AITemplateType = "product_image"
	TemplateTypeProductDesc    AITemplateType = "product_description"
	TemplateTypeMarketingEmail AITemplateType = "marketing_email"
	TemplateTypeBanner         AITemplateType = "banner"
	TemplateTypeStory          AITemplateType = "story"
	TemplateTypePost           AITemplateType = "post"
	TemplateTypeTelegram       AITemplateType = "telegram"
	TemplateTypeInstagram      AITemplateType = "instagram"
	TemplateTypeFacebook       AITemplateType = "facebook"
	TemplateTypeTwitter        AITemplateType = "twitter"
)

// SocialMediaTemplate represents social media specific template content
type SocialMediaTemplate struct {
	Platform    string            `json:"platform"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	ImageURL    string            `json:"image_url"`
	Price       float64           `json:"price"`
	Currency    string            `json:"currency"`
	CTAText     string            `json:"cta_text"`
	CTALink     string            `json:"cta_link"`
	Hashtags    []string          `json:"hashtags"`
	Colors      map[string]string `json:"colors"`
	Fonts       map[string]string `json:"fonts"`
	Layout      string            `json:"layout"`
	Dimensions  ImageDimensions   `json:"dimensions"`
}

// ProductImageTemplate represents product image template
type ProductImageTemplate struct {
	BackgroundType  string            `json:"background_type"`
	BackgroundColor string            `json:"background_color"`
	BackgroundImage string            `json:"background_image"`
	Filters         []string          `json:"filters"`
	Effects         []string          `json:"effects"`
	Overlays        []TemplateOverlay `json:"overlays"`
	Watermark       *Watermark        `json:"watermark,omitempty"`
	Dimensions      ImageDimensions   `json:"dimensions"`
}

// TemplateOverlay represents overlay elements on template
type TemplateOverlay struct {
	Type     string  `json:"type"` // text, image, shape
	Content  string  `json:"content"`
	X        int     `json:"x"`
	Y        int     `json:"y"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	Color    string  `json:"color"`
	FontSize int     `json:"font_size"`
	FontName string  `json:"font_name"`
	Opacity  float64 `json:"opacity"`
}

// Watermark represents watermark settings
type Watermark struct {
	Type     string  `json:"type"` // text, image
	Content  string  `json:"content"`
	Position string  `json:"position"` // top-left, top-right, bottom-left, bottom-right, center
	Opacity  float64 `json:"opacity"`
	Size     int     `json:"size"`
}

// AITemplateUsage tracks template usage
type AITemplateUsage struct {
	ID         int64     `json:"id" db:"id"`
	TemplateID int64     `json:"template_id" db:"template_id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	ProductID  *int64    `json:"product_id,omitempty" db:"product_id"`
	Platform   string    `json:"platform" db:"platform"`
	UsedAt     time.Time `json:"used_at" db:"used_at"`
	Success    bool      `json:"success" db:"success"`
	OutputURL  string    `json:"output_url" db:"output_url"`
}

// AITemplateRating represents user ratings for templates
type AITemplateRating struct {
	ID         int64     `json:"id" db:"id"`
	TemplateID int64     `json:"template_id" db:"template_id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	Rating     int       `json:"rating" db:"rating"` // 1-5
	Comment    string    `json:"comment" db:"comment"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}