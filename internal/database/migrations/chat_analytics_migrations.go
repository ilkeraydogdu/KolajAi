package migrations

// Chat and Analytics Tables Migration
var CreateChatAnalyticsTables = `
-- Chat Sessions Table
CREATE TABLE IF NOT EXISTS chat_sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id INTEGER NOT NULL,
    title VARCHAR(500) NOT NULL,
    context VARCHAR(50) DEFAULT 'general',
    status VARCHAR(20) DEFAULT 'active',
    metadata TEXT DEFAULT '{}',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_message DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Chat Messages Table
CREATE TABLE IF NOT EXISTS chat_messages (
    id VARCHAR(255) PRIMARY KEY,
    session_id VARCHAR(255) NOT NULL,
    user_id INTEGER NOT NULL,
    role VARCHAR(20) NOT NULL, -- user, assistant, system
    content TEXT NOT NULL,
    message_type VARCHAR(50) DEFAULT 'text',
    metadata TEXT DEFAULT '{}',
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    tokens INTEGER DEFAULT 0,
    FOREIGN KEY (session_id) REFERENCES chat_sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- User Sessions Table (for analytics)
CREATE TABLE IF NOT EXISTS user_sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id INTEGER,
    session_token VARCHAR(255) UNIQUE,
    ip_address VARCHAR(45),
    user_agent TEXT,
    device_type VARCHAR(50),
    browser VARCHAR(100),
    os VARCHAR(100),
    country VARCHAR(100),
    city VARCHAR(100),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_activity DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT 1,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Page Views Table (for analytics)
CREATE TABLE IF NOT EXISTS page_views (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id VARCHAR(255),
    user_id INTEGER,
    path VARCHAR(500) NOT NULL,
    title VARCHAR(500),
    referrer VARCHAR(500),
    query_params TEXT,
    duration INTEGER DEFAULT 0, -- in seconds
    bounce BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES user_sessions(id) ON DELETE SET NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Events Table (for analytics)
CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id VARCHAR(255),
    user_id INTEGER,
    event_type VARCHAR(100) NOT NULL,
    event_name VARCHAR(200) NOT NULL,
    event_data TEXT DEFAULT '{}',
    page_path VARCHAR(500),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES user_sessions(id) ON DELETE SET NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Business Metrics Cache Table
CREATE TABLE IF NOT EXISTS business_metrics_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    metric_type VARCHAR(100) NOT NULL,
    period_type VARCHAR(50) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    data TEXT NOT NULL, -- JSON data
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL,
    UNIQUE(metric_type, period_type, start_date, end_date)
);

-- Customer Segments Table
CREATE TABLE IF NOT EXISTS customer_segments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    segment_name VARCHAR(100) NOT NULL,
    segment_score DECIMAL(5,2) DEFAULT 0,
    rfm_recency INTEGER DEFAULT 0,
    rfm_frequency INTEGER DEFAULT 0,
    rfm_monetary DECIMAL(10,2) DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id)
);

-- A/B Test Experiments Table
CREATE TABLE IF NOT EXISTS ab_experiments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'draft', -- draft, active, paused, completed
    traffic_allocation DECIMAL(3,2) DEFAULT 0.5, -- 0.0 to 1.0
    control_variant TEXT NOT NULL, -- JSON
    test_variants TEXT NOT NULL, -- JSON array
    success_metrics TEXT, -- JSON array
    start_date DATETIME,
    end_date DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- A/B Test Assignments Table
CREATE TABLE IF NOT EXISTS ab_assignments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    experiment_id INTEGER NOT NULL,
    user_id INTEGER,
    session_id VARCHAR(255),
    variant_name VARCHAR(100) NOT NULL,
    assigned_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (experiment_id) REFERENCES ab_experiments(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (session_id) REFERENCES user_sessions(id) ON DELETE SET NULL,
    UNIQUE(experiment_id, user_id),
    UNIQUE(experiment_id, session_id)
);

-- A/B Test Results Table
CREATE TABLE IF NOT EXISTS ab_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    experiment_id INTEGER NOT NULL,
    variant_name VARCHAR(100) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL(15,4) NOT NULL,
    user_count INTEGER DEFAULT 0,
    conversion_count INTEGER DEFAULT 0,
    recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (experiment_id) REFERENCES ab_experiments(id) ON DELETE CASCADE
);

-- Product Views Table (for recommendations)
CREATE TABLE IF NOT EXISTS product_views (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    session_id VARCHAR(255),
    product_id INTEGER NOT NULL,
    view_duration INTEGER DEFAULT 0, -- in seconds
    source VARCHAR(100), -- search, category, recommendation, etc.
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (session_id) REFERENCES user_sessions(id) ON DELETE SET NULL,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

-- Search Queries Table (for analytics and improvements)
CREATE TABLE IF NOT EXISTS search_queries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    session_id VARCHAR(255),
    query VARCHAR(500) NOT NULL,
    results_count INTEGER DEFAULT 0,
    clicked_result_position INTEGER,
    clicked_product_id INTEGER,
    no_results BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (session_id) REFERENCES user_sessions(id) ON DELETE SET NULL,
    FOREIGN KEY (clicked_product_id) REFERENCES products(id) ON DELETE SET NULL
);

-- Cart Abandonment Table
CREATE TABLE IF NOT EXISTS cart_abandonment (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    session_id VARCHAR(255),
    cart_data TEXT NOT NULL, -- JSON of cart contents
    cart_value DECIMAL(10,2) DEFAULT 0,
    abandonment_stage VARCHAR(100), -- product_page, cart, checkout, payment
    recovery_email_sent BOOLEAN DEFAULT 0,
    recovered BOOLEAN DEFAULT 0,
    recovered_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (session_id) REFERENCES user_sessions(id) ON DELETE SET NULL
);

-- Email Campaign Analytics
CREATE TABLE IF NOT EXISTS email_campaigns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(200) NOT NULL,
    subject VARCHAR(500) NOT NULL,
    template_id VARCHAR(100),
    segment_criteria TEXT, -- JSON
    status VARCHAR(20) DEFAULT 'draft', -- draft, scheduled, sending, sent, completed
    scheduled_at DATETIME,
    sent_at DATETIME,
    total_recipients INTEGER DEFAULT 0,
    delivered_count INTEGER DEFAULT 0,
    opened_count INTEGER DEFAULT 0,
    clicked_count INTEGER DEFAULT 0,
    unsubscribed_count INTEGER DEFAULT 0,
    bounced_count INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Email Campaign Recipients
CREATE TABLE IF NOT EXISTS email_campaign_recipients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    email VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending', -- pending, sent, delivered, opened, clicked, bounced, unsubscribed
    sent_at DATETIME,
    delivered_at DATETIME,
    opened_at DATETIME,
    clicked_at DATETIME,
    unsubscribed_at DATETIME,
    bounce_reason TEXT,
    FOREIGN KEY (campaign_id) REFERENCES email_campaigns(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Real-time Analytics Aggregations (for performance)
CREATE TABLE IF NOT EXISTS analytics_aggregations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    metric_name VARCHAR(100) NOT NULL,
    dimension VARCHAR(100), -- date, hour, product_id, category_id, etc.
    dimension_value VARCHAR(255),
    metric_value DECIMAL(15,4) NOT NULL,
    count_value INTEGER DEFAULT 0,
    aggregation_date DATE NOT NULL,
    aggregation_hour INTEGER, -- 0-23
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(metric_name, dimension, dimension_value, aggregation_date, aggregation_hour)
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_chat_sessions_user_id ON chat_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_chat_sessions_status ON chat_sessions(status);
CREATE INDEX IF NOT EXISTS idx_chat_sessions_updated_at ON chat_sessions(updated_at);

CREATE INDEX IF NOT EXISTS idx_chat_messages_session_id ON chat_messages(session_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_user_id ON chat_messages(user_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_timestamp ON chat_messages(timestamp);
CREATE INDEX IF NOT EXISTS idx_chat_messages_role ON chat_messages(role);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_last_activity ON user_sessions(last_activity);
CREATE INDEX IF NOT EXISTS idx_user_sessions_is_active ON user_sessions(is_active);
CREATE INDEX IF NOT EXISTS idx_user_sessions_created_at ON user_sessions(created_at);

CREATE INDEX IF NOT EXISTS idx_page_views_session_id ON page_views(session_id);
CREATE INDEX IF NOT EXISTS idx_page_views_user_id ON page_views(user_id);
CREATE INDEX IF NOT EXISTS idx_page_views_created_at ON page_views(created_at);
CREATE INDEX IF NOT EXISTS idx_page_views_path ON page_views(path);

CREATE INDEX IF NOT EXISTS idx_events_session_id ON events(session_id);
CREATE INDEX IF NOT EXISTS idx_events_user_id ON events(user_id);
CREATE INDEX IF NOT EXISTS idx_events_event_type ON events(event_type);
CREATE INDEX IF NOT EXISTS idx_events_created_at ON events(created_at);

CREATE INDEX IF NOT EXISTS idx_business_metrics_cache_type_period ON business_metrics_cache(metric_type, period_type);
CREATE INDEX IF NOT EXISTS idx_business_metrics_cache_dates ON business_metrics_cache(start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_business_metrics_cache_expires_at ON business_metrics_cache(expires_at);

CREATE INDEX IF NOT EXISTS idx_customer_segments_user_id ON customer_segments(user_id);
CREATE INDEX IF NOT EXISTS idx_customer_segments_segment_name ON customer_segments(segment_name);
CREATE INDEX IF NOT EXISTS idx_customer_segments_updated_at ON customer_segments(updated_at);

CREATE INDEX IF NOT EXISTS idx_product_views_user_id ON product_views(user_id);
CREATE INDEX IF NOT EXISTS idx_product_views_product_id ON product_views(product_id);
CREATE INDEX IF NOT EXISTS idx_product_views_created_at ON product_views(created_at);
CREATE INDEX IF NOT EXISTS idx_product_views_session_id ON product_views(session_id);

CREATE INDEX IF NOT EXISTS idx_search_queries_user_id ON search_queries(user_id);
CREATE INDEX IF NOT EXISTS idx_search_queries_query ON search_queries(query);
CREATE INDEX IF NOT EXISTS idx_search_queries_created_at ON search_queries(created_at);
CREATE INDEX IF NOT EXISTS idx_search_queries_no_results ON search_queries(no_results);

CREATE INDEX IF NOT EXISTS idx_cart_abandonment_user_id ON cart_abandonment(user_id);
CREATE INDEX IF NOT EXISTS idx_cart_abandonment_created_at ON cart_abandonment(created_at);
CREATE INDEX IF NOT EXISTS idx_cart_abandonment_recovered ON cart_abandonment(recovered);

CREATE INDEX IF NOT EXISTS idx_email_campaigns_status ON email_campaigns(status);
CREATE INDEX IF NOT EXISTS idx_email_campaigns_scheduled_at ON email_campaigns(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_email_campaigns_sent_at ON email_campaigns(sent_at);

CREATE INDEX IF NOT EXISTS idx_email_campaign_recipients_campaign_id ON email_campaign_recipients(campaign_id);
CREATE INDEX IF NOT EXISTS idx_email_campaign_recipients_user_id ON email_campaign_recipients(user_id);
CREATE INDEX IF NOT EXISTS idx_email_campaign_recipients_status ON email_campaign_recipients(status);

CREATE INDEX IF NOT EXISTS idx_analytics_aggregations_metric ON analytics_aggregations(metric_name);
CREATE INDEX IF NOT EXISTS idx_analytics_aggregations_date ON analytics_aggregations(aggregation_date);
CREATE INDEX IF NOT EXISTS idx_analytics_aggregations_dimension ON analytics_aggregations(dimension, dimension_value);

-- Triggers for updated_at timestamps
CREATE TRIGGER IF NOT EXISTS update_chat_sessions_updated_at 
    AFTER UPDATE ON chat_sessions
    BEGIN
        UPDATE chat_sessions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_user_sessions_last_activity 
    AFTER UPDATE ON user_sessions
    BEGIN
        UPDATE user_sessions SET last_activity = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_business_metrics_cache_updated_at 
    AFTER UPDATE ON business_metrics_cache
    BEGIN
        UPDATE business_metrics_cache SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_customer_segments_updated_at 
    AFTER UPDATE ON customer_segments
    BEGIN
        UPDATE customer_segments SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_ab_experiments_updated_at 
    AFTER UPDATE ON ab_experiments
    BEGIN
        UPDATE ab_experiments SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_email_campaigns_updated_at 
    AFTER UPDATE ON email_campaigns
    BEGIN
        UPDATE email_campaigns SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_analytics_aggregations_updated_at 
    AFTER UPDATE ON analytics_aggregations
    BEGIN
        UPDATE analytics_aggregations SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;
`
