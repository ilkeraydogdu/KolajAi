package services

import (
	"fmt"
	"kolajAi/internal/database"
	"kolajAi/internal/models"
	"math"
	"sort"
	"strings"
)

// AIAnalyticsService provides advanced AI analytics and insights
type AIAnalyticsService struct {
	repo           database.SimpleRepository
	productService *ProductService
	orderService   *OrderService
}

// NewAIAnalyticsService creates a new AI analytics service
func NewAIAnalyticsService(repo database.SimpleRepository, productService *ProductService, orderService *OrderService) *AIAnalyticsService {
	return &AIAnalyticsService{
		repo:           repo,
		productService: productService,
		orderService:   orderService,
	}
}

// MarketTrend represents market trend analysis
type MarketTrend struct {
	CategoryID     int     `json:"category_id"`
	CategoryName   string  `json:"category_name"`
	TrendScore     float64 `json:"trend_score"` // -1 to 1 (declining to growing)
	GrowthRate     float64 `json:"growth_rate"` // Percentage growth
	PopularityRank int     `json:"popularity_rank"`
	AveragePrice   float64 `json:"average_price"`
	TotalProducts  int     `json:"total_products"`
	TotalSales     int     `json:"total_sales"`
	Prediction     string  `json:"prediction"`
}

// ProductInsight represents AI-powered product insights
type ProductInsight struct {
	ProductID        int      `json:"product_id"`
	ProductName      string   `json:"product_name"`
	PerformanceScore float64  `json:"performance_score"` // 0 to 1
	SentimentScore   float64  `json:"sentiment_score"`   // -1 to 1
	MarketPosition   string   `json:"market_position"`   // "leader", "challenger", "follower", "niche"
	Recommendations  []string `json:"recommendations"`
	RiskFactors      []string `json:"risk_factors"`
	Opportunities    []string `json:"opportunities"`
}

// CustomerSegment represents customer behavior analysis
type CustomerSegment struct {
	SegmentID           string   `json:"segment_id"`
	SegmentName         string   `json:"segment_name"`
	CustomerCount       int      `json:"customer_count"`
	AverageSpend        float64  `json:"average_spend"`
	PreferredCategories []string `json:"preferred_categories"`
	BehaviorProfile     string   `json:"behavior_profile"`
	GrowthPotential     float64  `json:"growth_potential"` // 0 to 1
	Characteristics     []string `json:"characteristics"`
}

// PricingStrategy represents AI-powered pricing recommendations
type PricingStrategy struct {
	ProductID         int     `json:"product_id"`
	CurrentPrice      float64 `json:"current_price"`
	OptimalPriceRange struct {
		Min float64 `json:"min"`
		Max float64 `json:"max"`
	} `json:"optimal_price_range"`
	RecommendedPrice float64 `json:"recommended_price"`
	ExpectedImpact   struct {
		SalesChange   float64 `json:"sales_change"`   // Percentage
		RevenueChange float64 `json:"revenue_change"` // Percentage
		MarketShare   float64 `json:"market_share"`   // Percentage
	} `json:"expected_impact"`
	Strategy      string   `json:"strategy"` // "penetration", "skimming", "competitive", "value"
	Justification []string `json:"justification"`
}

// AnalyzeMarketTrends analyzes market trends across categories
func (s *AIAnalyticsService) AnalyzeMarketTrends() ([]*MarketTrend, error) {
	categories, err := s.productService.GetAllCategories()
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	trends := make([]*MarketTrend, 0)

	for _, category := range categories {
		trend, err := s.analyzeCategoryTrend(category)
		if err != nil {
			continue // Skip categories with insufficient data
		}
		trends = append(trends, trend)
	}

	// Sort by trend score (highest first)
	sort.Slice(trends, func(i, j int) bool {
		return trends[i].TrendScore > trends[j].TrendScore
	})

	// Assign popularity ranks
	for i, trend := range trends {
		trend.PopularityRank = i + 1
	}

	return trends, nil
}

// analyzeCategoryTrend analyzes trend for a specific category
func (s *AIAnalyticsService) analyzeCategoryTrend(category models.Category) (*MarketTrend, error) {
	// Get products in this category
	allProducts, err := s.productService.GetAllProducts(1000, 0)
	if err != nil {
		return nil, err
	}

	categoryProducts := make([]models.Product, 0)
	for _, product := range allProducts {
		if product.CategoryID == int(category.ID) {
			categoryProducts = append(categoryProducts, product)
		}
	}

	if len(categoryProducts) == 0 {
		return nil, fmt.Errorf("no products in category")
	}

	// Calculate metrics
	totalProducts := len(categoryProducts)
	totalViews := 0
	totalSales := 0
	totalPrice := 0.0
	recentViews := 0 // Views in last 30 days (simulated)

	for _, product := range categoryProducts {
		totalViews += product.ViewCount
		totalSales += product.SalesCount
		totalPrice += product.Price

		// Simulate recent activity (in real implementation, you'd check actual dates)
		if product.ViewCount > 50 {
			recentViews += product.ViewCount / 2 // Assume half are recent
		}
	}

	averagePrice := totalPrice / float64(totalProducts)

	// Calculate trend score (-1 to 1)
	// This is a simplified calculation - in reality you'd use time series analysis
	trendScore := 0.0
	if totalViews > 0 {
		recentActivityRatio := float64(recentViews) / float64(totalViews)
		trendScore = (recentActivityRatio - 0.3) * 2 // Normalize to -1 to 1 range
		if trendScore > 1 {
			trendScore = 1
		} else if trendScore < -1 {
			trendScore = -1
		}
	}

	// Calculate growth rate (simplified)
	growthRate := trendScore * 100 // Convert to percentage

	// Generate prediction
	prediction := s.generateTrendPrediction(trendScore, totalSales, totalViews)

	return &MarketTrend{
						CategoryID: int(category.ID),
		CategoryName:  category.Name,
		TrendScore:    trendScore,
		GrowthRate:    growthRate,
		AveragePrice:  averagePrice,
		TotalProducts: totalProducts,
		TotalSales:    totalSales,
		Prediction:    prediction,
	}, nil
}

// generateTrendPrediction generates a human-readable trend prediction
func (s *AIAnalyticsService) generateTrendPrediction(trendScore float64, _, _ int) string {
	if trendScore > 0.5 {
		return "Güçlü büyüme trendi - Yatırım fırsatı"
	} else if trendScore > 0.2 {
		return "Pozitif trend - Büyüme potansiyeli var"
	} else if trendScore > -0.2 {
		return "Stabil pazar - Mevcut konumu koruma"
	} else if trendScore > -0.5 {
		return "Düşüş trendi - Dikkatli izleme gerekli"
	} else {
		return "Güçlü düşüş trendi - Strateji değişikliği öneriliyor"
	}
}

// AnalyzeProductInsights provides comprehensive product analysis
func (s *AIAnalyticsService) AnalyzeProductInsights(productID int) (*ProductInsight, error) {
	product, err := s.productService.GetProductByID(productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Calculate performance score
	performanceScore := s.calculateProductPerformance(product)

	// Analyze sentiment (simplified - in reality you'd analyze reviews/comments)
	sentimentScore := s.analyzeSentiment(product)

	// Determine market position
	marketPosition := s.determineMarketPosition(product, performanceScore)

	// Generate recommendations
	recommendations := s.generateProductRecommendations(product, performanceScore, sentimentScore)

	// Identify risk factors
	riskFactors := s.identifyRiskFactors(product, performanceScore, sentimentScore)

	// Find opportunities
	opportunities := s.findOpportunities(product, performanceScore, sentimentScore)

	return &ProductInsight{
		ProductID:        productID,
		ProductName:      product.Name,
		PerformanceScore: performanceScore,
		SentimentScore:   sentimentScore,
		MarketPosition:   marketPosition,
		Recommendations:  recommendations,
		RiskFactors:      riskFactors,
		Opportunities:    opportunities,
	}, nil
}

// calculateProductPerformance calculates a performance score for a product
func (s *AIAnalyticsService) calculateProductPerformance(product *models.Product) float64 {
	score := 0.0

	// Sales performance (40% of score)
	if product.SalesCount > 0 {
		salesScore := math.Min(float64(product.SalesCount)/100.0, 1.0) // Normalize to 0-1
		score += salesScore * 0.4
	}

	// View performance (30% of score)
	if product.ViewCount > 0 {
		viewScore := math.Min(float64(product.ViewCount)/1000.0, 1.0) // Normalize to 0-1
		score += viewScore * 0.3
	}

	// Stock management (20% of score)
	if product.Stock > 0 {
		stockScore := 1.0
		if product.Stock < product.MinStock {
			stockScore = 0.5 // Penalty for low stock
		}
		score += stockScore * 0.2
	}

	// Rating performance (10% of score)
	if product.Rating > 0 {
		ratingScore := product.Rating / 5.0 // Normalize to 0-1
		score += ratingScore * 0.1
	}

	return math.Min(score, 1.0)
}

// analyzeSentiment analyzes product sentiment (simplified implementation)
func (s *AIAnalyticsService) analyzeSentiment(product *models.Product) float64 {
	// This is a simplified sentiment analysis
	// In a real implementation, you'd analyze actual reviews and comments

	sentiment := 0.0

	// Base sentiment from rating
	if product.Rating > 0 {
		sentiment = (product.Rating - 3.0) / 2.0 // Convert 1-5 rating to -1 to 1 scale
	}

	// Adjust based on sales vs views ratio
	if product.ViewCount > 0 {
		conversionRate := float64(product.SalesCount) / float64(product.ViewCount)
		if conversionRate > 0.1 { // Good conversion rate
			sentiment += 0.2
		} else if conversionRate < 0.02 { // Poor conversion rate
			sentiment -= 0.2
		}
	}

	// Adjust based on stock levels
	if product.Stock == 0 {
		sentiment -= 0.1 // Out of stock is negative
	}

	// Clamp to -1 to 1 range
	if sentiment > 1 {
		sentiment = 1
	} else if sentiment < -1 {
		sentiment = -1
	}

	return sentiment
}

// determineMarketPosition determines the market position of a product
func (s *AIAnalyticsService) determineMarketPosition(_ *models.Product, performanceScore float64) string {
	if performanceScore > 0.8 {
		return "leader"
	} else if performanceScore > 0.6 {
		return "challenger"
	} else if performanceScore > 0.3 {
		return "follower"
	} else {
		return "niche"
	}
}

// generateProductRecommendations generates actionable recommendations
func (s *AIAnalyticsService) generateProductRecommendations(product *models.Product, performance, sentiment float64) []string {
	recommendations := make([]string, 0)

	if performance < 0.5 {
		recommendations = append(recommendations, "Ürün tanıtımını artırın ve pazarlama stratejisini gözden geçirin")
	}

	if sentiment < 0 {
		recommendations = append(recommendations, "Müşteri geri bildirimlerini analiz edin ve ürün kalitesini iyileştirin")
	}

	if product.Stock < product.MinStock {
		recommendations = append(recommendations, "Stok seviyesini artırın - talep var ancak stok yetersiz")
	}

	if product.ViewCount > product.SalesCount*20 { // High views, low sales
		recommendations = append(recommendations, "Fiyatlandırma stratejisini gözden geçirin - çok görüntüleniyor ama satılmıyor")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Mevcut performans iyi - stratejinizi sürdürün")
	}

	return recommendations
}

// identifyRiskFactors identifies potential risks
func (s *AIAnalyticsService) identifyRiskFactors(product *models.Product, performance, sentiment float64) []string {
	risks := make([]string, 0)

	if sentiment < -0.3 {
		risks = append(risks, "Düşük müşteri memnuniyeti - marka imajına zarar verebilir")
	}

	if product.Stock == 0 {
		risks = append(risks, "Stok tükendi - müşteri kaybı riski")
	}

	if performance < 0.3 {
		risks = append(risks, "Düşük performans - karlılık riski")
	}

	if product.ViewCount < 10 { // Very low visibility
		risks = append(risks, "Düşük görünürlük - pazar payı kaybı riski")
	}

	return risks
}

// findOpportunities identifies growth opportunities
func (s *AIAnalyticsService) findOpportunities(product *models.Product, performance, sentiment float64) []string {
	opportunities := make([]string, 0)

	if sentiment > 0.3 && performance < 0.7 {
		opportunities = append(opportunities, "Yüksek müşteri memnuniyeti - pazarlama yatırımı ile büyüme fırsatı")
	}

	if product.ViewCount > 100 && product.SalesCount < 10 {
		opportunities = append(opportunities, "Yüksek ilgi - fiyat optimizasyonu ile satış artırma fırsatı")
	}

	if performance > 0.7 {
		opportunities = append(opportunities, "Güçlü performans - premium fiyatlandırma fırsatı")
	}

	if strings.Contains(strings.ToLower(product.Tags), "trend") {
		opportunities = append(opportunities, "Trend ürün - hızlı büyüme potansiyeli")
	}

	return opportunities
}

// AnalyzeCustomerSegments analyzes customer behavior and segments
func (s *AIAnalyticsService) AnalyzeCustomerSegments() ([]*CustomerSegment, error) {
	// This is a simplified implementation
	// In a real system, you'd analyze actual customer data

	segments := []*CustomerSegment{
		{
			SegmentID:           "high_value",
			SegmentName:         "Yüksek Değerli Müşteriler",
			CustomerCount:       150,
			AverageSpend:        500.0,
			PreferredCategories: []string{"Elektronik", "Moda", "Ev & Yaşam"},
			BehaviorProfile:     "Sık alışveriş yapan, kaliteye önem veren müşteriler",
			GrowthPotential:     0.8,
			Characteristics:     []string{"Yüksek gelir", "Marka sadakati", "Kalite odaklı"},
		},
		{
			SegmentID:           "price_sensitive",
			SegmentName:         "Fiyat Hassas Müşteriler",
			CustomerCount:       300,
			AverageSpend:        150.0,
			PreferredCategories: []string{"Gıda", "Temizlik", "Temel İhtiyaçlar"},
			BehaviorProfile:     "İndirim ve kampanyaları takip eden, fiyat karşılaştırması yapan müşteriler",
			GrowthPotential:     0.6,
			Characteristics:     []string{"Fiyat odaklı", "Kampanya takipçisi", "Pratik çözüm arayan"},
		},
		{
			SegmentID:           "tech_enthusiasts",
			SegmentName:         "Teknoloji Meraklıları",
			CustomerCount:       200,
			AverageSpend:        350.0,
			PreferredCategories: []string{"Elektronik", "Bilgisayar", "Akıllı Cihazlar"},
			BehaviorProfile:     "Yeni teknolojileri erken benimseyen, araştırma yapan müşteriler",
			GrowthPotential:     0.9,
			Characteristics:     []string{"İnovasyon odaklı", "Araştırmacı", "Sosyal medya aktif"},
		},
	}

	return segments, nil
}

// GeneratePricingStrategy generates AI-powered pricing recommendations
func (s *AIAnalyticsService) GeneratePricingStrategy(productID int) (*PricingStrategy, error) {
	product, err := s.productService.GetProductByID(productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Get similar products for market analysis
	similarProducts, err := s.getSimilarProductsForPricing(product, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get similar products: %w", err)
	}

	// Calculate optimal price range
	minPrice, maxPrice := s.calculateOptimalPriceRange(product, similarProducts)

	// Determine recommended price
	recommendedPrice := s.calculateRecommendedPrice(product, similarProducts, minPrice, maxPrice)

	// Calculate expected impact
	salesChange, revenueChange, marketShare := s.calculatePriceImpact(product, recommendedPrice)

	// Determine strategy
	strategy := s.determinePricingStrategy(product, recommendedPrice, similarProducts)

	// Generate justification
	justification := s.generatePricingJustification(product, recommendedPrice, strategy)

	return &PricingStrategy{
		ProductID:    productID,
		CurrentPrice: product.Price,
		OptimalPriceRange: struct {
			Min float64 `json:"min"`
			Max float64 `json:"max"`
		}{
			Min: minPrice,
			Max: maxPrice,
		},
		RecommendedPrice: recommendedPrice,
		ExpectedImpact: struct {
			SalesChange   float64 `json:"sales_change"`
			RevenueChange float64 `json:"revenue_change"`
			MarketShare   float64 `json:"market_share"`
		}{
			SalesChange:   salesChange,
			RevenueChange: revenueChange,
			MarketShare:   marketShare,
		},
		Strategy:      strategy,
		Justification: justification,
	}, nil
}

// Helper methods for pricing strategy
func (s *AIAnalyticsService) getSimilarProductsForPricing(product *models.Product, limit int) ([]*models.Product, error) {
	// Reuse the similar products logic from the main AI service
	allProducts, err := s.productService.GetAllProducts(500, 0)
	if err != nil {
		return nil, err
	}

	scored := make([]struct {
		product *models.Product
		score   float64
	}, 0)

	for _, p := range allProducts {
		if p.ID == product.ID || p.Status != "active" {
			continue
		}

		score := 0.0
		if p.CategoryID == product.CategoryID {
			score += 0.6
		}

		priceDiff := math.Abs(p.Price - product.Price)
		maxPrice := math.Max(p.Price, product.Price)
		if maxPrice > 0 {
			priceScore := 1.0 - (priceDiff / maxPrice)
			score += priceScore * 0.4
		}

		if score > 0.3 {
			scored = append(scored, struct {
				product *models.Product
				score   float64
			}{product: &p, score: score})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	result := make([]*models.Product, 0, limit)
	for i, ps := range scored {
		if i >= limit {
			break
		}
		result = append(result, ps.product)
	}

	return result, nil
}

func (s *AIAnalyticsService) calculateOptimalPriceRange(product *models.Product, similarProducts []*models.Product) (float64, float64) {
	if len(similarProducts) == 0 {
		return product.Price * 0.8, product.Price * 1.2
	}

	prices := make([]float64, len(similarProducts))
	for i, p := range similarProducts {
		prices[i] = p.Price
	}

	sort.Float64s(prices)

	minPrice := prices[0] * 0.9
	maxPrice := prices[len(prices)-1] * 1.1

	return minPrice, maxPrice
}

func (s *AIAnalyticsService) calculateRecommendedPrice(product *models.Product, similarProducts []*models.Product, minPrice, maxPrice float64) float64 {
	if len(similarProducts) == 0 {
		return product.Price
	}

	totalPrice := 0.0
	for _, p := range similarProducts {
		totalPrice += p.Price
	}

	marketAverage := totalPrice / float64(len(similarProducts))

	// Adjust based on product performance
	performanceScore := s.calculateProductPerformance(product)
	adjustment := (performanceScore - 0.5) * 0.2 // ±20% adjustment

	recommendedPrice := marketAverage * (1 + adjustment)

	// Ensure it's within the optimal range
	if recommendedPrice < minPrice {
		recommendedPrice = minPrice
	} else if recommendedPrice > maxPrice {
		recommendedPrice = maxPrice
	}

	return math.Round(recommendedPrice*100) / 100
}

func (s *AIAnalyticsService) calculatePriceImpact(product *models.Product, newPrice float64) (float64, float64, float64) {
	priceChange := (newPrice - product.Price) / product.Price

	// Simplified elasticity calculation
	// In reality, you'd use historical data and more sophisticated models
	elasticity := -1.5 // Assume price elasticity of demand

	salesChange := elasticity * priceChange * 100
	revenueChange := (1+priceChange)*(1+salesChange/100) - 1
	revenueChange *= 100

	// Market share impact (simplified)
	marketShare := 5.0 + (priceChange * -10) // Lower price = higher market share
	if marketShare < 0 {
		marketShare = 0
	} else if marketShare > 20 {
		marketShare = 20
	}

	return salesChange, revenueChange, marketShare
}

func (s *AIAnalyticsService) determinePricingStrategy(_ *models.Product, recommendedPrice float64, similarProducts []*models.Product) string {
	if len(similarProducts) == 0 {
		return "value"
	}

	totalPrice := 0.0
	for _, p := range similarProducts {
		totalPrice += p.Price
	}
	marketAverage := totalPrice / float64(len(similarProducts))

	if recommendedPrice < marketAverage*0.9 {
		return "penetration"
	} else if recommendedPrice > marketAverage*1.1 {
		return "skimming"
	} else if math.Abs(recommendedPrice-marketAverage) < marketAverage*0.05 {
		return "competitive"
	} else {
		return "value"
	}
}

func (s *AIAnalyticsService) generatePricingJustification(product *models.Product, recommendedPrice float64, strategy string) []string {
	justification := make([]string, 0)

	priceChange := recommendedPrice - product.Price
	changePercent := (priceChange / product.Price) * 100

	switch strategy {
	case "penetration":
		justification = append(justification, "Pazar penetrasyon stratejisi - düşük fiyat ile pazar payı artırma")
		justification = append(justification, fmt.Sprintf("%.1f%% fiyat düşürme ile rekabet avantajı", -changePercent))
	case "skimming":
		justification = append(justification, "Fiyat skimming stratejisi - premium konumlandırma")
		justification = append(justification, fmt.Sprintf("%.1f%% fiyat artışı ile yüksek kar marjı", changePercent))
	case "competitive":
		justification = append(justification, "Rekabetçi fiyatlandırma - pazar ortalamasında konumlandırma")
		justification = append(justification, "Mevcut pazar koşullarına uygun fiyat seviyesi")
	case "value":
		justification = append(justification, "Değer odaklı fiyatlandırma - kalite-fiyat dengesi")
		justification = append(justification, "Ürün değeri ile uyumlu fiyat seviyesi")
	}

	if product.ViewCount > product.SalesCount*10 {
		justification = append(justification, "Yüksek ilgi ancak düşük satış - fiyat optimizasyonu gerekli")
	}

	return justification
}
