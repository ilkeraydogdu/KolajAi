package migrations

// CreateUsersTable migration
var CreateUsersTable = `
CREATE TABLE IF NOT EXISTS users (
	id INT AUTO_INCREMENT PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
	email VARCHAR(255) NOT NULL UNIQUE,
	password VARCHAR(255) NOT NULL,
	phone VARCHAR(20),
	is_active BOOLEAN DEFAULT FALSE,
	is_admin BOOLEAN DEFAULT FALSE,
	verification_token VARCHAR(255),
	reset_token VARCHAR(255),
	token_expires_at TIMESTAMP NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	INDEX idx_users_email (email),
	INDEX idx_users_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`
