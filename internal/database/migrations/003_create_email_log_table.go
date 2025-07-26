package migrations

// CreateEmailLogTable migration
var CreateEmailLogTable = `
CREATE TABLE IF NOT EXISTS email_log (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER,
	email_to VARCHAR(100) NOT NULL,
	subject VARCHAR(255) NOT NULL,
	email_type VARCHAR(50) NOT NULL,
	status VARCHAR(20) NOT NULL,
	error_message TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);
`
