package migrations

// CreateSessionsTable migration
var CreateSessionsTable = `
CREATE TABLE IF NOT EXISTS sessions (
	id TEXT PRIMARY KEY,
	user_id INTEGER,
	user_agent TEXT,
	ip_address TEXT,
	login_time DATETIME,
	last_activity DATETIME,
	expires_at DATETIME,
	is_active INTEGER DEFAULT 1,
	device_info TEXT,
	permissions TEXT,
	preferences TEXT,
	data TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`
