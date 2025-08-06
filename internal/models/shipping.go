package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// ShippingMethod represents a shipping method
type ShippingMethod struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// Basic Information
	Name        string `json:"name" gorm:"size:200;not null" validate:"required,min=2,max=200"`
	Description string `json:"description" gorm:"type:text"`
	Code        string `json:"code" gorm:"size:50;unique;not null" validate:"required"`

	// Provider Information
	Provider            ShippingProvider `json:"provider" gorm:"not null"`
	ProviderServiceCode string           `json:"provider_service_code" gorm:"size:100"`

	// Pricing
	BaseCost              float64  `json:"base_cost" gorm:"type:decimal(15,2);not null"`
	Currency              string   `json:"currency" gorm:"size:3;default:'TRY'"`
	FreeShippingThreshold *float64 `json:"free_shipping_threshold" gorm:"type:decimal(15,2)"`

	// Delivery Time
	MinDeliveryDays int    `json:"min_delivery_days" gorm:"not null"`
	MaxDeliveryDays int    `json:"max_delivery_days" gorm:"not null"`
	DeliveryTime    string `json:"delivery_time" gorm:"size:100"` // e.g., "1-3 business days"

	// Availability
	IsActive  bool `json:"is_active" gorm:"default:true"`
	IsDefault bool `json:"is_default" gorm:"default:false"`

	// Geographic Coverage
	AvailableCountries []string `json:"available_countries" gorm:"type:json"`
	ExcludedCountries  []string `json:"excluded_countries" gorm:"type:json"`
	AvailableCities    []string `json:"available_cities" gorm:"type:json"`
	ExcludedCities     []string `json:"excluded_cities" gorm:"type:json"`

	// Weight and Size Limits
	MaxWeight *float64 `json:"max_weight" gorm:"type:decimal(10,3)"` // in kg
	MaxLength *float64 `json:"max_length" gorm:"type:decimal(10,2)"` // in cm
	MaxWidth  *float64 `json:"max_width" gorm:"type:decimal(10,2)"`  // in cm
	MaxHeight *float64 `json:"max_height" gorm:"type:decimal(10,2)"` // in cm

	// Features
	HasTracking          bool `json:"has_tracking" gorm:"default:true"`
	HasInsurance         bool `json:"has_insurance" gorm:"default:false"`
	HasCOD               bool `json:"has_cod" gorm:"default:false"` // Cash on Delivery
	HasSignature         bool `json:"has_signature" gorm:"default:false"`
	HasScheduledDelivery bool `json:"has_scheduled_delivery" gorm:"default:false"`

	// Vendor Restrictions
	VendorID *uint   `json:"vendor_id" gorm:"index"`
	Vendor   *Vendor `json:"vendor,omitempty" gorm:"foreignKey:VendorID"`

	// Metadata
	Metadata ShippingMethodMetadata `json:"metadata" gorm:"type:json"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// ShippingProvider represents shipping service providers
type ShippingProvider string

const (
	ShippingProviderPTT      ShippingProvider = "ptt"
	ShippingProviderMNG      ShippingProvider = "mng"
	ShippingProviderYurtici  ShippingProvider = "yurtici"
	ShippingProviderAras     ShippingProvider = "aras"
	ShippingProviderUPS      ShippingProvider = "ups"
	ShippingProviderDHL      ShippingProvider = "dhl"
	ShippingProviderFedEx    ShippingProvider = "fedex"
	ShippingProviderSendeo   ShippingProvider = "sendeo"
	ShippingProviderTrendyol ShippingProvider = "trendyol"
	ShippingProviderHepsijet ShippingProvider = "hepsijet"
	ShippingProviderInternal ShippingProvider = "internal"
)

// ShippingMethodMetadata holds additional shipping method data
type ShippingMethodMetadata struct {
	ApiEndpoint    string                 `json:"api_endpoint,omitempty"`
	ApiCredentials map[string]string      `json:"api_credentials,omitempty"`
	CustomSettings map[string]interface{} `json:"custom_settings,omitempty"`
	InternalNotes  string                 `json:"internal_notes,omitempty"`
}

// Shipment represents a shipment
type Shipment struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// Order Relationship
	OrderID uint  `json:"order_id" gorm:"index;not null"`
	Order   Order `json:"order" gorm:"foreignKey:OrderID"`

	// Customer and Address
	CustomerID uint     `json:"customer_id" gorm:"index;not null"`
	Customer   Customer `json:"customer" gorm:"foreignKey:CustomerID"`

	// Shipping Method
	ShippingMethodID uint           `json:"shipping_method_id" gorm:"index;not null"`
	ShippingMethod   ShippingMethod `json:"shipping_method" gorm:"foreignKey:ShippingMethodID"`

	// Addresses
	FromAddress ShippingAddress `json:"from_address" gorm:"type:json"`
	ToAddress   ShippingAddress `json:"to_address" gorm:"type:json"`

	// Tracking Information
	TrackingNumber     string         `json:"tracking_number" gorm:"size:100;index"`
	ProviderTrackingID string         `json:"provider_tracking_id" gorm:"size:100"`
	Status             ShipmentStatus `json:"status" gorm:"default:'pending'"`

	// Package Information
	PackageCount int     `json:"package_count" gorm:"default:1"`
	TotalWeight  float64 `json:"total_weight" gorm:"type:decimal(10,3)"` // in kg
	TotalVolume  float64 `json:"total_volume" gorm:"type:decimal(10,3)"` // in cubic cm

	// Costs
	ShippingCost  float64 `json:"shipping_cost" gorm:"type:decimal(15,2);not null"`
	InsuranceCost float64 `json:"insurance_cost" gorm:"type:decimal(15,2);default:0"`
	CODCost       float64 `json:"cod_cost" gorm:"type:decimal(15,2);default:0"`
	TotalCost     float64 `json:"total_cost" gorm:"type:decimal(15,2);not null"`
	Currency      string  `json:"currency" gorm:"size:3;default:'TRY'"`

	// Delivery Information
	EstimatedDeliveryDate *time.Time `json:"estimated_delivery_date"`
	ActualDeliveryDate    *time.Time `json:"actual_delivery_date"`
	DeliveryAttempts      int        `json:"delivery_attempts" gorm:"default:0"`

	// Special Services
	HasInsurance      bool    `json:"has_insurance" gorm:"default:false"`
	InsuranceValue    float64 `json:"insurance_value" gorm:"type:decimal(15,2);default:0"`
	HasCOD            bool    `json:"has_cod" gorm:"default:false"`
	CODAmount         float64 `json:"cod_amount" gorm:"type:decimal(15,2);default:0"`
	RequiresSignature bool    `json:"requires_signature" gorm:"default:false"`

	// Delivery Instructions
	DeliveryInstructions string `json:"delivery_instructions" gorm:"type:text"`
	SpecialInstructions  string `json:"special_instructions" gorm:"type:text"`

	// Provider Response
	ProviderResponse ShipmentProviderResponse `json:"provider_response" gorm:"type:json"`

	// Labels and Documents
	ShippingLabel   string `json:"shipping_label" gorm:"type:text"`   // Base64 encoded
	InvoiceDocument string `json:"invoice_document" gorm:"type:text"` // Base64 encoded
	CustomsDocument string `json:"customs_document" gorm:"type:text"` // Base64 encoded

	// Tracking Events
	TrackingEvents []ShipmentTrackingEvent `json:"tracking_events,omitempty" gorm:"foreignKey:ShipmentID"`

	// Error Information
	ErrorCode    string `json:"error_code" gorm:"size:100"`
	ErrorMessage string `json:"error_message" gorm:"size:500"`

	// Timestamps
	ShippedAt   *time.Time `json:"shipped_at"`
	DeliveredAt *time.Time `json:"delivered_at"`
	CancelledAt *time.Time `json:"cancelled_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// ShipmentStatus represents shipment status
type ShipmentStatus string

const (
	ShipmentStatusPending        ShipmentStatus = "pending"
	ShipmentStatusProcessing     ShipmentStatus = "processing"
	ShipmentStatusShipped        ShipmentStatus = "shipped"
	ShipmentStatusInTransit      ShipmentStatus = "in_transit"
	ShipmentStatusOutForDelivery ShipmentStatus = "out_for_delivery"
	ShipmentStatusDelivered      ShipmentStatus = "delivered"
	ShipmentStatusReturned       ShipmentStatus = "returned"
	ShipmentStatusCancelled      ShipmentStatus = "cancelled"
	ShipmentStatusLost           ShipmentStatus = "lost"
	ShipmentStatusDamaged        ShipmentStatus = "damaged"
	ShipmentStatusException      ShipmentStatus = "exception"
)

// ShippingAddress represents shipping address information
type ShippingAddress struct {
	FirstName    string   `json:"first_name"`
	LastName     string   `json:"last_name"`
	CompanyName  string   `json:"company_name,omitempty"`
	AddressLine1 string   `json:"address_line1"`
	AddressLine2 string   `json:"address_line2,omitempty"`
	City         string   `json:"city"`
	State        string   `json:"state,omitempty"`
	PostalCode   string   `json:"postal_code"`
	Country      string   `json:"country"`
	Phone        string   `json:"phone,omitempty"`
	Email        string   `json:"email,omitempty"`
	Latitude     *float64 `json:"latitude,omitempty"`
	Longitude    *float64 `json:"longitude,omitempty"`
}

// ShipmentProviderResponse holds provider API response
type ShipmentProviderResponse struct {
	RawResponse     map[string]interface{} `json:"raw_response,omitempty"`
	ProviderStatus  string                 `json:"provider_status,omitempty"`
	ProviderCode    string                 `json:"provider_code,omitempty"`
	ProviderMessage string                 `json:"provider_message,omitempty"`
	APICallTime     time.Time              `json:"api_call_time,omitempty"`
}

// ShipmentTrackingEvent represents a tracking event
type ShipmentTrackingEvent struct {
	ID         uint     `json:"id" gorm:"primaryKey"`
	ShipmentID uint     `json:"shipment_id" gorm:"index;not null"`
	Shipment   Shipment `json:"shipment" gorm:"foreignKey:ShipmentID"`

	// Event Details
	Status      ShipmentStatus `json:"status" gorm:"not null"`
	Description string         `json:"description" gorm:"size:500"`
	Location    string         `json:"location" gorm:"size:200"`

	// Provider Information
	ProviderEventCode string `json:"provider_event_code" gorm:"size:100"`
	ProviderEventName string `json:"provider_event_name" gorm:"size:200"`

	// Timing
	EventDate time.Time `json:"event_date" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
}

// ShippingRate represents shipping rate calculation
type ShippingRate struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	ShippingMethodID uint           `json:"shipping_method_id" gorm:"index;not null"`
	ShippingMethod   ShippingMethod `json:"shipping_method" gorm:"foreignKey:ShippingMethodID"`

	// Rate Conditions
	MinWeight      *float64 `json:"min_weight" gorm:"type:decimal(10,3)"`
	MaxWeight      *float64 `json:"max_weight" gorm:"type:decimal(10,3)"`
	MinOrderAmount *float64 `json:"min_order_amount" gorm:"type:decimal(15,2)"`
	MaxOrderAmount *float64 `json:"max_order_amount" gorm:"type:decimal(15,2)"`

	// Geographic Conditions
	Country         string `json:"country" gorm:"size:2"`
	State           string `json:"state" gorm:"size:100"`
	City            string `json:"city" gorm:"size:100"`
	PostalCodeRange string `json:"postal_code_range" gorm:"size:100"`

	// Rate Calculation
	RateType       RateType `json:"rate_type" gorm:"not null"`
	BaseRate       float64  `json:"base_rate" gorm:"type:decimal(15,2);not null"`
	PerKgRate      float64  `json:"per_kg_rate" gorm:"type:decimal(15,2);default:0"`
	PerItemRate    float64  `json:"per_item_rate" gorm:"type:decimal(15,2);default:0"`
	PercentageRate float64  `json:"percentage_rate" gorm:"type:decimal(5,4);default:0"`

	// Priority
	Priority int  `json:"priority" gorm:"default:0"`
	IsActive bool `json:"is_active" gorm:"default:true"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// RateType represents shipping rate calculation type
type RateType string

const (
	RateTypeFlat       RateType = "flat"       // Fixed rate
	RateTypeWeight     RateType = "weight"     // Based on weight
	RateTypeQuantity   RateType = "quantity"   // Based on item count
	RateTypePercentage RateType = "percentage" // Percentage of order value
	RateTypeTiered     RateType = "tiered"     // Tiered pricing
)

// Implement driver.Valuer interfaces
func (sma ShippingMethodMetadata) Value() (driver.Value, error) {
	return json.Marshal(sma)
}

func (sma *ShippingMethodMetadata) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, sma)
}

func (sa ShippingAddress) Value() (driver.Value, error) {
	return json.Marshal(sa)
}

func (sa *ShippingAddress) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, sa)
}

func (spr ShipmentProviderResponse) Value() (driver.Value, error) {
	return json.Marshal(spr)
}

func (spr *ShipmentProviderResponse) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, spr)
}

// TableName methods
func (ShippingMethod) TableName() string {
	return "shipping_methods"
}

func (Shipment) TableName() string {
	return "shipments"
}

func (ShipmentTrackingEvent) TableName() string {
	return "shipment_tracking_events"
}

func (ShippingRate) TableName() string {
	return "shipping_rates"
}

// ShippingMethod methods
func (sm *ShippingMethod) IsAvailableInCountry(country string) bool {
	// If no restrictions, available everywhere
	if len(sm.AvailableCountries) == 0 && len(sm.ExcludedCountries) == 0 {
		return true
	}

	// Check excluded countries first
	for _, excluded := range sm.ExcludedCountries {
		if excluded == country {
			return false
		}
	}

	// If available countries specified, check inclusion
	if len(sm.AvailableCountries) > 0 {
		for _, available := range sm.AvailableCountries {
			if available == country {
				return true
			}
		}
		return false
	}

	return true
}

func (sm *ShippingMethod) CanHandleWeight(weight float64) bool {
	return sm.MaxWeight == nil || weight <= *sm.MaxWeight
}

func (sm *ShippingMethod) CanHandleDimensions(length, width, height float64) bool {
	if sm.MaxLength != nil && length > *sm.MaxLength {
		return false
	}
	if sm.MaxWidth != nil && width > *sm.MaxWidth {
		return false
	}
	if sm.MaxHeight != nil && height > *sm.MaxHeight {
		return false
	}
	return true
}

func (sm *ShippingMethod) GetEstimatedDeliveryDate() time.Time {
	now := time.Now()
	// Use average of min and max delivery days
	avgDays := (sm.MinDeliveryDays + sm.MaxDeliveryDays) / 2
	return now.AddDate(0, 0, avgDays)
}

// Shipment methods
func (s *Shipment) IsDelivered() bool {
	return s.Status == ShipmentStatusDelivered
}

func (s *Shipment) IsCancelled() bool {
	return s.Status == ShipmentStatusCancelled
}

func (s *Shipment) IsInTransit() bool {
	return s.Status == ShipmentStatusInTransit || s.Status == ShipmentStatusOutForDelivery
}

func (s *Shipment) GetLatestTrackingEvent() *ShipmentTrackingEvent {
	if len(s.TrackingEvents) == 0 {
		return nil
	}

	latest := &s.TrackingEvents[0]
	for i := 1; i < len(s.TrackingEvents); i++ {
		if s.TrackingEvents[i].EventDate.After(latest.EventDate) {
			latest = &s.TrackingEvents[i]
		}
	}
	return latest
}

func (s *Shipment) GetFullAddress() string {
	addr := s.ToAddress
	fullAddr := addr.AddressLine1
	if addr.AddressLine2 != "" {
		fullAddr += ", " + addr.AddressLine2
	}
	fullAddr += ", " + addr.City
	if addr.State != "" {
		fullAddr += ", " + addr.State
	}
	fullAddr += " " + addr.PostalCode + ", " + addr.Country
	return fullAddr
}

// ShippingRate methods
func (sr *ShippingRate) CalculateRate(weight float64, orderAmount float64, itemCount int) float64 {
	switch sr.RateType {
	case RateTypeFlat:
		return sr.BaseRate

	case RateTypeWeight:
		return sr.BaseRate + (weight * sr.PerKgRate)

	case RateTypeQuantity:
		return sr.BaseRate + (float64(itemCount) * sr.PerItemRate)

	case RateTypePercentage:
		return orderAmount * (sr.PercentageRate / 100)

	default:
		return sr.BaseRate
	}
}

func (sr *ShippingRate) IsApplicable(weight float64, orderAmount float64, country, state, city string) bool {
	// Check weight limits
	if sr.MinWeight != nil && weight < *sr.MinWeight {
		return false
	}
	if sr.MaxWeight != nil && weight > *sr.MaxWeight {
		return false
	}

	// Check order amount limits
	if sr.MinOrderAmount != nil && orderAmount < *sr.MinOrderAmount {
		return false
	}
	if sr.MaxOrderAmount != nil && orderAmount > *sr.MaxOrderAmount {
		return false
	}

	// Check geographic conditions
	if sr.Country != "" && sr.Country != country {
		return false
	}
	if sr.State != "" && sr.State != state {
		return false
	}
	if sr.City != "" && sr.City != city {
		return false
	}

	return sr.IsActive
}
