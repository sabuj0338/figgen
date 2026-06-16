package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type AnthropicProvider struct {
	model  string
	apiKey string
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewAnthropicProvider(modelName string) (*AnthropicProvider, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable is required")
	}

	return &AnthropicProvider{
		model:  modelName,
		apiKey: apiKey,
	}, nil
}

func (a *AnthropicProvider) GenerateJSON(ctx context.Context, prompt string) (string, error) {
	prompt = prompt + "\n\nIMPORTANT: You must respond ONLY with valid JSON and no other text."
	
	reqBody := anthropicRequest{
		Model:     a.model,
		MaxTokens: 8192,
		System:    "You are an expert AI software architect and coder. You MUST respond with ONLY valid JSON.",
		Messages: []anthropicMessage{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("anthropic request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var resData anthropicResponse
	if err := json.Unmarshal(bodyBytes, &resData); err != nil {
		return "", fmt.Errorf("failed to parse anthropic response: %w", err)
	}

	if resData.Error != nil {
		return "", fmt.Errorf("anthropic error: %s", resData.Error.Message)
	}

	if len(resData.Content) == 0 {
		return "", fmt.Errorf("no response content from Anthropic")
	}

	return CleanJSONResponse(resData.Content[0].Text), nil
}
