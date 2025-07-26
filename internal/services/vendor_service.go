package services

import (
	"fmt"
	"kolajAi/internal/database"
	"kolajAi/internal/models"
	"time"
)

type VendorService struct {
	repo database.SimpleRepository
}

func NewVendorService(repo database.SimpleRepository) *VendorService {
	return &VendorService{repo: repo}
}

// CreateVendor creates a new vendor
func (s *VendorService) CreateVendor(vendor *models.Vendor) error {
	vendor.CreatedAt = time.Now()
	vendor.UpdatedAt = time.Now()
	vendor.Status = "pending"
	vendor.Rating = 0.0
	vendor.TotalSales = 0.0
	vendor.Commission = 5.0

	id, err := s.repo.CreateStruct("vendors", vendor)
	if err != nil {
		return fmt.Errorf("failed to create vendor: %w", err)
	}
	vendor.ID = int(id)
	return nil
}

// GetVendorByID retrieves a vendor by ID
func (s *VendorService) GetVendorByID(id int) (*models.Vendor, error) {
	var vendor models.Vendor
	err := s.repo.FindByID("vendors", id, &vendor)
	if err != nil {
		return nil, fmt.Errorf("failed to get vendor: %w", err)
	}
	return &vendor, nil
}

// GetVendorByUserID retrieves a vendor by user ID
func (s *VendorService) GetVendorByUserID(userID int) (*models.Vendor, error) {
	var vendor models.Vendor
	conditions := map[string]interface{}{"user_id": userID}
	err := s.repo.FindOne("vendors", &vendor, conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to get vendor by user ID: %w", err)
	}
	return &vendor, nil
}

// UpdateVendor updates a vendor
func (s *VendorService) UpdateVendor(id int, vendor *models.Vendor) error {
	vendor.UpdatedAt = time.Now()
	err := s.repo.Update("vendors", id, vendor)
	if err != nil {
		return fmt.Errorf("failed to update vendor: %w", err)
	}
	return nil
}

// ApproveVendor approves a vendor
func (s *VendorService) ApproveVendor(id int) error {
	vendor := &models.Vendor{
		Status:    "approved",
		UpdatedAt: time.Now(),
	}
	return s.UpdateVendor(id, vendor)
}

// RejectVendor rejects a vendor
func (s *VendorService) RejectVendor(id int) error {
	vendor := &models.Vendor{
		Status:    "rejected",
		UpdatedAt: time.Now(),
	}
	return s.UpdateVendor(id, vendor)
}

// SuspendVendor suspends a vendor
func (s *VendorService) SuspendVendor(id int) error {
	vendor := &models.Vendor{
		Status:    "suspended",
		UpdatedAt: time.Now(),
	}
	return s.UpdateVendor(id, vendor)
}

// GetAllVendors retrieves all vendors with pagination (overloaded for admin)
func (s *VendorService) GetAllVendors(limit, offset int) ([]models.Vendor, error) {
	var vendors []models.Vendor
	conditions := map[string]interface{}{}
	err := s.repo.FindAll("vendors", &vendors, conditions, "created_at DESC", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get all vendors: %w", err)
	}
	return vendors, nil
}

// GetPendingVendors returns vendors waiting for approval
func (s *VendorService) GetPendingVendors() ([]models.Vendor, error) {
	var vendors []models.Vendor
	conditions := map[string]interface{}{"status": "pending"}
	err := s.repo.FindAll("vendors", &vendors, conditions, "created_at ASC", 10, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending vendors: %w", err)
	}
	return vendors, nil
}

// GetVendorStats returns vendor statistics
func (s *VendorService) GetVendorStats(vendorID int) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Get total products
	productCount, err := s.repo.Count("products", map[string]interface{}{"vendor_id": vendorID})
	if err == nil {
		stats["total_products"] = productCount
	}

	// Get total orders
	orderCount, err := s.repo.Count("order_items", map[string]interface{}{"vendor_id": vendorID})
	if err == nil {
		stats["total_orders"] = orderCount
	}

	// Get vendor info
	vendor, err := s.GetVendorByID(vendorID)
	if err == nil {
		stats["total_sales"] = vendor.TotalSales
		stats["rating"] = vendor.Rating
	}

	return stats, nil
}

// UpdateVendorRating updates vendor rating
func (s *VendorService) UpdateVendorRating(vendorID int, rating float64) error {
	vendor := &models.Vendor{
		Rating:    rating,
		UpdatedAt: time.Now(),
	}
	return s.UpdateVendor(vendorID, vendor)
}

// AddVendorSale adds to vendor's total sales
func (s *VendorService) AddVendorSale(vendorID int, amount float64) error {
	vendor, err := s.GetVendorByID(vendorID)
	if err != nil {
		return err
	}

	vendor.TotalSales += amount
	vendor.UpdatedAt = time.Now()
	
	return s.UpdateVendor(vendorID, vendor)
}

// GetVendorCount returns the total number of vendors
func (s *VendorService) GetVendorCount() (int64, error) {
	conditions := map[string]interface{}{}
	return s.repo.Count("vendors", conditions)
}