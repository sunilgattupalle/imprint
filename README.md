# Imprint Model (`imprint-model`)

> Blueprint and Go CLI to build a portable imprint of a person—carrying their judgment, taste, and voice into any AI tool as it is used.

Most "personalization" files mix three unrelated things into one prompt: **who the person is**, **company-specific repo conventions**, and **generic workarounds for LLM bugs**. The result cannot leave the repo it was written in and degrades whenever the task changes.

`imprint-model` separates those layers, mines evidence from local coding-agent archives ([AgentsView](https://github.com/wesm/agentsview)) plus authored work, and compiles a ~600-token `IMPRINT.md` and task overlays that travel across tools and workspaces.

---

## 1. The Construct

### Three Layers (Only L1 Is the Imprint)

| Layer | Definition | Destination |
|---|---|---|
| **L1. Personal Imprint** | Portable cognitive, dialectical, structural, stylistic, and delegation traits of **the person** across any workspace | `<instance>/IMPRINT.md`, `imprint.json`, `overlays/` (tracked) |
| **L2. Workspace Bindings** | How that imprint binds to a **specific company, stack, or repo** | `<instance>/.local/WORKSPACE.md` (gitignored) |
| **L3. Agent Baseline** | Universal competence **every** engineer expects (e.g., verify before claiming done) | `<instance>/.local/demoted_non_imprint.json` (excluded from imprint) |

### Two Admission Gates

A candidate rule enters **L1** only if it passes both gates and records its counter-evidence boundary:

1. **Portability Gate (`portability_check`):** Still holds if the person switches companies or works on an open-source repo tomorrow. Zero company/workspace proper nouns may appear in any exported field.
2. **Non-Trivial Fork Gate (`peer_fork`):** Another strong, reasonable engineer would legitimately prefer the opposite. If no reasonable peer would prefer the opposite, it is **L3 Agent Baseline**, not personal imprint.
3. **Flip Boundary (`flips_when`):** Every rule states the exact condition where the person intentionally inverts it (e.g., terse in chat and executive summaries vs. never compressing away reasoning in deep evaluations).

### Five Portable Facets

| Facet | Question it answers |
|---|---|
| **`F1_epistemic_posture`** | How does this person decide what is true, well-abstracted, and properly evaluated? |
| **`F2_dialectical_style`** | When should the agent challenge vs. execute, and how are disagreements resolved? |
| **`F3_compression_and_structure`** | Where does this person want extreme compression vs. full analytical depth, and how are documents structured? |
| **`F4_voice_and_visuals`** | What sentence cadence, vocabulary, visual hierarchy, and diagram medium match their hand? |
| **`F5_delegation_and_risk`** | Which actions are autonomous, which require human sign-off, and how are long-running tasks coordinated? |

---

## 2. Two-Zone Privacy Boundary

An imprint instance directory enforces a hard split between raw local evidence and portable output:

```text
<instance>/
├── .gitignore                 # Ignores .local/
├── IMPRINT.md                 # Always-on portable imprint (~500–650 tokens)
├── imprint.json               # Validated L1 claims (aggregate counts only, zero quotes/UUIDs)
├── overlays/                  # Loaded on demand by task genre
│   ├── mode-architecture-and-specs.md
│   ├── mode-exec-memos-and-decks.md
│   ├── mode-visual-html-and-prototypes.md
│   ├── mode-coding-and-prs.md
│   └── mode-autonomous-orchestration.md
└── .local/                    # Gitignored — never leaves the machine
    ├── sources.yaml           # Local DB paths, authored file globs, disallowed workspace terms
    ├── authored/              # Remote docs/decks fetched via MCP
    ├── corpus/                # Deduplicated train/holdout JSONL + synthesis digest.md
    ├── evidence_ledger.json   # Full claims with verbatim quotes & session UUIDs
    ├── demoted_non_imprint.json
    ├── WORKSPACE.md           # L2 company/repo bindings
    └── eval/                  # 20% holdout evaluation report
```

---

## 3. CLI & Agent Workflow

Install the CLI:

```bash
go install github.com/sunilgattupalle/imprint-model/cmd/imprint@latest
```

The Go CLI handles deterministic collection, hygiene, extraction, gate enforcement, and holdout scoring. Your coding agent (guided by [`SKILL.md`](./SKILL.md)) handles source curation and claim synthesis:

```bash
# 1. Scaffold two-zone instance, wire ~/.agentsview/config.toml, discover local priors
imprint init <instance_dir> --owner-id "<id>" --owner-name "<Name>" --sync

# 2. Register authored manuscripts (.md, .txt, .docx, .pptx, .html), chat exports, and banned workspace terms
imprint add <instance_dir> \
  --authored "~/docs/*.md" \
  --authored "~/decks/*.pptx" \
  --chats "~/exports/chatgpt-*.json" \
  --disallow-term "CompanyName"

# 3. Deduplicate sessions, strip non-human role=user noise, split 80/20, and build .local/corpus/digest.md
imprint digest <instance_dir>

# 4. Agent reads .local/corpus/digest.md (per SKILL.md) and writes:
#    .local/evidence_ledger.json, .local/WORKSPACE.md, IMPRINT.md, and overlays/

# 5. Enforce Gate 1 & Gate 2, export quote-free imprint.json, and score against the 20% holdout split
imprint validate <instance_dir>
imprint eval <instance_dir>
```

---

## 4. Repository Layout

- [`SKILL.md`](./SKILL.md) — Agent workflow for building or updating a personal imprint instance.
- [`AGENTS.md`](./AGENTS.md) — Operating instructions and invariants for agents working in this repository.
- [`cmd/imprint/`](./cmd/imprint) & [`internal/`](./internal) — Go CLI implementation (`config`, `collector`, `ingest`, `corpus`, `digest`, `gate`, `eval`).
- [`schemas/`](./schemas) — JSON Schemas for local evidence claims (`claim.schema.json`) and exported quote-free imprints (`imprint.schema.json`).
- [`docs/`](./docs) — Method specifications:
  - [`01-sources.md`](./docs/01-sources.md) — Source taxonomy (`S1` session archives, `S2` authored artifacts, `S3` explicit priors)
  - [`02-hygiene.md`](./docs/02-hygiene.md) — Cross-store UUID deduplication, `<worker>`/event purification, and 80/20 split
  - [`03-dimensions.md`](./docs/03-dimensions.md) — The five portable facets and two admission gates
  - [`04-extraction.md`](./docs/04-extraction.md) — Four-pass synthesis and graduation rules
  - [`05-evaluation.md`](./docs/05-evaluation.md) — Held-out intervention coverage and maintenance loop
