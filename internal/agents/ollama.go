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

type OllamaProvider struct {
	baseURL string
	model   string
}

type ollamaOptions struct {
	NumCtx     int `json:"num_ctx"`
	NumPredict int `json:"num_predict"`
}

type ollamaRequest struct {
	Model   string         `json:"model"`
	Prompt  string         `json:"prompt"`
	Stream  bool           `json:"stream"`
	Format  string         `json:"format,omitempty"`
	Options ollamaOptions  `json:"options"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
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

func (o *OllamaProvider) GenerateJSON(ctx context.Context, prompt string) (string, error) {
	// Add instruction to output JSON if missing
	prompt = prompt + "\n\nIMPORTANT: You must respond ONLY with valid JSON."

	reqBody := ollamaRequest{
		Model:  o.model,
		Prompt: prompt,
		Stream: false,
		Format: "json", // Forces JSON output in modern Ollama versions
		Options: ollamaOptions{
			NumCtx:     16384, // Maximize context window for Figma/React syntax
			NumPredict: 4096,  // Allow massive code outputs without truncation
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request failed (is Ollama running?): %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var resData ollamaResponse
	if err := json.Unmarshal(bodyBytes, &resData); err != nil {
		return "", fmt.Errorf("failed to parse ollama response: %w", err)
	}

	if resData.Error != "" {
		return "", fmt.Errorf("ollama error: %s", resData.Error)
	}

	return CleanJSONResponse(resData.Response), nil
}
