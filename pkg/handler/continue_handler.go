package handler

import (
	"context"
	"fmt"
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
	
	// Use a placeholder storage ID since we might not have it
	storageID := h.generateStorageID()
	
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
		
		// Extract parameters from metadata if available
		parameters := make(map[string]interface{})
		if params, ok := metadata["parameters"].(map[string]interface{}); ok {
			parameters = params
		}
		
		// Extract model info
		modelInfo := map[string]string{
			"name": result.ModelName,
		}
		if modelName, ok := metadata["model_name"].(string); ok {
			modelInfo["name"] = modelName
		}
		if model, ok := metadata["model"].(string); ok {
			modelInfo["id"] = model
		}
		
		// Build comprehensive metrics
		metrics := map[string]interface{}{
			"generation_time": result.Metrics.GenerationTime,
			"file_size":       result.Metrics.FileSize,
		}
		
		// Add metadata fields to metrics for frontend
		if prompt, ok := metadata["prompt"].(string); ok {
			metrics["prompt"] = prompt
		}
		// Prefer actual extracted resolution over requested
		if actualRes, ok := metadata["actual_resolution"].(string); ok && actualRes != "" {
			metrics["resolution"] = actualRes
		} else if resolution, ok := metadata["resolution"].(string); ok {
			metrics["resolution"] = resolution
		}
		// Prefer actual extracted duration over requested
		if actualDur, ok := metadata["actual_duration"].(float64); ok && actualDur > 0 {
			metrics["duration"] = actualDur
		} else if duration, ok := metadata["duration"].(float64); ok {
			metrics["duration"] = duration
		}
		if genType, ok := metadata["generation_type"].(string); ok {
			metrics["generation_type"] = genType
		}
		if format, ok := metadata["format"].(string); ok {
			metrics["format"] = format
		}
		// Add thumbnail path if available
		if thumbnailPath, ok := metadata["thumbnail_path"].(string); ok && thumbnailPath != "" {
			metrics["thumbnail_path"] = thumbnailPath
			metrics["thumbnail_available"] = true
		} else {
			metrics["thumbnail_available"] = false
		}
		
		// Operation completed - build success response
		response := responses.BuildSuccessResponse(
			"continue_operation",
			result.ID,
			map[string]string{
				"output": result.FilePath,
			},
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