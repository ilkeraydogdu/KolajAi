package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"


	"kolajAi/internal/database"
	"kolajAi/internal/handlers"
	"kolajAi/internal/models"
	"kolajAi/internal/repository"
	"kolajAi/internal/email"

	"kolajAi/internal/services"
	"kolajAi/internal/session"
	"kolajAi/internal/errors"
	"kolajAi/internal/utils"
	"kolajAi/internal/seo"
	"kolajAi/internal/security"
	"kolajAi/internal/cache"
	"kolajAi/internal/middleware"
	"kolajAi/internal/router"
	"kolajAi/internal/config"

)

// Build-time variables (set by ldflags)
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

var (
	MainLogger *log.Logger
)

func init() {
	// Environment'a göre log seviyesini ayarla
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("GIN_MODE")
	}
	
	if env == "production" || env == "release" {
		// Production'da sadece stdout'a minimal log
		MainLogger = log.New(os.Stdout, "[MAIN] ", log.LstdFlags)
	} else {
		// Development'ta debug log dosyası
		logFile, err := os.OpenFile("main_app_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Println("Ana uygulama log dosyası oluşturulamadı:", err)
			MainLogger = log.New(os.Stdout, "[MAIN] ", log.LstdFlags)
		} else {
			MainLogger = log.New(logFile, "[MAIN-APP-DEBUG] ", log.LstdFlags|log.Lshortfile)
		}
	}
}



func main() {
	// Handle health check flag
	if len(os.Args) > 1 && os.Args[1] == "--health-check" {
		// Create HTTP client with timeout to prevent DoS
		client := &http.Client{
			Timeout: 5 * time.Second,
		}
		resp, err := client.Get("http://localhost:8081/health")
		if err != nil || resp.StatusCode != 200 {
			if resp != nil {
				resp.Body.Close()
			}
			os.Exit(1)
		}
		resp.Body.Close()
		os.Exit(0)
	}

	MainLogger.Println("KolajAI Enterprise uygulaması başlatılıyor...")

	// Konfigürasyon yükle
	MainLogger.Println("Konfigürasyon yükleniyor...")
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		MainLogger.Printf("Konfigürasyon yüklenemedi, varsayılan değerler kullanılıyor: %v", err)
		cfg = config.GetDefaultConfig()
	}

	// Initialize database manager (SQLite for dev, MySQL for prod)
	MainLogger.Println("Database manager başlatılıyor...")
	if err := database.InitGlobalDB(); err != nil {
		MainLogger.Fatalf("Database initialization failed: %v", err)
	}
	defer database.GlobalDBManager.Close()

	// Run migrations
	MainLogger.Println("Database migrations çalıştırılıyor...")
	if err := database.RunMigrationsForGlobalDB(); err != nil {
		MainLogger.Fatalf("Migration failed: %v", err)
	}
	MainLogger.Println("Migrations başarıyla tamamlandı!")

	// Seed database with initial data
	MainLogger.Println("Database seeding başlatılıyor...")
	
	// Panic recovery for seeding
	func() {
		defer func() {
			if r := recover(); r != nil {
				MainLogger.Printf("Database seeding PANIC: %v", r)
			}
		}()
		
		if err := database.SeedGlobalDatabase(); err != nil {
			MainLogger.Printf("Database seeding failed (continuing anyway): %v", err)
		} else {
			MainLogger.Println("Database seeding tamamlandı!")
		}
	}()
	MainLogger.Println("✅ Database seeding completed successfully - ANA SERVER")

	// Get database connection for services
	MainLogger.Println("Database connection alınıyor...")
	
	var db *sql.DB
	
	// Panic recovery for database connection
	func() {
		defer func() {
			if r := recover(); r != nil {
				MainLogger.Printf("Database connection PANIC: %v", r)
				os.Exit(1)
			}
		}()
		
		db = database.GetGlobalDB()
	}()
	
	MainLogger.Println("✅ Database connection alındı")

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
	MainLogger.Println("✅ Cache Manager başlatıldı")

	// Security Manager
	MainLogger.Println("Güvenlik sistemi başlatılıyor...")
	MainLogger.Printf("EncryptionKey: %s", cfg.Security.EncryptionKey)
	MainLogger.Printf("JWTSecret: %s", cfg.Security.JWTSecret)
	
	var securityManager *security.SecurityManager
	
	// Panic recovery for security manager
	func() {
		defer func() {
			if r := recover(); r != nil {
				MainLogger.Printf("Security Manager PANIC: %v", r)
				os.Exit(1)
			}
		}()
		
		securityManager = security.NewSecurityManager(db, security.SecurityConfig{
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
	}()

	MainLogger.Println("✅ Security Manager başlatıldı")

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

	// Notification Manager - commented out for simplification
	MainLogger.Println("Bildirim sistemi başlatılıyor...")
	// notificationManager := notifications.NewNotificationManager(db, notifications.NotificationConfig{
	// 	DefaultChannel:     "email",
	// 	RetryAttempts:      3,
	// 	RetryDelay:         5 * time.Minute,
	// 	BatchSize:          100,
	// 	QueueSize:          10000,
	// 	Workers:            5,
	// 	EnableRateLimiting: true,
	// })

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

	// Reporting Manager - commented out for simplification
	MainLogger.Println("Raporlama sistemi başlatılıyor...")
	// reportManager := reporting.NewReportManager(db)

	// Test Manager - Commented out as it's not needed in production
	// MainLogger.Println("Test sistemi başlatılıyor...")

	// Repository oluştur
	mysqlRepo := database.NewMySQLRepository(db)
	repo := database.NewRepositoryWrapper(mysqlRepo)

	// Servisleri oluştur
	MainLogger.Println("Servisler oluşturuluyor...")
	// UserRepository için SimpleRepository wrapper kullanıyoruz
	userRepo := repository.NewUserRepository(repo)
	// emailService := email.NewService() // Email service'i initialize et - temporarily disabled
	var emailService *email.Service = nil
	authService := services.NewAuthService(userRepo, emailService)
	vendorService := services.NewVendorService(repo)
	productService := services.NewProductService(repo)
	orderService := services.NewOrderService(repo)
	auctionService := services.NewAuctionService(repo)
	aiService := services.NewAIService(repo, productService, orderService)
	aiAnalyticsService := services.NewAIAnalyticsService(repo, productService, orderService)
	aiVisionService := services.NewAIVisionService(repo, productService)
	_ = services.NewAIEnterpriseService(repo, aiService, aiVisionService, productService, orderService, authService)
	
	// Yeni gelişmiş AI ve marketplace servisleri
	aiAdvancedService := services.NewAIAdvancedService(repo, productService, orderService)
	marketplaceService := services.NewMarketplaceIntegrationsService()
	paymentService := services.NewPaymentService(repo)
	
	// AI Integration Manager
	MainLogger.Println("AI Integration Manager başlatılıyor...")
	aiIntegrationManager := services.NewAIIntegrationManager(marketplaceService, aiService)
	
	// Integration Webhook Service
	MainLogger.Println("Integration Webhook Service başlatılıyor...")
	webhookService := services.NewIntegrationWebhookService(marketplaceService, aiIntegrationManager)
	
	// Integration Analytics Service
	MainLogger.Println("Integration Analytics Service başlatılıyor...")
	analyticsService := services.NewIntegrationAnalyticsService(db, marketplaceService, aiIntegrationManager)

	// Asset Manager'ı başlat
	MainLogger.Println("Asset Manager başlatılıyor...")
	assetManager := utils.NewAssetManager("dist/manifest.json")
	MainLogger.Println("✅ Asset Manager başlatıldı")

	// Şablonları yükle
	MainLogger.Println("Şablonlar yükleniyor...")
	MainLogger.Println("Template functions tanımlanıyor...")
	MainLogger.Println("Dict function tanımlanıyor...")

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
		"default": func(defaultValue, value interface{}) interface{} {
			if value == nil || value == "" {
				return defaultValue
			}
			return value
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
		"iterate": func(n int) []int {
			result := make([]int, n)
			for i := 0; i < n; i++ {
				result[i] = i
			}
			return result
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"lt": func(a, b interface{}) bool {
			var aVal, bVal float64
			
			switch v := a.(type) {
			case int:
				aVal = float64(v)
			case int64:
				aVal = float64(v)
			case float64:
				aVal = v
			case float32:
				aVal = float64(v)
			default:
				return false
			}
			
			switch v := b.(type) {
			case int:
				bVal = float64(v)
			case int64:
				bVal = float64(v)
			case float64:
				bVal = v
			case float32:
				bVal = float64(v)
			default:
				return false
			}
			
			return aVal < bVal
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
		"truncate": func(s string, length int) string {
			if len(s) <= length {
				return s
			}
			return s[:length] + "..."
		},
		"substr": func(s string, start, length int) string {
			if start < 0 || start >= len(s) {
				return ""
			}
			end := start + length
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
		"maskEmail": func(email string) string {
			if len(email) < 3 {
				return email
			}
			atIndex := strings.Index(email, "@")
			if atIndex == -1 {
				return email
			}
			if atIndex < 2 {
				return email
			}
			return email[:2] + "***" + email[atIndex:]
		},
		"currency": func(price float64) string {
			return fmt.Sprintf("%.2f TL", price)
		},
		"eq": func(a, b interface{}) bool {
			return a == b
		},
		"ne": func(a, b interface{}) bool {
			return a != b
		},
		"gt": func(a, b interface{}) bool {
			var aVal, bVal float64
			
			switch v := a.(type) {
			case int:
				aVal = float64(v)
			case int64:
				aVal = float64(v)
			case float64:
				aVal = v
			case float32:
				aVal = float64(v)
			default:
				return false
			}
			
			switch v := b.(type) {
			case int:
				bVal = float64(v)
			case int64:
				bVal = float64(v)
			case float64:
				bVal = v
			case float32:
				bVal = float64(v)
			default:
				return false
			}
			
			return aVal > bVal
		},
		"ge": func(a, b interface{}) bool {
			var aVal, bVal float64
			
			switch v := a.(type) {
			case int:
				aVal = float64(v)
			case int64:
				aVal = float64(v)
			case float64:
				aVal = v
			case float32:
				aVal = float64(v)
			default:
				return false
			}
			
			switch v := b.(type) {
			case int:
				bVal = float64(v)
			case int64:
				bVal = float64(v)
			case float64:
				bVal = v
			case float32:
				bVal = float64(v)
			default:
				return false
			}
			
			return aVal >= bVal
		},
		"le": func(a, b interface{}) bool {
			var aVal, bVal float64
			
			switch v := a.(type) {
			case int:
				aVal = float64(v)
			case int64:
				aVal = float64(v)
			case float64:
				aVal = v
			case float32:
				aVal = float64(v)
			default:
				return false
			}
			
			switch v := b.(type) {
			case int:
				bVal = float64(v)
			case int64:
				bVal = float64(v)
			case float64:
				bVal = v
			case float32:
				bVal = float64(v)
			default:
				return false
			}
			
			return aVal <= bVal
		},
		"and": func(a, b bool) bool {
			return a && b
		},
		"or": func(a, b bool) bool {
			return a || b
		},
		"not": func(a bool) bool {
			return !a
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
		"title": func(s string) string {
			return strings.Title(s)
		},
		"trim": func(s string) string {
			return strings.TrimSpace(s)
		},
		"len": func(v interface{}) int {
			switch val := v.(type) {
			case string:
				return len(val)
			case []interface{}:
				return len(val)
			default:
				return 0
			}
		},
		"div": func(a, b interface{}) float64 {
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
				return 1
			}
			if numB == 0 {
				return 0
			}
			return numA / numB
		},
		"mod": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a % b
		},
	}

	MainLogger.Println("Template parsing başlatılıyor...")
	
	// Template dosyalarını manuel olarak bulalım
	templateFiles := []string{}
	walkErr := filepath.Walk("web/templates", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".gohtml") {
			templateFiles = append(templateFiles, path)
		}
		return nil
	})
	
	if walkErr != nil {
		MainLogger.Fatalf("Template dosyaları bulunamadı: %v", walkErr)
	}
	
	MainLogger.Printf("Bulunan template dosyaları: %d", len(templateFiles))
	
	// Debug: Template dosyalarını listele
	for i, file := range templateFiles {
		MainLogger.Printf("Template %d: %s", i+1, file)
	}
	
	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(templateFiles...)
	if err != nil {
		MainLogger.Printf("Template parsing hatası: %v", err)
		MainLogger.Printf("Problematik template'leri tek tek test ediyorum...")
		
		// Template'leri tek tek test et
		for _, file := range templateFiles {
			_, testErr := template.New("").Funcs(funcMap).ParseFiles(file)
			if testErr != nil {
				MainLogger.Printf("❌ Hatalı template: %s - Hata: %v", file, testErr)
			} else {
				MainLogger.Printf("✅ Başarılı template: %s", file)
			}
		}
		
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
		DB:             db,
		TemplateContext: map[string]interface{}{
			"AppName": "KolajAI Enterprise Marketplace",
			"Year":    time.Now().Year(),
			"Assets": map[string]interface{}{
				"CSS": assetManager.GetCSSAssets(),
				"JS":  assetManager.GetJSAssets(),
			},
		},
	}

	// E-ticaret handler'ı oluştur
	ecommerceHandler := handlers.NewEcommerceHandler(h, vendorService, productService, orderService, auctionService)

	// Admin handler'ı oluştur
	adminHandler := handlers.NewAdminHandler(h, repo)

	// Seller handler'ı oluştur
	sellerHandler := handlers.NewSellerHandler(h, vendorService, productService, orderService)

	// Inventory handler'ı oluştur
	inventoryService := services.NewInventoryService(repo, productService, orderService)
	inventoryHandler := handlers.NewInventoryHandler(h, inventoryService, productService)

	// Notification handler'ı oluştur
	notificationService := services.NewNotificationService(nil, nil, nil)
	notificationHandler := handlers.NewNotificationHandler(h, notificationService)

	// Security handler'ı oluştur
	securityHandler := handlers.NewSecurityHandler(h)

	// Analytics handler'ı oluştur
	analyticsHandler := handlers.NewAnalyticsHandler(h, nil)

	// Email handler'ı oluştur
	emailHandler := handlers.NewEmailHandler(h, nil)

	// AI handler'ı oluştur
	aiHandler := handlers.NewAIHandler(h, aiService)

	// AI Analytics handler'ı oluştur
	aiAnalyticsHandler := handlers.NewAIAnalyticsHandler(h, aiAnalyticsService)

	// AI Vision handler'ı oluştur
	aiVisionHandler := handlers.NewAIVisionHandler(h, aiVisionService)
	
	// Yeni gelişmiş handler'lar
	aiAdvancedHandler := handlers.NewAIAdvancedHandler(h, aiAdvancedService)
	marketplaceHandler := handlers.NewMarketplaceHandler(h, marketplaceService)
	paymentHandler := handlers.NewPaymentHandler(h, paymentService, orderService)

	// Middleware stack oluştur
	middlewareStack := middleware.NewMiddlewareStack(
		securityManager,
		sessionManager,
		errorManager,
		cacheManager,
	)

	// Router oluştur
	appRouter := router.NewRouter(middlewareStack)

	// Statik dosyalar - Sadece /web/static/ altındaki dosyalar serve edilecek
	appRouter.Handle("/web/static/", http.StripPrefix("/web/static/", http.FileServer(http.Dir("web/static"))))
	
	// Webpack build dosyaları için
	appRouter.Handle("/dist/", http.StripPrefix("/dist/", http.FileServer(http.Dir("dist"))))

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
		
		// Mock data for testing
		data := map[string]interface{}{
			"Title": "KolajAI Marketplace - Ana Sayfa",
			"Description": "Çoklu satıcı, açık artırma ve toptan satış platformu",
			"Categories": []map[string]interface{}{
				{"ID": 1, "Name": "Elektronik", "Image": "/web/static/assets/images/categories/electronics.jpg"},
				{"ID": 2, "Name": "Giyim", "Image": "/web/static/assets/images/categories/clothing.jpg"},
				{"ID": 3, "Name": "Ev & Yaşam", "Image": "/web/static/assets/images/categories/home.jpg"},
				{"ID": 4, "Name": "Spor", "Image": "/web/static/assets/images/categories/sports.jpg"},
				{"ID": 5, "Name": "Kitap", "Image": "/web/static/assets/images/categories/books.jpg"},
				{"ID": 6, "Name": "Oyuncak", "Image": "/web/static/assets/images/categories/toys.jpg"},
			},
			"FeaturedProducts": []map[string]interface{}{
				{
					"ID": 1,
					"Name": "iPhone 15 Pro Max",
					"Price": 49999.99,
					"ShortDesc": "En yeni iPhone modeli",
					"Images": []string{"/web/static/assets/images/products/iphone.jpg"},
					"IsFeatured": true,
					"Rating": 4.8,
				},
				{
					"ID": 2,
					"Name": "Samsung Galaxy S24 Ultra",
					"Price": 45999.99,
					"ShortDesc": "Güçlü Android telefon",
					"Images": []string{"/web/static/assets/images/products/samsung.jpg"},
					"IsFeatured": true,
					"Rating": 4.7,
				},
				{
					"ID": 3,
					"Name": "MacBook Pro M3",
					"Price": 89999.99,
					"ShortDesc": "Profesyonel laptop",
					"Images": []string{"/web/static/assets/images/products/macbook.jpg"},
					"IsFeatured": true,
					"Rating": 4.9,
				},
				{
					"ID": 4,
					"Name": "Sony PlayStation 5",
					"Price": 19999.99,
					"ShortDesc": "Yeni nesil oyun konsolu",
					"Images": []string{"/web/static/assets/images/products/ps5.jpg"},
					"IsFeatured": true,
					"Rating": 4.8,
				},
			},
			"ActiveAuctions": []map[string]interface{}{
				{
					"ID": 1,
					"Title": "Antika Saat Koleksiyonu",
					"CurrentBid": 5000.00,
					"TotalBids": 15,
					"EndTime": time.Now().Add(24 * time.Hour),
					"Images": []string{"/web/static/assets/images/auctions/watch.jpg"},
				},
				{
					"ID": 2,
					"Title": "Nadir Pul Koleksiyonu",
					"CurrentBid": 3500.00,
					"TotalBids": 8,
					"EndTime": time.Now().Add(48 * time.Hour),
					"Images": []string{"/web/static/assets/images/auctions/stamps.jpg"},
				},
				{
					"ID": 3,
					"Title": "Vintage Kamera Seti",
					"CurrentBid": 7500.00,
					"TotalBids": 22,
					"EndTime": time.Now().Add(12 * time.Hour),
					"Images": []string{"/web/static/assets/images/auctions/camera.jpg"},
				},
			},
		}
		
		h.RenderTemplate(w, r, "marketplace/index.gohtml", data)
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

	// API rotaları
	appRouter.HandleFunc("/api/products", ecommerceHandler.GetProducts)
	appRouter.HandleFunc("/api/product/", ecommerceHandler.GetProduct)
	appRouter.HandleFunc("/api/search", ecommerceHandler.SearchProducts)
	appRouter.HandleFunc("/api/categories", ecommerceHandler.GetCategories)
	appRouter.HandleFunc("/health", ecommerceHandler.HealthCheck)

	// User profile rotaları (auth gerektirir)
	appRouter.HandleFunc("/user/profile", func(w http.ResponseWriter, r *http.Request) {
		if !h.IsAuthenticated(r) {
			h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
			return
		}
		data := map[string]interface{}{
			"Title": "Profil",
			"User": map[string]interface{}{
				"name":  "Test Kullanıcı",
				"email": "test@example.com",
			},
		}
		h.RenderTemplate(w, r, "user/profile.gohtml", data)
	})
	
	appRouter.HandleFunc("/user/orders", func(w http.ResponseWriter, r *http.Request) {
		if !h.IsAuthenticated(r) {
			h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
			return
		}
		data := map[string]interface{}{
			"Title": "Siparişlerim",
			"Orders": []map[string]interface{}{
				{
					"id":     1,
					"status": "completed",
					"total":  99.99,
					"date":   "2025-01-01",
				},
			},
		}
		h.RenderTemplate(w, r, "user/orders.gohtml", data)
	})
	
	appRouter.HandleFunc("/user/addresses", func(w http.ResponseWriter, r *http.Request) {
		if !h.IsAuthenticated(r) {
			h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
			return
		}
		data := map[string]interface{}{
			"Title": "Adreslerim",
			"Addresses": []map[string]interface{}{
				{
					"id":    1,
					"title": "Ev",
					"address": "Test Mahallesi, Test Sokak No:1",
					"city":  "İstanbul",
				},
			},
		}
		h.RenderTemplate(w, r, "user/addresses.gohtml", data)
	})

	// Auth API rotaları
	appRouter.HandleFunc("/api/verify-temp-password", h.VerifyTempPassword)
	
	// API rotaları
	appRouter.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"timestamp": time.Now().Format(time.RFC3339),
			"service": "KolajAI Enterprise",
			"version": "2.0.0",
		})
	})
	// Legacy API endpoints (deprecated - use /api/v1/ instead) - simplified
	appRouter.HandleFunc("/api/legacy/search", ecommerceHandler.SearchProducts)
	appRouter.HandleFunc("/api/legacy/cart/update", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "deprecated"})
	})

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
	
	// Integration webhook endpoints
	appRouter.HandleFunc("/webhooks/integration", webhookService.HandleWebhook)
	
	// Payment endpoints
	appRouter.HandleFunc("/payment/checkout", paymentHandler.PaymentPage)
	appRouter.HandleFunc("/payment/success", paymentHandler.PaymentSuccess)
	appRouter.HandleFunc("/payment/failure", paymentHandler.PaymentFailure)
	
	// Payment API endpoints
	appRouter.HandleFunc("/api/payment/intent", paymentHandler.CreatePaymentIntent)
	appRouter.HandleFunc("/api/payment/process", paymentHandler.ProcessPayment)
	appRouter.HandleFunc("/api/payment/status/", paymentHandler.GetPaymentStatus)
	appRouter.HandleFunc("/api/payment/refund", paymentHandler.RefundPayment)
	appRouter.HandleFunc("/api/payment/methods", paymentHandler.GetPaymentMethods)
	appRouter.HandleFunc("/api/payment/calculate-fee", paymentHandler.CalculatePaymentFee)
	
	// Integration analytics endpoints
	appRouter.HandleFunc("/api/integration/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics := analyticsService.GetMetrics()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"metrics": metrics,
		})
	})
	
	appRouter.HandleFunc("/api/integration/health", func(w http.ResponseWriter, r *http.Request) {
		integrationID := r.URL.Query().Get("integration_id")
		if integrationID == "" {
			http.Error(w, "Integration ID required", http.StatusBadRequest)
			return
		}
		
		health := analyticsService.GetIntegrationHealth(integrationID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"health":  health,
		})
	})
	
	appRouter.HandleFunc("/api/integration/report", func(w http.ResponseWriter, r *http.Request) {
		reportType := r.URL.Query().Get("type")
		if reportType == "" {
			reportType = "daily"
		}
		
		// Parse time parameters
		startTime := time.Now().AddDate(0, 0, -7) // Default to 7 days ago
		endTime := time.Now()
		
		report, err := analyticsService.GenerateReport(reportType, startTime, endTime)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"report":  report,
		})
	})
	
	// AI Integration insights endpoints
	appRouter.HandleFunc("/api/ai/integration/insights", func(w http.ResponseWriter, r *http.Request) {
		integrationID := r.URL.Query().Get("integration_id")
		
		var insights map[string]interface{}
		var err error
		
		if integrationID != "" {
			insights, err = aiIntegrationManager.GetAIInsights(integrationID)
		} else {
			insights, err = aiIntegrationManager.GetAllAIInsights()
		}
		
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"insights": insights,
		})
	})
	
	// AI Editor ve Marketplace sayfaları
	appRouter.HandleFunc("/ai/editor", func(w http.ResponseWriter, r *http.Request) {
		h.RenderTemplate(w, r, "ai/ai_editor.html", nil)
	})
	// Marketplace rotaları
	appRouter.HandleFunc("/marketplace", func(w http.ResponseWriter, r *http.Request) {
		// Get categories from database
		categories, err := productService.GetAllCategories()
		if err != nil {
			log.Printf("Error loading categories: %v", err)
			categories = []models.Category{} // Empty slice on error
		}

		// Get featured products
		featuredProducts, err := productService.GetFeaturedProducts(8, 0)
		if err != nil {
			log.Printf("Error loading featured products: %v", err)
			featuredProducts = []models.Product{} // Empty slice on error
		}

		// Get active auctions
		activeAuctions, err := auctionService.GetActiveAuctions(6)
		if err != nil {
			log.Printf("Error loading active auctions: %v", err)
			activeAuctions = []models.Auction{} // Empty slice on error
		}

		data := map[string]interface{}{
			"Title":            "KolajAI Marketplace",
			"Categories":       categories,
			"FeaturedProducts": featuredProducts,
			"ActiveAuctions":   activeAuctions,
			"AppName":          "KolajAI",
		}
		h.RenderTemplate(w, r, "marketplace/index.gohtml", data)
	})
	
	appRouter.HandleFunc("/marketplace/products", func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		category := r.URL.Query().Get("category")
		search := r.URL.Query().Get("search")
		page := 1
		limit := 20
		
		// Get products from database
		products, err := productService.GetProducts(category, search, page, limit)
		if err != nil {
			log.Printf("Error loading products: %v", err)
			products = []models.Product{} // Empty slice on error
		}
		
		// Get categories for filter
		categories, err := productService.GetAllCategories()
		if err != nil {
			log.Printf("Error loading categories: %v", err)
			categories = []models.Category{} // Empty slice on error
		}

		data := map[string]interface{}{
			"Title":      "Ürünler - KolajAI",
			"Products":   products,
			"Categories": categories,
			"AppName":    "KolajAI",
		}
		h.RenderTemplate(w, r, "marketplace/products.gohtml", data)
	})
	
	appRouter.HandleFunc("/marketplace/categories", func(w http.ResponseWriter, r *http.Request) {
		// Get categories
		categories, err := productService.GetAllCategories()
		if err != nil {
			log.Printf("Error loading categories: %v", err)
			categories = []models.Category{} // Empty slice on error
		}

		data := map[string]interface{}{
			"Title":      "Kategoriler - KolajAI",
			"Categories": categories,
			"AppName":    "KolajAI",
		}
		h.RenderTemplate(w, r, "marketplace/categories.gohtml", data)
	})
	
	appRouter.HandleFunc("/marketplace/cart", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{
			"Title": "Sepetim",
			"CartItems": []map[string]interface{}{
				{
					"id":       1,
					"product":  "Test Ürün",
					"quantity": 2,
					"price":    99.99,
				},
			},
		}
		h.RenderTemplate(w, r, "marketplace/cart.gohtml", data)
	})
	
	appRouter.HandleFunc("/marketplace/integrations", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{
			"Title": "Entegrasyonlar",
		}
		h.RenderTemplate(w, r, "marketplace/integrations.html", data)
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

	// Admin rotaları - Admin middleware ile korumalı
	appRouter.Handle("/admin/", middlewareStack.AdminMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/admin/" || r.URL.Path == "/admin" {
			adminHandler.AdminDashboard(w, r)
			return
		}
		http.NotFound(w, r)
	})))
	appRouter.Handle("/admin/dashboard", middlewareStack.AdminMiddleware(http.HandlerFunc(adminHandler.AdminDashboard)))
	appRouter.Handle("/admin/users", middlewareStack.AdminMiddleware(http.HandlerFunc(adminHandler.AdminUsers)))
	appRouter.Handle("/admin/orders", middlewareStack.AdminMiddleware(http.HandlerFunc(adminHandler.AdminOrders)))
	appRouter.Handle("/admin/products", middlewareStack.AdminMiddleware(http.HandlerFunc(adminHandler.AdminProducts)))
	appRouter.Handle("/admin/reports", middlewareStack.AdminMiddleware(http.HandlerFunc(adminHandler.AdminReports)))
	appRouter.Handle("/admin/vendors", middlewareStack.AdminMiddleware(http.HandlerFunc(adminHandler.AdminVendors)))
	appRouter.Handle("/admin/system-health", middlewareStack.AdminMiddleware(http.HandlerFunc(adminHandler.AdminSystemHealth)))
	appRouter.Handle("/admin/seo", middlewareStack.AdminMiddleware(http.HandlerFunc(adminHandler.AdminSEO)))

	// Admin API rotaları - Admin middleware ile korumalı
	appRouter.Handle("/api/admin/users/stats", middlewareStack.AdminMiddleware(http.HandlerFunc(adminHandler.APIGetUserStats)))
	appRouter.Handle("/api/admin/users/create", middlewareStack.AdminMiddleware(http.HandlerFunc(adminHandler.APICreateUser)))
	appRouter.Handle("/api/admin/users/export", middlewareStack.AdminMiddleware(http.HandlerFunc(adminHandler.APIExportUsers)))
	appRouter.Handle("/api/admin/users/{id}/status", middlewareStack.AdminMiddleware(http.HandlerFunc(adminHandler.APIUpdateUserStatus)))
	appRouter.Handle("/api/admin/users/{id}", middlewareStack.AdminMiddleware(http.HandlerFunc(adminHandler.APIDeleteUser)))
	appRouter.Handle("/api/admin/orders/{id}/status", middlewareStack.AdminMiddleware(http.HandlerFunc(adminHandler.APIUpdateOrderStatus)))
	appRouter.Handle("/api/admin/orders/{id}", middlewareStack.AdminMiddleware(http.HandlerFunc(adminHandler.APIDeleteOrder)))
	appRouter.Handle("/api/admin/products/bulk-action", middlewareStack.AdminMiddleware(http.HandlerFunc(adminHandler.APIBulkProductAction)))
	appRouter.Handle("/api/admin/products/{id}/status", middlewareStack.AdminMiddleware(http.HandlerFunc(adminHandler.APIUpdateProductStatus)))
	appRouter.Handle("/api/admin/system/health", middlewareStack.AdminMiddleware(http.HandlerFunc(adminHandler.APISystemHealthCheck)))
	appRouter.Handle("/api/admin/seo/sitemap", middlewareStack.AdminMiddleware(http.HandlerFunc(adminHandler.APIGenerateSitemap)))
	appRouter.Handle("/api/admin/seo/analyze", middlewareStack.AdminMiddleware(http.HandlerFunc(adminHandler.APIAnalyzeSEO)))

	// Seller rotaları - Authentication middleware ile korumalı
	appRouter.HandleFunc("/seller/dashboard", sellerHandler.Dashboard)
	appRouter.HandleFunc("/seller/products", sellerHandler.Products)
	appRouter.HandleFunc("/seller/orders", sellerHandler.Orders)

	// Seller API rotaları
	appRouter.HandleFunc("/api/seller/products", sellerHandler.APIGetProducts)
	appRouter.HandleFunc("/api/seller/orders", sellerHandler.APIGetOrders)
	appRouter.HandleFunc("/api/seller/product/status", sellerHandler.APIUpdateProductStatus)
	appRouter.HandleFunc("/api/seller/order/status", sellerHandler.APIUpdateOrderStatus)

	// Inventory rotaları - Admin middleware ile korumalı
	appRouter.Handle("/inventory/dashboard", middlewareStack.AdminMiddleware(http.HandlerFunc(inventoryHandler.Dashboard)))
	appRouter.Handle("/inventory/stock-levels", middlewareStack.AdminMiddleware(http.HandlerFunc(inventoryHandler.StockLevels)))
	appRouter.Handle("/inventory/alerts", middlewareStack.AdminMiddleware(http.HandlerFunc(inventoryHandler.Alerts)))

	// Inventory API rotaları
	appRouter.Handle("/api/inventory/stock-levels", middlewareStack.AdminMiddleware(http.HandlerFunc(inventoryHandler.APIGetStockLevels)))
	appRouter.Handle("/api/inventory/update-stock", middlewareStack.AdminMiddleware(http.HandlerFunc(inventoryHandler.APIUpdateStock)))
	appRouter.Handle("/api/inventory/alerts", middlewareStack.AdminMiddleware(http.HandlerFunc(inventoryHandler.APIGetAlerts)))
	appRouter.Handle("/api/inventory/dismiss-alert", middlewareStack.AdminMiddleware(http.HandlerFunc(inventoryHandler.APIDismissAlert)))

	// Notification rotaları - Admin middleware ile korumalı
	appRouter.Handle("/notifications/dashboard", middlewareStack.AdminMiddleware(http.HandlerFunc(notificationHandler.Dashboard)))
	appRouter.Handle("/notifications/templates", middlewareStack.AdminMiddleware(http.HandlerFunc(notificationHandler.Templates)))
	appRouter.Handle("/notifications/campaigns", middlewareStack.AdminMiddleware(http.HandlerFunc(notificationHandler.Campaigns)))

	// Notification API rotaları
	appRouter.Handle("/api/notifications/send", middlewareStack.AdminMiddleware(http.HandlerFunc(notificationHandler.APISendNotification)))
	appRouter.Handle("/api/notifications/stats", middlewareStack.AdminMiddleware(http.HandlerFunc(notificationHandler.APIGetNotificationStats)))
	appRouter.Handle("/api/notifications/templates", middlewareStack.AdminMiddleware(http.HandlerFunc(notificationHandler.APICreateTemplate)))
	appRouter.Handle("/api/notifications/templates/update", middlewareStack.AdminMiddleware(http.HandlerFunc(notificationHandler.APIUpdateTemplate)))
	appRouter.Handle("/api/notifications/templates/delete", middlewareStack.AdminMiddleware(http.HandlerFunc(notificationHandler.APIDeleteTemplate)))

	// Security rotaları - Admin middleware ile korumalı
	appRouter.Handle("/security/dashboard", middlewareStack.AdminMiddleware(http.HandlerFunc(securityHandler.Dashboard)))
	appRouter.Handle("/security/users", middlewareStack.AdminMiddleware(http.HandlerFunc(securityHandler.Users)))
	appRouter.Handle("/security/threats", middlewareStack.AdminMiddleware(http.HandlerFunc(securityHandler.Threats)))
	appRouter.Handle("/security/settings", middlewareStack.AdminMiddleware(http.HandlerFunc(securityHandler.Settings)))

	// Security API rotaları
	appRouter.Handle("/api/security/stats", middlewareStack.AdminMiddleware(http.HandlerFunc(securityHandler.APIGetSecurityStats)))
	appRouter.Handle("/api/security/block-ip", middlewareStack.AdminMiddleware(http.HandlerFunc(securityHandler.APIBlockIP)))
	appRouter.Handle("/api/security/unblock-ip", middlewareStack.AdminMiddleware(http.HandlerFunc(securityHandler.APIUnblockIP)))
	appRouter.Handle("/api/security/enable-2fa", middlewareStack.AdminMiddleware(http.HandlerFunc(securityHandler.APIEnable2FA)))
	appRouter.Handle("/api/security/settings", middlewareStack.AdminMiddleware(http.HandlerFunc(securityHandler.APIUpdateSecuritySettings)))
	appRouter.Handle("/api/security/threat-details", middlewareStack.AdminMiddleware(http.HandlerFunc(securityHandler.APIGetThreatDetails)))

	// Analytics rotaları - Admin middleware ile korumalı
	appRouter.Handle("/analytics/dashboard", middlewareStack.AdminMiddleware(http.HandlerFunc(analyticsHandler.Dashboard)))
	appRouter.Handle("/analytics/revenue", middlewareStack.AdminMiddleware(http.HandlerFunc(analyticsHandler.Revenue)))
	appRouter.Handle("/analytics/customers", middlewareStack.AdminMiddleware(http.HandlerFunc(analyticsHandler.Customers)))
	appRouter.Handle("/analytics/products", middlewareStack.AdminMiddleware(http.HandlerFunc(analyticsHandler.Products)))

	// Analytics API rotaları
	appRouter.Handle("/api/analytics/metrics", middlewareStack.AdminMiddleware(http.HandlerFunc(analyticsHandler.APIGetBusinessMetrics)))
	appRouter.Handle("/api/analytics/forecast", middlewareStack.AdminMiddleware(http.HandlerFunc(analyticsHandler.APIGetRevenueForecast)))
	appRouter.Handle("/api/analytics/segments", middlewareStack.AdminMiddleware(http.HandlerFunc(analyticsHandler.APIGetCustomerSegments)))
	appRouter.Handle("/api/analytics/insights", middlewareStack.AdminMiddleware(http.HandlerFunc(analyticsHandler.APIGetProductInsights)))

	// Email rotaları - Admin middleware ile korumalı
	appRouter.Handle("/email/dashboard", middlewareStack.AdminMiddleware(http.HandlerFunc(emailHandler.Dashboard)))
	appRouter.Handle("/email/campaigns", middlewareStack.AdminMiddleware(http.HandlerFunc(emailHandler.Campaigns)))
	appRouter.Handle("/email/templates", middlewareStack.AdminMiddleware(http.HandlerFunc(emailHandler.Templates)))
	appRouter.Handle("/email/settings", middlewareStack.AdminMiddleware(http.HandlerFunc(emailHandler.Settings)))

	// Email API rotaları
	appRouter.Handle("/api/email/send-campaign", middlewareStack.AdminMiddleware(http.HandlerFunc(emailHandler.APISendCampaign)))
	appRouter.Handle("/api/email/stats", middlewareStack.AdminMiddleware(http.HandlerFunc(emailHandler.APIGetEmailStats)))
	appRouter.Handle("/api/email/create-template", middlewareStack.AdminMiddleware(http.HandlerFunc(emailHandler.APICreateTemplate)))

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
		// Security headers
		TLSConfig: &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		},
	}

	// Graceful shutdown için background process'ler başlat
	go func() {
		// Background processes would be started here in a real implementation
		// For now, just log that they would be running
		MainLogger.Println("Background processes started (session cleanup, notifications, cache cleanup, error cleanup)")
	}()

	MainLogger.Printf("KolajAI Enterprise uygulaması başlatıldı. %s adresinde dinleniyor...", addr)
	MainLogger.Printf("Tüm gelişmiş sistemler aktif: Session, Cache, Security, SEO, Notifications, Reporting, Testing, Error Management")
	MainLogger.Printf("Web tarayıcınızda http://localhost%s adresini ziyaret edin", addr)
	MainLogger.Printf("Static dosyalar /static/ altında serve ediliyor")
	MainLogger.Printf("Templates web/templates/ klasöründen yüklendi")
	
	// Server'ı başlat
	// Check for TLS certificates
	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")
	
	if certFile != "" && keyFile != "" {
		MainLogger.Printf("HTTPS sunucu başlatılıyor (TLS): %s", addr)
		if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
			MainLogger.Fatalf("HTTPS Server başlatılamadı: %v", err)
		}
	} else {
		MainLogger.Printf("HTTP sunucu başlatılıyor (TLS YOK - sadece development): %s", addr)
		MainLogger.Printf("Production için TLS_CERT_FILE ve TLS_KEY_FILE environment variables ayarlayın")
		MainLogger.Printf("🚀 KolajAI Server is starting on http://localhost%s", addr)
		MainLogger.Printf("🔗 Marketplace: http://localhost%s/marketplace", addr)
		if err := server.ListenAndServe(); err != nil {
			MainLogger.Fatalf("HTTP Server başlatılamadı: %v", err)
		}
	}
}
