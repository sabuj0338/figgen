Below is your **final consolidated SRS + Implementation Plan (BRD-style)** for your project. You can directly give this to Codex or use it as your system blueprint.

---

# 🚀 Figma → Code AI Agent System (Sabuj Engineering Workflow)

## SRS + Implementation Plan + Architecture Guide

## 1. Project Overview

This project is an **AI-powered CLI + Agent system** that converts **Figma designs into production-ready Next.js code** using:

- Figma MCP (design data extraction)
- Pre-built Next.js boilerplate
- Standardized architecture rules
- AI multi-step reasoning pipeline
- Component mapping system
- Visual verification loop

### 🎯 Goal

Automate frontend development so that:

> A developer only provides a Figma URL + config file → system generates full production-ready UI code aligned with existing architecture.

---

## 2. Core Idea

Instead of direct “Figma → Code”, the system uses:

```text
Figma → MCP → Design JSON
      → AI Architecture Planner
      → Component Mapper
      → Code Generator
      → Visual QA Loop
      → Final Refactor Output
```

---

## 3. Key Innovation

### 🧠 Not a prompt system — an ENGINEERING SYSTEM

It is NOT:

- Single prompt generator
- Screenshot-to-code tool

It IS:

- Multi-stage AI engineering pipeline
- Architecture-aware code generator
- Project-context-driven system

---

## 4. Required Inputs

### 4.1 Figma Input

- Figma URL
- MCP connection (figma-developer-mcp)

### 4.2 Project Boilerplate (GitHub)

Pre-built Next.js template:

- Next.js 16 App Router
- TypeScript
- Tailwind
- shadcn/ui (locked)
- Zustand (UI state only)
- React Query (server state)
- next-intl
- RBAC system
- Standard folder structure

---

### 4.3 Config File (Critical)

`sabuj.config.yaml`

```yaml
framework: nextjs16

dependencies:
  ui:
    - shadcn
    - tailwindcss
    - lucide-react

  state:
    - zustand
    - react-query

  forms:
    - react-hook-form
    - zod

rules:
  - never_edit_shadcn_ui
  - use_react_query_for_api
  - use_zustand_only_for_ui
  - follow_folder_structure
```

---

## 5. System Architecture

### CLI Tool

```bash
sabuj-ui generate --figma <url> --page dashboard --config sabuj.config.yaml
```

---

### Internal Pipeline

```text
1. Load Boilerplate (GitHub template)
2. Load config file
3. Load architecture rules
4. Fetch Figma via MCP
5. Parse design structure
6. Map to existing components
7. Generate implementation plan
8. Generate code (component-by-component)
9. Run visual comparison (Playwright)
10. Auto-fix differences
11. Final output project
```

---

## 6. System Modules

### 6.1 Project Analyzer

- Reads Next.js boilerplate
- Understands folder structure
- Loads coding standards

---

### 6.2 Figma MCP Analyzer

- Extracts:
  - layout hierarchy
  - spacing
  - components
  - typography
  - assets

---

### 6.3 Component Mapper

- Maps Figma elements to:
  - shadcn/ui
  - custom components

- Prevents duplicate component creation

---

### 6.4 Architecture Planner

Outputs:

- component tree
- state requirements
- API hooks needed
- page structure

---

### 6.5 Code Generator

- Generates one component at a time
- Uses React Query + Zustand rules
- Follows folder structure strictly

---

### 6.6 Visual QA Engine (Future Phase)

- Playwright screenshot capture
- Compare with Figma
- Generate fix instructions

---

## 7. Folder Structure (Fixed Standard)

```text
src/
├── app/
├── components/
│   ├── ui/
│   ├── common/
│   ├── layout/
│   └── forms/
├── store/
├── hooks/api/
├── services/
├── lib/
├── types/
├── auth/
├── providers/
├── constants/
├── proxy.ts
```

---

## 8. Development Phases

### ✅ Phase 1 — Foundation (MVP)

- Create Next.js boilerplate repo
- Lock architecture rules
- Create config system (`sabuj.config.yaml`)
- CLI tool setup (`sabuj-ui`)
- Figma MCP integration
- Basic code generation (single page)

---

### 🚀 Phase 2 — Smart Mapping Engine

- Component mapper (Figma → shadcn/ui)
- Architecture planner AI
- Multi-component generation
- Folder-aware file writing

---

### 🚀 Phase 3 — Production CLI Tool

- `sabuj-ui generate` command
- Project cloning system from GitHub
- Auto dependency installation
- Config-driven generation

---

### 🚀 Phase 4 — Visual QA System

- Playwright integration
- Screenshot comparison with Figma
- Auto fix loop
- Pixel alignment improvement

---

### 🚀 Phase 5 — Multi-Agent System

- Planner agent
- Code agent
- Reviewer agent
- Refactor agent

---

### 🚀 Phase 6 — SaaS Platform (Optional Future)

- Web UI dashboard
- Figma integration UI
- Team collaboration
- Project history tracking

---

## 9. Key Design Rules

- Never edit `shadcn/ui` components
- React Query = all server state
- Zustand = only UI state
- No manual API fetching
- Always follow config file
- One component at a time generation
- Architecture-first before code

---

## 10. Final Outcome

This system enables:

### Before:

Manual frontend development + design translation

### After:

```text
Figma URL + CLI command → Full production Next.js UI
```

---

## 11. Core Philosophy

> “AI should not guess architecture. It should follow engineering rules.”

This system ensures:

- deterministic structure
- reusable architecture
- consistent code output
- scalable frontend generation

---

If you want next step, I can help you design:

- CLI folder structure
- Agent system architecture (LangGraph / OpenAI Agents)
- Prompt engineering for each phase
- or GitHub boilerplate repo structure

Just tell me.
