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

type OllamaProvider struct {
	baseURL string
	model   string
}

type ollamaOptions struct {
	NumCtx     int `json:"num_ctx"`
	NumPredict int `json:"num_predict"`
}

type ollamaRequest struct {
	Model   string        `json:"model"`
	Prompt  string        `json:"prompt"`
	Stream  bool          `json:"stream"`
	Format  string        `json:"format,omitempty"`
	Options ollamaOptions `json:"options"`
}

type ollamaResponse struct {
	Response        string `json:"response"`
	DoneReason      string `json:"done_reason"`
	PromptEvalCount int    `json:"prompt_eval_count"`
	EvalCount       int    `json:"eval_count"`
	Error           string `json:"error,omitempty"`
}

func NewOllamaProvider(modelName string) (*OllamaProvider, error) {
	baseURL := os.Getenv("OLLAMA_HOST")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &OllamaProvider{
		baseURL: baseURL,
		model:   modelName,
	}, nil
}

func (o *OllamaProvider) GenerateJSON(ctx context.Context, req GenerateRequest) (*GenerateResult, error) {
	numPredict := req.MaxOutputTokens
	if numPredict == 0 {
		numPredict = 4096
	}

	reqBody := ollamaRequest{
		Model:  o.model,
		Prompt: req.CombinedPrompt(),
		Stream: false,
		Format: "json", // Forces JSON output in modern Ollama versions
		Options: ollamaOptions{
			NumCtx:     16384,
			NumPredict: numPredict,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed (is Ollama running?): %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var resData ollamaResponse
	if err := json.Unmarshal(bodyBytes, &resData); err != nil {
		return nil, fmt.Errorf("failed to parse ollama response: %w", err)
	}

	if resData.Error != "" {
		return nil, fmt.Errorf("ollama error: %s", resData.Error)
	}

	usage := Usage{
		InputTokens:  resData.PromptEvalCount,
		OutputTokens: resData.EvalCount,
	}

	telemetry.Add(telemetry.Record{
		Stage: req.Stage, Label: req.Label, Provider: "ollama", Model: o.model,
		InputTokens: usage.InputTokens, OutputTokens: usage.OutputTokens,
	})

	return &GenerateResult{
		Text:         CleanJSONResponse(resData.Response),
		Usage:        usage,
		Truncated:    resData.DoneReason == "length",
		FinishReason: resData.DoneReason,
	}, nil
}
