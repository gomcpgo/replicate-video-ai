package storage

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// Storage handles file operations for videos
type Storage struct {
	rootFolder string
	debug      bool
}

// NewStorage creates a new storage instance
func NewStorage(rootFolder string, debug bool) *Storage {
	return &Storage{
		rootFolder: rootFolder,
		debug:      debug,
	}
}

// GenerateStorageID creates a unique storage ID
func (s *Storage) GenerateStorageID() string {
	// Generate a short unique ID (8 characters)
	fullUUID := uuid.New().String()
	return strings.ReplaceAll(fullUUID, "-", "")[:8]
}

// CreateStorageFolder creates a folder for storing video and metadata
func (s *Storage) CreateStorageFolder(storageID string) (string, error) {
	folderPath := filepath.Join(s.rootFolder, storageID)
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create storage folder: %w", err)
	}
	return folderPath, nil
}

// SaveVideoFromURL downloads and saves a video from URL
func (s *Storage) SaveVideoFromURL(url string, storageID string, filename string) (string, int64, error) {
	// Create storage folder
	folderPath, err := s.CreateStorageFolder(storageID)
	if err != nil {
		return "", 0, err
	}

	// Determine file extension from URL or default to mp4
	ext := ".mp4"
	if strings.Contains(url, ".webm") {
		ext = ".webm"
	} else if strings.Contains(url, ".gif") {
		ext = ".gif"
	}

	// Use provided filename or default
	if filename == "" {
		filename = "video"
	}
	if !strings.Contains(filename, ".") {
		filename = filename + ext
	}

	outputPath := filepath.Join(folderPath, filename)

	// Download the video
	// Note: Debug logging disabled in MCP mode to avoid stdout pollution

	resp, err := http.Get(url)
	if err != nil {
		return "", 0, fmt.Errorf("failed to download video: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("failed to download video: status %d", resp.StatusCode)
	}

	// Create the output file
	out, err := os.Create(outputPath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	// Copy the video data
	size, err := io.Copy(out, resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("failed to save video: %w", err)
	}

	// Note: Debug logging disabled in MCP mode to avoid stdout pollution

	return outputPath, size, nil
}

// SaveMetadata saves generation metadata to YAML file
func (s *Storage) SaveMetadata(storageID string, metadata map[string]interface{}) error {
	// Ensure storage folder exists
	folderPath, err := s.CreateStorageFolder(storageID)
	if err != nil {
		return err
	}
	metadataPath := filepath.Join(folderPath, "metadata.yaml")

	// Add timestamp
	metadata["generated_at"] = time.Now().Format(time.RFC3339)

	data, err := yaml.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	// Note: Debug logging disabled in MCP mode to avoid stdout pollution

	return nil
}

// SaveInputImage saves the input image for I2V generation
func (s *Storage) SaveInputImage(storageID string, imagePath string) (string, error) {
	folderPath := filepath.Join(s.rootFolder, storageID)
	
	// Read the input image
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read input image: %w", err)
	}

	// Determine extension
	ext := filepath.Ext(imagePath)
	if ext == "" {
		ext = ".jpg"
	}

	// Save to storage folder
	outputPath := filepath.Join(folderPath, "input"+ext)
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to save input image: %w", err)
	}

	// Note: Debug logging disabled in MCP mode to avoid stdout pollution

	return outputPath, nil
}

// ImageToDataURL converts an image file to a data URL
func (s *Storage) ImageToDataURL(imagePath string) (string, error) {
	// Read the image file
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image file: %w", err)
	}

	// Determine MIME type based on extension
	ext := strings.ToLower(filepath.Ext(imagePath))
	var mimeType string
	switch ext {
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".png":
		mimeType = "image/png"
	case ".webp":
		mimeType = "image/webp"
	case ".gif":
		mimeType = "image/gif"
	default:
		// Try to detect from content
		contentType := http.DetectContentType(data)
		if strings.HasPrefix(contentType, "image/") {
			mimeType = contentType
		} else {
			mimeType = "image/jpeg" // Default fallback
		}
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(data)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)

	// Note: Debug logging disabled in MCP mode to avoid stdout pollution

	return dataURL, nil
}

// GetStoragePath returns the full path for a storage ID
func (s *Storage) GetStoragePath(storageID string) string {
	return filepath.Join(s.rootFolder, storageID)
}