package handler

import (
	"context"
	"fmt"
	"os"

	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/gomcpgo/replicate_video_ai/pkg/generation"
)

// handleGenerateVideoFromText handles text-to-video generation
func (h *ReplicateVideoHandler) handleGenerateVideoFromText(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	// Note: Debug logging disabled in MCP mode
	
	// Extract and validate parameters
	params, err := h.extractTextToVideoParams(args)
	if err != nil {
		return h.errorResponse("generate_video_from_text", "invalid_parameters", err.Error(), nil)
	}
	
	// Generate video (async by default)
	result, err := h.generator.GenerateTextToVideo(ctx, params)
	if err != nil {
		return h.errorResponse("generate_video_from_text", "generation_failed", err.Error(), nil)
	}
	
	// Return processing response (async)
	return h.processingResponse(
		"generate_video_from_text",
		result.PredictionID,
		result.ID,
		30,
	)
}

// handleGenerateVideoFromImage handles image-to-video generation
func (h *ReplicateVideoHandler) handleGenerateVideoFromImage(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	// Note: Debug logging disabled in MCP mode
	
	// Extract and validate parameters
	params, err := h.extractImageToVideoParams(args)
	if err != nil {
		return h.errorResponse("generate_video_from_image", "invalid_parameters", err.Error(), nil)
	}
	
	// Validate image file exists
	if _, err := os.Stat(params.ImagePath); os.IsNotExist(err) {
		return h.errorResponse("generate_video_from_image", "file_not_found", 
			fmt.Sprintf("Image file not found: %s", params.ImagePath), nil)
	}
	
	// Generate video (async by default)
	result, err := h.generator.GenerateImageToVideo(ctx, params)
	if err != nil {
		return h.errorResponse("generate_video_from_image", "generation_failed", err.Error(), nil)
	}
	
	// Return processing response (async)
	return h.processingResponse(
		"generate_video_from_image",
		result.PredictionID,
		result.ID,
		30,
	)
}

// extractTextToVideoParams extracts and validates T2V parameters
func (h *ReplicateVideoHandler) extractTextToVideoParams(args map[string]interface{}) (generation.VideoParams, error) {
	var params generation.VideoParams
	
	// Required: prompt
	prompt, ok := args["prompt"].(string)
	if !ok || prompt == "" {
		return params, fmt.Errorf("prompt parameter is required and must be a non-empty string")
	}
	params.Prompt = prompt
	
	// Optional: model (default: wan-t2v-fast)
	if model, ok := args["model"].(string); ok && model != "" {
		params.Model = model
	} else {
		params.Model = "wan-t2v-fast"
	}
	
	// Validate model supports T2V
	if !generation.IsTextToVideoModel(params.Model) {
		return params, fmt.Errorf("model %s does not support text-to-video generation", params.Model)
	}
	
	// Optional: resolution
	if resolution, ok := args["resolution"].(string); ok && resolution != "" {
		params.Resolution = resolution
	}
	
	// Optional: aspect_ratio
	if aspectRatio, ok := args["aspect_ratio"].(string); ok && aspectRatio != "" {
		params.AspectRatio = aspectRatio
	}
	
	// Optional: duration (for Kling)
	if durationFloat, ok := args["duration"].(float64); ok {
		duration := int(durationFloat)
		if duration < 5 || duration > 10 {
			return params, fmt.Errorf("duration must be between 5 and 10 seconds")
		}
		params.Duration = duration
	}
	
	// Optional: negative_prompt (for Veo3, Kling)
	if negativePrompt, ok := args["negative_prompt"].(string); ok {
		params.NegativePrompt = negativePrompt
	}
	
	// Optional: filename
	if filename, ok := args["filename"].(string); ok {
		params.Filename = filename
	}
	
	return params, nil
}

// extractImageToVideoParams extracts and validates I2V parameters
func (h *ReplicateVideoHandler) extractImageToVideoParams(args map[string]interface{}) (generation.VideoParams, error) {
	var params generation.VideoParams
	
	// Required: image_path
	imagePath, ok := args["image_path"].(string)
	if !ok || imagePath == "" {
		return params, fmt.Errorf("image_path parameter is required and must be a non-empty string")
	}
	params.ImagePath = imagePath
	
	// Required: prompt
	prompt, ok := args["prompt"].(string)
	if !ok || prompt == "" {
		return params, fmt.Errorf("prompt parameter is required and must be a non-empty string")
	}
	params.Prompt = prompt
	
	// Optional: model (default: wan-i2v-fast)
	if model, ok := args["model"].(string); ok && model != "" {
		params.Model = model
	} else {
		params.Model = "wan-i2v-fast"
	}
	
	// Validate model supports I2V
	if !generation.IsImageToVideoModel(params.Model) {
		return params, fmt.Errorf("model %s does not support image-to-video generation", params.Model)
	}
	
	// Optional: resolution
	if resolution, ok := args["resolution"].(string); ok && resolution != "" {
		params.Resolution = resolution
	}
	
	// Optional: duration (for Kling)
	if durationFloat, ok := args["duration"].(float64); ok {
		duration := int(durationFloat)
		if duration < 5 || duration > 10 {
			return params, fmt.Errorf("duration must be between 5 and 10 seconds")
		}
		params.Duration = duration
	}
	
	// Optional: negative_prompt (for Veo3, Kling)
	if negativePrompt, ok := args["negative_prompt"].(string); ok {
		params.NegativePrompt = negativePrompt
	}
	
	// Optional: filename
	if filename, ok := args["filename"].(string); ok {
		params.Filename = filename
	}
	
	return params, nil
}