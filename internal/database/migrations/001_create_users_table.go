package migrations

// CreateUsersTable migration
var CreateUsersTable = `
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE,
	password TEXT NOT NULL,
	phone TEXT,
	is_active INTEGER DEFAULT 0,
	is_admin INTEGER DEFAULT 0,
	verification_token TEXT,
	reset_token TEXT,
	token_expires_at DATETIME,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`
