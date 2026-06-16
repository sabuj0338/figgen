package agents

import (
	"context"
	"fmt"
	"strings"
)

// LLMProvider is the universal interface for any AI model
type LLMProvider interface {
	GenerateJSON(ctx context.Context, prompt string) (string, error)
}

// CleanJSONResponse is a utility to remove markdown formatting from JSON outputs
func CleanJSONResponse(raw string) string {
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	return strings.TrimSpace(raw)
}

// NewProvider is a factory that returns the correct LLMProvider based on the user's choice
func NewProvider(ctx context.Context, providerName string, modelName string) (LLMProvider, error) {
	switch providerName {
	case "gemini":
		if modelName == "" {
			modelName = "gemini-2.5-flash"
		}
		return NewGeminiProvider(ctx, modelName)
	case "ollama":
		if modelName == "" {
			modelName = "qwen2.5-coder:1.5b"
		}
		return NewOllamaProvider(modelName)
	case "openai":
		if modelName == "" {
			modelName = "gpt-4o-mini"
		}
		return NewOpenAIProvider(modelName)
	case "anthropic":
		if modelName == "" {
			modelName = "claude-3-5-sonnet-20240620"
		}
		return NewAnthropicProvider(modelName)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}
}
