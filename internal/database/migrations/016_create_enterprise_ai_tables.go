package migrations

// CreateEnterpriseAITables migration
var CreateEnterpriseAITables = `
-- Customer service requests table
CREATE TABLE IF NOT EXISTS customer_service_requests (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	type VARCHAR(50) NOT NULL, -- 'complaint', 'question', 'suggestion', 'technical'
	subject VARCHAR(500) NOT NULL,
	description TEXT NOT NULL,
	priority VARCHAR(20) NOT NULL DEFAULT 'medium', -- 'low', 'medium', 'high', 'urgent'
	status VARCHAR(20) NOT NULL DEFAULT 'open', -- 'open', 'in_progress', 'resolved', 'closed'
	category VARCHAR(50),
	tags TEXT, -- JSON array
	attachments TEXT, -- JSON array of image IDs
	ai_analysis TEXT, -- JSON object with AI analysis
	assigned_to INTEGER,
	resolution TEXT,
	satisfaction_score REAL DEFAULT 0.0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	resolved_at DATETIME,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	FOREIGN KEY (assigned_to) REFERENCES users(id) ON DELETE SET NULL
);

-- Content moderation results table
CREATE TABLE IF NOT EXISTS content_moderation_results (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	content_id VARCHAR(255) NOT NULL,
	content_type VARCHAR(50) NOT NULL, -- 'text', 'image', 'video'
	user_id INTEGER,
	is_appropriate BOOLEAN NOT NULL DEFAULT TRUE,
	confidence_score REAL NOT NULL DEFAULT 0.0,
	violations TEXT, -- JSON array of violations
	recommendations TEXT, -- JSON array of recommendations
	action_required VARCHAR(20) NOT NULL DEFAULT 'none', -- 'none', 'review', 'block', 'remove'
	action_taken VARCHAR(20) DEFAULT 'none',
	reviewed_by INTEGER,
	processed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	reviewed_at DATETIME,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
	FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL
);

-- Business insights table
CREATE TABLE IF NOT EXISTS business_insights (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	type VARCHAR(50) NOT NULL, -- 'trend', 'opportunity', 'risk', 'recommendation'
	title VARCHAR(500) NOT NULL,
	description TEXT NOT NULL,
	impact VARCHAR(20) NOT NULL, -- 'low', 'medium', 'high'
	confidence REAL NOT NULL DEFAULT 0.0,
	data TEXT, -- JSON object with supporting data
	action_items TEXT, -- JSON array of action items
	category VARCHAR(50) NOT NULL,
	status VARCHAR(20) DEFAULT 'active', -- 'active', 'acknowledged', 'resolved', 'dismissed'
	acknowledged_by INTEGER,
	acknowledged_at DATETIME,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	expires_at DATETIME,
	FOREIGN KEY (acknowledged_by) REFERENCES users(id) ON DELETE SET NULL
);

-- Automated tasks table
CREATE TABLE IF NOT EXISTS automated_tasks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	type VARCHAR(100) NOT NULL, -- 'content_optimization', 'inventory_management', etc.
	status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'running', 'completed', 'failed'
	progress REAL DEFAULT 0.0, -- 0 to 1
	results TEXT, -- JSON object with results
	error_message TEXT,
	scheduled_at DATETIME NOT NULL,
	started_at DATETIME,
	completed_at DATETIME,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- AI performance metrics tracking
CREATE TABLE IF NOT EXISTS ai_performance_metrics (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	service_type VARCHAR(50) NOT NULL, -- 'vision', 'nlp', 'recommendation', etc.
	operation VARCHAR(100) NOT NULL, -- 'image_analysis', 'sentiment_analysis', etc.
	processing_time_ms INTEGER NOT NULL,
	accuracy_score REAL,
	confidence_score REAL,
	success BOOLEAN NOT NULL DEFAULT TRUE,
	error_type VARCHAR(100),
	user_id INTEGER,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Enterprise AI configuration table
CREATE TABLE IF NOT EXISTS ai_configuration (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	service_name VARCHAR(100) NOT NULL UNIQUE,
	configuration TEXT NOT NULL, -- JSON configuration
	is_enabled BOOLEAN DEFAULT TRUE,
	version VARCHAR(20) DEFAULT '1.0',
	last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_by INTEGER,
	FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL
);

-- AI training data table for continuous learning
CREATE TABLE IF NOT EXISTS ai_training_data (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	data_type VARCHAR(50) NOT NULL, -- 'image_classification', 'sentiment', 'category_prediction'
	input_data TEXT NOT NULL, -- JSON or base64 encoded data
	expected_output TEXT NOT NULL, -- JSON with expected results
	actual_output TEXT, -- JSON with actual AI output
	confidence_score REAL,
	is_correct BOOLEAN,
	user_feedback TEXT,
	source VARCHAR(50) DEFAULT 'user_feedback', -- 'user_feedback', 'admin_review', 'automated'
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	reviewed_at DATETIME,
	reviewed_by INTEGER,
	FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_customer_service_user ON customer_service_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_customer_service_status ON customer_service_requests(status);
CREATE INDEX IF NOT EXISTS idx_customer_service_priority ON customer_service_requests(priority);
CREATE INDEX IF NOT EXISTS idx_customer_service_category ON customer_service_requests(category);
CREATE INDEX IF NOT EXISTS idx_customer_service_created ON customer_service_requests(created_at);
CREATE INDEX IF NOT EXISTS idx_customer_service_assigned ON customer_service_requests(assigned_to);

CREATE INDEX IF NOT EXISTS idx_content_moderation_content ON content_moderation_results(content_id);
CREATE INDEX IF NOT EXISTS idx_content_moderation_type ON content_moderation_results(content_type);
CREATE INDEX IF NOT EXISTS idx_content_moderation_user ON content_moderation_results(user_id);
CREATE INDEX IF NOT EXISTS idx_content_moderation_appropriate ON content_moderation_results(is_appropriate);
CREATE INDEX IF NOT EXISTS idx_content_moderation_action ON content_moderation_results(action_required);
CREATE INDEX IF NOT EXISTS idx_content_moderation_processed ON content_moderation_results(processed_at);

CREATE INDEX IF NOT EXISTS idx_business_insights_type ON business_insights(type);
CREATE INDEX IF NOT EXISTS idx_business_insights_impact ON business_insights(impact);
CREATE INDEX IF NOT EXISTS idx_business_insights_category ON business_insights(category);
CREATE INDEX IF NOT EXISTS idx_business_insights_status ON business_insights(status);
CREATE INDEX IF NOT EXISTS idx_business_insights_created ON business_insights(created_at);
CREATE INDEX IF NOT EXISTS idx_business_insights_expires ON business_insights(expires_at);

CREATE INDEX IF NOT EXISTS idx_automated_tasks_type ON automated_tasks(type);
CREATE INDEX IF NOT EXISTS idx_automated_tasks_status ON automated_tasks(status);
CREATE INDEX IF NOT EXISTS idx_automated_tasks_scheduled ON automated_tasks(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_automated_tasks_completed ON automated_tasks(completed_at);

CREATE INDEX IF NOT EXISTS idx_ai_performance_service ON ai_performance_metrics(service_type);
CREATE INDEX IF NOT EXISTS idx_ai_performance_operation ON ai_performance_metrics(operation);
CREATE INDEX IF NOT EXISTS idx_ai_performance_user ON ai_performance_metrics(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_performance_created ON ai_performance_metrics(created_at);
CREATE INDEX IF NOT EXISTS idx_ai_performance_success ON ai_performance_metrics(success);

CREATE INDEX IF NOT EXISTS idx_ai_training_type ON ai_training_data(data_type);
CREATE INDEX IF NOT EXISTS idx_ai_training_correct ON ai_training_data(is_correct);
CREATE INDEX IF NOT EXISTS idx_ai_training_source ON ai_training_data(source);
CREATE INDEX IF NOT EXISTS idx_ai_training_created ON ai_training_data(created_at);
CREATE INDEX IF NOT EXISTS idx_ai_training_reviewed ON ai_training_data(reviewed_by);

-- Insert initial AI configuration
INSERT OR IGNORE INTO ai_configuration (service_name, configuration, is_enabled, version) 
VALUES ('vision_service', '{"max_file_size": 10485760, "allowed_types": ["image/jpeg", "image/png", "image/gif", "image/webp"], "quality_threshold": 0.5}', TRUE, '1.0');

INSERT OR IGNORE INTO ai_configuration (service_name, configuration, is_enabled, version) 
VALUES ('enterprise_service', '{"sentiment_threshold": 0.5, "urgency_threshold": 0.7, "auto_assignment": true, "moderation_enabled": true}', TRUE, '1.0');

INSERT OR IGNORE INTO ai_configuration (service_name, configuration, is_enabled, version) 
VALUES ('content_moderation', '{"spam_threshold": 0.8, "inappropriate_threshold": 0.9, "offensive_threshold": 0.85, "auto_action": false}', TRUE, '1.0');

INSERT OR IGNORE INTO ai_configuration (service_name, configuration, is_enabled, version) 
VALUES ('business_insights', '{"update_frequency": "daily", "confidence_threshold": 0.7, "max_insights": 50, "retention_days": 30}', TRUE, '1.0');
`