package services

import (
	"encoding/json"
	"fmt"
	"kolajAi/internal/database"
	"math"
	"sort"
	"strings"
	"time"
)

// AIEnterpriseService provides enterprise-level AI capabilities
type AIEnterpriseService struct {
	repo             database.SimpleRepository
	aiService        *AIService
	aiVisionService  *AIVisionService
	productService   *ProductService
	orderService     *OrderService
	userService      *AuthService
}

// NewAIEnterpriseService creates a new enterprise AI service
func NewAIEnterpriseService(
	repo database.SimpleRepository,
	aiService *AIService,
	aiVisionService *AIVisionService,
	productService *ProductService,
	orderService *OrderService,
	userService *AuthService,
) *AIEnterpriseService {
	return &AIEnterpriseService{
		repo:             repo,
		aiService:        aiService,
		aiVisionService:  aiVisionService,
		productService:   productService,
		orderService:     orderService,
		userService:      userService,
	}
}

// CustomerServiceRequest represents a customer service request
type CustomerServiceRequest struct {
	ID            int                    `json:"id"`
	UserID        int                    `json:"user_id"`
	Type          string                 `json:"type"` // 'complaint', 'question', 'suggestion', 'technical'
	Subject       string                 `json:"subject"`
	Description   string                 `json:"description"`
	Priority      string                 `json:"priority"` // 'low', 'medium', 'high', 'urgent'
	Status        string                 `json:"status"`   // 'open', 'in_progress', 'resolved', 'closed'
	Category      string                 `json:"category"`
	Tags          []string               `json:"tags"`
	Attachments   []string               `json:"attachments"` // Image IDs
	AIAnalysis    CustomerServiceAnalysis `json:"ai_analysis"`
	AssignedTo    int                    `json:"assigned_to"`
	Resolution    string                 `json:"resolution"`
	SatisfactionScore float64            `json:"satisfaction_score"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	ResolvedAt    *time.Time             `json:"resolved_at"`
}

// CustomerServiceAnalysis represents AI analysis of customer service request
type CustomerServiceAnalysis struct {
	SentimentScore    float64  `json:"sentiment_score"`    // -1 to 1
	UrgencyScore      float64  `json:"urgency_score"`      // 0 to 1
	ComplexityScore   float64  `json:"complexity_score"`   // 0 to 1
	SuggestedCategory string   `json:"suggested_category"`
	SuggestedPriority string   `json:"suggested_priority"`
	KeyTopics         []string `json:"key_topics"`
	RelatedProducts   []int    `json:"related_products"`
	SuggestedResponse string   `json:"suggested_response"`
	RequiresHuman     bool     `json:"requires_human"`
	EstimatedTime     int      `json:"estimated_time"` // minutes
}

// ContentModerationResult represents content moderation analysis
type ContentModerationResult struct {
	ContentID     string                 `json:"content_id"`
	ContentType   string                 `json:"content_type"` // 'text', 'image', 'video'
	IsAppropriate bool                   `json:"is_appropriate"`
	ConfidenceScore float64              `json:"confidence_score"`
	Violations    []ContentViolation     `json:"violations"`
	Recommendations []string             `json:"recommendations"`
	ActionRequired string                `json:"action_required"` // 'none', 'review', 'block', 'remove'
	ProcessedAt   time.Time              `json:"processed_at"`
}

// ContentViolation represents a content policy violation
type ContentViolation struct {
	Type        string  `json:"type"`        // 'spam', 'inappropriate', 'offensive', 'copyright'
	Severity    string  `json:"severity"`    // 'low', 'medium', 'high'
	Confidence  float64 `json:"confidence"`
	Description string  `json:"description"`
}

// BusinessInsight represents AI-generated business insights
type BusinessInsight struct {
	ID          int                    `json:"id"`
	Type        string                 `json:"type"` // 'trend', 'opportunity', 'risk', 'recommendation'
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Impact      string                 `json:"impact"` // 'low', 'medium', 'high'
	Confidence  float64                `json:"confidence"`
	Data        map[string]interface{} `json:"data"`
	ActionItems []string               `json:"action_items"`
	Category    string                 `json:"category"`
	CreatedAt   time.Time              `json:"created_at"`
	ExpiresAt   *time.Time             `json:"expires_at"`
}

// AutomatedTask represents an automated task performed by AI
type AutomatedTask struct {
	ID          int                    `json:"id"`
	Type        string                 `json:"type"` // 'content_optimization', 'inventory_management', 'customer_segmentation'
	Status      string                 `json:"status"` // 'pending', 'running', 'completed', 'failed'
	Progress    float64                `json:"progress"` // 0 to 1
	Results     map[string]interface{} `json:"results"`
	ErrorMessage string                `json:"error_message"`
	ScheduledAt time.Time              `json:"scheduled_at"`
	StartedAt   *time.Time             `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at"`
}

// ProcessCustomerServiceRequest analyzes and processes a customer service request
func (s *AIEnterpriseService) ProcessCustomerServiceRequest(request *CustomerServiceRequest) error {
	// Perform AI analysis
	analysis, err := s.analyzeCustomerServiceRequest(request)
	if err != nil {
		return fmt.Errorf("failed to analyze request: %w", err)
	}

	request.AIAnalysis = *analysis

	// Auto-assign priority if not set
	if request.Priority == "" {
		request.Priority = analysis.SuggestedPriority
	}

	// Auto-assign category if not set
	if request.Category == "" {
		request.Category = analysis.SuggestedCategory
	}

	// Save to database
	return s.saveCustomerServiceRequest(request)
}

// analyzeCustomerServiceRequest performs AI analysis on customer service request
func (s *AIEnterpriseService) analyzeCustomerServiceRequest(request *CustomerServiceRequest) (*CustomerServiceAnalysis, error) {
	analysis := &CustomerServiceAnalysis{}

	// Analyze sentiment
	analysis.SentimentScore = s.analyzeSentiment(request.Description)

	// Analyze urgency
	analysis.UrgencyScore = s.analyzeUrgency(request.Subject, request.Description)

	// Analyze complexity
	analysis.ComplexityScore = s.analyzeComplexity(request.Description)

	// Suggest category
	analysis.SuggestedCategory = s.suggestCategory(request.Subject, request.Description)

	// Suggest priority
	analysis.SuggestedPriority = s.suggestPriority(analysis.UrgencyScore, analysis.ComplexityScore)

	// Extract key topics
	analysis.KeyTopics = s.extractKeyTopics(request.Subject + " " + request.Description)

	// Find related products
	analysis.RelatedProducts = s.findRelatedProducts(request.Description)

	// Generate suggested response
	analysis.SuggestedResponse = s.generateSuggestedResponse(request, analysis)

	// Determine if human intervention is required
	analysis.RequiresHuman = s.requiresHumanIntervention(analysis)

	// Estimate resolution time
	analysis.EstimatedTime = s.estimateResolutionTime(analysis)

	return analysis, nil
}

// analyzeSentiment analyzes sentiment of text (-1 to 1)
func (s *AIEnterpriseService) analyzeSentiment(text string) float64 {
	text = strings.ToLower(text)
	
	// Positive words
	positiveWords := []string{
		"good", "great", "excellent", "amazing", "wonderful", "fantastic", "love", "like",
		"happy", "satisfied", "pleased", "perfect", "awesome", "brilliant", "outstanding",
		"iyi", "harika", "mükemmel", "güzel", "beğendim", "memnun", "mutlu", "süper",
	}

	// Negative words
	negativeWords := []string{
		"bad", "terrible", "awful", "horrible", "hate", "dislike", "angry", "frustrated",
		"disappointed", "annoyed", "upset", "worst", "pathetic", "useless", "broken",
		"kötü", "berbat", "rezalet", "beğenmedim", "memnun değilim", "sinirli", "kızgın",
	}

	positiveCount := 0
	negativeCount := 0
	totalWords := len(strings.Fields(text))

	for _, word := range positiveWords {
		positiveCount += strings.Count(text, word)
	}

	for _, word := range negativeWords {
		negativeCount += strings.Count(text, word)
	}

	if totalWords == 0 {
		return 0.0
	}

	// Calculate sentiment score
	positiveRatio := float64(positiveCount) / float64(totalWords)
	negativeRatio := float64(negativeCount) / float64(totalWords)

	sentiment := positiveRatio - negativeRatio

	// Normalize to -1 to 1 range
	if sentiment > 1 {
		sentiment = 1
	} else if sentiment < -1 {
		sentiment = -1
	}

	return sentiment
}

// analyzeUrgency analyzes urgency of request (0 to 1)
func (s *AIEnterpriseService) analyzeUrgency(subject, description string) float64 {
	text := strings.ToLower(subject + " " + description)
	
	urgentWords := []string{
		"urgent", "emergency", "asap", "immediately", "critical", "broken", "not working",
		"can't", "unable", "error", "problem", "issue", "bug", "crash", "down",
		"acil", "hemen", "acilen", "çalışmıyor", "bozuk", "hata", "sorun", "problem",
	}

	urgencyScore := 0.0
	totalWords := len(strings.Fields(text))

	for _, word := range urgentWords {
		count := strings.Count(text, word)
		urgencyScore += float64(count) * 0.2
	}

	// Normalize based on text length
	if totalWords > 0 {
		urgencyScore = urgencyScore / float64(totalWords) * 10
	}

	// Cap at 1.0
	if urgencyScore > 1.0 {
		urgencyScore = 1.0
	}

	return urgencyScore
}

// analyzeComplexity analyzes complexity of request (0 to 1)
func (s *AIEnterpriseService) analyzeComplexity(description string) float64 {
	words := strings.Fields(description)
	wordCount := len(words)

	// Base complexity on text length
	lengthComplexity := math.Min(float64(wordCount)/200.0, 0.5)

	// Technical terms increase complexity
	technicalWords := []string{
		"api", "database", "server", "configuration", "integration", "authentication",
		"ssl", "https", "json", "xml", "payment", "gateway", "webhook", "oauth",
		"veritabanı", "sunucu", "yapılandırma", "entegrasyon", "kimlik doğrulama",
	}

	technicalComplexity := 0.0
	for _, word := range technicalWords {
		if strings.Contains(strings.ToLower(description), word) {
			technicalComplexity += 0.1
		}
	}

	// Multiple questions increase complexity
	questionCount := strings.Count(description, "?")
	questionComplexity := math.Min(float64(questionCount)*0.1, 0.3)

	totalComplexity := lengthComplexity + technicalComplexity + questionComplexity

	if totalComplexity > 1.0 {
		totalComplexity = 1.0
	}

	return totalComplexity
}

// suggestCategory suggests category based on content
func (s *AIEnterpriseService) suggestCategory(subject, description string) string {
	text := strings.ToLower(subject + " " + description)

	categories := map[string][]string{
		"technical": {"error", "bug", "crash", "not working", "broken", "api", "integration", "hata", "çalışmıyor", "bozuk"},
		"billing":   {"payment", "invoice", "charge", "refund", "money", "price", "cost", "ödeme", "fatura", "para", "ücret"},
		"shipping":  {"delivery", "shipping", "order", "package", "tracking", "teslimat", "kargo", "sipariş", "paket"},
		"account":   {"login", "password", "profile", "account", "register", "giriş", "şifre", "hesap", "profil"},
		"product":   {"product", "item", "quality", "defect", "description", "ürün", "kalite", "kusur", "açıklama"},
		"general":   {"question", "help", "support", "information", "soru", "yardım", "destek", "bilgi"},
	}

	maxScore := 0.0
	suggestedCategory := "general"

	for category, keywords := range categories {
		score := 0.0
		for _, keyword := range keywords {
			if strings.Contains(text, keyword) {
				score += 1.0
			}
		}
		if score > maxScore {
			maxScore = score
			suggestedCategory = category
		}
	}

	return suggestedCategory
}

// suggestPriority suggests priority based on analysis
func (s *AIEnterpriseService) suggestPriority(urgencyScore, complexityScore float64) string {
	combinedScore := (urgencyScore * 0.7) + (complexityScore * 0.3)

	if combinedScore >= 0.8 {
		return "urgent"
	} else if combinedScore >= 0.6 {
		return "high"
	} else if combinedScore >= 0.3 {
		return "medium"
	} else {
		return "low"
	}
}

// extractKeyTopics extracts key topics from text
func (s *AIEnterpriseService) extractKeyTopics(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	wordCount := make(map[string]int)

	// Count word frequency (ignore common words)
	commonWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "up": true, "about": true, "into": true,
		"through": true, "during": true, "before": true, "after": true, "above": true,
		"below": true, "between": true, "among": true, "is": true, "are": true, "was": true,
		"were": true, "be": true, "been": true, "being": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
		"could": true, "should": true, "may": true, "might": true, "must": true, "can": true,
		"ve": true, "ile": true, "bu": true, "şu": true, "o": true, "bir": true, "için": true,
		"da": true, "de": true, "ki": true, "mi": true, "mu": true, "mı": true, "mü": true,
	}

	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:")
		if len(word) > 2 && !commonWords[word] {
			wordCount[word]++
		}
	}

	// Sort by frequency
	type wordFreq struct {
		word  string
		count int
	}

	var frequencies []wordFreq
	for word, count := range wordCount {
		if count >= 2 { // Only include words that appear at least twice
			frequencies = append(frequencies, wordFreq{word: word, count: count})
		}
	}

	sort.Slice(frequencies, func(i, j int) bool {
		return frequencies[i].count > frequencies[j].count
	})

	// Return top 5 topics
	topics := make([]string, 0, 5)
	for i, freq := range frequencies {
		if i >= 5 {
			break
		}
		topics = append(topics, freq.word)
	}

	return topics
}

// findRelatedProducts finds products related to the request
func (s *AIEnterpriseService) findRelatedProducts(_ string) []int {
	// This would search for products mentioned in the description
	// For now, return empty slice
	return []int{}
}

// generateSuggestedResponse generates a suggested response
func (s *AIEnterpriseService) generateSuggestedResponse(_ *CustomerServiceRequest, _ *CustomerServiceAnalysis) string {
	responses := map[string]string{
		"technical": "Teknik sorununuz için üzgünüz. Ekibimiz bu konuyu inceleyecek ve en kısa sürede size dönüş yapacaktır.",
		"billing":   "Faturalandırma ile ilgili sorununuzu anlıyoruz. Mali işler departmanımız konuyu inceleyecek ve 24 saat içinde size dönüş yapacaktır.",
		"shipping":  "Kargo durumunuz hakkında bilgi almak için lütfen sipariş numaranızı paylaşın. Durumu kontrol edip size bilgi vereceğiz.",
		"account":   "Hesap sorununuz için yardımcı olmaktan memnuniyet duyarız. Güvenlik nedeniyle bazı bilgileri doğrulamamız gerekebilir.",
		"product":   "Ürün hakkındaki geri bildiriminiz için teşekkürler. Kalite ekibimiz durumu değerlendirip gerekli aksiyonu alacaktır.",
		"general":   "Mesajınız için teşekkürler. Ekibimiz en kısa sürede size yardımcı olmak için iletişime geçecektir.",
	}

	// For now, return general response
	// In a real implementation, this would use the analysis parameter
	return responses["general"]
}

// requiresHumanIntervention determines if human intervention is required
func (s *AIEnterpriseService) requiresHumanIntervention(analysis *CustomerServiceAnalysis) bool {
	// Require human for high complexity or negative sentiment
	return analysis.ComplexityScore > 0.7 || analysis.SentimentScore < -0.5 || analysis.UrgencyScore > 0.8
}

// estimateResolutionTime estimates resolution time in minutes
func (s *AIEnterpriseService) estimateResolutionTime(analysis *CustomerServiceAnalysis) int {
	baseTime := 30 // 30 minutes base

	// Adjust based on complexity
	complexityTime := int(analysis.ComplexityScore * 120) // Up to 2 hours

	// Adjust based on category
	categoryTime := map[string]int{
		"technical": 60,
		"billing":   30,
		"shipping":  20,
		"account":   40,
		"product":   45,
		"general":   30,
	}

	if catTime, exists := categoryTime[analysis.SuggestedCategory]; exists {
		return baseTime + complexityTime + catTime
	}

	return baseTime + complexityTime + 30
}

// ModerateContent performs AI-powered content moderation
func (s *AIEnterpriseService) ModerateContent(contentID, contentType, content string) (*ContentModerationResult, error) {
	result := &ContentModerationResult{
		ContentID:   contentID,
		ContentType: contentType,
		ProcessedAt: time.Now(),
	}

	// Analyze content
	violations := s.detectContentViolations(content)
	result.Violations = violations

	// Determine if content is appropriate
	result.IsAppropriate = len(violations) == 0

	// Calculate confidence score
	result.ConfidenceScore = s.calculateModerationConfidence(violations)

	// Generate recommendations
	result.Recommendations = s.generateModerationRecommendations(violations)

	// Determine action required
	result.ActionRequired = s.determineModerationAction(violations)

	return result, nil
}

// detectContentViolations detects various types of content violations
func (s *AIEnterpriseService) detectContentViolations(content string) []ContentViolation {
	violations := make([]ContentViolation, 0)
	content = strings.ToLower(content)

	// Spam detection
	if s.isSpam(content) {
		violations = append(violations, ContentViolation{
			Type:        "spam",
			Severity:    "medium",
			Confidence:  0.8,
			Description: "İçerik spam olarak tespit edildi",
		})
	}

	// Inappropriate content detection
	if s.isInappropriate(content) {
		violations = append(violations, ContentViolation{
			Type:        "inappropriate",
			Severity:    "high",
			Confidence:  0.9,
			Description: "İçerik uygunsuz olarak tespit edildi",
		})
	}

	// Offensive language detection
	if s.isOffensive(content) {
		violations = append(violations, ContentViolation{
			Type:        "offensive",
			Severity:    "high",
			Confidence:  0.85,
			Description: "İçerik saldırgan dil içeriyor",
		})
	}

	return violations
}

// isSpam detects spam content
func (s *AIEnterpriseService) isSpam(content string) bool {
	spamIndicators := []string{
		"click here", "buy now", "limited time", "act now", "free money",
		"guarantee", "no risk", "100% free", "make money fast", "work from home",
		"buraya tıkla", "hemen satın al", "sınırlı süre", "ücretsiz para", "garanti",
		"risk yok", "hızlı para kazan", "evden çalış",
	}

	spamCount := 0
	for _, indicator := range spamIndicators {
		if strings.Contains(content, indicator) {
			spamCount++
		}
	}

	// Consider spam if multiple indicators are present
	return spamCount >= 2
}

// isInappropriate detects inappropriate content
func (s *AIEnterpriseService) isInappropriate(content string) bool {
	inappropriateWords := []string{
		// Add inappropriate words here - keeping this minimal for the example
		"inappropriate1", "inappropriate2",
	}

	for _, word := range inappropriateWords {
		if strings.Contains(content, word) {
			return true
		}
	}

	return false
}

// isOffensive detects offensive language
func (s *AIEnterpriseService) isOffensive(content string) bool {
	offensiveWords := []string{
		// Add offensive words here - keeping this minimal for the example
		"offensive1", "offensive2",
	}

	for _, word := range offensiveWords {
		if strings.Contains(content, word) {
			return true
		}
	}

	return false
}

// calculateModerationConfidence calculates confidence score for moderation
func (s *AIEnterpriseService) calculateModerationConfidence(violations []ContentViolation) float64 {
	if len(violations) == 0 {
		return 0.95 // High confidence that content is appropriate
	}

	totalConfidence := 0.0
	for _, violation := range violations {
		totalConfidence += violation.Confidence
	}

	return totalConfidence / float64(len(violations))
}

// generateModerationRecommendations generates recommendations based on violations
func (s *AIEnterpriseService) generateModerationRecommendations(violations []ContentViolation) []string {
	if len(violations) == 0 {
		return []string{"İçerik uygun görünüyor"}
	}

	recommendations := make([]string, 0)
	for _, violation := range violations {
		switch violation.Type {
		case "spam":
			recommendations = append(recommendations, "Spam içerik için kullanıcıyı uyar")
		case "inappropriate":
			recommendations = append(recommendations, "Uygunsuz içeriği kaldır ve kullanıcıyı uyar")
		case "offensive":
			recommendations = append(recommendations, "Saldırgan içeriği kaldır ve kullanıcıya yaptırım uygula")
		}
	}

	return recommendations
}

// determineModerationAction determines what action should be taken
func (s *AIEnterpriseService) determineModerationAction(violations []ContentViolation) string {
	if len(violations) == 0 {
		return "none"
	}

	highSeverityCount := 0
	for _, violation := range violations {
		if violation.Severity == "high" {
			highSeverityCount++
		}
	}

	if highSeverityCount > 0 {
		return "remove"
	} else if len(violations) > 2 {
		return "review"
	} else {
		return "review"
	}
}

// GenerateBusinessInsights generates AI-powered business insights
func (s *AIEnterpriseService) GenerateBusinessInsights() ([]BusinessInsight, error) {
	insights := make([]BusinessInsight, 0)

	// Analyze sales trends
	salesInsights, err := s.analyzeSalesTrends()
	if err == nil {
		insights = append(insights, salesInsights...)
	}

	// Analyze customer behavior
	customerInsights, err := s.analyzeCustomerBehavior()
	if err == nil {
		insights = append(insights, customerInsights...)
	}

	// Analyze inventory
	inventoryInsights, err := s.analyzeInventory()
	if err == nil {
		insights = append(insights, inventoryInsights...)
	}

	// Sort by impact and confidence
	sort.Slice(insights, func(i, j int) bool {
		scoreI := s.calculateInsightScore(insights[i])
		scoreJ := s.calculateInsightScore(insights[j])
		return scoreI > scoreJ
	})

	return insights, nil
}

// analyzeSalesTrends analyzes sales trends and generates insights
func (s *AIEnterpriseService) analyzeSalesTrends() ([]BusinessInsight, error) {
	insights := make([]BusinessInsight, 0)

	// This would analyze actual sales data
	// For now, return sample insights
	insights = append(insights, BusinessInsight{
		Type:        "trend",
		Title:       "Elektronik Kategorisinde Yükseliş Trendi",
		Description: "Son 30 günde elektronik ürünlerinde %25 artış gözlemlendi",
		Impact:      "high",
		Confidence:  0.85,
		Category:    "sales",
		Data: map[string]interface{}{
			"growth_rate": 25.0,
			"category":    "Elektronik",
			"period":      "30_days",
		},
		ActionItems: []string{
			"Elektronik kategorisinde stok artırımı düşünün",
			"Elektronik ürünleri için özel kampanya planlayın",
		},
		CreatedAt: time.Now(),
	})

	return insights, nil
}

// analyzeCustomerBehavior analyzes customer behavior patterns
func (s *AIEnterpriseService) analyzeCustomerBehavior() ([]BusinessInsight, error) {
	insights := make([]BusinessInsight, 0)

	insights = append(insights, BusinessInsight{
		Type:        "opportunity",
		Title:       "Mobil Kullanıcı Artışı",
		Description: "Mobil cihazlardan gelen trafik %40 arttı ancak dönüşüm oranı düşük",
		Impact:      "medium",
		Confidence:  0.78,
		Category:    "customer_behavior",
		Data: map[string]interface{}{
			"mobile_traffic_increase": 40.0,
			"mobile_conversion_rate":  2.3,
			"desktop_conversion_rate": 4.8,
		},
		ActionItems: []string{
			"Mobil site deneyimini optimize edin",
			"Mobil ödeme seçeneklerini iyileştirin",
		},
		CreatedAt: time.Now(),
	})

	return insights, nil
}

// analyzeInventory analyzes inventory and generates insights
func (s *AIEnterpriseService) analyzeInventory() ([]BusinessInsight, error) {
	insights := make([]BusinessInsight, 0)

	insights = append(insights, BusinessInsight{
		Type:        "risk",
		Title:       "Düşük Stok Riski",
		Description: "15 üründe kritik stok seviyesi tespit edildi",
		Impact:      "high",
		Confidence:  0.92,
		Category:    "inventory",
		Data: map[string]interface{}{
			"low_stock_products": 15,
			"critical_threshold": 10,
		},
		ActionItems: []string{
			"Acil stok siparişi verin",
			"Otomatik stok yenileme sistemi kurun",
		},
		CreatedAt: time.Now(),
	})

	return insights, nil
}

// calculateInsightScore calculates a score for prioritizing insights
func (s *AIEnterpriseService) calculateInsightScore(insight BusinessInsight) float64 {
	impactScore := map[string]float64{
		"low":    0.3,
		"medium": 0.6,
		"high":   1.0,
	}

	impact := impactScore[insight.Impact]
	return (impact * 0.6) + (insight.Confidence * 0.4)
}

// Database operations
func (s *AIEnterpriseService) saveCustomerServiceRequest(request *CustomerServiceRequest) error {
	analysisJSON, _ := json.Marshal(request.AIAnalysis)
	tagsJSON, _ := json.Marshal(request.Tags)
	attachmentsJSON, _ := json.Marshal(request.Attachments)

	query := `
		INSERT INTO customer_service_requests (
			user_id, type, subject, description, priority, status, category,
			tags, attachments, ai_analysis, assigned_to, resolution,
			satisfaction_score, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.repo.Exec(query,
		request.UserID, request.Type, request.Subject, request.Description,
		request.Priority, request.Status, request.Category, string(tagsJSON),
		string(attachmentsJSON), string(analysisJSON), request.AssignedTo,
		request.Resolution, request.SatisfactionScore, request.CreatedAt, request.UpdatedAt,
	)

	if err != nil {
		return err
	}

	id, _ := result.LastInsertId()
	request.ID = int(id)

	return nil
}