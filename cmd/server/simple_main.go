package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
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

// Sample data structures
type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Image       string  `json:"image"`
	Category    string  `json:"category"`
	Rating      float64 `json:"rating"`
	Stock       int     `json:"stock"`
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Sample data
var (
	sampleProducts = []Product{
		{ID: 1, Name: "Akıllı Telefon", Description: "Son teknoloji akıllı telefon", Price: 5999.99, Image: "/static/images/phone.jpg", Category: "Elektronik", Rating: 4.5, Stock: 25},
		{ID: 2, Name: "Laptop", Description: "Gaming laptop", Price: 12999.99, Image: "/static/images/laptop.jpg", Category: "Bilgisayar", Rating: 4.8, Stock: 15},
		{ID: 3, Name: "Kulaklık", Description: "Kablosuz kulaklık", Price: 299.99, Image: "/static/images/headphones.jpg", Category: "Ses", Rating: 4.2, Stock: 50},
		{ID: 4, Name: "Tablet", Description: "10 inç tablet", Price: 2999.99, Image: "/static/images/tablet.jpg", Category: "Elektronik", Rating: 4.0, Stock: 30},
	}

	sampleCategories = []Category{
		{ID: 1, Name: "Elektronik"},
		{ID: 2, Name: "Bilgisayar"},
		{ID: 3, Name: "Ses"},
		{ID: 4, Name: "Giyim"},
		{ID: 5, Name: "Ev & Bahçe"},
		{ID: 6, Name: "Spor"},
	}
)

func main() {
	MainLogger.Println("KolajAI Marketplace başlatılıyor...")

	// Template fonksiyonlarını tanımla
	funcMap := template.FuncMap{
		"formatPrice": func(price float64) string {
			return fmt.Sprintf("%.2f TL", price)
		},
		"formatDate": func(t time.Time) string {
			return t.Format("02.01.2006 15:04")
		},
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
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"rand": func() int {
			return time.Now().Nanosecond() % 1000
		},
	}

	// Şablonları yükle
	tmpl, err := template.New("").Funcs(funcMap).ParseGlob("web/templates/**/*.gohtml")
	if err != nil {
		MainLogger.Fatalf("Şablonlar yüklenemedi: %v", err)
	}

	// Router oluştur
	router := http.NewServeMux()

	// Statik dosyalar
	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Ana sayfa
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		data := map[string]interface{}{
			"AppName":          "KolajAI Marketplace",
			"Year":             time.Now().Year(),
			"FeaturedProducts": sampleProducts,
			"Categories":       sampleCategories,
		}

		err := tmpl.ExecuteTemplate(w, "marketplace/index", data)
		if err != nil {
			MainLogger.Printf("Template hatası: %v", err)
			http.Error(w, "Sayfa yüklenemedi", http.StatusInternalServerError)
		}
	})

	// Ürünler sayfası
	router.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{
			"AppName":    "KolajAI Marketplace",
			"Year":       time.Now().Year(),
			"Products":   sampleProducts,
			"Categories": sampleCategories,
		}

		err := tmpl.ExecuteTemplate(w, "marketplace/products", data)
		if err != nil {
			MainLogger.Printf("Template hatası: %v", err)
			http.Error(w, "Sayfa yüklenemedi", http.StatusInternalServerError)
		}
	})

	// Sepet sayfası
	router.HandleFunc("/cart", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{
			"AppName":   "KolajAI Marketplace",
			"Year":      time.Now().Year(),
			"CartItems": []interface{}{},
			"CartTotal": 0.0,
		}

		err := tmpl.ExecuteTemplate(w, "marketplace/cart", data)
		if err != nil {
			MainLogger.Printf("Template hatası: %v", err)
			http.Error(w, "Sayfa yüklenemedi", http.StatusInternalServerError)
		}
	})

	// API - Ürün arama
	router.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		var results []Product

		for _, product := range sampleProducts {
			if query == "" || 
			   strings.Contains(strings.ToLower(product.Name), strings.ToLower(query)) ||
			   strings.Contains(strings.ToLower(product.Description), strings.ToLower(query)) {
				results = append(results, product)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `[`)
		for i, product := range results {
			if i > 0 {
				fmt.Fprintf(w, `,`)
			}
			fmt.Fprintf(w, `{"id":%d,"name":"%s","price":%.2f,"image":"%s"}`, 
				product.ID, product.Name, product.Price, product.Image)
		}
		fmt.Fprintf(w, `]`)
	})

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

	MainLogger.Printf("KolajAI Marketplace başlatıldı. %s adresinde dinleniyor...", addr)
	MainLogger.Fatal(server.ListenAndServe())
}