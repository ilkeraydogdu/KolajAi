package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"kolajAi/internal/database"
	"kolajAi/internal/models"
	"kolajAi/internal/repository"
	"kolajAi/internal/services"
)

func main() {
	log.Println("🚀 KolajAI Minimal Server başlatılıyor...")

	// Initialize database manager (SQLite for dev, MySQL for prod)
	log.Println("Database manager başlatılıyor...")
	if err := database.InitGlobalDB(); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer database.GlobalDBManager.Close()

	// Run migrations
	log.Println("Database migrations çalıştırılıyor...")
	if err := database.RunMigrationsForGlobalDB(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Seed database with initial data
	log.Println("Database seeding başlatılıyor...")
	if err := database.SeedGlobalDatabase(); err != nil {
		log.Printf("Database seeding failed (continuing anyway): %v", err)
	}

	// Get database connection for services
	db := database.GetGlobalDB()

	// Repository ve servisler
	mysqlRepo := database.NewMySQLRepository(db)
	repo := repository.NewBaseRepository(mysqlRepo)
	productService := services.NewProductService(repo)
	auctionService := services.NewAuctionService(repo)

	// Template fonksiyonları
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
		"formatPrice": func(price float64) string {
			return fmt.Sprintf("%.2f TL", price)
		},
		"currency": func(price float64) string {
			return fmt.Sprintf("%.2f TL", price)
		},
		"formatDate": func(t time.Time) string {
			return t.Format("02.01.2006 15:04")
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
		"rand": func() int {
			return time.Now().Nanosecond() % 1000
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
				result[i] = i + 1
			}
			return result
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"lt": func(a, b int) bool {
			return a < b
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
		"add": func(a, b int) int {
			return a + b
		},
		"mul": func(a, b int) int {
			return a * b
		},
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"mod": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a % b
		},
		"eq": func(a, b interface{}) bool {
			return a == b
		},
		"ne": func(a, b interface{}) bool {
			return a != b
		},
		"gt": func(a, b int) bool {
			return a > b
		},
		"ge": func(a, b int) bool {
			return a >= b
		},
		"le": func(a, b int) bool {
			return a <= b
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
	}

	// Template dosyalarını bulalım
	templateFiles := []string{}
	err := filepath.Walk("web/templates", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".gohtml") {
			templateFiles = append(templateFiles, path)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Template dosyaları bulunamadı: %v", err)
	}

	log.Printf("Bulunan template dosyaları: %d", len(templateFiles))

	// Template'leri yükle - sadece index sayfası için gerekli olanlar
	essentialTemplates := []string{}
	for _, file := range templateFiles {
		if strings.Contains(file, "marketplace/index.gohtml") || 
		   strings.Contains(file, "layout/base.gohtml") ||
		   strings.Contains(file, "layout/header.gohtml") ||
		   strings.Contains(file, "layout/footer.gohtml") {
			essentialTemplates = append(essentialTemplates, file)
		}
	}
	
	log.Printf("Essential template dosyaları: %d", len(essentialTemplates))
	
	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(essentialTemplates...)
	if err != nil {
		log.Printf("Template parse hatası: %v", err)
		log.Printf("Essential template dosyaları:")
		for i, file := range essentialTemplates {
			log.Printf("  %d: %s", i+1, file)
		}
		log.Fatalf("Essential şablonları yüklenemedi: %v", err)
	}

	log.Println("✅ Template'ler başarıyla yüklendi!")

	// HTTP handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/marketplace", http.StatusFound)
	})

	http.HandleFunc("/marketplace", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Marketplace sayfası istendi")

		// Get categories from database
		categories, err := productService.GetAllCategories()
		if err != nil {
			log.Printf("Error loading categories: %v", err)
			categories = []models.Category{}
		}

		// Get featured products
		featuredProducts, err := productService.GetFeaturedProducts(8, 0)
		if err != nil {
			log.Printf("Error loading featured products: %v", err)
			featuredProducts = []models.Product{}
		}

		// Get active auctions
		activeAuctions, err := auctionService.GetActiveAuctions(6)
		if err != nil {
			log.Printf("Error loading active auctions: %v", err)
			activeAuctions = []models.Auction{}
		}

		data := map[string]interface{}{
			"Title":            "KolajAI Marketplace",
			"Categories":       categories,
			"FeaturedProducts": featuredProducts,
			"ActiveAuctions":   activeAuctions,
			"AppName":          "KolajAI",
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = tmpl.ExecuteTemplate(w, "marketplace/index", data)
		if err != nil {
			log.Printf("Template render hatası: %v", err)
			http.Error(w, fmt.Sprintf("Template rendering error: %v", err), http.StatusInternalServerError)
			return
		}
	})

	// Static files
	http.Handle("/web/static/", http.StripPrefix("/web/static/", http.FileServer(http.Dir("web/static"))))

	// Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start server
	addr := ":8082"
	log.Printf("🚀 KolajAI Server is starting on http://localhost%s", addr)
	log.Printf("🔗 Marketplace: http://localhost%s/marketplace", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("HTTP Server başlatılamadı: %v", err)
	}
}