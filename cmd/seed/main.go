package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"kolajAi/internal/database"
	"kolajAi/internal/models"
	"kolajAi/internal/repository"
)

func main() {
	fmt.Println("KolajAI Data Seeder başlatılıyor...")

	// Veritabanı bağlantısı - Test için basit connection
	connectionString := "kolajai:password@tcp(localhost:3306)/kolajai?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := database.NewConnection(connectionString)
	if err != nil {
		log.Printf("MySQL bağlantısı kurulamadı: %v", err)
		log.Printf("Veritabanı bağlantısı olmadan seed işlemi yapılamaz")
		return
	}
	defer db.Close()

	// Repository ve servisler
	mysqlRepo := database.NewMySQLRepository(db)
	repo := repository.NewBaseRepository(mysqlRepo)

	// Kategoriler ekle
	fmt.Println("Kategoriler ekleniyor...")
	categories := []models.Category{
		{Name: "Elektronik", Description: "Elektronik ürünler", Image: "/static/images/categories/electronics.jpg", IsActive: true, SortOrder: 1},
		{Name: "Giyim", Description: "Giyim ve aksesuar", Image: "/static/images/categories/clothing.jpg", IsActive: true, SortOrder: 2},
		{Name: "Ev & Yaşam", Description: "Ev dekorasyonu ve yaşam ürünleri", Image: "/static/images/categories/home.jpg", IsActive: true, SortOrder: 3},
		{Name: "Spor", Description: "Spor malzemeleri", Image: "/static/images/categories/sports.jpg", IsActive: true, SortOrder: 4},
		{Name: "Kitap", Description: "Kitaplar ve eğitim materyalleri", Image: "/static/images/categories/books.jpg", IsActive: true, SortOrder: 5},
		{Name: "Sağlık", Description: "Sağlık ve kişisel bakım", Image: "/static/images/categories/health.jpg", IsActive: true, SortOrder: 6},
	}

	for i, category := range categories {
		id, err := repo.CreateStruct("categories", &category)
		if err != nil {
			log.Printf("Kategori eklenirken hata: %v", err)
		} else {
			categories[i].ID = uint(id)
			fmt.Printf("Kategori eklendi: %s\n", category.Name)
		}
	}

	// Kullanıcılar ekle
	fmt.Println("Kullanıcılar ekleniyor...")

	// Admin kullanıcı
	defaultAdminPassword := os.Getenv("ADMIN_PASSWORD")
	if defaultAdminPassword == "" {
		defaultAdminPassword = "admin123" // Fallback, production'da mutlaka env var kullanılmalı
		fmt.Println("WARNING: ADMIN_PASSWORD environment variable not set, using default password")
	}
	// Use stronger bcrypt cost for production security (12 instead of default 10)
	adminPassword, _ := bcrypt.GenerateFromPassword([]byte(defaultAdminPassword), 12)
	adminUser := models.User{
		Name:      "Admin User",
		Email:     "admin@kolajAi.com",
		Password:  string(adminPassword),
		Phone:     "0532 000 0000",
		IsActive:  true,
		IsAdmin:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	adminID, err := repo.CreateStruct("users", &adminUser)
	if err != nil {
		log.Printf("Admin kullanıcı eklenirken hata: %v", err)
	} else {
		adminUser.ID = adminID
		fmt.Println("Admin kullanıcı eklendi: admin@kolajAi.com")
	}

	// Satıcı kullanıcıları
	vendors := []struct {
		User   models.User
		Vendor models.Vendor
	}{
		{
			User: models.User{
				Name: "Ahmet Yılmaz", Email: "vendor1@kolajAi.com", Password: string(adminPassword), Phone: "0532 123 4567",
				IsActive: true, IsAdmin: false, CreatedAt: time.Now(), UpdatedAt: time.Now(),
			},
			Vendor: models.Vendor{
				CompanyName: "Yılmaz Elektronik", BusinessID: "1234567890", Phone: "0532 123 4567",
				Address: "İstanbul, Türkiye", City: "İstanbul", Country: "Türkiye", Status: "approved",
				CreatedAt: time.Now(), UpdatedAt: time.Now(),
			},
		},
		{
			User: models.User{
				Name: "Fatma Kaya", Email: "vendor2@kolajAi.com", Password: string(adminPassword), Phone: "0533 987 6543",
				IsActive: true, IsAdmin: false, CreatedAt: time.Now(), UpdatedAt: time.Now(),
			},
			Vendor: models.Vendor{
				CompanyName: "Kaya Giyim", BusinessID: "0987654321", Phone: "0533 987 6543",
				Address: "Ankara, Türkiye", City: "Ankara", Country: "Türkiye", Status: "approved",
				CreatedAt: time.Now(), UpdatedAt: time.Now(),
			},
		},
	}

	for _, v := range vendors {
		userID, err := repo.CreateStruct("users", &v.User)
		if err != nil {
			log.Printf("Satıcı kullanıcı eklenirken hata: %v", err)
			continue
		}

		v.Vendor.UserID = int(userID)
		vendorID, err := repo.CreateStruct("vendors", &v.Vendor)
		if err != nil {
			log.Printf("Satıcı bilgisi eklenirken hata: %v", err)
		} else {
			v.Vendor.ID = int(vendorID)
			fmt.Printf("Satıcı eklendi: %s (%s)\n", v.Vendor.CompanyName, v.User.Email)
		}
	}

	// Normal kullanıcılar
	normalUsers := []models.User{
		{Name: "Mehmet Demir", Email: "user1@kolajAi.com", Password: string(adminPassword), Phone: "0534 111 1111", IsActive: true, IsAdmin: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Name: "Ayşe Çelik", Email: "user2@kolajAi.com", Password: string(adminPassword), Phone: "0535 222 2222", IsActive: true, IsAdmin: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Name: "Ali Öz", Email: "user3@kolajAi.com", Password: string(adminPassword), Phone: "0536 333 3333", IsActive: true, IsAdmin: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, user := range normalUsers {
		userID, err := repo.CreateStruct("users", &user)
		if err != nil {
			log.Printf("Kullanıcı eklenirken hata: %v", err)
		} else {
			user.ID = userID
			fmt.Printf("Kullanıcı eklendi: %s (%s)\n", user.Name, user.Email)
		}
	}

	// Ürünler ekle
	fmt.Println("Ürünler ekleniyor...")
	products := []models.Product{
		// Elektronik ürünler
		{
			VendorID: 1, CategoryID: 1, Name: "iPhone 15 Pro", Description: "Apple iPhone 15 Pro 128GB - Titanyum renk. A17 Pro çip, ProRAW fotoğraf, Action Button.",
			ShortDesc: "Apple'ın en yeni Pro modeli", SKU: "IPH15PRO128-TITANIUM", Price: 45000.00, ComparePrice: 50000.00,
			Stock: 50, Status: "active", IsFeatured: true, AllowReviews: true, Rating: 4.8, ReviewCount: 234,
			Tags: "iphone,apple,telefon,akıllı telefon,pro,titanium", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		{
			VendorID: 1, CategoryID: 1, Name: "Samsung Galaxy S24 Ultra", Description: "Samsung Galaxy S24 Ultra 256GB - Phantom Black. 200MP kamera, S Pen dahil, AI özellikleri.",
			ShortDesc: "Samsung'un amiral gemisi", SKU: "SAMS24U256-BLACK", Price: 42000.00, ComparePrice: 47000.00,
			Stock: 30, Status: "active", IsFeatured: true, AllowReviews: true, Rating: 4.7, ReviewCount: 189,
			Tags: "samsung,galaxy,s24,ultra,telefon,android,s pen", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		{
			VendorID: 1, CategoryID: 1, Name: "MacBook Air M3", Description: "Apple MacBook Air 13.6 inç M3 çip 8GB RAM 256GB SSD - Gece Yarısı. Retina ekran, 18 saat pil.",
			ShortDesc: "Güçlü ve hafif laptop", SKU: "MBA-M3-256-MIDNIGHT", Price: 32999.00, ComparePrice: 36999.00,
			Stock: 15, Status: "active", IsFeatured: true, AllowReviews: true, Rating: 4.9, ReviewCount: 156,
			Tags: "macbook,air,m3,laptop,apple,ultrabook", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		{
			VendorID: 1, CategoryID: 1, Name: "iPad Pro 11 inç", Description: "iPad Pro 11 inç M4 çip 128GB WiFi - Space Black. Liquid Retina ekran, Apple Pencil desteği.",
			ShortDesc: "Profesyonel tablet deneyimi", SKU: "IPADPRO11-M4-128-BLACK", Price: 28999.00, ComparePrice: 32999.00,
			Stock: 25, Status: "active", IsFeatured: false, AllowReviews: true, Rating: 4.6, ReviewCount: 98,
			Tags: "ipad,pro,tablet,apple,m4,pencil", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		// Giyim ürünleri
		{
			VendorID: 2, CategoryID: 2, Name: "Levi's 501 Original Jeans", Description: "Klasik Levi's 501 Original kot pantolon. %100 pamuk, straight fit, vintage yıkama.",
			ShortDesc: "Zamansız klasik kot pantolon", SKU: "LEVIS-501-32-34-VINTAGE", Price: 599.99, ComparePrice: 699.99,
			Stock: 100, Status: "active", IsFeatured: false, AllowReviews: true, Rating: 4.5, ReviewCount: 267,
			Tags: "levis,501,kot,pantolon,erkek,klasik,vintage", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		{
			VendorID: 2, CategoryID: 2, Name: "Nike Dri-FIT Tişört", Description: "Nike Dri-FIT teknolojili spor tişört. Nefes alabilir kumaş, nem emici özellik.",
			ShortDesc: "Performans odaklı spor tişört", SKU: "NIKE-DRIFIT-L-BLACK", Price: 199.99, ComparePrice: 249.99,
			Stock: 78, Status: "active", IsFeatured: false, AllowReviews: true, Rating: 4.3, ReviewCount: 145,
			Tags: "nike,tişört,dri-fit,spor,erkek,performance", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		{
			VendorID: 2, CategoryID: 2, Name: "Zara Kadın Blazer", Description: "Zara kadın blazer ceket. Ofis ve günlük kullanım için ideal, modern kesim.",
			ShortDesc: "Şık ve modern blazer", SKU: "ZARA-BLAZER-M-NAVY", Price: 899.99, ComparePrice: 1199.99,
			Stock: 25, Status: "active", IsFeatured: true, AllowReviews: true, Rating: 4.4, ReviewCount: 89,
			Tags: "zara,blazer,ceket,kadın,ofis,şık,modern", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		// Ev & Yaşam ürünleri
		{
			VendorID: 1, CategoryID: 3, Name: "Dyson V15 Detect", Description: "Dyson V15 Detect kablosuz süpürge. Lazer toz algılama, 60 dakika çalışma süresi.",
			ShortDesc: "Gelişmiş kablosuz süpürge", SKU: "DYSON-V15-DETECT-GOLD", Price: 4299.99, ComparePrice: 4799.99,
			Stock: 8, Status: "active", IsFeatured: true, AllowReviews: true, Rating: 4.8, ReviewCount: 123,
			Tags: "dyson,süpürge,kablosuz,v15,detect,temizlik", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		{
			VendorID: 2, CategoryID: 3, Name: "Philips Hue Akıllı Ampul", Description: "Philips Hue White and Color akıllı LED ampul. 16 milyon renk, sesli kontrol.",
			ShortDesc: "Akıllı ev aydınlatması", SKU: "PHILIPS-HUE-E27-COLOR", Price: 399.99, ComparePrice: 499.99,
			Stock: 45, Status: "active", IsFeatured: false, AllowReviews: true, Rating: 4.6, ReviewCount: 78,
			Tags: "philips,hue,akıllı,ampul,led,renk,ev", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		{
			VendorID: 1, CategoryID: 3, Name: "Akıllı TV 55\"", Description: "4K UHD Smart TV",
			ShortDesc: "Büyük ekran deneyimi", SKU: "SMART-TV-55", Price: 12000.00, ComparePrice: 15000.00,
			Stock: 15, Status: "active", IsFeatured: true, AllowReviews: true,
			Tags: "tv,televizyon,smart,4k", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		{
			VendorID: 1, CategoryID: 4, Name: "Fitness Bisikleti", Description: "Ev tipi fitness bisikleti",
			ShortDesc: "Evde spor yapın", SKU: "FITNESS-BIKE-001", Price: 2500.00, ComparePrice: 3000.00,
			Stock: 10, Status: "active", IsFeatured: false, AllowReviews: true,
			Tags: "bisiklet,fitness,spor,egzersiz", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
	}

	for _, product := range products {
		productID, err := repo.CreateStruct("products", &product)
		if err != nil {
			log.Printf("Ürün eklenirken hata: %v", err)
		} else {
			product.ID = int(productID)
			fmt.Printf("Ürün eklendi: %s (₺%.2f)\n", product.Name, product.Price)
		}
	}

	fmt.Println("Data seeding tamamlandı!")
	fmt.Println("\nGiriş bilgileri:")
	fmt.Println("Admin: admin@kolajAi.com / admin123")
	fmt.Println("Satıcı 1: vendor1@kolajAi.com / admin123")
	fmt.Println("Satıcı 2: vendor2@kolajAi.com / admin123")
	fmt.Println("Kullanıcı 1: user1@kolajAi.com / admin123")
}
