package main

import (
	"os"
	"testing"
	"time"

	"kolajAi/internal/database"
	"kolajAi/internal/database/migrations"
	"kolajAi/internal/handlers"
	"kolajAi/internal/repository"
	"kolajAi/internal/services"
)

func setupTestDB(t *testing.T) *database.MySQLRepository {
	// Create a test database
	testDB := "test_integration.db"

	// Clean up any existing test database
	os.Remove(testDB)

	db, err := database.NewSQLiteConnection(testDB)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Run migrations
	migrationService := migrations.NewMigrationService(db, "kolajAi")
	if err := migrationService.RunMigrations(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create repository
	mysqlRepo := database.NewMySQLRepository(db)

	// Clean up function
	t.Cleanup(func() {
		db.Close()
		os.Remove(testDB)
	})

	return mysqlRepo
}

func TestMainPageIntegration(t *testing.T) {
	// Setup test database
	mysqlRepo := setupTestDB(t)
	repo := repository.NewBaseRepository(mysqlRepo)

	// Create services
	vendorService := services.NewVendorService(repo)
	productService := services.NewProductService(repo)
	orderService := services.NewOrderService(repo)
	auctionService := services.NewAuctionService(repo)

	// Test that services are created successfully
	if vendorService == nil {
		t.Error("VendorService should not be nil")
	}
	if productService == nil {
		t.Error("ProductService should not be nil")
	}
	if orderService == nil {
		t.Error("OrderService should not be nil")
	}
	if auctionService == nil {
		t.Error("AuctionService should not be nil")
	}

	// Test that we can create handlers
	sessionManager := handlers.NewSessionManager("test-secret-key")
	if sessionManager == nil {
		t.Error("SessionManager should not be nil")
	}

	// Create base handler
	h := &handlers.Handler{
		SessionManager: sessionManager,
		TemplateContext: map[string]interface{}{
			"AppName": "KolajAI Test",
			"Year":    time.Now().Year(),
		},
	}

	// Create ecommerce handler
	ecommerceHandler := handlers.NewEcommerceHandler(h, vendorService, productService, orderService, auctionService)
	if ecommerceHandler == nil {
		t.Error("EcommerceHandler should not be nil")
	}

	t.Log("Integration test components created successfully")
}

func TestServiceIntegration(t *testing.T) {
	// Setup test database
	mysqlRepo := setupTestDB(t)
	repo := repository.NewBaseRepository(mysqlRepo)

	// Test that we can create all services
	productService := services.NewProductService(repo)
	orderService := services.NewOrderService(repo)
	aiService := services.NewAIService(repo, productService, orderService)

	if productService == nil {
		t.Error("ProductService should not be nil")
	}
	if orderService == nil {
		t.Error("OrderService should not be nil")
	}
	if aiService == nil {
		t.Error("AIService should not be nil")
	}

	// Test that we can create handlers
	sessionManager := handlers.NewSessionManager("test-secret-key")
	h := &handlers.Handler{
		SessionManager: sessionManager,
		TemplateContext: map[string]interface{}{
			"AppName": "KolajAI Test",
			"Year":    time.Now().Year(),
		},
	}

	// Create AI handler
	aiHandler := handlers.NewAIHandler(h, aiService)
	if aiHandler == nil {
		t.Error("AIHandler should not be nil")
	}

	t.Log("Service integration test completed successfully")
}

func TestDatabaseConnectionIntegration(t *testing.T) {
	// Test that we can connect to the database and perform basic operations
	testDB := "test_db_integration.db"
	defer os.Remove(testDB)

	db, err := database.NewSQLiteConnection(testDB)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test basic query
	_, err = db.Exec("CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Test insert
	_, err = db.Exec("INSERT INTO test_table (name) VALUES (?)", "test")
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Test select
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_table").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query test data: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 row, got %d", count)
	}
}

func TestEnterpriseIntegrationServices(t *testing.T) {
	// Setup test database
	testDB := "test_enterprise_integration.db"
	defer os.Remove(testDB)

	db, err := database.NewSQLiteConnection(testDB)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Run migrations
	migrationService := migrations.NewMigrationService(db, "kolajAi")
	if err := migrationService.RunMigrations(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	mysqlRepo := database.NewMySQLRepository(db)
	repo := repository.NewBaseRepository(mysqlRepo)

	// Test marketplace integrations service
	marketplaceService := services.NewMarketplaceIntegrationsService()
	if marketplaceService == nil {
		t.Error("MarketplaceIntegrationsService should not be nil")
	}

	// Test getting all integrations
	integrations := marketplaceService.GetAllIntegrations()
	if len(integrations) == 0 {
		t.Error("Should have some integrations configured")
	}

	// Test AI integration manager
	productService := services.NewProductService(repo)
	orderService := services.NewOrderService(repo)
	aiService := services.NewAIService(repo, productService, orderService)
	
	aiIntegrationManager := services.NewAIIntegrationManager(marketplaceService, aiService)
	if aiIntegrationManager == nil {
		t.Error("AIIntegrationManager should not be nil")
	}

	// Test webhook service
	webhookService := services.NewIntegrationWebhookService(marketplaceService, aiIntegrationManager)
	if webhookService == nil {
		t.Error("IntegrationWebhookService should not be nil")
	}

	// Test analytics service
	analyticsService := services.NewIntegrationAnalyticsService(db, marketplaceService, aiIntegrationManager)
	if analyticsService == nil {
		t.Error("IntegrationAnalyticsService should not be nil")
	}

	// Test getting metrics
	metrics := analyticsService.GetMetrics()
	if metrics == nil {
		t.Error("Metrics should not be nil")
	}

	t.Log("Enterprise integration services test completed successfully")
}
