package handler

import (
	"context"
	"encoding/json"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

// ListTools returns the available MCP tools
func (h *ReplicateVideoHandler) ListTools(ctx context.Context) (*protocol.ListToolsResponse, error) {
	tools := []protocol.Tool{
		{
			Name:        "generate_video_from_text",
			Description: "Generate a video from a text prompt. Models: wan-t2v-fast (default, fast/cheap), veo3 (premium with audio), kling-master (high quality, supports 5/10s duration)",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"prompt": {
						"type": "string",
						"description": "Text description of the video to generate"
					},
					"model": {
						"type": "string",
						"description": "Model to use: wan-t2v-fast, veo3, kling-master",
						"default": "wan-t2v-fast"
					},
					"duration": {
						"type": "integer",
						"description": "Video duration in seconds (5 or 10, only for kling-master)",
						"minimum": 5,
						"maximum": 10
					},
					"resolution": {
						"type": "string",
						"description": "Video resolution: 480p, 720p, 1080p (model-dependent)",
						"default": "720p"
					},
					"aspect_ratio": {
						"type": "string",
						"description": "Aspect ratio: 16:9, 9:16, 1:1",
						"default": "16:9"
					},
					"negative_prompt": {
						"type": "string",
						"description": "What to avoid in the video (supported by veo3, kling-master)"
					},
					"filename": {
						"type": "string",
						"description": "Optional output filename"
					}
				},
				"required": ["prompt"]
			}`),
		},
		{
			Name:        "generate_video_from_image",
			Description: "Generate a video from an image with motion prompt. Models: wan-i2v-fast (default, fast/cheap), veo3 (preserves style), kling-master (high quality, 5/10s duration)",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"image_path": {
						"type": "string",
						"description": "Path to the input image (local file path)"
					},
					"prompt": {
						"type": "string",
						"description": "Description of how to animate the image"
					},
					"model": {
						"type": "string",
						"description": "Model to use: wan-i2v-fast, veo3, kling-master",
						"default": "wan-i2v-fast"
					},
					"duration": {
						"type": "integer",
						"description": "Video duration in seconds (only for kling-master: 5 or 10)"
					},
					"resolution": {
						"type": "string",
						"description": "Video resolution (model-dependent)",
						"default": "720p"
					},
					"negative_prompt": {
						"type": "string",
						"description": "What to avoid in the video (supported by veo3, kling-master)"
					},
					"filename": {
						"type": "string",
						"description": "Optional output filename"
					}
				},
				"required": ["image_path", "prompt"]
			}`),
		},
		{
			Name:        "continue_operation",
			Description: "Continue checking status of async video generation",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"prediction_id": {
						"type": "string",
						"description": "The prediction ID from initial generation"
					},
					"wait_time": {
						"type": "number",
						"description": "How long to wait in seconds (5-60)",
						"default": 30
					}
				},
				"required": ["prediction_id"]
			}`),
		},
	}

	return &protocol.ListToolsResponse{
		Tools: tools,
	}, nil
}