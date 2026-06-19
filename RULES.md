# 🚀 Figgen Development Rules

This document outlines the strict technical guidelines and architectural rules for developing the `figgen` CLI project. **All AI agents must strictly adhere to these rules.**

## 1. Core Technology Stack
- **Language:** Golang (1.21+ recommended)
- **CLI Framework:** `github.com/spf13/cobra`
- **AI Model:** Multi-provider via a shared `LLMProvider` interface (`internal/agents`). Supported providers: Gemini (`github.com/google/generative-ai-go/genai`), OpenAI, Anthropic, and Ollama.
- **Configuration:** YAML (`gopkg.in/yaml.v3`)
- **Figma Integration:** Semantic design context is fetched via the Model Context Protocol using the `figma-developer-mcp` server (`internal/mcp`). The standard `net/http` Figma REST API client (`internal/figma`) is still used for downloading image and SVG assets.
- **Visual QA:** `github.com/playwright-community/playwright-go` (Future Phase)

## 2. Directory Structure & Architecture
Strictly follow standard Go project layout:
```text
/
├── cmd/           # Cobra CLI commands (plan, run, listen, status, retry, usage)
├── internal/      # Private application code
│   ├── config/    # YAML parsing logic
│   ├── figma/     # Figma REST client, types, and token-aware data pruning
│   ├── mcp/       # Model Context Protocol client (figma-developer-mcp)
│   ├── agents/    # AI interaction logic (multi-provider planner/coder)
│   ├── github/    # Git cloning and dependency bootstrap
│   ├── executor/  # Post-generation: deps, shadcn, prettier
│   ├── state/     # Stateful task tracker (tasks.json / tasks.md)
│   ├── telemetry/ # Per-call LLM token usage logging (.figgen/usage.json)
│   ├── logger/    # Console logging helpers
│   └── filesystem/# File writing and structuring
├── main.go        # Entry point
└── go.mod         # Module definition
```

## 3. Coding Guidelines
- **Idiomatic Go:** Always write clean, idiomatic Golang. Use `gofmt`.
- **Error Handling:** Do not ignore errors. Always return or log them contextually using `fmt.Errorf`.
- **Concurrency:** Use goroutines and channels safely when orchestrating multiple AI agent tasks. Ensure proper `sync.WaitGroup` usage.
- **Dependency Management:** Keep external dependencies minimal. Rely on the Go standard library as much as possible.

## 4. AI Agent Principles
- **No Hallucination:** Agents should only generate React/Next.js code based strictly on the provided config constraints.
- **Stateless Prompts:** Pass full context to the AI for each generation step.
- **Iterative Generation:** Generate components one by one safely, avoiding memory bloat.

## 5. Next.js Output Rules (Target Repository)
When writing to the generated repository:
- **Never edit shadcn/ui components.**
- **Use React Query** for all server state.
- **Use Zustand** only for UI state.
- **Strict Folder Structure:** Output code must perfectly align with standard Next.js App Router folders (`src/app`, `src/components/ui`, etc.).
