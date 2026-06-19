# Figma → Code AI Agent Development Tracker

## Phase 1: Foundation (MVP)
- [x] Initialize Go module (`go mod init github.com/sabujislam/figgen`)
- [x] Setup Cobra CLI framework (`cmd/root.go`, `cmd/plan.go`, `cmd/run.go`)
- [x] Implement Configuration Parser (`internal/config/parser.go`)
- [x] Implement GitHub Cloner (`internal/github/cloner.go`)
- [x] Implement Figma API Extractor (`internal/figma/client.go`)
- [x] Setup Gemini AI Client wrapper (`internal/agents/gemini.go`)
- [x] Implement AI Planner Agent (`internal/agents/planner.go`)
- [x] Implement AI Coder Agent (`internal/agents/coder.go`)
- [x] Implement File System Writer (`internal/filesystem/writer.go`)

## Phase 1.5: Stateful Task Engine (Agentic Architecture)
- [x] Create `internal/state/manager.go`
- [x] Create `cmd/plan.go` (extract and plan tasks)
- [x] Create `cmd/run.go` (execute tasks iteratively)
- [x] Create `cmd/status.go` (inspect task progress)
- [x] Create `cmd/retry.go` (reset failed tasks to pending)

## Phase 1.6: Token Optimization (see docs/TOKEN_OPTIMIZATION_GUIDE.md)
- [x] Telemetry: per-call token usage logging + `figgen usage`
- [x] Condense coder rules (drop 28 KB guide from prompts)
- [x] Provider JSON mode (gemini/openai/anthropic/ollama)
- [x] Static-prefix prompt/context caching
- [x] Batched planner calls (`--batch-tokens`)
- [x] Pruning child cap (`--max-children`) + deterministic shadcn mapping
- [x] Per-stage model routing + output token caps

## Phase 2: Smart Mapping Engine
- [x] Map Figma elements to shadcn/ui equivalents (deterministic heuristic)
- [ ] Advanced architecture planner prompting
- [ ] Multi-component generation logic

## Phase 3: Production CLI Tool
- [ ] Auto dependency installation routines
- [ ] Config-driven comprehensive generation

## Phase 4: Visual QA System
- [ ] Integrate `playwright-go`
- [ ] Screenshot capture logic
- [ ] Auto-fix feedback loop

## Phase 5: Advanced Multi-Agent System
- [ ] Refine concurrent agent states (Reviewer, Refactor agents)

## Phase 6: SaaS Platform
- [ ] Wrap CLI logic into web server
- [ ] Web dashboard interface
