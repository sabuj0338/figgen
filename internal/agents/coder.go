package agents

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sabujislam/figgen/internal/config"
)

type CoderResponse struct {
	Code string `json:"code"`
}

func RunCoderForComponent(ctx context.Context, ai LLMProvider, cfg *config.Config, comp ComponentPlan) (string, error) {
	configJSON, _ := json.MarshalIndent(cfg, "", "  ")

	prompt := fmt.Sprintf(`You are an expert Next.js/React Developer.
Write the exact TypeScript React code for the following component plan.
Do NOT output markdown formatting like "` + "```tsx" + `". ONLY output pure JSON containing the exact file content in the "code" field.

Configuration Rules:
%s

Component Plan:
Name: %s
Description: %s
Props: %v
Is Shadcn: %v

Output JSON format:
{
  "code": "import React from 'react';\n..."
}`, string(configJSON), comp.Name, comp.Description, comp.Props, comp.IsShadcn)

	rawJSON, err := ai.GenerateJSON(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("ai coder failed: %w", err)
	}

	var codeResp CoderResponse
	if err := json.Unmarshal([]byte(rawJSON), &codeResp); err != nil {
		return "", fmt.Errorf("failed to parse AI JSON code response: %w\nRaw: %s", err, rawJSON)
	}

	return codeResp.Code, nil
}

func RunCoderForPage(ctx context.Context, ai LLMProvider, cfg *config.Config, page PagePlan) (string, error) {
	configJSON, _ := json.MarshalIndent(cfg, "", "  ")

	prompt := fmt.Sprintf(`You are an expert Next.js/React Developer.
Write the exact TypeScript React code for the following Next.js App Router Page.
Do NOT output markdown formatting like "` + "```tsx" + `". ONLY output pure JSON containing the exact file content in the "code" field.

Configuration Rules:
%s

Page Plan:
Name: %s
Required Components: %v

Output JSON format:
{
  "code": "import React from 'react';\n..."
}`, string(configJSON), page.Name, page.Components)

	rawJSON, err := ai.GenerateJSON(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("ai coder failed: %w", err)
	}

	var codeResp CoderResponse
	if err := json.Unmarshal([]byte(rawJSON), &codeResp); err != nil {
		return "", fmt.Errorf("failed to parse AI JSON code response: %w\nRaw: %s", err, rawJSON)
	}

	return codeResp.Code, nil
}
