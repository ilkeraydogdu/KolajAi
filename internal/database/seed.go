package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// Seed represents a database seed operation
type Seed struct {
	Name string
	Run  func(*sql.DB) error
}

// Seeder handles database seeding operations
type Seeder struct {
	db *sql.DB
}

// NewSeeder creates a new seeder
func NewSeeder(db *sql.DB) *Seeder {
	return &Seeder{db: db}
}

// ApplySeeds applies multiple seed operations
func (s *Seeder) ApplySeeds(seeds []Seed) error {
	// Create seeds table if it doesn't exist
	err := s.createSeedsTable()
	if err != nil {
		return fmt.Errorf("failed to create seeds table: %w", err)
	}

	for _, seed := range seeds {
		err := s.ApplySeed(seed)
		if err != nil {
			return err
		}
	}

	return nil
}

// createSeedsTable creates the seeds table if it doesn't exist
func (s *Seeder) createSeedsTable() error {
	// Bu işlem için doğrudan SQL kullanmak gerekiyor çünkü tablo oluşturma
	// işlemi için QueryBuilder'da uygun bir metod bulunmuyor
	query := `
	CREATE TABLE IF NOT EXISTS seeds (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL UNIQUE,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("failed to create seeds table: %w", err)
	}

	return nil
}

// ApplySeed applies a single seed operation
func (s *Seeder) ApplySeed(seed Seed) error {
	// QueryBuilder kullanarak seed kontrolü yap
	qb := NewQueryBuilder("seeds")
	query, args := qb.Where("name", Equal, seed.Name).BuildCount()

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	var count int64
	err = stmt.QueryRow(args...).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check seed status: %w", err)
	}

	if count > 0 {
		// Seed already applied
		return nil
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Apply seed
	err = seed.Run(s.db)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to apply seed: %w", err)
	}

	// QueryBuilder kullanarak seed kaydı ekle
	insertQb := NewQueryBuilder("seeds")
	insertQuery, insertArgs := insertQb.BuildInsert(map[string]interface{}{
		"name": seed.Name,
	})

	stmt, err = tx.Prepare(insertQuery)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(insertArgs...)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record seed: %w", err)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit seed: %w", err)
	}

	log.Printf("Applied seed: %s", seed.Name)
	return nil
}

// CreateDefaultAdminUser creates a default admin user if no users exist
func CreateDefaultAdminUser(db *sql.DB) error {
	// QueryBuilder kullanarak kullanıcı kontrolü yap
	log.Printf("Checking for existing users...")
	qb := NewQueryBuilder("users")
	query, args := qb.BuildCount()

	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	var count int64
	err = stmt.QueryRow(args...).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check users: %w", err)
	}

	if count > 0 {
		// Users already exist, skip
		log.Printf("Found %d existing users, skipping default admin creation", count)
		return nil
	}

	log.Printf("No users found in database, creating default admin user...")

	// Hash the default password
	// Get admin password from environment variable
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin123" // Fallback for development only
	}
	// Use stronger bcrypt cost for production security (12 instead of default 10)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), 12)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// QueryBuilder kullanarak admin kullanıcısı ekle
	insertQb := NewQueryBuilder("users")
	insertQuery, insertArgs := insertQb.BuildInsert(map[string]interface{}{
		"name":      "Admin User",
		"email":     "admin@kolajAi.com",
		"password":  string(hashedPassword),
		"phone":     "05555555555",
		"is_active": 1,
		"is_admin":  1,
	})

	stmt, err = db.Prepare(insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(insertArgs...)
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	id, _ := result.LastInsertId()
	log.Printf("Created default admin user with ID %d: admin@kolajAi.com / admin123", id)
	return nil
}
