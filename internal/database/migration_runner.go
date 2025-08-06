package database

import (
	"database/sql"
	"fmt"
	"log"

	"kolajAi/internal/database/migrations"
)

// MigrationRunner handles database migrations
type MigrationRunner struct {
	db     *sql.DB
	dbType DatabaseType
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(db *sql.DB, dbType DatabaseType) *MigrationRunner {
	return &MigrationRunner{
		db:     db,
		dbType: dbType,
	}
}

// RunMigrations runs migrations based on database type
func (mr *MigrationRunner) RunMigrations() error {
	log.Printf("🔄 Starting migrations for %s database", mr.dbType)

	switch mr.dbType {
	case SQLite:
		return migrations.RunSQLiteMigrations(mr.db)
	case MySQL:
		return mr.runMySQLMigrations()
	default:
		return fmt.Errorf("unsupported database type: %s", mr.dbType)
	}
}

// runMySQLMigrations runs MySQL specific migrations
func (mr *MigrationRunner) runMySQLMigrations() error {
	log.Println("🔄 Running MySQL migrations...")

	// MySQL migration files are in the migrations directory
	// For now, we'll use a simple approach
	mysqlMigrations := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			phone VARCHAR(20),
			is_active BOOLEAN DEFAULT TRUE,
			is_admin BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_users_email (email),
			INDEX idx_users_active (is_active)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Categories table
		`CREATE TABLE IF NOT EXISTS categories (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			slug VARCHAR(250) UNIQUE NOT NULL,
			description TEXT,
			parent_id INT NULL,
			image VARCHAR(500),
			is_active BOOLEAN DEFAULT TRUE,
			is_visible BOOLEAN DEFAULT TRUE,
			is_featured BOOLEAN DEFAULT FALSE,
			sort_order INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (parent_id) REFERENCES categories(id) ON DELETE CASCADE,
			INDEX idx_categories_parent (parent_id),
			INDEX idx_categories_active (is_active),
			INDEX idx_categories_slug (slug),
			INDEX idx_categories_sort (sort_order)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Vendors table
		`CREATE TABLE IF NOT EXISTS vendors (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			company_name VARCHAR(255) NOT NULL,
			business_id VARCHAR(50) UNIQUE,
			phone VARCHAR(20),
			address TEXT,
			city VARCHAR(100),
			country VARCHAR(100),
			status ENUM('pending', 'approved', 'rejected', 'suspended') DEFAULT 'pending',
			commission_rate DECIMAL(5,2) DEFAULT 10.00,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_vendors_user (user_id),
			INDEX idx_vendors_status (status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Products table
		`CREATE TABLE IF NOT EXISTS products (
			id INT AUTO_INCREMENT PRIMARY KEY,
			vendor_id INT NOT NULL,
			category_id INT NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			short_desc VARCHAR(500),
			sku VARCHAR(100) UNIQUE NOT NULL,
			price DECIMAL(10,2) NOT NULL DEFAULT 0.00,
			compare_price DECIMAL(10,2) DEFAULT 0.00,
			cost_price DECIMAL(10,2) DEFAULT 0.00,
			wholesale_price DECIMAL(10,2) DEFAULT 0.00,
			min_wholesale_qty INT DEFAULT 1,
			stock INT NOT NULL DEFAULT 0,
			min_stock INT DEFAULT 0,
			weight DECIMAL(8,2) DEFAULT 0.00,
			dimensions VARCHAR(100),
			status ENUM('draft', 'active', 'inactive', 'out_of_stock') DEFAULT 'draft',
			is_digital BOOLEAN DEFAULT FALSE,
			is_featured BOOLEAN DEFAULT FALSE,
			allow_reviews BOOLEAN DEFAULT TRUE,
			meta_title VARCHAR(255),
			meta_desc VARCHAR(500),
			tags VARCHAR(1000),
			view_count INT DEFAULT 0,
			sales_count INT DEFAULT 0,
			rating DECIMAL(3,2) DEFAULT 0.00,
			review_count INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (vendor_id) REFERENCES vendors(id) ON DELETE CASCADE,
			FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT,
			INDEX idx_products_vendor (vendor_id),
			INDEX idx_products_category (category_id),
			INDEX idx_products_status (status),
			INDEX idx_products_featured (is_featured),
			INDEX idx_products_sku (sku),
			INDEX idx_products_rating (rating),
			FULLTEXT idx_products_search (name, description, tags)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Product images table
		`CREATE TABLE IF NOT EXISTS product_images (
			id INT AUTO_INCREMENT PRIMARY KEY,
			product_id INT NOT NULL,
			image_url VARCHAR(500) NOT NULL,
			alt_text VARCHAR(255),
			sort_order INT DEFAULT 0,
			is_primary BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
			INDEX idx_product_images_product (product_id),
			INDEX idx_product_images_primary (is_primary)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Auctions table
		`CREATE TABLE IF NOT EXISTS auctions (
			id INT AUTO_INCREMENT PRIMARY KEY,
			product_id INT NULL,
			vendor_id INT NOT NULL,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			starting_price DECIMAL(10,2) NOT NULL,
			reserve_price DECIMAL(10,2) DEFAULT 0.00,
			current_bid DECIMAL(10,2) DEFAULT 0.00,
			bid_increment DECIMAL(10,2) DEFAULT 1.00,
			total_bids INT DEFAULT 0,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP NOT NULL,
			status ENUM('draft', 'active', 'ended', 'cancelled') DEFAULT 'draft',
			winner_id INT NULL,
			view_count INT DEFAULT 0,
			is_reserve_met BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
			FOREIGN KEY (vendor_id) REFERENCES vendors(id) ON DELETE CASCADE,
			FOREIGN KEY (winner_id) REFERENCES users(id) ON DELETE SET NULL,
			INDEX idx_auctions_vendor (vendor_id),
			INDEX idx_auctions_status (status),
			INDEX idx_auctions_end_time (end_time),
			INDEX idx_auctions_product (product_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}

	for i, migration := range mysqlMigrations {
		log.Printf("Running MySQL migration %d/%d", i+1, len(mysqlMigrations))

		if _, err := mr.db.Exec(migration); err != nil {
			return fmt.Errorf("failed to run MySQL migration %d: %w", i+1, err)
		}
	}

	log.Println("✅ All MySQL migrations completed successfully")
	return nil
}

// RunMigrationsForGlobalDB runs migrations for the global database
func RunMigrationsForGlobalDB() error {
	if GlobalDBManager == nil {
		return fmt.Errorf("global database not initialized")
	}

	runner := NewMigrationRunner(GlobalDBManager.GetDB(), GlobalDBManager.GetType())
	return runner.RunMigrations()
}
