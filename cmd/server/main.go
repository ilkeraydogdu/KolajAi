package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"kolajAi/internal/database"
	"kolajAi/internal/database/migrations"
	"kolajAi/internal/handlers"
	"kolajAi/internal/services"
	_ "github.com/mattn/go-sqlite3"
)

var (
	MainLogger *log.Logger
)

func init() {
	// Ana uygulama için log dosyası oluştur
	logFile, err := os.OpenFile("main_app_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("Ana uygulama log dosyası oluşturulamadı:", err)
		MainLogger = log.New(os.Stdout, "[MAIN-APP-DEBUG] ", log.LstdFlags)
	} else {
		MainLogger = log.New(logFile, "[MAIN-APP-DEBUG] ", log.LstdFlags|log.Lshortfile)
	}
}

func main() {
	MainLogger.Println("KolajAI uygulaması başlatılıyor...")

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

	// Repository oluştur
	mysqlRepo := database.NewMySQLRepository(db)
	repo := database.NewRepositoryWrapper(mysqlRepo)

	// Servisleri oluştur
	MainLogger.Println("Servisler oluşturuluyor...")
	vendorService := services.NewVendorService(repo)
	productService := services.NewProductService(repo)
	orderService := services.NewOrderService(repo)
	auctionService := services.NewAuctionService(repo)

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
	}
	
	tmpl, err := template.New("").Funcs(funcMap).ParseGlob("web/templates/**/*.gohtml")
	if err != nil {
		MainLogger.Fatalf("Şablonlar yüklenemedi: %v", err)
	}
	MainLogger.Printf("Şablonlar başarıyla yüklendi!")

	// Handler'ları oluştur
	MainLogger.Println("Handler'lar oluşturuluyor...")

	// Session manager oluştur - güvenli bir anahtar kullan
	sessionManager := handlers.NewSessionManager("supersecretkey123")

	h := &handlers.Handler{
		Templates:      tmpl,
		SessionManager: sessionManager,
		TemplateContext: map[string]interface{}{
			"AppName": "KolajAI Marketplace",
			"Year":    time.Now().Year(),
		},
	}

	// E-ticaret handler'ı oluştur
	ecommerceHandler := handlers.NewEcommerceHandler(h, vendorService, productService, orderService, auctionService)

	// Router oluştur ve handler'ları ekle
	router := http.NewServeMux()

	// Statik dosyalar
	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Ana sayfa - Marketplace
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		ecommerceHandler.Marketplace(w, r)
	})

	// Auth işlemleri
	router.HandleFunc("/login", h.Login)
	router.HandleFunc("/register", h.Register)
	router.HandleFunc("/forgot-password", h.ForgotPassword)
	router.HandleFunc("/reset-password", h.ResetPassword)

	// Auth gerektiren sayfalar
	router.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		if !h.IsAuthenticated(r) {
			h.RedirectWithFlash(w, r, "/login", "Lütfen önce giriş yapın")
			return
		}
		h.Dashboard(w, r)
	})

	router.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		if !h.IsAuthenticated(r) {
			h.RedirectWithFlash(w, r, "/login", "Zaten çıkış yapılmış")
			return
		}
		h.Logout(w, r)
	})

	// E-ticaret rotaları
	router.HandleFunc("/products", ecommerceHandler.Products)
	router.HandleFunc("/product/", ecommerceHandler.ProductDetail)
	router.HandleFunc("/cart", ecommerceHandler.Cart)
	router.HandleFunc("/add-to-cart", ecommerceHandler.AddToCart)
	
	// Açık artırma rotaları
	router.HandleFunc("/auctions", ecommerceHandler.Auctions)
	router.HandleFunc("/auction/", ecommerceHandler.AuctionDetail)
	router.HandleFunc("/place-bid", ecommerceHandler.PlaceBid)
	
	// Satıcı rotaları
	router.HandleFunc("/vendor/dashboard", ecommerceHandler.VendorDashboard)
	
	// API rotaları
	router.HandleFunc("/api/search", ecommerceHandler.APISearchProducts)
	router.HandleFunc("/api/cart/update", ecommerceHandler.APIUpdateCart)

	// Favicon
	router.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/static/assets/images/favicon-32x32.png")
	})

	// Sunucuyu başlat
	addr := ":8081"
	MainLogger.Printf("Sunucu başlatılıyor: %s", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	MainLogger.Printf("KolajAI uygulaması başlatıldı. %s adresinde dinleniyor...", addr)
	MainLogger.Fatal(server.ListenAndServe())
}
