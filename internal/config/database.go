package config

import (
	"fmt"
	"strings"
	"time"
)

// LegacyColumnConfig holds column configuration (for backward compatibility)
type LegacyColumnConfig struct {
	Name       string
	Type       string
	IsPrimary  bool
	IsAutoIncr bool
	IsUnique   bool
	NotNull    bool
	Default    interface{}
}

// LegacyRelationshipConfig holds relationship configuration (for backward compatibility)
type LegacyRelationshipConfig struct {
	Type         string
	TargetTable  string
	ForeignField string
	ThroughTable string
}

// LegacyIndexConfig holds index configuration (for backward compatibility)
type LegacyIndexConfig struct {
	Name    string
	Columns []string
	Type    string
}

// LegacyTableConfig holds table configuration (for backward compatibility)
type LegacyTableConfig struct {
	Name          string
	PrimaryKey    string
	Columns       map[string]*LegacyColumnConfig
	Relationships map[string]*LegacyRelationshipConfig
	Indexes       []*LegacyIndexConfig
}

// LegacyDatabaseConfig holds database configuration (for backward compatibility)
type LegacyDatabaseConfig struct {
	Host            string
	Port            int
	Username        string
	Password        string
	Database        string
	Charset         string
	ParseTime       bool
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	Tables          map[string]*LegacyTableConfig
}

// NewLegacyDatabaseConfig creates a new database configuration
func NewLegacyDatabaseConfig() *LegacyDatabaseConfig {
	return &LegacyDatabaseConfig{
		Host:            "localhost",
		Port:            3306,
		Username:        "root",
		Password:        "",
		Database:        "kolajai",
		Charset:         "utf8mb4",
		ParseTime:       true,
		MaxOpenConns:    25,
		MaxIdleConns:    25,
		ConnMaxLifetime: 5 * time.Minute,
		Tables:          make(map[string]*LegacyTableConfig),
	}
}

// GetDSN returns the database connection string
func (c *LegacyDatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%v",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.Charset,
		c.ParseTime,
	)
}

// AddTable adds a table configuration
func (c *LegacyDatabaseConfig) AddTable(name string, config LegacyTableConfig) {
	c.Tables[name] = &config
}

// GetTable returns a table configuration
func (c *LegacyDatabaseConfig) GetTable(name string) (LegacyTableConfig, bool) {
	config, exists := c.Tables[name]
	return *config, exists
}

// GetColumnList returns a list of column names for a table
func (c *LegacyDatabaseConfig) GetColumnList(tableName string) []string {
	if table, exists := c.Tables[tableName]; exists {
		columns := make([]string, 0, len(table.Columns))
		for colName := range table.Columns {
			columns = append(columns, colName)
		}
		return columns
	}
	return nil
}

// GetRelationConfig returns the relationship configuration for a table and relation
func (c *LegacyDatabaseConfig) GetRelationConfig(tableName, relationName string) (*LegacyRelationshipConfig, bool) {
	if table, exists := c.Tables[tableName]; exists {
		if relation, exists := table.Relationships[relationName]; exists {
			return relation, true
		}
	}
	return nil, false
}

// InitializeTables initializes table configurations
func (c *LegacyDatabaseConfig) InitializeTables() {
	// Users table
	c.Tables["users"] = &LegacyTableConfig{
		Name:       "users",
		PrimaryKey: "id",
		Columns: map[string]*LegacyColumnConfig{
			"id": {
				Name:       "id",
				Type:       "INT",
				IsPrimary:  true,
				IsAutoIncr: true,
			},
			"username": {
				Name:     "username",
				Type:     "VARCHAR(255)",
				IsUnique: true,
			},
			"email": {
				Name:     "email",
				Type:     "VARCHAR(255)",
				IsUnique: true,
			},
			"password": {
				Name: "password",
				Type: "VARCHAR(255)",
			},
			"active": {
				Name:    "active",
				Type:    "BOOLEAN",
				Default: true,
			},
			"role": {
				Name:    "role",
				Type:    "VARCHAR(50)",
				Default: "user",
			},
			"created_at": {
				Name:    "created_at",
				Type:    "TIMESTAMP",
				Default: "CURRENT_TIMESTAMP",
			},
			"updated_at": {
				Name:    "updated_at",
				Type:    "TIMESTAMP",
				Default: "CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP",
			},
		},
		Relationships: map[string]*LegacyRelationshipConfig{
			"posts": {
				Type:         "has_many",
				TargetTable:  "posts",
				ForeignField: "user_id",
			},
			"profile": {
				Type:         "has_one",
				TargetTable:  "profiles",
				ForeignField: "user_id",
			},
		},
		Indexes: []*LegacyIndexConfig{
			{
				Name:    "idx_users_username",
				Columns: []string{"username"},
				Type:    "UNIQUE",
			},
			{
				Name:    "idx_users_email",
				Columns: []string{"email"},
				Type:    "UNIQUE",
			},
		},
	}

	// Posts table
	c.Tables["posts"] = &LegacyTableConfig{
		Name:       "posts",
		PrimaryKey: "id",
		Columns: map[string]*LegacyColumnConfig{
			"id": {
				Name:       "id",
				Type:       "INT",
				IsPrimary:  true,
				IsAutoIncr: true,
			},
			"user_id": {
				Name: "user_id",
				Type: "INT",
			},
			"title": {
				Name: "title",
				Type: "VARCHAR(255)",
			},
			"content": {
				Name: "content",
				Type: "TEXT",
			},
			"status": {
				Name:    "status",
				Type:    "VARCHAR(50)",
				Default: "draft",
			},
			"created_at": {
				Name:    "created_at",
				Type:    "TIMESTAMP",
				Default: "CURRENT_TIMESTAMP",
			},
			"updated_at": {
				Name:    "updated_at",
				Type:    "TIMESTAMP",
				Default: "CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP",
			},
		},
		Relationships: map[string]*LegacyRelationshipConfig{
			"user": {
				Type:         "belongs_to",
				TargetTable:  "users",
				ForeignField: "user_id",
			},
			"comments": {
				Type:         "has_many",
				TargetTable:  "comments",
				ForeignField: "post_id",
			},
			"categories": {
				Type:         "belongs_to_many",
				TargetTable:  "categories",
				ThroughTable: "post_categories",
				ForeignField: "post_id",
			},
		},
		Indexes: []*LegacyIndexConfig{
			{
				Name:    "idx_posts_user_id",
				Columns: []string{"user_id"},
				Type:    "INDEX",
			},
			{
				Name:    "idx_posts_status",
				Columns: []string{"status"},
				Type:    "INDEX",
			},
		},
	}

	// Comments table
	c.Tables["comments"] = &LegacyTableConfig{
		Name:       "comments",
		PrimaryKey: "id",
		Columns: map[string]*LegacyColumnConfig{
			"id": {
				Name:       "id",
				Type:       "INT",
				IsPrimary:  true,
				IsAutoIncr: true,
			},
			"post_id": {
				Name: "post_id",
				Type: "INT",
			},
			"user_id": {
				Name: "user_id",
				Type: "INT",
			},
			"content": {
				Name: "content",
				Type: "TEXT",
			},
			"status": {
				Name:    "status",
				Type:    "VARCHAR(50)",
				Default: "pending",
			},
			"created_at": {
				Name:    "created_at",
				Type:    "TIMESTAMP",
				Default: "CURRENT_TIMESTAMP",
			},
			"updated_at": {
				Name:    "updated_at",
				Type:    "TIMESTAMP",
				Default: "CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP",
			},
		},
		Relationships: map[string]*LegacyRelationshipConfig{
			"post": {
				Type:         "belongs_to",
				TargetTable:  "posts",
				ForeignField: "post_id",
			},
			"user": {
				Type:         "belongs_to",
				TargetTable:  "users",
				ForeignField: "user_id",
			},
		},
		Indexes: []*LegacyIndexConfig{
			{
				Name:    "idx_comments_post_id",
				Columns: []string{"post_id"},
				Type:    "INDEX",
			},
			{
				Name:    "idx_comments_user_id",
				Columns: []string{"user_id"},
				Type:    "INDEX",
			},
		},
	}

	// Categories table
	c.Tables["categories"] = &LegacyTableConfig{
		Name:       "categories",
		PrimaryKey: "id",
		Columns: map[string]*LegacyColumnConfig{
			"id": {
				Name:       "id",
				Type:       "INT",
				IsPrimary:  true,
				IsAutoIncr: true,
			},
			"name": {
				Name: "name",
				Type: "VARCHAR(255)",
			},
			"slug": {
				Name: "slug",
				Type: "VARCHAR(255)",
			},
			"description": {
				Name: "description",
				Type: "TEXT",
			},
			"parent_id": {
				Name: "parent_id",
				Type: "INT",
			},
			"created_at": {
				Name:    "created_at",
				Type:    "TIMESTAMP",
				Default: "CURRENT_TIMESTAMP",
			},
			"updated_at": {
				Name:    "updated_at",
				Type:    "TIMESTAMP",
				Default: "CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP",
			},
		},
		Relationships: map[string]*LegacyRelationshipConfig{
			"parent": {
				Type:         "belongs_to",
				TargetTable:  "categories",
				ForeignField: "parent_id",
			},
			"children": {
				Type:         "has_many",
				TargetTable:  "categories",
				ForeignField: "parent_id",
			},
			"posts": {
				Type:         "belongs_to_many",
				TargetTable:  "posts",
				ThroughTable: "post_categories",
				ForeignField: "category_id",
			},
		},
		Indexes: []*LegacyIndexConfig{
			{
				Name:    "idx_categories_slug",
				Columns: []string{"slug"},
				Type:    "UNIQUE",
			},
			{
				Name:    "idx_categories_parent_id",
				Columns: []string{"parent_id"},
				Type:    "INDEX",
			},
		},
	}
}

// GetExtendedDatabaseConfig extends the base DatabaseConfig with default values
func GetExtendedDatabaseConfig() (DatabaseConfigExtended, bool) {
	baseConfig, ok := GetDatabaseConfig()
	if !ok {
		return DatabaseConfigExtended{}, false
	}

	// Extend with additional fields
	extended := DatabaseConfigExtended{
		Driver:          baseConfig.Driver,
		Host:            baseConfig.Host,
		Port:            baseConfig.Port,
		Username:        baseConfig.Username,
		Password:        baseConfig.Password,
		Database:        baseConfig.Database,
		MaxOpenConns:    baseConfig.MaxOpenConns,
		MaxIdleConns:    baseConfig.MaxIdleConns,
		ConnMaxLifetime: baseConfig.ConnMaxLifetime,
		Charset:         "utf8mb4",
		ParseTime:       true,
		Tables:          getDefaultTables(),
	}

	return extended, true
}

// BuildConnectionString builds a MySQL connection string
func (c *DatabaseConfigExtended) BuildConnectionString() string {
	// Base DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", c.Username, c.Password, c.Host, c.Port, c.Database)

	// Add query parameters
	params := []string{}
	if c.Charset != "" {
		params = append(params, "charset="+c.Charset)
	}
	if c.ParseTime {
		params = append(params, "parseTime=true")
	}

	// Add more params as needed
	params = append(params, "loc=Local")

	// Join parameters
	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}

	return dsn
}

// BuildRootConnectionString builds a connection string without database name
func (c *DatabaseConfigExtended) BuildRootConnectionString() string {
	// Base DSN without database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", c.Username, c.Password, c.Host, c.Port)

	// Add query parameters
	params := []string{}
	if c.Charset != "" {
		params = append(params, "charset="+c.Charset)
	}
	if c.ParseTime {
		params = append(params, "parseTime=true")
	}

	// Add more params as needed
	params = append(params, "loc=Local")

	// Join parameters
	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}

	return dsn
}

// getDefaultTables returns the default tables configuration for new schema
func getDefaultTables() []TableConfig {
	return []TableConfig{
		{
			Name: "users",
			Columns: []struct {
				Name       string `json:"name"`
				Type       string `json:"type"`
				PrimaryKey bool   `json:"primary_key,omitempty"`
				Nullable   bool   `json:"nullable,omitempty"`
				Default    string `json:"default,omitempty"`
			}{
				{Name: "id", Type: "int", PrimaryKey: true},
				{Name: "name", Type: "varchar(100)"},
				{Name: "email", Type: "varchar(100)"},
				{Name: "password", Type: "varchar(255)"},
				{Name: "is_admin", Type: "tinyint(1)", Default: "0"},
				{Name: "created_at", Type: "timestamp", Default: "CURRENT_TIMESTAMP"},
				{Name: "updated_at", Type: "timestamp", Nullable: true},
			},
			Indexes: []struct {
				Name    string   `json:"name"`
				Columns []string `json:"columns"`
				Unique  bool     `json:"unique,omitempty"`
			}{
				{Name: "idx_users_email", Columns: []string{"email"}, Unique: true},
			},
		},
		{
			Name: "sessions",
			Columns: []struct {
				Name       string `json:"name"`
				Type       string `json:"type"`
				PrimaryKey bool   `json:"primary_key,omitempty"`
				Nullable   bool   `json:"nullable,omitempty"`
				Default    string `json:"default,omitempty"`
			}{
				{Name: "id", Type: "varchar(255)", PrimaryKey: true},
				{Name: "user_id", Type: "int"},
				{Name: "data", Type: "text"},
				{Name: "created_at", Type: "timestamp", Default: "CURRENT_TIMESTAMP"},
				{Name: "expires_at", Type: "timestamp"},
			},
			Indexes: []struct {
				Name    string   `json:"name"`
				Columns []string `json:"columns"`
				Unique  bool     `json:"unique,omitempty"`
			}{
				{Name: "idx_sessions_user_id", Columns: []string{"user_id"}},
			},
		},
		{
			Name: "notifications",
			Columns: []struct {
				Name       string `json:"name"`
				Type       string `json:"type"`
				PrimaryKey bool   `json:"primary_key,omitempty"`
				Nullable   bool   `json:"nullable,omitempty"`
				Default    string `json:"default,omitempty"`
			}{
				{Name: "id", Type: "int", PrimaryKey: true},
				{Name: "user_id", Type: "int"},
				{Name: "type", Type: "varchar(50)"},
				{Name: "title", Type: "varchar(255)"},
				{Name: "message", Type: "text"},
				{Name: "is_read", Type: "tinyint(1)", Default: "0"},
				{Name: "created_at", Type: "timestamp", Default: "CURRENT_TIMESTAMP"},
			},
			Indexes: []struct {
				Name    string   `json:"name"`
				Columns []string `json:"columns"`
				Unique  bool     `json:"unique,omitempty"`
			}{
				{Name: "idx_notifications_user_id", Columns: []string{"user_id"}},
			},
		},
		{
			Name: "posts",
			Columns: []struct {
				Name       string `json:"name"`
				Type       string `json:"type"`
				PrimaryKey bool   `json:"primary_key,omitempty"`
				Nullable   bool   `json:"nullable,omitempty"`
				Default    string `json:"default,omitempty"`
			}{
				{Name: "id", Type: "int", PrimaryKey: true},
				{Name: "user_id", Type: "int"},
				{Name: "title", Type: "varchar(255)"},
				{Name: "content", Type: "text"},
				{Name: "status", Type: "varchar(50)", Default: "'draft'"},
				{Name: "created_at", Type: "timestamp", Default: "CURRENT_TIMESTAMP"},
				{Name: "updated_at", Type: "timestamp", Nullable: true},
			},
			Indexes: []struct {
				Name    string   `json:"name"`
				Columns []string `json:"columns"`
				Unique  bool     `json:"unique,omitempty"`
			}{
				{Name: "idx_posts_user_id", Columns: []string{"user_id"}},
			},
		},
	}
}

// DatabaseConfigExtended extends the base DatabaseConfig with additional fields
type DatabaseConfigExtended struct {
	Driver          string        `json:"driver"`
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	Username        string        `json:"username"`
	Password        string        `json:"password"`
	Database        string        `json:"database"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	Charset         string        `json:"charset"`
	ParseTime       bool          `json:"parse_time"`
	Tables          []TableConfig `json:"tables"`
}

// CreateTableSQL generates SQL to create a table
func (t *TableConfig) CreateTableSQL() string {
	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n", t.Name)

	// Add columns
	cols := []string{}
	for _, col := range t.Columns {
		colDef := fmt.Sprintf("  `%s` %s", col.Name, col.Type)

		if col.PrimaryKey {
			colDef += " PRIMARY KEY AUTO_INCREMENT"
		} else {
			if !col.Nullable {
				colDef += " NOT NULL"
			} else {
				colDef += " NULL"
			}

			if col.Default != "" {
				colDef += " DEFAULT " + col.Default
			}
		}

		cols = append(cols, colDef)
	}

	// Add indexes
	for _, idx := range t.Indexes {
		idxCols := []string{}
		for _, col := range idx.Columns {
			idxCols = append(idxCols, "`"+col+"`")
		}

		idxType := "INDEX"
		if idx.Unique {
			idxType = "UNIQUE INDEX"
		}

		cols = append(cols, fmt.Sprintf("  %s `%s` (%s)", idxType, idx.Name, strings.Join(idxCols, ", ")))
	}

	sql += strings.Join(cols, ",\n")
	sql += "\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;"

	return sql
}
