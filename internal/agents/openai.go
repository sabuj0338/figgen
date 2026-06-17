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

type OpenAIProvider struct {
	model  string
	apiKey string
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model          string                 `json:"model"`
	Messages       []openAIMessage        `json:"messages"`
	ResponseFormat map[string]interface{} `json:"response_format,omitempty"`
	MaxTokens      int                    `json:"max_tokens,omitempty"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewOpenAIProvider(modelName string) (*OpenAIProvider, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	return &OpenAIProvider{
		model:  modelName,
		apiKey: apiKey,
	}, nil
}

func (o *OpenAIProvider) GenerateJSON(ctx context.Context, prompt string) (string, error) {
	reqBody := openAIRequest{
		Model: o.model,
		Messages: []openAIMessage{
			{Role: "system", Content: "You are an expert AI software architect and coder. You MUST respond with ONLY valid JSON."},
			{Role: "user", Content: prompt},
		},
		MaxTokens: 8192,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	reqURL := baseURL + "/chat/completions"

	fmt.Printf("DEBUG: Using baseURL: %s\n", baseURL)

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var resData openAIResponse
	if err := json.Unmarshal(bodyBytes, &resData); err != nil {
		return "", fmt.Errorf("failed to parse openai response: %w", err)
	}

	if resData.Error != nil {
		return "", fmt.Errorf("openai error: %s", resData.Error.Message)
	}

	if len(resData.Choices) == 0 {
		return "", fmt.Errorf("no response choices from OpenAI")
	}

	return CleanJSONResponse(resData.Choices[0].Message.Content), nil
}
