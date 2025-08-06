package services

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"kolajAi/internal/database"
	"kolajAi/internal/models"
	"log"
	"math"
	"mime/multipart"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "image/gif"
)

// AIVisionService provides advanced AI-powered image recognition and categorization
type AIVisionService struct {
	repo           database.SimpleRepository
	productService *ProductService
	uploadPath     string
	maxFileSize    int64
	allowedTypes   []string
}

// NewAIVisionService creates a new AI vision service
func NewAIVisionService(repo database.SimpleRepository, productService *ProductService) *AIVisionService {
	return &AIVisionService{
		repo:           repo,
		productService: productService,
		uploadPath:     "web/static/uploads/images",
		maxFileSize:    10 * 1024 * 1024, // 10MB
		allowedTypes:   []string{"image/jpeg", "image/png", "image/gif", "image/webp"},
	}
}

// ImageAnalysisResult represents the result of image analysis
type ImageAnalysisResult struct {
	ImageID             string                 `json:"image_id"`
	UserID              int                    `json:"user_id"`
	OriginalFilename    string                 `json:"original_filename"`
	StoredFilename      string                 `json:"stored_filename"`
	FileSize            int64                  `json:"file_size"`
	Dimensions          ImageDimensions        `json:"dimensions"`
	Format              string                 `json:"format"`
	Hash                string                 `json:"hash"`
	DetectedObjects     []DetectedObject       `json:"detected_objects"`
	CategoryPredictions []CategoryPrediction   `json:"category_predictions"`
	ColorAnalysis       ColorAnalysis          `json:"color_analysis"`
	QualityScore        float64                `json:"quality_score"`
	Tags                []string               `json:"tags"`
	Metadata            map[string]interface{} `json:"metadata"`
	ProcessingTime      time.Duration          `json:"processing_time"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

// ImageDimensions represents image dimensions
type ImageDimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// DetectedObject represents an object detected in the image
type DetectedObject struct {
	Label       string      `json:"label"`
	Confidence  float64     `json:"confidence"`
	BoundingBox BoundingBox `json:"bounding_box"`
	Category    string      `json:"category"`
}

// BoundingBox represents object location in the image
type BoundingBox struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ColorAnalysis represents color analysis of the image
type ColorAnalysis struct {
	DominantColors []DominantColor `json:"dominant_colors"`
	ColorScheme    string          `json:"color_scheme"` // warm, cool, neutral, vibrant
	Brightness     float64         `json:"brightness"`   // 0-1
	Contrast       float64         `json:"contrast"`     // 0-1
	Saturation     float64         `json:"saturation"`   // 0-1
}

// DominantColor represents a dominant color in the image
type DominantColor struct {
	Color      string  `json:"color"`      // hex color
	Percentage float64 `json:"percentage"` // percentage of image
	RGB        RGB     `json:"rgb"`
}

// RGB represents RGB color values
type RGB struct {
	R int `json:"r"`
	G int `json:"g"`
	B int `json:"b"`
}

// UserImageLibrary represents a user's organized image library
type UserImageLibrary struct {
	UserID      int                   `json:"user_id"`
	Categories  map[string][]string   `json:"categories"` // category -> image_ids
	Tags        map[string][]string   `json:"tags"`       // tag -> image_ids
	Collections map[string]Collection `json:"collections"`
	TotalImages int                   `json:"total_images"`
	TotalSize   int64                 `json:"total_size"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

// Collection represents a user-defined image collection
type Collection struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ImageIDs    []string  `json:"image_ids"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SmartSearchQuery represents an advanced image search query
type SmartSearchQuery struct {
	UserID        int       `json:"user_id"`
	Query         string    `json:"query"`
	Categories    []string  `json:"categories"`
	Tags          []string  `json:"tags"`
	Colors        []string  `json:"colors"`
	ObjectTypes   []string  `json:"object_types"`
	DateRange     DateRange `json:"date_range"`
	QualityFilter string    `json:"quality_filter"` // high, medium, low, any
	SizeFilter    string    `json:"size_filter"`    // large, medium, small, any
	SortBy        string    `json:"sort_by"`        // relevance, date, quality, size
	Limit         int       `json:"limit"`
	Offset        int       `json:"offset"`
}

// DateRange represents a date range for filtering
type DateRange struct {
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
}

// ImageSearchResult represents search results
type ImageSearchResult struct {
	Images      []ImageAnalysisResult `json:"images"`
	TotalCount  int                   `json:"total_count"`
	ProcessTime time.Duration         `json:"process_time"`
	Query       SmartSearchQuery      `json:"query"`
	Suggestions []string              `json:"suggestions"`
}

// ProcessUploadedImage processes an uploaded image with AI analysis
func (s *AIVisionService) ProcessUploadedImage(userID int, file multipart.File, header *multipart.FileHeader) (*ImageAnalysisResult, error) {
	startTime := time.Now()

	// Validate file
	if err := s.validateFile(header); err != nil {
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Generate unique filename and hash
	hash := s.generateFileHash(fileContent)
	imageID := s.generateImageID(userID, hash)
	storedFilename := s.generateStoredFilename(imageID, header.Filename)

	// Check if image already exists for this user
	existingResult, err := s.getImageByHash(userID, hash)
	if err == nil && existingResult != nil {
		return existingResult, nil // Return existing analysis
	}

	// Decode image for analysis
	img, format, err := image.Decode(bytes.NewReader(fileContent))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Create upload directory if it doesn't exist
	uploadDir := filepath.Join(s.uploadPath, strconv.Itoa(userID))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Save file
	filePath := filepath.Join(uploadDir, storedFilename)
	if err := s.saveFile(fileContent, filePath); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Perform AI analysis
	result := &ImageAnalysisResult{
		ImageID:          imageID,
		UserID:           userID,
		OriginalFilename: header.Filename,
		StoredFilename:   storedFilename,
		FileSize:         header.Size,
		Format:           format,
		Hash:             hash,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Analyze image dimensions
	result.Dimensions = ImageDimensions{
		Width:  img.Bounds().Dx(),
		Height: img.Bounds().Dy(),
	}

	// Perform object detection
	result.DetectedObjects = s.detectObjects(img)

	// Predict categories
	result.CategoryPredictions = s.predictCategories(img, result.DetectedObjects)

	// Analyze colors
	result.ColorAnalysis = s.analyzeColors(img)

	// Calculate quality score
	result.QualityScore = s.calculateQualityScore(img, result.ColorAnalysis)

	// Generate tags
	result.Tags = s.generateTags(result.DetectedObjects, result.CategoryPredictions, result.ColorAnalysis)

	// Generate metadata
	result.Metadata = s.generateMetadata(img, result)

	result.ProcessingTime = time.Since(startTime)

	// Save analysis to database
	if err := s.saveImageAnalysis(result); err != nil {
		return nil, fmt.Errorf("failed to save analysis: %w", err)
	}

	// Update user's image library
	if err := s.updateUserLibrary(userID, result); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: Failed to update user library: %v\n", err)
	}

	return result, nil
}

// detectObjects performs object detection on the image
func (s *AIVisionService) detectObjects(img image.Image) []DetectedObject {
	// This is a simplified object detection implementation
	// In a production environment, you would integrate with:
	// - TensorFlow/PyTorch models
	// - OpenCV
	// - Cloud vision APIs (Google Vision, AWS Rekognition, Azure Computer Vision)
	// - Open source models like YOLO, SSD, etc.

	objects := make([]DetectedObject, 0)
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Analyze image characteristics to detect common objects
	colorAnalysis := s.analyzeColors(img)

	// Simple heuristic-based detection (replace with real ML models)
	if s.hasTextCharacteristics(img) {
		objects = append(objects, DetectedObject{
			Label:       "text",
			Confidence:  0.75,
			BoundingBox: BoundingBox{X: 0, Y: 0, Width: width, Height: height},
			Category:    "document",
		})
	}

	if s.hasProductCharacteristics(img, colorAnalysis) {
		objects = append(objects, DetectedObject{
			Label:       "product",
			Confidence:  0.80,
			BoundingBox: BoundingBox{X: width / 4, Y: height / 4, Width: width / 2, Height: height / 2},
			Category:    "commerce",
		})
	}

	if s.hasPersonCharacteristics(img) {
		objects = append(objects, DetectedObject{
			Label:       "person",
			Confidence:  0.70,
			BoundingBox: BoundingBox{X: width / 3, Y: height / 6, Width: width / 3, Height: height * 2 / 3},
			Category:    "people",
		})
	}

	if s.hasVehicleCharacteristics(img, colorAnalysis) {
		objects = append(objects, DetectedObject{
			Label:       "vehicle",
			Confidence:  0.65,
			BoundingBox: BoundingBox{X: width / 6, Y: height / 3, Width: width * 2 / 3, Height: height / 3},
			Category:    "transportation",
		})
	}

	return objects
}

// predictCategories predicts product categories based on image analysis
func (s *AIVisionService) predictCategories(img image.Image, objects []DetectedObject) []CategoryPrediction {
	predictions := make([]CategoryPrediction, 0)

	// Get available categories
	categories, err := s.productService.GetAllCategories()
	if err != nil {
		return predictions
	}

	// Analyze objects and image characteristics
	for _, category := range categories {
		confidence := s.calculateCategoryConfidence(img, objects, category)
		if confidence > 0.1 {
			predictions = append(predictions, CategoryPrediction{
				CategoryID:   int(category.ID),
				CategoryName: category.Name,
				Confidence:   confidence,
			})
		}
	}

	// Sort by confidence
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].Confidence > predictions[j].Confidence
	})

	// Limit to top 5
	if len(predictions) > 5 {
		predictions = predictions[:5]
	}

	return predictions
}

// calculateCategoryConfidence calculates confidence for category prediction
func (s *AIVisionService) calculateCategoryConfidence(img image.Image, objects []DetectedObject, category models.Category) float64 {
	confidence := 0.0
	categoryLower := strings.ToLower(category.Name)

	// Object-based confidence
	for _, obj := range objects {
		objCategory := strings.ToLower(obj.Category)
		objLabel := strings.ToLower(obj.Label)

		if strings.Contains(categoryLower, objCategory) || strings.Contains(categoryLower, objLabel) {
			confidence += obj.Confidence * 0.4
		}

		// Specific category mappings
		switch categoryLower {
		case "elektronik":
			if objCategory == "electronics" || objLabel == "device" {
				confidence += 0.3
			}
		case "giyim":
			if objCategory == "clothing" || objLabel == "apparel" {
				confidence += 0.3
			}
		case "ev & yaşam":
			if objCategory == "home" || objLabel == "furniture" {
				confidence += 0.3
			}
		case "spor":
			if objCategory == "sports" || objLabel == "equipment" {
				confidence += 0.3
			}
		}
	}

	// Color-based confidence
	colorAnalysis := s.analyzeColors(img)
	switch categoryLower {
	case "moda", "giyim":
		if colorAnalysis.ColorScheme == "vibrant" {
			confidence += 0.1
		}
	case "ev & yaşam":
		if colorAnalysis.ColorScheme == "neutral" {
			confidence += 0.1
		}
	}

	return math.Min(confidence, 1.0)
}

// analyzeColors performs color analysis on the image
func (s *AIVisionService) analyzeColors(img image.Image) ColorAnalysis {
	bounds := img.Bounds()
	colorCounts := make(map[string]int)
	totalPixels := 0
	totalR, totalG, totalB := 0, 0, 0

	// Sample pixels for performance (every 10th pixel)
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 10 {
		for x := bounds.Min.X; x < bounds.Max.X; x += 10 {
			r, g, b, _ := img.At(x, y).RGBA()
			r, g, b = r>>8, g>>8, b>>8 // Convert to 8-bit

			totalR += int(r)
			totalG += int(g)
			totalB += int(b)
			totalPixels++

			// Quantize color to reduce variations
			quantizedColor := s.quantizeColor(int(r), int(g), int(b))
			colorCounts[quantizedColor]++
		}
	}

	// Find dominant colors
	dominantColors := s.findDominantColors(colorCounts, totalPixels)

	// Calculate average brightness
	avgR := float64(totalR) / float64(totalPixels)
	avgG := float64(totalG) / float64(totalPixels)
	avgB := float64(totalB) / float64(totalPixels)
	brightness := (avgR + avgG + avgB) / (3 * 255)

	// Calculate contrast (simplified)
	contrast := s.calculateContrast(dominantColors)

	// Calculate saturation (simplified)
	saturation := s.calculateSaturation(dominantColors)

	// Determine color scheme
	colorScheme := s.determineColorScheme(dominantColors, brightness, saturation)

	return ColorAnalysis{
		DominantColors: dominantColors,
		ColorScheme:    colorScheme,
		Brightness:     brightness,
		Contrast:       contrast,
		Saturation:     saturation,
	}
}

// quantizeColor quantizes RGB values to reduce color variations
func (s *AIVisionService) quantizeColor(r, g, b int) string {
	// Quantize to 32 levels per channel
	qR := (r / 8) * 8
	qG := (g / 8) * 8
	qB := (b / 8) * 8
	return fmt.Sprintf("#%02x%02x%02x", qR, qG, qB)
}

// findDominantColors finds the most dominant colors in the image
func (s *AIVisionService) findDominantColors(colorCounts map[string]int, totalPixels int) []DominantColor {
	type colorCount struct {
		color string
		count int
	}

	colors := make([]colorCount, 0, len(colorCounts))
	for color, count := range colorCounts {
		colors = append(colors, colorCount{color: color, count: count})
	}

	// Sort by count
	sort.Slice(colors, func(i, j int) bool {
		return colors[i].count > colors[j].count
	})

	// Take top 5 colors
	dominantColors := make([]DominantColor, 0, 5)
	for i, cc := range colors {
		if i >= 5 {
			break
		}

		percentage := float64(cc.count) / float64(totalPixels) * 100
		r, g, b := s.hexToRGB(cc.color)

		dominantColors = append(dominantColors, DominantColor{
			Color:      cc.color,
			Percentage: percentage,
			RGB:        RGB{R: r, G: g, B: b},
		})
	}

	return dominantColors
}

// hexToRGB converts hex color to RGB
func (s *AIVisionService) hexToRGB(hex string) (int, int, int) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, 0, 0
	}

	r, _ := strconv.ParseInt(hex[0:2], 16, 64)
	g, _ := strconv.ParseInt(hex[2:4], 16, 64)
	b, _ := strconv.ParseInt(hex[4:6], 16, 64)

	return int(r), int(g), int(b)
}

// calculateContrast calculates image contrast
func (s *AIVisionService) calculateContrast(colors []DominantColor) float64 {
	if len(colors) < 2 {
		return 0.0
	}

	// Simple contrast calculation based on brightness difference
	maxBrightness := 0.0
	minBrightness := 1.0

	for _, color := range colors {
		brightness := (float64(color.RGB.R) + float64(color.RGB.G) + float64(color.RGB.B)) / (3 * 255)
		if brightness > maxBrightness {
			maxBrightness = brightness
		}
		if brightness < minBrightness {
			minBrightness = brightness
		}
	}

	return maxBrightness - minBrightness
}

// calculateSaturation calculates average saturation
func (s *AIVisionService) calculateSaturation(colors []DominantColor) float64 {
	if len(colors) == 0 {
		return 0.0
	}

	totalSaturation := 0.0
	for _, color := range colors {
		r, g, b := float64(color.RGB.R)/255, float64(color.RGB.G)/255, float64(color.RGB.B)/255
		max := math.Max(math.Max(r, g), b)
		min := math.Min(math.Min(r, g), b)

		var saturation float64
		if max != 0 {
			saturation = (max - min) / max
		}

		totalSaturation += saturation * color.Percentage / 100
	}

	return totalSaturation
}

// determineColorScheme determines the overall color scheme
func (s *AIVisionService) determineColorScheme(colors []DominantColor, brightness, saturation float64) string {
	if saturation > 0.6 {
		return "vibrant"
	} else if brightness > 0.7 {
		return "light"
	} else if brightness < 0.3 {
		return "dark"
	} else if saturation < 0.2 {
		return "neutral"
	} else {
		// Analyze temperature
		warmColors := 0.0
		coolColors := 0.0

		for _, color := range colors {
			r, g, b := color.RGB.R, color.RGB.G, color.RGB.B
			if r > b && (r > g || (r+g) > b*2) {
				warmColors += color.Percentage
			} else if b > r && (b > g || (b+g) > r*2) {
				coolColors += color.Percentage
			}
		}

		if warmColors > coolColors {
			return "warm"
		} else if coolColors > warmColors {
			return "cool"
		}
	}

	return "balanced"
}

// calculateQualityScore calculates image quality score
func (s *AIVisionService) calculateQualityScore(img image.Image, colorAnalysis ColorAnalysis) float64 {
	score := 0.0
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Resolution score (30%)
	totalPixels := width * height
	resolutionScore := math.Min(float64(totalPixels)/1000000.0, 1.0) // Normalize to 1MP
	score += resolutionScore * 0.3

	// Aspect ratio score (10%)
	aspectRatio := float64(width) / float64(height)
	aspectScore := 1.0
	if aspectRatio < 0.3 || aspectRatio > 3.0 {
		aspectScore = 0.5 // Penalize extreme aspect ratios
	}
	score += aspectScore * 0.1

	// Color diversity score (20%)
	colorDiversityScore := math.Min(float64(len(colorAnalysis.DominantColors))/5.0, 1.0)
	score += colorDiversityScore * 0.2

	// Contrast score (20%)
	score += colorAnalysis.Contrast * 0.2

	// Brightness score (10%)
	brightnessScore := 1.0 - math.Abs(colorAnalysis.Brightness-0.5)*2 // Optimal at 0.5
	score += brightnessScore * 0.1

	// Saturation score (10%)
	saturationScore := math.Min(colorAnalysis.Saturation*2, 1.0) // Prefer some saturation
	score += saturationScore * 0.1

	return math.Min(score, 1.0)
}

// generateTags generates descriptive tags for the image
func (s *AIVisionService) generateTags(objects []DetectedObject, categories []CategoryPrediction, colorAnalysis ColorAnalysis) []string {
	tags := make([]string, 0)
	tagSet := make(map[string]bool) // To avoid duplicates

	// Object-based tags
	for _, obj := range objects {
		if obj.Confidence > 0.5 {
			tag := strings.ToLower(obj.Label)
			if !tagSet[tag] {
				tags = append(tags, tag)
				tagSet[tag] = true
			}
		}
	}

	// Category-based tags
	for _, cat := range categories {
		if cat.Confidence > 0.3 {
			tag := strings.ToLower(cat.CategoryName)
			if !tagSet[tag] {
				tags = append(tags, tag)
				tagSet[tag] = true
			}
		}
	}

	// Color-based tags
	colorTag := colorAnalysis.ColorScheme
	if !tagSet[colorTag] {
		tags = append(tags, colorTag)
		tagSet[colorTag] = true
	}

	// Dominant color tags
	for _, color := range colorAnalysis.DominantColors {
		if color.Percentage > 20 { // Only significant colors
			colorName := s.getColorName(color.RGB)
			if colorName != "" && !tagSet[colorName] {
				tags = append(tags, colorName)
				tagSet[colorName] = true
			}
		}
	}

	return tags
}

// getColorName gets a human-readable color name from RGB
func (s *AIVisionService) getColorName(rgb RGB) string {
	r, g, b := rgb.R, rgb.G, rgb.B

	// Simple color name mapping
	if r > 200 && g < 100 && b < 100 {
		return "kırmızı"
	} else if g > 200 && r < 100 && b < 100 {
		return "yeşil"
	} else if b > 200 && r < 100 && g < 100 {
		return "mavi"
	} else if r > 200 && g > 200 && b < 100 {
		return "sarı"
	} else if r > 150 && g < 100 && b > 150 {
		return "mor"
	} else if r > 200 && g > 100 && b < 100 {
		return "turuncu"
	} else if r < 100 && g < 100 && b < 100 {
		return "siyah"
	} else if r > 200 && g > 200 && b > 200 {
		return "beyaz"
	} else if r > 100 && g > 100 && b > 100 && r < 150 && g < 150 && b < 150 {
		return "gri"
	}

	return ""
}

// generateMetadata generates additional metadata for the image
func (s *AIVisionService) generateMetadata(img image.Image, result *ImageAnalysisResult) map[string]interface{} {
	metadata := make(map[string]interface{})

	bounds := img.Bounds()
	metadata["width"] = bounds.Dx()
	metadata["height"] = bounds.Dy()
	metadata["aspect_ratio"] = float64(bounds.Dx()) / float64(bounds.Dy())
	metadata["total_pixels"] = bounds.Dx() * bounds.Dy()
	metadata["format"] = result.Format
	metadata["file_size"] = result.FileSize
	metadata["quality_score"] = result.QualityScore
	metadata["processing_time_ms"] = result.ProcessingTime.Milliseconds()

	return metadata
}

// Helper methods for object detection heuristics
func (s *AIVisionService) hasTextCharacteristics(img image.Image) bool {
	// Simplified text detection heuristic
	// In production, use OCR libraries like Tesseract
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Check for high contrast regions that might indicate text
	highContrastRegions := 0
	sampleSize := 100

	for i := 0; i < sampleSize; i++ {
		x := bounds.Min.X + (width * i / sampleSize)
		y := bounds.Min.Y + height/2

		if x < bounds.Max.X && y < bounds.Max.Y {
			r1, g1, b1, _ := img.At(x, y).RGBA()
			if x+1 < bounds.Max.X {
				r2, g2, b2, _ := img.At(x+1, y).RGBA()

				diff := math.Abs(float64(r1-r2)) + math.Abs(float64(g1-g2)) + math.Abs(float64(b1-b2))
				if diff > 30000 { // High contrast threshold
					highContrastRegions++
				}
			}
		}
	}

	return float64(highContrastRegions)/float64(sampleSize) > 0.3
}

func (s *AIVisionService) hasProductCharacteristics(img image.Image, colorAnalysis ColorAnalysis) bool {
	// Products typically have:
	// - Good lighting (brightness between 0.3-0.8)
	// - Reasonable contrast
	// - Not too many dominant colors (clean background)

	return colorAnalysis.Brightness > 0.3 &&
		colorAnalysis.Brightness < 0.8 &&
		colorAnalysis.Contrast > 0.2 &&
		len(colorAnalysis.DominantColors) <= 4
}

func (s *AIVisionService) hasPersonCharacteristics(img image.Image) bool {
	// Simplified person detection heuristic
	// In production, use face detection libraries or ML models
	bounds := img.Bounds()
	_, height := bounds.Dx(), bounds.Dy()

	// Check for skin-tone colors in upper portion of image
	skinTonePixels := 0
	totalSamples := 0

	for y := bounds.Min.Y; y < bounds.Min.Y+height/3; y += 5 {
		for x := bounds.Min.X; x < bounds.Max.X; x += 5 {
			r, g, b, _ := img.At(x, y).RGBA()
			r, g, b = r>>8, g>>8, b>>8

			if s.isSkinTone(int(r), int(g), int(b)) {
				skinTonePixels++
			}
			totalSamples++
		}
	}

	return totalSamples > 0 && float64(skinTonePixels)/float64(totalSamples) > 0.1
}

func (s *AIVisionService) isSkinTone(r, g, b int) bool {
	// Simplified skin tone detection
	return r > 95 && g > 40 && b > 20 &&
		r > g && r > b &&
		r-g > 15 && r-b > 15
}

func (s *AIVisionService) hasVehicleCharacteristics(img image.Image, colorAnalysis ColorAnalysis) bool {
	// Vehicles typically have:
	// - Strong geometric shapes
	// - Metallic colors
	// - High contrast edges

	return colorAnalysis.Contrast > 0.4 &&
		(s.hasMetallicColors(colorAnalysis.DominantColors) ||
			colorAnalysis.ColorScheme == "neutral")
}

func (s *AIVisionService) hasMetallicColors(colors []DominantColor) bool {
	for _, color := range colors {
		r, g, b := color.RGB.R, color.RGB.G, color.RGB.B
		// Check for metallic grays, silvers
		if math.Abs(float64(r-g)) < 30 && math.Abs(float64(g-b)) < 30 && math.Abs(float64(r-b)) < 30 {
			if r > 100 && r < 200 { // Gray range
				return true
			}
		}
	}
	return false
}

// Utility methods
func (s *AIVisionService) validateFile(header *multipart.FileHeader) error {
	// Check file size
	if header.Size > s.maxFileSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size %d", header.Size, s.maxFileSize)
	}

	// Check file type
	contentType := header.Header.Get("Content-Type")
	allowed := false
	for _, allowedType := range s.allowedTypes {
		if contentType == allowedType {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf("file type %s not allowed", contentType)
	}

	return nil
}

func (s *AIVisionService) generateFileHash(content []byte) string {
	hash := md5.Sum(content)
	return hex.EncodeToString(hash[:])
}

func (s *AIVisionService) generateImageID(userID int, hash string) string {
	return fmt.Sprintf("%d_%s_%d", userID, hash[:8], time.Now().Unix())
}

func (s *AIVisionService) generateStoredFilename(imageID, originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	return fmt.Sprintf("%s%s", imageID, ext)
}

func (s *AIVisionService) saveFile(content []byte, filePath string) error {
	return os.WriteFile(filePath, content, 0644)
}

// Database operations
func (s *AIVisionService) saveImageAnalysis(result *ImageAnalysisResult) error {
	// Convert complex fields to JSON
	objectsJSON, _ := json.Marshal(result.DetectedObjects)
	categoriesJSON, _ := json.Marshal(result.CategoryPredictions)
	colorsJSON, _ := json.Marshal(result.ColorAnalysis)
	tagsJSON, _ := json.Marshal(result.Tags)
	metadataJSON, _ := json.Marshal(result.Metadata)

	query := `
		INSERT INTO ai_image_analysis (
			image_id, user_id, original_filename, stored_filename, file_size,
			width, height, format, hash, detected_objects, category_predictions,
			color_analysis, quality_score, tags, metadata, processing_time_ms,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.repo.Exec(query,
		result.ImageID, result.UserID, result.OriginalFilename, result.StoredFilename,
		result.FileSize, result.Dimensions.Width, result.Dimensions.Height,
		result.Format, result.Hash, string(objectsJSON), string(categoriesJSON),
		string(colorsJSON), result.QualityScore, string(tagsJSON),
		string(metadataJSON), result.ProcessingTime.Milliseconds(),
		result.CreatedAt, result.UpdatedAt,
	)

	return err
}

func (s *AIVisionService) getImageByHash(userID int, hash string) (*ImageAnalysisResult, error) {
	query := `
		SELECT image_id, user_id, original_filename, stored_filename, file_size,
			   width, height, format, hash, detected_objects, category_predictions,
			   color_analysis, quality_score, tags, metadata, processing_time_ms,
			   created_at, updated_at
		FROM ai_image_analysis 
		WHERE user_id = ? AND hash = ?
		LIMIT 1
	`

	var result ImageAnalysisResult
	var objectsJSON, categoriesJSON, colorsJSON, tagsJSON, metadataJSON string
	var processingTimeMs int64

	err := s.repo.QueryRow(query, userID, hash).Scan(
		&result.ImageID, &result.UserID, &result.OriginalFilename, &result.StoredFilename,
		&result.FileSize, &result.Dimensions.Width, &result.Dimensions.Height,
		&result.Format, &result.Hash, &objectsJSON, &categoriesJSON,
		&colorsJSON, &result.QualityScore, &tagsJSON, &metadataJSON,
		&processingTimeMs, &result.CreatedAt, &result.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	json.Unmarshal([]byte(objectsJSON), &result.DetectedObjects)
	json.Unmarshal([]byte(categoriesJSON), &result.CategoryPredictions)
	json.Unmarshal([]byte(colorsJSON), &result.ColorAnalysis)
	json.Unmarshal([]byte(tagsJSON), &result.Tags)
	json.Unmarshal([]byte(metadataJSON), &result.Metadata)
	result.ProcessingTime = time.Duration(processingTimeMs) * time.Millisecond

	return &result, nil
}

func (s *AIVisionService) updateUserLibrary(userID int, result *ImageAnalysisResult) error {
	// This would update the user's image library organization
	// For now, we'll implement a simple version

	// Update category associations
	for _, cat := range result.CategoryPredictions {
		if cat.Confidence > 0.5 {
			query := `
				INSERT OR IGNORE INTO user_image_categories (user_id, image_id, category_id, confidence)
				VALUES (?, ?, ?, ?)
			`
			s.repo.Exec(query, userID, result.ImageID, cat.CategoryID, cat.Confidence)
		}
	}

	// Update tag associations
	for _, tag := range result.Tags {
		query := `
			INSERT OR IGNORE INTO user_image_tags (user_id, image_id, tag)
			VALUES (?, ?, ?)
		`
		s.repo.Exec(query, userID, result.ImageID, tag)
	}

	return nil
}

// SmartImageSearch performs advanced AI-powered image search
func (s *AIVisionService) SmartImageSearch(query SmartSearchQuery) (*ImageSearchResult, error) {
	startTime := time.Now()

	// Build search conditions
	conditions := []string{"user_id = ?"}
	args := []interface{}{query.UserID}

	// Text search in tags and categories
	if query.Query != "" {
		conditions = append(conditions, "(tags LIKE ? OR category_predictions LIKE ?)")
		searchPattern := "%" + strings.ToLower(query.Query) + "%"
		args = append(args, searchPattern, searchPattern)
	}

	// Category filter
	if len(query.Categories) > 0 {
		categoryConditions := make([]string, len(query.Categories))
		for i, cat := range query.Categories {
			categoryConditions[i] = "category_predictions LIKE ?"
			args = append(args, "%"+cat+"%")
		}
		conditions = append(conditions, "("+strings.Join(categoryConditions, " OR ")+")")
	}

	// Tag filter
	if len(query.Tags) > 0 {
		tagConditions := make([]string, len(query.Tags))
		for i, tag := range query.Tags {
			tagConditions[i] = "tags LIKE ?"
			args = append(args, "%"+tag+"%")
		}
		conditions = append(conditions, "("+strings.Join(tagConditions, " OR ")+")")
	}

	// Quality filter
	switch query.QualityFilter {
	case "high":
		conditions = append(conditions, "quality_score >= 0.8")
	case "medium":
		conditions = append(conditions, "quality_score >= 0.5 AND quality_score < 0.8")
	case "low":
		conditions = append(conditions, "quality_score < 0.5")
	}

	// Size filter
	switch query.SizeFilter {
	case "large":
		conditions = append(conditions, "width * height >= 1000000") // 1MP+
	case "medium":
		conditions = append(conditions, "width * height >= 300000 AND width * height < 1000000")
	case "small":
		conditions = append(conditions, "width * height < 300000")
	}

	// Date range filter
	if query.DateRange.StartDate != nil {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, *query.DateRange.StartDate)
	}
	if query.DateRange.EndDate != nil {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, *query.DateRange.EndDate)
	}

	// Build ORDER BY clause
	orderBy := "created_at DESC" // Default
	switch query.SortBy {
	case "quality":
		orderBy = "quality_score DESC"
	case "size":
		orderBy = "width * height DESC"
	case "relevance":
		if query.Query != "" {
			// Simple relevance scoring - could be improved with full-text search
			orderBy = "quality_score DESC"
		}
	}

	// Count total results
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ai_image_analysis WHERE %s", strings.Join(conditions, " AND "))
	var totalCount int
	err := s.repo.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count results: %w", err)
	}

	// Get paginated results
	searchQuery := fmt.Sprintf(`
		SELECT image_id, user_id, original_filename, stored_filename, file_size,
			   width, height, format, hash, detected_objects, category_predictions,
			   color_analysis, quality_score, tags, metadata, processing_time_ms,
			   created_at, updated_at
		FROM ai_image_analysis 
		WHERE %s
		ORDER BY %s
		LIMIT ? OFFSET ?
	`, strings.Join(conditions, " AND "), orderBy)

	args = append(args, query.Limit, query.Offset)

	rows, err := s.repo.Query(searchQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	results := make([]ImageAnalysisResult, 0)
	for rows.Next() {
		var result ImageAnalysisResult
		var objectsJSON, categoriesJSON, colorsJSON, tagsJSON, metadataJSON string
		var processingTimeMs int64

		err := rows.Scan(
			&result.ImageID, &result.UserID, &result.OriginalFilename, &result.StoredFilename,
			&result.FileSize, &result.Dimensions.Width, &result.Dimensions.Height,
			&result.Format, &result.Hash, &objectsJSON, &categoriesJSON,
			&colorsJSON, &result.QualityScore, &tagsJSON, &metadataJSON,
			&processingTimeMs, &result.CreatedAt, &result.UpdatedAt,
		)
		if err != nil {
			continue
		}

		// Parse JSON fields
		json.Unmarshal([]byte(objectsJSON), &result.DetectedObjects)
		json.Unmarshal([]byte(categoriesJSON), &result.CategoryPredictions)
		json.Unmarshal([]byte(colorsJSON), &result.ColorAnalysis)
		json.Unmarshal([]byte(tagsJSON), &result.Tags)
		json.Unmarshal([]byte(metadataJSON), &result.Metadata)
		result.ProcessingTime = time.Duration(processingTimeMs) * time.Millisecond

		results = append(results, result)
	}

	// Generate search suggestions
	suggestions := s.generateSearchSuggestions(query, results)

	return &ImageSearchResult{
		Images:      results,
		TotalCount:  totalCount,
		ProcessTime: time.Since(startTime),
		Query:       query,
		Suggestions: suggestions,
	}, nil
}

// generateSearchSuggestions generates search suggestions based on current query and results
func (s *AIVisionService) generateSearchSuggestions(query SmartSearchQuery, results []ImageAnalysisResult) []string {
	suggestions := make([]string, 0)
	tagCounts := make(map[string]int)

	// Analyze tags from current results
	for _, result := range results {
		for _, tag := range result.Tags {
			if !contains(query.Tags, tag) && tag != strings.ToLower(query.Query) {
				tagCounts[tag]++
			}
		}
	}

	// Sort tags by frequency
	type tagCount struct {
		tag   string
		count int
	}

	tagList := make([]tagCount, 0, len(tagCounts))
	for tag, count := range tagCounts {
		tagList = append(tagList, tagCount{tag: tag, count: count})
	}

	sort.Slice(tagList, func(i, j int) bool {
		return tagList[i].count > tagList[j].count
	})

	// Add top tags as suggestions
	for i, tc := range tagList {
		if i >= 5 {
			break
		}
		if tc.count >= 2 { // Only suggest tags that appear multiple times
			suggestions = append(suggestions, tc.tag)
		}
	}

	return suggestions
}

// GetUserImageLibrary retrieves and organizes a user's image library
func (s *AIVisionService) GetUserImageLibrary(userID int) (*UserImageLibrary, error) {
	library := &UserImageLibrary{
		UserID:      userID,
		Categories:  make(map[string][]string),
		Tags:        make(map[string][]string),
		Collections: make(map[string]Collection),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Get all user images
	query := `
		SELECT image_id, tags, category_predictions, file_size
		FROM ai_image_analysis 
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := s.repo.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user images: %w", err)
	}
	defer rows.Close()

	totalSize := int64(0)
	imageCount := 0

	for rows.Next() {
		var imageID, tagsJSON, categoriesJSON string
		var fileSize int64

		err := rows.Scan(&imageID, &tagsJSON, &categoriesJSON, &fileSize)
		if err != nil {
			continue
		}

		imageCount++
		totalSize += fileSize

		// Parse tags
		var tags []string
		if err := json.Unmarshal([]byte(tagsJSON), &tags); err == nil {
			for _, tag := range tags {
				if library.Tags[tag] == nil {
					library.Tags[tag] = make([]string, 0)
				}
				library.Tags[tag] = append(library.Tags[tag], imageID)
			}
		}

		// Parse categories
		var categories []CategoryPrediction
		if err := json.Unmarshal([]byte(categoriesJSON), &categories); err == nil {
			for _, cat := range categories {
				if cat.Confidence > 0.5 {
					if library.Categories[cat.CategoryName] == nil {
						library.Categories[cat.CategoryName] = make([]string, 0)
					}
					library.Categories[cat.CategoryName] = append(library.Categories[cat.CategoryName], imageID)
				}
			}
		}
	}

	library.TotalImages = imageCount
	library.TotalSize = totalSize

	// Load user collections
	collections, err := s.getUserCollections(userID)
	if err == nil {
		for _, collection := range collections {
			library.Collections[collection.ID] = collection
		}
	}

	return library, nil
}

// getUserCollections retrieves user's custom image collections
func (s *AIVisionService) getUserCollections(userID int) ([]Collection, error) {
	query := `
		SELECT collection_id, name, description, image_ids, is_public, created_at, updated_at
		FROM user_image_collections 
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := s.repo.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	collections := make([]Collection, 0)
	for rows.Next() {
		var collection Collection
		var imageIDsJSON string

		err := rows.Scan(
			&collection.ID, &collection.Name, &collection.Description,
			&imageIDsJSON, &collection.IsPublic, &collection.CreatedAt, &collection.UpdatedAt,
		)
		if err != nil {
			continue
		}

		// Parse image IDs
		if err := json.Unmarshal([]byte(imageIDsJSON), &collection.ImageIDs); err != nil {
			collection.ImageIDs = make([]string, 0)
		}

		collections = append(collections, collection)
	}

	return collections, nil
}

// CreateImageCollection creates a new image collection for a user
func (s *AIVisionService) CreateImageCollection(userID int, name, description string, imageIDs []string, isPublic bool) (*Collection, error) {
	collectionID := fmt.Sprintf("coll_%d_%d", userID, time.Now().Unix())

	collection := &Collection{
		ID:          collectionID,
		Name:        name,
		Description: description,
		ImageIDs:    imageIDs,
		IsPublic:    isPublic,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	imageIDsJSON, _ := json.Marshal(imageIDs)

	query := `
		INSERT INTO user_image_collections (
			user_id, collection_id, name, description, image_ids, is_public, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.repo.Exec(query,
		userID, collection.ID, collection.Name, collection.Description,
		string(imageIDsJSON), collection.IsPublic, collection.CreatedAt, collection.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	return collection, nil
}

// UpdateImageCollection updates an existing image collection
func (s *AIVisionService) UpdateImageCollection(userID int, collectionID string, name, description string, imageIDs []string, isPublic bool) error {
	imageIDsJSON, _ := json.Marshal(imageIDs)

	query := `
		UPDATE user_image_collections 
		SET name = ?, description = ?, image_ids = ?, is_public = ?, updated_at = ?
		WHERE user_id = ? AND collection_id = ?
	`

	result, err := s.repo.Exec(query,
		name, description, string(imageIDsJSON), isPublic, time.Now(),
		userID, collectionID,
	)

	if err != nil {
		return fmt.Errorf("failed to update collection: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("collection not found or not owned by user")
	}

	return nil
}

// DeleteImageCollection deletes an image collection
func (s *AIVisionService) DeleteImageCollection(userID int, collectionID string) error {
	query := `DELETE FROM user_image_collections WHERE user_id = ? AND collection_id = ?`

	result, err := s.repo.Exec(query, userID, collectionID)
	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("collection not found or not owned by user")
	}

	return nil
}

// GetImageAnalysis retrieves detailed analysis for a specific image
func (s *AIVisionService) GetImageAnalysis(userID int, imageID string) (*ImageAnalysisResult, error) {
	query := `
		SELECT image_id, user_id, original_filename, stored_filename, file_size,
			   width, height, format, hash, detected_objects, category_predictions,
			   color_analysis, quality_score, tags, metadata, processing_time_ms,
			   created_at, updated_at
		FROM ai_image_analysis 
		WHERE user_id = ? AND image_id = ?
		LIMIT 1
	`

	var result ImageAnalysisResult
	var objectsJSON, categoriesJSON, colorsJSON, tagsJSON, metadataJSON string
	var processingTimeMs int64

	err := s.repo.QueryRow(query, userID, imageID).Scan(
		&result.ImageID, &result.UserID, &result.OriginalFilename, &result.StoredFilename,
		&result.FileSize, &result.Dimensions.Width, &result.Dimensions.Height,
		&result.Format, &result.Hash, &objectsJSON, &categoriesJSON,
		&colorsJSON, &result.QualityScore, &tagsJSON, &metadataJSON,
		&processingTimeMs, &result.CreatedAt, &result.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}

	// Parse JSON fields
	json.Unmarshal([]byte(objectsJSON), &result.DetectedObjects)
	json.Unmarshal([]byte(categoriesJSON), &result.CategoryPredictions)
	json.Unmarshal([]byte(colorsJSON), &result.ColorAnalysis)
	json.Unmarshal([]byte(tagsJSON), &result.Tags)
	json.Unmarshal([]byte(metadataJSON), &result.Metadata)
	result.ProcessingTime = time.Duration(processingTimeMs) * time.Millisecond

	return &result, nil
}

// DeleteImage deletes an image and its analysis
func (s *AIVisionService) DeleteImage(userID int, imageID string) error {
	// Get image info first
	analysis, err := s.GetImageAnalysis(userID, imageID)
	if err != nil {
		return fmt.Errorf("image not found: %w", err)
	}

	// Delete physical file
	filePath := filepath.Join(s.uploadPath, strconv.Itoa(userID), analysis.StoredFilename)
	if err := os.Remove(filePath); err != nil {
		// Log error but continue with database cleanup
		fmt.Printf("Warning: Failed to delete physical file %s: %v\n", filePath, err)
	}

	// Delete from database
	tx, err := s.repo.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete from main analysis table
	_, err = tx.Exec("DELETE FROM ai_image_analysis WHERE user_id = ? AND image_id = ?", userID, imageID)
	if err != nil {
		return fmt.Errorf("failed to delete image analysis: %w", err)
	}

	// Delete from category associations
	_, err = tx.Exec("DELETE FROM user_image_categories WHERE user_id = ? AND image_id = ?", userID, imageID)
	if err != nil {
		// Log error but continue
		fmt.Printf("Warning: Failed to delete category associations: %v\n", err)
	}

	// Delete from tag associations
	_, err = tx.Exec("DELETE FROM user_image_tags WHERE user_id = ? AND image_id = ?", userID, imageID)
	if err != nil {
		// Log error but continue
		fmt.Printf("Warning: Failed to delete tag associations: %v\n", err)
	}

	return tx.Commit()
}

// GetImagesByCategory retrieves images by category for a user
func (s *AIVisionService) GetImagesByCategory(userID int, categoryName string, limit, offset int) ([]ImageAnalysisResult, error) {
	query := `
		SELECT image_id, user_id, original_filename, stored_filename, file_size,
			   width, height, format, hash, detected_objects, category_predictions,
			   color_analysis, quality_score, tags, metadata, processing_time_ms,
			   created_at, updated_at
		FROM ai_image_analysis 
		WHERE user_id = ? AND category_predictions LIKE ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.repo.Query(query, userID, "%"+categoryName+"%", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get images by category: %w", err)
	}
	defer rows.Close()

	results := make([]ImageAnalysisResult, 0)
	for rows.Next() {
		var result ImageAnalysisResult
		var objectsJSON, categoriesJSON, colorsJSON, tagsJSON, metadataJSON string
		var processingTimeMs int64

		err := rows.Scan(
			&result.ImageID, &result.UserID, &result.OriginalFilename, &result.StoredFilename,
			&result.FileSize, &result.Dimensions.Width, &result.Dimensions.Height,
			&result.Format, &result.Hash, &objectsJSON, &categoriesJSON,
			&colorsJSON, &result.QualityScore, &tagsJSON, &metadataJSON,
			&processingTimeMs, &result.CreatedAt, &result.UpdatedAt,
		)
		if err != nil {
			continue
		}

		// Parse JSON fields
		json.Unmarshal([]byte(objectsJSON), &result.DetectedObjects)
		json.Unmarshal([]byte(categoriesJSON), &result.CategoryPredictions)
		json.Unmarshal([]byte(colorsJSON), &result.ColorAnalysis)
		json.Unmarshal([]byte(tagsJSON), &result.Tags)
		json.Unmarshal([]byte(metadataJSON), &result.Metadata)
		result.ProcessingTime = time.Duration(processingTimeMs) * time.Millisecond

		results = append(results, result)
	}

	return results, nil
}

// GetImagesByTag retrieves images by tag for a user
func (s *AIVisionService) GetImagesByTag(userID int, tag string, limit, offset int) ([]ImageAnalysisResult, error) {
	query := `
		SELECT image_id, user_id, original_filename, stored_filename, file_size,
			   width, height, format, hash, detected_objects, category_predictions,
			   color_analysis, quality_score, tags, metadata, processing_time_ms,
			   created_at, updated_at
		FROM ai_image_analysis 
		WHERE user_id = ? AND tags LIKE ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.repo.Query(query, userID, "%"+tag+"%", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get images by tag: %w", err)
	}
	defer rows.Close()

	results := make([]ImageAnalysisResult, 0)
	for rows.Next() {
		var result ImageAnalysisResult
		var objectsJSON, categoriesJSON, colorsJSON, tagsJSON, metadataJSON string
		var processingTimeMs int64

		err := rows.Scan(
			&result.ImageID, &result.UserID, &result.OriginalFilename, &result.StoredFilename,
			&result.FileSize, &result.Dimensions.Width, &result.Dimensions.Height,
			&result.Format, &result.Hash, &objectsJSON, &categoriesJSON,
			&colorsJSON, &result.QualityScore, &tagsJSON, &metadataJSON,
			&processingTimeMs, &result.CreatedAt, &result.UpdatedAt,
		)
		if err != nil {
			continue
		}

		// Parse JSON fields
		json.Unmarshal([]byte(objectsJSON), &result.DetectedObjects)
		json.Unmarshal([]byte(categoriesJSON), &result.CategoryPredictions)
		json.Unmarshal([]byte(colorsJSON), &result.ColorAnalysis)
		json.Unmarshal([]byte(tagsJSON), &result.Tags)
		json.Unmarshal([]byte(metadataJSON), &result.Metadata)
		result.ProcessingTime = time.Duration(processingTimeMs) * time.Millisecond

		results = append(results, result)
	}

	return results, nil
}

// GetUserImageStats provides statistics about user's image library
func (s *AIVisionService) GetUserImageStats(userID int) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total images and size
	query := `
		SELECT COUNT(*), COALESCE(SUM(file_size), 0), COALESCE(AVG(quality_score), 0)
		FROM ai_image_analysis 
		WHERE user_id = ?
	`

	var totalImages int
	var totalSize int64
	var avgQuality float64

	err := s.repo.QueryRow(query, userID).Scan(&totalImages, &totalSize, &avgQuality)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic stats: %w", err)
	}

	stats["total_images"] = totalImages
	stats["total_size"] = totalSize
	stats["average_quality"] = avgQuality

	// Format distribution
	formatQuery := `
		SELECT format, COUNT(*) 
		FROM ai_image_analysis 
		WHERE user_id = ? 
		GROUP BY format
	`

	rows, err := s.repo.Query(formatQuery, userID)
	if err == nil {
		formats := make(map[string]int)
		for rows.Next() {
			var format string
			var count int
			if rows.Scan(&format, &count) == nil {
				formats[format] = count
			}
		}
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
		stats["format_distribution"] = formats
	}

	// Quality distribution
	qualityQuery := `
		SELECT 
			SUM(CASE WHEN quality_score >= 0.8 THEN 1 ELSE 0 END) as high,
			SUM(CASE WHEN quality_score >= 0.5 AND quality_score < 0.8 THEN 1 ELSE 0 END) as medium,
			SUM(CASE WHEN quality_score < 0.5 THEN 1 ELSE 0 END) as low
		FROM ai_image_analysis 
		WHERE user_id = ?
	`

	var highQuality, mediumQuality, lowQuality int
	err = s.repo.QueryRow(qualityQuery, userID).Scan(&highQuality, &mediumQuality, &lowQuality)
	if err == nil {
		stats["quality_distribution"] = map[string]int{
			"high":   highQuality,
			"medium": mediumQuality,
			"low":    lowQuality,
		}
	}

	// Recent activity (last 7 days)
	recentQuery := `
		SELECT COUNT(*) 
		FROM ai_image_analysis 
		WHERE user_id = ? AND created_at >= datetime('now', '-7 days')
	`

	var recentImages int
	err = s.repo.QueryRow(recentQuery, userID).Scan(&recentImages)
	if err == nil {
		stats["recent_uploads"] = recentImages
	}

	return stats, nil
}

// SuggestProductCategories suggests product categories based on image analysis
func (s *AIVisionService) SuggestProductCategories(imageID string, userID int) ([]CategoryPrediction, error) {
	analysis, err := s.GetImageAnalysis(userID, imageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image analysis: %w", err)
	}

	// Enhanced category prediction based on detected objects and colors
	suggestions := make([]CategoryPrediction, 0)

	// Use existing predictions as base
	for _, pred := range analysis.CategoryPredictions {
		suggestions = append(suggestions, pred)
	}

	// Add additional suggestions based on detected objects
	for _, obj := range analysis.DetectedObjects {
		if obj.Confidence > 0.6 {
			// Map objects to additional categories
			additionalCategories := s.mapObjectToCategories(obj.Label)
			for _, catName := range additionalCategories {
				// Check if category exists
				categories, err := s.productService.GetAllCategories()
				if err == nil {
					for _, cat := range categories {
						if strings.EqualFold(cat.Name, catName) {
							// Check if not already in suggestions
							found := false
							for _, existing := range suggestions {
								if existing.CategoryID == int(cat.ID) {
									found = true
									break
								}
							}
							if !found {
								suggestions = append(suggestions, CategoryPrediction{
									CategoryID:   int(cat.ID),
									CategoryName: cat.Name,
									Confidence:   obj.Confidence * 0.8, // Slightly lower confidence for derived suggestions
								})
							}
						}
					}
				}
			}
		}
	}

	// Sort by confidence
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Confidence > suggestions[j].Confidence
	})

	// Limit to top 10
	if len(suggestions) > 10 {
		suggestions = suggestions[:10]
	}

	return suggestions, nil
}

// mapObjectToCategories maps detected objects to potential product categories
func (s *AIVisionService) mapObjectToCategories(objectLabel string) []string {
	objectLower := strings.ToLower(objectLabel)

	categoryMap := map[string][]string{
		"person":    {"Giyim", "Moda", "Aksesuar"},
		"product":   {"Genel", "Elektronik", "Ev & Yaşam"},
		"text":      {"Kitap", "Eğitim", "Ofis"},
		"vehicle":   {"Otomotiv", "Ulaşım", "Spor"},
		"device":    {"Elektronik", "Teknoloji", "Bilgisayar"},
		"clothing":  {"Giyim", "Moda", "Tekstil"},
		"furniture": {"Ev & Yaşam", "Mobilya", "Dekorasyon"},
		"food":      {"Gıda", "İçecek", "Organik"},
		"book":      {"Kitap", "Eğitim", "Kültür"},
		"toy":       {"Oyuncak", "Çocuk", "Eğlence"},
		"jewelry":   {"Mücevher", "Aksesuar", "Moda"},
		"shoe":      {"Ayakkabı", "Giyim", "Spor"},
		"bag":       {"Çanta", "Aksesuar", "Moda"},
		"watch":     {"Saat", "Aksesuar", "Elektronik"},
		"phone":     {"Telefon", "Elektronik", "Teknoloji"},
		"computer":  {"Bilgisayar", "Elektronik", "Teknoloji"},
		"camera":    {"Kamera", "Elektronik", "Fotoğraf"},
		"tool":      {"Araç", "Bahçe", "İnşaat"},
		"plant":     {"Bitki", "Bahçe", "Dekorasyon"},
		"animal":    {"Pet", "Hayvan", "Bakım"},
	}

	if categories, exists := categoryMap[objectLower]; exists {
		return categories
	}

	// Default fallback
	return []string{"Genel"}
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
