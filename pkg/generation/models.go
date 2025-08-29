package generation

// ModelConfig holds configuration for a video model
type ModelConfig struct {
	ID          string
	Name        string
	Type        string // "t2v", "i2v", or "both"
	DefaultRes  string
	MaxDuration int
	Features    []string
}

// ModelAliases maps short aliases to full model names
var ModelAliases = map[string]string{
	"wan-t2v-fast": "wan-video/wan-2.2-t2v-fast",
	"wan-i2v-fast": "wan-video/wan-2.2-i2v-fast",
	"veo3":         "google/veo-3",
	"kling-master": "kwaivgi/kling-v2.1-master",
	"wan-i2v-full": "wan-video/wan-2.2-i2v-a14b",
	"kling":        "kwaivgi/kling-v2.1",
}

// ModelConfigs holds configuration for each model
var ModelConfigs = map[string]ModelConfig{
	"wan-t2v-fast": {
		ID:          "wan-video/wan-2.2-t2v-fast",
		Name:        "Wan 2.2 Fast Text-to-Video",
		Type:        "t2v",
		DefaultRes:  "480p",
		MaxDuration: 0, // Uses frames instead
		Features:    []string{"fast", "affordable", "go_fast"},
	},
	"wan-i2v-fast": {
		ID:          "wan-video/wan-2.2-i2v-fast",
		Name:        "Wan 2.2 Fast Image-to-Video",
		Type:        "i2v",
		DefaultRes:  "480p",
		MaxDuration: 0, // Uses frames instead
		Features:    []string{"fast", "affordable", "go_fast"},
	},
	"veo3": {
		ID:          "google/veo-3",
		Name:        "Google Veo 3",
		Type:        "both",
		DefaultRes:  "720p",
		MaxDuration: 0,
		Features:    []string{"premium", "audio", "style_preservation", "negative_prompt"},
	},
	"kling-master": {
		ID:          "kwaivgi/kling-v2.1-master",
		Name:        "Kling 2.1 Master",
		Type:        "both",
		DefaultRes:  "1080p",
		MaxDuration: 10,
		Features:    []string{"high_quality", "duration_control", "negative_prompt"},
	},
}

// GetModelID returns the full model ID from an alias
func GetModelID(alias string) string {
	if id, ok := ModelAliases[alias]; ok {
		return id
	}
	// If not an alias, assume it's already a full model ID
	return alias
}

// GetModelConfig returns the configuration for a model
func GetModelConfig(alias string) (ModelConfig, bool) {
	config, ok := ModelConfigs[alias]
	return config, ok
}

// IsTextToVideoModel checks if a model supports text-to-video
func IsTextToVideoModel(alias string) bool {
	if config, ok := ModelConfigs[alias]; ok {
		return config.Type == "t2v" || config.Type == "both"
	}
	return false
}

// IsImageToVideoModel checks if a model supports image-to-video
func IsImageToVideoModel(alias string) bool {
	if config, ok := ModelConfigs[alias]; ok {
		return config.Type == "i2v" || config.Type == "both"
	}
	return false
}