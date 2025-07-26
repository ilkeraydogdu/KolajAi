package migrations

const CreateWholesaleTable = `
CREATE TABLE IF NOT EXISTS wholesale_customers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    company_name VARCHAR(255) NOT NULL,
    tax_id VARCHAR(50),
    business_type VARCHAR(100),
    yearly_volume DECIMAL(15,2) DEFAULT 0.00,
    credit_limit DECIMAL(15,2) DEFAULT 0.00,
    payment_terms INT DEFAULT 30,
    discount_tier ENUM('bronze', 'silver', 'gold', 'platinum') DEFAULT 'bronze',
    status ENUM('pending', 'approved', 'suspended', 'rejected') DEFAULT 'pending',
    approved_by INT,
    approved_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_wholesale_customer_user (user_id),
    INDEX idx_wholesale_customer_status (status),
    INDEX idx_wholesale_customer_tier (discount_tier)
);

CREATE TABLE IF NOT EXISTS wholesale_prices (
    id INT AUTO_INCREMENT PRIMARY KEY,
    product_id INT NOT NULL,
    min_qty INT NOT NULL,
    max_qty INT,
    price DECIMAL(10,2) NOT NULL,
    discount DECIMAL(5,2) DEFAULT 0.00,
    tier ENUM('bronze', 'silver', 'gold', 'platinum', 'all') DEFAULT 'all',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    INDEX idx_wholesale_price_product (product_id),
    INDEX idx_wholesale_price_tier (tier),
    INDEX idx_wholesale_price_active (is_active)
);

CREATE TABLE IF NOT EXISTS wholesale_orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    customer_id INT NOT NULL,
    order_number VARCHAR(50) UNIQUE NOT NULL,
    status ENUM('draft', 'pending', 'confirmed', 'processing', 'shipped', 'delivered', 'cancelled') DEFAULT 'draft',
    payment_status ENUM('pending', 'paid', 'partial', 'overdue') DEFAULT 'pending',
    payment_terms INT NOT NULL,
    due_date TIMESTAMP NOT NULL,
    sub_total DECIMAL(15,2) NOT NULL,
    discount_amount DECIMAL(15,2) DEFAULT 0.00,
    tax_amount DECIMAL(15,2) DEFAULT 0.00,
    shipping_cost DECIMAL(15,2) DEFAULT 0.00,
    total_amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'TRY',
    notes TEXT,
    internal_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (customer_id) REFERENCES wholesale_customers(id) ON DELETE CASCADE,
    INDEX idx_wholesale_order_customer (customer_id),
    INDEX idx_wholesale_order_status (status),
    INDEX idx_wholesale_order_payment (payment_status),
    INDEX idx_wholesale_order_number (order_number)
);

CREATE TABLE IF NOT EXISTS wholesale_order_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id INT NOT NULL,
    product_id INT NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    product_sku VARCHAR(100) NOT NULL,
    quantity INT NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    discount_rate DECIMAL(5,2) DEFAULT 0.00,
    total_price DECIMAL(15,2) NOT NULL,
    status ENUM('pending', 'confirmed', 'shipped', 'delivered') DEFAULT 'pending',
    FOREIGN KEY (order_id) REFERENCES wholesale_orders(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT,
    INDEX idx_wholesale_order_item_order (order_id),
    INDEX idx_wholesale_order_item_product (product_id)
);

CREATE TABLE IF NOT EXISTS wholesale_quotes (
    id INT AUTO_INCREMENT PRIMARY KEY,
    customer_id INT NOT NULL,
    vendor_id INT NOT NULL,
    quote_number VARCHAR(50) UNIQUE NOT NULL,
    status ENUM('draft', 'sent', 'accepted', 'rejected', 'expired') DEFAULT 'draft',
    valid_until TIMESTAMP NOT NULL,
    sub_total DECIMAL(15,2) NOT NULL,
    discount_amount DECIMAL(15,2) DEFAULT 0.00,
    total_amount DECIMAL(15,2) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (customer_id) REFERENCES wholesale_customers(id) ON DELETE CASCADE,
    FOREIGN KEY (vendor_id) REFERENCES vendors(id) ON DELETE CASCADE,
    INDEX idx_wholesale_quote_customer (customer_id),
    INDEX idx_wholesale_quote_vendor (vendor_id),
    INDEX idx_wholesale_quote_status (status),
    INDEX idx_wholesale_quote_number (quote_number)
);

CREATE TABLE IF NOT EXISTS wholesale_quote_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    quote_id INT NOT NULL,
    product_id INT NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    product_sku VARCHAR(100) NOT NULL,
    quantity INT NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    discount_rate DECIMAL(5,2) DEFAULT 0.00,
    total_price DECIMAL(15,2) NOT NULL,
    FOREIGN KEY (quote_id) REFERENCES wholesale_quotes(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT,
    INDEX idx_wholesale_quote_item_quote (quote_id),
    INDEX idx_wholesale_quote_item_product (product_id)
);`