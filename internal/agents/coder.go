package agents

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sabujislam/figgen/internal/config"
)

type CoderResponse struct {
	Code             string                 `json:"code"`
	Dependencies     []string               `json:"dependencies"`
	ShadcnComponents []string               `json:"shadcn_components"`
	Translations     map[string]interface{} `json:"translations"`
}

func RunCoderForComponent(ctx context.Context, ai LLMProvider, cfg *config.Config, comp ComponentPlan) (*CoderResponse, error) {
	configJSON, _ := json.Marshal(cfg)

	prompt := fmt.Sprintf(`You are an expert Next.js/React Developer.
Write the exact TypeScript React code for the following component plan.
Do NOT output markdown formatting like "` + "```tsx" + `". ONLY output pure JSON.
Identify any external NPM dependencies you use (like "date-fns" or "lucide-react") and any shadcn/ui components you import (like "button" or "calendar").
CRITICAL: For navigation, strictly use "import Link from 'next/link'" and "import { useRouter } from 'next/navigation'". Do NOT use @/i18n/navigation.
CRITICAL: If you use next-intl useTranslations(), provide the English text used for each key in the "translations" map.
CRITICAL: Use standard HTML <img src="/placeholder.svg" /> for all images as we do not have static image files.

Configuration Rules:
%s

Engineering Guidelines:
%s

Component Plan:
Name: %s
Description: %s
Props: %v
Is Shadcn: %v

Output JSON format:
{
  "code": "import React from 'react';\n...",
  "dependencies": ["lucide-react"],
  "shadcn_components": ["button"],
  "translations": { "title": "Welcome", "description": "Description text" }
}`, string(configJSON), cfg.CoderRulesContent, comp.Name, comp.Description, comp.Props, comp.IsShadcn)

	rawJSON, err := ai.GenerateJSON(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("ai coder failed: %w", err)
	}

	var codeResp CoderResponse
	if err := json.Unmarshal([]byte(rawJSON), &codeResp); err != nil {
		return nil, fmt.Errorf("failed to parse AI JSON code response: %w\nRaw: %s", err, rawJSON)
	}

	return &codeResp, nil
}

func RunCoderForPage(ctx context.Context, ai LLMProvider, cfg *config.Config, page PagePlan) (*CoderResponse, error) {
	configJSON, _ := json.Marshal(cfg)

	prompt := fmt.Sprintf(`You are an expert Next.js/React Developer.
Write the exact TypeScript React code for the following Next.js App Router Page.
Do NOT output markdown formatting like "` + "```tsx" + `". ONLY output pure JSON.
Identify any external NPM dependencies you use (like "date-fns" or "lucide-react") and any shadcn/ui components you import (like "button" or "calendar").
CRITICAL: For navigation, strictly use "import Link from 'next/link'" and "import { useRouter } from 'next/navigation'". Do NOT use @/i18n/navigation.
CRITICAL: If you use next-intl useTranslations(), provide the English text used for each key in the "translations" map.
CRITICAL: Use standard HTML <img src="/placeholder.svg" /> for all images as we do not have static image files.

Configuration Rules:
%s

Engineering Guidelines:
%s

Page Plan:
Name: %s
Required Components: %v

Output JSON format:
{
  "code": "import React from 'react';\n...",
  "dependencies": ["lucide-react"],
  "shadcn_components": ["button"],
  "translations": { "title": "Welcome", "description": "Description text" }
}`, string(configJSON), cfg.CoderRulesContent, page.Name, page.Components)

	rawJSON, err := ai.GenerateJSON(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("ai coder failed: %w", err)
	}

	var codeResp CoderResponse
	if err := json.Unmarshal([]byte(rawJSON), &codeResp); err != nil {
		return nil, fmt.Errorf("failed to parse AI JSON code response: %w\nRaw: %s", err, rawJSON)
	}

	return &codeResp, nil
}
