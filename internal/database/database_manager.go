package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	MySQL  DatabaseType = "mysql"
	SQLite DatabaseType = "sqlite3"
)

// DatabaseManager manages database connections
type DatabaseManager struct {
	DB       *sql.DB
	DBType   DatabaseType
	ConnStr  string
	IsActive bool
}

// NewDatabaseManager creates a new database manager
func NewDatabaseManager() *DatabaseManager {
	return &DatabaseManager{}
}

// InitializeDatabase initializes the database based on environment
func (dm *DatabaseManager) InitializeDatabase() error {
	// Check if we're in development mode (use SQLite) or production (use MySQL)
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("GIN_MODE")
	}

	// Development: Use SQLite
	if env == "development" || env == "" {
		return dm.initSQLite()
	}

	// Production: Try MySQL first, fallback to SQLite
	if err := dm.initMySQL(); err != nil {
		log.Printf("MySQL connection failed, falling back to SQLite: %v", err)
		return dm.initSQLite()
	}

	return nil
}

// initSQLite initializes SQLite connection
func (dm *DatabaseManager) initSQLite() error {
	// Ensure data directory exists
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "kolajai.db")
	connStr := fmt.Sprintf("file:%s?cache=shared&mode=rwc&_journal_mode=WAL&_foreign_keys=on", dbPath)

	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	// Configure SQLite
	db.SetMaxOpenConns(1) // SQLite doesn't support multiple writers
	db.SetMaxIdleConns(1)

	dm.DB = db
	dm.DBType = SQLite
	dm.ConnStr = connStr
	dm.IsActive = true

	log.Printf("✅ SQLite database initialized: %s", dbPath)
	return nil
}

// initMySQL initializes MySQL connection
func (dm *DatabaseManager) initMySQL() error {
	// Get MySQL configuration from environment
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "3306"
	}
	if user == "" {
		user = "kolajai"
	}
	if dbname == "" {
		dbname = "kolajai"
	}

	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
		user, password, host, port, dbname)

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return fmt.Errorf("failed to open MySQL database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping MySQL database: %w", err)
	}

	// Configure MySQL
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)

	dm.DB = db
	dm.DBType = MySQL
	dm.ConnStr = connStr
	dm.IsActive = true

	log.Printf("✅ MySQL database initialized: %s@%s:%s/%s", user, host, port, dbname)
	return nil
}

// Close closes the database connection
func (dm *DatabaseManager) Close() error {
	if dm.DB != nil {
		dm.IsActive = false
		return dm.DB.Close()
	}
	return nil
}

// GetDB returns the database connection
func (dm *DatabaseManager) GetDB() *sql.DB {
	return dm.DB
}

// GetType returns the database type
func (dm *DatabaseManager) GetType() DatabaseType {
	return dm.DBType
}

// IsMySQL returns true if using MySQL
func (dm *DatabaseManager) IsMySQL() bool {
	return dm.DBType == MySQL
}

// IsSQLite returns true if using SQLite
func (dm *DatabaseManager) IsSQLite() bool {
	return dm.DBType == SQLite
}

// GetConnectionString returns the connection string
func (dm *DatabaseManager) GetConnectionString() string {
	return dm.ConnStr
}

// Global database manager instance
var GlobalDBManager *DatabaseManager

// InitGlobalDB initializes the global database manager
func InitGlobalDB() error {
	GlobalDBManager = NewDatabaseManager()
	return GlobalDBManager.InitializeDatabase()
}

// GetGlobalDB returns the global database connection
func GetGlobalDB() *sql.DB {
	if GlobalDBManager == nil {
		log.Fatal("Database not initialized. Call InitGlobalDB() first.")
	}
	return GlobalDBManager.GetDB()
}