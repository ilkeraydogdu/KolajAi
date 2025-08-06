package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// AICredit represents AI credits for users
type AICredit struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Credits   int       `json:"credits" db:"credits"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// AICreditTransaction represents AI credit transactions
type AICreditTransaction struct {
	ID            int64     `json:"id" db:"id"`
	UserID        int64     `json:"user_id" db:"user_id"`
	Type          string    `json:"type" db:"type"` // purchase, deduct, refund
	Amount        int       `json:"amount" db:"amount"`
	Description   string    `json:"description" db:"description"`
	ReferenceType string    `json:"reference_type" db:"reference_type"` // image_generation, content_generation, etc.
	ReferenceID   int64     `json:"reference_id" db:"reference_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// AIGeneratedContent represents AI-generated content
type AIGeneratedContent struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Type      string    `json:"type" db:"type"` // image, text, template
	Model     string    `json:"model" db:"model"`
	Prompt    string    `json:"prompt" db:"prompt"`
	Content   string    `json:"content" db:"content"`
	Metadata  JSONB     `json:"metadata" db:"metadata"`
	Credits   int       `json:"credits" db:"credits"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// AITemplate represents AI-generated templates
type AITemplate struct {
	ID         int64     `json:"id" db:"id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	Name       string    `json:"name" db:"name"`
	Type       string    `json:"type" db:"type"` // instagram_post, telegram_ad, etc.
	Design     JSONB     `json:"design" db:"design"`
	Thumbnail  string    `json:"thumbnail" db:"thumbnail"`
	IsPublic   bool      `json:"is_public" db:"is_public"`
	UsageCount int       `json:"usage_count" db:"usage_count"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// AIChatSession represents AI chat sessions
type AIChatSession struct {
	ID        string    `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Context   string    `json:"context" db:"context"`
	Messages  JSONB     `json:"messages" db:"messages"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// MarketplaceIntegrationConfig represents marketplace integration configurations
type MarketplaceIntegrationConfig struct {
	ID            int64     `json:"id" db:"id"`
	UserID        int64     `json:"user_id" db:"user_id"`
	IntegrationID string    `json:"integration_id" db:"integration_id"`
	Credentials   JSONB     `json:"credentials" db:"credentials"`
	Settings      JSONB     `json:"settings" db:"settings"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	LastSync      time.Time `json:"last_sync" db:"last_sync"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// MarketplaceSyncLog represents marketplace sync logs
type MarketplaceSyncLog struct {
	ID            int64     `json:"id" db:"id"`
	UserID        int64     `json:"user_id" db:"user_id"`
	IntegrationID string    `json:"integration_id" db:"integration_id"`
	SyncType      string    `json:"sync_type" db:"sync_type"` // product, order, inventory
	Status        string    `json:"status" db:"status"`       // success, failed, partial
	Details       JSONB     `json:"details" db:"details"`
	ErrorMessage  string    `json:"error_message" db:"error_message"`
	ItemsCount    int       `json:"items_count" db:"items_count"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// JSONB is a custom type for JSON data in database
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}
