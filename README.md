# Imprint Model (`imprint-model`)

> Blueprint and Go CLI to build a portable imprint of a person, carrying their judgment, taste, and voice into any AI tool as it is used.

---

## 1. Why This Matters

A builder does not experience AI tools as stateless API calls. They experience repeating the same corrections across every tool and session:
- *"Don't just agree with me, stress-test this."*
- *"You compressed out the actual technical mechanics."*
- *"Stop using AI buzzwords and em-dashes in every sentence."*
- *"Align on the mental model before writing 500 lines of code."*
- *"Never merge the PR yourself; I own the merge button."*

Hand-written `CLAUDE.md` or `AGENTS.md` files rarely solve this for long. First, people write down how they *think* they work rather than the boundaries they actually enforce across hundreds of sessions. Second, hand-written rules mix three different layers into one prompt: **who the person is**, **company-specific repo conventions**, and **generic workarounds for LLM bugs**. That file cannot leave the repo it was written in, and rules meant for terse chat replies break deep technical specs when applied blindly.

`imprint-model` turns your real historical traces (coding-agent transcripts via [AgentsView](https://github.com/wesm/agentsview) plus your own authored manuscripts and decks) into a verified, portable personal imprint with explicit boundary conditions (`flips_when`) for when each rule inverts.

---

## 2. What You Do With an Imprint

Generating an imprint produces three lightweight artifacts that travel across every AI tool you use:

| Artifact | Budget | Role |
|---|---|---|
| **`IMPRINT.md`** | ~500–650 tokens | **Always-on operating manual.** Loaded into every session so the agent starts with your epistemic bar, dialectical style, compression rules, prose cadence, and autonomy gates. |
| **`overlays/mode-*.md`** | ~250–400 tokens each | **Task-specific depth.** Attached on demand when the task matches (`mode-architecture-and-specs.md`, `mode-exec-memos-and-decks.md`, `mode-visual-html-and-prototypes.md`, `mode-coding-and-prs.md`, `mode-autonomous-orchestration.md`). |
| **`.local/WORKSPACE.md`** | Local only | **Company/repo bindings.** Loaded alongside `IMPRINT.md` only when working inside that workspace's repositories. |

### Wiring Across Your Tools

- **Claude Code CLI & Claude Desktop (Code / Cowork tabs):**
  Reference your imprint in `~/.claude/CLAUDE.md` so every local session inherits it automatically:
  ```markdown
  @~/path/to/your-imprint/IMPRINT.md
  @~/path/to/your-imprint/.local/WORKSPACE.md
  ```
  When starting a spec, memo, deck, or visual prototype, `@`-mention the matching file in `overlays/mode-*.md`.
- **Augment (`auggie`) & Cursor:**
  Add `IMPRINT.md` to your global user rules (or workspace `AGENTS.md` / `.cursor/rules`), and attach the matching `overlays/mode-*.md` when shifting into architecture, deck, or prototype work.
- **Claude Desktop (Chat / Projects) & ChatGPT:**
  Paste `IMPRINT.md` into **Settings -> Personal Preferences / Custom Instructions**, and attach specific `overlays/mode-*.md` files to individual Projects as project knowledge.
- **Custom Agents & Evaluation Harnesses:**
  Read `<instance>/imprint.json` programmatically to select rules by `facet`, `scope`, `applies_when`, and `flips_when`.

---

## 3. The Construct

### Three Layers (Only L1 Is the Imprint)

| Layer | Definition | Destination |
|---|---|---|
| **L1. Personal Imprint** | Portable cognitive, dialectical, structural, stylistic, and delegation traits of **the person** across any workspace | `<instance>/IMPRINT.md`, `imprint.json`, `overlays/` (tracked) |
| **L2. Workspace Bindings** | How that imprint binds to a **specific company, stack, or repo** | `<instance>/.local/WORKSPACE.md` (gitignored) |
| **L3. Agent Baseline** | Universal competence **every** engineer expects (e.g., verify before claiming done) | `<instance>/.local/demoted_non_imprint.json` (excluded from imprint) |

### Two Admission Gates and the Flip Boundary

A candidate rule graduates into **L1** only if it passes both gates and states its flip condition:

1. **Portability Gate (`portability_check`):** Still holds if the person switches companies or works on an open-source repo tomorrow. Zero company/workspace proper nouns may appear in any exported field.
2. **Non-Trivial Fork Gate (`peer_fork`):** Another strong, reasonable engineer would legitimately prefer the opposite. If no reasonable peer would prefer the opposite, it is **L3 Agent Baseline**, not personal imprint.
3. **Flip Boundary (`flips_when`):** Every rule states the exact condition where the person intentionally inverts it (for example: terse in chat and executive summaries vs. never compressing away reasoning in deep evaluations; SVG diagrams in published docs vs. ASCII in interactive chat).

### Five Portable Facets

| Facet | Question it answers |
|---|---|
| **`F1_epistemic_posture`** | How does this person decide what is true, well-abstracted, and properly evaluated? |
| **`F2_dialectical_style`** | When should the agent challenge vs. execute, and how are disagreements resolved? |
| **`F3_compression_and_structure`** | Where does this person want extreme compression vs. full analytical depth, and how are documents structured? |
| **`F4_voice_and_visuals`** | What sentence cadence, vocabulary, visual hierarchy, and diagram medium match their hand? |
| **`F5_delegation_and_risk`** | Which actions are autonomous, which require human sign-off, and how are long-running tasks coordinated? |

### Two-Zone Privacy Boundary

Raw session transcripts, verbatim quote ledgers, session UUIDs, and company bindings stay strictly inside `.local/` (gitignored). The tracked instance root contains only portable rules and aggregate provenance counts:

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
└── .local/                    # Gitignored: never leaves the machine
    ├── sources.yaml           # Local DB paths, authored file globs, disallowed workspace terms
    ├── authored/              # Remote docs/decks fetched via MCP
    ├── corpus/                # Deduplicated train/holdout JSONL + synthesis digest.md
    ├── evidence_ledger.json   # Full claims with verbatim quotes & session UUIDs
    ├── demoted_non_imprint.json
    ├── WORKSPACE.md           # L2 company/repo bindings
    └── eval/                  # 20% holdout evaluation report
```

---

## 4. How an Imprint Is Generated and Maintained

Install the Go CLI:

```bash
go install github.com/sunilgattupalle/imprint-model/cmd/imprint@latest
```

The `imprint` CLI handles deterministic collection, deduplication, document parsing, gate validation, and holdout scoring. Your coding agent (guided by [`SKILL.md`](./SKILL.md)) handles source curation and claim synthesis:

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

## 5. Repository Layout

- [`SKILL.md`](./SKILL.md): Step-by-step agent skill for generating or refreshing an imprint instance.
- [`AGENTS.md`](./AGENTS.md): Operating instructions and invariants for agents working in this repository.
- [`cmd/imprint/`](./cmd/imprint) & [`internal/`](./internal): Go CLI packages (`config`, `collector`, `ingest`, `corpus`, `digest`, `gate`, `eval`).
- [`schemas/`](./schemas): JSON Schemas for local evidence claims (`claim.schema.json`) and exported quote-free imprints (`imprint.schema.json`).
- [`docs/`](./docs): Detailed method specifications (`01-sources.md` through `05-evaluation.md`).
