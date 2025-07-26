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
			VendorID:        1,
			CategoryID:      1,
			Name:            "iPhone 15 Pro",
			Description:     "Apple iPhone 15 Pro - 128GB",
			ShortDesc:       "En yeni iPhone modeli",
			SKU:             "IPHONE15PRO128",
			Price:           45000.00,
			ComparePrice:    50000.00,
			CostPrice:       40000.00,
			WholesalePrice:  42000.00,
			MinWholesaleQty: 5,
			Stock:           10,
			MinStock:        2,
			Weight:          0.2,
			Dimensions:      "15.0x7.0x0.8 cm",
			Status:          "active",
			IsDigital:       false,
			IsFeatured:      true,
			AllowReviews:    true,
			MetaTitle:       "iPhone 15 Pro - KolajAI",
			MetaDesc:        "Apple iPhone 15 Pro satın al",
			Tags:            "iphone,apple,telefon",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			VendorID:        1,
			CategoryID:      1,
			Name:            "Samsung Galaxy S24",
			Description:     "Samsung Galaxy S24 - 256GB",
			ShortDesc:       "Premium Android telefon",
			SKU:             "GALAXYS24256",
			Price:           35000.00,
			ComparePrice:    40000.00,
			CostPrice:       30000.00,
			WholesalePrice:  32000.00,
			MinWholesaleQty: 5,
			Stock:           15,
			MinStock:        3,
			Weight:          0.18,
			Dimensions:      "14.5x7.2x0.9 cm",
			Status:          "active",
			IsDigital:       false,
			IsFeatured:      true,
			AllowReviews:    true,
			MetaTitle:       "Samsung Galaxy S24 - KolajAI",
			MetaDesc:        "Samsung Galaxy S24 satın al",
			Tags:            "samsung,galaxy,android",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			VendorID:        1,
			CategoryID:      2,
			Name:            "Nike Air Max",
			Description:     "Nike Air Max ayakkabı - Siyah",
			ShortDesc:       "Rahat spor ayakkabı",
			SKU:             "NIKEAIRMAX01",
			Price:           2500.00,
			ComparePrice:    3000.00,
			CostPrice:       2000.00,
			WholesalePrice:  2200.00,
			MinWholesaleQty: 10,
			Stock:           25,
			MinStock:        5,
			Weight:          0.8,
			Dimensions:      "30x12x10 cm",
			Status:          "active",
			IsDigital:       false,
			IsFeatured:      true,
			AllowReviews:    true,
			MetaTitle:       "Nike Air Max - KolajAI",
			MetaDesc:        "Nike Air Max ayakkabı satın al",
			Tags:            "nike,ayakkabı,spor",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	// Insert products
	for _, product := range products {
		_, err := db.Exec(`
			INSERT OR REPLACE INTO products (
				vendor_id, category_id, name, description, short_desc, sku, 
				price, compare_price, cost_price, wholesale_price, min_wholesale_qty,
				stock, min_stock, weight, dimensions, status, is_digital, is_featured, 
				allow_reviews, meta_title, meta_desc, tags, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, product.VendorID, product.CategoryID, product.Name, product.Description, product.ShortDesc, product.SKU,
		   product.Price, product.ComparePrice, product.CostPrice, product.WholesalePrice, product.MinWholesaleQty,
		   product.Stock, product.MinStock, product.Weight, product.Dimensions, product.Status, product.IsDigital, product.IsFeatured,
		   product.AllowReviews, product.MetaTitle, product.MetaDesc, product.Tags, product.CreatedAt, product.UpdatedAt)
		if err != nil {
			log.Printf("Error inserting product %s: %v", product.Name, err)
		} else {
			log.Printf("Inserted product: %s", product.Name)
		}
	}

	log.Println("Seed data inserted successfully!")
}