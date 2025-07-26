package migrations

// CreateSessionsTable migration
var CreateSessionsTable = `
CREATE TABLE IF NOT EXISTS sessions (
	id VARCHAR(100) PRIMARY KEY,
	user_id INT NOT NULL,
	data TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	expires_at TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`
