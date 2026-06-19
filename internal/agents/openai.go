package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/sabujislam/figgen/internal/telemetry"
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
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens        int `json:"prompt_tokens"`
		CompletionTokens    int `json:"completion_tokens"`
		PromptTokensDetails struct {
			CachedTokens int `json:"cached_tokens"`
		} `json:"prompt_tokens_details"`
	} `json:"usage"`
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

func (o *OpenAIProvider) GenerateJSON(ctx context.Context, req GenerateRequest) (*GenerateResult, error) {
	maxTokens := req.MaxOutputTokens
	if maxTokens == 0 {
		maxTokens = 8192
	}

	// System message carries the static prefix first; OpenAI automatically
	// caches identical prompt prefixes over ~1k tokens server-side.
	reqBody := openAIRequest{
		Model: o.model,
		Messages: []openAIMessage{
			{Role: "system", Content: req.StaticPrefix},
			{Role: "user", Content: req.Dynamic},
		},
		ResponseFormat: map[string]interface{}{"type": "json_object"},
		MaxTokens:      maxTokens,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	reqURL := baseURL + "/chat/completions"

	httpReq, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var resData openAIResponse
	if err := json.Unmarshal(bodyBytes, &resData); err != nil {
		return nil, fmt.Errorf("failed to parse openai response: %w", err)
	}

	if resData.Error != nil {
		return nil, fmt.Errorf("openai error: %s", resData.Error.Message)
	}

	if len(resData.Choices) == 0 {
		return nil, fmt.Errorf("no response choices from OpenAI")
	}

	usage := Usage{
		InputTokens:  resData.Usage.PromptTokens,
		OutputTokens: resData.Usage.CompletionTokens,
		CachedTokens: resData.Usage.PromptTokensDetails.CachedTokens,
	}

	telemetry.Add(telemetry.Record{
		Stage: req.Stage, Label: req.Label, Provider: "openai", Model: o.model,
		InputTokens: usage.InputTokens, OutputTokens: usage.OutputTokens, CachedTokens: usage.CachedTokens,
	})

	finishReason := resData.Choices[0].FinishReason
	return &GenerateResult{
		Text:         CleanJSONResponse(resData.Choices[0].Message.Content),
		Usage:        usage,
		Truncated:    finishReason == "length",
		FinishReason: finishReason,
	}, nil
}
