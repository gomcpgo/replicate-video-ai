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
		// Operation completed - build success response
		response := responses.BuildSuccessResponse(
			"continue_operation",
			result.ID,
			map[string]string{
				"output": result.FilePath,
			},
			map[string]string{
				"name": result.ModelName,
			},
			map[string]interface{}{}, // Parameters not available in this context
			map[string]interface{}{
				"generation_time": result.Metrics.GenerationTime,
				"file_size":       result.Metrics.FileSize,
			},
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