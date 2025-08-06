package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Seeder handles database seeding
type Seeder struct {
	db     *sql.DB
	dbType DatabaseType
}

// NewSeeder creates a new seeder
func NewSeeder(db *sql.DB, dbType DatabaseType) *Seeder {
	return &Seeder{
		db:     db,
		dbType: dbType,
	}
}

// SeedDatabase seeds the database with initial data
func (s *Seeder) SeedDatabase() error {
	log.Println("🌱 Starting database seeding...")

	// Seed in order due to foreign key constraints
	if err := s.seedUsers(); err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	if err := s.seedCategories(); err != nil {
		return fmt.Errorf("failed to seed categories: %w", err)
	}

	if err := s.seedVendors(); err != nil {
		return fmt.Errorf("failed to seed vendors: %w", err)
	}

	if err := s.seedProducts(); err != nil {
		return fmt.Errorf("failed to seed products: %w", err)
	}

	if err := s.seedAuctions(); err != nil {
		return fmt.Errorf("failed to seed auctions: %w", err)
	}

	log.Println("✅ Database seeding completed successfully")
	return nil
}

// seedUsers seeds initial users
func (s *Seeder) seedUsers() error {
	log.Println("👥 Seeding users...")

	// Check if users already exist
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("Users already exist, skipping...")
		return nil
	}

	// Get admin password from environment
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin123"
		log.Println("⚠️ Using default admin password. Set ADMIN_PASSWORD environment variable for security.")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	users := []struct {
		name     string
		email    string
		phone    string
		isAdmin  bool
		isActive bool
	}{
		{"Admin User", "admin@kolajai.com", "+90 532 000 0000", true, true},
		{"Ahmet Yılmaz", "vendor1@kolajai.com", "+90 532 123 4567", false, true},
		{"Fatma Kaya", "vendor2@kolajai.com", "+90 533 987 6543", false, true},
		{"Mehmet Demir", "user1@kolajai.com", "+90 534 111 1111", false, true},
		{"Ayşe Çelik", "user2@kolajai.com", "+90 535 222 2222", false, true},
		{"Ali Öz", "user3@kolajai.com", "+90 536 333 3333", false, true},
	}

	for _, user := range users {
		_, err := s.db.Exec(`
			INSERT INTO users (name, email, password, phone, is_admin, is_active, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			user.name, user.email, string(hashedPassword), user.phone, user.isAdmin, user.isActive, time.Now(), time.Now())
		
		if err != nil {
			return fmt.Errorf("failed to insert user %s: %w", user.email, err)
		}
		
		log.Printf("✅ Created user: %s (%s)", user.name, user.email)
	}

	return nil
}

// seedCategories seeds initial categories
func (s *Seeder) seedCategories() error {
	log.Println("📂 Seeding categories...")

	// Check if categories already exist
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("Categories already exist, skipping...")
		return nil
	}

	categories := []struct {
		name        string
		slug        string
		description string
		image       string
		sortOrder   int
	}{
		{"Elektronik", "elektronik", "Telefon, bilgisayar, tablet ve diğer elektronik ürünler", "/web/static/images/categories/electronics.jpg", 1},
		{"Giyim & Moda", "giyim-moda", "Kadın, erkek ve çocuk giyim ürünleri", "/web/static/images/categories/clothing.jpg", 2},
		{"Ev & Yaşam", "ev-yasam", "Ev dekorasyonu, mutfak eşyaları ve yaşam ürünleri", "/web/static/images/categories/home.jpg", 3},
		{"Spor & Outdoor", "spor-outdoor", "Spor malzemeleri ve outdoor aktivite ürünleri", "/web/static/images/categories/sports.jpg", 4},
		{"Kitap & Medya", "kitap-medya", "Kitaplar, dergiler ve dijital medya", "/web/static/images/categories/books.jpg", 5},
		{"Sağlık & Kişisel Bakım", "saglik-kisisel-bakim", "Sağlık ürünleri ve kişisel bakım malzemeleri", "/web/static/images/categories/health.jpg", 6},
		{"Otomobil & Motosiklet", "otomobil-motosiklet", "Araç aksesuarları ve yedek parçalar", "/web/static/images/categories/automotive.jpg", 7},
		{"Bahçe & Yapı Market", "bahce-yapi-market", "Bahçıvanlık ve yapı malzemeleri", "/web/static/images/categories/garden.jpg", 8},
	}

	for _, cat := range categories {
		_, err := s.db.Exec(`
			INSERT INTO categories (name, slug, description, image, is_active, is_visible, sort_order, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			cat.name, cat.slug, cat.description, cat.image, true, true, cat.sortOrder, time.Now(), time.Now())
		
		if err != nil {
			return fmt.Errorf("failed to insert category %s: %w", cat.name, err)
		}
		
		log.Printf("✅ Created category: %s", cat.name)
	}

	return nil
}

// seedVendors seeds initial vendors
func (s *Seeder) seedVendors() error {
	log.Println("🏪 Seeding vendors...")

	// Check if vendors already exist
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM vendors").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("Vendors already exist, skipping...")
		return nil
	}

	vendors := []struct {
		userID      int
		companyName string
		businessID  string
		phone       string
		address     string
		city        string
		country     string
	}{
		{2, "Yılmaz Elektronik", "1234567890", "+90 532 123 4567", "Fatih Mah. Elektronik Cad. No:123", "İstanbul", "Türkiye"},
		{3, "Kaya Tekstil", "0987654321", "+90 533 987 6543", "Merkez Mah. Tekstil Sok. No:45", "İzmir", "Türkiye"},
	}

	for _, vendor := range vendors {
		_, err := s.db.Exec(`
			INSERT INTO vendors (user_id, company_name, business_id, phone, address, city, country, status, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			vendor.userID, vendor.companyName, vendor.businessID, vendor.phone, vendor.address, vendor.city, vendor.country, "approved", time.Now(), time.Now())
		
		if err != nil {
			return fmt.Errorf("failed to insert vendor %s: %w", vendor.companyName, err)
		}
		
		log.Printf("✅ Created vendor: %s", vendor.companyName)
	}

	return nil
}

// seedProducts seeds initial products
func (s *Seeder) seedProducts() error {
	log.Println("📦 Seeding products...")

	// Check if products already exist
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("Products already exist, skipping...")
		return nil
	}

	products := []struct {
		vendorID     int
		categoryID   int
		name         string
		description  string
		shortDesc    string
		sku          string
		price        float64
		comparePrice float64
		stock        int
		isFeatured   bool
		rating       float64
		reviewCount  int
		tags         string
	}{
		{1, 1, "iPhone 15 Pro 128GB", "Apple iPhone 15 Pro 128GB Titanyum renk. A17 Pro çip ile üstün performans, ProRAW fotoğraf çekimi, Action Button ve USB-C.", "Apple'ın en yeni Pro modeli", "IPHONE15PRO-128-TI", 52999.00, 56999.00, 25, true, 4.8, 234, "iphone,apple,telefon,akıllı telefon,pro,titanyum"},
		{1, 1, "Samsung Galaxy S24 Ultra", "Samsung Galaxy S24 Ultra 256GB Phantom Black. 200MP kamera, S Pen dahil, AI destekli özellikler ve 6.8 inç Dynamic AMOLED ekran.", "Samsung'un amiral gemisi", "GALAXY-S24U-256-PB", 48999.00, 52999.00, 18, true, 4.7, 189, "samsung,galaxy,s24,ultra,telefon,android,s pen"},
		{1, 1, "MacBook Air M3 13 inç", "Apple MacBook Air 13.6 inç M3 çip 8GB RAM 256GB SSD Gece Yarısı. Liquid Retina ekran, Touch ID, 18 saate varan pil ömrü.", "Güçlü ve hafif laptop", "MBA-M3-256-MN", 36999.00, 39999.00, 12, true, 4.9, 156, "macbook,air,m3,laptop,apple,ultrabook"},
		{1, 1, "iPad Pro 11 inç M4", "iPad Pro 11 inç M4 çip 128GB WiFi Space Black. Ultra Retina XDR ekran, Apple Pencil Pro desteği, profesyonel performans.", "Profesyonel tablet deneyimi", "IPADPRO11-M4-128-SB", 31999.00, 34999.00, 15, false, 4.6, 98, "ipad,pro,tablet,apple,m4,pencil"},
		{2, 2, "Levi's 501 Original Jeans", "Klasik Levi's 501 Original kot pantolon. %100 pamuk, straight fit, vintage yıkama. Zamansız stil ve dayanıklılık.", "Zamansız klasik kot pantolon", "LEVIS-501-32-34-VW", 649.99, 749.99, 67, false, 4.5, 267, "levis,501,kot,pantolon,erkek,klasik,vintage"},
		{2, 2, "Nike Dri-FIT Tişört", "Nike Dri-FIT teknolojili spor tişört. Nefes alabilir kumaş, nem emici özellik, günlük spor ve antrenman için ideal.", "Performans odaklı spor tişört", "NIKE-DRIFIT-L-BLK", 229.99, 279.99, 89, false, 4.3, 145, "nike,tişört,dri-fit,spor,erkek,performance"},
		{2, 2, "Zara Kadın Blazer Ceket", "Zara kadın blazer ceket. Ofis ve günlük kullanım için ideal, modern kesim, yüksek kalite kumaş.", "Şık ve modern blazer", "ZARA-BLAZER-M-NV", 999.99, 1299.99, 23, true, 4.4, 89, "zara,blazer,ceket,kadın,ofis,şık,modern"},
		{1, 3, "Dyson V15 Detect", "Dyson V15 Detect kablosuz süpürge. Lazer toz algılama teknolojisi, 60 dakika çalışma süresi, HEPA filtre.", "Gelişmiş kablosuz süpürge", "DYSON-V15-DETECT-GD", 4799.99, 5299.99, 6, true, 4.8, 123, "dyson,süpürge,kablosuz,v15,detect,temizlik"},
		{2, 3, "Philips Hue Akıllı Ampul", "Philips Hue White and Color akıllı LED ampul. 16 milyon renk, sesli kontrol, uygulama kontrolü, enerji tasarrufu.", "Akıllı ev aydınlatması", "PHILIPS-HUE-E27-CLR", 449.99, 549.99, 34, false, 4.6, 78, "philips,hue,akıllı,ampul,led,renk,ev"},
		{1, 4, "Nike Air Max 270", "Nike Air Max 270 spor ayakkabı. Max Air yastıklama teknolojisi, mesh üst yapı, günlük kullanım için rahat ve şık.", "Rahat ve şık spor ayakkabı", "NIKE-AIRMAX270-42-WHT", 999.99, 1199.99, 45, false, 4.4, 189, "nike,air max,270,ayakkabı,spor,erkek"},
		{2, 5, "Sapiens: İnsanlığın Kısa Tarihi", "Yuval Noah Harari'nin çığır açan eseri. İnsanlığın 70.000 yıllık serüvenini anlatan, dünya çapında bestseller kitap.", "Dünya çapında bestseller", "BOOK-SAPIENS-TR", 99.99, 129.99, 156, false, 4.6, 267, "sapiens,harari,kitap,tarih,insanlık"},
		{1, 6, "Oral-B Genius X", "Oral-B Genius X yapay zeka destekli elektrikli diş fırçası. 6 temizleme modu, basınç sensörü, akıllı rehberlik.", "AI destekli diş bakımı", "ORALB-GENIUSX-WHT", 2199.99, 2599.99, 19, true, 4.5, 89, "oral-b,diş fırçası,elektrikli,genius,ai"},
	}

	for _, prod := range products {
		_, err := s.db.Exec(`
			INSERT INTO products (vendor_id, category_id, name, description, short_desc, sku, price, compare_price, stock, status, is_featured, rating, review_count, tags, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			prod.vendorID, prod.categoryID, prod.name, prod.description, prod.shortDesc, prod.sku, prod.price, prod.comparePrice, prod.stock, "active", prod.isFeatured, prod.rating, prod.reviewCount, prod.tags, time.Now(), time.Now())
		
		if err != nil {
			return fmt.Errorf("failed to insert product %s: %w", prod.name, err)
		}
		
		log.Printf("✅ Created product: %s", prod.name)
	}

	return nil
}

// seedAuctions seeds initial auctions
func (s *Seeder) seedAuctions() error {
	log.Println("🔨 Seeding auctions...")

	// Check if auctions already exist
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM auctions").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("Auctions already exist, skipping...")
		return nil
	}

	// Create auctions that end in the future
	endTime := time.Now().Add(7 * 24 * time.Hour) // 7 days from now
	startTime := time.Now().Add(-1 * time.Hour)   // Started 1 hour ago

	auctions := []struct {
		vendorID      int
		title         string
		description   string
		startingPrice float64
		currentBid    float64
		totalBids     int
	}{
		{1, "Vintage Apple iPhone 4S - Koleksiyonluk", "2011 yılından kalma orijinal Apple iPhone 4S. Kutulu, aksesuarlı, çalışır durumda. Teknoloji koleksiyoncuları için ideal.", 2500.00, 3200.00, 15},
		{2, "Limited Edition Nike Air Jordan 1", "Sınırlı sayıda üretilen Nike Air Jordan 1 ayakkabı. Orijinal kutusu ve sertifikası mevcut. Sneaker koleksiyoncuları için.", 8000.00, 9500.00, 23},
		{1, "Antika Rolex Submariner Saati", "1980'lerden kalma orijinal Rolex Submariner saati. Servisi yapılmış, çalışır durumda. Saat koleksiyoncuları için.", 45000.00, 52000.00, 8},
		{2, "Nadir Bulunan Türk Halısı", "El dokuması, 100 yıllık antika Türk halısı. Müze kalitesinde, özel koleksiyondan çıkmış. Sanat eseri niteliğinde.", 15000.00, 18500.00, 12},
	}

	for _, auction := range auctions {
		_, err := s.db.Exec(`
			INSERT INTO auctions (vendor_id, title, description, starting_price, current_bid, total_bids, start_time, end_time, status, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			auction.vendorID, auction.title, auction.description, auction.startingPrice, auction.currentBid, auction.totalBids, startTime, endTime, "active", time.Now(), time.Now())
		
		if err != nil {
			return fmt.Errorf("failed to insert auction %s: %w", auction.title, err)
		}
		
		log.Printf("✅ Created auction: %s", auction.title)
	}

	return nil
}

// SeedGlobalDatabase seeds the global database
func SeedGlobalDatabase() error {
	if GlobalDBManager == nil {
		return fmt.Errorf("global database not initialized")
	}

	seeder := NewSeeder(GlobalDBManager.GetDB(), GlobalDBManager.GetType())
	return seeder.SeedDatabase()
}