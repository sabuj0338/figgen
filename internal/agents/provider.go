package agents

import (
	"context"
	"fmt"
	"strings"
)

// Stage identifiers for telemetry/routing.
const (
	StagePlan = "plan"
	StageCode = "code"
)

// Debug toggles writing of raw model responses to disk for troubleshooting.
// It is set from the --debug CLI flag.
var Debug = false

// GenerateRequest is a structured prompt split into a large, cacheable static
// prefix and a small per-task dynamic body. Keeping the static portion byte
// identical across calls maximizes provider prompt/context cache hits.
type GenerateRequest struct {
	// StaticPrefix holds rules, config, instructions, and the output schema.
	// It should be identical across as many calls as possible.
	StaticPrefix string
	// Dynamic holds the per-task content (Figma node context, plan, etc.).
	Dynamic string
	// Stage is StagePlan or StageCode (telemetry + routing).
	Stage string
	// Label identifies the task for telemetry (component/page/chunk name).
	Label string
	// MaxOutputTokens caps generated tokens (0 = provider default).
	MaxOutputTokens int
}

// Usage captures token accounting reported by the provider.
type Usage struct {
	InputTokens  int
	OutputTokens int
	CachedTokens int
}

// GenerateResult is the model output plus usage accounting.
type GenerateResult struct {
	Text  string
	Usage Usage
	// Truncated is true when the provider stopped generating because it hit the
	// output token cap (Anthropic stop_reason "max_tokens", OpenAI finish_reason
	// "length", Gemini MaxTokens, Ollama done_reason "length"). A truncated
	// response is almost always invalid JSON, so callers should treat it as a
	// recoverable error rather than a parse bug.
	Truncated bool
	// FinishReason is the raw provider-reported stop reason, kept for diagnostics.
	FinishReason string
}

// LLMProvider is the universal interface for any AI model.
type LLMProvider interface {
	GenerateJSON(ctx context.Context, req GenerateRequest) (*GenerateResult, error)
}

// CombinedPrompt joins the static prefix and dynamic body for providers that
// take a single prompt string (e.g. Ollama).
func (r GenerateRequest) CombinedPrompt() string {
	if r.StaticPrefix == "" {
		return r.Dynamic
	}
	return r.StaticPrefix + "\n\n" + r.Dynamic
}

// CleanJSONResponse removes markdown fences from JSON outputs (fallback for
// providers/models that ignore JSON mode).
func CleanJSONResponse(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	return strings.TrimSpace(raw)
}

// NewProvider returns the correct LLMProvider for the given provider/model.
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
