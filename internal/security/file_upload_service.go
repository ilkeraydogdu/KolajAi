package security

import (
	"crypto/md5"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// FileUploadService handles secure file upload operations
type FileUploadService struct {
	db            *sql.DB
	config        FileUploadConfig
	storageDir    string
	quarantineDir string
}

// FileUploadConfig holds file upload security configuration
type FileUploadConfig struct {
	MaxFileSize       int64             `json:"max_file_size"`       // bytes
	AllowedTypes      []string          `json:"allowed_types"`       // MIME types
	AllowedExtensions []string          `json:"allowed_extensions"`  // file extensions
	BlockedTypes      []string          `json:"blocked_types"`       // blocked MIME types
	BlockedExtensions []string          `json:"blocked_extensions"`  // blocked extensions
	ScanForViruses    bool              `json:"scan_for_viruses"`
	ScanForMalware    bool              `json:"scan_for_malware"`
	CheckMagicBytes   bool              `json:"check_magic_bytes"`
	RenameFiles       bool              `json:"rename_files"`
	CreateThumbnails  bool              `json:"create_thumbnails"`
	MaxDimensions     ImageDimensions   `json:"max_dimensions"`
	CompressionRules  CompressionRules  `json:"compression_rules"`
	StorageRules      StorageRules      `json:"storage_rules"`
}

// ImageDimensions represents maximum image dimensions
type ImageDimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// CompressionRules represents file compression rules
type CompressionRules struct {
	CompressImages bool    `json:"compress_images"`
	ImageQuality   int     `json:"image_quality"`   // 1-100
	CompressRatio  float64 `json:"compress_ratio"`  // 0.1-1.0
}

// StorageRules represents file storage rules
type StorageRules struct {
	UseSecureNaming   bool   `json:"use_secure_naming"`
	DirectoryStructure string `json:"directory_structure"` // "date", "user", "type"
	EncryptFiles      bool   `json:"encrypt_files"`
	BackupFiles       bool   `json:"backup_files"`
}

// UploadResult represents file upload result
type UploadResult struct {
	Success      bool                   `json:"success"`
	FileID       string                 `json:"file_id,omitempty"`
	OriginalName string                 `json:"original_name"`
	SecureName   string                 `json:"secure_name"`
	MimeType     string                 `json:"mime_type"`
	Size         int64                  `json:"size"`
	Path         string                 `json:"path"`
	URL          string                 `json:"url,omitempty"`
	Checksum     string                 `json:"checksum"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Warnings     []string               `json:"warnings,omitempty"`
	Error        string                 `json:"error,omitempty"`
	ScanResults  *ScanResults           `json:"scan_results,omitempty"`
}

// ScanResults represents security scan results
type ScanResults struct {
	VirusScan   *FileScanResult `json:"virus_scan,omitempty"`
	MalwareScan *FileScanResult `json:"malware_scan,omitempty"`
	SafetyScore int             `json:"safety_score"` // 0-100
}

// FileScanResult represents individual scan result
type FileScanResult struct {
	Clean       bool      `json:"clean"`
	Threats     []string  `json:"threats,omitempty"`
	Scanner     string    `json:"scanner"`
	ScannedAt   time.Time `json:"scanned_at"`
	ScanTime    int64     `json:"scan_time_ms"`
}

// FileRecord represents uploaded file record
type FileRecord struct {
	ID           string                 `json:"id"`
	UserID       int64                  `json:"user_id"`
	OriginalName string                 `json:"original_name"`
	SecureName   string                 `json:"secure_name"`
	MimeType     string                 `json:"mime_type"`
	Size         int64                  `json:"size"`
	Path         string                 `json:"path"`
	Checksum     string                 `json:"checksum"`
	Status       string                 `json:"status"` // "uploaded", "scanned", "approved", "quarantined", "deleted"
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	UploadedAt   time.Time              `json:"uploaded_at"`
	ScannedAt    *time.Time             `json:"scanned_at,omitempty"`
	ApprovedAt   *time.Time             `json:"approved_at,omitempty"`
}

// NewFileUploadService creates a new file upload service
func NewFileUploadService(db *sql.DB, storageDir, quarantineDir string, config FileUploadConfig) *FileUploadService {
	return &FileUploadService{
		db:            db,
		config:        config,
		storageDir:    storageDir,
		quarantineDir: quarantineDir,
	}
}

// ValidateFile validates file before upload
func (f *FileUploadService) ValidateFile(filename string, size int64, content io.Reader) error {
	// Check file size
	if size > f.config.MaxFileSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size %d", size, f.config.MaxFileSize)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(filename))
	if !f.isExtensionAllowed(ext) {
		return fmt.Errorf("file extension %s is not allowed", ext)
	}

	if f.isExtensionBlocked(ext) {
		return fmt.Errorf("file extension %s is blocked", ext)
	}

	// Check MIME type
	if f.config.CheckMagicBytes {
		buffer := make([]byte, 512)
		n, err := content.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read file content: %w", err)
		}

		mimeType := http.DetectContentType(buffer[:n])
		if !f.isMimeTypeAllowed(mimeType) {
			return fmt.Errorf("MIME type %s is not allowed", mimeType)
		}

		if f.isMimeTypeBlocked(mimeType) {
			return fmt.Errorf("MIME type %s is blocked", mimeType)
		}

		// Reset reader if possible
		if seeker, ok := content.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}
	}

	return nil
}

// UploadFile handles secure file upload
func (f *FileUploadService) UploadFile(userID int64, filename string, content io.Reader) (*UploadResult, error) {
	result := &UploadResult{
		OriginalName: filename,
	}

	// Read file content
	data, err := io.ReadAll(content)
	if err != nil {
		result.Error = fmt.Sprintf("failed to read file content: %v", err)
		return result, err
	}

	// Validate file
	if err := f.ValidateFile(filename, int64(len(data)), strings.NewReader(string(data))); err != nil {
		result.Error = err.Error()
		return result, err
	}

	// Generate secure filename
	secureName := f.generateSecureFilename(filename)
	result.SecureName = secureName

	// Determine storage path
	storagePath := f.generateStoragePath(userID, secureName)
	result.Path = storagePath

	// Calculate checksums
	_ = md5.Sum(data) // Keep for future use
	sha256Hash := sha256.Sum256(data)
	result.Checksum = hex.EncodeToString(sha256Hash[:])

	// Detect MIME type
	mimeType := http.DetectContentType(data)
	result.MimeType = mimeType
	result.Size = int64(len(data))

	// Create directory if not exists
	dir := filepath.Dir(storagePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		result.Error = fmt.Sprintf("failed to create storage directory: %v", err)
		return result, err
	}

	// Write file to storage
	if err := os.WriteFile(storagePath, data, 0644); err != nil {
		result.Error = fmt.Sprintf("failed to write file: %v", err)
		return result, err
	}

	// Generate file ID
	fileID := f.generateFileID()
	result.FileID = fileID

	// Extract metadata
	metadata := f.extractMetadata(data, mimeType)
	result.Metadata = metadata

	// Perform security scans
	if f.config.ScanForViruses || f.config.ScanForMalware {
		scanResults, err := f.performSecurityScan(storagePath)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Security scan failed: %v", err))
		} else {
			result.ScanResults = scanResults
			
			// Check if file is safe
			if !scanResults.VirusScan.Clean || !scanResults.MalwareScan.Clean {
				// Move to quarantine
				quarantinePath := filepath.Join(f.quarantineDir, secureName)
				if err := os.Rename(storagePath, quarantinePath); err != nil {
					result.Error = fmt.Sprintf("failed to quarantine infected file: %v", err)
					return result, err
				}
				result.Error = "file contains threats and has been quarantined"
				return result, errors.New("file contains threats")
			}
		}
	}

	// Save file record to database
	fileRecord := &FileRecord{
		ID:           fileID,
		UserID:       userID,
		OriginalName: filename,
		SecureName:   secureName,
		MimeType:     mimeType,
		Size:         int64(len(data)),
		Path:         storagePath,
		Checksum:     result.Checksum,
		Status:       "approved",
		Metadata:     metadata,
		UploadedAt:   time.Now(),
	}

	if err := f.saveFileRecord(fileRecord); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to save file record: %v", err))
	}

	result.Success = true
	return result, nil
}

// GetFile retrieves file information
func (f *FileUploadService) GetFile(fileID string) (*FileRecord, error) {
	var record FileRecord
	var metadataJSON sql.NullString

	err := f.db.QueryRow(`
		SELECT id, user_id, original_name, secure_name, mime_type, size, path, 
			   checksum, status, metadata, uploaded_at, scanned_at, approved_at
		FROM uploaded_files WHERE id = ?
	`, fileID).Scan(
		&record.ID, &record.UserID, &record.OriginalName, &record.SecureName,
		&record.MimeType, &record.Size, &record.Path, &record.Checksum,
		&record.Status, &metadataJSON, &record.UploadedAt,
		&record.ScannedAt, &record.ApprovedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get file record: %w", err)
	}

	// Parse metadata JSON
	if metadataJSON.Valid {
		// Parse JSON metadata
		// json.Unmarshal([]byte(metadataJSON.String), &record.Metadata)
	}

	return &record, nil
}

// DeleteFile deletes file and record
func (f *FileUploadService) DeleteFile(fileID string, userID int64) error {
	// Get file record
	record, err := f.GetFile(fileID)
	if err != nil {
		return err
	}

	// Check ownership
	if record.UserID != userID {
		return errors.New("unauthorized to delete this file")
	}

	// Delete physical file
	if err := os.Remove(record.Path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete physical file: %w", err)
	}

	// Delete database record
	_, err = f.db.Exec("DELETE FROM uploaded_files WHERE id = ?", fileID)
	if err != nil {
		return fmt.Errorf("failed to delete file record: %w", err)
	}

	return nil
}

// Helper methods

func (f *FileUploadService) isExtensionAllowed(ext string) bool {
	if len(f.config.AllowedExtensions) == 0 {
		return true // Allow all if no restrictions
	}

	for _, allowed := range f.config.AllowedExtensions {
		if strings.EqualFold(ext, allowed) {
			return true
		}
	}
	return false
}

func (f *FileUploadService) isExtensionBlocked(ext string) bool {
	for _, blocked := range f.config.BlockedExtensions {
		if strings.EqualFold(ext, blocked) {
			return true
		}
	}
	return false
}

func (f *FileUploadService) isMimeTypeAllowed(mimeType string) bool {
	if len(f.config.AllowedTypes) == 0 {
		return true // Allow all if no restrictions
	}

	for _, allowed := range f.config.AllowedTypes {
		if strings.HasPrefix(mimeType, allowed) {
			return true
		}
	}
	return false
}

func (f *FileUploadService) isMimeTypeBlocked(mimeType string) bool {
	for _, blocked := range f.config.BlockedTypes {
		if strings.HasPrefix(mimeType, blocked) {
			return true
		}
	}
	return false
}

func (f *FileUploadService) generateSecureFilename(originalName string) string {
	if !f.config.RenameFiles {
		// Sanitize original name
		return f.sanitizeFilename(originalName)
	}

	// Generate UUID-like name
	timestamp := time.Now().UnixNano()
	ext := filepath.Ext(originalName)
	
	// Create hash of original name + timestamp
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s_%d", originalName, timestamp)))
	name := hex.EncodeToString(hash[:8]) // Use first 8 bytes
	
	return name + ext
}

func (f *FileUploadService) sanitizeFilename(filename string) string {
	// Remove dangerous characters
	reg := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	sanitized := reg.ReplaceAllString(filename, "_")
	
	// Limit length
	if len(sanitized) > 255 {
		ext := filepath.Ext(sanitized)
		name := sanitized[:255-len(ext)]
		sanitized = name + ext
	}
	
	return sanitized
}

func (f *FileUploadService) generateStoragePath(userID int64, filename string) string {
	switch f.config.StorageRules.DirectoryStructure {
	case "date":
		now := time.Now()
		return filepath.Join(f.storageDir, 
			fmt.Sprintf("%d", now.Year()),
			fmt.Sprintf("%02d", now.Month()),
			fmt.Sprintf("%02d", now.Day()),
			filename)
	case "user":
		return filepath.Join(f.storageDir, 
			fmt.Sprintf("user_%d", userID),
			filename)
	case "type":
		ext := strings.ToLower(filepath.Ext(filename))
		return filepath.Join(f.storageDir, 
			strings.TrimPrefix(ext, "."),
			filename)
	default:
		return filepath.Join(f.storageDir, filename)
	}
}

func (f *FileUploadService) generateFileID() string {
	timestamp := time.Now().UnixNano()
	hash := sha256.Sum256([]byte(fmt.Sprintf("file_%d", timestamp)))
	return hex.EncodeToString(hash[:16])
}

func (f *FileUploadService) extractMetadata(data []byte, mimeType string) map[string]interface{} {
	metadata := make(map[string]interface{})
	
	metadata["size"] = len(data)
	metadata["mime_type"] = mimeType
	metadata["uploaded_at"] = time.Now()
	
	// Extract image metadata if it's an image
	if strings.HasPrefix(mimeType, "image/") {
		// TODO: Extract EXIF data, dimensions, etc.
		metadata["type"] = "image"
	}
	
	// Extract document metadata if it's a document
	if strings.HasPrefix(mimeType, "application/pdf") {
		metadata["type"] = "document"
		metadata["format"] = "pdf"
	}
	
	return metadata
}

func (f *FileUploadService) performSecurityScan(filePath string) (*ScanResults, error) {
	results := &ScanResults{
		SafetyScore: 100, // Start with perfect score
	}

	// Virus scan
	if f.config.ScanForViruses {
		virusScan := &FileScanResult{
			Clean:     true,
			Scanner:   "built-in",
			ScannedAt: time.Now(),
		}
		
		// TODO: Integrate with actual antivirus engine (ClamAV, etc.)
		// For now, just check for suspicious patterns
		if f.containsSuspiciousPatterns(filePath) {
			virusScan.Clean = false
			virusScan.Threats = []string{"suspicious_pattern_detected"}
			results.SafetyScore -= 50
		}
		
		results.VirusScan = virusScan
	}

	// Malware scan
	if f.config.ScanForMalware {
		malwareScan := &FileScanResult{
			Clean:     true,
			Scanner:   "built-in",
			ScannedAt: time.Now(),
		}
		
		// TODO: Integrate with malware detection engine
		// For now, just check file signatures
		if f.containsMalwareSignatures(filePath) {
			malwareScan.Clean = false
			malwareScan.Threats = []string{"malware_signature_detected"}
			results.SafetyScore -= 30
		}
		
		results.MalwareScan = malwareScan
	}

	return results, nil
}

func (f *FileUploadService) containsSuspiciousPatterns(filePath string) bool {
	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	content := string(data)
	
	// Check for suspicious patterns
	suspiciousPatterns := []string{
		"<script",
		"javascript:",
		"vbscript:",
		"onload=",
		"onerror=",
		"eval(",
		"document.write",
		"innerHTML",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(strings.ToLower(content), pattern) {
			return true
		}
	}

	return false
}

func (f *FileUploadService) containsMalwareSignatures(filePath string) bool {
	// Read first 1KB for signature checking
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}

	content := buffer[:n]
	
	// Check for known malware signatures (simplified)
	malwareSignatures := [][]byte{
		[]byte("MZ"), // PE executable header
		[]byte("\x7fELF"), // ELF executable header
	}

	// Only flag executables if they're not in allowed types
	for _, sig := range malwareSignatures {
		if len(content) >= len(sig) && string(content[:len(sig)]) == string(sig) {
			// Check if executables are allowed
			mimeType := http.DetectContentType(content)
			if !f.isMimeTypeAllowed(mimeType) {
				return true
			}
		}
	}

	return false
}

func (f *FileUploadService) saveFileRecord(record *FileRecord) error {
	// Convert metadata to JSON
	// metadataJSON, _ := json.Marshal(record.Metadata)

	_, err := f.db.Exec(`
		INSERT INTO uploaded_files (id, user_id, original_name, secure_name, mime_type, 
									size, path, checksum, status, metadata, uploaded_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, record.ID, record.UserID, record.OriginalName, record.SecureName,
		record.MimeType, record.Size, record.Path, record.Checksum,
		record.Status, "{}", record.UploadedAt)

	return err
}