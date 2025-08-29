package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gomcpgo/replicate_video_ai/pkg/types"
)

const (
	replicateAPIURL = "https://api.replicate.com/v1"
)

// ReplicateClient handles communication with the Replicate API
type ReplicateClient struct {
	apiToken   string
	httpClient *http.Client
	debug      bool
}

// NewReplicateClient creates a new Replicate API client
func NewReplicateClient(apiToken string, debug bool) *ReplicateClient {
	return &ReplicateClient{
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		debug: debug,
	}
}

// CreatePrediction creates a new prediction on Replicate
func (c *ReplicateClient) CreatePrediction(ctx context.Context, modelVersion string, input map[string]interface{}) (*types.ReplicatePredictionResponse, error) {
	var url string
	var body []byte
	var err error

	// Check if modelVersion contains a version hash (has colon)
	if strings.Contains(modelVersion, ":") {
		// Use version endpoint for specific versions
		req := types.ReplicatePredictionRequest{
			Version: modelVersion,
			Input:   input,
		}
		body, err = json.Marshal(req)
		url = fmt.Sprintf("%s/predictions", replicateAPIURL)
	} else {
		// Use deployment endpoint for latest version
		reqBody := map[string]interface{}{
			"input": input,
		}
		body, err = json.Marshal(reqBody)
		url = fmt.Sprintf("%s/models/%s/predictions", replicateAPIURL, modelVersion)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Note: Debug logging disabled in MCP mode to avoid stdout pollution

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Note: Debug logging disabled in MCP mode to avoid stdout pollution

	// Handle specific error codes
	if resp.StatusCode == http.StatusPaymentRequired {
		var errorResp map[string]interface{}
		if err := json.Unmarshal(respBody, &errorResp); err == nil {
			if detail, ok := errorResp["detail"].(string); ok {
				return nil, fmt.Errorf("billing issue: %s", detail)
			}
		}
		return nil, fmt.Errorf("billing issue (status 402): %s", string(respBody))
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var prediction types.ReplicatePredictionResponse
	if err := json.Unmarshal(respBody, &prediction); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &prediction, nil
}

// GetPrediction gets the status of a prediction
func (c *ReplicateClient) GetPrediction(ctx context.Context, predictionID string) (*types.ReplicatePredictionResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/predictions/%s", replicateAPIURL, predictionID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var prediction types.ReplicatePredictionResponse
	if err := json.Unmarshal(respBody, &prediction); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &prediction, nil
}

// WaitForCompletion waits for a prediction to complete or timeout
func (c *ReplicateClient) WaitForCompletion(ctx context.Context, predictionID string, timeout time.Duration) (*types.ReplicatePredictionResponse, error) {
	// Note: Debug logging disabled in MCP mode to avoid stdout pollution

	// If timeout is 0, use context deadline or a very long timeout
	var deadline time.Time
	if timeout == 0 {
		if d, ok := ctx.Deadline(); ok {
			deadline = d
		} else {
			deadline = time.Now().Add(10 * time.Minute)
		}
	} else {
		deadline = time.Now().Add(timeout)
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	pollCount := 0

	for {
		select {
		case <-ctx.Done():
			// Note: Debug logging disabled in MCP mode to avoid stdout pollution
			return nil, ctx.Err()
		case <-ticker.C:
			pollCount++
			if time.Now().After(deadline) {
				// Note: Debug logging disabled in MCP mode to avoid stdout pollution
				prediction, _ := c.GetPrediction(ctx, predictionID)
				return prediction, fmt.Errorf("operation timed out after %v", timeout)
			}

			// Note: Debug logging disabled in MCP mode to avoid stdout pollution
			prediction, err := c.GetPrediction(ctx, predictionID)
			if err != nil {
				// Note: Debug logging disabled in MCP mode to avoid stdout pollution
				return nil, err
			}

			// Note: Debug logging disabled in MCP mode to avoid stdout pollution
			switch prediction.Status {
			case types.StatusSucceeded:
				// Note: Debug logging disabled in MCP mode to avoid stdout pollution
				return prediction, nil
			case types.StatusFailed:
				errMsg := "prediction failed"
				if prediction.Error != nil {
					if errStr, ok := prediction.Error.(string); ok {
						errMsg = errStr
					} else if errMap, ok := prediction.Error.(map[string]interface{}); ok {
						if msg, exists := errMap["message"]; exists {
							errMsg = fmt.Sprintf("%v", msg)
						}
					}
				}
				return prediction, fmt.Errorf(errMsg)
			case types.StatusCanceled:
				return prediction, fmt.Errorf("prediction was canceled")
			}
			// Continue polling for "starting" or "processing" status
		}
	}
}

// CancelPrediction cancels a running prediction
func (c *ReplicateClient) CancelPrediction(ctx context.Context, predictionID string) error {
	httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/predictions/%s/cancel", replicateAPIURL, predictionID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to cancel prediction (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}