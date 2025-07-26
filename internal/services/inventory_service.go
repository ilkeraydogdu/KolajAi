package services

import (
	"fmt"
	"kolajAi/internal/database"
	"kolajAi/internal/models"
	"math"
	"sort"
	"time"
)

// InventoryService provides inventory management with AI predictions
type InventoryService struct {
	repo           database.SimpleRepository
	productService *ProductService
	orderService   *OrderService
}

// NewInventoryService creates a new inventory service
func NewInventoryService(repo database.SimpleRepository, productService *ProductService, orderService *OrderService) *InventoryService {
	return &InventoryService{
		repo:           repo,
		productService: productService,
		orderService:   orderService,
	}
}

// StockPrediction represents AI-powered stock level predictions
type StockPrediction struct {
	ProductID          int     `json:"product_id"`
	ProductName        string  `json:"product_name"`
	CurrentStock       int     `json:"current_stock"`
	MinStock           int     `json:"min_stock"`
	PredictedDemand    float64 `json:"predicted_demand"` // Units per day
	DaysUntilStockout  int     `json:"days_until_stockout"`
	RecommendedReorder int     `json:"recommended_reorder"`
	ReorderUrgency     string  `json:"reorder_urgency"`  // "critical", "high", "medium", "low"
	ConfidenceLevel    float64 `json:"confidence_level"` // 0 to 1
	SeasonalFactor     float64 `json:"seasonal_factor"`  // Seasonal adjustment
	TrendFactor        float64 `json:"trend_factor"`     // Growth trend factor
}

// InventoryAlert represents inventory alerts and notifications
type InventoryAlert struct {
	ID          string    `json:"id"`
	ProductID   int       `json:"product_id"`
	ProductName string    `json:"product_name"`
	AlertType   string    `json:"alert_type"` // "low_stock", "out_of_stock", "overstock", "reorder"
	Severity    string    `json:"severity"`   // "critical", "high", "medium", "low"
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
	IsRead      bool      `json:"is_read"`
}

// InventoryOptimization represents inventory optimization recommendations
type InventoryOptimization struct {
	ProductID             int      `json:"product_id"`
	ProductName           string   `json:"product_name"`
	CurrentStock          int      `json:"current_stock"`
	OptimalStock          int      `json:"optimal_stock"`
	OptimalMinStock       int      `json:"optimal_min_stock"`
	OptimalMaxStock       int      `json:"optimal_max_stock"`
	CarryingCost          float64  `json:"carrying_cost"`          // Cost of holding inventory
	StockoutCost          float64  `json:"stockout_cost"`          // Cost of being out of stock
	OptimizationPotential float64  `json:"optimization_potential"` // Potential cost savings
	Recommendations       []string `json:"recommendations"`
}

// SupplierPerformance represents supplier performance metrics
type SupplierPerformance struct {
	SupplierID      int      `json:"supplier_id"`
	SupplierName    string   `json:"supplier_name"`
	DeliveryScore   float64  `json:"delivery_score"` // 0 to 1
	QualityScore    float64  `json:"quality_score"`  // 0 to 1
	PriceScore      float64  `json:"price_score"`    // 0 to 1
	OverallScore    float64  `json:"overall_score"`  // 0 to 1
	OrderCount      int      `json:"order_count"`
	OnTimeDelivery  float64  `json:"on_time_delivery"`  // Percentage
	AverageLeadTime int      `json:"average_lead_time"` // Days
	Recommendations []string `json:"recommendations"`
}

// PredictStockLevels predicts future stock levels for all products
func (s *InventoryService) PredictStockLevels() ([]*StockPrediction, error) {
	products, err := s.productService.GetAllProducts(1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	predictions := make([]*StockPrediction, 0)

	for _, product := range products {
		if product.Status != "active" {
			continue
		}

		prediction, err := s.predictProductStock(&product)
		if err != nil {
			continue // Skip products with insufficient data
		}

		predictions = append(predictions, prediction)
	}

	// Sort by urgency (critical first)
	sort.Slice(predictions, func(i, j int) bool {
		urgencyOrder := map[string]int{
			"critical": 0,
			"high":     1,
			"medium":   2,
			"low":      3,
		}
		return urgencyOrder[predictions[i].ReorderUrgency] < urgencyOrder[predictions[j].ReorderUrgency]
	})

	return predictions, nil
}

// predictProductStock predicts stock levels for a specific product
func (s *InventoryService) predictProductStock(product *models.Product) (*StockPrediction, error) {
	// Calculate historical demand
	dailyDemand := s.calculateDailyDemand(product)

	// Apply seasonal and trend factors
	seasonalFactor := s.calculateSeasonalFactor(product)
	trendFactor := s.calculateTrendFactor(product)

	// Adjust demand prediction
	predictedDemand := dailyDemand * seasonalFactor * trendFactor

	// Calculate days until stockout
	daysUntilStockout := 0
	if predictedDemand > 0 {
		daysUntilStockout = int(float64(product.Stock) / predictedDemand)
	} else {
		daysUntilStockout = 999 // Very high number if no demand
	}

	// Calculate recommended reorder quantity
	recommendedReorder := s.calculateReorderQuantity(product, predictedDemand)

	// Determine urgency
	urgency := s.determineReorderUrgency(daysUntilStockout, product.Stock, product.MinStock)

	// Calculate confidence level
	confidence := s.calculatePredictionConfidence(product)

	return &StockPrediction{
		ProductID:          product.ID,
		ProductName:        product.Name,
		CurrentStock:       product.Stock,
		MinStock:           product.MinStock,
		PredictedDemand:    predictedDemand,
		DaysUntilStockout:  daysUntilStockout,
		RecommendedReorder: recommendedReorder,
		ReorderUrgency:     urgency,
		ConfidenceLevel:    confidence,
		SeasonalFactor:     seasonalFactor,
		TrendFactor:        trendFactor,
	}, nil
}

// calculateDailyDemand calculates average daily demand for a product
func (s *InventoryService) calculateDailyDemand(product *models.Product) float64 {
	// This is a simplified calculation
	// In a real implementation, you'd analyze historical sales data

	if product.SalesCount == 0 {
		return 0.1 // Very low demand for products with no sales
	}

	// Assume the product has been available for 30 days (simplified)
	// In reality, you'd calculate from the actual creation date
	daysSinceCreation := 30.0

	// Calculate daily demand based on total sales
	dailyDemand := float64(product.SalesCount) / daysSinceCreation

	// Add some randomness based on view count (interest level)
	interestFactor := math.Min(float64(product.ViewCount)/1000.0, 2.0) + 0.5
	dailyDemand *= interestFactor

	return math.Max(dailyDemand, 0.1) // Minimum demand
}

// calculateSeasonalFactor calculates seasonal adjustment factor
func (s *InventoryService) calculateSeasonalFactor(product *models.Product) float64 {
	// This is a simplified seasonal calculation
	// In reality, you'd analyze historical seasonal patterns

	now := time.Now()
	month := now.Month()

	// Simple seasonal factors by month
	seasonalFactors := map[time.Month]float64{
		time.January:   0.8, // Post-holiday slowdown
		time.February:  0.9, // Winter
		time.March:     1.0, // Spring starts
		time.April:     1.1, // Spring
		time.May:       1.2, // Spring peak
		time.June:      1.1, // Early summer
		time.July:      1.0, // Summer
		time.August:    0.9, // Late summer
		time.September: 1.1, // Back to school
		time.October:   1.2, // Fall shopping
		time.November:  1.4, // Pre-holiday
		time.December:  1.5, // Holiday peak
	}

	return seasonalFactors[month]
}

// calculateTrendFactor calculates growth trend factor
func (s *InventoryService) calculateTrendFactor(product *models.Product) float64 {
	// This is a simplified trend calculation
	// In reality, you'd analyze time series data

	// Use view count as a proxy for trend
	if product.ViewCount > 500 {
		return 1.3 // Growing trend
	} else if product.ViewCount > 100 {
		return 1.1 // Moderate growth
	} else if product.ViewCount > 50 {
		return 1.0 // Stable
	} else {
		return 0.8 // Declining
	}
}

// calculateReorderQuantity calculates optimal reorder quantity
func (s *InventoryService) calculateReorderQuantity(product *models.Product, dailyDemand float64) int {
	// Economic Order Quantity (EOQ) simplified calculation
	// EOQ = sqrt(2 * D * S / H)
	// D = annual demand, S = ordering cost, H = holding cost

	annualDemand := dailyDemand * 365
	orderingCost := 50.0               // Simplified ordering cost
	holdingCost := product.Price * 0.2 // 20% of product price per year

	if holdingCost <= 0 {
		holdingCost = 1.0 // Minimum holding cost
	}

	eoq := math.Sqrt(2 * annualDemand * orderingCost / holdingCost)

	// Ensure minimum order quantity
	minOrder := int(dailyDemand * 30) // 30 days worth
	if int(eoq) < minOrder {
		return minOrder
	}

	return int(eoq)
}

// determineReorderUrgency determines the urgency level for reordering
func (s *InventoryService) determineReorderUrgency(daysUntilStockout, currentStock, minStock int) string {
	if currentStock <= 0 {
		return "critical"
	} else if currentStock <= minStock {
		return "high"
	} else if daysUntilStockout <= 7 {
		return "high"
	} else if daysUntilStockout <= 14 {
		return "medium"
	} else {
		return "low"
	}
}

// calculatePredictionConfidence calculates confidence level for predictions
func (s *InventoryService) calculatePredictionConfidence(product *models.Product) float64 {
	confidence := 0.5 // Base confidence

	// More sales history = higher confidence
	if product.SalesCount > 50 {
		confidence += 0.3
	} else if product.SalesCount > 10 {
		confidence += 0.2
	} else if product.SalesCount > 0 {
		confidence += 0.1
	}

	// More views = higher confidence in demand pattern
	if product.ViewCount > 1000 {
		confidence += 0.2
	} else if product.ViewCount > 100 {
		confidence += 0.1
	}

	return math.Min(confidence, 1.0)
}

// GenerateInventoryAlerts generates inventory alerts based on current stock levels
func (s *InventoryService) GenerateInventoryAlerts() ([]*InventoryAlert, error) {
	products, err := s.productService.GetAllProducts(1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	alerts := make([]*InventoryAlert, 0)

	for _, product := range products {
		if product.Status != "active" {
			continue
		}

		// Generate alerts for this product
		productAlerts := s.generateProductAlerts(&product)
		alerts = append(alerts, productAlerts...)
	}

	// Sort by severity
	sort.Slice(alerts, func(i, j int) bool {
		severityOrder := map[string]int{
			"critical": 0,
			"high":     1,
			"medium":   2,
			"low":      3,
		}
		return severityOrder[alerts[i].Severity] < severityOrder[alerts[j].Severity]
	})

	return alerts, nil
}

// generateProductAlerts generates alerts for a specific product
func (s *InventoryService) generateProductAlerts(product *models.Product) []*InventoryAlert {
	alerts := make([]*InventoryAlert, 0)
	now := time.Now()

	// Out of stock alert
	if product.Stock <= 0 {
		alerts = append(alerts, &InventoryAlert{
			ID:          fmt.Sprintf("out_of_stock_%d_%d", product.ID, now.Unix()),
			ProductID:   product.ID,
			ProductName: product.Name,
			AlertType:   "out_of_stock",
			Severity:    "critical",
			Message:     fmt.Sprintf("Ürün '%s' stokta yok!", product.Name),
			CreatedAt:   now,
			IsRead:      false,
		})
	}

	// Low stock alert
	if product.Stock > 0 && product.Stock <= product.MinStock {
		alerts = append(alerts, &InventoryAlert{
			ID:          fmt.Sprintf("low_stock_%d_%d", product.ID, now.Unix()),
			ProductID:   product.ID,
			ProductName: product.Name,
			AlertType:   "low_stock",
			Severity:    "high",
			Message:     fmt.Sprintf("Ürün '%s' stoku düşük: %d adet kaldı", product.Name, product.Stock),
			CreatedAt:   now,
			IsRead:      false,
		})
	}

	// Predict future stockout
	prediction, err := s.predictProductStock(product)
	if err == nil && prediction.DaysUntilStockout <= 7 && prediction.DaysUntilStockout > 0 {
		alerts = append(alerts, &InventoryAlert{
			ID:          fmt.Sprintf("predicted_stockout_%d_%d", product.ID, now.Unix()),
			ProductID:   product.ID,
			ProductName: product.Name,
			AlertType:   "reorder",
			Severity:    "medium",
			Message:     fmt.Sprintf("Ürün '%s' tahmini %d gün içinde tükenecek", product.Name, prediction.DaysUntilStockout),
			CreatedAt:   now,
			IsRead:      false,
		})
	}

	return alerts
}

// OptimizeInventory provides inventory optimization recommendations
func (s *InventoryService) OptimizeInventory() ([]*InventoryOptimization, error) {
	products, err := s.productService.GetAllProducts(1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	optimizations := make([]*InventoryOptimization, 0)

	for _, product := range products {
		if product.Status != "active" {
			continue
		}

		optimization := s.optimizeProductInventory(&product)
		if optimization != nil {
			optimizations = append(optimizations, optimization)
		}
	}

	// Sort by optimization potential (highest savings first)
	sort.Slice(optimizations, func(i, j int) bool {
		return optimizations[i].OptimizationPotential > optimizations[j].OptimizationPotential
	})

	return optimizations, nil
}

// optimizeProductInventory optimizes inventory for a specific product
func (s *InventoryService) optimizeProductInventory(product *models.Product) *InventoryOptimization {
	// Calculate daily demand
	dailyDemand := s.calculateDailyDemand(product)

	// Calculate optimal stock levels
	optimalStock := int(dailyDemand * 30)    // 30 days worth
	optimalMinStock := int(dailyDemand * 7)  // 7 days worth
	optimalMaxStock := int(dailyDemand * 60) // 60 days worth

	// Calculate costs
	carryingCost := s.calculateCarryingCost(product, optimalStock)
	stockoutCost := s.calculateStockoutCost(product, dailyDemand)

	// Calculate optimization potential
	currentCost := s.calculateCarryingCost(product, product.Stock)
	optimizationPotential := math.Abs(currentCost - carryingCost)

	// Generate recommendations
	recommendations := s.generateInventoryRecommendations(product, optimalStock, optimalMinStock, optimalMaxStock)

	return &InventoryOptimization{
		ProductID:             product.ID,
		ProductName:           product.Name,
		CurrentStock:          product.Stock,
		OptimalStock:          optimalStock,
		OptimalMinStock:       optimalMinStock,
		OptimalMaxStock:       optimalMaxStock,
		CarryingCost:          carryingCost,
		StockoutCost:          stockoutCost,
		OptimizationPotential: optimizationPotential,
		Recommendations:       recommendations,
	}
}

// calculateCarryingCost calculates the cost of carrying inventory
func (s *InventoryService) calculateCarryingCost(product *models.Product, stockLevel int) float64 {
	// Carrying cost = (Average inventory value) * (Carrying rate)
	averageInventoryValue := float64(stockLevel) * product.Price / 2
	carryingRate := 0.25 // 25% annual carrying cost

	return averageInventoryValue * carryingRate / 365 // Daily carrying cost
}

// calculateStockoutCost calculates the cost of being out of stock
func (s *InventoryService) calculateStockoutCost(product *models.Product, dailyDemand float64) float64 {
	// Stockout cost = Lost sales + Customer dissatisfaction
	// Simplified: Assume we lose the profit margin on lost sales
	profitMargin := product.Price * 0.3 // Assume 30% profit margin

	return dailyDemand * profitMargin
}

// generateInventoryRecommendations generates optimization recommendations
func (s *InventoryService) generateInventoryRecommendations(product *models.Product, optimalStock, optimalMinStock, optimalMaxStock int) []string {
	recommendations := make([]string, 0)

	if product.Stock > optimalMaxStock {
		recommendations = append(recommendations, fmt.Sprintf("Stok seviyesi çok yüksek - %d adet azaltın", product.Stock-optimalStock))
	} else if product.Stock < optimalMinStock {
		recommendations = append(recommendations, fmt.Sprintf("Stok seviyesi çok düşük - %d adet ekleyin", optimalStock-product.Stock))
	}

	if product.MinStock != optimalMinStock {
		recommendations = append(recommendations, fmt.Sprintf("Minimum stok seviyesini %d olarak güncelleyin", optimalMinStock))
	}

	// Add general recommendations
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Mevcut stok seviyeleri optimal aralıkta")
	}

	return recommendations
}

// AnalyzeSupplierPerformance analyzes supplier performance metrics
func (s *InventoryService) AnalyzeSupplierPerformance() ([]*SupplierPerformance, error) {
	// This is a simplified implementation
	// In a real system, you'd analyze actual supplier data

	suppliers := []*SupplierPerformance{
		{
			SupplierID:      1,
			SupplierName:    "ABC Tedarikçi",
			DeliveryScore:   0.92,
			QualityScore:    0.88,
			PriceScore:      0.85,
			OverallScore:    0.88,
			OrderCount:      45,
			OnTimeDelivery:  92.0,
			AverageLeadTime: 5,
			Recommendations: []string{
				"Kalite kontrolünü artırın",
				"Fiyat rekabetçiliğini iyileştirin",
			},
		},
		{
			SupplierID:      2,
			SupplierName:    "XYZ Tedarik",
			DeliveryScore:   0.95,
			QualityScore:    0.95,
			PriceScore:      0.78,
			OverallScore:    0.89,
			OrderCount:      32,
			OnTimeDelivery:  95.0,
			AverageLeadTime: 3,
			Recommendations: []string{
				"Mükemmel performans - tercih edilen tedarikçi",
				"Fiyat müzakereleri yapılabilir",
			},
		},
		{
			SupplierID:      3,
			SupplierName:    "DEF Supply",
			DeliveryScore:   0.75,
			QualityScore:    0.82,
			PriceScore:      0.92,
			OverallScore:    0.83,
			OrderCount:      28,
			OnTimeDelivery:  75.0,
			AverageLeadTime: 8,
			Recommendations: []string{
				"Teslimat performansını iyileştirin",
				"Lead time'ı kısaltın",
				"Fiyat avantajını koruyun",
			},
		},
	}

	// Sort by overall score
	sort.Slice(suppliers, func(i, j int) bool {
		return suppliers[i].OverallScore > suppliers[j].OverallScore
	})

	return suppliers, nil
}
