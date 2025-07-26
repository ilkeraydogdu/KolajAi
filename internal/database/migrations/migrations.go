package migrations

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Migration represents a database migration
type Migration struct {
	Version string
	SQL     string
}

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

// ApplyMigration applies a single migration
func (m *MigrationService) ApplyMigration(migration Migration) error {
	// Execute migration SQL
	_, err := m.db.Exec(migration.SQL)
	if err != nil {
		return fmt.Errorf("error executing migration SQL: %w", err)
	}

	// Record migration as applied
	_, err = m.db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", migration.Version)
	if err != nil {
		return fmt.Errorf("error recording migration: %w", err)
	}

	return nil
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
	fmt.Println("Starting migrations...")
	
	// Ensure migration table exists
	if err := m.EnsureMigrationTable(); err != nil {
		return fmt.Errorf("failed to ensure migration table: %w", err)
	}
	fmt.Println("Migration table ensured")

	// Get applied migrations
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}
	fmt.Printf("Found %d applied migrations\n", len(applied))

	// Define all migrations
	migrations := []Migration{
		{Version: "001_create_users_table", SQL: CreateUsersTable},
		{Version: "002_create_sessions_table", SQL: CreateSessionsTable},
		{Version: "003_create_email_log_table", SQL: CreateEmailLogTable},
		{Version: "004_create_user_profiles_table", SQL: CreateUserProfilesTable},
		{Version: "005_create_vendors_table", SQL: CreateVendorsTable},
		{Version: "006_create_products_table", SQL: CreateProductsTable},
		{Version: "007_create_orders_table", SQL: CreateOrdersTable},
		{Version: "008_create_auctions_table", SQL: CreateAuctionsTable},
		{Version: "009_create_wholesale_table", SQL: CreateWholesaleTable},
	}

	// Apply migrations
	for _, migration := range migrations {
		if !applied[migration.Version] {
			fmt.Printf("Applying migration: %s\n", migration.Version)
			if err := m.ApplyMigration(migration); err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
			}
			fmt.Printf("Successfully applied migration: %s\n", migration.Version)
		} else {
			fmt.Printf("Migration already applied: %s\n", migration.Version)
		}
	}

	fmt.Println("All migrations completed successfully")
	return nil
}
