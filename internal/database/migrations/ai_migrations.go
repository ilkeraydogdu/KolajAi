package migrations

// CreateAIAdvancedTables creates AI-related tables
var CreateAIAdvancedTables = `
		-- AI Credits table
		CREATE TABLE IF NOT EXISTS ai_credits (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			credits INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		-- AI Credit Transactions table
		CREATE TABLE IF NOT EXISTS ai_credit_transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			type VARCHAR(50) NOT NULL,
			amount INTEGER NOT NULL,
			description TEXT,
			reference_type VARCHAR(50),
			reference_id INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		-- AI Generated Content table
		CREATE TABLE IF NOT EXISTS ai_generated_content (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			type VARCHAR(50) NOT NULL,
			model VARCHAR(100) NOT NULL,
			prompt TEXT,
			content TEXT,
			metadata TEXT,
			credits INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		-- AI Templates table
		CREATE TABLE IF NOT EXISTS ai_templates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL,
			design TEXT,
			thumbnail VARCHAR(500),
			is_public BOOLEAN DEFAULT FALSE,
			usage_count INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		-- AI Chat Sessions table
		CREATE TABLE IF NOT EXISTS ai_chat_sessions (
			id VARCHAR(100) PRIMARY KEY,
			user_id INTEGER NOT NULL,
			context VARCHAR(100),
			messages TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		-- Marketplace Integration Configs table
		CREATE TABLE IF NOT EXISTS marketplace_integration_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			integration_id VARCHAR(100) NOT NULL,
			credentials TEXT,
			settings TEXT,
			is_active BOOLEAN DEFAULT TRUE,
			last_sync TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		-- Marketplace Sync Logs table
		CREATE TABLE IF NOT EXISTS marketplace_sync_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			integration_id VARCHAR(100) NOT NULL,
			sync_type VARCHAR(50) NOT NULL,
			status VARCHAR(50) NOT NULL,
			details TEXT,
			error_message TEXT,
			items_count INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		-- Create indexes
		CREATE INDEX idx_ai_credits_user_id ON ai_credits(user_id);
		CREATE INDEX idx_ai_credit_transactions_user_id ON ai_credit_transactions(user_id);
		CREATE INDEX idx_ai_generated_content_user_id ON ai_generated_content(user_id);
		CREATE INDEX idx_ai_templates_user_id ON ai_templates(user_id);
		CREATE INDEX idx_ai_chat_sessions_user_id ON ai_chat_sessions(user_id);
		CREATE INDEX idx_marketplace_configs_user_id ON marketplace_integration_configs(user_id);
		CREATE INDEX idx_marketplace_logs_user_id ON marketplace_sync_logs(user_id);

		-- Initialize AI credits for existing users
		INSERT INTO ai_credits (user_id, credits)
		SELECT id, 100 FROM users
		WHERE id NOT IN (SELECT user_id FROM ai_credits);
	`
