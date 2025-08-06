package migrations

import (
	"database/sql"
	"fmt"
	"log"
)

// SQLiteMigrations contains all SQLite migration queries
var SQLiteMigrations = []string{
	// Users table
	`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		phone VARCHAR(20),
		is_active BOOLEAN DEFAULT 1,
		is_admin BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	// Categories table
	`CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(200) NOT NULL,
		slug VARCHAR(250) UNIQUE NOT NULL,
		description TEXT,
		parent_id INTEGER REFERENCES categories(id) ON DELETE CASCADE,
		image VARCHAR(500),
		is_active BOOLEAN DEFAULT 1,
		is_visible BOOLEAN DEFAULT 1,
		is_featured BOOLEAN DEFAULT 0,
		sort_order INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	// Vendors table
	`CREATE TABLE IF NOT EXISTS vendors (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		company_name VARCHAR(255) NOT NULL,
		business_id VARCHAR(50) UNIQUE,
		phone VARCHAR(20),
		address TEXT,
		city VARCHAR(100),
		country VARCHAR(100),
		status VARCHAR(20) DEFAULT 'pending',
		commission_rate DECIMAL(5,2) DEFAULT 10.00,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	// Products table
	`CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		vendor_id INTEGER NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
		category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		short_desc VARCHAR(500),
		sku VARCHAR(100) UNIQUE NOT NULL,
		price DECIMAL(10,2) NOT NULL DEFAULT 0.00,
		compare_price DECIMAL(10,2) DEFAULT 0.00,
		cost_price DECIMAL(10,2) DEFAULT 0.00,
		wholesale_price DECIMAL(10,2) DEFAULT 0.00,
		min_wholesale_qty INTEGER DEFAULT 1,
		stock INTEGER NOT NULL DEFAULT 0,
		min_stock INTEGER DEFAULT 0,
		weight DECIMAL(8,2) DEFAULT 0.00,
		dimensions VARCHAR(100),
		status VARCHAR(20) DEFAULT 'draft',
		is_digital BOOLEAN DEFAULT 0,
		is_featured BOOLEAN DEFAULT 0,
		allow_reviews BOOLEAN DEFAULT 1,
		meta_title VARCHAR(255),
		meta_desc VARCHAR(500),
		tags VARCHAR(1000),
		view_count INTEGER DEFAULT 0,
		sales_count INTEGER DEFAULT 0,
		rating DECIMAL(3,2) DEFAULT 0.00,
		review_count INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	// Product images table
	`CREATE TABLE IF NOT EXISTS product_images (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
		image_url VARCHAR(500) NOT NULL,
		alt_text VARCHAR(255),
		sort_order INTEGER DEFAULT 0,
		is_primary BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	// Auctions table
	`CREATE TABLE IF NOT EXISTS auctions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
		vendor_id INTEGER NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
		title VARCHAR(255) NOT NULL,
		description TEXT,
		starting_price DECIMAL(10,2) NOT NULL,
		reserve_price DECIMAL(10,2) DEFAULT 0.00,
		current_bid DECIMAL(10,2) DEFAULT 0.00,
		bid_increment DECIMAL(10,2) DEFAULT 1.00,
		total_bids INTEGER DEFAULT 0,
		start_time DATETIME NOT NULL,
		end_time DATETIME NOT NULL,
		status VARCHAR(20) DEFAULT 'draft',
		winner_id INTEGER REFERENCES users(id),
		view_count INTEGER DEFAULT 0,
		is_reserve_met BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	// Orders table
	`CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		order_number VARCHAR(50) UNIQUE NOT NULL,
		status VARCHAR(20) DEFAULT 'pending',
		subtotal DECIMAL(10,2) NOT NULL DEFAULT 0.00,
		tax_amount DECIMAL(10,2) DEFAULT 0.00,
		shipping_amount DECIMAL(10,2) DEFAULT 0.00,
		discount_amount DECIMAL(10,2) DEFAULT 0.00,
		total_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
		currency VARCHAR(3) DEFAULT 'TRY',
		payment_status VARCHAR(20) DEFAULT 'pending',
		payment_method VARCHAR(50),
		shipping_address TEXT,
		billing_address TEXT,
		notes TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	// Order items table
	`CREATE TABLE IF NOT EXISTS order_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
		product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
		vendor_id INTEGER NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
		product_name VARCHAR(255) NOT NULL,
		product_sku VARCHAR(100) NOT NULL,
		quantity INTEGER NOT NULL DEFAULT 1,
		unit_price DECIMAL(10,2) NOT NULL,
		total_price DECIMAL(10,2) NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	// Sessions table
	`CREATE TABLE IF NOT EXISTS sessions (
		id VARCHAR(255) PRIMARY KEY,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		data TEXT,
		expires_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	// Create indexes for better performance
	`CREATE INDEX IF NOT EXISTS idx_categories_parent ON categories(parent_id)`,
	`CREATE INDEX IF NOT EXISTS idx_categories_active ON categories(is_active)`,
	`CREATE INDEX IF NOT EXISTS idx_categories_slug ON categories(slug)`,
	`CREATE INDEX IF NOT EXISTS idx_products_vendor ON products(vendor_id)`,
	`CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id)`,
	`CREATE INDEX IF NOT EXISTS idx_products_status ON products(status)`,
	`CREATE INDEX IF NOT EXISTS idx_products_featured ON products(is_featured)`,
	`CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku)`,
	`CREATE INDEX IF NOT EXISTS idx_product_images_product ON product_images(product_id)`,
	`CREATE INDEX IF NOT EXISTS idx_auctions_vendor ON auctions(vendor_id)`,
	`CREATE INDEX IF NOT EXISTS idx_auctions_status ON auctions(status)`,
	`CREATE INDEX IF NOT EXISTS idx_auctions_end_time ON auctions(end_time)`,
	`CREATE INDEX IF NOT EXISTS idx_orders_user ON orders(user_id)`,
	`CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)`,
	`CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id)`,
	`CREATE INDEX IF NOT EXISTS idx_order_items_product ON order_items(product_id)`,
	`CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id)`,
	`CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at)`,
}

// RunSQLiteMigrations runs all SQLite migrations
func RunSQLiteMigrations(db *sql.DB) error {
	log.Println("🔄 Running SQLite migrations...")

	for i, migration := range SQLiteMigrations {
		log.Printf("Running migration %d/%d", i+1, len(SQLiteMigrations))

		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to run migration %d: %w", i+1, err)
		}
	}

	log.Println("✅ All SQLite migrations completed successfully")
	return nil
}
