package main

import (
	"context"
	"fmt"
	"os"

	"github.com/instructor-ai/instructor-go/pkg/instructor"
	"google.golang.org/genai"
)

type ImageInfo struct {
	Description string   `json:"description" jsonschema:"title=Description,description=What is visible in the image"`
	Objects     []string `json:"objects" jsonschema:"title=Objects,description=List of objects detected in the image"`
	Colors      []string `json:"colors" jsonschema:"title=Colors,description=Main colors present in the image"`
	Mood        string   `json:"mood" jsonschema:"title=Mood,description=Overall mood of the image,enum=bright,enum=dark,enum=cheerful,enum=melancholic"`
}

func main() {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  os.Getenv("GOOGLE_API_KEY"),
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create Gemini client: %v", err))
	}

	instructorClient := instructor.FromGemini(
		client,
		instructor.WithMode(instructor.ModeJSON),
		instructor.WithMaxRetries(3),
	)

	imageBytes, err := os.ReadFile("examples/gemini/sample.jpg")
	if err != nil {
		fmt.Printf("Error reading image: %v\n", err)
		fmt.Println("Please add a sample image at examples/gemini/sample.jpg")
		return
	}

	contents := []*genai.Content{
		{
			Parts: []*genai.Part{
				{Text: "Analyze this image and provide structured information about what you see."},
				{InlineData: &genai.Blob{
					Data:     imageBytes,
					MIMEType: "image/jpeg",
				}},
			},
			Role: "user",
		},
	}

	request := instructor.GeminiRequest{
		Model:    "gemini-2.5-flash",
		Contents: contents,
	}

	var imageInfo ImageInfo
	_, err = instructorClient.CreateChatCompletion(ctx, request, &imageInfo)
	if err != nil {
		fmt.Printf("Error analyzing image: %v\n", err)
		return
	}

	fmt.Println("📸 Image Analysis Results:")
	fmt.Printf("Description: %s\n", imageInfo.Description)
	fmt.Printf("Mood: %s\n", imageInfo.Mood)
	fmt.Printf("Objects: %v\n", imageInfo.Objects)
	fmt.Printf("Colors: %v\n", imageInfo.Colors)
	/*
		📸 Image Analysis Results:
		Description: The image features three cartoon gopher characters, the mascot for the Go programming language, arranged horizontally against a light blue background. Each gopher is light blue with large white eyes and black pupils, a small light brown or beige snout with visible teeth, and small limbs. The left gopher is angled to its right, showing one eye; the center gopher faces directly forward, showing both eyes; and the right gopher is angled to its left, showing one eye. The art style is simple and friendly with a slight textured or crayon-like appearance.
		Mood: cheerful
		Objects: [gopher cartoon character mascot]
		Colors: [light blue blue white black beige]
	*/
}
