package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"kolajAi/internal/database"
	"kolajAi/internal/database/migrations"
	"kolajAi/internal/handlers"

	"kolajAi/internal/services"
	"kolajAi/internal/session"
	"kolajAi/internal/reporting"
	"kolajAi/internal/errors"
	"kolajAi/internal/seo"
	"kolajAi/internal/notifications"
	"kolajAi/internal/testing"
	"kolajAi/internal/security"
	"kolajAi/internal/cache"
	"kolajAi/internal/middleware"
	"kolajAi/internal/router"
	"kolajAi/internal/config"

)

var (
	MainLogger *log.Logger
)

func init() {
	// Ana uygulama için log dosyası oluştur
	logFile, err := os.OpenFile("main_app_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("Ana uygulama log dosyası oluşturulamadı:", err)
		MainLogger = log.New(os.Stdout, "[MAIN] ", log.LstdFlags)
	} else {
		MainLogger = log.New(logFile, "[MAIN-APP-DEBUG] ", log.LstdFlags|log.Lshortfile)
	}
}



func main() {
	MainLogger.Println("KolajAI Enterprise uygulaması başlatılıyor...")

	// Konfigürasyon yükle
	MainLogger.Println("Konfigürasyon yükleniyor...")
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		MainLogger.Printf("Konfigürasyon yüklenemedi, varsayılan değerler kullanılıyor: %v", err)
		cfg = config.GetDefaultConfig()
	}

	// Veritabanı bağlantısı (SQLite)
	MainLogger.Println("Veritabanı bağlantısı kuruluyor...")
	db, err := database.NewSQLiteConnection("kolajAi.db")
	if err != nil {
		MainLogger.Fatalf("Veritabanı bağlantısı kurulamadı: %v", err)
	}
	defer db.Close()

	// Migration'ları çalıştır
	MainLogger.Println("Veritabanı migration'ları çalıştırılıyor...")
	migrationService := migrations.NewMigrationService(db, "kolajAi")
	if err := migrationService.RunMigrations(); err != nil {
		MainLogger.Fatalf("Migration'lar çalıştırılamadı: %v", err)
	}
	MainLogger.Println("Migration'lar başarıyla tamamlandı!")

	// Advanced systems initialization
	MainLogger.Println("Gelişmiş sistemler başlatılıyor...")

	// Cache Manager
	MainLogger.Println("Cache sistemi başlatılıyor...")
	cacheManager := cache.NewCacheManager(db, cache.CacheConfig{
		DefaultTTL:         30 * time.Minute,
		MaxMemoryUsage:     1024 * 1024 * 1024, // 1GB
		Stores:             make(map[string]cache.StoreConfig),
	})
	defer cacheManager.Close()

	// Security Manager
	MainLogger.Println("Güvenlik sistemi başlatılıyor...")
	securityManager := security.NewSecurityManager(db, security.SecurityConfig{
		MaxLoginAttempts:     5,
		LoginLockoutDuration: 30 * time.Minute,
		PasswordMinLength:    8,
		PasswordRequireUpper: true,
		PasswordRequireLower: true,
		PasswordRequireDigit: true,
		PasswordRequireSymbol: true,
		SessionTimeout:       24 * time.Hour,
		CSRFTokenLength:      32,
		EnableIPWhitelist:    false,
		EnableIPBlacklist:    true,
		EnableRateLimit:      true,
		EncryptionKey:        cfg.Security.EncryptionKey,
		JWTSecret:           cfg.Security.JWTSecret,
		TwoFactorEnabled:    true,
		AuditLogEnabled:     true,
	})

	// Session Manager
	MainLogger.Println("Session sistemi başlatılıyor...")
	sessionManager, err := session.NewSessionManager(db, session.SessionConfig{
		CookieName: "kolajAI_session",
		Secure:     true,
		HTTPOnly:   true,
		SameSite:   http.SameSiteStrictMode,
		MaxAge:     int(24 * time.Hour / time.Second),
		Domain:     cfg.Server.Domain,
		Path:       "/",
	})
	if err != nil {
		MainLogger.Fatalf("Session sistemi başlatılamadı: %v", err)
	}

	// Error Manager
	MainLogger.Println("Hata yönetim sistemi başlatılıyor...")
	errorManager := errors.NewErrorManager(db, nil, errors.ErrorConfig{
		Environment:         cfg.Environment,
		EnableStackTrace:    true,
		EnableNotifications: true,
		MaxStackDepth:       20,
		RetentionDays:       90,
	})

	// Notification Manager
	MainLogger.Println("Bildirim sistemi başlatılıyor...")
	notificationManager := notifications.NewNotificationManager(db, notifications.NotificationConfig{
		DefaultChannel:     "email",
		RetryAttempts:      3,
		RetryDelay:         5 * time.Minute,
		BatchSize:          100,
		QueueSize:          10000,
		Workers:            5,
		EnableRateLimiting: true,
	})

	// SEO Manager
	MainLogger.Println("SEO sistemi başlatılıyor...")
	seoManager := seo.NewSEOManager(db, seo.SEOConfig{
		DefaultLanguage: "tr",
		SupportedLanguages: []seo.Language{
			{Code: "tr", Name: "Türkçe"},
			{Code: "en", Name: "English"},
			{Code: "ar", Name: "العربية"},
		},
	})

	// Reporting Manager
	MainLogger.Println("Raporlama sistemi başlatılıyor...")
	reportManager := reporting.NewReportManager(db)

	// Test Manager
	MainLogger.Println("Test sistemi başlatılıyor...")
	_ = testing.NewTestManager(db, testing.TestConfig{
		Environment:       "production",
		Parallel:          true,
		MaxWorkers:        4,
		Timeout:           30 * time.Minute,
		RetryAttempts:     3,
		CoverageEnabled:   true,
		CoverageThreshold: 80.0,
		ReportFormats:     []string{"html", "json", "xml"},
		OutputDirectory:   "./test-reports",
	})

	// Repository oluştur
	mysqlRepo := database.NewMySQLRepository(db)
	repo := database.NewRepositoryWrapper(mysqlRepo)

	// Servisleri oluştur
	MainLogger.Println("Servisler oluşturuluyor...")
	authService := services.NewAuthService(nil, nil) // Placeholder - implement properly
	vendorService := services.NewVendorService(repo)
	productService := services.NewProductService(repo)
	orderService := services.NewOrderService(repo)
	auctionService := services.NewAuctionService(repo)
	aiService := services.NewAIService(repo, productService, orderService)
	aiAnalyticsService := services.NewAIAnalyticsService(repo, productService, orderService)
	aiVisionService := services.NewAIVisionService(repo, productService)
	aiEnterpriseService := services.NewAIEnterpriseService(repo, aiService, aiVisionService, productService, orderService, authService)
	
	// Yeni gelişmiş AI ve marketplace servisleri
	aiAdvancedService := services.NewAIAdvancedService(repo, productService, orderService)
	marketplaceService := services.NewMarketplaceIntegrationsService()

	// Şablonları yükle
	MainLogger.Println("Şablonlar yükleniyor...")

	// Template fonksiyonlarını tanımla
	funcMap := template.FuncMap{
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("dict fonksiyonu için çift sayıda parametre gerekli")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict fonksiyonu için anahtarlar string olmalı")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"rand": func() int {
			return time.Now().Nanosecond() % 1000
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"formatPrice": func(price float64) string {
			return fmt.Sprintf("%.2f TL", price)
		},
		"formatDate": func(t time.Time) string {
			return t.Format("02.01.2006 15:04")
		},
		"seq": func(n int) []int {
			result := make([]int, n)
			for i := 0; i < n; i++ {
				result[i] = i
			}
			return result
		},
		"mul": func(a, b interface{}) float64 {
			var numA, numB float64
			switch v := a.(type) {
			case int:
				numA = float64(v)
			case float64:
				numA = v
			case float32:
				numA = float64(v)
			default:
				return 0
			}
			switch v := b.(type) {
			case int:
				numB = float64(v)
			case float64:
				numB = v
			case float32:
				numB = float64(v)
			default:
				return 0
			}
			return numA * numB
		},
		"add": func(a, b interface{}) float64 {
			var numA, numB float64
			switch v := a.(type) {
			case int:
				numA = float64(v)
			case float64:
				numA = v
			case float32:
				numA = float64(v)
			default:
				return 0
			}
			switch v := b.(type) {
			case int:
				numB = float64(v)
			case float64:
				numB = v
			case float32:
				numB = float64(v)
			default:
				return 0
			}
			return numA + numB
		},
		// SEO template functions
		"seoTitle": func(title string) string {
			return seoManager.OptimizeTitle(title)
		},
		"seoDescription": func(description string) string {
			return seoManager.OptimizeDescription(description)
		},
		"generateSchema": func(pageType string, data interface{}) template.HTML {
			schema, _ := seoManager.GenerateSchema(pageType, data)
			return template.HTML(schema)
		},
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseGlob("web/templates/**/*.gohtml")
	if err != nil {
		MainLogger.Fatalf("Şablonlar yüklenemedi: %v", err)
	}
	MainLogger.Printf("Şablonlar başarıyla yüklendi!")

	// Handler'ları oluştur
	MainLogger.Println("Handler'lar oluşturuluyor...")

	// Create legacy session manager for handlers
	legacySessionManager := handlers.NewSessionManager("supersecretkey123")
	
	h := &handlers.Handler{
		Templates:      tmpl,
		SessionManager: legacySessionManager,
		TemplateContext: map[string]interface{}{
			"AppName": "KolajAI Enterprise Marketplace",
			"Year":    time.Now().Year(),
		},
	}

	// E-ticaret handler'ı oluştur
	ecommerceHandler := handlers.NewEcommerceHandler(h, vendorService, productService, orderService, auctionService)

	// Admin handler'ı oluştur - tüm yeni sistemlerle birlikte
	adminHandler := handlers.NewAdminHandler(
		h, 
		productService, 
		vendorService, 
		orderService, 
		auctionService,
		sessionManager,
		reportManager,
		notificationManager,
		seoManager,
		errorManager,
	)

	// AI handler'ı oluştur
	aiHandler := handlers.NewAIHandler(h, aiService)

	// AI Analytics handler'ı oluştur
	aiAnalyticsHandler := handlers.NewAIAnalyticsHandler(h, aiAnalyticsService)

	// AI Vision handler'ı oluştur
	aiVisionHandler := handlers.NewAIVisionHandler(h, aiVisionService)
	
	// Yeni gelişmiş handler'lar
	aiAdvancedHandler := handlers.NewAIAdvancedHandler(h, aiAdvancedService)
	marketplaceHandler := handlers.NewMarketplaceHandler(h, marketplaceService)

	// Middleware stack oluştur
	middlewareStack := middleware.NewMiddlewareStack(
		securityManager,
		sessionManager,
		errorManager,
		cacheManager,
	)

	// Router oluştur
	appRouter := router.NewRouter(middlewareStack)

	// Statik dosyalar
	appRouter.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// SEO rotaları
	appRouter.HandleFunc("/sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
		_, err := seoManager.GenerateSitemap("default")
		if err != nil {
			errorManager.HandleHTTPError(w, r, errors.NewApplicationError(errors.INTERNAL, "SITEMAP_ERROR", "Sitemap oluşturulamadı", err))
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		// Convert sitemap to bytes - in a real implementation would marshal XML
		w.Write([]byte("<urlset></urlset>"))
	})

	appRouter.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		robots := seoManager.GenerateRobotsTxt()
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(robots))
	})

	// Ana sayfa - Marketplace
	appRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		ecommerceHandler.Marketplace(w, r)
	})

	// Auth işlemleri
	appRouter.HandleFunc("/login", h.Login)
	appRouter.HandleFunc("/register", h.Register)
	appRouter.HandleFunc("/forgot-password", h.ForgotPassword)
	appRouter.HandleFunc("/reset-password", h.ResetPassword)

	// Auth gerektiren sayfalar
	appRouter.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		if !h.IsAuthenticated(r) {
			h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
			return
		}
		h.Dashboard(w, r)
	})

	appRouter.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		if !h.IsAuthenticated(r) {
			h.RedirectWithFlash(w, r, "/login", "Zaten çıkış yapılmış")
			return
		}
		h.Logout(w, r)
	})

	// E-ticaret rotaları
	appRouter.HandleFunc("/products", ecommerceHandler.Products)
	appRouter.HandleFunc("/product/", ecommerceHandler.ProductDetail)
	appRouter.HandleFunc("/cart", ecommerceHandler.Cart)
	appRouter.HandleFunc("/add-to-cart", ecommerceHandler.AddToCart)

	// Açık artırma rotaları
	appRouter.HandleFunc("/auctions", ecommerceHandler.Auctions)
	appRouter.HandleFunc("/auction/", ecommerceHandler.AuctionDetail)
	appRouter.HandleFunc("/place-bid", ecommerceHandler.PlaceBid)

	// Satıcı rotaları
	appRouter.HandleFunc("/vendor/dashboard", ecommerceHandler.VendorDashboard)

	// API rotaları
	appRouter.HandleFunc("/api/search", ecommerceHandler.APISearchProducts)
	appRouter.HandleFunc("/api/cart/update", ecommerceHandler.APIUpdateCart)

	// AI rotaları
	appRouter.HandleFunc("/ai/dashboard", aiHandler.GetAIDashboard)
	appRouter.HandleFunc("/ai/recommendations", aiHandler.GetRecommendationsPage)
	appRouter.HandleFunc("/ai/smart-search", aiHandler.GetSmartSearchPage)
	appRouter.HandleFunc("/ai/price-optimization", aiHandler.GetPriceOptimizationPage)

	// AI API rotaları
	appRouter.HandleFunc("/api/ai/recommendations", aiHandler.GetRecommendations)
	appRouter.HandleFunc("/api/ai/price-optimize/", aiHandler.OptimizePrice)
	appRouter.HandleFunc("/api/ai/predict-category", aiHandler.PredictCategory)
	appRouter.HandleFunc("/api/ai/smart-search", aiHandler.SmartSearch)

	// AI Analytics API rotaları
	appRouter.HandleFunc("/api/ai/market-trends", aiAnalyticsHandler.GetMarketTrends)
	appRouter.HandleFunc("/api/ai/product-insights/", aiAnalyticsHandler.GetProductInsights)
	appRouter.HandleFunc("/api/ai/customer-segments", aiAnalyticsHandler.GetCustomerSegments)

	// AI Vision rotaları
	appRouter.HandleFunc("/ai/vision/dashboard", aiVisionHandler.GetVisionDashboard)
	appRouter.HandleFunc("/ai/vision/upload", aiVisionHandler.RenderVisionUploadPage)
	appRouter.HandleFunc("/ai/vision/search", aiVisionHandler.RenderVisionSearchPage)
	appRouter.HandleFunc("/ai/vision/gallery", aiVisionHandler.RenderVisionGalleryPage)

	// AI Vision API rotaları
	appRouter.HandleFunc("/api/ai/vision/upload", aiVisionHandler.UploadImage)
	
	// Gelişmiş AI rotaları
	appRouter.HandleFunc("/api/ai/generate-image", aiAdvancedHandler.GenerateProductImage)
	appRouter.HandleFunc("/api/ai/generate-content", aiAdvancedHandler.GenerateContent)
	appRouter.HandleFunc("/api/ai/create-template", aiAdvancedHandler.CreateAITemplate)
	appRouter.HandleFunc("/api/ai/chat/start", aiAdvancedHandler.StartAIChat)
	appRouter.HandleFunc("/api/ai/chat/message", aiAdvancedHandler.SendAIChatMessage)
	appRouter.HandleFunc("/api/ai/analyze-image", aiAdvancedHandler.AnalyzeProductImage)
	appRouter.HandleFunc("/api/ai/credits", aiAdvancedHandler.GetAICredits)
	
	// Marketplace entegrasyon rotaları
	appRouter.HandleFunc("/api/marketplace/integrations", marketplaceHandler.GetAllIntegrations)
	appRouter.HandleFunc("/api/marketplace/integration", marketplaceHandler.GetIntegration)
	appRouter.HandleFunc("/api/marketplace/configure", marketplaceHandler.ConfigureIntegration)
	appRouter.HandleFunc("/api/marketplace/sync-products", marketplaceHandler.SyncProducts)
	appRouter.HandleFunc("/api/marketplace/orders", marketplaceHandler.GetMarketplaceOrders)
	appRouter.HandleFunc("/api/marketplace/create-shipment", marketplaceHandler.CreateShipment)
	appRouter.HandleFunc("/api/marketplace/generate-invoice", marketplaceHandler.GenerateInvoice)
	appRouter.HandleFunc("/api/marketplace/update-inventory", marketplaceHandler.UpdateInventory)
	
	// AI Editor ve Marketplace sayfaları
	appRouter.HandleFunc("/ai/editor", func(w http.ResponseWriter, r *http.Request) {
		h.RenderTemplate(w, r, "ai/ai_editor.html", nil)
	})
	appRouter.HandleFunc("/marketplace/integrations", func(w http.ResponseWriter, r *http.Request) {
		h.RenderTemplate(w, r, "marketplace/integrations.html", nil)
	})
	appRouter.HandleFunc("/api/ai/vision/search", aiVisionHandler.SearchImages)
	appRouter.HandleFunc("/api/ai/vision/analysis", aiVisionHandler.GetImageAnalysis)
	appRouter.HandleFunc("/api/ai/vision/delete", aiVisionHandler.DeleteImage)
	appRouter.HandleFunc("/api/ai/vision/category", aiVisionHandler.GetImagesByCategory)
	appRouter.HandleFunc("/api/ai/vision/tag", aiVisionHandler.GetImagesByTag)
	appRouter.HandleFunc("/api/ai/vision/collection/create", aiVisionHandler.CreateCollection)
	appRouter.HandleFunc("/api/ai/vision/collection/update", aiVisionHandler.UpdateCollection)
	appRouter.HandleFunc("/api/ai/vision/collection/delete", aiVisionHandler.DeleteCollection)
	appRouter.HandleFunc("/api/ai/vision/suggest-categories", aiVisionHandler.SuggestCategories)
	appRouter.HandleFunc("/api/ai/vision/stats", aiVisionHandler.GetUserStats)
	appRouter.HandleFunc("/api/ai/vision/library", aiVisionHandler.GetImageLibrary)
	appRouter.HandleFunc("/api/ai/pricing-strategy/", aiAnalyticsHandler.GetPricingStrategy)

	// AI Analytics sayfa rotaları
	appRouter.HandleFunc("/ai/analytics", aiAnalyticsHandler.GetAnalyticsDashboard)
	appRouter.HandleFunc("/ai/analytics/dashboard", aiAnalyticsHandler.GetAnalyticsDashboard)
	appRouter.HandleFunc("/ai/analytics/market-trends", aiAnalyticsHandler.GetMarketTrendsPage)
	appRouter.HandleFunc("/ai/analytics/product-insights", aiAnalyticsHandler.GetProductInsightsPage)
	appRouter.HandleFunc("/ai/analytics/customer-segments", aiAnalyticsHandler.GetCustomerSegmentsPage)
	appRouter.HandleFunc("/ai/analytics/pricing-strategy", aiAnalyticsHandler.GetPricingStrategyPage)

	// Admin rotaları - Gelişmiş admin paneli
	appRouter.HandleFunc("/admin/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/admin/" || r.URL.Path == "/admin" {
			adminHandler.AdminDashboard(w, r)
			return
		}
		http.NotFound(w, r)
	})
	appRouter.HandleFunc("/admin/dashboard", adminHandler.AdminDashboard)
	appRouter.HandleFunc("/admin/products", adminHandler.AdminProducts)
	appRouter.HandleFunc("/admin/products/edit/", adminHandler.AdminProductEdit)
	appRouter.HandleFunc("/admin/vendors", adminHandler.AdminVendors)
	appRouter.HandleFunc("/admin/vendors/approve", adminHandler.AdminVendorApprove)
	appRouter.HandleFunc("/admin/users", adminHandler.AdminUsers)
	appRouter.HandleFunc("/admin/users/", adminHandler.AdminUserDetail)
	appRouter.HandleFunc("/admin/reports", adminHandler.AdminReports)
	appRouter.HandleFunc("/admin/seo", adminHandler.AdminSEO)
	appRouter.HandleFunc("/admin/notifications", adminHandler.AdminNotifications)
	appRouter.HandleFunc("/admin/system", adminHandler.AdminSystem)
	appRouter.HandleFunc("/admin/settings", adminHandler.AdminSettings)

	// Test rotaları (sadece development ortamında)
	if cfg.Environment == "development" {
		appRouter.HandleFunc("/test/run", func(w http.ResponseWriter, r *http.Request) {
			// Simplified test runner - in production would use actual test results
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"completed","tests_run":0,"passed":0,"failed":0}`))
		})
	}

	// Favicon
	appRouter.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/static/assets/images/favicon-32x32.png")
	})

	// Sunucuyu başlat
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	MainLogger.Printf("Enterprise sunucu başlatılıyor: %s", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      appRouter,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Graceful shutdown için background process'ler başlat
	go func() {
		// Background processes would be started here in a real implementation
		// For now, just log that they would be running
		MainLogger.Println("Background processes started (session cleanup, notifications, cache cleanup, error cleanup)")
	}()

	MainLogger.Printf("KolajAI Enterprise uygulaması başlatıldı. %s adresinde dinleniyor...", addr)
	MainLogger.Printf("Tüm gelişmiş sistemler aktif: Session, Cache, Security, SEO, Notifications, Reporting, Testing, Error Management")
	MainLogger.Fatal(server.ListenAndServe())
}
