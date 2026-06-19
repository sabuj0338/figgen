package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/sabujislam/figgen/internal/telemetry"
)

type AnthropicProvider struct {
	model  string
	apiKey string
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicTextBlock struct {
	Type         string            `json:"type"`
	Text         string            `json:"text"`
	CacheControl map[string]string `json:"cache_control,omitempty"`
}

type anthropicRequest struct {
	Model     string               `json:"model"`
	MaxTokens int                  `json:"max_tokens"`
	System    []anthropicTextBlock `json:"system,omitempty"`
	Messages  []anthropicMessage   `json:"messages"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens              int `json:"input_tokens"`
		OutputTokens             int `json:"output_tokens"`
		CacheReadInputTokens     int `json:"cache_read_input_tokens"`
		CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	} `json:"usage"`
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

// anthropicMaxTokensCeiling returns the hard output-token limit for a model.
// Requesting more than a model supports returns an HTTP 400, so we clamp the
// caller's request to the model's known ceiling. Legacy Claude 3 / 3.5 models
// cap at 8192; 3.7 and 4.x models support far larger outputs.
func anthropicMaxTokensCeiling(model string) int {
	switch {
	case strings.Contains(model, "claude-3-7"),
		strings.Contains(model, "claude-sonnet-4"),
		strings.Contains(model, "claude-opus-4"),
		strings.Contains(model, "claude-haiku-4"),
		strings.Contains(model, "claude-4"):
		return 64000
	default:
		// Claude 3 and 3.5 family.
		return 8192
	}
}

func (a *AnthropicProvider) GenerateJSON(ctx context.Context, req GenerateRequest) (*GenerateResult, error) {
	maxTokens := req.MaxOutputTokens
	if maxTokens == 0 {
		maxTokens = 8192
	}
	if ceiling := anthropicMaxTokensCeiling(a.model); maxTokens > ceiling {
		maxTokens = ceiling
	}

	// Mark the static prefix as a cache breakpoint so repeated calls read it
	// from Anthropic's prompt cache instead of re-billing full price.
	system := []anthropicTextBlock{
		{
			Type:         "text",
			Text:         req.StaticPrefix,
			CacheControl: map[string]string{"type": "ephemeral"},
		},
	}

	reqBody := anthropicRequest{
		Model:     a.model,
		MaxTokens: maxTokens,
		System:    system,
		Messages: []anthropicMessage{
			{Role: "user", Content: req.Dynamic},
			// Prefill forces the model to start a JSON object immediately.
			{Role: "assistant", Content: "{"},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", a.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var resData anthropicResponse
	if err := json.Unmarshal(bodyBytes, &resData); err != nil {
		return nil, fmt.Errorf("failed to parse anthropic response: %w", err)
	}

	if resData.Error != nil {
		return nil, fmt.Errorf("anthropic error: %s", resData.Error.Message)
	}

	if len(resData.Content) == 0 {
		return nil, fmt.Errorf("no response content from Anthropic")
	}

	// Re-attach the prefilled opening brace that started the JSON object.
	text := resData.Content[0].Text
	if !strings.HasPrefix(strings.TrimSpace(text), "{") {
		text = "{" + text
	}

	usage := Usage{
		InputTokens:  resData.Usage.InputTokens + resData.Usage.CacheCreationInputTokens + resData.Usage.CacheReadInputTokens,
		OutputTokens: resData.Usage.OutputTokens,
		CachedTokens: resData.Usage.CacheReadInputTokens,
	}

	telemetry.Add(telemetry.Record{
		Stage: req.Stage, Label: req.Label, Provider: "anthropic", Model: a.model,
		InputTokens: usage.InputTokens, OutputTokens: usage.OutputTokens, CachedTokens: usage.CachedTokens,
	})

	return &GenerateResult{
		Text:         CleanJSONResponse(text),
		Usage:        usage,
		Truncated:    resData.StopReason == "max_tokens",
		FinishReason: resData.StopReason,
	}, nil
}
