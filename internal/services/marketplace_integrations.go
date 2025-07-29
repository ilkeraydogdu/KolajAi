package services

import (
	"fmt"
	"sync"
	"time"
)

// MarketplaceIntegration represents a marketplace integration
type MarketplaceIntegration struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // turkish, international, ecommerce_platform, social_media, accounting, cargo
	Region      string                 `json:"region"`
	IsActive    bool                   `json:"is_active"`
	Config      map[string]interface{} `json:"config"`
	Credentials map[string]string      `json:"credentials"`
	Features    []string               `json:"features"`
}

// MarketplaceIntegrationsService manages all marketplace integrations
type MarketplaceIntegrationsService struct {
	integrations map[string]*MarketplaceIntegration
	rateLimiter  *RateLimitManager
	circuitBreakers map[string]*CircuitBreaker
	metrics      map[string]*IntegrationMetrics
	monitoring   *MonitoringService
	mu           sync.RWMutex
}

// NewMarketplaceIntegrationsService creates a new marketplace integrations service
func NewMarketplaceIntegrationsService() *MarketplaceIntegrationsService {
	// Create monitoring config
	monitoringConfig := &MonitoringConfig{
		CheckInterval: 30 * time.Second,
		AlertThresholds: AlertThresholds{
			ErrorRateThreshold:     0.1,  // 10%
			ResponseTimeThreshold:  5 * time.Second,
			SuccessRateThreshold:   0.9,  // 90%
			CircuitBreakerThreshold: 5,
		},
		NotificationChannels: []string{"email", "slack"},
		RetentionPeriod:     24 * time.Hour,
	}

	service := &MarketplaceIntegrationsService{
		integrations:     make(map[string]*MarketplaceIntegration),
		rateLimiter:      NewRateLimitManager(),
		circuitBreakers:  make(map[string]*CircuitBreaker),
		metrics:          make(map[string]*IntegrationMetrics),
		monitoring:       NewMonitoringService(monitoringConfig),
	}
	
	// Initialize all integrations
	service.initializeTurkishMarketplaces()
	service.initializeInternationalMarketplaces()
	service.initializeEcommercePlatforms()
	service.initializeSocialMediaIntegrations()
	service.initializeAccountingIntegrations()
	service.initializeCargoIntegrations()
	
	// Initialize rate limits and circuit breakers for all integrations
	service.initializeRateLimits()
	service.initializeCircuitBreakers()
	
	return service
}

// initializeTurkishMarketplaces initializes Turkish marketplace integrations
func (s *MarketplaceIntegrationsService) initializeTurkishMarketplaces() {
	turkishMarketplaces := []struct {
		id   string
		name string
	}{
		{"trendyol", "Trendyol"},
		{"hepsiburada", "Hepsiburada"},
		{"ciceksepeti", "ÇiçekSepeti"},
		{"amazon_tr", "Amazon Türkiye"},
		{"pttavm", "PttAvm"},
		{"n11", "N11"},
		{"n11pro", "N11Pro"},
		{"akakce", "Akakçe"},
		{"cimri", "Cimri"},
		{"modanisa", "Modanisa"},
		{"farmazon", "Farmazon"},
		{"flo", "Flo"},
		{"bunadeger", "BunaDeğer"},
		{"lazimbana", "Lazım Bana"},
		{"allesgo", "Allesgo"},
		{"pazarama", "Pazarama"},
		{"vodafone_hersey", "Vodafone Her Şey Yanımda"},
		{"farmaborsa", "Farmaborsa"},
		{"getircarsi", "GetirÇarşı"},
		{"ecza1", "Ecza1"},
		{"turkcell_pasaj", "Turkcell Pasaj"},
		{"teknosa", "Teknosa"},
		{"idefix", "İdefix"},
		{"koctas", "Koçtaş"},
		{"pempati", "Pempati"},
		{"lcw", "LCW"},
		{"alisgidis", "AlışGidiş"},
		{"beymen", "Beymen"},
		{"novadan", "Novadan"},
		{"magazanolsun", "MagazanOlsun"},
	}
	
	for _, mp := range turkishMarketplaces {
		s.integrations[mp.id] = &MarketplaceIntegration{
			ID:       mp.id,
			Name:     mp.name,
			Type:     "turkish",
			Region:   "TR",
			IsActive: true,
			Config: map[string]interface{}{
				"api_version":      "v1",
				"rate_limit":       100,
				"sync_interval":    300,
				"max_products":     50000,
				"supports_variants": true,
			},
			Credentials: map[string]string{
				"api_key":    "",
				"api_secret": "",
				"store_id":   "",
			},
			Features: []string{
				"product_sync",
				"order_sync",
				"inventory_sync",
				"price_sync",
				"category_mapping",
				"bulk_operations",
				"real_time_notifications",
			},
		}
	}
	
	// Add retail sales modules
	retailModules := []struct {
		id   string
		name string
	}{
		{"prapazar_store", "PraPazar Mağazası"},
		{"prastore", "PraStore Mağazası"},
	}
	
	for _, module := range retailModules {
		s.integrations[module.id] = &MarketplaceIntegration{
			ID:       module.id,
			Name:     module.name,
			Type:     "retail_module",
			Region:   "TR",
			IsActive: true,
			Config: map[string]interface{}{
				"api_version":      "v1",
				"pos_integration":  true,
				"inventory_sync":   true,
				"offline_support":  true,
				"mobile_app":       true,
			},
			Credentials: map[string]string{
				"store_code":   "",
				"pos_key":      "",
				"terminal_id":  "",
			},
			Features: []string{
				"pos_integration",
				"inventory_sync",
				"sales_reporting",
				"customer_management",
				"offline_mode",
				"mobile_support",
				"receipt_printing",
			},
		}
	}
}

// initializeInternationalMarketplaces initializes international marketplace integrations
func (s *MarketplaceIntegrationsService) initializeInternationalMarketplaces() {
	internationalMarketplaces := []struct {
		id     string
		name   string
		region string
	}{
		{"amazon_us", "Amazon US", "US"},
		{"amazon_uk", "Amazon UK", "UK"},
		{"amazon_de", "Amazon Germany", "DE"},
		{"amazon_fr", "Amazon France", "FR"},
		{"amazon_nl", "Amazon Netherlands", "NL"},
		{"amazon_it", "Amazon Italy", "IT"},
		{"amazon_ca", "Amazon Canada", "CA"},
		{"amazon_ae", "Amazon UAE", "AE"},
		{"amazon_es", "Amazon Spain", "ES"},
		{"ebay", "eBay", "GLOBAL"},
		{"aliexpress", "AliExpress", "GLOBAL"},
		{"etsy", "Etsy", "GLOBAL"},
		{"ozon", "Ozon", "RU"},
		{"joom", "Joom", "GLOBAL"},
		{"fruugo", "Fruugo", "GLOBAL"},
		{"allegro", "Allegro", "PL"},
		{"hepsiglobal", "HepsiGlobal", "GLOBAL"},
		{"bolcom", "Bol.com", "NL"},
		{"onbuy", "OnBuy", "UK"},
		{"wayfair", "Wayfair", "US"},
		{"zoodmall", "ZoodMall", "GLOBAL"},
		{"walmart", "Walmart", "US"},
		{"jumia", "Jumia", "AFRICA"},
		{"zalando", "Zalando", "EU"},
		{"cdiscount", "Cdiscount", "FR"},
		{"wish", "Wish", "GLOBAL"},
		{"otto", "Otto", "DE"},
		{"rakuten", "Rakuten", "JP"},
	}
	
	for _, mp := range internationalMarketplaces {
		s.integrations[mp.id] = &MarketplaceIntegration{
			ID:       mp.id,
			Name:     mp.name,
			Type:     "international",
			Region:   mp.region,
			IsActive: true,
			Config: map[string]interface{}{
				"api_version":        "v2",
				"rate_limit":         50,
				"sync_interval":      600,
				"max_products":       100000,
				"supports_variants":  true,
				"multi_currency":     true,
				"multi_language":     true,
				"shipping_templates": true,
			},
			Credentials: map[string]string{
				"api_key":     "",
				"api_secret":  "",
				"merchant_id": "",
				"auth_token":  "",
			},
			Features: []string{
				"product_sync",
				"order_sync",
				"inventory_sync",
				"price_sync",
				"category_mapping",
				"bulk_operations",
				"real_time_notifications",
				"multi_currency",
				"shipping_calculation",
				"tax_calculation",
				"return_management",
			},
		}
	}
}

// initializeEcommercePlatforms initializes e-commerce platform integrations
func (s *MarketplaceIntegrationsService) initializeEcommercePlatforms() {
	platforms := []struct {
		id   string
		name string
	}{
		{"tsoft", "T-soft"},
		{"ticimax", "Ticimax"},
		{"ideasoft", "İdeasoft"},
		{"platinmarket", "Platin Market"},
		{"woocommerce", "WooCommerce"},
		{"opencart", "OpenCart"},
		{"shopphp", "ShopPHP"},
		{"shopify", "Shopify"},
		{"prestashop", "PrestaShop"},
		{"magento", "Magento"},
		{"ethica", "Ethica"},
		{"ikas", "İkas"},
	}
	
	for _, platform := range platforms {
		s.integrations[platform.id] = &MarketplaceIntegration{
			ID:       platform.id,
			Name:     platform.name,
			Type:     "ecommerce_platform",
			Region:   "GLOBAL",
			IsActive: true,
			Config: map[string]interface{}{
				"api_version":      "latest",
				"sync_mode":        "bidirectional",
				"webhook_support":  true,
				"custom_fields":    true,
				"plugin_available": true,
			},
			Credentials: map[string]string{
				"api_url":      "",
				"api_key":      "",
				"api_secret":   "",
				"access_token": "",
			},
			Features: []string{
				"full_sync",
				"real_time_sync",
				"webhook_integration",
				"custom_fields",
				"multi_store",
				"seo_sync",
				"customer_sync",
				"payment_sync",
			},
		}
	}
}

// initializeSocialMediaIntegrations initializes social media integrations
func (s *MarketplaceIntegrationsService) initializeSocialMediaIntegrations() {
	socialMedia := []struct {
		id   string
		name string
	}{
		{"facebook_shop", "Facebook Shop"},
		{"google_merchant", "Google Merchant Center"},
		{"instagram_shop", "Instagram Mağaza"},
	}
	
	for _, sm := range socialMedia {
		s.integrations[sm.id] = &MarketplaceIntegration{
			ID:       sm.id,
			Name:     sm.name,
			Type:     "social_media",
			Region:   "GLOBAL",
			IsActive: true,
			Config: map[string]interface{}{
				"api_version":     "latest",
				"catalog_support": true,
				"pixel_tracking":  true,
				"dynamic_ads":     true,
			},
			Credentials: map[string]string{
				"app_id":       "",
				"app_secret":   "",
				"access_token": "",
				"pixel_id":     "",
			},
			Features: []string{
				"catalog_sync",
				"pixel_tracking",
				"dynamic_remarketing",
				"shopping_tags",
				"checkout_integration",
			},
		}
	}
}

// initializeAccountingIntegrations initializes accounting and ERP integrations
func (s *MarketplaceIntegrationsService) initializeAccountingIntegrations() {
	// E-Fatura integrations
	efaturaProviders := []string{
		"qnb_efatura", "n11_faturam", "nilvera", "uyumsoft", "trendyol_efatura",
		"foriba", "digital_planet", "turkcell_efatura", "smart_fatura", "edm_efatura",
		"ice_efatura", "izibiz", "mysoft", "faturamix", "nesbilgi_efatura",
	}
	
	for _, provider := range efaturaProviders {
		s.integrations[provider] = &MarketplaceIntegration{
			ID:       provider,
			Name:     provider,
			Type:     "efatura",
			Region:   "TR",
			IsActive: true,
			Config: map[string]interface{}{
				"api_version": "v1",
				"gib_support": true,
			},
			Credentials: map[string]string{
				"username": "",
				"password": "",
				"api_key":  "",
			},
			Features: []string{
				"invoice_creation",
				"invoice_sending",
				"gib_integration",
				"archive_support",
			},
		}
	}
	
	// Accounting & ERP systems
	accountingSystems := []string{
		"logo", "mikro", "netsis", "netsim", "dia", "nethesap",
		"zirve", "akinsoft", "vega", "nebim", "barsoft", "sentez",
	}
	
	for _, system := range accountingSystems {
		s.integrations[system] = &MarketplaceIntegration{
			ID:       system,
			Name:     system,
			Type:     "accounting",
			Region:   "TR",
			IsActive: true,
			Config: map[string]interface{}{
				"api_version":    "latest",
				"db_integration": true,
			},
			Credentials: map[string]string{
				"server":   "",
				"database": "",
				"username": "",
				"password": "",
			},
			Features: []string{
				"invoice_sync",
				"stock_sync",
				"customer_sync",
				"financial_reports",
			},
		}
	}
	
	// Pre-accounting systems
	preAccountingSystems := []struct {
		id   string
		name string
	}{
		{"pranomi", "PraNomi"},
		{"parasut", "Paraşüt"},
		{"bizim_hesap", "Bizim Hesap"},
		{"uyumsoft_on_muhasebe", "Uyumsoft Ön Muhasebe"},
		{"odoo_muhasebe", "Odoo Muhasebe"},
	}
	
	for _, system := range preAccountingSystems {
		s.integrations[system.id] = &MarketplaceIntegration{
			ID:       system.id,
			Name:     system.name,
			Type:     "pre_accounting",
			Region:   "TR",
			IsActive: true,
			Config: map[string]interface{}{
				"api_version":   "latest",
				"cloud_based":   true,
				"mobile_app":    true,
			},
			Credentials: map[string]string{
				"api_key":      "",
				"api_secret":   "",
				"company_id":   "",
			},
			Features: []string{
				"invoice_management",
				"expense_tracking",
				"bank_integration",
				"report_generation",
			},
		}
	}
}

// initializeCargoIntegrations initializes cargo and fulfillment integrations
func (s *MarketplaceIntegrationsService) initializeCargoIntegrations() {
	// Cargo companies
	cargoCompanies := []struct {
		id   string
		name string
	}{
		{"yurtici", "Yurtiçi Kargo"},
		{"aras", "Aras Kargo"},
		{"mng", "MNG Kargo"},
		{"ptt", "PTT Kargo"},
		{"ups", "UPS"},
		{"surat", "Sürat Kargo"},
		{"foodman", "FoodMan Lojistik"},
		{"cdek", "Cdek"},
		{"sendeo", "Sendeo"},
		{"pts", "PTS Kargo"},
		{"fedex", "FedEx"},
		{"shipentegra", "ShipEntegra"},
		{"dhl", "DHL"},
		{"hepsijet", "HepsiJet"},
		{"tnt", "TNT"},
		{"ekol", "Ekol Logistics"},
		{"kolaygelsin", "Kolay Gelsin"},
	}
	
	for _, cargo := range cargoCompanies {
		s.integrations[cargo.id] = &MarketplaceIntegration{
			ID:       cargo.id,
			Name:     cargo.name,
			Type:     "cargo",
			Region:   "TR",
			IsActive: true,
			Config: map[string]interface{}{
				"api_version":      "v1",
				"tracking_support": true,
				"label_printing":   true,
				"bulk_shipping":    true,
			},
			Credentials: map[string]string{
				"customer_code": "",
				"username":      "",
				"password":      "",
				"api_key":       "",
			},
			Features: []string{
				"shipment_creation",
				"tracking",
				"label_printing",
				"bulk_operations",
				"pickup_scheduling",
				"delivery_notification",
			},
		}
	}
	
	// Fulfillment services
	fulfillmentServices := []struct {
		id   string
		name string
	}{
		{"oplog", "Oplog Fulfillment"},
		{"hepsilojistik", "Hepsilojistik"},
		{"n11depom", "N11Depom"},
		{"navlungo", "Navlungo Fulfillment"},
	}
	
	for _, fulfillment := range fulfillmentServices {
		s.integrations[fulfillment.id] = &MarketplaceIntegration{
			ID:       fulfillment.id,
			Name:     fulfillment.name,
			Type:     "fulfillment",
			Region:   "TR",
			IsActive: true,
			Config: map[string]interface{}{
				"api_version":        "v1",
				"warehouse_support":  true,
				"inventory_tracking": true,
				"order_fulfillment":  true,
			},
			Credentials: map[string]string{
				"api_key":      "",
				"api_secret":   "",
				"warehouse_id": "",
			},
			Features: []string{
				"inventory_management",
				"order_fulfillment",
				"warehouse_operations",
				"return_handling",
				"reporting",
			},
		}
	}
}

// GetIntegration returns a specific integration by ID
func (s *MarketplaceIntegrationsService) GetIntegration(id string) (*MarketplaceIntegration, error) {
	integration, exists := s.integrations[id]
	if !exists {
		return nil, fmt.Errorf("integration %s not found", id)
	}
	return integration, nil
}

// GetAllIntegrations returns all integrations
func (s *MarketplaceIntegrationsService) GetAllIntegrations() map[string]*MarketplaceIntegration {
	return s.integrations
}

// GetIntegrationsByType returns integrations filtered by type
func (s *MarketplaceIntegrationsService) GetIntegrationsByType(integrationType string) []*MarketplaceIntegration {
	var result []*MarketplaceIntegration
	for _, integration := range s.integrations {
		if integration.Type == integrationType {
			result = append(result, integration)
		}
	}
	return result
}

// ConfigureIntegration configures an integration with credentials
func (s *MarketplaceIntegrationsService) ConfigureIntegration(id string, credentials map[string]string) error {
	integration, err := s.GetIntegration(id)
	if err != nil {
		return err
	}
	
	// Update credentials
	for key, value := range credentials {
		integration.Credentials[key] = value
	}
	
	// Test connection
	if err := s.testIntegrationConnection(integration); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	
	return nil
}

// testIntegrationConnection tests if the integration can connect successfully
func (s *MarketplaceIntegrationsService) testIntegrationConnection(integration *MarketplaceIntegration) error {
	// This would implement actual API connection tests for each integration
	// For now, we'll simulate it
	return nil
}

// SyncProducts syncs products with a marketplace with enhanced error handling
func (s *MarketplaceIntegrationsService) SyncProducts(integrationID string, products []interface{}) error {
	startTime := time.Now()
	metrics := s.ensureMetrics(integrationID)
	
	// Check rate limit
	if err := s.rateLimiter.CheckRateLimit(integrationID); err != nil {
		metrics.RecordRequest(false, time.Since(startTime))
		return fmt.Errorf("rate limit exceeded: %v", err)
	}
	
	// Get circuit breaker
	breaker, exists := s.getCircuitBreaker(integrationID)
	if !exists {
		metrics.RecordRequest(false, time.Since(startTime))
		return fmt.Errorf("circuit breaker not found for integration: %s", integrationID)
	}
	
	// Execute with circuit breaker protection
	var syncErr error
	err := breaker.Execute(func() error {
		integration, err := s.GetIntegration(integrationID)
		if err != nil {
			return err
		}
		
		// Sync products based on integration type with retry mechanism
		retryConfig := NewRetryConfig(3, 1*time.Second)
		syncErr = RetryOperation(func() error {
			switch integration.Type {
			case "turkish":
				return s.syncToTurkishMarketplace(integration, products)
			case "international":
				return s.syncToInternationalMarketplace(integration, products)
			case "ecommerce_platform":
				return s.syncToEcommercePlatform(integration, products)
			case "social_media":
				return s.syncToSocialMedia(integration, products)
			default:
				return fmt.Errorf("unsupported integration type: %s", integration.Type)
			}
		}, retryConfig)
		
		return syncErr
	})
	
	// Record metrics
	responseTime := time.Since(startTime)
	success := err == nil
	metrics.RecordRequest(success, responseTime)
	
	// Update monitoring
	s.monitoring.MonitorIntegration(integrationID, metrics)
	
	// Update health check
	healthStatus := HealthStatusHealthy
	if err != nil {
		healthStatus = HealthStatusUnhealthy
	}
	s.monitoring.UpdateHealthCheck(integrationID, healthStatus, responseTime, err)
	
	return err
}

// syncToTurkishMarketplace syncs products to Turkish marketplaces
func (s *MarketplaceIntegrationsService) syncToTurkishMarketplace(integration *MarketplaceIntegration, products []interface{}) error {
	// Implement specific sync logic for each Turkish marketplace
	// This would include API calls, data transformation, etc.
	return nil
}

// syncToInternationalMarketplace syncs products to international marketplaces
func (s *MarketplaceIntegrationsService) syncToInternationalMarketplace(integration *MarketplaceIntegration, products []interface{}) error {
	// Implement specific sync logic for each international marketplace
	// This would include API calls, data transformation, currency conversion, etc.
	return nil
}

// syncToEcommercePlatform syncs products to e-commerce platforms
func (s *MarketplaceIntegrationsService) syncToEcommercePlatform(integration *MarketplaceIntegration, products []interface{}) error {
	// Implement specific sync logic for each e-commerce platform
	// This would include API calls, plugin integration, etc.
	return nil
}

// syncToSocialMedia syncs products to social media platforms
func (s *MarketplaceIntegrationsService) syncToSocialMedia(integration *MarketplaceIntegration, products []interface{}) error {
	// Implement specific sync logic for social media platforms
	// This would include catalog creation, pixel setup, etc.
	return nil
}

// ProcessOrder processes an order from a marketplace with enhanced error handling
func (s *MarketplaceIntegrationsService) ProcessOrder(integrationID string, orderData interface{}) error {
	startTime := time.Now()
	metrics := s.ensureMetrics(integrationID)
	
	// Check rate limit
	if err := s.rateLimiter.CheckRateLimit(integrationID); err != nil {
		metrics.RecordRequest(false, time.Since(startTime))
		return fmt.Errorf("rate limit exceeded: %v", err)
	}
	
	// Get circuit breaker
	breaker, exists := s.getCircuitBreaker(integrationID)
	if !exists {
		metrics.RecordRequest(false, time.Since(startTime))
		return fmt.Errorf("circuit breaker not found for integration: %s", integrationID)
	}
	
	// Execute with circuit breaker protection
	err := breaker.Execute(func() error {
		_, err := s.GetIntegration(integrationID)
		if err != nil {
			return err
		}
		
		// Process order based on integration type with retry mechanism
		retryConfig := NewRetryConfig(3, 1*time.Second)
		return RetryOperation(func() error {
			// Process order based on integration type
			// This would include order validation, inventory update, notification, etc.
			return nil
		}, retryConfig)
	})
	
	// Record metrics
	responseTime := time.Since(startTime)
	success := err == nil
	metrics.RecordRequest(success, responseTime)
	
	// Update monitoring
	s.monitoring.MonitorIntegration(integrationID, metrics)
	
	// Update health check
	healthStatus := HealthStatusHealthy
	if err != nil {
		healthStatus = HealthStatusUnhealthy
	}
	s.monitoring.UpdateHealthCheck(integrationID, healthStatus, responseTime, err)
	
	return err
}

// UpdateInventory updates inventory across integrated marketplaces
func (s *MarketplaceIntegrationsService) UpdateInventory(productID string, quantity int) error {
	// Update inventory across all active integrations
	for _, integration := range s.integrations {
		if integration.IsActive {
			// Send inventory update to each marketplace
			// This would be done asynchronously in production
		}
	}
	return nil
}

// GetMarketplaceOrders retrieves orders from a specific marketplace
func (s *MarketplaceIntegrationsService) GetMarketplaceOrders(integrationID string, since time.Time) ([]interface{}, error) {
	_, err := s.GetIntegration(integrationID)
	if err != nil {
		return nil, err
	}
	
	// Fetch orders from marketplace API
	// This would implement actual API calls
	
	return []interface{}{}, nil
}

// CreateShipment creates a shipment with a cargo integration
func (s *MarketplaceIntegrationsService) CreateShipment(cargoID string, shipmentData interface{}) (string, error) {
	integration, err := s.GetIntegration(cargoID)
	if err != nil {
		return "", err
	}
	
	if integration.Type != "cargo" {
		return "", fmt.Errorf("integration %s is not a cargo service", cargoID)
	}
	
	// Create shipment with cargo API
	// Return tracking number
	
	return "TRACK123456", nil
}

// GenerateInvoice generates an invoice using e-fatura integration
func (s *MarketplaceIntegrationsService) GenerateInvoice(efaturaID string, invoiceData interface{}) (string, error) {
	integration, err := s.GetIntegration(efaturaID)
	if err != nil {
		return "", err
	}
	
	if integration.Type != "efatura" {
		return "", fmt.Errorf("integration %s is not an e-fatura service", efaturaID)
	}
	
	// Generate invoice with e-fatura API
	// Return invoice number
	
	return "INV2024001", nil
}

// initializeRateLimits initializes rate limits for all integrations
func (s *MarketplaceIntegrationsService) initializeRateLimits() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Set default rate limits for different integration types
	defaultLimits := map[string]*RateLimit{
		"turkish": {
			RequestsPerMinute: 60,
			RequestsPerHour:   1000,
			RequestsPerDay:    10000,
			WindowSize:        time.Minute,
			BurstSize:         100,
		},
		"international": {
			RequestsPerMinute: 30,
			RequestsPerHour:   500,
			RequestsPerDay:    5000,
			WindowSize:        time.Minute,
			BurstSize:         50,
		},
		"ecommerce_platform": {
			RequestsPerMinute: 100,
			RequestsPerHour:   2000,
			RequestsPerDay:    20000,
			WindowSize:        time.Minute,
			BurstSize:         200,
		},
		"social_media": {
			RequestsPerMinute: 20,
			RequestsPerHour:   300,
			RequestsPerDay:    3000,
			WindowSize:        time.Minute,
			BurstSize:         30,
		},
		"accounting": {
			RequestsPerMinute: 10,
			RequestsPerHour:   100,
			RequestsPerDay:    1000,
			WindowSize:        time.Minute,
			BurstSize:         20,
		},
		"cargo": {
			RequestsPerMinute: 40,
			RequestsPerHour:   600,
			RequestsPerDay:    6000,
			WindowSize:        time.Minute,
			BurstSize:         80,
		},
	}

	// Apply rate limits to all integrations
	for integrationID, integration := range s.integrations {
		if limit, exists := defaultLimits[integration.Type]; exists {
			limit.IntegrationID = integrationID
			s.rateLimiter.SetRateLimit(integrationID, limit)
		}
	}
}

// initializeCircuitBreakers initializes circuit breakers for all integrations
func (s *MarketplaceIntegrationsService) initializeCircuitBreakers() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Set default circuit breaker configurations
	defaultConfigs := map[string]struct {
		threshold int
		timeout   time.Duration
	}{
		"turkish":            {5, 30 * time.Second},
		"international":      {3, 60 * time.Second},
		"ecommerce_platform": {10, 20 * time.Second},
		"social_media":       {2, 120 * time.Second},
		"accounting":         {3, 60 * time.Second},
		"cargo":              {5, 45 * time.Second},
	}

	// Create circuit breakers for all integrations
	for integrationID, integration := range s.integrations {
		if config, exists := defaultConfigs[integration.Type]; exists {
			s.circuitBreakers[integrationID] = NewCircuitBreaker(config.threshold, config.timeout)
		}
	}
}

// getCircuitBreaker gets circuit breaker for an integration
func (s *MarketplaceIntegrationsService) getCircuitBreaker(integrationID string) (*CircuitBreaker, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	breaker, exists := s.circuitBreakers[integrationID]
	return breaker, exists
}

// getMetrics gets metrics for an integration
func (s *MarketplaceIntegrationsService) getMetrics(integrationID string) (*IntegrationMetrics, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	metrics, exists := s.metrics[integrationID]
	return metrics, exists
}

// ensureMetrics ensures metrics exist for an integration
func (s *MarketplaceIntegrationsService) ensureMetrics(integrationID string) *IntegrationMetrics {
	s.mu.Lock()
	defer s.mu.Unlock()

	if metrics, exists := s.metrics[integrationID]; exists {
		return metrics
	}

	metrics := NewIntegrationMetrics(integrationID)
	s.metrics[integrationID] = metrics
	return metrics
}

// GetMonitoringService returns the monitoring service
func (s *MarketplaceIntegrationsService) GetMonitoringService() *MonitoringService {
	return s.monitoring
}

// GetAlerts returns all active alerts
func (s *MarketplaceIntegrationsService) GetAlerts() []*Alert {
	return s.monitoring.GetAlerts()
}

// GetAlertsByIntegration returns alerts for a specific integration
func (s *MarketplaceIntegrationsService) GetAlertsByIntegration(integrationID string) []*Alert {
	return s.monitoring.GetAlertsByIntegration(integrationID)
}

// ResolveAlert resolves an alert
func (s *MarketplaceIntegrationsService) ResolveAlert(alertID string) error {
	return s.monitoring.ResolveAlert(alertID)
}

// GetHealthStatus returns health status for an integration
func (s *MarketplaceIntegrationsService) GetHealthStatus(integrationID string) *HealthCheck {
	return s.monitoring.GetHealthStatus(integrationID)
}

// GetMetrics returns metrics for an integration
func (s *MarketplaceIntegrationsService) GetMetrics(integrationID string) (*IntegrationMetrics, bool) {
	return s.monitoring.GetMetrics(integrationID)
}

// GetAllMetrics returns all metrics
func (s *MarketplaceIntegrationsService) GetAllMetrics() map[string]*IntegrationMetrics {
	return s.monitoring.GetAllMetrics()
}