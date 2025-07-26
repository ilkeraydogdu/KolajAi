package main

import (
	"log"
	"time"

	"kolajAi/internal/database"
	"kolajAi/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Connect to database
	db, err := database.NewSQLiteConnection("kolajAi.db")
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	// Create a sample user first
	_, err = db.Exec(`
		INSERT OR IGNORE INTO users (id, name, email, password, phone, is_active, is_admin, created_at, updated_at)
		VALUES (1, 'Test Vendor', 'vendor@test.com', 'hashed_password', '0532 123 4567', 1, 0, ?, ?)
	`, time.Now(), time.Now())
	if err != nil {
		log.Printf("Error inserting user: %v", err)
	} else {
		log.Println("Inserted user")
	}

	// Create a sample vendor
	_, err = db.Exec(`
		INSERT OR IGNORE INTO vendors (id, user_id, company_name, business_id, phone, address, city, country, status, created_at, updated_at)
		VALUES (1, 1, 'Test Company', '1234567890', '0532 123 4567', 'Test Address', 'Istanbul', 'Turkey', 'approved', ?, ?)
	`, time.Now(), time.Now())
	if err != nil {
		log.Printf("Error inserting vendor: %v", err)
	} else {
		log.Println("Inserted vendor")
	}

	// Create sample categories
	categories := []models.Category{
		{
			Name:        "Elektronik",
			Description: "Elektronik ürünler",
			IsActive:    true,
			SortOrder:   1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "Giyim",
			Description: "Giyim ürünleri",
			IsActive:    true,
			SortOrder:   2,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "Ev & Bahçe",
			Description: "Ev ve bahçe ürünleri",
			IsActive:    true,
			SortOrder:   3,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// Insert categories
	for _, category := range categories {
		_, err := db.Exec(`
			INSERT OR IGNORE INTO categories (name, description, is_active, sort_order, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, category.Name, category.Description, category.IsActive, category.SortOrder, category.CreatedAt, category.UpdatedAt)
		if err != nil {
			log.Printf("Error inserting category %s: %v", category.Name, err)
		} else {
			log.Printf("Inserted category: %s", category.Name)
		}
	}

	// Create sample products
	products := []models.Product{
		{
			VendorID:    1,
			CategoryID:  1,
			Name:        "iPhone 15 Pro",
			Description: "Apple iPhone 15 Pro - 128GB",
			ShortDesc:   "En yeni iPhone modeli",
			SKU:         "IPHONE15PRO128",
			Price:       45000.00,
			ComparePrice: 50000.00,
			Stock:       10,
			Status:      "active",
			IsFeatured:  true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			VendorID:    1,
			CategoryID:  1,
			Name:        "Samsung Galaxy S24",
			Description: "Samsung Galaxy S24 - 256GB",
			ShortDesc:   "Premium Android telefon",
			SKU:         "GALAXYS24256",
			Price:       35000.00,
			ComparePrice: 40000.00,
			Stock:       15,
			Status:      "active",
			IsFeatured:  true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			VendorID:    1,
			CategoryID:  2,
			Name:        "Nike Air Max",
			Description: "Nike Air Max ayakkabı - Siyah",
			ShortDesc:   "Rahat spor ayakkabı",
			SKU:         "NIKEAIRMAX01",
			Price:       2500.00,
			ComparePrice: 3000.00,
			Stock:       25,
			Status:      "active",
			IsFeatured:  true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// Insert products
	for _, product := range products {
		_, err := db.Exec(`
			INSERT OR IGNORE INTO products (vendor_id, category_id, name, description, short_desc, sku, price, compare_price, stock, status, is_featured, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, product.VendorID, product.CategoryID, product.Name, product.Description, product.ShortDesc, product.SKU, product.Price, product.ComparePrice, product.Stock, product.Status, product.IsFeatured, product.CreatedAt, product.UpdatedAt)
		if err != nil {
			log.Printf("Error inserting product %s: %v", product.Name, err)
		} else {
			log.Printf("Inserted product: %s", product.Name)
		}
	}

	log.Println("Seed data inserted successfully!")
}