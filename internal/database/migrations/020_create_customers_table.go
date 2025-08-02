package migrations

// CreateCustomersTable SQL for creating customers and addresses tables
const CreateCustomersTable = `
CREATE TABLE IF NOT EXISTS customers (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	first_name VARCHAR(100) NOT NULL,
	last_name VARCHAR(100) NOT NULL,
	email VARCHAR(255) NOT NULL,
	phone VARCHAR(20),
	date_of_birth DATE,
	gender VARCHAR(10) CHECK (gender IN ('male', 'female', 'other')),
	customer_type VARCHAR(20) DEFAULT 'individual' CHECK (customer_type IN ('individual', 'corporate')),
	company_name VARCHAR(255),
	tax_number VARCHAR(50),
	preferred_language VARCHAR(10) DEFAULT 'tr',
	newsletter BOOLEAN DEFAULT FALSE,
	sms_notifications BOOLEAN DEFAULT TRUE,
	email_notifications BOOLEAN DEFAULT TRUE,
	loyalty_points INTEGER DEFAULT 0,
	total_spent DECIMAL(10,2) DEFAULT 0.00,
	order_count INTEGER DEFAULT 0,
	last_order_date DATETIME,
	status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'blocked')),
	notes TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS addresses (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	customer_id INTEGER NOT NULL,
	type VARCHAR(20) DEFAULT 'both' CHECK (type IN ('billing', 'shipping', 'both')),
	title VARCHAR(50) DEFAULT 'Home',
	first_name VARCHAR(100) NOT NULL,
	last_name VARCHAR(100) NOT NULL,
	company VARCHAR(255),
	address_line1 VARCHAR(255) NOT NULL,
	address_line2 VARCHAR(255),
	city VARCHAR(100) NOT NULL,
	state VARCHAR(100),
	postal_code VARCHAR(20),
	country VARCHAR(100) NOT NULL DEFAULT 'Turkey',
	phone VARCHAR(20),
	is_default BOOLEAN DEFAULT FALSE,
	is_active BOOLEAN DEFAULT TRUE,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_customers_user_id ON customers(user_id);
CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email);
CREATE INDEX IF NOT EXISTS idx_customers_status ON customers(status);
CREATE INDEX IF NOT EXISTS idx_customers_customer_type ON customers(customer_type);
CREATE INDEX IF NOT EXISTS idx_addresses_customer_id ON addresses(customer_id);
CREATE INDEX IF NOT EXISTS idx_addresses_type ON addresses(type);
CREATE INDEX IF NOT EXISTS idx_addresses_is_default ON addresses(is_default);

CREATE TRIGGER IF NOT EXISTS update_customers_updated_at 
AFTER UPDATE ON customers
FOR EACH ROW
BEGIN
	UPDATE customers SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_addresses_updated_at 
AFTER UPDATE ON addresses
FOR EACH ROW
BEGIN
	UPDATE addresses SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
`