package migrations

import (
	"database/sql"
)

// AddPerformanceIndexes adds indexes for performance optimization
func AddPerformanceIndexes(db *sql.DB) error {
	queries := []string{
		// User indexes
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_users_role ON users(role)`,

		// Product indexes
		`CREATE INDEX IF NOT EXISTS idx_products_vendor_id ON products(vendor_id)`,
		`CREATE INDEX IF NOT EXISTS idx_products_category ON products(category)`,
		`CREATE INDEX IF NOT EXISTS idx_products_price ON products(price)`,
		`CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_products_status ON products(status)`,
		`CREATE INDEX IF NOT EXISTS idx_products_stock ON products(stock)`,

		// Order indexes
		`CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_vendor_id ON orders(vendor_id)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_total_amount ON orders(total_amount)`,

		// Session indexes
		`CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at)`,

		// Vendor indexes
		`CREATE INDEX IF NOT EXISTS idx_vendors_user_id ON vendors(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_vendors_status ON vendors(status)`,
		`CREATE INDEX IF NOT EXISTS idx_vendors_created_at ON vendors(created_at)`,

		// Auction indexes
		`CREATE INDEX IF NOT EXISTS idx_auctions_product_id ON auctions(product_id)`,
		`CREATE INDEX IF NOT EXISTS idx_auctions_vendor_id ON auctions(vendor_id)`,
		`CREATE INDEX IF NOT EXISTS idx_auctions_status ON auctions(status)`,
		`CREATE INDEX IF NOT EXISTS idx_auctions_end_time ON auctions(end_time)`,

		// Wholesale indexes
		`CREATE INDEX IF NOT EXISTS idx_wholesale_deals_product_id ON wholesale_deals(product_id)`,
		`CREATE INDEX IF NOT EXISTS idx_wholesale_deals_vendor_id ON wholesale_deals(vendor_id)`,
		`CREATE INDEX IF NOT EXISTS idx_wholesale_deals_is_active ON wholesale_deals(is_active)`,

		// Integration indexes
		`CREATE INDEX IF NOT EXISTS idx_integration_logs_integration_id ON integration_logs(integration_id)`,
		`CREATE INDEX IF NOT EXISTS idx_integration_logs_timestamp ON integration_logs(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_integration_logs_status ON integration_logs(status)`,

		// AI indexes
		`CREATE INDEX IF NOT EXISTS idx_ai_usage_logs_user_id ON ai_usage_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ai_usage_logs_service_type ON ai_usage_logs(service_type)`,
		`CREATE INDEX IF NOT EXISTS idx_ai_usage_logs_created_at ON ai_usage_logs(created_at)`,

		// Composite indexes for common queries
		`CREATE INDEX IF NOT EXISTS idx_products_vendor_category ON products(vendor_id, category)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_user_status ON orders(user_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_vendor_status ON orders(vendor_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_products_category_price ON products(category, price)`,
		`CREATE INDEX IF NOT EXISTS idx_integration_logs_integration_status ON integration_logs(integration_id, status)`,

		// Full-text search indexes (for MySQL)
		// Note: SQLite doesn't support FULLTEXT indexes, so these are commented out
		// `CREATE FULLTEXT INDEX IF NOT EXISTS idx_products_fulltext ON products(name, description)`,
		// `CREATE FULLTEXT INDEX IF NOT EXISTS idx_vendors_fulltext ON vendors(company_name, description)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	// Analyze tables to update statistics (SQLite specific)
	if _, err := db.Exec("ANALYZE"); err != nil {
		return err
	}

	return nil
}

// RemovePerformanceIndexes removes performance indexes (for rollback)
func RemovePerformanceIndexes(db *sql.DB) error {
	queries := []string{
		`DROP INDEX IF EXISTS idx_users_email`,
		`DROP INDEX IF EXISTS idx_users_created_at`,
		`DROP INDEX IF EXISTS idx_users_role`,
		`DROP INDEX IF EXISTS idx_products_vendor_id`,
		`DROP INDEX IF EXISTS idx_products_category`,
		`DROP INDEX IF EXISTS idx_products_price`,
		`DROP INDEX IF EXISTS idx_products_created_at`,
		`DROP INDEX IF EXISTS idx_products_status`,
		`DROP INDEX IF EXISTS idx_products_stock`,
		`DROP INDEX IF EXISTS idx_orders_user_id`,
		`DROP INDEX IF EXISTS idx_orders_vendor_id`,
		`DROP INDEX IF EXISTS idx_orders_status`,
		`DROP INDEX IF EXISTS idx_orders_created_at`,
		`DROP INDEX IF EXISTS idx_orders_total_amount`,
		`DROP INDEX IF EXISTS idx_sessions_token`,
		`DROP INDEX IF EXISTS idx_sessions_user_id`,
		`DROP INDEX IF EXISTS idx_sessions_expires_at`,
		`DROP INDEX IF EXISTS idx_vendors_user_id`,
		`DROP INDEX IF EXISTS idx_vendors_status`,
		`DROP INDEX IF EXISTS idx_vendors_created_at`,
		`DROP INDEX IF EXISTS idx_auctions_product_id`,
		`DROP INDEX IF EXISTS idx_auctions_vendor_id`,
		`DROP INDEX IF EXISTS idx_auctions_status`,
		`DROP INDEX IF EXISTS idx_auctions_end_time`,
		`DROP INDEX IF EXISTS idx_wholesale_deals_product_id`,
		`DROP INDEX IF EXISTS idx_wholesale_deals_vendor_id`,
		`DROP INDEX IF EXISTS idx_wholesale_deals_is_active`,
		`DROP INDEX IF EXISTS idx_integration_logs_integration_id`,
		`DROP INDEX IF EXISTS idx_integration_logs_timestamp`,
		`DROP INDEX IF EXISTS idx_integration_logs_status`,
		`DROP INDEX IF EXISTS idx_ai_usage_logs_user_id`,
		`DROP INDEX IF EXISTS idx_ai_usage_logs_service_type`,
		`DROP INDEX IF EXISTS idx_ai_usage_logs_created_at`,
		`DROP INDEX IF EXISTS idx_products_vendor_category`,
		`DROP INDEX IF EXISTS idx_orders_user_status`,
		`DROP INDEX IF EXISTS idx_orders_vendor_status`,
		`DROP INDEX IF EXISTS idx_products_category_price`,
		`DROP INDEX IF EXISTS idx_integration_logs_integration_status`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}
