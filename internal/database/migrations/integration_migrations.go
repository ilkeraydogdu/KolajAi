package migrations

import (
	"database/sql"
	"fmt"
)

// CreateIntegrationTables creates tables for the integration system
func CreateIntegrationTables(db *sql.DB) error {
	queries := []string{
		// Integration credentials table
		`CREATE TABLE IF NOT EXISTS integration_credentials (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			integration_id VARCHAR(100) UNIQUE NOT NULL,
			encrypted_data BLOB NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			rotated_at TIMESTAMP,
			INDEX idx_integration_id (integration_id)
		)`,

		// Integration configurations table
		`CREATE TABLE IF NOT EXISTS integration_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			integration_id VARCHAR(100) UNIQUE NOT NULL,
			integration_type VARCHAR(50) NOT NULL,
			provider VARCHAR(100) NOT NULL,
			name VARCHAR(255) NOT NULL,
			status VARCHAR(20) DEFAULT 'inactive',
			config JSON,
			metadata JSON,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_integration_type (integration_type),
			INDEX idx_provider (provider),
			INDEX idx_status (status)
		)`,

		// Integration audit log table
		`CREATE TABLE IF NOT EXISTS integration_audit_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			integration_id VARCHAR(100) NOT NULL,
			action VARCHAR(50) NOT NULL,
			user_id INTEGER,
			ip_address VARCHAR(45),
			user_agent TEXT,
			request_data JSON,
			response_data JSON,
			error_message TEXT,
			status_code INTEGER,
			duration_ms INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_integration_audit (integration_id, created_at),
			INDEX idx_action (action),
			INDEX idx_user_id (user_id)
		)`,

		// Webhook events table
		`CREATE TABLE IF NOT EXISTS webhook_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			event_id VARCHAR(100) UNIQUE NOT NULL,
			integration_id VARCHAR(100) NOT NULL,
			event_type VARCHAR(50) NOT NULL,
			payload JSON NOT NULL,
			headers JSON,
			signature VARCHAR(500),
			status VARCHAR(20) DEFAULT 'pending',
			processed_at TIMESTAMP,
			retry_count INTEGER DEFAULT 0,
			error_message TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_webhook_integration (integration_id),
			INDEX idx_webhook_status (status),
			INDEX idx_webhook_created (created_at)
		)`,

		// Integration metrics table
		`CREATE TABLE IF NOT EXISTS integration_metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			integration_id VARCHAR(100) NOT NULL,
			metric_type VARCHAR(50) NOT NULL,
			metric_name VARCHAR(100) NOT NULL,
			value DECIMAL(20,4),
			tags JSON,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_metrics_integration (integration_id, timestamp),
			INDEX idx_metrics_type (metric_type)
		)`,

		// API rate limits table
		`CREATE TABLE IF NOT EXISTS integration_rate_limits (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			integration_id VARCHAR(100) NOT NULL,
			endpoint VARCHAR(255),
			requests_per_minute INTEGER,
			requests_remaining INTEGER,
			resets_at TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY unique_integration_endpoint (integration_id, endpoint)
		)`,

		// Integration health checks table
		`CREATE TABLE IF NOT EXISTS integration_health_checks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			integration_id VARCHAR(100) NOT NULL,
			status VARCHAR(20) NOT NULL,
			response_time_ms INTEGER,
			error_message TEXT,
			checked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_health_integration (integration_id, checked_at)
		)`,

		// Payment transactions table (for payment integrations)
		`CREATE TABLE IF NOT EXISTS payment_transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			transaction_id VARCHAR(100) UNIQUE NOT NULL,
			integration_id VARCHAR(100) NOT NULL,
			order_id INTEGER,
			payment_method VARCHAR(50),
			amount DECIMAL(10,2) NOT NULL,
			currency VARCHAR(3) NOT NULL,
			status VARCHAR(20) NOT NULL,
			gateway_response JSON,
			metadata JSON,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_payment_order (order_id),
			INDEX idx_payment_status (status),
			INDEX idx_payment_created (created_at)
		)`,

		// Integration user mappings table
		`CREATE TABLE IF NOT EXISTS integration_user_mappings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			integration_id VARCHAR(100) NOT NULL,
			external_user_id VARCHAR(255),
			access_token TEXT,
			refresh_token TEXT,
			token_expires_at TIMESTAMP,
			metadata JSON,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY unique_user_integration (user_id, integration_id),
			INDEX idx_external_user (external_user_id)
		)`,

		// Integration queue jobs table
		`CREATE TABLE IF NOT EXISTS integration_queue_jobs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			job_id VARCHAR(100) UNIQUE NOT NULL,
			integration_id VARCHAR(100) NOT NULL,
			job_type VARCHAR(50) NOT NULL,
			payload JSON NOT NULL,
			status VARCHAR(20) DEFAULT 'pending',
			priority INTEGER DEFAULT 0,
			retry_count INTEGER DEFAULT 0,
			max_retries INTEGER DEFAULT 3,
			error_message TEXT,
			scheduled_at TIMESTAMP,
			started_at TIMESTAMP,
			completed_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_queue_status (status, priority, scheduled_at),
			INDEX idx_queue_integration (integration_id)
		)`,
	}

	// Execute each query
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	return nil
}

// DropIntegrationTables drops all integration-related tables
func DropIntegrationTables(db *sql.DB) error {
	tables := []string{
		"integration_queue_jobs",
		"integration_user_mappings",
		"payment_transactions",
		"integration_health_checks",
		"integration_rate_limits",
		"integration_metrics",
		"webhook_events",
		"integration_audit_logs",
		"integration_configs",
		"integration_credentials",
	}

	for _, table := range tables {
		query := fmt.Sprintf("DROP TABLE IF EXISTS %s", table)
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}
