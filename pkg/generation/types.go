package generation

// VideoParams holds parameters for video generation
type VideoParams struct {
	// Common parameters
	Prompt      string
	Model       string
	Resolution  string
	AspectRatio string
	Filename    string

	// Text-to-video specific
	NegativePrompt string
	Duration       int // For Kling

	// Image-to-video specific
	ImagePath       string
	NumFrames       int // For Wan
	FramesPerSecond int

	// Model-specific optimizations
	GoFast      bool    // For Wan fast models
	SampleShift float64 // For Wan tuning
}

// VideoResult holds the result of video generation
type VideoResult struct {
	ID           string
	FilePath     string
	Model        string
	ModelName    string
	PredictionID string
	Parameters   map[string]interface{}
	Metrics      VideoMetrics
	Status       string
}

// VideoMetrics holds metrics about the generated video
type VideoMetrics struct {
	GenerationTime float64
	FileSize       int64
	Duration       float64
	Resolution     string
	FrameCount     int
}