package registry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"kolajAi/internal/integrations"
	"kolajAi/internal/errors"
	"kolajAi/internal/security"
)

// IntegrationRegistry manages all available integrations
type IntegrationRegistry struct {
	integrationDefinitions map[string]*IntegrationDefinition
	providers              map[string]IntegrationProvider
	credentialManager      *security.CredentialManager
	mutex                  sync.RWMutex
}

// IntegrationDefinition defines the structure of an integration
type IntegrationDefinition struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	DisplayName  string                 `json:"display_name"`
	Category     string                 `json:"category"`
	Type         string                 `json:"type"`
	Region       string                 `json:"region"`
	Country      string                 `json:"country"`
	Description  string                 `json:"description"`
	Website      string                 `json:"website"`
	LogoURL      string                 `json:"logo_url"`
	Features     []string               `json:"features"`
	Requirements []string               `json:"requirements"`
	IsActive     bool                   `json:"is_active"`
	IsProduction bool                   `json:"is_production"`
	Priority     int                    `json:"priority"`
	Config       map[string]interface{} `json:"config"`
	Metadata     map[string]string      `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// IntegrationProvider interface for all integration providers
type IntegrationProvider interface {
	Initialize(ctx context.Context, credentials integrations.Credentials, config map[string]interface{}) error
	GetName() string
	GetType() string
	IsHealthy() bool
	GetMetrics() map[string]interface{}
	HealthCheck(ctx context.Context) error
}

// NewIntegrationRegistry creates a new integration registry
func NewIntegrationRegistry(credentialManager *security.CredentialManager) *IntegrationRegistry {
	registry := &IntegrationRegistry{
		integrationDefinitions: make(map[string]*IntegrationDefinition),
		providers:              make(map[string]IntegrationProvider),
		credentialManager:      credentialManager,
	}

	// Initialize all integrations
	registry.initializeAllIntegrations()
	
	return registry
}

// initializeAllIntegrations initializes all 129 integrations
func (r *IntegrationRegistry) initializeAllIntegrations() {
	r.initializeTurkishMarketplaces()
	r.initializeInternationalMarketplaces()
	r.initializeEcommercePlatforms()
	r.initializeSocialMediaIntegrations()
	r.initializeEInvoiceIntegrations()
	r.initializeAccountingERPIntegrations()
	r.initializePreAccountingIntegrations()
	r.initializeCargoIntegrations()
	r.initializeFulfillmentIntegrations()
	r.initializeRetailIntegrations()
}

// initializeTurkishMarketplaces initializes Turkish marketplace integrations (30 total)
func (r *IntegrationRegistry) initializeTurkishMarketplaces() {
	turkishMarketplaces := []*IntegrationDefinition{
		// Major Turkish Marketplaces
		{
			ID: "trendyol", Name: "trendyol", DisplayName: "Trendyol",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Turkey's leading e-commerce platform",
			Website: "https://www.trendyol.com", IsActive: true, IsProduction: true, Priority: 1,
			Features: []string{"product_sync", "order_sync", "inventory_sync", "price_sync", "webhooks"},
		},
		{
			ID: "hepsiburada", Name: "hepsiburada", DisplayName: "Hepsiburada",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Turkey's largest online shopping platform",
			Website: "https://www.hepsiburada.com", IsActive: true, IsProduction: true, Priority: 2,
			Features: []string{"product_sync", "order_sync", "inventory_sync", "variants", "webhooks"},
		},
		{
			ID: "n11", Name: "n11", DisplayName: "N11",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Popular Turkish online marketplace",
			Website: "https://www.n11.com", IsActive: true, IsProduction: true, Priority: 3,
			Features: []string{"product_sync", "order_sync", "inventory_sync", "categories"},
		},
		{
			ID: "amazon_tr", Name: "amazon_tr", DisplayName: "Amazon Türkiye",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Amazon Turkey marketplace",
			Website: "https://www.amazon.com.tr", IsActive: true, IsProduction: true, Priority: 4,
			Features: []string{"product_sync", "order_sync", "fba", "sp_api", "aws_auth"},
		},
		{
			ID: "ciceksepeti", Name: "ciceksepeti", DisplayName: "ÇiçekSepeti",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Leading flower and gift marketplace in Turkey",
			Website: "https://www.ciceksepeti.com", IsActive: true, IsProduction: true, Priority: 5,
			Features: []string{"product_sync", "order_sync", "category_mapping"},
		},
		{
			ID: "sahibinden", Name: "sahibinden", DisplayName: "Sahibinden",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Turkey's largest classified ads platform",
			Website: "https://www.sahibinden.com", IsActive: true, IsProduction: false, Priority: 6,
			Features: []string{"listing_sync", "contact_management", "location_based"},
		},
		{
			ID: "letgo", Name: "letgo", DisplayName: "Letgo",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Mobile marketplace for buying and selling",
			Website: "https://www.letgo.com", IsActive: true, IsProduction: false, Priority: 7,
			Features: []string{"mobile_sync", "image_recognition", "chat_integration"},
		},
		{
			ID: "dolap", Name: "dolap", DisplayName: "Dolap",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Fashion marketplace for second-hand items",
			Website: "https://www.dolap.com", IsActive: true, IsProduction: false, Priority: 8,
			Features: []string{"fashion_sync", "brand_verification", "social_features"},
		},
		{
			ID: "pazarama", Name: "pazarama", DisplayName: "Pazarama",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Turkish online marketplace",
			Website: "https://www.pazarama.com", IsActive: true, IsProduction: false, Priority: 9,
			Features: []string{"product_sync", "order_sync", "inventory_sync"},
		},
		{
			ID: "gittigidiyor", Name: "gittigidiyor", DisplayName: "GittiGidiyor",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "eBay Turkey - Auction and fixed price marketplace",
			Website: "https://www.gittigidiyor.com", IsActive: false, IsProduction: false, Priority: 10,
			Features: []string{"auction_sync", "fixed_price", "paypal_integration"},
		},
		// Additional Turkish Marketplaces
		{
			ID: "modanisa", Name: "modanisa", DisplayName: "Modanisa",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Modest fashion marketplace",
			Website: "https://www.modanisa.com", IsActive: true, IsProduction: false, Priority: 11,
			Features: []string{"fashion_sync", "modest_fashion", "international_shipping"},
		},
		{
			ID: "koton", Name: "koton", DisplayName: "Koton",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Turkish fashion retailer marketplace",
			Website: "https://www.koton.com", IsActive: true, IsProduction: false, Priority: 12,
			Features: []string{"fashion_sync", "seasonal_collections", "size_charts"},
		},
		{
			ID: "lcw", Name: "lcw", DisplayName: "LCW",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "LC Waikiki online marketplace",
			Website: "https://www.lcw.com", IsActive: true, IsProduction: false, Priority: 13,
			Features: []string{"fashion_sync", "family_collections", "affordable_fashion"},
		},
		{
			ID: "defacto", Name: "defacto", DisplayName: "DeFacto",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Turkish fashion brand marketplace",
			Website: "https://www.defacto.com.tr", IsActive: true, IsProduction: false, Priority: 14,
			Features: []string{"fashion_sync", "brand_collections", "trendy_fashion"},
		},
		{
			ID: "boyner", Name: "boyner", DisplayName: "Boyner",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Premium fashion and lifestyle marketplace",
			Website: "https://www.boyner.com.tr", IsActive: true, IsProduction: false, Priority: 15,
			Features: []string{"premium_fashion", "lifestyle_products", "brand_partnerships"},
		},
		{
			ID: "teknosa", Name: "teknosa", DisplayName: "Teknosa",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Electronics and technology marketplace",
			Website: "https://www.teknosa.com", IsActive: true, IsProduction: false, Priority: 16,
			Features: []string{"electronics_sync", "tech_specs", "warranty_management"},
		},
		{
			ID: "mediamarkt", Name: "mediamarkt", DisplayName: "MediaMarkt",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Electronics retail marketplace",
			Website: "https://www.mediamarkt.com.tr", IsActive: true, IsProduction: false, Priority: 17,
			Features: []string{"electronics_sync", "installation_services", "extended_warranty"},
		},
		{
			ID: "vatan", Name: "vatan", DisplayName: "Vatan Bilgisayar",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Computer and technology marketplace",
			Website: "https://www.vatanbilgisayar.com", IsActive: true, IsProduction: false, Priority: 18,
			Features: []string{"tech_sync", "computer_specs", "gaming_products"},
		},
		{
			ID: "kitapyurdu", Name: "kitapyurdu", DisplayName: "Kitapyurdu",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Book marketplace in Turkey",
			Website: "https://www.kitapyurdu.com", IsActive: true, IsProduction: false, Priority: 19,
			Features: []string{"book_sync", "author_management", "isbn_tracking"},
		},
		{
			ID: "dr", Name: "dr", DisplayName: "D&R",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Books, music, and entertainment marketplace",
			Website: "https://www.dr.com.tr", IsActive: true, IsProduction: false, Priority: 20,
			Features: []string{"media_sync", "entertainment_products", "digital_content"},
		},
		{
			ID: "superstep", Name: "superstep", DisplayName: "Superstep",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Sports and lifestyle marketplace",
			Website: "https://www.superstep.com.tr", IsActive: true, IsProduction: false, Priority: 21,
			Features: []string{"sports_sync", "sneaker_collections", "lifestyle_brands"},
		},
		{
			ID: "intersport", Name: "intersport", DisplayName: "Intersport",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Sports equipment marketplace",
			Website: "https://www.intersport.com.tr", IsActive: true, IsProduction: false, Priority: 22,
			Features: []string{"sports_equipment", "fitness_products", "outdoor_gear"},
		},
		{
			ID: "decathlon", Name: "decathlon", DisplayName: "Decathlon",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Sports and outdoor equipment marketplace",
			Website: "https://www.decathlon.com.tr", IsActive: true, IsProduction: false, Priority: 23,
			Features: []string{"outdoor_sports", "equipment_rental", "sports_services"},
		},
		{
			ID: "gratis", Name: "gratis", DisplayName: "Gratis",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Beauty and personal care marketplace",
			Website: "https://www.gratis.com", IsActive: true, IsProduction: false, Priority: 24,
			Features: []string{"beauty_sync", "cosmetics", "personal_care"},
		},
		{
			ID: "sephora", Name: "sephora", DisplayName: "Sephora",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Premium beauty marketplace",
			Website: "https://www.sephora.com.tr", IsActive: true, IsProduction: false, Priority: 25,
			Features: []string{"premium_beauty", "brand_partnerships", "beauty_services"},
		},
		{
			ID: "ebebek", Name: "ebebek", DisplayName: "Ebebek",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Baby and kids products marketplace",
			Website: "https://www.ebebek.com", IsActive: true, IsProduction: false, Priority: 26,
			Features: []string{"baby_products", "kids_fashion", "parenting_essentials"},
		},
		{
			ID: "english_home", Name: "english_home", DisplayName: "English Home",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Home decoration and lifestyle marketplace",
			Website: "https://www.englishhome.com", IsActive: true, IsProduction: false, Priority: 27,
			Features: []string{"home_decor", "furniture", "lifestyle_products"},
		},
		{
			ID: "madame_coco", Name: "madame_coco", DisplayName: "Madame Coco",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Home accessories and decoration marketplace",
			Website: "https://www.madamecoco.com.tr", IsActive: true, IsProduction: false, Priority: 28,
			Features: []string{"home_accessories", "decoration", "gift_items"},
		},
		{
			ID: "koçtaş", Name: "koctas", DisplayName: "Koçtaş",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Home improvement and DIY marketplace",
			Website: "https://www.koctas.com.tr", IsActive: true, IsProduction: false, Priority: 29,
			Features: []string{"diy_products", "home_improvement", "garden_supplies"},
		},
		{
			ID: "bauhaus", Name: "bauhaus", DisplayName: "Bauhaus",
			Category: "marketplace", Type: "turkish", Region: "Turkey", Country: "TR",
			Description: "Construction and home improvement marketplace",
			Website: "https://www.bauhaus.com.tr", IsActive: true, IsProduction: false, Priority: 30,
			Features: []string{"construction_materials", "tools", "professional_services"},
		},
	}

	for _, def := range turkishMarketplaces {
		def.CreatedAt = time.Now()
		def.UpdatedAt = time.Now()
		r.integrationDefinitions[def.ID] = def
	}
}

// initializeInternationalMarketplaces initializes international marketplace integrations (29 total)
func (r *IntegrationRegistry) initializeInternationalMarketplaces() {
	internationalMarketplaces := []*IntegrationDefinition{
		// Major International Marketplaces
		{
			ID: "amazon_us", Name: "amazon_us", DisplayName: "Amazon US",
			Category: "marketplace", Type: "international", Region: "North America", Country: "US",
			Description: "Amazon United States marketplace",
			Website: "https://www.amazon.com", IsActive: true, IsProduction: false, Priority: 1,
			Features: []string{"sp_api", "fba", "advertising", "brand_registry"},
		},
		{
			ID: "amazon_uk", Name: "amazon_uk", DisplayName: "Amazon UK",
			Category: "marketplace", Type: "international", Region: "Europe", Country: "UK",
			Description: "Amazon United Kingdom marketplace",
			Website: "https://www.amazon.co.uk", IsActive: true, IsProduction: false, Priority: 2,
			Features: []string{"sp_api", "fba", "vat_services", "pan_eu"},
		},
		{
			ID: "amazon_de", Name: "amazon_de", DisplayName: "Amazon Germany",
			Category: "marketplace", Type: "international", Region: "Europe", Country: "DE",
			Description: "Amazon Germany marketplace",
			Website: "https://www.amazon.de", IsActive: true, IsProduction: false, Priority: 3,
			Features: []string{"sp_api", "fba", "german_compliance", "european_expansion"},
		},
		{
			ID: "amazon_fr", Name: "amazon_fr", DisplayName: "Amazon France",
			Category: "marketplace", Type: "international", Region: "Europe", Country: "FR",
			Description: "Amazon France marketplace",
			Website: "https://www.amazon.fr", IsActive: true, IsProduction: false, Priority: 4,
			Features: []string{"sp_api", "fba", "french_regulations", "european_shipping"},
		},
		{
			ID: "amazon_it", Name: "amazon_it", DisplayName: "Amazon Italy",
			Category: "marketplace", Type: "international", Region: "Europe", Country: "IT",
			Description: "Amazon Italy marketplace",
			Website: "https://www.amazon.it", IsActive: true, IsProduction: false, Priority: 5,
			Features: []string{"sp_api", "fba", "italian_market", "mediterranean_shipping"},
		},
		{
			ID: "amazon_es", Name: "amazon_es", DisplayName: "Amazon Spain",
			Category: "marketplace", Type: "international", Region: "Europe", Country: "ES",
			Description: "Amazon Spain marketplace",
			Website: "https://www.amazon.es", IsActive: true, IsProduction: false, Priority: 6,
			Features: []string{"sp_api", "fba", "spanish_market", "iberian_logistics"},
		},
		{
			ID: "ebay_us", Name: "ebay_us", DisplayName: "eBay US",
			Category: "marketplace", Type: "international", Region: "North America", Country: "US",
			Description: "eBay United States marketplace",
			Website: "https://www.ebay.com", IsActive: true, IsProduction: false, Priority: 7,
			Features: []string{"auction_format", "buy_it_now", "managed_payments", "global_shipping"},
		},
		{
			ID: "ebay_uk", Name: "ebay_uk", DisplayName: "eBay UK",
			Category: "marketplace", Type: "international", Region: "Europe", Country: "UK",
			Description: "eBay United Kingdom marketplace",
			Website: "https://www.ebay.co.uk", IsActive: true, IsProduction: false, Priority: 8,
			Features: []string{"auction_format", "buy_it_now", "uk_shipping", "brexit_compliance"},
		},
		{
			ID: "ebay_de", Name: "ebay_de", DisplayName: "eBay Germany",
			Category: "marketplace", Type: "international", Region: "Europe", Country: "DE",
			Description: "eBay Germany marketplace",
			Website: "https://www.ebay.de", IsActive: true, IsProduction: false, Priority: 9,
			Features: []string{"auction_format", "buy_it_now", "german_regulations", "eu_shipping"},
		},
		{
			ID: "etsy", Name: "etsy", DisplayName: "Etsy",
			Category: "marketplace", Type: "international", Region: "Global", Country: "US",
			Description: "Handmade and vintage marketplace",
			Website: "https://www.etsy.com", IsActive: true, IsProduction: false, Priority: 10,
			Features: []string{"handmade_products", "vintage_items", "digital_downloads", "pattern_integration"},
		},
		{
			ID: "walmart", Name: "walmart", DisplayName: "Walmart Marketplace",
			Category: "marketplace", Type: "international", Region: "North America", Country: "US",
			Description: "Walmart online marketplace",
			Website: "https://marketplace.walmart.com", IsActive: true, IsProduction: false, Priority: 11,
			Features: []string{"pro_seller", "wfs", "advertising", "grocery_delivery"},
		},
		{
			ID: "aliexpress", Name: "aliexpress", DisplayName: "AliExpress",
			Category: "marketplace", Type: "international", Region: "Asia", Country: "CN",
			Description: "Global online retail marketplace",
			Website: "https://www.aliexpress.com", IsActive: true, IsProduction: false, Priority: 12,
			Features: []string{"dropshipping", "bulk_orders", "buyer_protection", "global_shipping"},
		},
		{
			ID: "alibaba", Name: "alibaba", DisplayName: "Alibaba",
			Category: "marketplace", Type: "international", Region: "Asia", Country: "CN",
			Description: "B2B wholesale marketplace",
			Website: "https://www.alibaba.com", IsActive: true, IsProduction: false, Priority: 13,
			Features: []string{"b2b_wholesale", "trade_assurance", "supplier_verification", "bulk_pricing"},
		},
		{
			ID: "shopee_sg", Name: "shopee_sg", DisplayName: "Shopee Singapore",
			Category: "marketplace", Type: "international", Region: "Southeast Asia", Country: "SG",
			Description: "Leading e-commerce platform in Southeast Asia",
			Website: "https://shopee.sg", IsActive: true, IsProduction: false, Priority: 14,
			Features: []string{"social_commerce", "live_streaming", "games", "sea_logistics"},
		},
		{
			ID: "lazada", Name: "lazada", DisplayName: "Lazada",
			Category: "marketplace", Type: "international", Region: "Southeast Asia", Country: "SG",
			Description: "Southeast Asia's leading online shopping platform",
			Website: "https://www.lazada.com", IsActive: true, IsProduction: false, Priority: 15,
			Features: []string{"cross_border", "flash_sales", "live_streaming", "logistics_solutions"},
		},
		{
			ID: "rakuten", Name: "rakuten", DisplayName: "Rakuten",
			Category: "marketplace", Type: "international", Region: "Asia", Country: "JP",
			Description: "Japan's largest e-commerce marketplace",
			Website: "https://www.rakuten.com", IsActive: true, IsProduction: false, Priority: 16,
			Features: []string{"loyalty_points", "ichiba", "books", "travel_services"},
		},
		{
			ID: "mercadolibre", Name: "mercadolibre", DisplayName: "MercadoLibre",
			Category: "marketplace", Type: "international", Region: "Latin America", Country: "AR",
			Description: "Latin America's leading e-commerce platform",
			Website: "https://www.mercadolibre.com", IsActive: true, IsProduction: false, Priority: 17,
			Features: []string{"mercado_pago", "mercado_envios", "classified_ads", "real_estate"},
		},
		{
			ID: "flipkart", Name: "flipkart", DisplayName: "Flipkart",
			Category: "marketplace", Type: "international", Region: "Asia", Country: "IN",
			Description: "India's leading e-commerce marketplace",
			Website: "https://www.flipkart.com", IsActive: true, IsProduction: false, Priority: 18,
			Features: []string{"big_billion_days", "flipkart_assured", "grocery", "fashion"},
		},
		{
			ID: "amazon_in", Name: "amazon_in", DisplayName: "Amazon India",
			Category: "marketplace", Type: "international", Region: "Asia", Country: "IN",
			Description: "Amazon India marketplace",
			Website: "https://www.amazon.in", IsActive: true, IsProduction: false, Priority: 19,
			Features: []string{"sp_api", "easy_ship", "amazon_pay", "prime_delivery"},
		},
		{
			ID: "jd", Name: "jd", DisplayName: "JD.com",
			Category: "marketplace", Type: "international", Region: "Asia", Country: "CN",
			Description: "Chinese e-commerce platform",
			Website: "https://www.jd.com", IsActive: true, IsProduction: false, Priority: 20,
			Features: []string{"jd_logistics", "jd_finance", "fresh_products", "electronics"},
		},
		{
			ID: "tmall", Name: "tmall", DisplayName: "Tmall",
			Category: "marketplace", Type: "international", Region: "Asia", Country: "CN",
			Description: "B2C platform operated by Alibaba Group",
			Website: "https://www.tmall.com", IsActive: true, IsProduction: false, Priority: 21,
			Features: []string{"brand_stores", "luxury_pavilion", "global_import", "singles_day"},
		},
		{
			ID: "cdiscount", Name: "cdiscount", DisplayName: "Cdiscount",
			Category: "marketplace", Type: "international", Region: "Europe", Country: "FR",
			Description: "French e-commerce marketplace",
			Website: "https://www.cdiscount.com", IsActive: true, IsProduction: false, Priority: 22,
			Features: []string{"marketplace", "fulfilment", "advertising", "mobile_services"},
		},
		{
			ID: "bol", Name: "bol", DisplayName: "Bol.com",
			Category: "marketplace", Type: "international", Region: "Europe", Country: "NL",
			Description: "Dutch online marketplace",
			Website: "https://www.bol.com", IsActive: true, IsProduction: false, Priority: 23,
			Features: []string{"plaza", "fulfillment", "advertising", "subscription_services"},
		},
		{
			ID: "zalando", Name: "zalando", DisplayName: "Zalando",
			Category: "marketplace", Type: "international", Region: "Europe", Country: "DE",
			Description: "European fashion and lifestyle platform",
			Website: "https://www.zalando.com", IsActive: true, IsProduction: false, Priority: 24,
			Features: []string{"fashion_store", "connected_retail", "logistics", "advertising"},
		},
		{
			ID: "otto", Name: "otto", DisplayName: "OTTO",
			Category: "marketplace", Type: "international", Region: "Europe", Country: "DE",
			Description: "German online marketplace",
			Website: "https://www.otto.de", IsActive: true, IsProduction: false, Priority: 25,
			Features: []string{"marketplace", "fulfillment", "fashion", "home_living"},
		},
		{
			ID: "real", Name: "real", DisplayName: "Real.de",
			Category: "marketplace", Type: "international", Region: "Europe", Country: "DE",
			Description: "German online marketplace",
			Website: "https://www.real.de", IsActive: true, IsProduction: false, Priority: 26,
			Features: []string{"marketplace", "grocery", "electronics", "fulfillment"},
		},
		{
			ID: "allegro", Name: "allegro", DisplayName: "Allegro",
			Category: "marketplace", Type: "international", Region: "Europe", Country: "PL",
			Description: "Poland's largest e-commerce platform",
			Website: "https://www.allegro.pl", IsActive: true, IsProduction: false, Priority: 27,
			Features: []string{"one_fulfillment", "smart", "advertising", "allegro_pay"},
		},
		{
			ID: "emag", Name: "emag", DisplayName: "eMAG",
			Category: "marketplace", Type: "international", Region: "Europe", Country: "RO",
			Description: "Leading e-commerce platform in Eastern Europe",
			Website: "https://www.emag.ro", IsActive: true, IsProduction: false, Priority: 28,
			Features: []string{"marketplace", "genius", "easy_box", "showroom"},
		},
		{
			ID: "ozon", Name: "ozon", DisplayName: "Ozon",
			Category: "marketplace", Type: "international", Region: "Europe", Country: "RU",
			Description: "Russian e-commerce marketplace",
			Website: "https://www.ozon.ru", IsActive: true, IsProduction: false, Priority: 29,
			Features: []string{"fulfillment", "express_delivery", "premium", "travel_services"},
		},
	}

	for _, def := range internationalMarketplaces {
		def.CreatedAt = time.Now()
		def.UpdatedAt = time.Now()
		r.integrationDefinitions[def.ID] = def
	}
}

// initializeEcommercePlatforms initializes e-commerce platform integrations (12 total)
func (r *IntegrationRegistry) initializeEcommercePlatforms() {
	ecommercePlatforms := []*IntegrationDefinition{
		{
			ID: "shopify", Name: "shopify", DisplayName: "Shopify",
			Category: "ecommerce_platform", Type: "saas", Region: "Global", Country: "CA",
			Description: "Leading e-commerce platform",
			Website: "https://www.shopify.com", IsActive: true, IsProduction: true, Priority: 1,
			Features: []string{"store_sync", "product_sync", "order_sync", "inventory_sync", "webhooks"},
		},
		{
			ID: "woocommerce", Name: "woocommerce", DisplayName: "WooCommerce",
			Category: "ecommerce_platform", Type: "wordpress", Region: "Global", Country: "US",
			Description: "WordPress e-commerce plugin",
			Website: "https://woocommerce.com", IsActive: true, IsProduction: true, Priority: 2,
			Features: []string{"rest_api", "webhooks", "extensions", "payment_gateways"},
		},
		{
			ID: "magento", Name: "magento", DisplayName: "Magento",
			Category: "ecommerce_platform", Type: "open_source", Region: "Global", Country: "US",
			Description: "Open-source e-commerce platform",
			Website: "https://magento.com", IsActive: true, IsProduction: true, Priority: 3,
			Features: []string{"rest_api", "graphql", "multi_store", "b2b_features"},
		},
		{
			ID: "opencart", Name: "opencart", DisplayName: "OpenCart",
			Category: "ecommerce_platform", Type: "open_source", Region: "Global", Country: "HK",
			Description: "Free open source e-commerce platform",
			Website: "https://www.opencart.com", IsActive: true, IsProduction: false, Priority: 4,
			Features: []string{"rest_api", "multi_store", "extensions", "themes"},
		},
		{
			ID: "prestashop", Name: "prestashop", DisplayName: "PrestaShop",
			Category: "ecommerce_platform", Type: "open_source", Region: "Europe", Country: "FR",
			Description: "Open source e-commerce solution",
			Website: "https://www.prestashop.com", IsActive: true, IsProduction: false, Priority: 5,
			Features: []string{"webservice_api", "modules", "themes", "multi_shop"},
		},
		{
			ID: "bigcommerce", Name: "bigcommerce", DisplayName: "BigCommerce",
			Category: "ecommerce_platform", Type: "saas", Region: "Global", Country: "US",
			Description: "SaaS e-commerce platform",
			Website: "https://www.bigcommerce.com", IsActive: true, IsProduction: false, Priority: 6,
			Features: []string{"rest_api", "storefront_api", "webhooks", "headless_commerce"},
		},
		{
			ID: "squarespace", Name: "squarespace", DisplayName: "Squarespace Commerce",
			Category: "ecommerce_platform", Type: "saas", Region: "Global", Country: "US",
			Description: "Website builder with e-commerce",
			Website: "https://www.squarespace.com", IsActive: true, IsProduction: false, Priority: 7,
			Features: []string{"commerce_api", "inventory_management", "order_management"},
		},
		{
			ID: "wix", Name: "wix", DisplayName: "Wix Stores",
			Category: "ecommerce_platform", Type: "saas", Region: "Global", Country: "IL",
			Description: "Website builder with e-commerce functionality",
			Website: "https://www.wix.com", IsActive: true, IsProduction: false, Priority: 8,
			Features: []string{"stores_api", "payment_processing", "shipping_integration"},
		},
		{
			ID: "volusion", Name: "volusion", DisplayName: "Volusion",
			Category: "ecommerce_platform", Type: "saas", Region: "North America", Country: "US",
			Description: "E-commerce platform for small businesses",
			Website: "https://www.volusion.com", IsActive: true, IsProduction: false, Priority: 9,
			Features: []string{"api_integration", "inventory_sync", "order_management"},
		},
		{
			ID: "3dcart", Name: "3dcart", DisplayName: "Shift4Shop",
			Category: "ecommerce_platform", Type: "saas", Region: "North America", Country: "US",
			Description: "Feature-rich e-commerce platform",
			Website: "https://www.shift4shop.com", IsActive: true, IsProduction: false, Priority: 10,
			Features: []string{"rest_api", "webhooks", "advanced_features", "seo_tools"},
		},
		{
			ID: "ecwid", Name: "ecwid", DisplayName: "Ecwid",
			Category: "ecommerce_platform", Type: "saas", Region: "Global", Country: "US",
			Description: "E-commerce platform for existing websites",
			Website: "https://www.ecwid.com", IsActive: true, IsProduction: false, Priority: 11,
			Features: []string{"rest_api", "instant_site", "social_selling", "mobile_responsive"},
		},
		{
			ID: "lightspeed", Name: "lightspeed", DisplayName: "Lightspeed eCom",
			Category: "ecommerce_platform", Type: "saas", Region: "Global", Country: "CA",
			Description: "E-commerce platform with POS integration",
			Website: "https://www.lightspeedhq.com", IsActive: true, IsProduction: false, Priority: 12,
			Features: []string{"rest_api", "pos_integration", "inventory_management", "omnichannel"},
		},
	}

	for _, def := range ecommercePlatforms {
		def.CreatedAt = time.Now()
		def.UpdatedAt = time.Now()
		r.integrationDefinitions[def.ID] = def
	}
}

// initializeSocialMediaIntegrations initializes social media integrations (3 total)
func (r *IntegrationRegistry) initializeSocialMediaIntegrations() {
	socialMediaIntegrations := []*IntegrationDefinition{
		{
			ID: "facebook_shop", Name: "facebook_shop", DisplayName: "Facebook Shop",
			Category: "social_media", Type: "social_commerce", Region: "Global", Country: "US",
			Description: "Facebook shopping integration",
			Website: "https://www.facebook.com/business/shops", IsActive: true, IsProduction: true, Priority: 1,
			Features: []string{"catalog_sync", "dynamic_ads", "pixel_tracking", "messenger_integration"},
		},
		{
			ID: "instagram_shopping", Name: "instagram_shopping", DisplayName: "Instagram Shopping",
			Category: "social_media", Type: "social_commerce", Region: "Global", Country: "US",
			Description: "Instagram shopping integration",
			Website: "https://business.instagram.com/shopping", IsActive: true, IsProduction: true, Priority: 2,
			Features: []string{"product_tags", "shopping_ads", "stories_shopping", "reels_shopping"},
		},
		{
			ID: "google_shopping", Name: "google_shopping", DisplayName: "Google Shopping",
			Category: "social_media", Type: "advertising", Region: "Global", Country: "US",
			Description: "Google Shopping and Merchant Center integration",
			Website: "https://www.google.com/retail/shopping", IsActive: true, IsProduction: true, Priority: 3,
			Features: []string{"merchant_center", "shopping_ads", "free_listings", "local_inventory"},
		},
	}

	for _, def := range socialMediaIntegrations {
		def.CreatedAt = time.Now()
		def.UpdatedAt = time.Now()
		r.integrationDefinitions[def.ID] = def
	}
}

// initializeEInvoiceIntegrations initializes e-invoice integrations (15 total)
func (r *IntegrationRegistry) initializeEInvoiceIntegrations() {
	eInvoiceIntegrations := []*IntegrationDefinition{
		{
			ID: "gib_einvoice", Name: "gib_einvoice", DisplayName: "GİB E-Fatura",
			Category: "einvoice", Type: "government", Region: "Turkey", Country: "TR",
			Description: "Turkish Revenue Administration e-invoice system",
			Website: "https://www.gib.gov.tr", IsActive: true, IsProduction: true, Priority: 1,
			Features: []string{"ubl_format", "digital_signature", "archive", "integration_test"},
		},
		{
			ID: "logo_einvoice", Name: "logo_einvoice", DisplayName: "Logo E-Fatura",
			Category: "einvoice", Type: "service_provider", Region: "Turkey", Country: "TR",
			Description: "Logo e-invoice service provider",
			Website: "https://www.logo.com.tr", IsActive: true, IsProduction: true, Priority: 2,
			Features: []string{"api_integration", "bulk_processing", "reporting", "compliance"},
		},
		{
			ID: "uyumsoft_einvoice", Name: "uyumsoft_einvoice", DisplayName: "UyumSoft E-Fatura",
			Category: "einvoice", Type: "service_provider", Region: "Turkey", Country: "TR",
			Description: "UyumSoft e-invoice service provider",
			Website: "https://www.uyumsoft.com.tr", IsActive: true, IsProduction: false, Priority: 3,
			Features: []string{"web_service", "xml_processing", "validation", "archiving"},
		},
		{
			ID: "elogo_einvoice", Name: "elogo_einvoice", DisplayName: "E-Logo E-Fatura",
			Category: "einvoice", Type: "service_provider", Region: "Turkey", Country: "TR",
			Description: "E-Logo e-invoice service provider",
			Website: "https://www.elogo.com.tr", IsActive: true, IsProduction: false, Priority: 4,
			Features: []string{"cloud_service", "mobile_app", "integration", "support"},
		},
		{
			ID: "foriba_einvoice", Name: "foriba_einvoice", DisplayName: "Foriba E-Fatura",
			Category: "einvoice", Type: "service_provider", Region: "Turkey", Country: "TR",
			Description: "Foriba e-invoice and e-document solutions",
			Website: "https://www.foriba.com", IsActive: true, IsProduction: false, Priority: 5,
			Features: []string{"comprehensive_api", "multi_country", "compliance", "analytics"},
		},
		{
			ID: "ziraat_einvoice", Name: "ziraat_einvoice", DisplayName: "Ziraat E-Fatura",
			Category: "einvoice", Type: "service_provider", Region: "Turkey", Country: "TR",
			Description: "Ziraat Teknoloji e-invoice services",
			Website: "https://www.ziraatteknoloji.com", IsActive: true, IsProduction: false, Priority: 6,
			Features: []string{"banking_integration", "secure_processing", "compliance", "reporting"},
		},
		{
			ID: "turkiye_finans_einvoice", Name: "turkiye_finans_einvoice", DisplayName: "Türkiye Finans E-Fatura",
			Category: "einvoice", Type: "service_provider", Region: "Turkey", Country: "TR",
			Description: "Türkiye Finans e-invoice services",
			Website: "https://www.turkiyefinans.com.tr", IsActive: true, IsProduction: false, Priority: 7,
			Features: []string{"islamic_finance", "compliance", "integration", "support"},
		},
		{
			ID: "parasoft_einvoice", Name: "parasoft_einvoice", DisplayName: "Parasoft E-Fatura",
			Category: "einvoice", Type: "service_provider", Region: "Turkey", Country: "TR",
			Description: "Parasoft e-invoice solutions",
			Website: "https://www.parasoft.com.tr", IsActive: true, IsProduction: false, Priority: 8,
			Features: []string{"software_solutions", "integration", "customization", "training"},
		},
		{
			ID: "innova_einvoice", Name: "innova_einvoice", DisplayName: "İnnova E-Fatura",
			Category: "einvoice", Type: "service_provider", Region: "Turkey", Country: "TR",
			Description: "İnnova e-invoice and digital transformation",
			Website: "https://www.innova.com.tr", IsActive: true, IsProduction: false, Priority: 9,
			Features: []string{"digital_transformation", "cloud_solutions", "integration", "consulting"},
		},
		{
			ID: "netsis_einvoice", Name: "netsis_einvoice", DisplayName: "Netsis E-Fatura",
			Category: "einvoice", Type: "service_provider", Region: "Turkey", Country: "TR",
			Description: "Netsis ERP integrated e-invoice",
			Website: "https://www.netsis.com.tr", IsActive: true, IsProduction: false, Priority: 10,
			Features: []string{"erp_integration", "workflow", "approval", "reporting"},
		},
		{
			ID: "mikro_einvoice", Name: "mikro_einvoice", DisplayName: "Mikro E-Fatura",
			Category: "einvoice", Type: "service_provider", Region: "Turkey", Country: "TR",
			Description: "Mikro ERP e-invoice integration",
			Website: "https://www.mikro.com.tr", IsActive: true, IsProduction: false, Priority: 11,
			Features: []string{"erp_native", "automation", "compliance", "reporting"},
		},
		{
			ID: "eta_einvoice", Name: "eta_einvoice", DisplayName: "ETA E-Fatura",
			Category: "einvoice", Type: "service_provider", Region: "Turkey", Country: "TR",
			Description: "ETA e-invoice service provider",
			Website: "https://www.eta.com.tr", IsActive: true, IsProduction: false, Priority: 12,
			Features: []string{"api_service", "bulk_processing", "validation", "archiving"},
		},
		{
			ID: "turkcell_einvoice", Name: "turkcell_einvoice", DisplayName: "Turkcell E-Fatura",
			Category: "einvoice", Type: "service_provider", Region: "Turkey", Country: "TR",
			Description: "Turkcell digital business e-invoice",
			Website: "https://www.turkcell.com.tr", IsActive: true, IsProduction: false, Priority: 13,
			Features: []string{"telecom_integration", "mobile_solutions", "cloud_service", "support"},
		},
		{
			ID: "vodafone_einvoice", Name: "vodafone_einvoice", DisplayName: "Vodafone E-Fatura",
			Category: "einvoice", Type: "service_provider", Region: "Turkey", Country: "TR",
			Description: "Vodafone business e-invoice services",
			Website: "https://www.vodafone.com.tr", IsActive: true, IsProduction: false, Priority: 14,
			Features: []string{"business_solutions", "integration", "mobile_access", "reporting"},
		},
		{
			ID: "avea_einvoice", Name: "avea_einvoice", DisplayName: "Avea E-Fatura",
			Category: "einvoice", Type: "service_provider", Region: "Turkey", Country: "TR",
			Description: "Avea (Türk Telekom) e-invoice services",
			Website: "https://www.turktelekom.com.tr", IsActive: true, IsProduction: false, Priority: 15,
			Features: []string{"telekom_integration", "enterprise_solutions", "api_access", "support"},
		},
	}

	for _, def := range eInvoiceIntegrations {
		def.CreatedAt = time.Now()
		def.UpdatedAt = time.Now()
		r.integrationDefinitions[def.ID] = def
	}
}

// Continue with other integration types...
// (I'll implement the rest in the next parts due to length constraints)

// GetIntegration returns a specific integration definition
func (r *IntegrationRegistry) GetIntegration(id string) (*IntegrationDefinition, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	integration, exists := r.integrationDefinitions[id]
	return integration, exists
}

// ListIntegrations returns all integration definitions
func (r *IntegrationRegistry) ListIntegrations() []*IntegrationDefinition {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	integrations := make([]*IntegrationDefinition, 0, len(r.integrationDefinitions))
	for _, integration := range r.integrationDefinitions {
		integrations = append(integrations, integration)
	}
	
	return integrations
}

// ListIntegrationsByCategory returns integrations filtered by category
func (r *IntegrationRegistry) ListIntegrationsByCategory(category string) []*IntegrationDefinition {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var integrations []*IntegrationDefinition
	for _, integration := range r.integrationDefinitions {
		if integration.Category == category {
			integrations = append(integrations, integration)
		}
	}
	
	return integrations
}

// GetIntegrationCount returns the total number of integrations
func (r *IntegrationRegistry) GetIntegrationCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	return len(r.integrationDefinitions)
}

// GetIntegrationCountByCategory returns count of integrations by category
func (r *IntegrationRegistry) GetIntegrationCountByCategory() map[string]int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	counts := make(map[string]int)
	for _, integration := range r.integrationDefinitions {
		counts[integration.Category]++
	}
	
	return counts
}

// RegisterProvider registers a provider for an integration
func (r *IntegrationRegistry) RegisterProvider(integrationID string, provider IntegrationProvider) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.integrationDefinitions[integrationID]; !exists {
		return fmt.Errorf("integration %s not found", integrationID)
	}
	
	r.providers[integrationID] = provider
	return nil
}

// GetProvider returns a provider for an integration
func (r *IntegrationRegistry) GetProvider(integrationID string) (IntegrationProvider, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	provider, exists := r.providers[integrationID]
	return provider, exists
}

// GetSummary returns a summary of all integrations
func (r *IntegrationRegistry) GetSummary() map[string]interface{} {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	summary := map[string]interface{}{
		"total_integrations": len(r.integrationDefinitions),
		"by_category":        r.GetIntegrationCountByCategory(),
		"active_count":       0,
		"production_ready":   0,
		"regions":            make(map[string]int),
		"types":              make(map[string]int),
	}
	
	regions := make(map[string]int)
	types := make(map[string]int)
	activeCount := 0
	productionReady := 0
	
	for _, integration := range r.integrationDefinitions {
		if integration.IsActive {
			activeCount++
		}
		if integration.IsProduction {
			productionReady++
		}
		regions[integration.Region]++
		types[integration.Type]++
	}
	
	summary["active_count"] = activeCount
	summary["production_ready"] = productionReady
	summary["regions"] = regions
	summary["types"] = types
	
	return summary
}

// initializeAccountingERPIntegrations initializes accounting/ERP integrations (12 total)
func (r *IntegrationRegistry) initializeAccountingERPIntegrations() {
	accountingERPIntegrations := []*IntegrationDefinition{
		{
			ID: "logo_erp", Name: "logo_erp", DisplayName: "Logo ERP",
			Category: "accounting_erp", Type: "erp", Region: "Turkey", Country: "TR",
			Description: "Logo Tiger ERP system integration",
			Website: "https://www.logo.com.tr", IsActive: true, IsProduction: true, Priority: 1,
			Features: []string{"financial_sync", "inventory_management", "customer_management", "reporting"},
		},
		{
			ID: "sap", Name: "sap", DisplayName: "SAP ERP",
			Category: "accounting_erp", Type: "erp", Region: "Global", Country: "DE",
			Description: "SAP enterprise resource planning integration",
			Website: "https://www.sap.com", IsActive: true, IsProduction: true, Priority: 2,
			Features: []string{"enterprise_integration", "financial_modules", "supply_chain", "analytics"},
		},
		{
			ID: "oracle_erp", Name: "oracle_erp", DisplayName: "Oracle ERP Cloud",
			Category: "accounting_erp", Type: "erp", Region: "Global", Country: "US",
			Description: "Oracle Cloud ERP integration",
			Website: "https://www.oracle.com", IsActive: true, IsProduction: true, Priority: 3,
			Features: []string{"cloud_erp", "financial_management", "procurement", "project_management"},
		},
		{
			ID: "microsoft_dynamics", Name: "microsoft_dynamics", DisplayName: "Microsoft Dynamics 365",
			Category: "accounting_erp", Type: "erp", Region: "Global", Country: "US",
			Description: "Microsoft Dynamics 365 Business Central integration",
			Website: "https://dynamics.microsoft.com", IsActive: true, IsProduction: false, Priority: 4,
			Features: []string{"business_central", "financial_management", "sales", "service"},
		},
		{
			ID: "netsis_erp", Name: "netsis_erp", DisplayName: "Netsis ERP",
			Category: "accounting_erp", Type: "erp", Region: "Turkey", Country: "TR",
			Description: "Netsis enterprise resource planning system",
			Website: "https://www.netsis.com.tr", IsActive: true, IsProduction: false, Priority: 5,
			Features: []string{"turkish_erp", "manufacturing", "distribution", "retail"},
		},
		{
			ID: "mikro_erp", Name: "mikro_erp", DisplayName: "Mikro ERP",
			Category: "accounting_erp", Type: "erp", Region: "Turkey", Country: "TR",
			Description: "Mikro enterprise business solutions",
			Website: "https://www.mikro.com.tr", IsActive: true, IsProduction: false, Priority: 6,
			Features: []string{"business_solutions", "financial_management", "crm", "hr"},
		},
		{
			ID: "eta_erp", Name: "eta_erp", DisplayName: "ETA ERP",
			Category: "accounting_erp", Type: "erp", Region: "Turkey", Country: "TR",
			Description: "ETA enterprise resource planning",
			Website: "https://www.eta.com.tr", IsActive: true, IsProduction: false, Priority: 7,
			Features: []string{"manufacturing_erp", "quality_management", "maintenance", "reporting"},
		},
		{
			ID: "quickbooks", Name: "quickbooks", DisplayName: "QuickBooks",
			Category: "accounting_erp", Type: "accounting", Region: "Global", Country: "US",
			Description: "QuickBooks accounting software integration",
			Website: "https://quickbooks.intuit.com", IsActive: true, IsProduction: false, Priority: 8,
			Features: []string{"small_business", "invoicing", "expense_tracking", "payroll"},
		},
		{
			ID: "xero", Name: "xero", DisplayName: "Xero",
			Category: "accounting_erp", Type: "accounting", Region: "Global", Country: "NZ",
			Description: "Xero cloud accounting software",
			Website: "https://www.xero.com", IsActive: true, IsProduction: false, Priority: 9,
			Features: []string{"cloud_accounting", "bank_reconciliation", "invoicing", "reporting"},
		},
		{
			ID: "sage", Name: "sage", DisplayName: "Sage",
			Category: "accounting_erp", Type: "accounting", Region: "Global", Country: "UK",
			Description: "Sage accounting and business management",
			Website: "https://www.sage.com", IsActive: true, IsProduction: false, Priority: 10,
			Features: []string{"business_management", "payroll", "hr", "accounting"},
		},
		{
			ID: "freshbooks", Name: "freshbooks", DisplayName: "FreshBooks",
			Category: "accounting_erp", Type: "accounting", Region: "Global", Country: "CA",
			Description: "FreshBooks cloud accounting software",
			Website: "https://www.freshbooks.com", IsActive: true, IsProduction: false, Priority: 11,
			Features: []string{"invoicing", "time_tracking", "expense_management", "reporting"},
		},
		{
			ID: "wave", Name: "wave", DisplayName: "Wave Accounting",
			Category: "accounting_erp", Type: "accounting", Region: "Global", Country: "CA",
			Description: "Free accounting software for small businesses",
			Website: "https://www.waveapps.com", IsActive: true, IsProduction: false, Priority: 12,
			Features: []string{"free_accounting", "invoicing", "payments", "payroll"},
		},
	}

	for _, def := range accountingERPIntegrations {
		def.CreatedAt = time.Now()
		def.UpdatedAt = time.Now()
		r.integrationDefinitions[def.ID] = def
	}
}

// initializePreAccountingIntegrations initializes pre-accounting integrations (5 total)
func (r *IntegrationRegistry) initializePreAccountingIntegrations() {
	preAccountingIntegrations := []*IntegrationDefinition{
		{
			ID: "parasoft_preaccounting", Name: "parasoft_preaccounting", DisplayName: "Parasoft Ön Muhasebe",
			Category: "pre_accounting", Type: "pre_accounting", Region: "Turkey", Country: "TR",
			Description: "Parasoft pre-accounting solutions",
			Website: "https://www.parasoft.com.tr", IsActive: true, IsProduction: true, Priority: 1,
			Features: []string{"pre_accounting", "document_management", "workflow", "integration"},
		},
		{
			ID: "logo_preaccounting", Name: "logo_preaccounting", DisplayName: "Logo Ön Muhasebe",
			Category: "pre_accounting", Type: "pre_accounting", Region: "Turkey", Country: "TR",
			Description: "Logo pre-accounting module",
			Website: "https://www.logo.com.tr", IsActive: true, IsProduction: true, Priority: 2,
			Features: []string{"pre_accounting", "document_processing", "approval_workflow", "reporting"},
		},
		{
			ID: "eta_preaccounting", Name: "eta_preaccounting", DisplayName: "ETA Ön Muhasebe",
			Category: "pre_accounting", Type: "pre_accounting", Region: "Turkey", Country: "TR",
			Description: "ETA pre-accounting solutions",
			Website: "https://www.eta.com.tr", IsActive: true, IsProduction: false, Priority: 3,
			Features: []string{"document_workflow", "approval_process", "integration", "reporting"},
		},
		{
			ID: "mikro_preaccounting", Name: "mikro_preaccounting", DisplayName: "Mikro Ön Muhasebe",
			Category: "pre_accounting", Type: "pre_accounting", Region: "Turkey", Country: "TR",
			Description: "Mikro pre-accounting module",
			Website: "https://www.mikro.com.tr", IsActive: true, IsProduction: false, Priority: 4,
			Features: []string{"pre_accounting", "document_management", "workflow", "erp_integration"},
		},
		{
			ID: "netsis_preaccounting", Name: "netsis_preaccounting", DisplayName: "Netsis Ön Muhasebe",
			Category: "pre_accounting", Type: "pre_accounting", Region: "Turkey", Country: "TR",
			Description: "Netsis pre-accounting solutions",
			Website: "https://www.netsis.com.tr", IsActive: true, IsProduction: false, Priority: 5,
			Features: []string{"pre_accounting", "document_processing", "approval", "erp_sync"},
		},
	}

	for _, def := range preAccountingIntegrations {
		def.CreatedAt = time.Now()
		def.UpdatedAt = time.Now()
		r.integrationDefinitions[def.ID] = def
	}
}

// initializeCargoIntegrations initializes cargo integrations (17 total)
func (r *IntegrationRegistry) initializeCargoIntegrations() {
	cargoIntegrations := []*IntegrationDefinition{
		{
			ID: "yurtici_kargo", Name: "yurtici_kargo", DisplayName: "Yurtiçi Kargo",
			Category: "cargo", Type: "cargo_company", Region: "Turkey", Country: "TR",
			Description: "Turkey's leading cargo company",
			Website: "https://www.yurticikargo.com", IsActive: true, IsProduction: true, Priority: 1,
			Features: []string{"shipment_tracking", "label_printing", "pickup_scheduling", "delivery_notifications"},
		},
		{
			ID: "mng_kargo", Name: "mng_kargo", DisplayName: "MNG Kargo",
			Category: "cargo", Type: "cargo_company", Region: "Turkey", Country: "TR",
			Description: "MNG cargo and logistics services",
			Website: "https://www.mngkargo.com.tr", IsActive: true, IsProduction: true, Priority: 2,
			Features: []string{"cargo_tracking", "express_delivery", "international_shipping", "e_commerce_solutions"},
		},
		{
			ID: "aras_kargo", Name: "aras_kargo", DisplayName: "Aras Kargo",
			Category: "cargo", Type: "cargo_company", Region: "Turkey", Country: "TR",
			Description: "Aras cargo and express delivery",
			Website: "https://www.araskargo.com.tr", IsActive: true, IsProduction: true, Priority: 3,
			Features: []string{"express_delivery", "cargo_tracking", "same_day_delivery", "international_service"},
		},
		{
			ID: "ptt_kargo", Name: "ptt_kargo", DisplayName: "PTT Kargo",
			Category: "cargo", Type: "cargo_company", Region: "Turkey", Country: "TR",
			Description: "Turkish Post cargo services",
			Website: "https://www.ptt.gov.tr", IsActive: true, IsProduction: true, Priority: 4,
			Features: []string{"postal_services", "cargo_delivery", "government_integration", "nationwide_coverage"},
		},
		{
			ID: "ups_kargo", Name: "ups_kargo", DisplayName: "UPS Kargo",
			Category: "cargo", Type: "cargo_company", Region: "Turkey", Country: "US",
			Description: "UPS Turkey cargo services",
			Website: "https://www.ups.com/tr", IsActive: true, IsProduction: true, Priority: 5,
			Features: []string{"international_express", "supply_chain", "logistics_solutions", "tracking"},
		},
		{
			ID: "dhl_kargo", Name: "dhl_kargo", DisplayName: "DHL Kargo",
			Category: "cargo", Type: "cargo_company", Region: "Turkey", Country: "DE",
			Description: "DHL Turkey express delivery",
			Website: "https://www.dhl.com.tr", IsActive: true, IsProduction: false, Priority: 6,
			Features: []string{"express_worldwide", "supply_chain", "e_commerce", "tracking"},
		},
		{
			ID: "fedex_kargo", Name: "fedex_kargo", DisplayName: "FedEx Kargo",
			Category: "cargo", Type: "cargo_company", Region: "Turkey", Country: "US",
			Description: "FedEx Turkey express services",
			Website: "https://www.fedex.com/tr", IsActive: true, IsProduction: false, Priority: 7,
			Features: []string{"express_delivery", "international_shipping", "supply_chain", "tracking"},
		},
		{
			ID: "tnt_kargo", Name: "tnt_kargo", DisplayName: "TNT Kargo",
			Category: "cargo", Type: "cargo_company", Region: "Turkey", Country: "NL",
			Description: "TNT Turkey express delivery",
			Website: "https://www.tnt.com/express/tr_tr/site_home.html", IsActive: true, IsProduction: false, Priority: 8,
			Features: []string{"express_delivery", "road_network", "air_express", "tracking"},
		},
		{
			ID: "kargo_turk", Name: "kargo_turk", DisplayName: "Kargo Türk",
			Category: "cargo", Type: "cargo_company", Region: "Turkey", Country: "TR",
			Description: "Turkish cargo and logistics company",
			Website: "https://www.kargoturk.com", IsActive: true, IsProduction: false, Priority: 9,
			Features: []string{"domestic_cargo", "express_delivery", "logistics", "tracking"},
		},
		{
			ID: "sendeo", Name: "sendeo", DisplayName: "Sendeo",
			Category: "cargo", Type: "cargo_company", Region: "Turkey", Country: "TR",
			Description: "Digital cargo and logistics platform",
			Website: "https://www.sendeo.com", IsActive: true, IsProduction: false, Priority: 10,
			Features: []string{"digital_platform", "last_mile_delivery", "e_commerce", "api_integration"},
		},
		{
			ID: "horoz_lojistik", Name: "horoz_lojistik", DisplayName: "Horoz Lojistik",
			Category: "cargo", Type: "cargo_company", Region: "Turkey", Country: "TR",
			Description: "Horoz logistics and transportation",
			Website: "https://www.horozlojistik.com.tr", IsActive: true, IsProduction: false, Priority: 11,
			Features: []string{"logistics", "transportation", "warehousing", "distribution"},
		},
		{
			ID: "borusan_lojistik", Name: "borusan_lojistik", DisplayName: "Borusan Lojistik",
			Category: "cargo", Type: "cargo_company", Region: "Turkey", Country: "TR",
			Description: "Borusan logistics solutions",
			Website: "https://www.borusanlojistik.com", IsActive: true, IsProduction: false, Priority: 12,
			Features: []string{"integrated_logistics", "supply_chain", "warehousing", "transportation"},
		},
		{
			ID: "ekol_lojistik", Name: "ekol_lojistik", DisplayName: "Ekol Lojistik",
			Category: "cargo", Type: "cargo_company", Region: "Turkey", Country: "TR",
			Description: "Ekol international logistics",
			Website: "https://www.ekol.com", IsActive: true, IsProduction: false, Priority: 13,
			Features: []string{"international_logistics", "road_transport", "intermodal", "warehousing"},
		},
		{
			ID: "ceva_lojistik", Name: "ceva_lojistik", DisplayName: "CEVA Lojistik",
			Category: "cargo", Type: "cargo_company", Region: "Turkey", Country: "FR",
			Description: "CEVA Logistics Turkey operations",
			Website: "https://www.cevalogistics.com", IsActive: true, IsProduction: false, Priority: 14,
			Features: []string{"supply_chain", "contract_logistics", "freight_management", "e_solutions"},
		},
		{
			ID: "omsan_lojistik", Name: "omsan_lojistik", DisplayName: "Omsan Lojistik",
			Category: "cargo", Type: "cargo_company", Region: "Turkey", Country: "TR",
			Description: "Omsan logistics and transportation",
			Website: "https://www.omsan.com.tr", IsActive: true, IsProduction: false, Priority: 15,
			Features: []string{"logistics_solutions", "warehousing", "distribution", "transportation"},
		},
		{
			ID: "mars_lojistik", Name: "mars_lojistik", DisplayName: "Mars Lojistik",
			Category: "cargo", Type: "cargo_company", Region: "Turkey", Country: "TR",
			Description: "Mars logistics and cargo services",
			Website: "https://www.marslojistik.com.tr", IsActive: true, IsProduction: false, Priority: 16,
			Features: []string{"cargo_services", "logistics", "warehousing", "distribution"},
		},
		{
			ID: "trendyol_express", Name: "trendyol_express", DisplayName: "Trendyol Express",
			Category: "cargo", Type: "cargo_company", Region: "Turkey", Country: "TR",
			Description: "Trendyol's own delivery service",
			Website: "https://www.trendyolexpress.com", IsActive: true, IsProduction: false, Priority: 17,
			Features: []string{"e_commerce_delivery", "same_day_delivery", "express_service", "marketplace_integration"},
		},
	}

	for _, def := range cargoIntegrations {
		def.CreatedAt = time.Now()
		def.UpdatedAt = time.Now()
		r.integrationDefinitions[def.ID] = def
	}
}

// initializeFulfillmentIntegrations initializes fulfillment integrations (4 total)
func (r *IntegrationRegistry) initializeFulfillmentIntegrations() {
	fulfillmentIntegrations := []*IntegrationDefinition{
		{
			ID: "amazon_fba", Name: "amazon_fba", DisplayName: "Amazon FBA",
			Category: "fulfillment", Type: "fulfillment_service", Region: "Global", Country: "US",
			Description: "Amazon Fulfillment by Amazon service",
			Website: "https://services.amazon.com/fulfillment-by-amazon", IsActive: true, IsProduction: true, Priority: 1,
			Features: []string{"warehouse_management", "order_fulfillment", "customer_service", "returns_processing"},
		},
		{
			ID: "trendyol_fulfillment", Name: "trendyol_fulfillment", DisplayName: "Trendyol Fulfillment",
			Category: "fulfillment", Type: "fulfillment_service", Region: "Turkey", Country: "TR",
			Description: "Trendyol's fulfillment service",
			Website: "https://www.trendyol.com", IsActive: true, IsProduction: true, Priority: 2,
			Features: []string{"warehouse_storage", "order_processing", "shipping", "returns_management"},
		},
		{
			ID: "hepsiburada_fulfillment", Name: "hepsiburada_fulfillment", DisplayName: "HepsiJet Fulfillment",
			Category: "fulfillment", Type: "fulfillment_service", Region: "Turkey", Country: "TR",
			Description: "Hepsiburada's fulfillment and logistics service",
			Website: "https://www.hepsiburada.com", IsActive: true, IsProduction: false, Priority: 3,
			Features: []string{"logistics_service", "warehousing", "last_mile_delivery", "inventory_management"},
		},
		{
			ID: "shipbob", Name: "shipbob", DisplayName: "ShipBob",
			Category: "fulfillment", Type: "fulfillment_service", Region: "Global", Country: "US",
			Description: "E-commerce fulfillment service",
			Website: "https://www.shipbob.com", IsActive: true, IsProduction: false, Priority: 4,
			Features: []string{"fulfillment_network", "inventory_management", "shipping", "analytics"},
		},
	}

	for _, def := range fulfillmentIntegrations {
		def.CreatedAt = time.Now()
		def.UpdatedAt = time.Now()
		r.integrationDefinitions[def.ID] = def
	}
}

// initializeRetailIntegrations initializes retail integrations (2 total)
func (r *IntegrationRegistry) initializeRetailIntegrations() {
	retailIntegrations := []*IntegrationDefinition{
		{
			ID: "shopify_pos", Name: "shopify_pos", DisplayName: "Shopify POS",
			Category: "retail", Type: "pos_system", Region: "Global", Country: "CA",
			Description: "Shopify Point of Sale system",
			Website: "https://www.shopify.com/pos", IsActive: true, IsProduction: true, Priority: 1,
			Features: []string{"pos_integration", "inventory_sync", "omnichannel", "payment_processing"},
		},
		{
			ID: "square_pos", Name: "square_pos", DisplayName: "Square POS",
			Category: "retail", Type: "pos_system", Region: "Global", Country: "US",
			Description: "Square Point of Sale system",
			Website: "https://squareup.com", IsActive: true, IsProduction: true, Priority: 2,
			Features: []string{"pos_system", "payment_processing", "inventory_management", "analytics"},
		},
	}

	for _, def := range retailIntegrations {
		def.CreatedAt = time.Now()
		def.UpdatedAt = time.Now()
		r.integrationDefinitions[def.ID] = def
	}
}