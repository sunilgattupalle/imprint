---
name: imprint
description: Build, refresh, validate, and auto-link a person's portable cognitive, dialectical, writing, visual, and delegation imprint (~/.imprint) from their local coding-agent sessions and authored writings.
---

# Imprint Skill (`SKILL.md`)

Use this skill when the user wants to build, refresh, or review their personal **Imprint** (`~/.imprint/IMPRINT.md`, `imprint.json`, and `overlays/`) using the `imprint` Go CLI.

## Determinism Boundary

- **The `imprint` Go CLI** handles all deterministic data-plane work: initializing `~/.imprint`, scanning local session stores (Claude, Cursor IDE/CLI, Pi, OMP, Auggie, OpenCode, Devin) alongside `~/.agentsview/sessions.db`, deduplicating cross-store session UUIDs, stripping `<worker>`/`<*-events>` noise, extracting text and CSS tokens from `.docx`/`.pptx`/`.html`/`.md`, compiling the diversity-sampled digest (`F1`–`F5` seeded facets + `FX` emergent corrections), enforcing Gate 1 & Gate 2, scoring the 20% holdout split, and linking agent harnesses (`imprint link`).
- **You (the Agent)** handle source curation with the user, fetching remote authored pages via MCP (Confluence/Notion/GitHub), synthesizing portable claims and emergent dimensions, running the **Habit Reflection** check with the user, and writing `IMPRINT.md` + `overlays/`.

---

## Step-by-Step Workflow (Default Home: `~/.imprint`)

### Step 1: Register Pre-AI Writings (`S2A`), Authored Artifacts (`S2B`), & Disallowed Workspace Terms
Ask the user once:
1. *"Do you have 1–3 pieces written before AI tools (or written purely by hand) that best represent your natural voice, plus any recent specs, decks (`.pptx`/`.html`), or Confluence/Notion URLs?"*
2. *"What company, product, or internal repo names should be stripped from your portable imprint and kept in `.local/WORKSPACE.md`?"*

- Ensure `~/.imprint` exists (`imprint init` with no arguments).
- For **remote URLs** (e.g. Confluence pages), fetch their content using your MCP tools and save them as markdown into `~/.imprint/.local/authored/`.
- For **local files/globs and disallowed workspace terms**, register them via `imprint add`:
```bash
imprint add \
  --pre-ai "~/writings/pre-2023-*.md" \
  --authored "~/docs/*.md" \
  --authored "~/decks/*.pptx" \
  --disallow-term "CompanyName" \
  --disallow-term "InternalDSL"
```

### Step 2: Sync Sessions & Build the Synthesis Digest
Run one command:
```bash
imprint sync
```
This syncs local coding-agent sessions, deduplicates by session UUID, strips non-human `role=user` turns, splits sessions 80/20 into `train_turns.jsonl` and `holdout_turns.jsonl`, and writes **`~/.imprint/.local/corpus/digest.md`** containing:
- `S1`: Diversity-balanced training turns with assistant trigger context across the 5 seeded facets (`F1`–`F5`) plus **`FX: Emergent / Unclustered Corrections & Directives`**
- `S2A` (`[PRE-AI GOLD]`) & `S2B`: Stratified head+tail text and CSS tokens from `.md`, `.txt`, `.docx`, `.pptx`, and `.html` artifacts
- `S3`: Auto-discovered explicit style and operating priors

### Step 3: Synthesize L1 Portable Claims, Emergent Dimensions & Situational Overlays
Read `~/.imprint/.local/corpus/digest.md` and write:
1. `~/.imprint/.local/evidence_ledger.json`: Graduated **L1 Portable Imprint** claims (`IMP-F1-01` .. `IMP-F5-xx`, plus `IMP-E-xx` for any recurring **`F6_emergent`** dimension discovered in `FX`):
   - `F1_epistemic_posture`
   - `F2_dialectical_style`
   - `F3_compression_and_structure`
   - `F4_voice_and_visuals`
   - `F5_delegation_and_risk`
   - `F6_emergent` (any recurring builder-specific dimension not covered by `F1`–`F5`)
   Every claim must include:
   - `dimension`: Specific seeded or emergent dimension ID (e.g. `epistemic.abstraction_depth`, `dialectic.plan_before_code`, `emergent.functional_purity`).
   - **Gate 1 (`portability_check`):** zero company/workspace proper nouns in `title`, `portable_rule`, `applies_when`, or `flips_when` (put company bindings in `workspace_bindings`).
   - **Gate 2 (`peer_fork`):** state what a strong peer engineer would legitimately prefer instead.
   - **Situational Activation (`applies_when` & `flips_when`):** state when this dimension activates and the exact boundary condition where the user intentionally inverts it (for example, when they say *"go fast"* vs. *"go slow and align first"*).
2. `~/.imprint/.local/demoted_non_imprint.json` & `~/.imprint/.local/WORKSPACE.md`: Separated **L2 Workspace Bindings** and **L3 Agent Baseline** rules.
3. `~/.imprint/IMPRINT.md` (~500–650 tokens, including a situational activation router for `overlays/mode-*.md`) and `~/.imprint/overlays/mode-*.md`.

### Step 4: Habit Reflection Check (Observed Self vs. Aspirational Shift)
Habits in session logs are often automatic rather than deliberate. Run `imprint review`, show the user the discovered dimensions, and ask:
- *"Are there any observed habits here you want to keep as-is (`observed_keep`) vs. intentionally adjust or invert (`aspirational_shift`)?"*
Record the user's choices via `imprint review --set <CLAIM_ID>=observed_keep|aspirational_shift`.

### Step 5: Validate, Auto-Link Across All Harnesses & Evaluate Holdout
```bash
imprint validate
imprint eval
imprint brief --cwd . --prompt "verify situational brief" --json
```
- `imprint validate` enforces Gate 1 & Gate 2, exports quote-free `~/.imprint/imprint.json`, and **automatically links** `~/.imprint/IMPRINT.md` and `~/.imprint/.local/WORKSPACE.md` into Claude Code (`~/.claude/CLAUDE.md`), Pi (`~/.pi/agent/AGENTS.md`), Oh My Pi (`~/.omp/agent/AGENTS.md`), OpenCode (`~/.config/opencode/AGENTS.md`), and Hermes (`~/.hermes/SOUL.md`).
- `imprint eval` tests the claims against the untouched 20% holdout sessions and writes `~/.imprint/.local/eval/holdout_report.md`.
