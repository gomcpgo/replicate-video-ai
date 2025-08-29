package config

import "time"

// TimeoutConfig holds timeout configuration for video operations
type TimeoutConfig struct {
	InitialWait  time.Duration
	MaxWait      time.Duration
	PollInterval time.Duration
	TotalTimeout time.Duration
}

// LoadTimeouts returns default timeout configuration
func LoadTimeouts() TimeoutConfig {
	return TimeoutConfig{
		InitialWait:  30 * time.Second,
		MaxWait:      5 * time.Minute,
		PollInterval: 2 * time.Second,
		TotalTimeout: 10 * time.Minute,
	}
}