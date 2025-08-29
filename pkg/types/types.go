package types

// ReplicatePredictionRequest represents the request to create a prediction
type ReplicatePredictionRequest struct {
	Version string                 `json:"version,omitempty"`
	Input   map[string]interface{} `json:"input"`
}

// ReplicatePredictionResponse represents the response from Replicate API
type ReplicatePredictionResponse struct {
	ID          string                 `json:"id"`
	Version     string                 `json:"version"`
	Status      string                 `json:"status"`
	Input       map[string]interface{} `json:"input"`
	Output      interface{}            `json:"output"`
	Error       interface{}            `json:"error"`
	Logs        string                 `json:"logs"`
	MetricsEnd  interface{}            `json:"metrics_end"`
	CreatedAt   string                 `json:"created_at"`
	StartedAt   string                 `json:"started_at"`
	CompletedAt string                 `json:"completed_at"`
	URLs        map[string]string      `json:"urls"`
}

// Prediction status constants
const (
	StatusStarting   = "starting"
	StatusProcessing = "processing"
	StatusSucceeded  = "succeeded"
	StatusFailed     = "failed"
	StatusCanceled   = "canceled"
)