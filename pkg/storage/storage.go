package storage

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
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

// LoadMetadata loads metadata from a YAML file
func (s *Storage) LoadMetadata(storageID string) (map[string]interface{}, error) {
	folderPath := filepath.Join(s.rootFolder, storageID)
	metadataPath := filepath.Join(folderPath, "metadata.yaml")
	
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty map if metadata doesn't exist yet
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}
	
	var metadata map[string]interface{}
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	
	return metadata, nil
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

// GenerateThumbnail attempts to generate a thumbnail from video using ffmpeg
// Returns the thumbnail path if successful, empty string if ffmpeg is not available
func (s *Storage) GenerateThumbnail(storageID string, videoPath string) (string, error) {
	// Check if ffmpeg is available
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		log.Printf("WARNING: ffmpeg not found, skipping thumbnail generation: %v", err)
		return "", nil // Not an error, just degraded functionality
	}
	
	// Create thumbnail path
	folderPath := filepath.Join(s.rootFolder, storageID)
	thumbnailPath := filepath.Join(folderPath, "thumbnail.jpg")
	
	// Build ffmpeg command to extract frame at 2 seconds (or middle if shorter)
	// -ss 2: seek to 2 seconds
	// -i: input file
	// -vframes 1: extract 1 frame
	// -vf scale=320:-1: scale to 320px width, maintain aspect ratio
	// -q:v 2: JPEG quality (2 is good quality)
	cmd := exec.Command(ffmpegPath,
		"-ss", "2",
		"-i", videoPath,
		"-vframes", "1",
		"-vf", "scale=320:-1",
		"-q:v", "2",
		"-y", // Overwrite output file
		thumbnailPath,
	)
	
	// Run the command with timeout
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try extracting first frame instead if seeking to 2 seconds failed
		cmd = exec.Command(ffmpegPath,
			"-i", videoPath,
			"-vframes", "1",
			"-vf", "scale=320:-1",
			"-q:v", "2",
			"-y",
			thumbnailPath,
		)
		output, err = cmd.CombinedOutput()
		if err != nil {
			log.Printf("WARNING: Failed to generate thumbnail: %v, output: %s", err, string(output))
			return "", nil // Not a critical error
		}
	}
	
	// Verify thumbnail was created
	if _, err := os.Stat(thumbnailPath); os.IsNotExist(err) {
		log.Printf("WARNING: Thumbnail file was not created")
		return "", nil
	}
	
	log.Printf("Successfully generated thumbnail: %s", thumbnailPath)
	return thumbnailPath, nil
}

// ExtractVideoMetadata attempts to extract video metadata using ffmpeg
// Returns duration and resolution if successful
func (s *Storage) ExtractVideoMetadata(videoPath string) (duration float64, resolution string, err error) {
	// Check if ffprobe is available (comes with ffmpeg)
	ffprobePath, err := exec.LookPath("ffprobe")
	if err != nil {
		log.Printf("WARNING: ffprobe not found, skipping metadata extraction: %v", err)
		return 0, "", nil
	}
	
	// Get duration
	durationCmd := exec.Command(ffprobePath,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	)
	
	durationOutput, err := durationCmd.Output()
	if err != nil {
		log.Printf("WARNING: Failed to extract duration: %v", err)
	} else {
		// Parse duration string
		var d float64
		fmt.Sscanf(string(durationOutput), "%f", &d)
		duration = d
	}
	
	// Get resolution
	resCmd := exec.Command(ffprobePath,
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=s=x:p=0",
		videoPath,
	)
	
	resOutput, err := resCmd.Output()
	if err != nil {
		log.Printf("WARNING: Failed to extract resolution: %v", err)
	} else {
		resolution = strings.TrimSpace(string(resOutput))
	}
	
	return duration, resolution, nil
}