package responses

import (
	"encoding/json"
	"log"

	"github.com/gomcpgo/replicate_video_ai/pkg/types"
)

// BuildSuccessResponse creates a success response
func BuildSuccessResponse(operation, storageID string, paths map[string]string, model map[string]string, parameters map[string]interface{}, metrics map[string]interface{}, predictionID string) string {
	response := types.SuccessResponse{
		Success:      true,
		Operation:    operation,
		StorageID:    storageID,
		PredictionID: predictionID,
		Status:       "completed",
		Paths:        paths,
		Model:        model,
		Parameters:   parameters,
		Metrics:      metrics,
	}

	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Printf("ERROR: Failed to marshal success response: %v", err)
		return `{"success": false, "error": {"message": "Failed to format response"}}`
	}

	return string(data)
}

// BuildProcessingResponse creates a processing/async response
func BuildProcessingResponse(operation, predictionID, storageID string, waitTime int) string {
	response := types.ProcessingResponse{
		Success:      true,
		Status:       "processing",
		Operation:    operation,
		PredictionID: predictionID,
		StorageID:    storageID,
		Message:      "Video generation in progress. Use continue_operation to check status.",
		WaitTime:     waitTime,
	}

	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Printf("ERROR: Failed to marshal processing response: %v", err)
		return `{"success": false, "error": {"message": "Failed to format response"}}`
	}

	return string(data)
}

// BuildErrorResponse creates an error response
func BuildErrorResponse(operation, errorType, message string, details map[string]interface{}) string {
	response := types.ErrorResponse{
		Success:   false,
		Operation: operation,
		Error: types.ErrorDetails{
			Type:    errorType,
			Message: message,
			Details: details,
		},
	}

	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Printf("ERROR: Failed to marshal error response: %v", err)
		return `{"success": false, "error": {"message": "Failed to format error"}}`
	}

	return string(data)
}