# Replicate Video AI MCP Server

A Model Context Protocol (MCP) server for generating videos using Replicate's video AI models.

## Features

- **Text-to-Video Generation**: Create videos from text prompts
- **Image-to-Video Generation**: Animate still images with motion prompts
- **Multiple Models**: Support for Wan 2.2, Google Veo 3, and Kling 2.1
- **Async Operations**: Handle long-running video generation with status checking
- **Terminal Mode**: Built-in CLI for testing without MCP overhead

## Supported Models

| Alias | Model | Type | Features |
|-------|-------|------|----------|
| `wan-t2v-fast` | Wan 2.2 Fast T2V | Text-to-Video | Fast (~30s), affordable |
| `wan-i2v-fast` | Wan 2.2 Fast I2V | Image-to-Video | Fast animation |
| `veo3` | Google Veo 3 | Both | Premium quality with audio |
| `kling-master` | Kling 2.1 Master | Both | High quality, 5/10s duration |

## Setup

1. Set your Replicate API token:
```bash
export REPLICATE_API_TOKEN=your_token_here
```

Or create a `.env` file:
```bash
REPLICATE_API_TOKEN=your_token_here
```

2. Make the run script executable:
```bash
chmod +x run.sh
```

## Usage

### Terminal Mode

List available models:
```bash
./run.sh list-models
```

Generate text-to-video:
```bash
./run.sh t2v wan-t2v-fast "A sunset over the ocean"
./run.sh t2v veo3 "Dancing robot" --resolution 1080p
./run.sh t2v kling-master "City timelapse" --duration 10
```

Generate image-to-video:
```bash
./run.sh i2v wan-i2v-fast images/car.webp "Car driving forward"
./run.sh i2v veo3 images/test.jpg "Add rain effect"
```

Check generation status:
```bash
./run.sh continue <prediction_id>
```

Test async flow:
```bash
./run.sh test-async
```

Debug mode:
```bash
./run.sh debug t2v wan-t2v-fast "Test prompt"
```

### MCP Server Mode

Start the MCP server:
```bash
./run.sh run
```

## MCP Tools

### generate_video_from_text
Generate a video from a text prompt.

Parameters:
- `prompt` (required): Text description of the video
- `model`: Model to use (default: wan-t2v-fast)
- `resolution`: Video resolution (480p, 720p, 1080p)
- `aspect_ratio`: Aspect ratio (16:9, 9:16, 1:1)
- `duration`: Duration in seconds (for Kling only)
- `negative_prompt`: What to avoid (for Veo3, Kling)

### generate_video_from_image
Generate a video from an image with motion prompt.

Parameters:
- `image_path` (required): Path to input image
- `prompt` (required): How to animate the image
- `model`: Model to use (default: wan-i2v-fast)
- `resolution`: Video resolution
- `duration`: Duration (for Kling only)
- `negative_prompt`: What to avoid

### continue_operation
Check status of async video generation.

Parameters:
- `prediction_id` (required): The prediction ID
- `wait_time`: How long to wait (5-60 seconds)

## Output

Videos are saved to:
```
~/Library/Application Support/Savant/replicate_video_ai/<storage_id>/
├── video.mp4        # Generated video
├── metadata.yaml    # Generation parameters
└── input.jpg        # Input image (if I2V)
```

## Environment Variables

- `REPLICATE_API_TOKEN` (required): Your Replicate API token
- `REPLICATE_VIDEOS_ROOT_FOLDER`: Custom output directory
- `REPLICATE_VIDEO_DEBUG`: Enable debug mode (true/false)
- `REPLICATE_VIDEO_DEFAULT_TIMEOUT`: Default timeout in seconds
- `REPLICATE_VIDEO_POLL_INTERVAL`: Status check interval

## Development

Build the server:
```bash
./run.sh build
```

Run tests:
```bash
./run.sh test
```

## License

MIT