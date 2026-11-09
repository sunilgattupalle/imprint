---
name: imprint
description: Build, validate, and evaluate a person's portable cognitive, dialectical, writing, visual, and delegation imprint from their local coding-agent sessions (via AgentsView) and authored manuscripts.
---

# Imprint Builder Skill (`SKILL.md`)

Use this skill when the user wants to build or refresh their personal **Imprint** (`<instance>/IMPRINT.md`, `imprint.json`, and `overlays/`) using the `imprint` Go CLI.

## Determinism Boundary

- **The `imprint` Go CLI** handles all deterministic data-plane work: `agentsview` configuration, cross-store session UUID deduplication, stripping `<worker>`/`<cosmos-events>` noise, `.docx`/`.pptx`/`.html`/`.md` text & CSS token extraction, token-budgeted digest compilation, Gate 1 & Gate 2 schema enforcement, and 20% holdout evaluation.
- **You (the Agent)** handle source curation with the user, fetching remote authored pages via MCP (Confluence/Notion/GitHub), synthesizing the portable claims vs. workspace bindings, and compiling `IMPRINT.md` + `overlays/`.

---

## Step-by-Step Workflow

### Step 1: Initialize the Two-Zone Instance
Run `imprint init`:
```bash
imprint init <instance_dir> --owner-id "<id>" --owner-name "<Full Name>" --sync
```
This creates `<instance_dir>/.local/` (gitignored so raw session data and verbatim quotes never enter git), configures `~/.agentsview/config.toml` (including `~/.augment/sessions-v2` if present), and auto-discovers `CLAUDE.md` / `style.md` priors.

### Step 2: Register User-Authored Ground Truth (`S2`) & Disallowed Workspace Terms
Ask the user once:
1. *"Which 2–5 files, slide decks (`.pptx`/`.html`), Word docs (`.docx`), early `v1` drafts, or Confluence/Notion URLs best represent your own writing and thinking?"*
2. *"What company, product, or internal repo names should be stripped from your portable imprint and kept in `.local/WORKSPACE.md`?"*

- For **remote URLs** (e.g. Confluence pages), fetch their content using your MCP tools and save them as markdown into `<instance_dir>/.local/authored/`.
- For **local files/globs and disallowed workspace terms**, register them via `imprint add`:
```bash
imprint add <instance_dir> \
  --authored "~/path/to/authored-drafts/*.md" \
  --authored "~/path/to/deck.pptx" \
  --chats "~/path/to/chatgpt-export-*.json" \
  --disallow-term "CompanyName" \
  --disallow-term "InternalDSL"
```

### Step 3: Build the Corpus & Synthesis Digest
Run `imprint digest`:
```bash
imprint digest <instance_dir>
```
This deduplicates sessions, strips non-human `role=user` turns, splits sessions 80/20 into `train_turns.jsonl` and `holdout_turns.jsonl`, and writes a single token-budgeted **`<instance_dir>/.local/corpus/digest.md`** combining:
- `S1`: High-signal training turns clustered by facet (`F1`–`F5`)
- `S2`: Extracted text and CSS tokens from `.md`, `.txt`, `.docx`, `.pptx`, and `.html` authored artifacts
- `S3`: Explicit style and operating priors

### Step 4: Synthesize L1 Portable Claims vs. L2/L3 Local Ledger
Read `<instance_dir>/.local/corpus/digest.md` and write:
1. `<instance_dir>/.local/evidence_ledger.json` — Graduated **L1 Portable Imprint** claims (`IMP-F1-01` .. `IMP-F5-xx`) across the 5 facets:
   - `F1_epistemic_posture`
   - `F2_dialectical_style`
   - `F3_compression_and_structure`
   - `F4_voice_and_visuals`
   - `F5_delegation_and_risk`
   Every claim must pass:
   - **Gate 1 (`portability_check`):** zero company/workspace proper nouns in `title`, `portable_rule`, `applies_when`, or `flips_when` (put company bindings in `workspace_bindings`).
   - **Gate 2 (`peer_fork`):** state what a strong peer engineer would legitimately prefer instead.
   - **Counter-evidence (`flips_when`):** state the exact boundary condition where the user inverts the rule.
2. `<instance_dir>/.local/demoted_non_imprint.json` & `<instance_dir>/.local/WORKSPACE.md` — Separated **L2 Workspace Bindings** and **L3 Agent Baseline** rules.
3. `<instance_dir>/IMPRINT.md` (~500–650 tokens) and `<instance_dir>/overlays/mode-*.md`.

### Step 5: Validate Gates & Run Held-Out Evaluation
```bash
imprint validate <instance_dir>
imprint eval <instance_dir>
```
- `imprint validate` verifies all gates and exports the sanitized, quote-free `<instance_dir>/imprint.json`.
- `imprint eval` tests the claims against the untouched 20% holdout sessions and writes `<instance_dir>/.local/eval/holdout_report.md`.
