package types

// SuccessResponse represents a successful operation response
type SuccessResponse struct {
	Success      bool                   `json:"success"`
	Operation    string                 `json:"operation"`
	StorageID    string                 `json:"storage_id"`
	PredictionID string                 `json:"prediction_id,omitempty"`
	Status       string                 `json:"status"`
	Paths        map[string]string      `json:"paths"`
	Model        map[string]string      `json:"model"`
	Parameters   map[string]interface{} `json:"parameters"`
	Metrics      map[string]interface{} `json:"metrics,omitempty"`
	Message      string                 `json:"message,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success   bool                   `json:"success"`
	Operation string                 `json:"operation"`
	Error     ErrorDetails           `json:"error"`
}

// ErrorDetails contains error information
type ErrorDetails struct {
	Type    string                 `json:"type"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ProcessingResponse represents an async operation in progress
type ProcessingResponse struct {
	Success      bool   `json:"success"`
	Status       string `json:"status"`
	Operation    string `json:"operation"`
	PredictionID string `json:"prediction_id"`
	StorageID    string `json:"storage_id,omitempty"`
	Message      string `json:"message"`
	WaitTime     int    `json:"wait_time,omitempty"`
}