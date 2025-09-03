package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config holds the configuration for the Replicate Video AI server
type Config struct {
	ReplicateAPIToken   string
	VideosRootFolder    string
	DebugMode          bool
	DefaultTimeout     time.Duration
	PollInterval       time.Duration
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{
		DefaultTimeout: 5 * time.Minute,
		PollInterval:  2 * time.Second,
	}

	// Optional: API token (MCP server can start without it)
	cfg.ReplicateAPIToken = os.Getenv("REPLICATE_API_TOKEN")

	// Optional: Videos root folder
	cfg.VideosRootFolder = os.Getenv("REPLICATE_VIDEOS_ROOT_FOLDER")
	if cfg.VideosRootFolder == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		cfg.VideosRootFolder = filepath.Join(homeDir, "Library", "Application Support", "Savant", "replicate_video_ai")
	}

	// Create videos directory if it doesn't exist
	if err := os.MkdirAll(cfg.VideosRootFolder, 0755); err != nil {
		return nil, fmt.Errorf("failed to create videos directory: %w", err)
	}

	// Optional: Debug mode
	cfg.DebugMode = os.Getenv("REPLICATE_VIDEO_DEBUG") == "true"

	// Optional: Timeout
	if timeout := os.Getenv("REPLICATE_VIDEO_DEFAULT_TIMEOUT"); timeout != "" {
		duration, err := time.ParseDuration(timeout + "s")
		if err != nil {
			return nil, fmt.Errorf("invalid REPLICATE_VIDEO_DEFAULT_TIMEOUT: %w", err)
		}
		cfg.DefaultTimeout = duration
	}

	// Optional: Poll interval
	if interval := os.Getenv("REPLICATE_VIDEO_POLL_INTERVAL"); interval != "" {
		duration, err := time.ParseDuration(interval + "s")
		if err != nil {
			return nil, fmt.Errorf("invalid REPLICATE_VIDEO_POLL_INTERVAL: %w", err)
		}
		cfg.PollInterval = duration
	}

	return cfg, nil
}