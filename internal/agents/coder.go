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

func RunCoderForComponent(ctx context.Context, ai LLMProvider, cfg *config.Config, comp ComponentPlan, availableImages []string, mcpContext string) (*CoderResponse, error) {
	configJSON, _ := json.Marshal(cfg)

	prompt := fmt.Sprintf(`You are an expert Next.js/React Developer.
Write the exact TypeScript React code for the following component plan.
Do NOT output markdown formatting like "` + "```tsx" + `". ONLY output pure JSON.
Identify any external NPM dependencies you use (like "date-fns" or "lucide-react") and any shadcn/ui components you import (like "button" or "calendar").
CRITICAL: For navigation, strictly use "import Link from 'next/link'" and "import { useRouter } from 'next/navigation'". Do NOT use @/i18n/navigation.
CRITICAL: If you use next-intl useTranslations(), provide the English text used for each key in the "translations" map.
CRITICAL: Available images/vectors downloaded for this task: %v. Use standard HTML <img src="/images/filename.svg" /> or <img src="/images/filename.png" /> for these. If you need an icon that is NOT in the available list, use 'lucide-react'. For unknown illustrations, use <img src="/placeholder.svg" />.
CRITICAL: Figma provides layout data like "layoutMode" (HORIZONTAL=flex-row, VERTICAL=flex-col), "primaryAxisAlignItems" (justify-content), "counterAxisAlignItems" (align-items), and exact padding/gap ("itemSpacing"). You MUST strictly map these to Tailwind flex utilities (e.g., flex, flex-col, justify-between, items-center, gap-X, p-X) to match the exact design alignment.
CRITICAL: Figma provides text styling under "style" (e.g., fontSize, fontWeight, fontFamily, letterSpacing, lineHeightPx). You MUST map these precisely to Tailwind text classes (e.g., text-[16px], font-semibold, leading-[24px], tracking-wide).
CRITICAL: For shapes, inputs, and buttons, map Figma "cornerRadius", "fills", and "strokes" to Tailwind border-radius, background, and border classes (e.g., rounded-md, bg-[#FF0000], border border-[#00FF00]).
CRITICAL: Figma provides colors in RGB format from 0 to 1 (e.g., {"r": 0.5, "g": 0.5, "b": 0.5}). You MUST convert these to HEX and use Tailwind arbitrary values (e.g., bg-[#808080], text-[#808080]).
CRITICAL: Do NOT render the entire page or layout as a single <img> tag. You MUST build the UI structure (headers, text, buttons, layouts) using standard HTML elements and Tailwind CSS.

Configuration Rules:
%s

Engineering Guidelines:
%s

Component Plan:
Name: %s
Description: %s
Props: %v
Is Shadcn: %v

Figma Design Data (Semantic Context):
%s

Output JSON format:
{
  "code": "import React from 'react';\n...",
  "dependencies": ["lucide-react"],
  "shadcn_components": ["button"],
  "translations": { "title": "Welcome", "description": "Description text" }
} `, availableImages, string(configJSON), cfg.CoderRulesContent, comp.Name, comp.Description, comp.Props, comp.IsShadcn, mcpContext)

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

func RunCoderForPage(ctx context.Context, ai LLMProvider, cfg *config.Config, page PagePlan, availableImages []string, mcpContext string) (*CoderResponse, error) {
	configJSON, _ := json.Marshal(cfg)

	prompt := fmt.Sprintf(`You are an expert Next.js/React Developer.
Write the exact TypeScript React code for the following Next.js App Router Page.
Do NOT output markdown formatting like "` + "```tsx" + `". ONLY output pure JSON.
Identify any external NPM dependencies you use (like "date-fns" or "lucide-react") and any shadcn/ui components you import (like "button" or "calendar").
CRITICAL: For navigation, strictly use "import Link from 'next/link'" and "import { useRouter } from 'next/navigation'". Do NOT use @/i18n/navigation.
CRITICAL: If you use next-intl useTranslations(), provide the English text used for each key in the "translations" map.
CRITICAL: Available images/vectors downloaded for this task: %v. Use standard HTML <img src="/images/filename.svg" /> or <img src="/images/filename.png" /> for these. If you need an icon that is NOT in the available list, use 'lucide-react'. For unknown illustrations, use <img src="/placeholder.svg" />.
CRITICAL: Figma provides layout data like "layoutMode" (HORIZONTAL=flex-row, VERTICAL=flex-col), "primaryAxisAlignItems" (justify-content), "counterAxisAlignItems" (align-items), and exact padding/gap ("itemSpacing"). You MUST strictly map these to Tailwind flex utilities (e.g., flex, flex-col, justify-between, items-center, gap-X, p-X) to match the exact design alignment.
CRITICAL: Figma provides text styling under "style" (e.g., fontSize, fontWeight, fontFamily, letterSpacing, lineHeightPx). You MUST map these precisely to Tailwind text classes (e.g., text-[16px], font-semibold, leading-[24px], tracking-wide).
CRITICAL: For shapes, inputs, and buttons, map Figma "cornerRadius", "fills", and "strokes" to Tailwind border-radius, background, and border classes (e.g., rounded-md, bg-[#FF0000], border border-[#00FF00]).
CRITICAL: Figma provides colors in RGB format from 0 to 1 (e.g., {"r": 0.5, "g": 0.5, "b": 0.5}). You MUST convert these to HEX and use Tailwind arbitrary values (e.g., bg-[#808080], text-[#808080]).
CRITICAL: Do NOT render the entire page or layout as a single <img> tag. You MUST build the UI structure (headers, text, buttons, layouts) using standard HTML elements and Tailwind CSS.

Configuration Rules:
%s

Engineering Guidelines:
%s

Page Plan:
Name: %s
Required Components: %v

Figma Design Data (Semantic Context):
%s

Output JSON format:
{
  "code": "import React from 'react';\n...",
  "dependencies": ["lucide-react"],
  "shadcn_components": ["button"],
  "translations": { "title": "Welcome", "description": "Description text" }
}`, availableImages, string(configJSON), cfg.CoderRulesContent, page.Name, page.Components, mcpContext)

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

