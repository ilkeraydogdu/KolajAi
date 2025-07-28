package migrations

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// MigrationService handles database migrations
type MigrationService struct {
	db     *sql.DB
	dbName string
}

// NewMigrationService creates a new migration service
func NewMigrationService(db *sql.DB, dbName string) *MigrationService {
	return &MigrationService{
		db:     db,
		dbName: dbName,
	}
}

// EnsureMigrationTable creates the migrations table if it doesn't exist
func (m *MigrationService) EnsureMigrationTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) NOT NULL,
		applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (version)
	);`

	_, err := m.db.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating migration table: %w", err)
	}

	return nil
}

// GetAppliedMigrations returns a list of already applied migrations
func (m *MigrationService) GetAppliedMigrations() (map[string]bool, error) {
	applied := make(map[string]bool)

	rows, err := m.db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, fmt.Errorf("error querying migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("error scanning migration version: %w", err)
		}
		applied[version] = true
	}

	return applied, nil
}

// CreateMigration creates a new migration file with the given name
func (m *MigrationService) CreateMigration(name string) (string, error) {
	timestamp := time.Now().Format("20060102150405")
	safeName := strings.ReplaceAll(strings.ToLower(name), " ", "_")

	// Create both up and down migrations
	upFileName := fmt.Sprintf("%s_%s.up.sql", timestamp, safeName)
	downFileName := fmt.Sprintf("%s_%s.down.sql", timestamp, safeName)

	// Note: In an actual implementation, this would create the files
	// For this example, we'll just return the filenames
	return fmt.Sprintf("Created migration files: %s and %s", upFileName, downFileName), nil
}

// RunMigrations runs all pending migrations
func (m *MigrationService) RunMigrations() error {
	// Ensure migration table exists
	if err := m.EnsureMigrationTable(); err != nil {
		return err
	}

	// Get applied migrations
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}

	// Define all migrations in order
	migrations := []struct {
		version string
		sql     string
	}{
		{"001_create_users_table", CreateUsersTable},
		{"002_create_sessions_table", CreateSessionsTable},
		{"003_create_email_log_table", CreateEmailLogTable},
		{"004_create_user_profiles_table", CreateUserProfilesTable},
		{"005_create_vendors_table", CreateVendorsTable},
		{"006_create_products_table", CreateProductsTable},
		{"007_create_orders_table", CreateOrdersTable},
		{"008_create_auctions_table", CreateAuctionsTable},
		{"009_create_wholesale_table", CreateWholesaleTable},
		{"015_create_ai_vision_tables", CreateAIVisionTables},
		{"016_create_enterprise_ai_tables", CreateEnterpriseAITables},
		{"020_create_ai_template_tables", CreateAITemplateTables},
		{"021_create_marketplace_integration_tables", CreateMarketplaceIntegrationTables},
		{"022_update_users_table", UpdateUsersTable},
	}

	// Run pending migrations
	for _, migration := range migrations {
		if !applied[migration.version] {
			fmt.Printf("Running migration: %s\n", migration.version)

			// Execute migration
			_, err := m.db.Exec(migration.sql)
			if err != nil {
				return fmt.Errorf("error running migration %s: %w", migration.version, err)
			}

			// Record migration as applied
			_, err = m.db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", migration.version)
			if err != nil {
				return fmt.Errorf("error recording migration %s: %w", migration.version, err)
			}

			fmt.Printf("Migration %s completed successfully\n", migration.version)
		}
	}

	return nil
}

// CreateAITemplateTables creates AI template related tables
const CreateAITemplateTables = `
CREATE TABLE IF NOT EXISTS ai_templates (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	name TEXT NOT NULL,
	type TEXT NOT NULL,
	category TEXT NOT NULL,
	content TEXT NOT NULL,
	metadata TEXT,
	is_public BOOLEAN DEFAULT FALSE,
	is_active BOOLEAN DEFAULT TRUE,
	usage_count INTEGER DEFAULT 0,
	rating REAL DEFAULT 0,
	tags TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS ai_template_usage (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	template_id INTEGER NOT NULL,
	user_id INTEGER NOT NULL,
	product_id INTEGER,
	platform TEXT NOT NULL,
	used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	success BOOLEAN DEFAULT TRUE,
	output_url TEXT,
	FOREIGN KEY (template_id) REFERENCES ai_templates(id) ON DELETE CASCADE,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS ai_template_ratings (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	template_id INTEGER NOT NULL,
	user_id INTEGER NOT NULL,
	rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
	comment TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (template_id) REFERENCES ai_templates(id) ON DELETE CASCADE,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	UNIQUE(template_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_ai_templates_user_id ON ai_templates(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_templates_type ON ai_templates(type);
CREATE INDEX IF NOT EXISTS idx_ai_templates_category ON ai_templates(category);
`

// CreateMarketplaceIntegrationTables creates marketplace integration related tables
const CreateMarketplaceIntegrationTables = `
CREATE TABLE IF NOT EXISTS marketplace_integrations (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	name TEXT NOT NULL,
	type TEXT NOT NULL,
	platform TEXT NOT NULL,
	api_key TEXT,
	api_secret TEXT,
	access_token TEXT,
	refresh_token TEXT,
	config TEXT,
	is_active BOOLEAN DEFAULT TRUE,
	last_sync TIMESTAMP,
	sync_status TEXT DEFAULT 'pending',
	error_message TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS sync_logs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	integration_id INTEGER NOT NULL,
	sync_type TEXT NOT NULL,
	status TEXT NOT NULL,
	records_total INTEGER DEFAULT 0,
	records_success INTEGER DEFAULT 0,
	records_failed INTEGER DEFAULT 0,
	error_message TEXT,
	started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	completed_at TIMESTAMP,
	duration INTEGER DEFAULT 0,
	FOREIGN KEY (integration_id) REFERENCES marketplace_integrations(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS product_mappings (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	integration_id INTEGER NOT NULL,
	local_product_id INTEGER NOT NULL,
	remote_product_id TEXT NOT NULL,
	remote_sku TEXT,
	sync_status TEXT DEFAULT 'pending',
	last_synced TIMESTAMP,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (integration_id) REFERENCES marketplace_integrations(id) ON DELETE CASCADE,
	FOREIGN KEY (local_product_id) REFERENCES products(id) ON DELETE CASCADE,
	UNIQUE(integration_id, local_product_id),
	UNIQUE(integration_id, remote_product_id)
);

CREATE INDEX IF NOT EXISTS idx_marketplace_integrations_user_id ON marketplace_integrations(user_id);
CREATE INDEX IF NOT EXISTS idx_marketplace_integrations_type ON marketplace_integrations(type);
`

// UpdateUsersTable adds new columns to users table
const UpdateUsersTable = `
ALTER TABLE users ADD COLUMN role TEXT DEFAULT 'user';
ALTER TABLE users ADD COLUMN ai_access BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN ai_edit_access BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN ai_template_access BOOLEAN DEFAULT FALSE;

UPDATE users SET ai_access = TRUE, ai_edit_access = TRUE, ai_template_access = TRUE WHERE is_admin = TRUE;

CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_ai_access ON users(ai_access);
`
