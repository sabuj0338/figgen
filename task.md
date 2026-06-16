# Figma → Code AI Agent Development Tracker

## Phase 1: Foundation (MVP)
- [x] Initialize Go module (`go mod init github.com/sabujislam/figgen`)
- [x] Setup Cobra CLI framework (`cmd/root.go`, `cmd/generate.go`)
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

## Phase 2: Smart Mapping Engine
- [ ] Map Figma elements to shadcn/ui equivalents
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
