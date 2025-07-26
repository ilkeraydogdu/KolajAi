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
