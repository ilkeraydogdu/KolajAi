package migrations

// CreateEmailLogTable migration
var CreateEmailLogTable = `
CREATE TABLE IF NOT EXISTS email_log (
	id INT AUTO_INCREMENT PRIMARY KEY,
	user_id INT,
	email_to VARCHAR(100) NOT NULL,
	subject VARCHAR(255) NOT NULL,
	email_type VARCHAR(50) NOT NULL,
	status VARCHAR(20) NOT NULL,
	error_message TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`
