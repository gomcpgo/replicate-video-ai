#!/bin/bash

# Source .env file if it exists
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

case "$1" in
    "build")
        echo "Building Replicate Video AI server..."
        go build -o bin/replicate-video-ai ./cmd
        ;;
    
    "test")
        echo "Running tests..."
        go test ./pkg/...
        ;;
    
    "list-models"|"list")
        # Check for API token for terminal operations
        if [ -z "$REPLICATE_API_TOKEN" ]; then
            echo "Error: REPLICATE_API_TOKEN environment variable is not set"
            echo "Please set it in your environment or create a .env file"
            exit 1
        fi
        go run ./cmd -list
        ;;
    
    "t2v")
        # Text-to-video generation
        if [ -z "$REPLICATE_API_TOKEN" ]; then
            echo "Error: REPLICATE_API_TOKEN environment variable is not set"
            echo "Please set it in your environment or create a .env file"
            exit 1
        fi
        if [ -z "$2" ]; then
            echo "Usage: ./run.sh t2v <model> [prompt]"
            echo "Models: wan-t2v-fast, veo3, kling-master"
            exit 1
        fi
        go run ./cmd -t2v "$2" -p "${3:-}"
        ;;
    
    "i2v")
        # Image-to-video generation
        if [ -z "$REPLICATE_API_TOKEN" ]; then
            echo "Error: REPLICATE_API_TOKEN environment variable is not set"
            echo "Please set it in your environment or create a .env file"
            exit 1
        fi
        if [ -z "$2" ] || [ -z "$3" ]; then
            echo "Usage: ./run.sh i2v <model> <image_path> [prompt]"
            echo "Models: wan-i2v-fast, veo3, kling-master"
            exit 1
        fi
        go run ./cmd -i2v "$2" -image "$3" -p "${4:-}"
        ;;
    
    "continue")
        # Continue checking prediction
        if [ -z "$REPLICATE_API_TOKEN" ]; then
            echo "Error: REPLICATE_API_TOKEN environment variable is not set"
            echo "Please set it in your environment or create a .env file"
            exit 1
        fi
        if [ -z "$2" ]; then
            echo "Usage: ./run.sh continue <prediction_id>"
            exit 1
        fi
        go run ./cmd -continue "$2"
        ;;
    
    "test-async")
        if [ -z "$REPLICATE_API_TOKEN" ]; then
            echo "Error: REPLICATE_API_TOKEN environment variable is not set"
            echo "Please set it in your environment or create a .env file"
            exit 1
        fi
        go run ./cmd -test-async
        ;;
    
    "run"|"server")
        echo "Starting MCP server..."
        go run ./cmd
        ;;
    
    "debug")
        # Run with debug mode
        export REPLICATE_VIDEO_DEBUG=true
        shift
        ./run.sh "$@"
        ;;
    
    *)
        echo "Usage: $0 {build|test|list-models|t2v|i2v|continue|test-async|run|debug}"
        echo ""
        echo "Commands:"
        echo "  build       - Build the server binary"
        echo "  test        - Run tests"
        echo "  list-models - List available video models"
        echo "  t2v         - Generate text-to-video"
        echo "  i2v         - Generate image-to-video"
        echo "  continue    - Continue checking a prediction"
        echo "  test-async  - Test async generation flow"
        echo "  run         - Start MCP server"
        echo "  debug       - Run any command with debug mode"
        echo ""
        echo "Examples:"
        echo "  ./run.sh t2v wan-t2v-fast \"A sunset over the ocean\""
        echo "  ./run.sh i2v wan-i2v-fast images/car.webp \"Make the car drive\""
        echo "  ./run.sh continue abc123xyz"
        echo "  ./run.sh debug t2v wan-t2v-fast \"Test prompt\""
        ;;
esac

