package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sabujislam/figgen/internal/config"
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
	Category    string   `json:"category"`
	FigmaNodeID string   `json:"figma_node_id"`
}

type PagePlan struct {
	Name        string   `json:"name"`
	Components  []string `json:"components"`
	Category    string   `json:"category"`
	FigmaNodeID string   `json:"figma_node_id"`
}

func RunPlanner(ctx context.Context, ai LLMProvider, cfg *config.Config, mcpContext string) (*PlannerResponse, error) {
	// Serialize inputs for prompt without indentation to save tokens
	configJSON, _ := json.Marshal(cfg)

	prompt := fmt.Sprintf(`You are an expert Frontend Next.js Architecture Planner.
Analyze the following Figma Design semantic context and the provided project configuration.
Create a structured component generation plan. Group components and pages into logical modules using the "category" field (e.g. "Authentication", "Dashboard", "Settings"). For shared/generic UI elements, use the category "Global".
CRITICAL: You MUST break down large layouts into granular, reusable components (e.g. HeroSection, FeatureList, Footer). Do NOT output a single massive component for the entire page.

Configuration Rules:
%s

Engineering Guidelines:
%s

Figma Semantic Design Context (from MCP):
%s

Output JSON in this exact structure:
{
  "components": [
    {
      "name": "Button",
      "description": "Primary action button",
      "props": ["onClick", "variant"],
      "is_shadcn": true,
      "category": "Global",
      "figma_node_id": "0:1"
    }
  ],
  "pages": [
    {
      "name": "Dashboard",
      "components": ["Button", "Sidebar"],
      "category": "Dashboard",
      "figma_node_id": "0:2"
    }
  ]
}`, string(configJSON), cfg.PlannerRulesContent, mcpContext)

	rawJSON, err := ai.GenerateJSON(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("ai generation failed: %w", err)
	}

	var plan PlannerResponse
	if err := json.Unmarshal([]byte(rawJSON), &plan); err != nil {
		os.WriteFile("debug_planner.json", []byte(rawJSON), 0644)
		return nil, fmt.Errorf("failed to parse AI JSON response: %w", err)
	}

	return &plan, nil
}
