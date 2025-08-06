package seo

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// SEOManager handles comprehensive SEO management
type SEOManager struct {
	db       *sql.DB
	config   SEOConfig
	analyzer SEOAnalyzer
}

// SEOConfig holds SEO configuration
type SEOConfig struct {
	DefaultLanguage    string            `json:"default_language"`
	SupportedLanguages []Language        `json:"supported_languages"`
	SitemapConfig      SitemapConfig     `json:"sitemap_config"`
	RobotsConfig       RobotsConfig      `json:"robots_config"`
	MetaDefaults       MetaDefaults      `json:"meta_defaults"`
	SchemaConfig       SchemaConfig      `json:"schema_config"`
	URLPatterns        map[string]string `json:"url_patterns"`
	RedirectRules      []RedirectRule    `json:"redirect_rules"`
}

// Language represents a supported language
type Language struct {
	Code       string `json:"code"`        // "tr", "en", "de"
	Name       string `json:"name"`        // "Türkçe", "English", "Deutsch"
	LocaleName string `json:"locale_name"` // "tr-TR", "en-US", "de-DE"
	Direction  string `json:"direction"`   // "ltr", "rtl"
	Enabled    bool   `json:"enabled"`
	IsDefault  bool   `json:"is_default"`
	URLPrefix  string `json:"url_prefix"`  // "/tr", "/en", "/de"
	Currency   string `json:"currency"`    // "TRY", "USD", "EUR"
	DateFormat string `json:"date_format"` // "02.01.2006", "01/02/2006"
	TimeZone   string `json:"timezone"`    // "Europe/Istanbul", "America/New_York"
}

// SitemapConfig holds sitemap configuration
type SitemapConfig struct {
	Enabled            bool            `json:"enabled"`
	MaxURLsPerFile     int             `json:"max_urls_per_file"`
	UpdateFrequency    string          `json:"update_frequency"`
	Priority           float64         `json:"priority"`
	IncludeImages      bool            `json:"include_images"`
	IncludeVideos      bool            `json:"include_videos"`
	IncludeNews        bool            `json:"include_news"`
	ExcludePatterns    []string        `json:"exclude_patterns"`
	CustomSitemaps     []CustomSitemap `json:"custom_sitemaps"`
	CompressionEnabled bool            `json:"compression_enabled"`
}

// CustomSitemap represents custom sitemap configuration
type CustomSitemap struct {
	Name       string   `json:"name"`
	URLPattern string   `json:"url_pattern"`
	DataSource string   `json:"data_source"`
	UpdateFreq string   `json:"update_freq"`
	Priority   float64  `json:"priority"`
	Languages  []string `json:"languages"`
}

// RobotsConfig holds robots.txt configuration
type RobotsConfig struct {
	Enabled     bool            `json:"enabled"`
	UserAgents  []UserAgentRule `json:"user_agents"`
	SitemapURLs []string        `json:"sitemap_urls"`
	CrawlDelay  int             `json:"crawl_delay"`
	CustomRules []string        `json:"custom_rules"`
}

// UserAgentRule represents robots.txt user agent rules
type UserAgentRule struct {
	UserAgent  string   `json:"user_agent"`
	Allow      []string `json:"allow"`
	Disallow   []string `json:"disallow"`
	CrawlDelay int      `json:"crawl_delay"`
}

// MetaDefaults holds default meta tag values
type MetaDefaults struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Keywords    string            `json:"keywords"`
	Author      string            `json:"author"`
	Publisher   string            `json:"publisher"`
	Copyright   string            `json:"copyright"`
	Robots      string            `json:"robots"`
	Viewport    string            `json:"viewport"`
	CharSet     string            `json:"charset"`
	Language    string            `json:"language"`
	OpenGraph   OpenGraphDefaults `json:"open_graph"`
	TwitterCard TwitterDefaults   `json:"twitter_card"`
	CustomMeta  map[string]string `json:"custom_meta"`
}

// OpenGraphDefaults holds Open Graph default values
type OpenGraphDefaults struct {
	Type        string `json:"type"`
	SiteName    string `json:"site_name"`
	Image       string `json:"image"`
	ImageWidth  int    `json:"image_width"`
	ImageHeight int    `json:"image_height"`
	Locale      string `json:"locale"`
}

// TwitterDefaults holds Twitter Card default values
type TwitterDefaults struct {
	Card    string `json:"card"`
	Site    string `json:"site"`
	Creator string `json:"creator"`
	Image   string `json:"image"`
}

// SchemaConfig holds Schema.org configuration
type SchemaConfig struct {
	Enabled        bool                   `json:"enabled"`
	Organization   OrganizationSchema     `json:"organization"`
	WebSite        WebSiteSchema          `json:"website"`
	BreadcrumbList BreadcrumbListSchema   `json:"breadcrumb_list"`
	Product        ProductSchema          `json:"product"`
	Article        ArticleSchema          `json:"article"`
	CustomSchemas  map[string]interface{} `json:"custom_schemas"`
}

// Schema structures
type OrganizationSchema struct {
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Logo        string    `json:"logo"`
	ContactInfo []Contact `json:"contact_info"`
	SameAs      []string  `json:"same_as"`
}

type Contact struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type WebSiteSchema struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	Description     string `json:"description"`
	PotentialAction string `json:"potential_action"`
}

type BreadcrumbListSchema struct {
	Enabled bool `json:"enabled"`
}

type ProductSchema struct {
	Enabled bool `json:"enabled"`
}

type ArticleSchema struct {
	Enabled bool `json:"enabled"`
}

// RedirectRule represents URL redirect rules
type RedirectRule struct {
	From       string `json:"from"`
	To         string `json:"to"`
	StatusCode int    `json:"status_code"`
	Enabled    bool   `json:"enabled"`
}

// SEOPage represents a page with SEO data
type SEOPage struct {
	ID             int                    `json:"id"`
	URL            string                 `json:"url"`
	Language       string                 `json:"language"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Keywords       string                 `json:"keywords"`
	H1             string                 `json:"h1"`
	H2             []string               `json:"h2"`
	H3             []string               `json:"h3"`
	MetaTags       map[string]string      `json:"meta_tags"`
	OpenGraph      map[string]string      `json:"open_graph"`
	TwitterCard    map[string]string      `json:"twitter_card"`
	CanonicalURL   string                 `json:"canonical_url"`
	AlternateURLs  map[string]string      `json:"alternate_urls"`
	Schema         map[string]interface{} `json:"schema"`
	Images         []SEOImage             `json:"images"`
	InternalLinks  []string               `json:"internal_links"`
	ExternalLinks  []string               `json:"external_links"`
	WordCount      int                    `json:"word_count"`
	ReadingTime    int                    `json:"reading_time"`
	LastModified   time.Time              `json:"last_modified"`
	IndexStatus    string                 `json:"index_status"`
	SitemapInclude bool                   `json:"sitemap_include"`
	Priority       float64                `json:"priority"`
	ChangeFreq     string                 `json:"change_freq"`
	SEOScore       float64                `json:"seo_score"`
	Issues         []SEOIssue             `json:"issues"`
	Suggestions    []SEOSuggestion        `json:"suggestions"`
}

// SEOImage represents an image with SEO data
type SEOImage struct {
	URL      string `json:"url"`
	Alt      string `json:"alt"`
	Title    string `json:"title"`
	Caption  string `json:"caption"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	FileSize int    `json:"file_size"`
}

// SEOIssue represents an SEO issue
type SEOIssue struct {
	Type       string `json:"type"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	Element    string `json:"element"`
	Suggestion string `json:"suggestion"`
}

// SEOSuggestion represents an SEO improvement suggestion
type SEOSuggestion struct {
	Type           string `json:"type"`
	Priority       string `json:"priority"`
	Message        string `json:"message"`
	Action         string `json:"action"`
	ExpectedImpact string `json:"expected_impact"`
}

// SitemapURL represents a URL in sitemap
type SitemapURL struct {
	Loc        string             `xml:"loc"`
	LastMod    string             `xml:"lastmod,omitempty"`
	ChangeFreq string             `xml:"changefreq,omitempty"`
	Priority   float64            `xml:"priority,omitempty"`
	Images     []SitemapImage     `xml:"image:image,omitempty"`
	Videos     []SitemapVideo     `xml:"video:video,omitempty"`
	News       *SitemapNews       `xml:"news:news,omitempty"`
	Alternates []SitemapAlternate `xml:"xhtml:link,omitempty"`
}

// SitemapImage represents an image in sitemap
type SitemapImage struct {
	XMLName xml.Name `xml:"image:image"`
	Loc     string   `xml:"image:loc"`
	Caption string   `xml:"image:caption,omitempty"`
	Title   string   `xml:"image:title,omitempty"`
}

// SitemapVideo represents a video in sitemap
type SitemapVideo struct {
	XMLName      xml.Name `xml:"video:video"`
	ThumbnailLoc string   `xml:"video:thumbnail_loc"`
	Title        string   `xml:"video:title"`
	Description  string   `xml:"video:description"`
	ContentLoc   string   `xml:"video:content_loc,omitempty"`
	PlayerLoc    string   `xml:"video:player_loc,omitempty"`
	Duration     int      `xml:"video:duration,omitempty"`
}

// SitemapNews represents news in sitemap
type SitemapNews struct {
	XMLName     xml.Name        `xml:"news:news"`
	Publication NewsPublication `xml:"news:publication"`
	Title       string          `xml:"news:title"`
	PublishDate string          `xml:"news:publication_date"`
	Keywords    string          `xml:"news:keywords,omitempty"`
}

// NewsPublication represents news publication info
type NewsPublication struct {
	Name     string `xml:"news:name"`
	Language string `xml:"news:language"`
}

// SitemapAlternate represents alternate language URLs
type SitemapAlternate struct {
	XMLName  xml.Name `xml:"xhtml:link"`
	Rel      string   `xml:"rel,attr"`
	Hreflang string   `xml:"hreflang,attr"`
	Href     string   `xml:"href,attr"`
}

// Sitemap represents the main sitemap structure
type Sitemap struct {
	XMLName    xml.Name     `xml:"urlset"`
	Xmlns      string       `xml:"xmlns,attr"`
	XmlnsImage string       `xml:"xmlns:image,attr,omitempty"`
	XmlnsVideo string       `xml:"xmlns:video,attr,omitempty"`
	XmlnsNews  string       `xml:"xmlns:news,attr,omitempty"`
	XmlnsXhtml string       `xml:"xmlns:xhtml,attr,omitempty"`
	URLs       []SitemapURL `xml:"url"`
}

// SitemapIndex represents sitemap index
type SitemapIndex struct {
	XMLName  xml.Name            `xml:"sitemapindex"`
	Xmlns    string              `xml:"xmlns,attr"`
	Sitemaps []SitemapIndexEntry `xml:"sitemap"`
}

// SitemapIndexEntry represents an entry in sitemap index
type SitemapIndexEntry struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

// SEOAnalyzer analyzes SEO metrics
type SEOAnalyzer interface {
	AnalyzePage(url string, content string) (*SEOPage, error)
	CalculateSEOScore(page *SEOPage) float64
	FindIssues(page *SEOPage) []SEOIssue
	GenerateSuggestions(page *SEOPage) []SEOSuggestion
}

// NewSEOManager creates a new SEO manager
func NewSEOManager(db *sql.DB, config SEOConfig) *SEOManager {
	sm := &SEOManager{
		db:     db,
		config: config,
	}

	sm.createSEOTables()
	return sm
}

// OptimizeTitle optimizes a title for SEO
func (sm *SEOManager) OptimizeTitle(title string) string {
	// Simple title optimization - in production would be more sophisticated
	if len(title) > 60 {
		return title[:57] + "..."
	}
	return title
}

// OptimizeDescription optimizes a description for SEO
func (sm *SEOManager) OptimizeDescription(description string) string {
	// Simple description optimization
	if len(description) > 160 {
		return description[:157] + "..."
	}
	return description
}

// GenerateSchema generates schema markup
func (sm *SEOManager) GenerateSchema(pageType string, data interface{}) (string, error) {
	// Simple schema generation - in production would be more comprehensive
	return `{"@context":"http://schema.org","@type":"WebPage"}`, nil
}

// createSEOTables creates necessary tables for SEO management
func (sm *SEOManager) createSEOTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS seo_pages (
			id INT AUTO_INCREMENT PRIMARY KEY,
			url VARCHAR(512) NOT NULL,
			language VARCHAR(10) NOT NULL,
			title VARCHAR(255),
			description TEXT,
			keywords TEXT,
			h1 VARCHAR(255),
			h2_tags TEXT,
			h3_tags TEXT,
			meta_tags TEXT,
			open_graph TEXT,
			twitter_card TEXT,
			canonical_url VARCHAR(512),
			alternate_urls TEXT,
			schema_data TEXT,
			images TEXT,
			internal_links TEXT,
			external_links TEXT,
			word_count INT DEFAULT 0,
			reading_time INT DEFAULT 0,
			last_modified DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			index_status VARCHAR(20) DEFAULT 'index',
			sitemap_include BOOLEAN DEFAULT TRUE,
			priority DECIMAL(2,1) DEFAULT 0.5,
			change_freq VARCHAR(20) DEFAULT 'weekly',
			seo_score DECIMAL(3,1) DEFAULT 0.0,
			issues TEXT,
			suggestions TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY unique_url_lang (url, language),
			INDEX idx_language (language),
			INDEX idx_last_modified (last_modified),
			INDEX idx_seo_score (seo_score)
		)`,
		`CREATE TABLE IF NOT EXISTS seo_redirects (
			id INT AUTO_INCREMENT PRIMARY KEY,
			from_url VARCHAR(512) NOT NULL,
			to_url VARCHAR(512) NOT NULL,
			status_code INT DEFAULT 301,
			enabled BOOLEAN DEFAULT TRUE,
			hit_count INT DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_from_url (from_url),
			INDEX idx_enabled (enabled)
		)`,
		`CREATE TABLE IF NOT EXISTS seo_translations (
			id INT AUTO_INCREMENT PRIMARY KEY,
			entity_type VARCHAR(50) NOT NULL,
			entity_id INT NOT NULL,
			language VARCHAR(10) NOT NULL,
			field_name VARCHAR(100) NOT NULL,
			field_value TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY unique_translation (entity_type, entity_id, language, field_name),
			INDEX idx_entity (entity_type, entity_id),
			INDEX idx_language (language)
		)`,
		`CREATE TABLE IF NOT EXISTS seo_analytics (
			id INT AUTO_INCREMENT PRIMARY KEY,
			url VARCHAR(512) NOT NULL,
			language VARCHAR(10) NOT NULL,
			date DATE NOT NULL,
			impressions INT DEFAULT 0,
			clicks INT DEFAULT 0,
			ctr DECIMAL(5,2) DEFAULT 0.00,
			average_position DECIMAL(5,2) DEFAULT 0.00,
			bounce_rate DECIMAL(5,2) DEFAULT 0.00,
			avg_session_duration INT DEFAULT 0,
			page_views INT DEFAULT 0,
			unique_visitors INT DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY unique_analytics (url, language, date),
			INDEX idx_date (date),
			INDEX idx_url_lang (url, language)
		)`,
	}

	for _, query := range queries {
		if _, err := sm.db.Exec(query); err != nil {
			return fmt.Errorf("failed to create SEO table: %w", err)
		}
	}

	return nil
}

// GenerateSitemap generates XML sitemap for specified language
func (sm *SEOManager) GenerateSitemap(language string) (*Sitemap, error) {
	sitemap := &Sitemap{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  make([]SitemapURL, 0),
	}

	// Add namespace declarations if needed
	if sm.config.SitemapConfig.IncludeImages {
		sitemap.XmlnsImage = "http://www.google.com/schemas/sitemap-image/1.1"
	}
	if sm.config.SitemapConfig.IncludeVideos {
		sitemap.XmlnsVideo = "http://www.google.com/schemas/sitemap-video/1.1"
	}
	if sm.config.SitemapConfig.IncludeNews {
		sitemap.XmlnsNews = "http://www.google.com/schemas/sitemap-news/0.9"
	}
	sitemap.XmlnsXhtml = "http://www.w3.org/1999/xhtml"

	// Get pages for sitemap
	pages, err := sm.getSitemapPages(language)
	if err != nil {
		return nil, err
	}

	// Generate URLs
	for _, page := range pages {
		if !page.SitemapInclude {
			continue
		}

		sitemapURL := SitemapURL{
			Loc:        page.URL,
			LastMod:    page.LastModified.Format("2006-01-02T15:04:05Z07:00"),
			ChangeFreq: page.ChangeFreq,
			Priority:   page.Priority,
		}

		// Add alternate language URLs
		alternates := make([]SitemapAlternate, 0)
		for lang, altURL := range page.AlternateURLs {
			if lang != language {
				alternates = append(alternates, SitemapAlternate{
					Rel:      "alternate",
					Hreflang: lang,
					Href:     altURL,
				})
			}
		}
		sitemapURL.Alternates = alternates

		// Add images if enabled
		if sm.config.SitemapConfig.IncludeImages {
			images := make([]SitemapImage, 0)
			for _, img := range page.Images {
				images = append(images, SitemapImage{
					Loc:     img.URL,
					Caption: img.Caption,
					Title:   img.Title,
				})
			}
			sitemapURL.Images = images
		}

		sitemap.URLs = append(sitemap.URLs, sitemapURL)
	}

	return sitemap, nil
}

// GenerateRobotsTxt generates robots.txt content
func (sm *SEOManager) GenerateRobotsTxt() string {
	if !sm.config.RobotsConfig.Enabled {
		return ""
	}

	var robots strings.Builder

	// Add user agent rules
	for _, rule := range sm.config.RobotsConfig.UserAgents {
		robots.WriteString(fmt.Sprintf("User-agent: %s\n", rule.UserAgent))

		// Add allow rules
		for _, allow := range rule.Allow {
			robots.WriteString(fmt.Sprintf("Allow: %s\n", allow))
		}

		// Add disallow rules
		for _, disallow := range rule.Disallow {
			robots.WriteString(fmt.Sprintf("Disallow: %s\n", disallow))
		}

		// Add crawl delay if specified
		if rule.CrawlDelay > 0 {
			robots.WriteString(fmt.Sprintf("Crawl-delay: %d\n", rule.CrawlDelay))
		}

		robots.WriteString("\n")
	}

	// Add global crawl delay
	if sm.config.RobotsConfig.CrawlDelay > 0 {
		robots.WriteString(fmt.Sprintf("Crawl-delay: %d\n", sm.config.RobotsConfig.CrawlDelay))
	}

	// Add sitemap URLs
	for _, sitemapURL := range sm.config.RobotsConfig.SitemapURLs {
		robots.WriteString(fmt.Sprintf("Sitemap: %s\n", sitemapURL))
	}

	// Add custom rules
	for _, customRule := range sm.config.RobotsConfig.CustomRules {
		robots.WriteString(customRule + "\n")
	}

	return robots.String()
}

// GetTranslation retrieves translation for entity
func (sm *SEOManager) GetTranslation(entityType string, entityID int, language string, fieldName string) (string, error) {
	query := `
		SELECT field_value FROM seo_translations 
		WHERE entity_type = ? AND entity_id = ? AND language = ? AND field_name = ?
	`

	var value string
	err := sm.db.QueryRow(query, entityType, entityID, language, fieldName).Scan(&value)
	if err != nil {
		return "", err
	}

	return value, nil
}

// SetTranslation sets translation for entity
func (sm *SEOManager) SetTranslation(entityType string, entityID int, language string, fieldName string, value string) error {
	query := `
		INSERT INTO seo_translations (entity_type, entity_id, language, field_name, field_value)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE field_value = VALUES(field_value), updated_at = NOW()
	`

	_, err := sm.db.Exec(query, entityType, entityID, language, fieldName, value)
	return err
}

// GetAllTranslations retrieves all translations for entity
func (sm *SEOManager) GetAllTranslations(entityType string, entityID int) (map[string]map[string]string, error) {
	query := `
		SELECT language, field_name, field_value FROM seo_translations 
		WHERE entity_type = ? AND entity_id = ?
	`

	rows, err := sm.db.Query(query, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	translations := make(map[string]map[string]string)

	for rows.Next() {
		var language, fieldName, fieldValue string
		if err := rows.Scan(&language, &fieldName, &fieldValue); err != nil {
			continue
		}

		if translations[language] == nil {
			translations[language] = make(map[string]string)
		}
		translations[language][fieldName] = fieldValue
	}

	return translations, nil
}

// GenerateAlternateURLs generates alternate language URLs for a page
func (sm *SEOManager) GenerateAlternateURLs(baseURL string, currentLanguage string) map[string]string {
	alternates := make(map[string]string)

	for _, lang := range sm.config.SupportedLanguages {
		if !lang.Enabled || lang.Code == currentLanguage {
			continue
		}

		// Generate alternate URL based on URL pattern
		alternateURL := sm.generateLanguageURL(baseURL, lang.Code)
		alternates[lang.Code] = alternateURL
	}

	return alternates
}

// generateLanguageURL generates URL for specific language
func (sm *SEOManager) generateLanguageURL(baseURL string, language string) string {
	// Find language config
	var langConfig Language
	for _, lang := range sm.config.SupportedLanguages {
		if lang.Code == language {
			langConfig = lang
			break
		}
	}

	// Parse base URL
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	// Add language prefix if not default language
	if !langConfig.IsDefault && langConfig.URLPrefix != "" {
		parsedURL.Path = langConfig.URLPrefix + parsedURL.Path
	}

	return parsedURL.String()
}

// OptimizePage analyzes and optimizes a page for SEO
func (sm *SEOManager) OptimizePage(pageURL string, content string, language string) (*SEOPage, error) {
	// Analyze page content
	page, err := sm.analyzer.AnalyzePage(pageURL, content)
	if err != nil {
		return nil, err
	}

	page.Language = language

	// Calculate SEO score
	page.SEOScore = sm.analyzer.CalculateSEOScore(page)

	// Find issues
	page.Issues = sm.analyzer.FindIssues(page)

	// Generate suggestions
	page.Suggestions = sm.analyzer.GenerateSuggestions(page)

	// Generate alternate URLs
	page.AlternateURLs = sm.GenerateAlternateURLs(pageURL, language)

	// Save to database
	err = sm.saveSEOPage(page)
	if err != nil {
		return nil, err
	}

	return page, nil
}

// getSitemapPages retrieves pages for sitemap generation
func (sm *SEOManager) getSitemapPages(language string) ([]*SEOPage, error) {
	query := `
		SELECT url, title, description, canonical_url, alternate_urls, 
		       priority, change_freq, last_modified, sitemap_include
		FROM seo_pages 
		WHERE language = ? AND sitemap_include = TRUE AND index_status = 'index'
		ORDER BY priority DESC, last_modified DESC
	`

	rows, err := sm.db.Query(query, language)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pages := make([]*SEOPage, 0)

	for rows.Next() {
		page := &SEOPage{}
		var alternateURLsJSON string

		err := rows.Scan(
			&page.URL, &page.Title, &page.Description, &page.CanonicalURL,
			&alternateURLsJSON, &page.Priority, &page.ChangeFreq,
			&page.LastModified, &page.SitemapInclude,
		)
		if err != nil {
			continue
		}

		// Parse alternate URLs
		if alternateURLsJSON != "" {
			json.Unmarshal([]byte(alternateURLsJSON), &page.AlternateURLs)
		}

		pages = append(pages, page)
	}

	return pages, nil
}

// saveSEOPage saves SEO page data to database
func (sm *SEOManager) saveSEOPage(page *SEOPage) error {
	// Convert complex fields to JSON
	metaTagsJSON, _ := json.Marshal(page.MetaTags)
	openGraphJSON, _ := json.Marshal(page.OpenGraph)
	twitterCardJSON, _ := json.Marshal(page.TwitterCard)
	alternateURLsJSON, _ := json.Marshal(page.AlternateURLs)
	schemaJSON, _ := json.Marshal(page.Schema)
	imagesJSON, _ := json.Marshal(page.Images)
	h2JSON, _ := json.Marshal(page.H2)
	h3JSON, _ := json.Marshal(page.H3)
	internalLinksJSON, _ := json.Marshal(page.InternalLinks)
	externalLinksJSON, _ := json.Marshal(page.ExternalLinks)
	issuesJSON, _ := json.Marshal(page.Issues)
	suggestionsJSON, _ := json.Marshal(page.Suggestions)

	query := `
		INSERT INTO seo_pages (
			url, language, title, description, keywords, h1, h2_tags, h3_tags,
			meta_tags, open_graph, twitter_card, canonical_url, alternate_urls,
			schema_data, images, internal_links, external_links, word_count,
			reading_time, index_status, sitemap_include, priority, change_freq,
			seo_score, issues, suggestions
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		title = VALUES(title), description = VALUES(description),
		keywords = VALUES(keywords), h1 = VALUES(h1), h2_tags = VALUES(h2_tags),
		h3_tags = VALUES(h3_tags), meta_tags = VALUES(meta_tags),
		open_graph = VALUES(open_graph), twitter_card = VALUES(twitter_card),
		canonical_url = VALUES(canonical_url), alternate_urls = VALUES(alternate_urls),
		schema_data = VALUES(schema_data), images = VALUES(images),
		internal_links = VALUES(internal_links), external_links = VALUES(external_links),
		word_count = VALUES(word_count), reading_time = VALUES(reading_time),
		index_status = VALUES(index_status), sitemap_include = VALUES(sitemap_include),
		priority = VALUES(priority), change_freq = VALUES(change_freq),
		seo_score = VALUES(seo_score), issues = VALUES(issues),
		suggestions = VALUES(suggestions), updated_at = NOW()
	`

	_, err := sm.db.Exec(query,
		page.URL, page.Language, page.Title, page.Description, page.Keywords,
		page.H1, string(h2JSON), string(h3JSON), string(metaTagsJSON),
		string(openGraphJSON), string(twitterCardJSON), page.CanonicalURL,
		string(alternateURLsJSON), string(schemaJSON), string(imagesJSON),
		string(internalLinksJSON), string(externalLinksJSON), page.WordCount,
		page.ReadingTime, page.IndexStatus, page.SitemapInclude, page.Priority,
		page.ChangeFreq, page.SEOScore, string(issuesJSON), string(suggestionsJSON),
	)

	return err
}

// GetSEOAnalytics retrieves SEO analytics data
func (sm *SEOManager) GetSEOAnalytics(url string, language string, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	query := `
		SELECT date, impressions, clicks, ctr, average_position, bounce_rate,
		       avg_session_duration, page_views, unique_visitors
		FROM seo_analytics 
		WHERE url = ? AND language = ? AND date BETWEEN ? AND ?
		ORDER BY date ASC
	`

	rows, err := sm.db.Query(query, url, language, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	analytics := make([]map[string]interface{}, 0)

	for rows.Next() {
		var date time.Time
		var impressions, clicks, avgSessionDuration, pageViews, uniqueVisitors int
		var ctr, avgPosition, bounceRate float64

		err := rows.Scan(&date, &impressions, &clicks, &ctr, &avgPosition,
			&bounceRate, &avgSessionDuration, &pageViews, &uniqueVisitors)
		if err != nil {
			continue
		}

		analytics = append(analytics, map[string]interface{}{
			"date":                 date.Format("2006-01-02"),
			"impressions":          impressions,
			"clicks":               clicks,
			"ctr":                  ctr,
			"average_position":     avgPosition,
			"bounce_rate":          bounceRate,
			"avg_session_duration": avgSessionDuration,
			"page_views":           pageViews,
			"unique_visitors":      uniqueVisitors,
		})
	}

	return analytics, nil
}

// GenerateHreflangTags generates hreflang tags for a page
func (sm *SEOManager) GenerateHreflangTags(baseURL string, currentLanguage string) []map[string]string {
	tags := make([]map[string]string, 0)

	// Add current language
	for _, lang := range sm.config.SupportedLanguages {
		if lang.Code == currentLanguage {
			tags = append(tags, map[string]string{
				"rel":      "alternate",
				"hreflang": lang.LocaleName,
				"href":     baseURL,
			})
			break
		}
	}

	// Add alternate languages
	alternateURLs := sm.GenerateAlternateURLs(baseURL, currentLanguage)
	for langCode, altURL := range alternateURLs {
		for _, lang := range sm.config.SupportedLanguages {
			if lang.Code == langCode {
				tags = append(tags, map[string]string{
					"rel":      "alternate",
					"hreflang": lang.LocaleName,
					"href":     altURL,
				})
				break
			}
		}
	}

	// Add x-default for default language
	defaultLang := sm.getDefaultLanguage()
	if defaultLang != nil {
		defaultURL := sm.generateLanguageURL(baseURL, defaultLang.Code)
		tags = append(tags, map[string]string{
			"rel":      "alternate",
			"hreflang": "x-default",
			"href":     defaultURL,
		})
	}

	return tags
}

// getDefaultLanguage returns the default language configuration
func (sm *SEOManager) getDefaultLanguage() *Language {
	for _, lang := range sm.config.SupportedLanguages {
		if lang.IsDefault {
			return &lang
		}
	}
	return nil
}

// ValidateURL validates if URL follows SEO best practices
func (sm *SEOManager) ValidateURL(url string) []SEOIssue {
	issues := make([]SEOIssue, 0)

	// Check URL length
	if len(url) > 255 {
		issues = append(issues, SEOIssue{
			Type:       "url_length",
			Severity:   "medium",
			Message:    "URL too long (over 255 characters)",
			Suggestion: "Consider shortening the URL for better SEO",
		})
	}

	// Check for special characters
	if matched, _ := regexp.MatchString(`[^a-zA-Z0-9\-_/.]`, url); matched {
		issues = append(issues, SEOIssue{
			Type:       "url_characters",
			Severity:   "low",
			Message:    "URL contains special characters",
			Suggestion: "Use only alphanumeric characters, hyphens, and underscores",
		})
	}

	// Check for uppercase letters
	if matched, _ := regexp.MatchString(`[A-Z]`, url); matched {
		issues = append(issues, SEOIssue{
			Type:       "url_case",
			Severity:   "low",
			Message:    "URL contains uppercase letters",
			Suggestion: "Use lowercase letters for better consistency",
		})
	}

	return issues
}

// GenerateStructuredData generates Schema.org structured data
func (sm *SEOManager) GenerateStructuredData(dataType string, data map[string]interface{}) map[string]interface{} {
	schema := map[string]interface{}{
		"@context": "http://schema.org",
		"@type":    dataType,
	}

	// Add data fields
	for key, value := range data {
		schema[key] = value
	}

	return schema
}
