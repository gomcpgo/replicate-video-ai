package client

import (
	"context"
	"time"

	"github.com/gomcpgo/replicate_video_ai/pkg/types"
)

// Client defines the interface for Replicate API client
type Client interface {
	CreatePrediction(ctx context.Context, modelVersion string, input map[string]interface{}) (*types.ReplicatePredictionResponse, error)
	GetPrediction(ctx context.Context, predictionID string) (*types.ReplicatePredictionResponse, error)
	WaitForCompletion(ctx context.Context, predictionID string, timeout time.Duration) (*types.ReplicatePredictionResponse, error)
	CancelPrediction(ctx context.Context, predictionID string) error
}