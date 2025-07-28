package models

import (
	"encoding/json"
	"time"
)

// MarketplaceIntegration represents a marketplace integration
type MarketplaceIntegration struct {
	ID              int64                  `json:"id" db:"id"`
	UserID          int64                  `json:"user_id" db:"user_id"`
	Name            string                 `json:"name" db:"name"`
	Type            IntegrationType        `json:"type" db:"type"`
	Platform        string                 `json:"platform" db:"platform"`
	APIKey          string                 `json:"api_key" db:"api_key"`
	APISecret       string                 `json:"api_secret" db:"api_secret"`
	AccessToken     string                 `json:"access_token" db:"access_token"`
	RefreshToken    string                 `json:"refresh_token" db:"refresh_token"`
	Config          json.RawMessage        `json:"config" db:"config"`
	IsActive        bool                   `json:"is_active" db:"is_active"`
	LastSync        *time.Time             `json:"last_sync" db:"last_sync"`
	SyncStatus      SyncStatus             `json:"sync_status" db:"sync_status"`
	ErrorMessage    string                 `json:"error_message" db:"error_message"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// IntegrationType represents different integration types
type IntegrationType string

const (
	// Turkish Marketplaces
	IntegrationTrendyol      IntegrationType = "trendyol"
	IntegrationHepsiburada   IntegrationType = "hepsiburada"
	IntegrationCicekSepeti   IntegrationType = "ciceksepeti"
	IntegrationAmazonTR      IntegrationType = "amazon_tr"
	IntegrationPttAvm        IntegrationType = "pttavm"
	IntegrationN11           IntegrationType = "n11"
	IntegrationN11Pro        IntegrationType = "n11pro"
	IntegrationAkakce        IntegrationType = "akakce"
	IntegrationCimri         IntegrationType = "cimri"
	IntegrationModanisa      IntegrationType = "modanisa"
	IntegrationFarmazon      IntegrationType = "farmazon"
	IntegrationFlo           IntegrationType = "flo"
	IntegrationBunaDeger     IntegrationType = "bunadeger"
	IntegrationLazimBana     IntegrationType = "lazimbana"
	IntegrationAllesgo       IntegrationType = "allesgo"
	IntegrationPazarama      IntegrationType = "pazarama"
	IntegrationVodafone      IntegrationType = "vodafone"
	IntegrationFarmaborsa    IntegrationType = "farmaborsa"
	IntegrationGetirCarsi    IntegrationType = "getircarsi"
	IntegrationEcza1         IntegrationType = "ecza1"
	IntegrationTurkcellPasaj IntegrationType = "turkcellpasaj"
	IntegrationTeknosa       IntegrationType = "teknosa"
	IntegrationIdefix        IntegrationType = "idefix"
	IntegrationKoctas        IntegrationType = "koctas"
	IntegrationPempati       IntegrationType = "pempati"
	IntegrationLCW           IntegrationType = "lcw"
	IntegrationAlisGidis     IntegrationType = "alisgidis"
	IntegrationBeymen        IntegrationType = "beymen"
	IntegrationNovadan       IntegrationType = "novadan"
	IntegrationMagazanOlsun  IntegrationType = "magazanolsun"

	// International Marketplaces
	IntegrationAmazonUS      IntegrationType = "amazon_us"
	IntegrationAmazonUK      IntegrationType = "amazon_uk"
	IntegrationAmazonDE      IntegrationType = "amazon_de"
	IntegrationAmazonFR      IntegrationType = "amazon_fr"
	IntegrationAmazonNL      IntegrationType = "amazon_nl"
	IntegrationAmazonIT      IntegrationType = "amazon_it"
	IntegrationAmazonCA      IntegrationType = "amazon_ca"
	IntegrationAmazonAE      IntegrationType = "amazon_ae"
	IntegrationAmazonES      IntegrationType = "amazon_es"
	IntegrationEbay          IntegrationType = "ebay"
	IntegrationAliExpress    IntegrationType = "aliexpress"
	IntegrationEtsy          IntegrationType = "etsy"
	IntegrationOzon          IntegrationType = "ozon"
	IntegrationJoom          IntegrationType = "joom"
	IntegrationFruugo        IntegrationType = "fruugo"
	IntegrationAllegro       IntegrationType = "allegro"
	IntegrationHepsiGlobal   IntegrationType = "hepsiglobal"
	IntegrationBolcom        IntegrationType = "bolcom"
	IntegrationOnBuy         IntegrationType = "onbuy"
	IntegrationWayfair       IntegrationType = "wayfair"
	IntegrationZoodMall      IntegrationType = "zoodmall"
	IntegrationWalmart       IntegrationType = "walmart"
	IntegrationJumia         IntegrationType = "jumia"
	IntegrationZalando       IntegrationType = "zalando"
	IntegrationCdiscount     IntegrationType = "cdiscount"
	IntegrationWish          IntegrationType = "wish"
	IntegrationOtto          IntegrationType = "otto"
	IntegrationRakuten       IntegrationType = "rakuten"

	// E-commerce Platforms
	IntegrationTsoft         IntegrationType = "tsoft"
	IntegrationTicimax       IntegrationType = "ticimax"
	IntegrationIdeasoft      IntegrationType = "ideasoft"
	IntegrationPlatinMarket  IntegrationType = "platinmarket"
	IntegrationWooCommerce   IntegrationType = "woocommerce"
	IntegrationOpenCart      IntegrationType = "opencart"
	IntegrationShopPHP       IntegrationType = "shopphp"
	IntegrationShopify       IntegrationType = "shopify"
	IntegrationPrestaShop    IntegrationType = "prestashop"
	IntegrationMagento       IntegrationType = "magento"
	IntegrationEthica        IntegrationType = "ethica"
	IntegrationIkas          IntegrationType = "ikas"

	// Social Media
	IntegrationFacebookShop  IntegrationType = "facebook_shop"
	IntegrationGoogleMerchant IntegrationType = "google_merchant"
	IntegrationInstagramShop IntegrationType = "instagram_shop"

	// Accounting/ERP
	IntegrationLogo          IntegrationType = "logo"
	IntegrationMikro         IntegrationType = "mikro"
	IntegrationNetsis        IntegrationType = "netsis"
	IntegrationNetsim        IntegrationType = "netsim"
	IntegrationDia           IntegrationType = "dia"
	IntegrationNethesap      IntegrationType = "nethesap"
	IntegrationZirve         IntegrationType = "zirve"
	IntegrationAkinsoft      IntegrationType = "akinsoft"
	IntegrationVegaYazilim   IntegrationType = "vegayazilim"
	IntegrationNebim         IntegrationType = "nebim"
	IntegrationBarsoftMuhasebe IntegrationType = "barsoftmuhasebe"
	IntegrationSentez        IntegrationType = "sentez"

	// Cargo/Shipping
	IntegrationYurticiKargo  IntegrationType = "yurtici_kargo"
	IntegrationArasKargo     IntegrationType = "aras_kargo"
	IntegrationMNGKargo      IntegrationType = "mng_kargo"
	IntegrationPTTKargo      IntegrationType = "ptt_kargo"
	IntegrationUPS           IntegrationType = "ups"
	IntegrationSuratKargo    IntegrationType = "surat_kargo"
	IntegrationFoodManLojistik IntegrationType = "foodman_lojistik"
	IntegrationCdek          IntegrationType = "cdek"
	IntegrationSendeo        IntegrationType = "sendeo"
	IntegrationPTSKargo      IntegrationType = "pts_kargo"
	IntegrationFedEx         IntegrationType = "fedex"
	IntegrationShipEntegra   IntegrationType = "shipentegra"
	IntegrationDHL           IntegrationType = "dhl"
	IntegrationHepsiJet      IntegrationType = "hepsijet"
)

// SyncStatus represents synchronization status
type SyncStatus string

const (
	SyncStatusPending    SyncStatus = "pending"
	SyncStatusInProgress SyncStatus = "in_progress"
	SyncStatusCompleted  SyncStatus = "completed"
	SyncStatusFailed     SyncStatus = "failed"
	SyncStatusPartial    SyncStatus = "partial"
)

// IntegrationConfig represents platform-specific configuration
type IntegrationConfig struct {
	BaseURL         string            `json:"base_url"`
	APIVersion      string            `json:"api_version"`
	RateLimit       int               `json:"rate_limit"`
	Timeout         int               `json:"timeout"`
	RetryAttempts   int               `json:"retry_attempts"`
	CustomFields    map[string]string `json:"custom_fields"`
	Webhooks        []WebhookConfig   `json:"webhooks"`
	SyncSettings    SyncSettings      `json:"sync_settings"`
}

// WebhookConfig represents webhook configuration
type WebhookConfig struct {
	URL       string            `json:"url"`
	Events    []string          `json:"events"`
	Headers   map[string]string `json:"headers"`
	IsActive  bool              `json:"is_active"`
}

// SyncSettings represents synchronization settings
type SyncSettings struct {
	AutoSync        bool     `json:"auto_sync"`
	SyncInterval    int      `json:"sync_interval"` // minutes
	SyncProducts    bool     `json:"sync_products"`
	SyncOrders      bool     `json:"sync_orders"`
	SyncInventory   bool     `json:"sync_inventory"`
	SyncPrices      bool     `json:"sync_prices"`
	SyncCategories  bool     `json:"sync_categories"`
	FieldMappings   []FieldMapping `json:"field_mappings"`
}

// FieldMapping represents field mapping between systems
type FieldMapping struct {
	LocalField    string `json:"local_field"`
	RemoteField   string `json:"remote_field"`
	Transform     string `json:"transform"`
	DefaultValue  string `json:"default_value"`
}

// SyncLog represents synchronization log
type SyncLog struct {
	ID            int64      `json:"id" db:"id"`
	IntegrationID int64      `json:"integration_id" db:"integration_id"`
	SyncType      string     `json:"sync_type" db:"sync_type"`
	Status        SyncStatus `json:"status" db:"status"`
	RecordsTotal  int        `json:"records_total" db:"records_total"`
	RecordsSuccess int       `json:"records_success" db:"records_success"`
	RecordsFailed int        `json:"records_failed" db:"records_failed"`
	ErrorMessage  string     `json:"error_message" db:"error_message"`
	StartedAt     time.Time  `json:"started_at" db:"started_at"`
	CompletedAt   *time.Time `json:"completed_at" db:"completed_at"`
	Duration      int        `json:"duration" db:"duration"` // seconds
}

// ProductMapping represents product mapping between systems
type ProductMapping struct {
	ID            int64     `json:"id" db:"id"`
	IntegrationID int64     `json:"integration_id" db:"integration_id"`
	LocalProductID int64    `json:"local_product_id" db:"local_product_id"`
	RemoteProductID string  `json:"remote_product_id" db:"remote_product_id"`
	RemoteSKU     string    `json:"remote_sku" db:"remote_sku"`
	SyncStatus    SyncStatus `json:"sync_status" db:"sync_status"`
	LastSynced    *time.Time `json:"last_synced" db:"last_synced"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}