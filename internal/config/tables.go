package config

// DatabaseTables defines all table names used in the application
// This ensures consistency in table naming across the application
type DatabaseTables struct {
	Users              string
	UserRoles          string
	UserPermissions    string
	Roles              string
	Permissions        string
	RolePermissions    string
	Sessions           string
	PasswordResets     string
	EmailVerifications string
	LoginAttempts      string
	AuditLogs          string
	UserSettings       string
	UserProfiles       string
	SentEmails         string
	Notifications      string
	Projects           string
	ProjectMembers     string
	ProjectAssets      string
	Documents          string
	DocumentVersions   string
	Templates          string
	TemplateCategories string
	Media              string
	Tags               string
	Comments           string
	Products           string
	Orders             string
	OrderItems         string
	Payments           string
	Subscriptions      string
	Invoices           string
	Settings           string
}

// NewDatabaseTables returns a new DatabaseTables with default table names
func NewDatabaseTables() DatabaseTables {
	return DatabaseTables{
		Users:              "users",
		UserRoles:          "user_roles",
		UserPermissions:    "user_permissions",
		Roles:              "roles",
		Permissions:        "permissions",
		RolePermissions:    "role_permissions",
		Sessions:           "sessions",
		PasswordResets:     "password_resets",
		EmailVerifications: "email_verifications",
		LoginAttempts:      "login_attempts",
		AuditLogs:          "audit_logs",
		UserSettings:       "user_settings",
		UserProfiles:       "user_profiles",
		SentEmails:         "sent_emails",
		Notifications:      "notifications",
		Projects:           "projects",
		ProjectMembers:     "project_members",
		ProjectAssets:      "project_assets",
		Documents:          "documents",
		DocumentVersions:   "document_versions",
		Templates:          "templates",
		TemplateCategories: "template_categories",
		Media:              "media",
		Tags:               "tags",
		Comments:           "comments",
		Products:           "products",
		Orders:             "orders",
		OrderItems:         "order_items",
		Payments:           "payments",
		Subscriptions:      "subscriptions",
		Invoices:           "invoices",
		Settings:           "settings",
	}
}
