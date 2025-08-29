package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/gomcpgo/mcp/pkg/async"
	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/gomcpgo/replicate_video_ai/pkg/client"
	"github.com/gomcpgo/replicate_video_ai/pkg/config"
	"github.com/gomcpgo/replicate_video_ai/pkg/generation"
	"github.com/gomcpgo/replicate_video_ai/pkg/responses"
	"github.com/gomcpgo/replicate_video_ai/pkg/storage"
)

// ReplicateVideoHandler handles MCP requests for video operations
type ReplicateVideoHandler struct {
	generator *generation.Generator
	storage   *storage.Storage
	client    client.Client
	executor  *async.OperationExecutor
	timeouts  config.TimeoutConfig
	debug     bool
}

// NewReplicateVideoHandler creates a new handler instance
func NewReplicateVideoHandler(apiKey string, rootFolder string, debug bool) (*ReplicateVideoHandler, error) {
	// Initialize storage
	store := storage.NewStorage(rootFolder, debug)
	
	// Initialize Replicate client
	replicateClient := client.NewReplicateClient(apiKey, debug)
	
	// Initialize generator
	gen := generation.NewGenerator(replicateClient, store, debug)
	
	// Load timeout configuration
	timeouts := config.LoadTimeouts()
	
	// Initialize async executor
	executorConfig := async.ExecutorConfig{
		DefaultTimeout:  timeouts.InitialWait,
		MaxLifetime:     10 * time.Minute,
		RetentionPeriod: 5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
	}
	executor := async.NewExecutor(executorConfig)
	
	return &ReplicateVideoHandler{
		generator: gen,
		storage:   store,
		client:    replicateClient,
		executor:  executor,
		timeouts:  timeouts,
		debug:     debug,
	}, nil
}

// CallTool handles execution of video tools
func (h *ReplicateVideoHandler) CallTool(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResponse, error) {
	// Note: Debug logging disabled in MCP mode to avoid stdout pollution
	
	switch req.Name {
	// Generation tools
	case "generate_video_from_text":
		return h.handleGenerateVideoFromText(ctx, req.Arguments)
	case "generate_video_from_image":
		return h.handleGenerateVideoFromImage(ctx, req.Arguments)
		
	// Async operation management
	case "continue_operation":
		return h.handleContinueOperation(ctx, req.Arguments)
		
	default:
		return nil, fmt.Errorf("unknown tool: %s", req.Name)
	}
}

// Stop cleanly shuts down the handler
func (h *ReplicateVideoHandler) Stop() {
	if h.executor != nil {
		h.executor.Stop()
	}
}

// Helper methods for building responses

// errorResponse creates an error response
func (h *ReplicateVideoHandler) errorResponse(operation, errorType, message string, details map[string]interface{}) (*protocol.CallToolResponse, error) {
	response := responses.BuildErrorResponse(operation, errorType, message, details)
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{Type: "text", Text: response},
		},
		IsError: true,
	}, nil
}

// successResponse creates a success response
func (h *ReplicateVideoHandler) successResponse(response string) (*protocol.CallToolResponse, error) {
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{Type: "text", Text: response},
		},
	}, nil
}

// processingResponse creates a processing response
func (h *ReplicateVideoHandler) processingResponse(operation, predictionID, storageID string, waitTime int) (*protocol.CallToolResponse, error) {
	response := responses.BuildProcessingResponse(operation, predictionID, storageID, waitTime)
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{Type: "text", Text: response},
		},
	}, nil
}