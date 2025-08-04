package migrations

// CreateSessionsTable migration
var CreateSessionsTable = `
CREATE TABLE IF NOT EXISTS sessions (
	id VARCHAR(128) PRIMARY KEY,
	user_id INT,
	data TEXT,
	expires_at TIMESTAMP NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	INDEX idx_sessions_user (user_id),
	INDEX idx_sessions_expires (expires_at),
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`
