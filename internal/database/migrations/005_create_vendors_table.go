package migrations

const CreateVendorsTable = `
-- Create vendors table first (before products table)
CREATE TABLE IF NOT EXISTS vendors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    business_name VARCHAR(255) NOT NULL,
    business_type VARCHAR(50) DEFAULT 'individual',
    tax_number VARCHAR(50),
    tax_office VARCHAR(100),
    business_address TEXT,
    contact_person VARCHAR(255),
    phone VARCHAR(20),
    email VARCHAR(255),
    website VARCHAR(255),
    logo VARCHAR(500),
    description TEXT,
    bank_account_info TEXT,
    commission_rate DECIMAL(5,2) DEFAULT 10.00,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'suspended', 'rejected')),
    is_verified BOOLEAN DEFAULT FALSE,
    verification_documents TEXT,
    rating DECIMAL(3,2) DEFAULT 0.00,
    total_sales INTEGER DEFAULT 0,
    total_products INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_vendor_user ON vendors(user_id);
CREATE INDEX IF NOT EXISTS idx_vendor_status ON vendors(status);
CREATE INDEX IF NOT EXISTS idx_vendor_verified ON vendors(is_verified);
CREATE INDEX IF NOT EXISTS idx_vendor_rating ON vendors(rating);
CREATE INDEX IF NOT EXISTS idx_vendor_business_name ON vendors(business_name);

CREATE TABLE IF NOT EXISTS vendor_documents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    vendor_id INTEGER NOT NULL,
    type TEXT NOT NULL,
    file_name TEXT NOT NULL,
    file_path TEXT NOT NULL,
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (vendor_id) REFERENCES vendors(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_vendor_doc_vendor ON vendor_documents(vendor_id);
CREATE INDEX IF NOT EXISTS idx_vendor_doc_status ON vendor_documents(status);`
