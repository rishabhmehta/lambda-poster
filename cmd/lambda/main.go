package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"poster/internal/poster"
)

// Request represents the incoming request body
type Request struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
}

// Response represents the response body
type Response struct {
	Image   string `json:"image,omitempty"`
	Error   string `json:"error,omitempty"`
}

var generator *poster.Generator

func init() {
	var err error
	generator, err = poster.NewGenerator()
	if err != nil {
		log.Fatalf("failed to initialize generator: %v", err)
	}
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req Request
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return errorResponse(400, "invalid request body")
	}

	if req.Name == "" {
		return errorResponse(400, "name is required")
	}

	if req.AvatarURL == "" {
		return errorResponse(400, "avatarUrl is required")
	}

	imageBase64, err := generator.Generate(req.Name, req.AvatarURL)
	if err != nil {
		log.Printf("generation error: %v", err)
		return errorResponse(500, "failed to generate poster")
	}

	resp := Response{Image: imageBase64}
	body, _ := json.Marshal(resp)

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}, nil
}

func errorResponse(status int, message string) (events.APIGatewayProxyResponse, error) {
	resp := Response{Error: message}
	body, _ := json.Marshal(resp)

	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}, nil
}

func main() {
	lambda.Start(handler)
}
