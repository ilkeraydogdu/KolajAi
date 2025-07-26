package migrations

const CreateAuctionsTable = `
CREATE TABLE IF NOT EXISTS auctions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    vendor_id INT NOT NULL,
    product_id INT,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    starting_price DECIMAL(10,2) NOT NULL,
    reserve_price DECIMAL(10,2) DEFAULT 0.00,
    current_bid DECIMAL(10,2) DEFAULT 0.00,
    bid_increment DECIMAL(10,2) NOT NULL,
    buy_now_price DECIMAL(10,2) DEFAULT 0.00,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    status ENUM('draft', 'active', 'ended', 'cancelled') DEFAULT 'draft',
    winner_id INT,
    total_bids INT DEFAULT 0,
    view_count INT DEFAULT 0,
    is_reserve_met BOOLEAN DEFAULT FALSE,
    auto_extend BOOLEAN DEFAULT TRUE,
    extend_minutes INT DEFAULT 5,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (vendor_id) REFERENCES vendors(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE SET NULL,
    FOREIGN KEY (winner_id) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_auction_vendor (vendor_id),
    INDEX idx_auction_status (status),
    INDEX idx_auction_end_time (end_time),
    INDEX idx_auction_winner (winner_id)
);

CREATE TABLE IF NOT EXISTS auction_bids (
    id INT AUTO_INCREMENT PRIMARY KEY,
    auction_id INT NOT NULL,
    user_id INT NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    is_winning BOOLEAN DEFAULT FALSE,
    is_proxy BOOLEAN DEFAULT FALSE,
    max_amount DECIMAL(10,2) DEFAULT 0.00,
    ip_address VARCHAR(45),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (auction_id) REFERENCES auctions(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_auction_bid_auction (auction_id),
    INDEX idx_auction_bid_user (user_id),
    INDEX idx_auction_bid_amount (amount),
    INDEX idx_auction_bid_winning (is_winning)
);

CREATE TABLE IF NOT EXISTS auction_watchers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    auction_id INT NOT NULL,
    user_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (auction_id) REFERENCES auctions(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_watcher (auction_id, user_id),
    INDEX idx_auction_watcher_auction (auction_id),
    INDEX idx_auction_watcher_user (user_id)
);

CREATE TABLE IF NOT EXISTS auction_images (
    id INT AUTO_INCREMENT PRIMARY KEY,
    auction_id INT NOT NULL,
    image_url VARCHAR(500) NOT NULL,
    alt_text VARCHAR(255),
    sort_order INT DEFAULT 0,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (auction_id) REFERENCES auctions(id) ON DELETE CASCADE,
    INDEX idx_auction_image_auction (auction_id),
    INDEX idx_auction_image_primary (is_primary)
);

CREATE TABLE IF NOT EXISTS auction_questions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    auction_id INT NOT NULL,
    user_id INT NOT NULL,
    question TEXT NOT NULL,
    answer TEXT,
    is_public BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    answered_at TIMESTAMP NULL,
    FOREIGN KEY (auction_id) REFERENCES auctions(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_auction_question_auction (auction_id),
    INDEX idx_auction_question_user (user_id),
    INDEX idx_auction_question_public (is_public)
);`