package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"kolajAi/internal/models"
	"log"
)

// MockDataService handles mock data operations
type MockDataService struct {
	categories []models.Category
	products   []models.Product
	auctions   []models.Auction
}

// NewMockDataService creates a new mock data service
func NewMockDataService() *MockDataService {
	service := &MockDataService{}
	service.loadMockData()
	return service
}

// loadMockData loads mock data from JSON files
func (s *MockDataService) loadMockData() {
	// Load categories
	if data, err := ioutil.ReadFile("mock_categories.json"); err == nil {
		if err := json.Unmarshal(data, &s.categories); err != nil {
			log.Printf("Error unmarshaling categories: %v", err)
		}
	} else {
		log.Printf("Categories mock file not found, using default data")
		s.categories = s.getDefaultCategories()
	}

	// Load products
	if data, err := ioutil.ReadFile("mock_products.json"); err == nil {
		if err := json.Unmarshal(data, &s.products); err != nil {
			log.Printf("Error unmarshaling products: %v", err)
		}
	} else {
		log.Printf("Products mock file not found, using default data")
		s.products = s.getDefaultProducts()
	}

	// Load auctions
	if data, err := ioutil.ReadFile("mock_auctions.json"); err == nil {
		if err := json.Unmarshal(data, &s.auctions); err != nil {
			log.Printf("Error unmarshaling auctions: %v", err)
		}
	} else {
		log.Printf("Auctions mock file not found, using default data")
		s.auctions = s.getDefaultAuctions()
	}

	log.Printf("Mock data loaded: %d categories, %d products, %d auctions", 
		len(s.categories), len(s.products), len(s.auctions))
}

// GetAllCategories returns all categories
func (s *MockDataService) GetAllCategories() ([]models.Category, error) {
	return s.categories, nil
}

// GetFeaturedProducts returns featured products
func (s *MockDataService) GetFeaturedProducts(limit int) ([]models.Product, error) {
	var featured []models.Product
	count := 0
	for _, product := range s.products {
		if product.IsFeatured && count < limit {
			featured = append(featured, product)
			count++
		}
	}
	return featured, nil
}

// GetActiveAuctions returns active auctions
func (s *MockDataService) GetActiveAuctions(limit int) ([]models.Auction, error) {
	var active []models.Auction
	count := 0
	for _, auction := range s.auctions {
		if auction.Status == "active" && count < limit {
			active = append(active, auction)
			count++
		}
	}
	return active, nil
}

// GetProducts returns products with filtering and pagination
func (s *MockDataService) GetProducts(category, search string, page, limit int) ([]models.Product, error) {
	var filtered []models.Product
	
	for _, product := range s.products {
		// Category filter
		if category != "" && fmt.Sprintf("%d", product.CategoryID) != category {
			continue
		}
		
		// Search filter
		if search != "" {
			if !contains(product.Name, search) && !contains(product.Description, search) {
				continue
			}
		}
		
		filtered = append(filtered, product)
	}
	
	// Pagination
	start := (page - 1) * limit
	end := start + limit
	
	if start >= len(filtered) {
		return []models.Product{}, nil
	}
	
	if end > len(filtered) {
		end = len(filtered)
	}
	
	return filtered[start:end], nil
}

// GetProductByID returns a product by ID
func (s *MockDataService) GetProductByID(id int) (*models.Product, error) {
	for _, product := range s.products {
		if product.ID == id {
			return &product, nil
		}
	}
	return nil, fmt.Errorf("product not found")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(str, substr string) bool {
	return len(str) >= len(substr) && 
		(str == substr || 
		 len(substr) == 0 || 
		 (len(str) > 0 && (str[0:len(substr)] == substr || 
		  (len(str) > len(substr) && contains(str[1:], substr)))))
}

// Default data fallbacks
func (s *MockDataService) getDefaultCategories() []models.Category {
	return []models.Category{
		{ID: 1, Name: "Elektronik", Slug: "elektronik", IsActive: true, SortOrder: 1},
		{ID: 2, Name: "Giyim", Slug: "giyim", IsActive: true, SortOrder: 2},
		{ID: 3, Name: "Ev & Yaşam", Slug: "ev-yasam", IsActive: true, SortOrder: 3},
	}
}

func (s *MockDataService) getDefaultProducts() []models.Product {
	return []models.Product{
		{ID: 1, CategoryID: 1, Name: "iPhone 15 Pro", Description: "Apple iPhone 15 Pro", Price: 45000.00, Stock: 50, Status: "active", IsFeatured: true, Rating: 4.8},
		{ID: 2, CategoryID: 1, Name: "Samsung Galaxy S24", Description: "Samsung Galaxy S24", Price: 42000.00, Stock: 30, Status: "active", IsFeatured: true, Rating: 4.7},
	}
}

func (s *MockDataService) getDefaultAuctions() []models.Auction {
	return []models.Auction{
		{ID: 1, Title: "Vintage Rolex", Description: "1970'lerden kalma orijinal Rolex", StartingPrice: 15000.00, CurrentBid: 18500.00, TotalBids: 12, Status: "active"},
	}
}