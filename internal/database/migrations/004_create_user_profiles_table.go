package migrations

// CreateUserProfilesTable migration
var CreateUserProfilesTable = `
CREATE TABLE IF NOT EXISTS user_profiles (
	id INT AUTO_INCREMENT PRIMARY KEY,
	user_id INT NOT NULL UNIQUE,
	bio TEXT,
	avatar VARCHAR(255),
	company VARCHAR(100),
	website VARCHAR(255),
	location VARCHAR(100),
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`
