package models

import "time"

// Auction represents an auction
type Auction struct {
	ID            int       `json:"id" db:"id"`
	VendorID      int       `json:"vendor_id" db:"vendor_id"`
	ProductID     int       `json:"product_id" db:"product_id"`
	Title         string    `json:"title" db:"title"`
	Description   string    `json:"description" db:"description"`
	StartingPrice float64   `json:"starting_price" db:"starting_price"`
	ReservePrice  float64   `json:"reserve_price" db:"reserve_price"`
	CurrentBid    float64   `json:"current_bid" db:"current_bid"`
	BidIncrement  float64   `json:"bid_increment" db:"bid_increment"`
	BuyNowPrice   float64   `json:"buy_now_price" db:"buy_now_price"`
	StartTime     time.Time `json:"start_time" db:"start_time"`
	EndTime       time.Time `json:"end_time" db:"end_time"`
	Status        string    `json:"status" db:"status"` // draft, active, ended, cancelled
	WinnerID      *int      `json:"winner_id" db:"winner_id"`
	TotalBids     int       `json:"total_bids" db:"total_bids"`
	ViewCount     int       `json:"view_count" db:"view_count"`
	IsReserveMet  bool      `json:"is_reserve_met" db:"is_reserve_met"`
	AutoExtend    bool      `json:"auto_extend" db:"auto_extend"`
	ExtendMinutes int       `json:"extend_minutes" db:"extend_minutes"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	
	// Computed fields for templates (not stored in DB)
	Images        []string `json:"images,omitempty" db:"-"`
	Image         string   `json:"image,omitempty" db:"-"` // Primary image
}

// AuctionBid represents a bid in an auction
type AuctionBid struct {
	ID        int       `json:"id" db:"id"`
	AuctionID int       `json:"auction_id" db:"auction_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Amount    float64   `json:"amount" db:"amount"`
	IsWinning bool      `json:"is_winning" db:"is_winning"`
	IsProxy   bool      `json:"is_proxy" db:"is_proxy"`
	MaxAmount float64   `json:"max_amount" db:"max_amount"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// AuctionWatcher represents users watching an auction
type AuctionWatcher struct {
	ID        int       `json:"id" db:"id"`
	AuctionID int       `json:"auction_id" db:"auction_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// AuctionImage represents images for auction items
type AuctionImage struct {
	ID        int       `json:"id" db:"id"`
	AuctionID int       `json:"auction_id" db:"auction_id"`
	ImageURL  string    `json:"image_url" db:"image_url"`
	AltText   string    `json:"alt_text" db:"alt_text"`
	SortOrder int       `json:"sort_order" db:"sort_order"`
	IsPrimary bool      `json:"is_primary" db:"is_primary"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// AuctionQuestion represents questions asked about auction items
type AuctionQuestion struct {
	ID         int        `json:"id" db:"id"`
	AuctionID  int        `json:"auction_id" db:"auction_id"`
	UserID     int        `json:"user_id" db:"user_id"`
	Question   string     `json:"question" db:"question"`
	Answer     string     `json:"answer" db:"answer"`
	IsPublic   bool       `json:"is_public" db:"is_public"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	AnsweredAt *time.Time `json:"answered_at" db:"answered_at"`
}
