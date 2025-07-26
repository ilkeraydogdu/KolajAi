package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"kolajAi/internal/database/migrations"
)

// Migration represents a database migration
type Migration struct {
	ID        int
	Name      string
	SQL       string
	CreatedAt time.Time
}

// Migrator handles database migrations
type Migrator struct {
	db *sql.DB
}

// NewMigrator creates a new migrator
func NewMigrator(db *sql.DB) (*Migrator, error) {
	m := &Migrator{db: db}

	// Create migrations table if it doesn't exist
	err := m.createMigrationsTable()
	if err != nil {
		return nil, err
	}

	return m, nil
}

// createMigrationsTable creates the migrations table if it doesn't exist
func (m *Migrator) createMigrationsTable() error {
	// Bu işlem için doğrudan SQL kullanmak gerekiyor çünkü tablo oluşturma
	// işlemi için QueryBuilder'da uygun bir metod bulunmuyor
	query := `
	CREATE TABLE IF NOT EXISTS migrations (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL UNIQUE,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	stmt, err := m.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	return nil
}

// ApplyMigration applies a single migration
func (m *Migrator) ApplyMigration(migration Migration) error {
	// QueryBuilder kullanarak migration kontrolü yap
	qb := NewQueryBuilder("migrations")
	query, args := qb.Where("name", Equal, migration.Name).BuildCount()

	stmt, err := m.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	var count int64
	err = stmt.QueryRow(args...).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if count > 0 {
		// Migration already applied
		return nil
	}

	// Start transaction
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Migration SQL'i doğrudan uygulanmalı çünkü bu özel bir SQL ifadesi
	// ve QueryBuilder ile oluşturulamaz
	stmt, err = tx.Prepare(migration.SQL)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare migration statement: %w", err)
	}

	_, err = stmt.Exec()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to apply migration: %w", err)
	}
	stmt.Close()

	// QueryBuilder kullanarak migration kaydı ekle
	insertQb := NewQueryBuilder("migrations")
	insertQuery, insertArgs := insertQb.BuildInsert(map[string]interface{}{
		"name": migration.Name,
	})

	stmt, err = tx.Prepare(insertQuery)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}

	_, err = stmt.Exec(insertArgs...)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record migration: %w", err)
	}
	stmt.Close()

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	log.Printf("Applied migration: %s", migration.Name)
	return nil
}

// ApplyMigrations applies multiple migrations in order
func (m *Migrator) ApplyMigrations(migrations []Migration) error {
	for _, migration := range migrations {
		err := m.ApplyMigration(migration)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetAppliedMigrations returns all applied migrations
func (m *Migrator) GetAppliedMigrations() ([]string, error) {
	// QueryBuilder kullanarak uygulanmış migrasyonları al
	qb := NewQueryBuilder("migrations")
	query, args := qb.Select("name").OrderBy("id", Ascending).Build()

	stmt, err := m.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}
	defer rows.Close()

	var migrations []string
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			return nil, fmt.Errorf("failed to scan migration: %w", err)
		}
		migrations = append(migrations, name)
	}

	return migrations, nil
}

// ApplyCoreMigrations applies the core migrations that are built into the application
func (m *Migrator) ApplyCoreMigrations() error {
	coreMigrations := []Migration{
		{
			Name: "001_create_users_table",
			SQL:  migrations.CreateUsersTable,
		},
		{
			Name: "002_create_sessions_table",
			SQL:  migrations.CreateSessionsTable,
		},
		{
			Name: "003_create_email_log_table",
			SQL:  migrations.CreateEmailLogTable,
		},
		{
			Name: "004_create_user_profiles_table",
			SQL:  migrations.CreateUserProfilesTable,
		},
	}

	return m.ApplyMigrations(coreMigrations)
}
