package handler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/gomcpgo/replicate_video_ai/pkg/responses"
)

// handleContinueOperation handles the continue_operation tool
func (h *ReplicateVideoHandler) handleContinueOperation(ctx context.Context, args map[string]interface{}) (*protocol.CallToolResponse, error) {
	// Note: Debug logging disabled in MCP mode
	
	// Extract parameters - support both prediction_id (for backward compatibility) and operation_id
	var operationID string
	
	// First try prediction_id (for backward compatibility)
	if predID, ok := args["prediction_id"].(string); ok && predID != "" {
		operationID = predID
	} else if opID, ok := args["operation_id"].(string); ok && opID != "" {
		// Fall back to operation_id
		operationID = opID
	} else {
		return h.errorResponse("continue_operation", "invalid_parameters", "prediction_id or operation_id is required", nil)
	}
	
	waitTime := 30 * time.Second
	if wt, ok := args["wait_time"].(float64); ok {
		waitTime = time.Duration(wt) * time.Second
		if waitTime < 5*time.Second {
			waitTime = 5 * time.Second
		}
		if waitTime > 60*time.Second {
			waitTime = 60 * time.Second
		}
	}
	
	// Since we don't have a built-in async executor yet, let's handle this directly
	// by calling the generator's ContinueGeneration method
	
	// Find existing storage ID for this prediction ID
	storageID, err := h.findStorageIDForPrediction(operationID)
	if err != nil || storageID == "" {
		// If we can't find existing storage ID, generate a new one
		storageID = h.generateStorageID()
	}
	
	result, err := h.generator.ContinueGeneration(ctx, operationID, storageID, waitTime)
	if err != nil {
		// Check if it's still processing
		if result != nil && result.Status == "processing" {
			// Return processing response
			response := responses.BuildProcessingResponse(
				"continue_operation",
				operationID,
				result.ID,
				int(waitTime.Seconds()),
			)
			return &protocol.CallToolResponse{
				Content: []protocol.ToolContent{
					{Type: "text", Text: response},
				},
			}, nil
		}
		
		return h.errorResponse("continue_operation", "operation_failed", err.Error(), map[string]interface{}{
			"prediction_id": operationID,
		})
	}
	
	// Handle the result based on status
	switch result.Status {
	case "processing":
		// Still processing - return processing response
		response := responses.BuildProcessingResponse(
			"continue_operation",
			operationID,
			result.ID,
			int(waitTime.Seconds()),
		)
		
		return &protocol.CallToolResponse{
			Content: []protocol.ToolContent{
				{Type: "text", Text: response},
			},
		}, nil
		
	case "completed":
		// Load full metadata for the completed video
		metadata, err := h.storage.LoadMetadata(storageID)
		if err != nil {
			// Log but don't fail - use what we have
			metadata = make(map[string]interface{})
		}
		
		// Build paths with absolute paths from relative paths in metadata
		paths := make(map[string]string)
		basePath := h.storage.GetStoragePath(storageID)
		
		// Convert relative paths to absolute
		if metadataPaths, ok := metadata["paths"].(map[string]interface{}); ok {
			if output, ok := metadataPaths["output"].(string); ok {
				paths["output"] = filepath.Join(basePath, output)
			}
			if thumbnail, ok := metadataPaths["thumbnail"].(string); ok {
				paths["thumbnail"] = filepath.Join(basePath, thumbnail)
			}
		} else {
			// Fallback for old format
			paths["output"] = result.FilePath
		}
		
		// Extract parameters from metadata (includes prompt)
		parameters := make(map[string]interface{})
		if params, ok := metadata["parameters"].(map[string]interface{}); ok {
			parameters = params
		}
		// Ensure prompt is in parameters
		if prompt, ok := metadata["prompt"].(string); ok && prompt != "" {
			parameters["prompt"] = prompt
		}
		// Add other parameter fields
		if resolution, ok := metadata["resolution"].(string); ok {
			parameters["resolution"] = resolution
		}
		if aspectRatio, ok := metadata["aspect_ratio"].(string); ok {
			parameters["aspect_ratio"] = aspectRatio
		}
		if duration, ok := metadata["duration"].(int); ok {
			parameters["duration"] = duration
		}
		if negativePrompt, ok := metadata["negative_prompt"].(string); ok {
			parameters["negative_prompt"] = negativePrompt
		}
		
		// Build model info
		modelInfo := make(map[string]string)
		if modelID, ok := metadata["model"].(string); ok {
			modelInfo["id"] = modelID
		}
		if modelName, ok := metadata["model_name"].(string); ok {
			modelInfo["name"] = modelName
		} else if result.ModelName != "" {
			modelInfo["name"] = result.ModelName
		}
		
		// Build metrics (video metadata only, no prompt/params)
		metrics := map[string]interface{}{
			"generation_time": result.Metrics.GenerationTime,
			"file_size":       result.Metrics.FileSize,
		}
		
		// Add actual video metadata to metrics
		if actualRes, ok := metadata["actual_resolution"].(string); ok && actualRes != "" {
			metrics["actual_resolution"] = actualRes
		}
		if actualDur, ok := metadata["actual_duration"].(float64); ok && actualDur > 0 {
			metrics["actual_duration"] = actualDur
		}
		if genType, ok := metadata["generation_type"].(string); ok {
			metrics["generation_type"] = genType
		}
		if format, ok := metadata["format"].(string); ok {
			metrics["format"] = format
		}
		
		// Operation completed - build success response
		response := responses.BuildSuccessResponse(
			"continue_operation",
			result.ID,
			paths,
			modelInfo,
			parameters,
			metrics,
			result.PredictionID,
		)
		
		return &protocol.CallToolResponse{
			Content: []protocol.ToolContent{
				{Type: "text", Text: response},
			},
		}, nil
		
	default:
		return h.errorResponse("continue_operation", "unexpected_status", 
			fmt.Sprintf("unexpected operation status: %s", result.Status), 
			map[string]interface{}{
				"prediction_id": operationID,
				"status": result.Status,
			})
	}
}

// generateStorageID creates a unique storage ID for continue operations
func (h *ReplicateVideoHandler) generateStorageID() string {
	return h.storage.GenerateStorageID()
}

// findStorageIDForPrediction searches for existing storage ID with given prediction ID
func (h *ReplicateVideoHandler) findStorageIDForPrediction(predictionID string) (string, error) {
	// Get the root videos folder
	videosDir := h.storage.GetStoragePath("")
	
	// Read all subdirectories (storage IDs)
	entries, err := os.ReadDir(videosDir)
	if err != nil {
		return "", fmt.Errorf("failed to read videos directory: %w", err)
	}
	
	// Search through each storage directory
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		storageID := entry.Name()
		metadata, err := h.storage.LoadMetadata(storageID)
		if err != nil {
			continue // Skip if can't load metadata
		}
		
		// Check if this metadata matches the prediction ID
		if metaPredID, ok := metadata["prediction_id"].(string); ok && metaPredID == predictionID {
			return storageID, nil
		}
	}
	
	return "", fmt.Errorf("storage ID not found for prediction %s", predictionID)
}

// Helper functions to extract values from result map (for future async executor integration)
func getStringValue(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getFloatValue(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}

func getIntValue(m map[string]interface{}, key string) int64 {
	if v, ok := m[key].(int64); ok {
		return v
	}
	if v, ok := m[key].(float64); ok {
		return int64(v)
	}
	return 0
}

func getMapValue(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key].(map[string]interface{}); ok {
		return v
	}
	return make(map[string]interface{})
}