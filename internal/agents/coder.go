package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

	modeInstruction := ""
	if comp.IsShadcn {
		modeInstruction = "CRITICAL SHADCN MODE: This component maps directly to a Shadcn UI component. You MUST import the standard Shadcn UI React component (e.g., <Button variant=\"outline\">) and use it natively. Do not reconstruct standard Shadcn components from raw divs."
	} else {
		modeInstruction = "CRITICAL CUSTOM MODE (TOTAL FREEDOM): This is a custom or independent component. Do NOT force this into a standard Shadcn component block. You have total freedom to design this component perfectly matching the Figma context using raw HTML structure (e.g. <div>, <section>) and independent Tailwind styling."
	}

	prompt := fmt.Sprintf(`You are an expert Next.js/React Developer.
Write the exact TypeScript React code for the following component plan.
Do NOT output markdown formatting like "` + "```tsx" + `". ONLY output pure JSON.
CRITICAL JSON FORMATTING: You MUST output strictly valid JSON. Because the "code" field contains a large string of React code, you MUST properly escape ALL double quotes (\") and newlines (\n) inside the code string. Failure to escape double quotes will break the JSON parser.
CRITICAL LINTING & CLEAN CODE: Do NOT leave any unused imports, variables, or functions. If you do not use a variable or import, you MUST delete it. You MUST NOT use 'any' types anywhere; define proper TypeScript interfaces. If you use standard HTML <img> tags instead of next/image, you MUST place {/* eslint-disable-next-line @next/next/no-img-element */} on the line immediately preceding each <img> tag.
Identify any external NPM dependencies you use (like "date-fns" or "lucide-react") and any shadcn/ui components you import (like "button" or "calendar").
%s
CRITICAL: For navigation, strictly use "import Link from 'next/link'" and "import { useRouter } from 'next/navigation'". Do NOT use @/i18n/navigation.
CRITICAL NEXT-INTL NAMESPACE: Your translations will be injected into en.json under the explicit namespace "%s". You MUST initialize next-intl using exactly: const t = useTranslations("%s"); and access keys directly from it. Do NOT use any other namespace and do NOT nest your "translations" output block under another namespace key. You MUST extract all visible text from the Figma semantic context and place it in the translations map. Do not use dummy data.
CRITICAL: Available images/vectors downloaded for this task: %v. Use standard HTML <img src="/images/filename.svg" /> or <img src="/images/filename.png" /> for these. 
CRITICAL: Do NOT generate raw inline <svg>...</svg> tags under ANY circumstances. If a vector, icon, or illustration is missing from the available list, you MUST render a visible image placeholder like <img src="https://placehold.co/100x100" alt="Placeholder" /> so its physical space and layout are understandable. Do NOT attempt to draw vectors manually and do NOT simply skip the element.
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
  "translations": { "your_descriptive_key": "Actual text extracted from Figma context" }
} `, modeInstruction, strings.ToLower(comp.Name), strings.ToLower(comp.Name), availableImages, string(configJSON), cfg.CoderRulesContent, comp.Name, comp.Description, comp.Props, comp.IsShadcn, mcpContext)

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

func RunCoderForPage(ctx context.Context, ai LLMProvider, cfg *config.Config, page PagePlan, requiredComponentPlans map[string]ComponentPlan, availableImages []string, mcpContext string) (*CoderResponse, error) {
	configJSON, _ := json.Marshal(cfg)
	reqCompsJSON, _ := json.MarshalIndent(requiredComponentPlans, "", "  ")

	prompt := fmt.Sprintf(`You are an expert Next.js/React Developer.
Write the exact TypeScript React code for the following Next.js App Router Page.
Do NOT output markdown formatting like "` + "```tsx" + `". ONLY output pure JSON.
CRITICAL JSON FORMATTING: You MUST output strictly valid JSON. Because the "code" field contains a large string of React code, you MUST properly escape ALL double quotes (\") and newlines (\n) inside the code string. Failure to escape double quotes will break the JSON parser.
CRITICAL LINTING & CLEAN CODE: Do NOT leave any unused imports, variables, or functions. If you do not use a variable or import, you MUST delete it. You MUST NOT use 'any' types anywhere; define proper TypeScript interfaces. If you use standard HTML <img> tags instead of next/image, you MUST place {/* eslint-disable-next-line @next/next/no-img-element */} on the line immediately preceding each <img> tag.
Identify any external NPM dependencies you use (like "date-fns" or "lucide-react") and any shadcn/ui components you import (like "button" or "calendar").
CRITICAL: For navigation, strictly use "import Link from 'next/link'" and "import { useRouter } from 'next/navigation'". Do NOT use @/i18n/navigation.
CRITICAL NEXT-INTL NAMESPACE: Your translations will be injected into en.json under the explicit namespace "%s". You MUST initialize next-intl using exactly: const t = useTranslations("%s"); and access keys directly from it. Do NOT use any other namespace and do NOT nest your "translations" output block under another namespace key. You MUST extract all visible text from the Figma semantic context and place it in the translations map. Do not use dummy data.
CRITICAL: Available images/vectors downloaded for this task: %v. Use standard HTML <img src="/images/filename.svg" /> or <img src="/images/filename.png" /> for these. 
CRITICAL: Do NOT generate raw inline <svg>...</svg> tags under ANY circumstances. If a vector, icon, or illustration is missing from the available list, you MUST render a visible image placeholder like <img src="https://placehold.co/100x100" alt="Placeholder" /> so its physical space and layout are understandable. Do NOT attempt to draw vectors manually and do NOT simply skip the element.
CRITICAL: Figma provides layout data like "layoutMode" (HORIZONTAL=flex-row, VERTICAL=flex-col), "primaryAxisAlignItems" (justify-content), "counterAxisAlignItems" (align-items), and exact padding/gap ("itemSpacing"). You MUST strictly map these to Tailwind flex utilities (e.g., flex, flex-col, justify-between, items-center, gap-X, p-X) to match the exact design alignment.
CRITICAL: Figma provides text styling under "style" (e.g., fontSize, fontWeight, fontFamily, letterSpacing, lineHeightPx). You MUST map these precisely to Tailwind text classes (e.g., text-[16px], font-semibold, leading-[24px], tracking-wide).
CRITICAL: For shapes, inputs, and buttons, map Figma "cornerRadius", "fills", and "strokes" to Tailwind border-radius, background, and border classes (e.g., rounded-md, bg-[#FF0000], border border-[#00FF00]).
CRITICAL: Figma provides colors in RGB format from 0 to 1 (e.g., {"r": 0.5, "g": 0.5, "b": 0.5}). You MUST convert these to HEX and use Tailwind arbitrary values (e.g., bg-[#808080], text-[#808080]).
CRITICAL: Do NOT render the entire page or layout as a single <img> tag. You MUST build the UI structure (headers, text, buttons, layouts) using standard HTML elements and Tailwind CSS.
CRITICAL COMPONENTS IMPORT: This page requires the following previously generated components. Their specifications and expected props are defined below:
%s

You MUST import them from "@/components/common/[ComponentName]" or "@/components/ui/[ComponentName]" and render them in your page layout, passing ALL required props. Do NOT rewrite their internal HTML from scratch.

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
  "translations": { "your_descriptive_key": "Actual text extracted from Figma context" }
} `, strings.ToLower(page.Name), strings.ToLower(page.Name), availableImages, string(reqCompsJSON), string(configJSON), cfg.CoderRulesContent, page.Name, page.Components, mcpContext)

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

