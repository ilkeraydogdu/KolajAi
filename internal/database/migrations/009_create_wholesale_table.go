package migrations

const CreateWholesaleTable = `
CREATE TABLE IF NOT EXISTS wholesale_customers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    company_name VARCHAR(255) NOT NULL,
    tax_id VARCHAR(50),
    business_type VARCHAR(100),
    yearly_volume DECIMAL(15,2) DEFAULT 0.00,
    credit_limit DECIMAL(15,2) DEFAULT 0.00,
    payment_terms INTEGER DEFAULT 30,
    discount_tier VARCHAR(20) DEFAULT 'bronze' CHECK (discount_tier IN ('bronze', 'silver', 'gold', 'platinum')),
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'suspended', 'rejected')),
    approved_by INTEGER,
    approved_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_wholesale_customer_user ON wholesale_customers(user_id);
CREATE INDEX IF NOT EXISTS idx_wholesale_customer_status ON wholesale_customers(status);
CREATE INDEX IF NOT EXISTS idx_wholesale_customer_tier ON wholesale_customers(discount_tier);

CREATE TABLE IF NOT EXISTS wholesale_prices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER NOT NULL,
    min_qty INTEGER NOT NULL,
    max_qty INTEGER,
    price DECIMAL(10,2) NOT NULL,
    tier VARCHAR(20) DEFAULT 'all' CHECK (tier IN ('bronze', 'silver', 'gold', 'platinum', 'all')),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_wholesale_price_product ON wholesale_prices(product_id);
CREATE INDEX IF NOT EXISTS idx_wholesale_price_tier ON wholesale_prices(tier);
CREATE INDEX IF NOT EXISTS idx_wholesale_price_active ON wholesale_prices(is_active);

CREATE TABLE IF NOT EXISTS wholesale_orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    customer_id INTEGER NOT NULL,
    order_number VARCHAR(50) UNIQUE NOT NULL,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'pending', 'confirmed', 'processing', 'shipped', 'delivered', 'cancelled')),
    payment_status VARCHAR(20) DEFAULT 'pending' CHECK (payment_status IN ('pending', 'paid', 'partial', 'overdue')),
    payment_method VARCHAR(50),
    payment_terms INTEGER DEFAULT 30,
    sub_total DECIMAL(15,2) NOT NULL,
    discount_amount DECIMAL(15,2) DEFAULT 0.00,
    tax_amount DECIMAL(15,2) DEFAULT 0.00,
    shipping_cost DECIMAL(15,2) DEFAULT 0.00,
    total_amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'TRY',
    notes TEXT,
    due_date TIMESTAMP NULL,
    shipped_at TIMESTAMP NULL,
    delivered_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (customer_id) REFERENCES wholesale_customers(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_wholesale_order_customer ON wholesale_orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_wholesale_order_status ON wholesale_orders(status);
CREATE INDEX IF NOT EXISTS idx_wholesale_order_payment ON wholesale_orders(payment_status);
CREATE INDEX IF NOT EXISTS idx_wholesale_order_number ON wholesale_orders(order_number);

CREATE TABLE IF NOT EXISTS wholesale_order_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(15,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'shipped', 'delivered')),
    FOREIGN KEY (order_id) REFERENCES wholesale_orders(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_wholesale_order_item_order ON wholesale_order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_wholesale_order_item_product ON wholesale_order_items(product_id);

CREATE TABLE IF NOT EXISTS wholesale_quotes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    customer_id INTEGER NOT NULL,
    vendor_id INTEGER NOT NULL,
    quote_number VARCHAR(50) UNIQUE NOT NULL,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'sent', 'accepted', 'rejected', 'expired')),
    valid_until TIMESTAMP NOT NULL,
    total_amount DECIMAL(15,2) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (customer_id) REFERENCES wholesale_customers(id) ON DELETE CASCADE,
    FOREIGN KEY (vendor_id) REFERENCES vendors(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_wholesale_quote_customer ON wholesale_quotes(customer_id);
CREATE INDEX IF NOT EXISTS idx_wholesale_quote_vendor ON wholesale_quotes(vendor_id);
CREATE INDEX IF NOT EXISTS idx_wholesale_quote_status ON wholesale_quotes(status);
CREATE INDEX IF NOT EXISTS idx_wholesale_quote_number ON wholesale_quotes(quote_number);

CREATE TABLE IF NOT EXISTS wholesale_quote_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    quote_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(15,2) NOT NULL,
    FOREIGN KEY (quote_id) REFERENCES wholesale_quotes(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_wholesale_quote_item_quote ON wholesale_quote_items(quote_id);
CREATE INDEX IF NOT EXISTS idx_wholesale_quote_item_product ON wholesale_quote_items(product_id);
`