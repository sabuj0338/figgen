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

// coderSharedRules holds the generation-time constraints that are identical for
// every component and page. Kept in one place so the rules never drift between
// the component and page prompts, and so the block lives in the cacheable
// static prefix.
const coderSharedRules = `You are an expert Next.js/React Developer. Write exact TypeScript React code for the given plan.
CRITICAL LINTING & CLEAN CODE: No unused imports/variables/functions. Never use the 'any' type; define proper TypeScript interfaces. If you use a raw <img> instead of next/image, put {/* eslint-disable-next-line @next/next/no-img-element */} on the line immediately before each <img>.
CRITICAL NAVIGATION: Use "import Link from 'next/link'" and "import { useRouter } from 'next/navigation'". Do NOT use @/i18n/navigation.
CRITICAL I18N: Initialize next-intl exactly as const t = useTranslations("<namespace>") using the namespace given below, and access keys directly from it. Extract ALL visible text from the Figma context into the translations map. Do not use dummy data and do not nest under another namespace key.
CRITICAL IMAGES: Use the downloaded images/vectors listed below via standard <img src="/images/<filename>" />. Do NOT emit raw inline <svg>...</svg>. If a needed vector/icon/illustration is missing from the list, render <img src="https://placehold.co/100x100" alt="Placeholder" /> so its layout space is preserved. Never skip the element.
CRITICAL LAYOUT: Map Figma layout data to Tailwind: layoutMode HORIZONTAL=flex-row / VERTICAL=flex-col, primaryAxisAlignItems=justify-*, counterAxisAlignItems=items-*, itemSpacing=gap-*, padding=p-*.
CRITICAL TYPOGRAPHY: Map fontSize/fontWeight/lineHeight/letterSpacing (or the precomputed "tailwind_text" field) to Tailwind text classes (text-[16px], font-semibold, leading-[24px], tracking-wide).
CRITICAL SHAPES: Map cornerRadius/fills/strokes to rounded-*, bg-[#hex], border border-[#hex]. Colors are provided as hex.
CRITICAL STRUCTURE: Never render an entire page/layout as a single <img>. Build the UI with real HTML elements and Tailwind CSS.`

const coderOutputSchema = `Output a JSON object with this exact shape:
{"code": "import React from 'react';\n...", "dependencies": ["lucide-react"], "shadcn_components": ["button"], "translations": {"your_descriptive_key": "Actual text from Figma"}}`

// coderStaticPrefix is the cacheable instruction/rules/config/schema block,
// identical across all coder calls in a run.
func coderStaticPrefix(cfg *config.Config) string {
	configJSON, _ := json.Marshal(cfg)
	return fmt.Sprintf("%s\n\nConfiguration Rules:\n%s\n\nEngineering Guidelines:\n%s\n\n%s",
		coderSharedRules, string(configJSON), cfg.CoderRulesContent, coderOutputSchema)
}

func RunCoderForComponent(ctx context.Context, ai LLMProvider, cfg *config.Config, comp ComponentPlan, availableImages []string, mcpContext string) (*CoderResponse, error) {
	ns := strings.ToLower(comp.Name)

	var modeInstruction string
	if comp.IsShadcn {
		modeInstruction = "SHADCN MODE: This maps to a Shadcn UI component. Import and use the standard Shadcn component natively (e.g. <Button variant=\"outline\">). Do not reconstruct it from raw divs."
	} else {
		modeInstruction = "CUSTOM MODE: This is a custom component. Design it freely with raw HTML structure and Tailwind to match the Figma context. Do not force it into a standard Shadcn block."
	}

	dynamic := fmt.Sprintf(`%s
Translations namespace: %q
Available images/vectors: %v

Component Plan:
Name: %s
Description: %s
Props: %v
Is Shadcn: %v

Figma Design Data (Semantic Context):
%s`, modeInstruction, ns, availableImages, comp.Name, comp.Description, comp.Props, comp.IsShadcn, mcpContext)

	return runCoder(ctx, ai, cfg, comp.Name, dynamic)
}

func RunCoderForPage(ctx context.Context, ai LLMProvider, cfg *config.Config, page PagePlan, requiredComponentPlans map[string]ComponentPlan, availableImages []string, mcpContext string) (*CoderResponse, error) {
	ns := strings.ToLower(page.Name)
	reqCompsJSON, _ := json.Marshal(requiredComponentPlans)

	dynamic := fmt.Sprintf(`Build a Next.js App Router page.
Translations namespace: %q
Available images/vectors: %v

This page must import and render the following previously generated components from "@/components/common/<Name>" or "@/components/ui/<Name>", passing ALL required props. Do NOT rewrite their internals:
%s

Page Plan:
Name: %s
Required Components: %v

Figma Design Data (Semantic Context):
%s`, ns, availableImages, string(reqCompsJSON), page.Name, page.Components, mcpContext)

	return runCoder(ctx, ai, cfg, page.Name, dynamic)
}

func runCoder(ctx context.Context, ai LLMProvider, cfg *config.Config, label string, dynamic string) (*CoderResponse, error) {
	result, err := ai.GenerateJSON(ctx, GenerateRequest{
		StaticPrefix:    coderStaticPrefix(cfg),
		Dynamic:         dynamic,
		Stage:           StageCode,
		Label:           label,
		MaxOutputTokens: CoderMaxOutputTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("ai coder failed: %w", err)
	}

	// A truncated response is cut off mid-JSON, so parsing it would fail with a
	// misleading "unexpected end of JSON input". Detect it up front and return a
	// clear, actionable error instead.
	if result.Truncated {
		return nil, fmt.Errorf(
			"AI code response for %q was truncated at the output token limit (%d tokens, stop reason %q): the generated component is too large to fit. Use a model with a larger output window (e.g. gemini-2.5-pro, gpt-4o, claude-3-7-sonnet or claude-sonnet-4) or split the component into smaller pieces",
			label, CoderMaxOutputTokens, result.FinishReason)
	}

	var codeResp CoderResponse
	if err := json.Unmarshal([]byte(result.Text), &codeResp); err != nil {
		return nil, fmt.Errorf("failed to parse AI JSON code response: %w\nRaw: %s", err, result.Text)
	}

	return &codeResp, nil
}
