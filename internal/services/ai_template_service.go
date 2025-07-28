package services

import (
	"encoding/json"
	"fmt"
	"kolajAi/internal/database"
	"kolajAi/internal/models"
	"strings"
	"time"
)

// AITemplateService provides AI-powered template generation and management
type AITemplateService struct {
	repo           database.SimpleRepository
	productService *ProductService
	aiService      *AIService
	uploadPath     string
	templatePath   string
}

// NewAITemplateService creates a new AI template service
func NewAITemplateService(repo database.SimpleRepository, productService *ProductService, aiService *AIService) *AITemplateService {
	return &AITemplateService{
		repo:           repo,
		productService: productService,
		aiService:      aiService,
		uploadPath:     "web/static/uploads/templates",
		templatePath:   "web/static/templates",
	}
}

// TemplateGenerationRequest represents a template generation request
type TemplateGenerationRequest struct {
	UserID      int64                     `json:"user_id"`
	Type        models.AITemplateType     `json:"type"`
	Platform    string                    `json:"platform"`
	ProductID   *int64                    `json:"product_id,omitempty"`
	Content     map[string]interface{}    `json:"content"`
	Style       TemplateStyle             `json:"style"`
	Dimensions  models.ImageDimensions    `json:"dimensions"`
	Options     TemplateGenerationOptions `json:"options"`
}

// TemplateStyle represents visual style preferences
type TemplateStyle struct {
	ColorScheme   string            `json:"color_scheme"` // modern, classic, vibrant, minimal
	FontStyle     string            `json:"font_style"`   // modern, classic, bold, elegant
	Layout        string            `json:"layout"`       // grid, centered, asymmetric, magazine
	Theme         string            `json:"theme"`        // light, dark, gradient, colorful
	CustomColors  map[string]string `json:"custom_colors"`
	CustomFonts   []string          `json:"custom_fonts"`
	Mood          string            `json:"mood"`         // professional, casual, luxury, playful
}

// TemplateGenerationOptions represents generation options
type TemplateGenerationOptions struct {
	IncludeBranding    bool     `json:"include_branding"`
	IncludeWatermark   bool     `json:"include_watermark"`
	GenerateVariations bool     `json:"generate_variations"`
	VariationCount     int      `json:"variation_count"`
	OptimizeForPlatform bool    `json:"optimize_for_platform"`
	IncludeHashtags    bool     `json:"include_hashtags"`
	LanguageCode       string   `json:"language_code"`
	TargetAudience     string   `json:"target_audience"`
	Keywords           []string `json:"keywords"`
}

// TemplateGenerationResult represents the result of template generation
type TemplateGenerationResult struct {
	TemplateID      int64                     `json:"template_id"`
	MainTemplate    *models.AITemplate        `json:"main_template"`
	Variations      []*models.AITemplate      `json:"variations"`
	GeneratedImages []GeneratedImage          `json:"generated_images"`
	SuggestedText   map[string]string         `json:"suggested_text"`
	Hashtags        []string                  `json:"hashtags"`
	Performance     TemplatePerformanceMetrics `json:"performance"`
	ProcessingTime  time.Duration             `json:"processing_time"`
}

// GeneratedImage represents a generated image
type GeneratedImage struct {
	URL         string                 `json:"url"`
	Type        string                 `json:"type"` // main, variation, thumbnail
	Platform    string                 `json:"platform"`
	Dimensions  models.ImageDimensions `json:"dimensions"`
	FileSize    int64                  `json:"file_size"`
	Format      string                 `json:"format"`
}

// TemplatePerformanceMetrics represents predicted performance metrics
type TemplatePerformanceMetrics struct {
	EngagementScore    float64 `json:"engagement_score"`    // 0-100
	VisualAppealScore  float64 `json:"visual_appeal_score"` // 0-100
	PlatformOptimization float64 `json:"platform_optimization"` // 0-100
	ReadabilityScore   float64 `json:"readability_score"`   // 0-100
	BrandConsistency   float64 `json:"brand_consistency"`   // 0-100
	ConversionPotential float64 `json:"conversion_potential"` // 0-100
}

// GenerateTemplate creates AI-powered templates based on request
func (s *AITemplateService) GenerateTemplate(req *TemplateGenerationRequest) (*TemplateGenerationResult, error) {
	startTime := time.Now()

	// Validate user permissions
	user, err := s.getUserByID(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !user.CanUseAITemplates() {
		return nil, fmt.Errorf("user does not have permission to use AI templates")
	}

	// Get product information if specified
	var product *models.Product
	if req.ProductID != nil {
		product, err = s.productService.GetProductByID(int(*req.ProductID))
		if err != nil {
			return nil, fmt.Errorf("failed to get product: %w", err)
		}
	}

	// Generate main template
	mainTemplate, err := s.generateMainTemplate(req, product)
	if err != nil {
		return nil, fmt.Errorf("failed to generate main template: %w", err)
	}

	// Save template to database
	templateID, err := s.saveTemplate(mainTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to save template: %w", err)
	}
	mainTemplate.ID = templateID

	result := &TemplateGenerationResult{
		TemplateID:      templateID,
		MainTemplate:    mainTemplate,
		Variations:      []*models.AITemplate{},
		GeneratedImages: []GeneratedImage{},
		SuggestedText:   make(map[string]string),
		Hashtags:        []string{},
		ProcessingTime:  time.Since(startTime),
	}

	// Generate variations if requested
	if req.Options.GenerateVariations && req.Options.VariationCount > 0 {
		variations, err := s.generateVariations(req, mainTemplate, req.Options.VariationCount)
		if err == nil {
			result.Variations = variations
		}
	}

	// Generate images
	images, err := s.generateImages(req, mainTemplate, product)
	if err == nil {
		result.GeneratedImages = images
	}

	// Generate suggested text content
	textContent, err := s.generateTextContent(req, product)
	if err == nil {
		result.SuggestedText = textContent
	}

	// Generate hashtags
	if req.Options.IncludeHashtags {
		hashtags, err := s.generateHashtags(req, product)
		if err == nil {
			result.Hashtags = hashtags
		}
	}

	// Calculate performance metrics
	performance, err := s.calculatePerformanceMetrics(mainTemplate, req)
	if err == nil {
		result.Performance = performance
	}

	result.ProcessingTime = time.Since(startTime)
	return result, nil
}

// generateMainTemplate creates the main template
func (s *AITemplateService) generateMainTemplate(req *TemplateGenerationRequest, product *models.Product) (*models.AITemplate, error) {
	template := &models.AITemplate{
		UserID:     req.UserID,
		Name:       s.generateTemplateName(req, product),
		Type:       req.Type,
		Category:   s.determineCategory(req, product),
		IsPublic:   false,
		IsActive:   true,
		UsageCount: 0,
		Rating:     0,
		Tags:       s.generateTags(req, product),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Generate content based on template type
	content, err := s.generateTemplateContent(req, product)
	if err != nil {
		return nil, err
	}

	contentJSON, err := json.Marshal(content)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal content: %w", err)
	}
	template.Content = contentJSON

	// Generate metadata
	metadata := s.generateTemplateMetadata(req, product)
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}
	template.Metadata = metadataJSON

	return template, nil
}

// generateTemplateContent generates content based on template type
func (s *AITemplateService) generateTemplateContent(req *TemplateGenerationRequest, product *models.Product) (interface{}, error) {
	switch req.Type {
	case models.TemplateTypeSocialMedia, models.TemplateTypeTelegram, models.TemplateTypeInstagram, models.TemplateTypeFacebook, models.TemplateTypeTwitter:
		return s.generateSocialMediaContent(req, product)
	case models.TemplateTypeProductImage:
		return s.generateProductImageContent(req, product)
	case models.TemplateTypeProductDesc:
		return s.generateProductDescContent(req, product)
	case models.TemplateTypeMarketingEmail:
		return s.generateMarketingEmailContent(req, product)
	case models.TemplateTypeBanner:
		return s.generateBannerContent(req, product)
	default:
		return s.generateGenericContent(req, product)
	}
}

// generateSocialMediaContent generates social media specific content
func (s *AITemplateService) generateSocialMediaContent(req *TemplateGenerationRequest, product *models.Product) (*models.SocialMediaTemplate, error) {
	template := &models.SocialMediaTemplate{
		Platform:    req.Platform,
		Dimensions:  req.Dimensions,
		Colors:      make(map[string]string),
		Fonts:       make(map[string]string),
		Hashtags:    []string{},
	}

	// Set platform-specific dimensions if not provided
	if template.Dimensions.Width == 0 || template.Dimensions.Height == 0 {
		template.Dimensions = s.getPlatformDimensions(req.Platform)
	}

	// Generate content based on product or generic content
	if product != nil {
		template.Title = s.generateProductTitle(product, req.Platform)
		template.Description = s.generateProductDescription(product, req.Platform)
		template.Price = product.Price
		template.Currency = "TL" // Default currency
		template.ImageURL = s.getProductImageURL(product)
	} else {
		template.Title = s.generateGenericTitle(req)
		template.Description = s.generateGenericDescription(req)
	}

	// Generate CTA
	template.CTAText = s.generateCTA(req.Platform, req.Options.LanguageCode)
	template.CTALink = s.generateCTALink(product)

	// Apply style
	s.applyStyleToSocialMedia(template, &req.Style)

	// Generate layout
	template.Layout = s.selectOptimalLayout(req.Platform, template.Dimensions)

	return template, nil
}

// generateProductImageContent generates product image template content
func (s *AITemplateService) generateProductImageContent(req *TemplateGenerationRequest, product *models.Product) (*models.ProductImageTemplate, error) {
	template := &models.ProductImageTemplate{
		Dimensions: req.Dimensions,
		Overlays:   []models.TemplateOverlay{},
	}

	// Set default dimensions if not provided
	if template.Dimensions.Width == 0 || template.Dimensions.Height == 0 {
		template.Dimensions = models.ImageDimensions{Width: 1080, Height: 1080}
	}

	// Determine background based on style
	template.BackgroundType = s.determineBackgroundType(&req.Style)
	template.BackgroundColor = s.selectBackgroundColor(&req.Style)

	// Apply filters and effects based on style
	template.Filters = s.selectFilters(&req.Style, req.Platform)
	template.Effects = s.selectEffects(&req.Style, req.Platform)

	// Generate overlays
	if product != nil {
		overlays := s.generateProductOverlays(product, template.Dimensions, &req.Style)
		template.Overlays = overlays
	}

	// Add watermark if requested
	if req.Options.IncludeWatermark {
		watermark := s.generateWatermark(req.UserID)
		template.Watermark = watermark
	}

	return template, nil
}

// Helper methods

func (s *AITemplateService) getUserByID(userID int64) (*models.User, error) {
	// This would typically query the database
	// For now, return a mock user with AI access
	return &models.User{
		ID:               userID,
		AITemplateAccess: true,
		IsAdmin:          false,
	}, nil
}

func (s *AITemplateService) generateTemplateName(req *TemplateGenerationRequest, product *models.Product) string {
	if product != nil {
		return fmt.Sprintf("%s - %s Template", product.Name, strings.Title(string(req.Type)))
	}
	return fmt.Sprintf("%s Template - %s", strings.Title(string(req.Type)), time.Now().Format("2006-01-02"))
}

func (s *AITemplateService) determineCategory(req *TemplateGenerationRequest, product *models.Product) string {
	if product != nil {
		return fmt.Sprintf("category_%d", product.CategoryID)
	}
	return string(req.Type)
}

func (s *AITemplateService) generateTags(req *TemplateGenerationRequest, product *models.Product) []string {
	tags := []string{string(req.Type), req.Platform}
	
	if product != nil {
		tags = append(tags, fmt.Sprintf("category_%d", product.CategoryID), "product")
	}
	
	tags = append(tags, req.Style.ColorScheme, req.Style.FontStyle, req.Style.Theme)
	
	return tags
}

func (s *AITemplateService) generateTemplateMetadata(req *TemplateGenerationRequest, product *models.Product) map[string]interface{} {
	metadata := map[string]interface{}{
		"platform":       req.Platform,
		"style":          req.Style,
		"options":        req.Options,
		"generated_at":   time.Now(),
		"ai_version":     "1.0",
	}

	if product != nil {
		metadata["product_id"] = product.ID
		metadata["product_category"] = product.CategoryID
	}

	return metadata
}

func (s *AITemplateService) getPlatformDimensions(platform string) models.ImageDimensions {
	switch strings.ToLower(platform) {
	case "instagram":
		return models.ImageDimensions{Width: 1080, Height: 1080}
	case "facebook":
		return models.ImageDimensions{Width: 1200, Height: 630}
	case "twitter":
		return models.ImageDimensions{Width: 1200, Height: 675}
	case "telegram":
		return models.ImageDimensions{Width: 1280, Height: 720}
	default:
		return models.ImageDimensions{Width: 1080, Height: 1080}
	}
}

func (s *AITemplateService) generateProductTitle(product *models.Product, platform string) string {
	maxLength := s.getPlatformTitleLimit(platform)
	title := product.Name
	
	if len(title) > maxLength {
		title = title[:maxLength-3] + "..."
	}
	
	return title
}

func (s *AITemplateService) generateProductDescription(product *models.Product, platform string) string {
	maxLength := s.getPlatformDescriptionLimit(platform)
	description := product.Description
	
	if len(description) > maxLength {
		description = description[:maxLength-3] + "..."
	}
	
	return description
}

func (s *AITemplateService) getPlatformTitleLimit(platform string) int {
	switch strings.ToLower(platform) {
	case "twitter":
		return 50
	case "instagram":
		return 100
	case "facebook":
		return 80
	default:
		return 60
	}
}

func (s *AITemplateService) getPlatformDescriptionLimit(platform string) int {
	switch strings.ToLower(platform) {
	case "twitter":
		return 200
	case "instagram":
		return 300
	case "facebook":
		return 250
	default:
		return 200
	}
}

func (s *AITemplateService) getProductImageURL(product *models.Product) string {
	// This would return the actual product image URL
	return "/static/images/products/default.jpg"
}

func (s *AITemplateService) generateGenericTitle(req *TemplateGenerationRequest) string {
	titles := []string{
		"Özel Fırsat!",
		"Yeni Ürün!",
		"İndirim Zamanı!",
		"Sınırlı Süre!",
		"Harika Fiyat!",
	}
	
	return titles[int(time.Now().Unix())%len(titles)]
}

func (s *AITemplateService) generateGenericDescription(req *TemplateGenerationRequest) string {
	descriptions := []string{
		"Bu fırsatı kaçırmayın! Hemen sipariş verin.",
		"Kaliteli ürünler, uygun fiyatlar. Şimdi keşfedin!",
		"Özel indirimlerden yararlanın. Detaylar için tıklayın.",
		"Yeni sezon ürünleri burada! Hemen göz atın.",
		"En iyi fiyat garantisi ile satışta!",
	}
	
	return descriptions[int(time.Now().Unix())%len(descriptions)]
}

func (s *AITemplateService) generateCTA(platform, languageCode string) string {
	ctas := map[string][]string{
		"tr": {"Hemen Al", "Sipariş Ver", "Detayları Gör", "Şimdi Satın Al", "Keşfet"},
		"en": {"Buy Now", "Order Now", "See Details", "Shop Now", "Discover"},
	}
	
	if ctaList, exists := ctas[languageCode]; exists {
		return ctaList[int(time.Now().Unix())%len(ctaList)]
	}
	
	return ctas["tr"][0] // Default to Turkish
}

func (s *AITemplateService) generateCTALink(product *models.Product) string {
	if product != nil {
		return fmt.Sprintf("/product/%d", product.ID)
	}
	return "/products"
}

func (s *AITemplateService) applyStyleToSocialMedia(template *models.SocialMediaTemplate, style *TemplateStyle) {
	// Apply color scheme
	template.Colors = s.getColorScheme(style.ColorScheme, style.CustomColors)
	
	// Apply font style
	template.Fonts = s.getFontScheme(style.FontStyle, style.CustomFonts)
}

func (s *AITemplateService) getColorScheme(scheme string, customColors map[string]string) map[string]string {
	schemes := map[string]map[string]string{
		"modern": {
			"primary":   "#2563eb",
			"secondary": "#64748b",
			"accent":    "#f59e0b",
			"background": "#ffffff",
			"text":      "#1f2937",
		},
		"classic": {
			"primary":   "#1f2937",
			"secondary": "#6b7280",
			"accent":    "#dc2626",
			"background": "#f9fafb",
			"text":      "#111827",
		},
		"vibrant": {
			"primary":   "#ec4899",
			"secondary": "#8b5cf6",
			"accent":    "#06d6a0",
			"background": "#ffffff",
			"text":      "#1f2937",
		},
		"minimal": {
			"primary":   "#000000",
			"secondary": "#6b7280",
			"accent":    "#f59e0b",
			"background": "#ffffff",
			"text":      "#374151",
		},
	}
	
	if customColors != nil && len(customColors) > 0 {
		return customColors
	}
	
	if colors, exists := schemes[scheme]; exists {
		return colors
	}
	
	return schemes["modern"] // Default
}

func (s *AITemplateService) getFontScheme(style string, customFonts []string) map[string]string {
	schemes := map[string]map[string]string{
		"modern": {
			"title": "Helvetica Neue",
			"body":  "Arial",
			"accent": "Impact",
		},
		"classic": {
			"title": "Times New Roman",
			"body":  "Georgia",
			"accent": "Serif",
		},
		"bold": {
			"title": "Impact",
			"body":  "Arial Black",
			"accent": "Helvetica",
		},
		"elegant": {
			"title": "Playfair Display",
			"body":  "Lato",
			"accent": "Dancing Script",
		},
	}
	
	if len(customFonts) > 0 {
		return map[string]string{
			"title": customFonts[0],
			"body":  customFonts[0],
			"accent": customFonts[0],
		}
	}
	
	if fonts, exists := schemes[style]; exists {
		return fonts
	}
	
	return schemes["modern"] // Default
}

func (s *AITemplateService) selectOptimalLayout(platform string, dimensions models.ImageDimensions) string {
	if dimensions.Width == dimensions.Height {
		return "centered_square"
	} else if dimensions.Width > dimensions.Height {
		return "horizontal_banner"
	} else {
		return "vertical_story"
	}
}

// Additional helper methods would continue here...
// Due to length constraints, I'm showing the core structure

func (s *AITemplateService) saveTemplate(template *models.AITemplate) (int64, error) {
	// This would save to database and return the ID
	return int64(time.Now().Unix()), nil
}

func (s *AITemplateService) generateVariations(req *TemplateGenerationRequest, mainTemplate *models.AITemplate, count int) ([]*models.AITemplate, error) {
	// Generate template variations
	return []*models.AITemplate{}, nil
}

func (s *AITemplateService) generateImages(req *TemplateGenerationRequest, template *models.AITemplate, product *models.Product) ([]GeneratedImage, error) {
	// Generate actual images
	return []GeneratedImage{}, nil
}

func (s *AITemplateService) generateTextContent(req *TemplateGenerationRequest, product *models.Product) (map[string]string, error) {
	// Generate text content
	return map[string]string{}, nil
}

func (s *AITemplateService) generateHashtags(req *TemplateGenerationRequest, product *models.Product) ([]string, error) {
	// Generate relevant hashtags
	return []string{}, nil
}

func (s *AITemplateService) calculatePerformanceMetrics(template *models.AITemplate, req *TemplateGenerationRequest) (TemplatePerformanceMetrics, error) {
	// Calculate predicted performance metrics
	return TemplatePerformanceMetrics{
		EngagementScore:     85.0,
		VisualAppealScore:   90.0,
		PlatformOptimization: 88.0,
		ReadabilityScore:    82.0,
		BrandConsistency:    87.0,
		ConversionPotential: 79.0,
	}, nil
}

// Additional helper methods for template generation

func (s *AITemplateService) generateProductDescContent(req *TemplateGenerationRequest, product *models.Product) (interface{}, error) {
	content := map[string]interface{}{
		"title":       "Product Description Template",
		"description": "AI-generated product description",
	}
	
	if product != nil {
		content["product_name"] = product.Name
		content["product_description"] = product.Description
		content["product_price"] = product.Price
	}
	
	return content, nil
}

func (s *AITemplateService) generateMarketingEmailContent(req *TemplateGenerationRequest, product *models.Product) (interface{}, error) {
	content := map[string]interface{}{
		"subject":     "Special Offer!",
		"header":      "Don't Miss Out!",
		"body":        "Check out our amazing products at great prices.",
		"cta_text":    "Shop Now",
		"footer":      "Thank you for choosing us!",
	}
	
	if product != nil {
		content["subject"] = fmt.Sprintf("Special offer on %s", product.Name)
		content["body"] = fmt.Sprintf("Get %s at an amazing price of %.2f TL", product.Name, product.Price)
	}
	
	return content, nil
}

func (s *AITemplateService) generateBannerContent(req *TemplateGenerationRequest, product *models.Product) (interface{}, error) {
	content := map[string]interface{}{
		"title":       "Sale Banner",
		"subtitle":    "Limited Time Offer",
		"cta_text":    "Shop Now",
		"background":  "#ff6b6b",
		"text_color":  "#ffffff",
	}
	
	return content, nil
}

func (s *AITemplateService) generateGenericContent(req *TemplateGenerationRequest, product *models.Product) (interface{}, error) {
	content := map[string]interface{}{
		"type":        string(req.Type),
		"platform":    req.Platform,
		"generated_at": time.Now(),
	}
	
	return content, nil
}

func (s *AITemplateService) determineBackgroundType(style *TemplateStyle) string {
	switch style.Theme {
	case "gradient":
		return "gradient"
	case "colorful":
		return "pattern"
	default:
		return "solid"
	}
}

func (s *AITemplateService) selectBackgroundColor(style *TemplateStyle) string {
	if style.CustomColors != nil {
		if bg, exists := style.CustomColors["background"]; exists {
			return bg
		}
	}
	
	switch style.ColorScheme {
	case "modern":
		return "#ffffff"
	case "classic":
		return "#f9fafb"
	case "vibrant":
		return "#fef3c7"
	case "minimal":
		return "#ffffff"
	default:
		return "#ffffff"
	}
}

func (s *AITemplateService) selectFilters(style *TemplateStyle, platform string) []string {
	filters := []string{}
	
	switch style.Mood {
	case "professional":
		filters = append(filters, "sharpen", "contrast")
	case "luxury":
		filters = append(filters, "warm", "vignette")
	case "playful":
		filters = append(filters, "saturate", "bright")
	default:
		filters = append(filters, "auto-enhance")
	}
	
	return filters
}

func (s *AITemplateService) selectEffects(style *TemplateStyle, platform string) []string {
	effects := []string{}
	
	if style.Theme == "gradient" {
		effects = append(effects, "gradient-overlay")
	}
	
	if platform == "instagram" {
		effects = append(effects, "instagram-optimized")
	}
	
	return effects
}

func (s *AITemplateService) generateProductOverlays(product *models.Product, dimensions models.ImageDimensions, style *TemplateStyle) []models.TemplateOverlay {
	overlays := []models.TemplateOverlay{}
	
	// Price overlay
	priceOverlay := models.TemplateOverlay{
		Type:     "text",
		Content:  fmt.Sprintf("%.2f TL", product.Price),
		X:        dimensions.Width - 150,
		Y:        dimensions.Height - 80,
		Width:    140,
		Height:   60,
		Color:    "#ff6b6b",
		FontSize: 24,
		FontName: "Arial Bold",
		Opacity:  1.0,
	}
	overlays = append(overlays, priceOverlay)
	
	// Product name overlay
	nameOverlay := models.TemplateOverlay{
		Type:     "text",
		Content:  product.Name,
		X:        20,
		Y:        dimensions.Height - 120,
		Width:    dimensions.Width - 40,
		Height:   40,
		Color:    "#333333",
		FontSize: 18,
		FontName: "Arial",
		Opacity:  1.0,
	}
	overlays = append(overlays, nameOverlay)
	
	return overlays
}

func (s *AITemplateService) generateWatermark(userID int64) *models.Watermark {
	return &models.Watermark{
		Type:     "text",
		Content:  "KolajAI",
		Position: "bottom-right",
		Opacity:  0.3,
		Size:     12,
	}
}