package migrations

const CreateAuctionsTable = `
CREATE TABLE IF NOT EXISTS auctions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    vendor_id INTEGER NOT NULL,
    product_id INTEGER,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    starting_price DECIMAL(10,2) NOT NULL,
    reserve_price DECIMAL(10,2) DEFAULT 0.00,
    current_bid DECIMAL(10,2) DEFAULT 0.00,
    bid_increment DECIMAL(10,2) NOT NULL,
    buy_now_price DECIMAL(10,2) DEFAULT 0.00,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'ended', 'cancelled')),
    winner_id INTEGER,
    total_bids INTEGER DEFAULT 0,
    view_count INTEGER DEFAULT 0,
    is_reserve_met BOOLEAN DEFAULT FALSE,
    auto_extend BOOLEAN DEFAULT TRUE,
    extend_minutes INTEGER DEFAULT 5,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (vendor_id) REFERENCES vendors(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE SET NULL,
    FOREIGN KEY (winner_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_auction_vendor ON auctions(vendor_id);
CREATE INDEX IF NOT EXISTS idx_auction_status ON auctions(status);
CREATE INDEX IF NOT EXISTS idx_auction_end_time ON auctions(end_time);
CREATE INDEX IF NOT EXISTS idx_auction_winner ON auctions(winner_id);

CREATE TABLE IF NOT EXISTS auction_bids (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    auction_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    is_winning BOOLEAN DEFAULT FALSE,
    is_proxy BOOLEAN DEFAULT FALSE,
    max_amount DECIMAL(10,2) DEFAULT 0.00,
    ip_address VARCHAR(45),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (auction_id) REFERENCES auctions(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_auction_bid_auction ON auction_bids(auction_id);
CREATE INDEX IF NOT EXISTS idx_auction_bid_user ON auction_bids(user_id);
CREATE INDEX IF NOT EXISTS idx_auction_bid_amount ON auction_bids(amount);
CREATE INDEX IF NOT EXISTS idx_auction_bid_winning ON auction_bids(is_winning);

CREATE TABLE IF NOT EXISTS auction_watchers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    auction_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (auction_id) REFERENCES auctions(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (auction_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_auction_watcher_auction ON auction_watchers(auction_id);
CREATE INDEX IF NOT EXISTS idx_auction_watcher_user ON auction_watchers(user_id);

CREATE TABLE IF NOT EXISTS auction_images (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    auction_id INTEGER NOT NULL,
    image_url VARCHAR(500) NOT NULL,
    alt_text VARCHAR(255),
    sort_order INTEGER DEFAULT 0,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (auction_id) REFERENCES auctions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_auction_image_auction ON auction_images(auction_id);
CREATE INDEX IF NOT EXISTS idx_auction_image_primary ON auction_images(is_primary);

CREATE TABLE IF NOT EXISTS auction_questions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    auction_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    question TEXT NOT NULL,
    answer TEXT,
    is_public BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    answered_at TIMESTAMP NULL,
    FOREIGN KEY (auction_id) REFERENCES auctions(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_auction_question_auction ON auction_questions(auction_id);
CREATE INDEX IF NOT EXISTS idx_auction_question_user ON auction_questions(user_id);
CREATE INDEX IF NOT EXISTS idx_auction_question_public ON auction_questions(is_public);
`
