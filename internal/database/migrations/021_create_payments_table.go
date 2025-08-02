package migrations

// CreatePaymentsTable SQL for creating payment related tables
const CreatePaymentsTable = `
CREATE TABLE IF NOT EXISTS payments (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	order_id INTEGER NOT NULL,
	user_id INTEGER NOT NULL,
	payment_method VARCHAR(50) NOT NULL CHECK (payment_method IN ('credit_card', 'debit_card', 'bank_transfer', 'wallet', 'cash_on_delivery')),
	payment_provider VARCHAR(50),
	transaction_id VARCHAR(255) UNIQUE,
	provider_transaction_id VARCHAR(255),
	amount DECIMAL(10,2) NOT NULL,
	currency VARCHAR(10) NOT NULL DEFAULT 'TRY',
	fee DECIMAL(10,2) DEFAULT 0.00,
	net_amount DECIMAL(10,2),
	status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled', 'refunded', 'partial_refund')),
	payment_date DATETIME,
	failure_reason TEXT,
	refund_amount DECIMAL(10,2) DEFAULT 0.00,
	refund_date DATETIME,
	refund_reason TEXT,
	notes TEXT,
	metadata TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS payment_details (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	payment_id INTEGER NOT NULL,
	card_type VARCHAR(20),
	card_last_four VARCHAR(4),
	card_holder_name VARCHAR(255),
	expiry_month VARCHAR(2),
	expiry_year VARCHAR(4),
	bank_name VARCHAR(255),
	bank_code VARCHAR(20),
	iban VARCHAR(34),
	wallet_type VARCHAR(50),
	wallet_id VARCHAR(255),
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments(user_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_transaction_id ON payments(transaction_id);
CREATE INDEX IF NOT EXISTS idx_payment_details_payment_id ON payment_details(payment_id);
`