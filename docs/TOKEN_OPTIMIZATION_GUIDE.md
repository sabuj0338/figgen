# Figgen Token Optimization & Development Guide

> Status: **Phases 0–6 implemented.** This document is both the rationale and the
> operating manual for Figgen's token-efficiency features across the
> Figma → Code pipeline (Planning → Coding).

---

## 1. The Token Pipeline (where tokens are spent)

```text
Figma (MCP / Plugin)
        │  raw design tree (YAML/JSON)
        ▼
[PRUNE: PruneFigmaData]  ── internal/figma/prune.go   (+ --max-children cap)
        │  minified, style-stripped JSON
        ▼
[PLANNER]  ── internal/agents/planner.go     ← batched: many frames per call
        │  STATIC prefix (rules+config+schema) + DYNAMIC (pruned chunks)
        ▼
tasks.json / figma_context.json  ── internal/state, cmd/plan.go
        │
        ▼
[PRUNE: PruneForCoder]  ── internal/figma/prune.go
        │  styles kept, colors→hex, style→tailwind
        ▼
[CODER]  ── internal/agents/coder.go         ← 1 call per non-shadcn task
        │  STATIC prefix (cached) + DYNAMIC (node context + plan)
        ▼
Generated .tsx files     (shadcn atoms are installed, not generated)
```

Every prompt is now split into a **STATIC prefix** (identical across calls →
cacheable) and a small **DYNAMIC body** (per-task). All calls are metered to
`.figgen/usage.json`.

---

## 2. Token Cost: before vs after

Approximate input tokens injected per call.

| Asset | Before | After |
|-------|-------:|------:|
| Coder rules per call | ~8,185 (`AI_AGENT_RULES` + 28 KB setup guide) | **~400** (`coder_rules_condensed.md`) |
| Coder static instructions | ~800 (duplicated in 2 prompts) | one shared block in the cacheable prefix |
| Planner calls per design | **N** (one per top-level frame) | **⌈N·chunkSize / budget⌉** (batched) |
| JSON-format prose | present in every prompt | removed (native JSON mode) |
| Static prefix re-billing | full price every call | **cached** (provider prompt/context cache) |

> The dominant win: coder rules dropped from ~8.2k to ~0.4k tokens/call, and the
> remaining static prefix is cached after the first call.

---

## 3. Implemented Features (and how to use them)

### Phase 0 — Cleanup ✅
- Raw model dumps (`debug_fulltext.txt`, `debug_planner.json`) are gated behind
  `--debug` (`internal/agents/provider.go` `Debug`, wired in `cmd/root.go`).
- The coder's `CRITICAL …` rules are a single shared block —
  `coderSharedRules` in `internal/agents/coder.go` — used by both the component
  and page prompts (no more duplication/drift).

### Telemetry ✅ — measure before/after
- `internal/telemetry` appends per-call usage to `<out>/.figgen/usage.json`
  (stage, label, provider, model, input/output/cached tokens).
- View it: `figgen usage` (totals, per-stage, and cache-hit rate).

### Phase 1 — Slim coder rules ✅
- `rules/coder_rules_condensed.md` (~400 tokens) is the only coder ruleset in
  `figgen.config.yaml`. The 28 KB `nextjs-project-setup-guide-light.md` is kept
  for humans and **not** sent to the model.

### Phase 2 — Provider JSON mode ✅
- Gemini: `ResponseMIMEType=application/json`.
- OpenAI: `response_format: {type: json_object}`.
- Anthropic: assistant **prefill** with `{` to force a JSON object.
- Ollama: `format: "json"`.
- The verbose "escape all quotes / output pure JSON" prose was removed.

### Phase 3 — Prompt / context caching ✅
- Prompts are structured as `GenerateRequest{StaticPrefix, Dynamic}`
  (`internal/agents/provider.go`).
- Anthropic: `cache_control: ephemeral` on the system prefix block.
- OpenAI: automatic server-side prefix caching (prefix sent first); `cached_tokens` recorded.
- Gemini: static prefix placed in `SystemInstruction` to benefit from implicit
  caching; `CachedContentTokenCount` recorded.
- Ollama: local, keeps context warm; no remote cache needed.

### Phase 4 — Batched planner calls ✅
- `cmd/plan.go` accumulates pruned frames and calls the planner once per
  `--batch-tokens` budget (default ~3,000 tokens) instead of once per frame.

### Phase 5 — Context shaping ✅
- `--max-children N` caps children per Figma node during pruning, recording
  `_omitted_children` (`internal/figma/prune.go`).
- Deterministic shadcn mapping: `agents.ApplyShadcnHeuristics` /
  `IsKnownShadcn` force-mark standard atoms (button, input, dialog, …) as
  `is_shadcn` so the runner **installs** them instead of spending a codegen call.

### Phase 6 — Model routing & output caps ✅
- Per-stage models: `planner_model` / `coder_model` in `figgen.config.yaml`
  (precedence: `--model` flag → config → `DEFAULT_MODEL`).
- Output caps: `PlannerMaxOutputTokens=4096`, `CoderMaxOutputTokens=8192`
  (`internal/agents/planner.go`), passed through every provider.

---

## 4. Operating Manual (knobs)

```bash
# Plan with batching tuned to ~4k-token planner calls
./figgen plan --figma "<url>" --batch-tokens 4000

# Cap huge repeated lists (e.g. long tables) to 20 children per node
./figgen run --all --max-children 20

# Route planning to a cheap model, coding to a stronger one (figgen.config.yaml)
#   planner_model: "gemini-2.5-flash"
#   coder_model:   "gemini-2.5-pro"

# Inspect token spend after a run
./figgen usage

# Troubleshoot raw model output
./figgen run --debug
```

| Knob | Where | Effect |
|------|-------|--------|
| `--batch-tokens` | `plan` | Frames per planner call (↑ = fewer calls) |
| `--max-children` | global | Bounds tree size in prompts |
| `--debug` | global | Writes raw model responses to disk |
| `planner_model`/`coder_model` | config | Per-stage model routing |
| `--model` / `--provider` | plan/run/listen | Override per command |

---

## 5. Metrics to track (`figgen usage`)

| Metric | Goal after optimization |
|--------|-------------------------|
| tokens/coder-call (input) | −60–70% vs pre-Phase-1 |
| retry/parse-failure rate | ~0% (JSON mode) |
| cache hit rate (of input) | >80% on the static prefix once warm |
| planner calls/run | reduced via `--batch-tokens` |
| total codegen calls/run | reduced via shadcn install-skip |

Run a reference design, note `figgen usage`, change one knob, re-run, compare.

---

## 6. Pros / Cons of the current design

| Pros | Cons |
|------|------|
| Static/dynamic split makes caching automatic across providers | Cache benefits are provider-dependent (Gemini implicit only) |
| Condensed rules cut the biggest cost | Under-specified rules can reduce output quality — keep `coder_rules_condensed.md` curated |
| JSON mode removes retries and prose | Anthropic relies on prefill, not a strict schema |
| Batching cuts fixed per-call overhead | Larger planner prompts risk truncation — tune `--batch-tokens` |
| Deterministic shadcn skip removes whole calls | Mapping table (`knownShadcnComponents`) needs upkeep |
| Telemetry enables data-driven tuning | Usage is estimated from provider-reported counts |

---

## 7. Remaining / Optional Future Work

These were intentionally **not** implemented (higher cost, lower ROI, or they
*increase* tokens) and are safe to defer:

- [ ] **Gemini explicit `CachedContent`** (vs current implicit caching) — only
      worth it for very large, stable prefixes; has minimum-size thresholds.
- [ ] **Anthropic strict JSON schema** via tool-use (current prefill is reliable
      enough for these prompts).
- [ ] **Symbol/style dictionary** — extract repeated colors/typography into a
      legend referenced by token. Marginal now that colors are pre-hexed and
      styles pre-mapped to Tailwind.
- [ ] **Incremental regeneration** — only re-plan/re-code changed Figma nodes.
- [ ] **Dry-run cost estimate** (`plan --estimate`) before spending.
- [ ] **Visual QA / Reviewer agents** — *increase* tokens; add only after the
      above, and keep their prompts diff-scoped.

---

## 8. Golden Rule

> **Send the static stuff once, send the dynamic stuff small.**
> Rules/config/instructions are static → curate (Phase 1), put them in the
> cacheable prefix (Phase 3), and let JSON mode (Phase 2) handle formatting.
> Figma node context is dynamic → prune hard, cap depth (Phase 5), batch at the
> planner (Phase 4), and never send a node to more than one call.
