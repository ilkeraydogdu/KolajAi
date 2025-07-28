package migrations

// CreateAIVisionTables migration
var CreateAIVisionTables = `
-- Main AI image analysis table
CREATE TABLE IF NOT EXISTS ai_image_analysis (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	image_id VARCHAR(255) NOT NULL UNIQUE,
	user_id INTEGER NOT NULL,
	original_filename VARCHAR(500) NOT NULL,
	stored_filename VARCHAR(500) NOT NULL,
	file_size INTEGER NOT NULL,
	width INTEGER NOT NULL,
	height INTEGER NOT NULL,
	format VARCHAR(50) NOT NULL,
	hash VARCHAR(64) NOT NULL,
	detected_objects TEXT, -- JSON array of detected objects
	category_predictions TEXT, -- JSON array of category predictions
	color_analysis TEXT, -- JSON object of color analysis
	quality_score REAL NOT NULL DEFAULT 0.0,
	tags TEXT, -- JSON array of tags
	metadata TEXT, -- JSON object of additional metadata
	processing_time_ms INTEGER NOT NULL DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- User image categories association table
CREATE TABLE IF NOT EXISTS user_image_categories (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	image_id VARCHAR(255) NOT NULL,
	category_id INTEGER NOT NULL,
	confidence REAL NOT NULL DEFAULT 0.0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	FOREIGN KEY (image_id) REFERENCES ai_image_analysis(image_id) ON DELETE CASCADE,
	UNIQUE(user_id, image_id, category_id)
);

-- User image tags association table
CREATE TABLE IF NOT EXISTS user_image_tags (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	image_id VARCHAR(255) NOT NULL,
	tag VARCHAR(100) NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	FOREIGN KEY (image_id) REFERENCES ai_image_analysis(image_id) ON DELETE CASCADE,
	UNIQUE(user_id, image_id, tag)
);

-- User image collections table
CREATE TABLE IF NOT EXISTS user_image_collections (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	collection_id VARCHAR(255) NOT NULL UNIQUE,
	name VARCHAR(200) NOT NULL,
	description TEXT,
	image_ids TEXT, -- JSON array of image IDs
	is_public BOOLEAN DEFAULT FALSE,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- AI processing logs table for monitoring and debugging
CREATE TABLE IF NOT EXISTS ai_processing_logs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	image_id VARCHAR(255),
	operation VARCHAR(100) NOT NULL, -- 'upload', 'analysis', 'search', etc.
	status VARCHAR(50) NOT NULL, -- 'success', 'error', 'processing'
	processing_time_ms INTEGER,
	error_message TEXT,
	metadata TEXT, -- JSON object with additional info
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- AI model performance metrics table
CREATE TABLE IF NOT EXISTS ai_model_metrics (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	model_type VARCHAR(100) NOT NULL, -- 'object_detection', 'category_prediction', etc.
	model_version VARCHAR(50) NOT NULL,
	accuracy_score REAL,
	precision_score REAL,
	recall_score REAL,
	f1_score REAL,
	total_predictions INTEGER DEFAULT 0,
	correct_predictions INTEGER DEFAULT 0,
	last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
	metadata TEXT, -- JSON object with additional metrics
	UNIQUE(model_type, model_version)
);

-- User feedback on AI predictions (for model improvement)
CREATE TABLE IF NOT EXISTS ai_prediction_feedback (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	image_id VARCHAR(255) NOT NULL,
	prediction_type VARCHAR(100) NOT NULL, -- 'category', 'object', 'tag'
	predicted_value VARCHAR(200) NOT NULL,
	actual_value VARCHAR(200),
	is_correct BOOLEAN,
	confidence_score REAL,
	user_rating INTEGER, -- 1-5 rating from user
	feedback_text TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	FOREIGN KEY (image_id) REFERENCES ai_image_analysis(image_id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_ai_image_user ON ai_image_analysis(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_image_hash ON ai_image_analysis(hash);
CREATE INDEX IF NOT EXISTS idx_ai_image_created ON ai_image_analysis(created_at);
CREATE INDEX IF NOT EXISTS idx_ai_image_quality ON ai_image_analysis(quality_score);
CREATE INDEX IF NOT EXISTS idx_ai_image_format ON ai_image_analysis(format);
CREATE INDEX IF NOT EXISTS idx_ai_image_size ON ai_image_analysis(width, height);

CREATE INDEX IF NOT EXISTS idx_user_image_cat_user ON user_image_categories(user_id);
CREATE INDEX IF NOT EXISTS idx_user_image_cat_image ON user_image_categories(image_id);
CREATE INDEX IF NOT EXISTS idx_user_image_cat_category ON user_image_categories(category_id);
CREATE INDEX IF NOT EXISTS idx_user_image_cat_confidence ON user_image_categories(confidence);

CREATE INDEX IF NOT EXISTS idx_user_image_tag_user ON user_image_tags(user_id);
CREATE INDEX IF NOT EXISTS idx_user_image_tag_image ON user_image_tags(image_id);
CREATE INDEX IF NOT EXISTS idx_user_image_tag_tag ON user_image_tags(tag);

CREATE INDEX IF NOT EXISTS idx_user_collections_user ON user_image_collections(user_id);
CREATE INDEX IF NOT EXISTS idx_user_collections_public ON user_image_collections(is_public);
CREATE INDEX IF NOT EXISTS idx_user_collections_created ON user_image_collections(created_at);

CREATE INDEX IF NOT EXISTS idx_ai_logs_user ON ai_processing_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_logs_operation ON ai_processing_logs(operation);
CREATE INDEX IF NOT EXISTS idx_ai_logs_status ON ai_processing_logs(status);
CREATE INDEX IF NOT EXISTS idx_ai_logs_created ON ai_processing_logs(created_at);

CREATE INDEX IF NOT EXISTS idx_ai_feedback_user ON ai_prediction_feedback(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_feedback_image ON ai_prediction_feedback(image_id);
CREATE INDEX IF NOT EXISTS idx_ai_feedback_type ON ai_prediction_feedback(prediction_type);
CREATE INDEX IF NOT EXISTS idx_ai_feedback_correct ON ai_prediction_feedback(is_correct);

-- Insert initial AI model metrics
INSERT OR IGNORE INTO ai_model_metrics (model_type, model_version, accuracy_score, precision_score, recall_score, f1_score, metadata) 
VALUES ('object_detection', 'v1.0', 0.75, 0.72, 0.78, 0.75, '{"algorithm": "heuristic_based", "training_data": "internal"}');

INSERT OR IGNORE INTO ai_model_metrics (model_type, model_version, accuracy_score, precision_score, recall_score, f1_score, metadata) 
VALUES ('category_prediction', 'v1.0', 0.68, 0.65, 0.71, 0.68, '{"algorithm": "rule_based", "categories": "marketplace_specific"}');

INSERT OR IGNORE INTO ai_model_metrics (model_type, model_version, accuracy_score, precision_score, recall_score, f1_score, metadata) 
VALUES ('color_analysis', 'v1.0', 0.85, 0.83, 0.87, 0.85, '{"algorithm": "rgb_clustering", "color_space": "rgb"}');

INSERT OR IGNORE INTO ai_model_metrics (model_type, model_version, accuracy_score, precision_score, recall_score, f1_score, metadata) 
VALUES ('quality_assessment', 'v1.0', 0.72, 0.70, 0.74, 0.72, '{"algorithm": "composite_scoring", "factors": ["resolution", "contrast", "brightness", "saturation"]"}');
`