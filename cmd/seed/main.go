package main

import (
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"kolajAi/internal/database"
	"kolajAi/internal/models"
	"kolajAi/internal/repository"
)

func main() {
	fmt.Println("KolajAI Data Seeder başlatılıyor...")

	// Veritabanı bağlantısı
	db, err := database.NewSQLiteConnection("kolajAi.db")
	if err != nil {
		log.Fatalf("Veritabanı bağlantısı kurulamadı: %v", err)
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
			categories[i].ID = int(id)
			fmt.Printf("Kategori eklendi: %s\n", category.Name)
		}
	}

	// Kullanıcılar ekle
	fmt.Println("Kullanıcılar ekleniyor...")

	// Admin kullanıcı
	adminPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
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
		{
			VendorID: 1, CategoryID: 1, Name: "iPhone 15 Pro", Description: "Apple iPhone 15 Pro 128GB",
			ShortDesc: "En yeni iPhone modeli", SKU: "IPH15PRO128", Price: 45000.00, ComparePrice: 50000.00,
			Stock: 50, Status: "active", IsFeatured: true, AllowReviews: true,
			Tags: "iphone,apple,telefon,akıllı telefon", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		{
			VendorID: 1, CategoryID: 1, Name: "Samsung Galaxy S24", Description: "Samsung Galaxy S24 256GB",
			ShortDesc: "Samsung'un flagShip modeli", SKU: "SAMS24256", Price: 35000.00, ComparePrice: 40000.00,
			Stock: 30, Status: "active", IsFeatured: true, AllowReviews: true,
			Tags: "samsung,galaxy,telefon,android", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		{
			VendorID: 2, CategoryID: 2, Name: "Erkek Kot Pantolon", Description: "Slim fit erkek kot pantolon",
			ShortDesc: "Rahat ve şık kot pantolon", SKU: "ERKEK-KOT-001", Price: 299.99, ComparePrice: 399.99,
			Stock: 100, Status: "active", IsFeatured: false, AllowReviews: true,
			Tags: "kot,pantolon,erkek,giyim", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		{
			VendorID: 2, CategoryID: 2, Name: "Kadın Elbise", Description: "Şık kadın elbisesi",
			ShortDesc: "Özel günler için ideal", SKU: "KADIN-ELBISE-001", Price: 599.99, ComparePrice: 799.99,
			Stock: 25, Status: "active", IsFeatured: true, AllowReviews: true,
			Tags: "elbise,kadın,giyim,şık", CreatedAt: time.Now(), UpdatedAt: time.Now(),
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
