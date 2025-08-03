package migrations

// CreateUsersTable migration
var CreateUsersTable = `
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE,
	password TEXT NOT NULL,
	phone TEXT,
	is_active BOOLEAN DEFAULT FALSE,
	is_admin BOOLEAN DEFAULT FALSE,
	verification_token TEXT,
	reset_token TEXT,
	token_expires_at DATETIME,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`
