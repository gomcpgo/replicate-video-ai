package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gomcpgo/replicate_video_ai/pkg/client"
	"github.com/gomcpgo/replicate_video_ai/pkg/generation"
	"github.com/gomcpgo/replicate_video_ai/pkg/responses"
	"github.com/gomcpgo/replicate_video_ai/pkg/storage"
)

const version = "1.0.0"

func main() {
	// Parse command line flags
	var (
		listModels     bool
		versionFlag    bool
		t2vModel       string
		i2vModel       string
		prompt         string
		imagePath      string
		resolution     string
		aspectRatio    string
		duration       int
		negativePrompt string
		outputFile     string
		testAsync      bool
		continueID     string
		debugMode      bool
	)

	flag.BoolVar(&listModels, "list", false, "List all available models")
	flag.BoolVar(&versionFlag, "version", false, "Show version information")
	flag.StringVar(&t2vModel, "t2v", "", "Generate text-to-video with specified model")
	flag.StringVar(&i2vModel, "i2v", "", "Generate image-to-video with specified model")
	flag.StringVar(&prompt, "p", "", "Prompt for video generation")
	flag.StringVar(&imagePath, "image", "", "Input image path for I2V")
	flag.StringVar(&resolution, "resolution", "", "Video resolution (480p, 720p, 1080p)")
	flag.StringVar(&aspectRatio, "aspect", "", "Aspect ratio (16:9, 9:16, 1:1)")
	flag.IntVar(&duration, "duration", 0, "Video duration in seconds (5 or 10, for Kling)")
	flag.StringVar(&negativePrompt, "negative", "", "Negative prompt (what to avoid)")
	flag.StringVar(&outputFile, "output", "", "Output filename")
	flag.BoolVar(&testAsync, "test-async", false, "Test async video generation flow")
	flag.StringVar(&continueID, "continue", "", "Continue checking a prediction ID")
	flag.BoolVar(&debugMode, "debug", false, "Enable debug mode")

	flag.Parse()

	if versionFlag {
		fmt.Printf("Replicate Video AI MCP Server v%s\n", version)
		return
	}

	// Terminal mode operations
	if listModels || t2vModel != "" || i2vModel != "" || testAsync || continueID != "" {
		// Get API key from environment
		apiKey := os.Getenv("REPLICATE_API_TOKEN")
		if apiKey == "" {
			log.Fatal("REPLICATE_API_TOKEN environment variable is required")
		}

		// Get root folder from environment or use default
		rootFolder := os.Getenv("REPLICATE_VIDEOS_ROOT_FOLDER")
		if rootFolder == "" {
			homeDir, _ := os.UserHomeDir()
			rootFolder = fmt.Sprintf("%s/Library/Application Support/Savant/replicate_video_ai", homeDir)
		}

		// Enable debug mode from env or flag
		if os.Getenv("REPLICATE_VIDEO_DEBUG") == "true" {
			debugMode = true
		}

		// Create components
		replicateClient := client.NewReplicateClient(apiKey, debugMode)
		store := storage.NewStorage(rootFolder, debugMode)
		gen := generation.NewGenerator(replicateClient, store, debugMode)

		ctx := context.Background()

		// Handle terminal mode operations
		if listModels {
			listAvailableModels()
			return
		}

		if t2vModel != "" {
			runTextToVideo(ctx, gen, t2vModel, prompt, resolution, aspectRatio, duration, negativePrompt, outputFile)
			return
		}

		if i2vModel != "" {
			runImageToVideo(ctx, gen, i2vModel, imagePath, prompt, resolution, duration, negativePrompt, outputFile)
			return
		}

		if continueID != "" {
			runContinue(ctx, gen, continueID, "")
			return
		}

		if testAsync {
			runAsyncTest(ctx, gen)
			return
		}

		return
	}

	// MCP Server mode
	fmt.Println("MCP Server mode not implemented yet. Use terminal mode for testing.")
	fmt.Println("Run with --help to see available options.")
}

func listAvailableModels() {
	fmt.Println("\n=== Available Video Models ===")
	fmt.Println("\nText-to-Video Models:")
	fmt.Println("  wan-t2v-fast    - Wan 2.2 Fast T2V (default, ~30s generation)")
	fmt.Println("  veo3            - Google Veo 3 (premium with audio)")
	fmt.Println("  kling-master    - Kling 2.1 Master (high quality, 5/10s duration)")
	fmt.Println()
	fmt.Println("Image-to-Video Models:")
	fmt.Println("  wan-i2v-fast    - Wan 2.2 Fast I2V (default for I2V)")
	fmt.Println("  veo3            - Google Veo 3 (preserves image style)")
	fmt.Println("  kling-master    - Kling 2.1 Master (high quality animation)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ./run.sh t2v wan-t2v-fast \"A car driving on beach\"")
	fmt.Println("  ./run.sh i2v wan-i2v-fast /path/to/image.jpg \"Zoom in slowly\"")
	fmt.Println()
}

func runTextToVideo(ctx context.Context, gen *generation.Generator, model, prompt, resolution, aspectRatio string, duration int, negativePrompt, outputFile string) {
	if prompt == "" {
		prompt = "A beautiful sunset over mountains with a lake in the foreground, golden hour lighting"
	}

	fmt.Printf("Generating text-to-video with %s...\n", model)
	fmt.Printf("Prompt: %s\n", prompt)

	params := generation.VideoParams{
		Prompt:         prompt,
		Model:          model,
		Resolution:     resolution,
		AspectRatio:    aspectRatio,
		Duration:       duration,
		NegativePrompt: negativePrompt,
		Filename:       outputFile,
	}

	result, err := gen.GenerateTextToVideo(ctx, params)
	if err != nil {
		log.Fatalf("Text-to-video generation failed: %v", err)
	}

	// Print async response
	response := responses.BuildProcessingResponse(
		"text_to_video",
		result.PredictionID,
		result.ID,
		30,
	)
	fmt.Println(response)
	fmt.Printf("\n✓ Generation started. Prediction ID: %s\n", result.PredictionID)
	fmt.Printf("Storage ID: %s\n", result.ID)
	fmt.Printf("\nTo check status, run:\n")
	fmt.Printf("  ./run.sh continue %s\n", result.PredictionID)
}

func runImageToVideo(ctx context.Context, gen *generation.Generator, model, imagePath, prompt, resolution string, duration int, negativePrompt, outputFile string) {
	if imagePath == "" {
		log.Fatal("Image path is required for image-to-video generation")
	}

	if prompt == "" {
		prompt = "Bring the image to life with natural motion"
	}

	fmt.Printf("Generating image-to-video with %s...\n", model)
	fmt.Printf("Input image: %s\n", imagePath)
	fmt.Printf("Prompt: %s\n", prompt)

	params := generation.VideoParams{
		Prompt:         prompt,
		Model:          model,
		ImagePath:      imagePath,
		Resolution:     resolution,
		Duration:       duration,
		NegativePrompt: negativePrompt,
		Filename:       outputFile,
	}

	result, err := gen.GenerateImageToVideo(ctx, params)
	if err != nil {
		log.Fatalf("Image-to-video generation failed: %v", err)
	}

	// Print async response
	response := responses.BuildProcessingResponse(
		"image_to_video",
		result.PredictionID,
		result.ID,
		30,
	)
	fmt.Println(response)
	fmt.Printf("\n✓ Generation started. Prediction ID: %s\n", result.PredictionID)
	fmt.Printf("Storage ID: %s\n", result.ID)
	fmt.Printf("\nTo check status, run:\n")
	fmt.Printf("  ./run.sh continue %s\n", result.PredictionID)
}

func runContinue(ctx context.Context, gen *generation.Generator, predictionID, storageID string) {
	fmt.Printf("Checking status of prediction %s...\n", predictionID)

	// If no storage ID provided, use a placeholder
	if storageID == "" {
		storageID = "unknown"
	}

	// Wait up to 60 seconds
	result, err := gen.ContinueGeneration(ctx, predictionID, storageID, 60*time.Second)
	if err != nil {
		// Check if it's still processing
		if result != nil && result.Status == "processing" {
			fmt.Printf("Still processing... Try again later.\n")
			return
		}
		log.Fatalf("Failed to check status: %v", err)
	}

	if result.Status == "completed" && result.FilePath != "" {
		response := responses.BuildSuccessResponse(
			"continue_operation",
			result.ID,
			map[string]string{
				"output": result.FilePath,
			},
			map[string]string{},
			map[string]interface{}{},
			map[string]interface{}{
				"generation_time": result.Metrics.GenerationTime,
				"file_size":       result.Metrics.FileSize,
			},
			result.PredictionID,
		)
		fmt.Println(response)
		fmt.Printf("\n✓ Video saved to: %s\n", result.FilePath)
	} else {
		fmt.Printf("Status: %s\n", result.Status)
	}
}

func runAsyncTest(ctx context.Context, gen *generation.Generator) {
	fmt.Println("\n=== Testing Async Video Generation Flow ===")
	fmt.Println()

	// Step 1: Start generation
	fmt.Println("Step 1: Starting text-to-video generation...")
	params := generation.VideoParams{
		Prompt:      "A serene lake at sunset with birds flying overhead",
		Model:       "wan-t2v-fast",
		Resolution:  "480p",
		AspectRatio: "16:9",
	}

	result, err := gen.GenerateTextToVideo(ctx, params)
	if err != nil {
		log.Fatalf("Failed to start generation: %v", err)
	}

	fmt.Printf("✓ Generation started\n")
	fmt.Printf("  Prediction ID: %s\n", result.PredictionID)
	fmt.Printf("  Storage ID: %s\n", result.ID)
	fmt.Printf("  Status: %s\n", result.Status)
	fmt.Println()

	// Step 2: Wait and check status
	fmt.Println("Step 2: Waiting 10 seconds before checking status...")
	time.Sleep(10 * time.Second)

	fmt.Println("Step 3: Checking generation status...")
	finalResult, err := gen.ContinueGeneration(ctx, result.PredictionID, result.ID, 2*time.Minute)
	if err != nil {
		fmt.Printf("Generation not complete yet: %v\n", err)
		if finalResult != nil {
			fmt.Printf("Current status: %s\n", finalResult.Status)
		}
		fmt.Println("\nTry running the continue command manually:")
		fmt.Printf("  ./run.sh continue %s\n", result.PredictionID)
		return
	}

	// Step 3: Show results
	if finalResult.Status == "completed" && finalResult.FilePath != "" {
		fmt.Printf("✓ Video generation completed!\n")
		fmt.Printf("  Output path: %s\n", finalResult.FilePath)
		fmt.Printf("  File size: %d bytes\n", finalResult.Metrics.FileSize)
		fmt.Printf("  Generation time: %.2f seconds\n", finalResult.Metrics.GenerationTime)

		// Print formatted response
		response := responses.BuildSuccessResponse(
			"async_test",
			finalResult.ID,
			map[string]string{
				"output": finalResult.FilePath,
			},
			map[string]string{
				"name": "wan-t2v-fast",
			},
			convertParamsToMap(params),
			map[string]interface{}{
				"generation_time": finalResult.Metrics.GenerationTime,
				"file_size":       finalResult.Metrics.FileSize,
			},
			finalResult.PredictionID,
		)
		fmt.Println("\nFormatted response:")
		fmt.Println(response)
	} else {
		fmt.Printf("Unexpected status: %s\n", finalResult.Status)
	}

	fmt.Println("\n=== Async Test Complete ===")
}

// Helper to convert VideoParams to map for response
func convertParamsToMap(p generation.VideoParams) map[string]interface{} {
	params := make(map[string]interface{})
	params["prompt"] = p.Prompt
	if p.Resolution != "" {
		params["resolution"] = p.Resolution
	}
	if p.AspectRatio != "" {
		params["aspect_ratio"] = p.AspectRatio
	}
	if p.Duration > 0 {
		params["duration"] = p.Duration
	}
	if p.NegativePrompt != "" {
		params["negative_prompt"] = p.NegativePrompt
	}
	return params
}