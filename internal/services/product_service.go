package services

import (
	"fmt"
	"kolajAi/internal/database"
	"kolajAi/internal/models"
	"strings"
	"time"
)

type ProductService struct {
	repo database.SimpleRepository
}

func NewProductService(repo database.SimpleRepository) *ProductService {
	return &ProductService{repo: repo}
}

// CreateProduct creates a new product
func (s *ProductService) CreateProduct(product *models.Product) error {
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	if product.Status == "" {
		product.Status = "draft"
	}

	id, err := s.repo.CreateStruct("products", product)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}
	product.ID = int(id)
	return nil
}

// GetProductByID retrieves a product by ID
func (s *ProductService) GetProductByID(id int) (*models.Product, error) {
	var product models.Product
	err := s.repo.FindByID("products", id, &product)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return &product, nil
}

// UpdateProduct updates a product
func (s *ProductService) UpdateProduct(id int, product *models.Product) error {
	product.UpdatedAt = time.Now()
	err := s.repo.Update("products", id, product)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	return nil
}

// DeleteProduct soft deletes a product
func (s *ProductService) DeleteProduct(id int) error {
	err := s.repo.SoftDelete("products", id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	return nil
}

// GetProductsByVendor retrieves products by vendor ID
func (s *ProductService) GetProductsByVendor(vendorID int, limit, offset int) ([]models.Product, error) {
	var products []models.Product
	conditions := map[string]interface{}{"vendor_id": vendorID}

	err := s.repo.FindAll("products", &products, conditions, "created_at DESC", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get products by vendor: %w", err)
	}
	return products, nil
}

// GetProductsByCategory retrieves products by category ID
func (s *ProductService) GetProductsByCategory(categoryID int, limit, offset int) ([]models.Product, error) {
	var products []models.Product
	conditions := map[string]interface{}{
		"category_id": categoryID,
		"status":      "active",
	}

	err := s.repo.FindAll("products", &products, conditions, "created_at DESC", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get products by category: %w", err)
	}
	return products, nil
}

// SearchProducts searches for products
func (s *ProductService) SearchProducts(term string, limit, offset int) ([]models.Product, error) {
	var products []models.Product
	fields := []string{"name", "description", "tags"}

	err := s.repo.Search("products", fields, term, limit, offset, &products)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}
	return products, nil
}

// GetFeaturedProducts retrieves featured products
func (s *ProductService) GetFeaturedProducts(limit, offset int) ([]models.Product, error) {
	var products []models.Product
	conditions := map[string]interface{}{
		"is_featured": true,
		"status":      "active",
	}

	err := s.repo.FindAll("products", &products, conditions, "created_at DESC", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get featured products: %w", err)
	}
	return products, nil
}

// UpdateProductStock updates product stock
func (s *ProductService) UpdateProductStock(productID int, quantity int) error {
	product, err := s.GetProductByID(productID)
	if err != nil {
		return err
	}

	product.Stock = quantity
	product.UpdatedAt = time.Now()

	// Update status based on stock
	if quantity <= 0 {
		product.Status = "out_of_stock"
	} else if product.Status == "out_of_stock" {
		product.Status = "active"
	}

	return s.UpdateProduct(productID, product)
}

// IncrementProductViews increments product view count
func (s *ProductService) IncrementProductViews(productID int) error {
	product, err := s.GetProductByID(productID)
	if err != nil {
		return err
	}

	product.ViewCount++
	product.UpdatedAt = time.Now()

	return s.UpdateProduct(productID, product)
}

// IncrementProductSales increments product sales count
func (s *ProductService) IncrementProductSales(productID int, quantity int) error {
	product, err := s.GetProductByID(productID)
	if err != nil {
		return err
	}

	product.SalesCount += quantity
	product.Stock -= quantity
	product.UpdatedAt = time.Now()

	// Update status if out of stock
	if product.Stock <= 0 {
		product.Status = "out_of_stock"
	}

	return s.UpdateProduct(productID, product)
}

// CreateCategory creates a new category
func (s *ProductService) CreateCategory(category *models.Category) error {
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()

	id, err := s.repo.CreateStruct("categories", category)
	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}
	category.ID = uint(id)
	return nil
}

// GetAllCategories retrieves all categories
func (s *ProductService) GetAllCategories() ([]models.Category, error) {
	var categories []models.Category
	conditions := map[string]interface{}{"is_active": true}

	err := s.repo.FindAll("categories", &categories, conditions, "sort_order ASC", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	return categories, nil
}

// GetCategoryByID retrieves a category by ID
func (s *ProductService) GetCategoryByID(id int) (*models.Category, error) {
	var category models.Category
	err := s.repo.FindByID("categories", id, &category)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	return &category, nil
}

// AddProductImage adds an image to a product
func (s *ProductService) AddProductImage(image *models.ProductImage) error {
	image.CreatedAt = time.Now()

	id, err := s.repo.CreateStruct("product_images", image)
	if err != nil {
		return fmt.Errorf("failed to add product image: %w", err)
	}
	image.ID = int(id)
	return nil
}

// GetProductImages retrieves images for a product
func (s *ProductService) GetProductImages(productID int) ([]models.ProductImage, error) {
	var images []models.ProductImage
	conditions := map[string]interface{}{"product_id": productID}

	err := s.repo.FindAll("product_images", &images, conditions, "sort_order ASC", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get product images: %w", err)
	}
	return images, nil
}

// AddProductReview adds a review to a product
func (s *ProductService) AddProductReview(review *models.ProductReview) error {
	review.CreatedAt = time.Now()
	review.UpdatedAt = time.Now()
	if review.Status == "" {
		review.Status = "pending"
	}

	id, err := s.repo.CreateStruct("product_reviews", review)
	if err != nil {
		return fmt.Errorf("failed to add product review: %w", err)
	}
	review.ID = int(id)

	// Update product rating
	go s.updateProductRating(review.ProductID)

	return nil
}

// GetProductReviews retrieves reviews for a product
func (s *ProductService) GetProductReviews(productID int, limit, offset int) ([]models.ProductReview, error) {
	var reviews []models.ProductReview
	conditions := map[string]interface{}{
		"product_id": productID,
		"status":     "approved",
	}

	err := s.repo.FindAll("product_reviews", &reviews, conditions, "created_at DESC", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get product reviews: %w", err)
	}
	return reviews, nil
}

// updateProductRating updates the average rating for a product
func (s *ProductService) updateProductRating(productID int) {
	// This would typically be done with a SQL query to calculate average
	// For now, we'll implement a basic version
	reviews, err := s.GetProductReviews(productID, 0, 0)
	if err != nil {
		return
	}

	if len(reviews) == 0 {
		return
	}

	var totalRating float64
	for _, review := range reviews {
		totalRating += float64(review.Rating)
	}

	avgRating := totalRating / float64(len(reviews))

	product, err := s.GetProductByID(productID)
	if err != nil {
		return
	}

	product.Rating = avgRating
	product.ReviewCount = len(reviews)
	product.UpdatedAt = time.Now()

	s.UpdateProduct(productID, product)
}

// GetProductsBySKU retrieves a product by SKU
func (s *ProductService) GetProductBySKU(sku string) (*models.Product, error) {
	var product models.Product
	conditions := map[string]interface{}{"sku": sku}
	err := s.repo.FindOne("products", &product, conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to get product by SKU: %w", err)
	}
	return &product, nil
}

// GenerateSKU generates a unique SKU for a product
func (s *ProductService) GenerateSKU(productName string, vendorID int) string {
	// Simple SKU generation logic
	prefix := strings.ToUpper(strings.ReplaceAll(productName[:min(3, len(productName))], " ", ""))
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s-%d-%d", prefix, vendorID, timestamp)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetProductCount returns the total number of products
func (s *ProductService) GetProductCount(filters ...map[string]interface{}) (int64, error) {
	conditions := map[string]interface{}{}
	
	// If filters are provided, use them
	if len(filters) > 0 {
		for key, value := range filters[0] {
			if key == "category" && value != "" {
				conditions["category"] = value
			} else if key == "min_price" && value.(float64) > 0 {
				conditions["price >="] = value
			} else if key == "max_price" && value.(float64) > 0 {
				conditions["price <="] = value
			} else if key == "status" && value != "" {
				conditions["status"] = value
			} else if key == "vendor_id" && value.(int64) > 0 {
				conditions["vendor_id"] = value
			}
		}
	}
	
	return s.repo.Count("products", conditions)
}

// GetAllProducts returns all products with pagination
func (s *ProductService) GetAllProducts(limit, offset int) ([]models.Product, error) {
	var products []models.Product
	conditions := map[string]interface{}{}
	err := s.repo.FindAll("products", &products, conditions, "created_at DESC", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get all products: %w", err)
	}
	return products, nil
}

// GetRecentProducts returns recently added products
func (s *ProductService) GetRecentProducts(limit int) ([]models.Product, error) {
	var products []models.Product
	conditions := map[string]interface{}{}
	err := s.repo.FindAll("products", &products, conditions, "created_at DESC", limit, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent products: %w", err)
	}
	return products, nil
}

// GetProductsWithFilters gets products with various filters
func (s *ProductService) GetProductsWithFilters(filters map[string]interface{}, sortBy, sortOrder string, limit, offset int) ([]models.Product, error) {
	var products []models.Product
	
	// Build conditions from filters
	conditions := make(map[string]interface{})
	
	if category, ok := filters["category"]; ok && category != "" {
		conditions["category"] = category
	}
	
	if minPrice, ok := filters["min_price"]; ok && minPrice.(float64) > 0 {
		conditions["price >="] = minPrice
	}
	
	if maxPrice, ok := filters["max_price"]; ok && maxPrice.(float64) > 0 {
		conditions["price <="] = maxPrice
	}
	
	if status, ok := filters["status"]; ok && status != "" {
		conditions["status"] = status
	}
	
	if vendorID, ok := filters["vendor_id"]; ok && vendorID.(int64) > 0 {
		conditions["vendor_id"] = vendorID
	}
	
	// Build order by clause
	orderBy := "created_at DESC"
	if sortBy != "" {
		if sortOrder == "" {
			sortOrder = "ASC"
		}
		orderBy = fmt.Sprintf("%s %s", sortBy, sortOrder)
	}
	
	err := s.repo.FindAll("products", &products, conditions, orderBy, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get products with filters: %w", err)
	}
	
	return products, nil
}

// IncrementViewCount increments the view count for a product
func (s *ProductService) IncrementViewCount(productID int) error {
	// This would typically update the views column in the database
	// For now, we'll just return nil as a placeholder
	return nil
}
