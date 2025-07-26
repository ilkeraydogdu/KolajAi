package migrations

// CreateUsersTable migration
var CreateUsersTable = `
CREATE TABLE IF NOT EXISTS users (
	id INT AUTO_INCREMENT PRIMARY KEY,
	name VARCHAR(100) NOT NULL,
	email VARCHAR(100) NOT NULL UNIQUE,
	password VARCHAR(255) NOT NULL,
	phone VARCHAR(20),
	is_active TINYINT(1) DEFAULT 0,
	is_admin TINYINT(1) DEFAULT 0,
	verification_token VARCHAR(100),
	reset_token VARCHAR(100),
	token_expires_at TIMESTAMP NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`
