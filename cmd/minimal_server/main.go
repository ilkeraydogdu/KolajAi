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

	// Template'leri yükle - sadece marketplace template'leri
	marketplaceTemplates := []string{}
	for _, file := range templateFiles {
		if strings.Contains(file, "marketplace") || strings.Contains(file, "layout") || strings.Contains(file, "components") {
			marketplaceTemplates = append(marketplaceTemplates, file)
		}
	}
	
	log.Printf("Marketplace template dosyaları: %d", len(marketplaceTemplates))
	
	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(marketplaceTemplates...)
	if err != nil {
		log.Printf("Template parse hatası: %v", err)
		log.Printf("Marketplace template dosyaları:")
		for i, file := range marketplaceTemplates {
			log.Printf("  %d: %s", i+1, file)
		}
		log.Fatalf("Marketplace şablonları yüklenemedi: %v", err)
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
		err = tmpl.ExecuteTemplate(w, "marketplace/index.gohtml", data)
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
	addr := ":8081"
	log.Printf("🚀 KolajAI Server is starting on http://localhost%s", addr)
	log.Printf("🔗 Marketplace: http://localhost%s/marketplace", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("HTTP Server başlatılamadı: %v", err)
	}
}