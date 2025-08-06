package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"kolajAi/internal/database"
	"kolajAi/internal/models"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// Open Source AI Models Configuration
type OpenSourceAIConfig struct {
	HuggingFaceAPIKey string
	OllamaEndpoint    string
	LocalLLMEndpoint  string
	UseLocalModels    bool
}

// OpenSourceAIClient handles open-source AI model interactions
type OpenSourceAIClient struct {
	config     OpenSourceAIConfig
	httpClient *http.Client
}

// NewOpenSourceAIClient creates a new open-source AI client
func NewOpenSourceAIClient() *OpenSourceAIClient {
	return &OpenSourceAIClient{
		config: OpenSourceAIConfig{
			HuggingFaceAPIKey: os.Getenv("HUGGINGFACE_API_KEY"),
			OllamaEndpoint:    os.Getenv("OLLAMA_ENDPOINT"),
			LocalLLMEndpoint:  getEnvWithDefault("LOCAL_LLM_ENDPOINT", "http://localhost:11434"),
			UseLocalModels:    getEnvWithDefault("USE_LOCAL_AI", "true") == "true",
		},
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// getEnvWithDefault returns environment variable or default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// HuggingFace integration for product recommendations
func (client *OpenSourceAIClient) GetHuggingFaceRecommendations(productName, category string) ([]string, error) {
	if client.config.HuggingFaceAPIKey == "" {
		// Fallback to rule-based recommendations
		return client.getRuleBasedRecommendations(productName, category), nil
	}

	// Use Hugging Face Transformers API
	requestBody := map[string]interface{}{
		"inputs": fmt.Sprintf("Recommend similar products to: %s in category: %s", productName, category),
		"parameters": map[string]interface{}{
			"max_length":           100,
			"num_return_sequences": 5,
		},
	}

	jsonData, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "https://api-inference.huggingface.co/models/microsoft/DialoGPT-medium", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+client.config.HuggingFaceAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return client.getRuleBasedRecommendations(productName, category), nil
	}
	defer resp.Body.Close()

	var response []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return client.getRuleBasedRecommendations(productName, category), nil
	}

	recommendations := make([]string, 0)
	for _, item := range response {
		if text, ok := item["generated_text"].(string); ok {
			recommendations = append(recommendations, text)
		}
	}

	if len(recommendations) == 0 {
		return client.getRuleBasedRecommendations(productName, category), nil
	}

	return recommendations, nil
}

// Ollama integration for local LLM
func (client *OpenSourceAIClient) GetOllamaRecommendations(prompt string) (string, error) {
	if client.config.OllamaEndpoint == "" {
		return client.getDefaultResponse(prompt), nil
	}

	requestBody := map[string]interface{}{
		"model":  "llama2", // or "mistral", "codellama", etc.
		"prompt": prompt,
		"stream": false,
	}

	jsonData, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", client.config.OllamaEndpoint+"/api/generate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return client.getDefaultResponse(prompt), nil
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return client.getDefaultResponse(prompt), nil
	}

	if text, ok := response["response"].(string); ok {
		return text, nil
	}

	return client.getDefaultResponse(prompt), nil
}

// Rule-based fallback recommendations
func (client *OpenSourceAIClient) getRuleBasedRecommendations(productName, category string) []string {
	recommendations := make([]string, 0)

	// Category-based recommendations
	switch strings.ToLower(category) {
	case "electronics", "elektronik":
		recommendations = []string{
			"Akıllı telefon aksesuarları",
			"Kablosuz kulaklık",
			"Powerbank",
			"Bluetooth hoparlör",
			"Akıllı saat",
		}
	case "clothing", "giyim":
		recommendations = []string{
			"Benzer renk ve stil ürünler",
			"Mevsimlik kıyafetler",
			"Aksesuar ürünleri",
			"Ayakkabı modelleri",
			"Çanta çeşitleri",
		}
	case "home", "ev":
		recommendations = []string{
			"Ev dekorasyon ürünleri",
			"Mutfak gereçleri",
			"Temizlik malzemeleri",
			"Bahçe ürünleri",
			"Ev tekstili",
		}
	default:
		recommendations = []string{
			"Popüler ürünler",
			"İndirimli ürünler",
			"Yeni çıkan ürünler",
			"En çok satan ürünler",
			"Benzer kategorideki ürünler",
		}
	}

	return recommendations
}

// Default response for fallback scenarios
func (client *OpenSourceAIClient) getDefaultResponse(prompt string) string {
	responses := []string{
		"Bu ürün için size özel önerilerimiz hazırlanıyor...",
		"Benzer ürünleri incelemenizi öneririz.",
		"Bu kategorideki popüler ürünlere göz atabilirsiniz.",
		"Size özel kampanyalar için bildirimleri açmanızı öneririz.",
	}

	return responses[rand.Intn(len(responses))]
}

// AIService provides AI-powered features for the marketplace
type AIService struct {
	repo           database.SimpleRepository
	productService *ProductService
	orderService   *OrderService
	osAIClient     *OpenSourceAIClient
}

// NewAIService creates a new AI service
func NewAIService(repo database.SimpleRepository, productService *ProductService, orderService *OrderService) *AIService {
	return &AIService{
		repo:           repo,
		productService: productService,
		orderService:   orderService,
		osAIClient:     NewOpenSourceAIClient(),
	}
}

// ProductRecommendation represents a product recommendation with score
type AIProductRecommendation struct {
	Product *models.Product `json:"product"`
	Score   float64         `json:"score"`
	Reason  string          `json:"reason"`
}

// AIPriceOptimization represents price optimization suggestions
type AIPriceOptimization struct {
	ProductID          int     `json:"product_id"`
	CurrentPrice       float64 `json:"current_price"`
	SuggestedPrice     float64 `json:"suggested_price"`
	PriceChange        float64 `json:"price_change"`
	PriceChangePercent float64 `json:"price_change_percent"`
	Confidence         float64 `json:"confidence"`
	Reasoning          string  `json:"reasoning"`
}

// CategoryPrediction represents AI-powered category prediction
type CategoryPrediction struct {
	CategoryID   int     `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Confidence   float64 `json:"confidence"`
}

// SearchResult represents enhanced search results with AI scoring
type SearchResult struct {
	Products      []*models.Product `json:"products"`
	TotalCount    int               `json:"total_count"`
	SearchQuery   string            `json:"search_query"`
	ProcessedTime time.Duration     `json:"processed_time"`
	Suggestions   []string          `json:"suggestions"`
}

// GetPersonalizedRecommendations returns personalized product recommendations for a user
func (s *AIService) GetPersonalizedRecommendations(userID int, limit int) ([]*AIProductRecommendation, error) {
	startTime := time.Now()

	// Get user's order history
	userOrders, err := s.orderService.GetOrdersByUser(userID, 50, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get user orders: %w", err)
	}

	// Get user's purchased product categories and brands
	categoryScores := make(map[string]float64)
	brandScores := make(map[string]float64)
	priceRange := struct{ min, max, avg float64 }{math.MaxFloat64, 0, 0}
	totalSpent := 0.0
	orderCount := 0

	for _, order := range userOrders {
		orderCount++
		totalSpent += order.TotalAmount

		// Analyze order items (this would need to be implemented in order service)
		// For now, we'll use a simplified approach
		if order.TotalAmount < priceRange.min {
			priceRange.min = order.TotalAmount
		}
		if order.TotalAmount > priceRange.max {
			priceRange.max = order.TotalAmount
		}
	}

	if orderCount > 0 {
		priceRange.avg = totalSpent / float64(orderCount)
	}

	// Get all available products
	allProducts, err := s.productService.GetAllProducts(1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	recommendations := make([]*AIProductRecommendation, 0)

	// Use open-source AI for enhanced recommendations
	var aiEnhancedCategories []string
	if len(allProducts) > 0 {
		// Get AI-powered category recommendations based on user history
		sampleProduct := allProducts[0]
		categoryName := fmt.Sprintf("category_%d", sampleProduct.CategoryID) // Use category ID as fallback
		aiCategories, err := s.osAIClient.GetHuggingFaceRecommendations(sampleProduct.Name, categoryName)
		if err == nil {
			aiEnhancedCategories = aiCategories
		}
	}

	// Score products based on user preferences and AI insights
	for _, product := range allProducts {
		if product.Status != "active" {
			continue
		}

		score := s.calculateRecommendationScore(&product, categoryScores, brandScores, priceRange)

		// Boost score if product matches AI recommendations
		for _, aiCategory := range aiEnhancedCategories {
			productCategoryName := fmt.Sprintf("category_%d", product.CategoryID)
			if strings.Contains(strings.ToLower(product.Name), strings.ToLower(aiCategory)) ||
				strings.Contains(strings.ToLower(productCategoryName), strings.ToLower(aiCategory)) {
				score += 0.2 // AI boost
				break
			}
		}

		reason := s.generateRecommendationReason(&product, score)

		if score > 0.3 { // Minimum threshold
			recommendations = append(recommendations, &AIProductRecommendation{
				Product: &product,
				Score:   score,
				Reason:  reason,
			})
		}
	}

	// Sort by score descending
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	// Limit results
	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}

	fmt.Printf("AI Recommendations generated in %v for user %d\n", time.Since(startTime), userID)
	return recommendations, nil
}

// calculateRecommendationScore calculates a recommendation score for a product
func (s *AIService) calculateRecommendationScore(product *models.Product, _, _ map[string]float64, priceRange struct{ min, max, avg float64 }) float64 {
	score := 0.0

	// Base score from product popularity (view count, rating, etc.)
	if product.ViewCount > 0 {
		score += math.Log10(float64(product.ViewCount)) * 0.1
	}

	// Price compatibility score
	if priceRange.avg > 0 {
		priceDiff := math.Abs(product.Price - priceRange.avg)
		priceScore := 1.0 - (priceDiff / (priceRange.max - priceRange.min + 1))
		score += priceScore * 0.3
	}

	// Category preference score (we'll need to get category name by ID)
	// For now, skip this feature until we implement category lookup

	// Brand preference score (not available in current model)
	// For now, skip this feature

	// Stock availability boost
	if product.Stock > 0 {
		score += 0.1
	}

	// Recent products get a small boost
	daysSinceCreated := time.Since(product.CreatedAt).Hours() / 24
	if daysSinceCreated < 30 {
		score += (30 - daysSinceCreated) / 300 // Max 0.1 boost for newest products
	}

	return math.Min(score, 1.0) // Cap at 1.0
}

// generateRecommendationReason generates a human-readable reason for the recommendation
func (s *AIService) generateRecommendationReason(product *models.Product, score float64) string {
	reasons := []string{}

	if score > 0.8 {
		reasons = append(reasons, "Mükemmel eşleşme")
	} else if score > 0.6 {
		reasons = append(reasons, "Yüksek uyumluluk")
	} else if score > 0.4 {
		reasons = append(reasons, "İlginizi çekebilir")
	}

	if product.ViewCount > 100 {
		reasons = append(reasons, "Popüler ürün")
	}

	if product.Stock > 0 {
		reasons = append(reasons, "Stokta mevcut")
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "Size özel öneri")
	}

	return strings.Join(reasons, ", ")
}

// OptimizeProductPricing provides AI-powered price optimization suggestions
func (s *AIService) OptimizeProductPricing(productID int) (*AIPriceOptimization, error) {
	product, err := s.productService.GetProductByID(productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Get similar products for price comparison
	similarProducts, err := s.getSimilarProducts(product, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get similar products: %w", err)
	}

	// Calculate market average price
	totalPrice := 0.0
	validProducts := 0
	for _, similar := range similarProducts {
		if similar.Price > 0 {
			totalPrice += similar.Price
			validProducts++
		}
	}

	if validProducts == 0 {
		return &AIPriceOptimization{
			ProductID:          productID,
			CurrentPrice:       product.Price,
			SuggestedPrice:     product.Price,
			PriceChange:        0,
			PriceChangePercent: 0,
			Confidence:         0.1,
			Reasoning:          "Yeterli karşılaştırma verisi bulunamadı",
		}, nil
	}

	marketAverage := totalPrice / float64(validProducts)

	// Calculate suggested price based on various factors
	suggestedPrice := s.calculateOptimalPrice(product, marketAverage, similarProducts)

	priceChange := suggestedPrice - product.Price
	priceChangePercent := (priceChange / product.Price) * 100

	confidence := s.calculatePriceConfidence(validProducts, product.ViewCount)
	reasoning := s.generatePriceReasoning(product.Price, suggestedPrice, marketAverage)

	return &AIPriceOptimization{
		ProductID:          productID,
		CurrentPrice:       product.Price,
		SuggestedPrice:     suggestedPrice,
		PriceChange:        priceChange,
		PriceChangePercent: priceChangePercent,
		Confidence:         confidence,
		Reasoning:          reasoning,
	}, nil
}

// getSimilarProducts finds products similar to the given product
func (s *AIService) getSimilarProducts(product *models.Product, limit int) ([]*models.Product, error) {
	// This is a simplified similarity calculation
	// In a real implementation, you might use more sophisticated ML algorithms

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

		// Category match (by ID)
		if p.CategoryID == product.CategoryID {
			score += 0.4
		}

		// Brand match (not available in current model, skip for now)

		// Price similarity
		priceDiff := math.Abs(p.Price - product.Price)
		maxPrice := math.Max(p.Price, product.Price)
		if maxPrice > 0 {
			priceScore := 1.0 - (priceDiff / maxPrice)
			score += priceScore * 0.3
		}

		if score > 0.2 { // Minimum similarity threshold
			scored = append(scored, struct {
				product *models.Product
				score   float64
			}{product: &p, score: score})
		}
	}

	// Sort by similarity score
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Extract products
	result := make([]*models.Product, 0, limit)
	for i, ps := range scored {
		if i >= limit {
			break
		}
		result = append(result, ps.product)
	}

	return result, nil
}

// calculateOptimalPrice calculates the optimal price based on market data
func (s *AIService) calculateOptimalPrice(product *models.Product, marketAverage float64, _ []*models.Product) float64 {
	// Base price on market average
	suggestedPrice := marketAverage

	// Adjust based on product characteristics
	if product.ViewCount > 100 {
		// Popular products can command higher prices
		suggestedPrice *= 1.05
	}

	if product.Stock < 10 {
		// Low stock can justify higher prices
		suggestedPrice *= 1.03
	}

	// Don't suggest extreme price changes
	maxIncrease := product.Price * 1.2 // Max 20% increase
	maxDecrease := product.Price * 0.8 // Max 20% decrease

	if suggestedPrice > maxIncrease {
		suggestedPrice = maxIncrease
	} else if suggestedPrice < maxDecrease {
		suggestedPrice = maxDecrease
	}

	// Round to reasonable precision
	return math.Round(suggestedPrice*100) / 100
}

// calculatePriceConfidence calculates confidence level for price suggestions
func (s *AIService) calculatePriceConfidence(similarProductCount, viewCount int) float64 {
	confidence := 0.0

	// More similar products = higher confidence
	confidence += math.Min(float64(similarProductCount)/10.0, 0.5)

	// More views = higher confidence in demand
	confidence += math.Min(float64(viewCount)/1000.0, 0.3)

	// Base confidence
	confidence += 0.2

	return math.Min(confidence, 1.0)
}

// generatePriceReasoning generates human-readable reasoning for price suggestions
func (s *AIService) generatePriceReasoning(currentPrice, suggestedPrice, marketAverage float64) string {
	if math.Abs(suggestedPrice-currentPrice) < 0.01 {
		return "Mevcut fiyat optimal seviyede"
	}

	if suggestedPrice > currentPrice {
		if suggestedPrice > marketAverage {
			return "Pazar ortalamasının üzerinde fiyatlandırma öneriliyor - ürün kalitesi ve popülaritesi bunu destekliyor"
		}
		return "Pazar koşullarına göre fiyat artışı öneriliyor"
	} else {
		if suggestedPrice < marketAverage {
			return "Rekabetçi fiyatlandırma için fiyat düşürülmesi öneriliyor"
		}
		return "Pazar ortalamasına yakın fiyatlandırma öneriliyor"
	}
}

// PredictProductCategory predicts the most suitable category for a product
func (s *AIService) PredictProductCategory(productName, description string) ([]*CategoryPrediction, error) {
	// Get all categories
	categories, err := s.productService.GetAllCategories()
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	predictions := make([]*CategoryPrediction, 0)
	text := strings.ToLower(productName + " " + description)

	for _, category := range categories {
		confidence := s.calculateCategoryConfidence(text, strings.ToLower(category.Name))

		if confidence > 0.1 { // Minimum confidence threshold
			predictions = append(predictions, &CategoryPrediction{
				CategoryID:   int(category.ID),
				CategoryName: category.Name,
				Confidence:   confidence,
			})
		}
	}

	// Sort by confidence descending
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].Confidence > predictions[j].Confidence
	})

	// Limit to top 5 predictions
	if len(predictions) > 5 {
		predictions = predictions[:5]
	}

	return predictions, nil
}

// calculateCategoryConfidence calculates confidence for category prediction
func (s *AIService) calculateCategoryConfidence(text, categoryName string) float64 {
	// Simple keyword matching approach
	// In a real implementation, you'd use more sophisticated NLP/ML

	keywords := strings.Fields(categoryName)
	matches := 0

	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			matches++
		}
	}

	if len(keywords) == 0 {
		return 0.0
	}

	return float64(matches) / float64(len(keywords))
}

// SmartSearch performs AI-enhanced product search
func (s *AIService) SmartSearch(query string, limit, offset int) (*SearchResult, error) {
	startTime := time.Now()

	// Get all products for searching
	allProducts, err := s.productService.GetAllProducts(2000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	// Score and filter products based on search query
	scored := make([]struct {
		product *models.Product
		score   float64
	}, 0)
	queryLower := strings.ToLower(query)
	queryWords := strings.Fields(queryLower)

	for _, product := range allProducts {
		if product.Status != "active" {
			continue
		}

		score := s.calculateSearchScore(&product, queryLower, queryWords)

		if score > 0 {
			scored = append(scored, struct {
				product *models.Product
				score   float64
			}{product: &product, score: score})
		}
	}

	// Sort by relevance score
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Apply pagination
	totalCount := len(scored)
	start := offset
	end := offset + limit

	if start > totalCount {
		start = totalCount
	}
	if end > totalCount {
		end = totalCount
	}

	results := make([]*models.Product, 0, end-start)
	for i := start; i < end; i++ {
		results = append(results, scored[i].product)
	}

	// Generate search suggestions
	suggestions := s.generateSearchSuggestions(query, scored)

	return &SearchResult{
		Products:      results,
		TotalCount:    totalCount,
		SearchQuery:   query,
		ProcessedTime: time.Since(startTime),
		Suggestions:   suggestions,
	}, nil
}

// calculateSearchScore calculates relevance score for search results
func (s *AIService) calculateSearchScore(product *models.Product, query string, queryWords []string) float64 {
	score := 0.0

	productText := strings.ToLower(product.Name + " " + product.Description + " " + product.Tags)

	// Exact phrase match gets highest score
	if strings.Contains(productText, query) {
		score += 1.0
	}

	// Individual word matches
	wordMatches := 0
	for _, word := range queryWords {
		if strings.Contains(productText, word) {
			wordMatches++
		}
	}

	if len(queryWords) > 0 {
		wordScore := float64(wordMatches) / float64(len(queryWords))
		score += wordScore * 0.8
	}

	// Title match bonus
	if strings.Contains(strings.ToLower(product.Name), query) {
		score += 0.5
	}

	// Tags match bonus
	if strings.Contains(strings.ToLower(product.Tags), query) {
		score += 0.3
	}

	// Popularity boost
	if product.ViewCount > 0 {
		popularityBoost := math.Log10(float64(product.ViewCount)) * 0.1
		score += math.Min(popularityBoost, 0.2)
	}

	// Stock availability
	if product.Stock > 0 {
		score += 0.1
	}

	return score
}

// generateSearchSuggestions generates search suggestions based on results
func (s *AIService) generateSearchSuggestions(query string, results []struct {
	product *models.Product
	score   float64
}) []string {
	suggestions := make([]string, 0)

	// Extract common tags from top results
	tagCount := make(map[string]int)

	maxResults := 20
	if len(results) < maxResults {
		maxResults = len(results)
	}

	for i := 0; i < maxResults; i++ {
		product := results[i].product

		// Extract individual tags
		if product.Tags != "" {
			tags := strings.Split(product.Tags, ",")
			for _, tag := range tags {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					tagCount[tag]++
				}
			}
		}
	}

	// Add tag suggestions
	for tag, count := range tagCount {
		if count >= 2 && !strings.Contains(strings.ToLower(query), strings.ToLower(tag)) {
			suggestions = append(suggestions, query+" "+tag)
		}
	}

	// Limit suggestions
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}

	return suggestions
}

// EnhancedSearch performs AI-enhanced search
func (s *AIService) EnhancedSearch(query string, userID int, filters map[string]interface{}) ([]models.Product, error) {
	// This would implement AI-enhanced search logic
	// For now, we'll use the basic product search and enhance it

	// Get products using filters
	products, err := s.productService.GetProductsWithFilters(filters, "relevance", "DESC", 50, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	// Apply AI ranking based on query relevance
	rankedProducts := s.rankProductsByRelevance(products, query)

	// Limit results
	if len(rankedProducts) > 20 {
		rankedProducts = rankedProducts[:20]
	}

	return rankedProducts, nil
}

// rankProductsByRelevance ranks products by relevance to search query
func (s *AIService) rankProductsByRelevance(products []models.Product, query string) []models.Product {
	if query == "" {
		return products
	}

	queryLower := strings.ToLower(query)
	queryWords := strings.Fields(queryLower)

	// Score each product
	type scoredProduct struct {
		product models.Product
		score   float64
	}

	var scored []scoredProduct

	for _, product := range products {
		score := 0.0

		// Check name relevance
		nameLower := strings.ToLower(product.Name)
		for _, word := range queryWords {
			if strings.Contains(nameLower, word) {
				score += 2.0
			}
		}

		// Check description relevance
		descLower := strings.ToLower(product.Description)
		for _, word := range queryWords {
			if strings.Contains(descLower, word) {
				score += 1.0
			}
		}

		// Check tags relevance (since we don't have direct category name)
		tagsLower := strings.ToLower(product.Tags)
		for _, word := range queryWords {
			if strings.Contains(tagsLower, word) {
				score += 1.5
			}
		}

		scored = append(scored, scoredProduct{product: product, score: score})
	}

	// Sort by score
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Extract products
	var result []models.Product
	for _, sp := range scored {
		result = append(result, sp.product)
	}

	return result
}

// GetPriceOptimizations provides AI-powered price optimization suggestions
func (s *AIService) GetPriceOptimizations(userID int) (map[string]interface{}, error) {
	// This would implement AI-powered price optimization analysis
	// For now, we'll return sample data

	optimizations := map[string]interface{}{
		"recommendations": []map[string]interface{}{
			{
				"product_id":      1,
				"current_price":   99.99,
				"suggested_price": 89.99,
				"reason":          "Market analysis suggests 10% price reduction could increase sales by 25%",
				"confidence":      0.85,
				"expected_impact": map[string]interface{}{
					"sales_increase": 25.0,
					"revenue_change": 12.5,
				},
			},
			{
				"product_id":      2,
				"current_price":   149.99,
				"suggested_price": 159.99,
				"reason":          "Demand is high and competitor prices are higher",
				"confidence":      0.92,
				"expected_impact": map[string]interface{}{
					"sales_decrease": -5.0,
					"revenue_change": 15.0,
				},
			},
		},
		"market_trends": map[string]interface{}{
			"category_average": 125.50,
			"price_elasticity": -0.8,
			"demand_forecast":  "increasing",
		},
		"competitor_analysis": map[string]interface{}{
			"average_price": 130.00,
			"min_price":     85.00,
			"max_price":     200.00,
		},
	}

	return optimizations, nil
}

// GetAIDashboardStats returns statistics for AI dashboard
func (s *AIService) GetAIDashboardStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count total recommendations made
	var recommendationCount int64
	err := s.repo.QueryRow("SELECT COUNT(*) FROM ai_recommendations").Scan(&recommendationCount)
	if err != nil {
		recommendationCount = 0
	}
	stats["RecommendationsCount"] = recommendationCount

	// Count total search queries
	var searchQueryCount int64
	err = s.repo.QueryRow("SELECT COUNT(*) FROM search_logs WHERE created_at >= datetime('now', '-30 days')").Scan(&searchQueryCount)
	if err != nil {
		searchQueryCount = 0
	}
	stats["SearchQueriesCount"] = searchQueryCount

	// Count price optimizations
	var priceOptimizationCount int64
	err = s.repo.QueryRow("SELECT COUNT(*) FROM price_optimizations WHERE created_at >= datetime('now', '-30 days')").Scan(&priceOptimizationCount)
	if err != nil {
		priceOptimizationCount = 0
	}
	stats["PriceOptimizationsCount"] = priceOptimizationCount

	// AI model usage stats
	var modelUsageCount int64
	err = s.repo.QueryRow("SELECT COUNT(*) FROM ai_model_usage WHERE created_at >= datetime('now', '-7 days')").Scan(&modelUsageCount)
	if err != nil {
		modelUsageCount = 0
	}
	stats["ModelUsageCount"] = modelUsageCount

	return stats, nil
}
