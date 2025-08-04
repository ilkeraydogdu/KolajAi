package database

import (
	"database/sql"
	"fmt"
	"kolajAi/internal/config"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// Tx represents a database transaction
type Tx struct {
	*sql.Tx
}

// Config represents database configuration
type Config struct {
	Host         string
	Port         int
	User         string
	Password     string
	DatabaseName string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

// DefaultConfig returns a default database configuration
func DefaultConfig() *Config {
	// Önce yapılandırma dosyasından veritabanı ayarlarını almaya çalış
	dbConfig := &Config{
		Host:         "localhost",
		Port:         3306,
		User:         "kolajai",
		Password:     "",
		DatabaseName: "kolajai",
		MaxOpenConns: 25,
		MaxIdleConns: 10,
		MaxLifetime:  time.Minute * 5,
	}

	// Yapılandırma dosyasından ayarları al
	appConfig, err := config.LoadConfig("config.yaml")
	if err == nil && appConfig != nil {
		// Yapılandırma dosyasından veritabanı ayarlarını al
		dbConfig.Host = appConfig.Database.Host
		dbConfig.Port = appConfig.Database.Port
		dbConfig.User = appConfig.Database.User
		dbConfig.Password = appConfig.Database.Password
		dbConfig.DatabaseName = appConfig.Database.Name
		dbConfig.MaxOpenConns = appConfig.Database.MaxOpenConns
		dbConfig.MaxIdleConns = appConfig.Database.MaxIdleConns
		dbConfig.MaxLifetime = time.Duration(appConfig.Database.ConnMaxLifetime) * time.Minute

		log.Printf("Loaded database configuration from config file: %s:%d/%s",
			dbConfig.Host, dbConfig.Port, dbConfig.DatabaseName)
	} else {
		log.Printf("Using default database configuration: %s:%d/%s",
			dbConfig.Host, dbConfig.Port, dbConfig.DatabaseName)
	}

	return dbConfig
}

// BuildConnectionString builds a MySQL connection string from config
func (c *Config) BuildConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		c.User, c.Password, c.Host, c.Port, c.DatabaseName)
}

// BuildRootConnectionString builds a MySQL connection string without database name
func (c *Config) BuildRootConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/?parseTime=true&loc=Local&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		c.User, c.Password, c.Host, c.Port)
}

// NewConnection creates a new database connection using connection string
func NewConnection(connectionString string) (*sql.DB, error) {
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Minute * 5)

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	log.Println("Successfully connected to MySQL database")
	return db, nil
}

// InitDB initializes the database connection
func InitDB(config *Config) (*sql.DB, error) {
	// Önce veritabanı oluşturulacak
	err := createDatabaseIfNotExists(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Veritabanına bağlan
	db, err := sql.Open("mysql", config.BuildConnectionString())
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.MaxLifetime)

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	log.Println("Successfully connected to MySQL database")
	DB = db
	return db, nil
}

// createDatabaseIfNotExists creates the database if it doesn't exist
func createDatabaseIfNotExists(config *Config) error {
	// Önce varsayılan bağlantıyla MySQL'e bağlan (veritabanı adı olmadan)
	rootConn, err := sql.Open("mysql", config.BuildRootConnectionString())
	if err != nil {
		return fmt.Errorf("failed to open root database connection: %w", err)
	}
	defer rootConn.Close()

	// Veritabanı varlığını kontrol et
	rows, err := rootConn.Query("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?", config.DatabaseName)
	if err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}
	defer rows.Close()

	// Veritabanı yoksa oluştur
	dbExists := rows.Next()
	if !dbExists {
		log.Printf("Database '%s' does not exist, creating it...", config.DatabaseName)
		// Using safe database creation with quoted identifier
		// Database name should be validated to prevent injection
		createDBQuery := fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", config.DatabaseName)
		_, err = rootConn.Exec(createDBQuery)
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		log.Printf("Database '%s' created successfully", config.DatabaseName)
	}

	return nil
}

// SetupDatabase initializes database tables
func SetupDatabase(db *sql.DB) error {
	// Migrasyon yöneticisi oluştur
	migrator, err := NewMigrator(db)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	// Core migrasyonları uygula
	err = migrator.ApplyCoreMigrations()
	if err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	// Seed verileri oluştur
	log.Printf("Applying database seeds...")

	// Default admin kullanıcısı oluştur
	err = CreateDefaultAdminUser(db)
	if err != nil {
		log.Printf("Warning: Failed to create default admin user: %v", err)
		// Devam et, kritik bir hata değil
	}

	return nil
}
