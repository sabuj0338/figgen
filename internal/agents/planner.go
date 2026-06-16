package agents

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sabujislam/figgen/internal/config"
	"github.com/sabujislam/figgen/internal/figma"
)

type PlannerResponse struct {
	Components []ComponentPlan `json:"components"`
	Pages      []PagePlan      `json:"pages"`
}

type ComponentPlan struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Props       []string `json:"props"`
	IsShadcn    bool     `json:"is_shadcn"`
}

type PagePlan struct {
	Name       string   `json:"name"`
	Components []string `json:"components"`
}

func RunPlanner(ctx context.Context, ai LLMProvider, cfg *config.Config, figmaNode *figma.FileResponse) (*PlannerResponse, error) {
	// Serialize inputs for prompt
	configJSON, _ := json.MarshalIndent(cfg, "", "  ")
	figmaJSON, _ := json.Marshal(figmaNode)

	prompt := fmt.Sprintf(`You are an expert Frontend Next.js Architecture Planner.
Analyze the following Figma Design tree and the provided project configuration.
Create a structured component generation plan.

Configuration Rules:
%s

Figma Design JSON:
%s

Output JSON in this exact structure:
{
  "components": [
    {
      "name": "Button",
      "description": "Primary action button",
      "props": ["onClick", "variant"],
      "is_shadcn": true
    }
  ],
  "pages": [
    {
      "name": "Dashboard",
      "components": ["Button", "Sidebar"]
    }
  ]
}`, string(configJSON), string(figmaJSON))

	rawJSON, err := ai.GenerateJSON(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("ai generation failed: %w", err)
	}

	var plan PlannerResponse
	if err := json.Unmarshal([]byte(rawJSON), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse AI JSON response: %w\nRaw: %s", err, rawJSON)
	}

	return &plan, nil
}
