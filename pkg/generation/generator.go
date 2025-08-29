package generation

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gomcpgo/replicate_video_ai/pkg/client"
	"github.com/gomcpgo/replicate_video_ai/pkg/storage"
	"github.com/gomcpgo/replicate_video_ai/pkg/types"
)

// Generator handles video generation operations
type Generator struct {
	client  client.Client
	storage *storage.Storage
	debug   bool
}

// NewGenerator creates a new video generator
func NewGenerator(client client.Client, storage *storage.Storage, debug bool) *Generator {
	return &Generator{
		client:  client,
		storage: storage,
		debug:   debug,
	}
}

// GenerateTextToVideo generates a video from text prompt
func (g *Generator) GenerateTextToVideo(ctx context.Context, params VideoParams) (*VideoResult, error) {
	startTime := time.Now()

	// Get model configuration
	modelConfig, ok := GetModelConfig(params.Model)
	if !ok {
		return nil, fmt.Errorf("unknown model: %s", params.Model)
	}

	if !IsTextToVideoModel(params.Model) {
		return nil, fmt.Errorf("model %s does not support text-to-video", params.Model)
	}

	// Build input parameters based on model
	input := g.buildTextToVideoInput(params, modelConfig)

	// Create storage ID
	storageID := g.storage.GenerateStorageID()

	// Create prediction
	if g.debug {
		log.Printf("DEBUG: Creating T2V prediction with model %s", modelConfig.ID)
	}

	prediction, err := g.client.CreatePrediction(ctx, modelConfig.ID, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create prediction: %w", err)
	}

	// Save metadata immediately
	metadata := map[string]interface{}{
		"operation":     "text_to_video",
		"model":         params.Model,
		"model_id":      modelConfig.ID,
		"prompt":        params.Prompt,
		"parameters":    input,
		"prediction_id": prediction.ID,
		"status":        prediction.Status,
	}

	if err := g.storage.SaveMetadata(storageID, metadata); err != nil {
		log.Printf("WARNING: Failed to save metadata: %v", err)
	}

	// Return immediately with prediction ID (async by default)
	result := &VideoResult{
		ID:           storageID,
		Model:        params.Model,
		ModelName:    modelConfig.Name,
		PredictionID: prediction.ID,
		Parameters:   input,
		Status:       prediction.Status,
		Metrics: VideoMetrics{
			GenerationTime: time.Since(startTime).Seconds(),
		},
	}

	return result, nil
}

// GenerateImageToVideo generates a video from an image with motion prompt
func (g *Generator) GenerateImageToVideo(ctx context.Context, params VideoParams) (*VideoResult, error) {
	startTime := time.Now()

	// Get model configuration
	modelConfig, ok := GetModelConfig(params.Model)
	if !ok {
		return nil, fmt.Errorf("unknown model: %s", params.Model)
	}

	if !IsImageToVideoModel(params.Model) {
		return nil, fmt.Errorf("model %s does not support image-to-video", params.Model)
	}

	// Convert image to data URL
	dataURL, err := g.storage.ImageToDataURL(params.ImagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to convert image: %w", err)
	}

	// Build input parameters based on model
	input := g.buildImageToVideoInput(params, modelConfig, dataURL)

	// Create storage ID
	storageID := g.storage.GenerateStorageID()

	// Save input image
	if _, err := g.storage.SaveInputImage(storageID, params.ImagePath); err != nil {
		log.Printf("WARNING: Failed to save input image: %v", err)
	}

	// Create prediction
	if g.debug {
		log.Printf("DEBUG: Creating I2V prediction with model %s", modelConfig.ID)
	}

	prediction, err := g.client.CreatePrediction(ctx, modelConfig.ID, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create prediction: %w", err)
	}

	// Save metadata immediately
	metadata := map[string]interface{}{
		"operation":     "image_to_video",
		"model":         params.Model,
		"model_id":      modelConfig.ID,
		"prompt":        params.Prompt,
		"input_image":   params.ImagePath,
		"parameters":    input,
		"prediction_id": prediction.ID,
		"status":        prediction.Status,
	}

	if err := g.storage.SaveMetadata(storageID, metadata); err != nil {
		log.Printf("WARNING: Failed to save metadata: %v", err)
	}

	// Return immediately with prediction ID (async by default)
	result := &VideoResult{
		ID:           storageID,
		Model:        params.Model,
		ModelName:    modelConfig.Name,
		PredictionID: prediction.ID,
		Parameters:   input,
		Status:       prediction.Status,
		Metrics: VideoMetrics{
			GenerationTime: time.Since(startTime).Seconds(),
		},
	}

	return result, nil
}

// ContinueGeneration continues checking and downloading a video generation
func (g *Generator) ContinueGeneration(ctx context.Context, predictionID string, storageID string, waitTime time.Duration) (*VideoResult, error) {
	startTime := time.Now()

	// Wait for completion with timeout
	prediction, err := g.client.WaitForCompletion(ctx, predictionID, waitTime)
	if err != nil {
		// Check if we at least got a prediction back
		if prediction != nil {
			return &VideoResult{
				ID:           storageID,
				PredictionID: predictionID,
				Status:       prediction.Status,
				Metrics: VideoMetrics{
					GenerationTime: time.Since(startTime).Seconds(),
				},
			}, err
		}
		return nil, err
	}

	// Check if succeeded
	if prediction.Status != types.StatusSucceeded {
		return &VideoResult{
			ID:           storageID,
			PredictionID: predictionID,
			Status:       prediction.Status,
			Metrics: VideoMetrics{
				GenerationTime: time.Since(startTime).Seconds(),
			},
		}, fmt.Errorf("generation failed with status: %s", prediction.Status)
	}

	// Download video from output URL
	outputURL, ok := prediction.Output.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected output format: %T", prediction.Output)
	}

	// Save video
	videoPath, fileSize, err := g.storage.SaveVideoFromURL(outputURL, storageID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to save video: %w", err)
	}

	// Update metadata with completion info
	metadata := map[string]interface{}{
		"prediction_id": predictionID,
		"status":        "completed",
		"output_url":    outputURL,
		"output_path":   videoPath,
		"file_size":     fileSize,
		"completed_at":  time.Now().Format(time.RFC3339),
	}

	if err := g.storage.SaveMetadata(storageID, metadata); err != nil {
		log.Printf("WARNING: Failed to update metadata: %v", err)
	}

	result := &VideoResult{
		ID:           storageID,
		FilePath:     videoPath,
		PredictionID: predictionID,
		Status:       "completed",
		Metrics: VideoMetrics{
			GenerationTime: time.Since(startTime).Seconds(),
			FileSize:       fileSize,
		},
	}

	return result, nil
}

// buildTextToVideoInput builds input parameters for T2V generation
func (g *Generator) buildTextToVideoInput(params VideoParams, config ModelConfig) map[string]interface{} {
	input := make(map[string]interface{})
	input["prompt"] = params.Prompt

	// Handle resolution
	if params.Resolution != "" {
		input["resolution"] = params.Resolution
	} else {
		input["resolution"] = config.DefaultRes
	}

	// Handle aspect ratio
	if params.AspectRatio != "" {
		input["aspect_ratio"] = params.AspectRatio
	}

	// Model-specific parameters
	switch params.Model {
	case "wan-t2v-fast":
		input["go_fast"] = true
		input["num_frames"] = 81 // Default
		input["frames_per_second"] = 16
		input["sample_shift"] = 12
		input["optimize_prompt"] = false

	case "veo3":
		if params.NegativePrompt != "" {
			input["negative_prompt"] = params.NegativePrompt
		}

	case "kling-master":
		if params.Duration > 0 {
			input["duration"] = params.Duration
		} else {
			input["duration"] = 5 // Default
		}
		if params.NegativePrompt != "" {
			input["negative_prompt"] = params.NegativePrompt
		}
	}

	return input
}

// buildImageToVideoInput builds input parameters for I2V generation
func (g *Generator) buildImageToVideoInput(params VideoParams, config ModelConfig, dataURL string) map[string]interface{} {
	input := make(map[string]interface{})
	input["prompt"] = params.Prompt
	input["image"] = dataURL

	// Handle resolution
	if params.Resolution != "" {
		input["resolution"] = params.Resolution
	} else {
		input["resolution"] = config.DefaultRes
	}

	// Model-specific parameters
	switch params.Model {
	case "wan-i2v-fast":
		input["go_fast"] = true
		input["num_frames"] = 81 // Default
		input["frames_per_second"] = 16
		input["sample_shift"] = 12
		input["disable_safety_checker"] = false

	case "veo3":
		if params.NegativePrompt != "" {
			input["negative_prompt"] = params.NegativePrompt
		}

	case "kling-master":
		// For kling-master in I2V mode, it requires start_image
		delete(input, "image")
		input["start_image"] = dataURL
		if params.Duration > 0 {
			input["duration"] = params.Duration
		} else {
			input["duration"] = 5 // Default
		}
		if params.NegativePrompt != "" {
			input["negative_prompt"] = params.NegativePrompt
		}
	}

	return input
}