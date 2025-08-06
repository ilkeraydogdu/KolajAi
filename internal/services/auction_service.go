package services

import (
	"fmt"
	"kolajAi/internal/database"
	"kolajAi/internal/models"
	"time"
)

type AuctionService struct {
	repo database.SimpleRepository
}

func NewAuctionService(repo database.SimpleRepository) *AuctionService {
	return &AuctionService{repo: repo}
}

// GetActiveAuctions retrieves active auctions
func (s *AuctionService) GetActiveAuctions(limit int) ([]models.Auction, error) {
	var auctions []models.Auction
	conditions := map[string]interface{}{
		"status": "active",
	}
	err := s.repo.FindAll("auctions", &auctions, conditions, "end_time ASC", limit, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get active auctions: %w", err)
	}
	
	// Load images for all auctions
	for i := range auctions {
		images, err := s.GetAuctionImages(auctions[i].ID)
		if err == nil && len(images) > 0 {
			auctions[i].Images = make([]string, len(images))
			for j, img := range images {
				auctions[i].Images[j] = img.ImageURL
				if img.IsPrimary {
					auctions[i].Image = img.ImageURL
				}
			}
			// If no primary image set, use first image as primary
			if auctions[i].Image == "" && len(auctions[i].Images) > 0 {
				auctions[i].Image = auctions[i].Images[0]
			}
		}
	}
	
	return auctions, nil
}

// CreateAuction creates a new auction
func (s *AuctionService) CreateAuction(auction *models.Auction) error {
	auction.CreatedAt = time.Now()
	auction.UpdatedAt = time.Now()
	if auction.Status == "" {
		auction.Status = "draft"
	}
	auction.CurrentBid = auction.StartingPrice
	auction.TotalBids = 0
	auction.ViewCount = 0
	auction.IsReserveMet = false

	id, err := s.repo.CreateStruct("auctions", auction)
	if err != nil {
		return fmt.Errorf("failed to create auction: %w", err)
	}
	auction.ID = int(id)
	return nil
}

// GetAuctionByID retrieves an auction by ID
func (s *AuctionService) GetAuctionByID(id int) (*models.Auction, error) {
	var auction models.Auction
	err := s.repo.FindByID("auctions", id, &auction)
	if err != nil {
		return nil, fmt.Errorf("failed to get auction: %w", err)
	}
	
	// Load auction images
	images, err := s.GetAuctionImages(id)
	if err == nil && len(images) > 0 {
		auction.Images = make([]string, len(images))
		for i, img := range images {
			auction.Images[i] = img.ImageURL
			if img.IsPrimary {
				auction.Image = img.ImageURL
			}
		}
		// If no primary image set, use first image as primary
		if auction.Image == "" && len(auction.Images) > 0 {
			auction.Image = auction.Images[0]
		}
	}
	
	return &auction, nil
}

// UpdateAuction updates an auction
func (s *AuctionService) UpdateAuction(id int, auction *models.Auction) error {
	auction.UpdatedAt = time.Now()
	err := s.repo.Update("auctions", id, auction)
	if err != nil {
		return fmt.Errorf("failed to update auction: %w", err)
	}
	return nil
}

// GetAuctionBids retrieves bids for an auction
func (s *AuctionService) GetAuctionBids(auctionID int, limit, offset int) ([]models.AuctionBid, error) {
	var bids []models.AuctionBid
	conditions := map[string]interface{}{"auction_id": auctionID}

	err := s.repo.FindAll("auction_bids", &bids, conditions, "amount DESC", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get auction bids: %w", err)
	}
	return bids, nil
}

// GetAuctionsByVendor retrieves auctions by vendor ID
func (s *AuctionService) GetAuctionsByVendor(vendorID int, limit, offset int) ([]models.Auction, error) {
	var auctions []models.Auction
	conditions := map[string]interface{}{"vendor_id": vendorID}

	err := s.repo.FindAll("auctions", &auctions, conditions, "created_at DESC", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get auctions by vendor: %w", err)
	}
	return auctions, nil
}

// GetEndingAuctions retrieves auctions ending soon
func (s *AuctionService) GetEndingAuctions(hours int, limit, offset int) ([]models.Auction, error) {
	var auctions []models.Auction
	endTime := time.Now().Add(time.Duration(hours) * time.Hour)

	// This would need a custom query in a real implementation
	conditions := map[string]interface{}{"status": "active"}

	err := s.repo.FindAll("auctions", &auctions, conditions, "end_time ASC", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get ending auctions: %w", err)
	}

	// Filter by end time (in a real implementation, this would be done in SQL)
	var filtered []models.Auction
	for _, auction := range auctions {
		if auction.EndTime.Before(endTime) {
			filtered = append(filtered, auction)
		}
	}

	return filtered, nil
}

// StartAuction starts an auction
func (s *AuctionService) StartAuction(auctionID int) error {
	auction := &models.Auction{
		Status:    "active",
		UpdatedAt: time.Now(),
	}
	return s.UpdateAuction(auctionID, auction)
}

// EndAuction ends an auction
func (s *AuctionService) EndAuction(auctionID int) error {
	auction, err := s.GetAuctionByID(auctionID)
	if err != nil {
		return err
	}

	// Get the winning bid
	winningBid, err := s.GetWinningBid(auctionID)
	if err == nil && winningBid != nil {
		auction.WinnerID = &winningBid.UserID
	}

	auction.Status = "ended"
	auction.UpdatedAt = time.Now()

	return s.UpdateAuction(auctionID, auction)
}

// CancelAuction cancels an auction
func (s *AuctionService) CancelAuction(auctionID int) error {
	auction := &models.Auction{
		Status:    "cancelled",
		UpdatedAt: time.Now(),
	}
	return s.UpdateAuction(auctionID, auction)
}

// PlaceBid places a bid on an auction
func (s *AuctionService) PlaceBid(bid *models.AuctionBid) error {
	// Get auction details
	auction, err := s.GetAuctionByID(bid.AuctionID)
	if err != nil {
		return err
	}

	// Validate bid
	if err := s.validateBid(auction, bid); err != nil {
		return err
	}

	// Mark previous bids as not winning
	err = s.markPreviousBidsAsLosing(bid.AuctionID)
	if err != nil {
		return err
	}

	// Create the bid
	bid.CreatedAt = time.Now()
	bid.IsWinning = true

	id, err := s.repo.CreateStruct("auction_bids", bid)
	if err != nil {
		return fmt.Errorf("failed to place bid: %w", err)
	}
	bid.ID = int(id)

	// Update auction
	auction.CurrentBid = bid.Amount
	auction.TotalBids++
	if auction.ReservePrice > 0 && bid.Amount >= auction.ReservePrice {
		auction.IsReserveMet = true
	}

	// Auto-extend if needed
	if auction.AutoExtend && time.Until(auction.EndTime) < time.Duration(auction.ExtendMinutes)*time.Minute {
		auction.EndTime = auction.EndTime.Add(time.Duration(auction.ExtendMinutes) * time.Minute)
	}

	return s.UpdateAuction(bid.AuctionID, auction)
}

// validateBid validates a bid
func (s *AuctionService) validateBid(auction *models.Auction, bid *models.AuctionBid) error {
	// Check if auction is active
	if auction.Status != "active" {
		return fmt.Errorf("auction is not active")
	}

	// Check if auction has ended
	if time.Now().After(auction.EndTime) {
		return fmt.Errorf("auction has ended")
	}

	// Check if bid meets minimum increment
	minBid := auction.CurrentBid + auction.BidIncrement
	if bid.Amount < minBid {
		return fmt.Errorf("bid must be at least %.2f", minBid)
	}

	return nil
}

// markPreviousBidsAsLosing marks all previous bids as not winning
func (s *AuctionService) markPreviousBidsAsLosing(auctionID int) error {
	// In a real implementation, this would be a single SQL UPDATE query
	bids, err := s.GetAuctionBids(auctionID, 0, 0)
	if err != nil {
		return err
	}

	for _, bid := range bids {
		if bid.IsWinning {
			bid.IsWinning = false
			s.repo.Update("auction_bids", bid.ID, &bid)
		}
	}

	return nil
}

// GetWinningBid retrieves the winning bid for an auction
func (s *AuctionService) GetWinningBid(auctionID int) (*models.AuctionBid, error) {
	var bid models.AuctionBid
	conditions := map[string]interface{}{
		"auction_id": auctionID,
		"is_winning": true,
	}
	err := s.repo.FindOne("auction_bids", &bid, conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to get winning bid: %w", err)
	}
	return &bid, nil
}

// GetUserBids retrieves bids by a user
func (s *AuctionService) GetUserBids(userID int, limit, offset int) ([]models.AuctionBid, error) {
	var bids []models.AuctionBid
	conditions := map[string]interface{}{"user_id": userID}

	err := s.repo.FindAll("auction_bids", &bids, conditions, "created_at DESC", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user bids: %w", err)
	}
	return bids, nil
}

// AddWatcher adds a user to auction watchers
func (s *AuctionService) AddWatcher(watcher *models.AuctionWatcher) error {
	watcher.CreatedAt = time.Now()

	id, err := s.repo.CreateStruct("auction_watchers", watcher)
	if err != nil {
		return fmt.Errorf("failed to add watcher: %w", err)
	}
	watcher.ID = int(id)
	return nil
}

// RemoveWatcher removes a user from auction watchers
func (s *AuctionService) RemoveWatcher(auctionID, userID int) error {
	conditions := map[string]interface{}{
		"auction_id": auctionID,
		"user_id":    userID,
	}

	var watcher models.AuctionWatcher
	err := s.repo.FindOne("auction_watchers", &watcher, conditions)
	if err != nil {
		return fmt.Errorf("watcher not found: %w", err)
	}

	return s.repo.Delete("auction_watchers", watcher.ID)
}

// GetAuctionWatchers retrieves watchers for an auction
func (s *AuctionService) GetAuctionWatchers(auctionID int) ([]models.AuctionWatcher, error) {
	var watchers []models.AuctionWatcher
	conditions := map[string]interface{}{"auction_id": auctionID}

	err := s.repo.FindAll("auction_watchers", &watchers, conditions, "created_at DESC", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get auction watchers: %w", err)
	}
	return watchers, nil
}

// IsUserWatching checks if a user is watching an auction
func (s *AuctionService) IsUserWatching(auctionID, userID int) (bool, error) {
	conditions := map[string]interface{}{
		"auction_id": auctionID,
		"user_id":    userID,
	}

	exists, err := s.repo.Exists("auction_watchers", conditions)
	if err != nil {
		return false, fmt.Errorf("failed to check if user is watching: %w", err)
	}
	return exists, nil
}

// AddAuctionImage adds an image to an auction
func (s *AuctionService) AddAuctionImage(image *models.AuctionImage) error {
	image.CreatedAt = time.Now()

	id, err := s.repo.CreateStruct("auction_images", image)
	if err != nil {
		return fmt.Errorf("failed to add auction image: %w", err)
	}
	image.ID = int(id)
	return nil
}

// GetAuctionImages retrieves images for an auction
func (s *AuctionService) GetAuctionImages(auctionID int) ([]models.AuctionImage, error) {
	var images []models.AuctionImage
	conditions := map[string]interface{}{"auction_id": auctionID}

	err := s.repo.FindAll("auction_images", &images, conditions, "sort_order ASC", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get auction images: %w", err)
	}
	return images, nil
}

// IncrementAuctionViews increments auction view count
func (s *AuctionService) IncrementAuctionViews(auctionID int) error {
	auction, err := s.GetAuctionByID(auctionID)
	if err != nil {
		return err
	}

	auction.ViewCount++
	auction.UpdatedAt = time.Now()

	return s.UpdateAuction(auctionID, auction)
}

// GetAuctionStats returns auction statistics
func (s *AuctionService) GetAuctionStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total auctions
	totalAuctions, err := s.repo.Count("auctions", nil)
	if err == nil {
		stats["total_auctions"] = totalAuctions
	}

	// Active auctions
	activeAuctions, err := s.repo.Count("auctions", map[string]interface{}{"status": "active"})
	if err == nil {
		stats["active_auctions"] = activeAuctions
	}

	// Ended auctions
	endedAuctions, err := s.repo.Count("auctions", map[string]interface{}{"status": "ended"})
	if err == nil {
		stats["ended_auctions"] = endedAuctions
	}

	// Total bids
	totalBids, err := s.repo.Count("auction_bids", nil)
	if err == nil {
		stats["total_bids"] = totalBids
	}

	return stats, nil
}

// ProcessExpiredAuctions processes auctions that have expired
func (s *AuctionService) ProcessExpiredAuctions() error {
	// Get active auctions that have ended
	var auctions []models.Auction
	conditions := map[string]interface{}{"status": "active"}

	err := s.repo.FindAll("auctions", &auctions, conditions, "end_time ASC", 0, 0)
	if err != nil {
		return fmt.Errorf("failed to get active auctions: %w", err)
	}

	now := time.Now()
	for _, auction := range auctions {
		if auction.EndTime.Before(now) {
			err := s.EndAuction(auction.ID)
			if err != nil {
				fmt.Printf("Error ending auction %d: %v\n", auction.ID, err)
			}
		}
	}

	return nil
}

// GetActiveAuctionCount returns the number of active auctions
func (s *AuctionService) GetActiveAuctionCount() (int64, error) {
	conditions := map[string]interface{}{"status": "active"}
	return s.repo.Count("auctions", conditions)
}
