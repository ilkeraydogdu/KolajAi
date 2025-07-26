package email

import "time"

// EmailType represents different types of emails
type EmailType string

const (
	EmailTypeWelcome         EmailType = "welcome"
	EmailTypePasswordReset   EmailType = "password_reset"
	EmailTypeVerification    EmailType = "verification"
	EmailTypePasswordChanged EmailType = "password_changed"
	EmailTypeInvoice         EmailType = "invoice"
	EmailTypeNotification    EmailType = "notification"
	EmailTypeMarketing       EmailType = "marketing"
)

// EmailPriority represents email priority
type EmailPriority int

const (
	PriorityLow    EmailPriority = 1
	PriorityNormal EmailPriority = 2
	PriorityHigh   EmailPriority = 3
)

// Attachment represents a file attachment for an email
type Attachment struct {
	Filename string
	Content  []byte
	MIMEType string
}

// SocialLink represents a social media link
type SocialLink struct {
	Name string
	URL  string
	Last bool
}

// ActionButton represents a call to action button
type ActionButton struct {
	Text string
	URL  string
	Type string // primary, success, danger, etc.
}

// AlertBox represents an alert box in the email
type AlertBox struct {
	Type    string // success, danger, warning, info
	Title   string
	Content string
}

// EmailData represents data for an email
type EmailData struct {
	// Core fields
	Type     EmailType
	To       []string
	CC       []string
	BCC      []string
	Subject  string
	Priority EmailPriority
	Name     string

	// Header customization
	CompanyName string
	HeaderLogo  string
	HeaderBg    string

	// Content fields
	Title            string
	Greeting         string
	Paragraphs       []string
	Features         []string
	FeatureIntro     string
	Alert            *AlertBox
	PrimaryAction    *ActionButton
	SecondaryContent interface{} // HTML içerik olarak kullanılabilir

	// Footer customization
	SupportEmail    string
	Signature       string
	SocialLinks     []SocialLink
	UnsubscribeLink string
	PrivacyLink     string

	// Advanced options
	CustomCSS   string
	Attachments []Attachment
	SendAt      time.Time
	Metadata    map[string]string
}

// NewEmailData creates a new EmailData with defaults
func NewEmailData(emailType EmailType, to []string, subject, name string) *EmailData {
	return &EmailData{
		Type:         emailType,
		To:           to,
		Subject:      subject,
		Name:         name,
		Priority:     PriorityNormal,
		CompanyName:  "KolajAI | Pofuduk Dijital",
		SupportEmail: "destek@kolaj.ai",
	}
}

// SetAlert adds an alert box to the email
func (d *EmailData) SetAlert(alertType, title, content string) *EmailData {
	d.Alert = &AlertBox{
		Type:    alertType,
		Title:   title,
		Content: content,
	}
	return d
}

// AddParagraph adds a paragraph to the email content
func (d *EmailData) AddParagraph(paragraph string) *EmailData {
	d.Paragraphs = append(d.Paragraphs, paragraph)
	return d
}

// AddFeature adds a feature bullet point to the email
func (d *EmailData) AddFeature(feature string) *EmailData {
	d.Features = append(d.Features, feature)
	return d
}

// SetPrimaryAction sets the primary call-to-action button
func (d *EmailData) SetPrimaryAction(text, url, buttonType string) *EmailData {
	d.PrimaryAction = &ActionButton{
		Text: text,
		URL:  url,
		Type: buttonType,
	}
	return d
}

// AddSocialLink adds a social media link to the email footer
func (d *EmailData) AddSocialLink(name, url string) *EmailData {
	d.SocialLinks = append(d.SocialLinks, SocialLink{
		Name: name,
		URL:  url,
		Last: false,
	})
	// Mark the last one as last
	if len(d.SocialLinks) > 0 {
		for i := range d.SocialLinks {
			d.SocialLinks[i].Last = (i == len(d.SocialLinks)-1)
		}
	}
	return d
}

// AddAttachment adds an attachment to the email
func (d *EmailData) AddAttachment(filename string, content []byte, mimeType string) *EmailData {
	d.Attachments = append(d.Attachments, Attachment{
		Filename: filename,
		Content:  content,
		MIMEType: mimeType,
	})
	return d
}

// AddMetadata adds a metadata key-value pair
func (d *EmailData) AddMetadata(key, value string) *EmailData {
	if d.Metadata == nil {
		d.Metadata = make(map[string]string)
	}
	d.Metadata[key] = value
	return d
}

// GetCurrentYear returns the current year for copyright
func (d *EmailData) GetCurrentYear() int {
	return time.Now().Year()
}
