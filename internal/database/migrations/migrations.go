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
		{"017_create_ai_advanced_tables", AITablesMigration.Up},
		{"020_create_customers_table", CreateCustomersTable},
		{"021_create_payments_table", CreatePaymentsTable},
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
