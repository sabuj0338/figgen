package agents

import (
	"context"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/sabujislam/figgen/internal/telemetry"
	"google.golang.org/api/option"
)

type GeminiProvider struct {
	client *genai.Client
	model  string
	ctx    context.Context
}

func NewGeminiProvider(ctx context.Context, modelName string) (*GeminiProvider, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	return &GeminiProvider{
		client: client,
		model:  modelName,
		ctx:    ctx,
	}, nil
}

func (g *GeminiProvider) GenerateJSON(ctx context.Context, req GenerateRequest) (*GenerateResult, error) {
	model := g.client.GenerativeModel(g.model)

	// JSON mode removes the need for prose instructions about valid JSON.
	model.GenerationConfig.ResponseMIMEType = "application/json"
	if req.MaxOutputTokens > 0 {
		mt := int32(req.MaxOutputTokens)
		model.GenerationConfig.MaxOutputTokens = &mt
	}

	// The static prefix goes into the system instruction. Gemini 2.x applies
	// implicit caching to a stable system instruction + leading context, so
	// keeping this identical across calls earns cache discounts automatically.
	if req.StaticPrefix != "" {
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(req.StaticPrefix)},
		}
	}

	resp, err := model.GenerateContent(ctx, genai.Text(req.Dynamic))
	if err != nil {
		return nil, fmt.Errorf("gemini generation failed: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}

	var fullText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			fullText += string(txt)
		}
	}
	if fullText == "" {
		return nil, fmt.Errorf("unexpected AI response format")
	}

	if Debug {
		_ = os.WriteFile("debug_fulltext.txt", []byte(fullText), 0644)
	}

	usage := Usage{}
	if resp.UsageMetadata != nil {
		usage.InputTokens = int(resp.UsageMetadata.PromptTokenCount)
		usage.OutputTokens = int(resp.UsageMetadata.CandidatesTokenCount)
		usage.CachedTokens = int(resp.UsageMetadata.CachedContentTokenCount)
	}

	telemetry.Add(telemetry.Record{
		Stage: req.Stage, Label: req.Label, Provider: "gemini", Model: g.model,
		InputTokens: usage.InputTokens, OutputTokens: usage.OutputTokens, CachedTokens: usage.CachedTokens,
	})

	finishReason := resp.Candidates[0].FinishReason
	return &GenerateResult{
		Text:         CleanJSONResponse(fullText),
		Usage:        usage,
		Truncated:    finishReason == genai.FinishReasonMaxTokens,
		FinishReason: finishReason.String(),
	}, nil
}
