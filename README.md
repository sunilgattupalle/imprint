# Imprint (`imprint`)

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

`imprint` turns your real historical traces (coding-agent transcripts via [AgentsView](https://github.com/sunilgattupalle/agentsview) plus your own authored manuscripts and decks) into a verified, portable personal imprint with explicit boundary conditions (`flips_when`) for when each rule inverts.

---

## 2. What You Do With an Imprint

Generating an imprint produces three lightweight artifacts that travel across every AI tool you use:

| Artifact | Budget | Role |
|---|---|---|
| **`IMPRINT.md`** | ~500–650 tokens | **Always-on operating manual.** Loaded into every session so the agent starts with your epistemic bar, dialectical style, compression rules, prose cadence, and autonomy gates. |
| **`overlays/mode-*.md`** | ~250–400 tokens each | **Task-specific depth.** Attached on demand when the task matches (`mode-architecture-and-specs.md`, `mode-exec-memos-and-decks.md`, `mode-visual-html-and-prototypes.md`, `mode-coding-and-prs.md`, `mode-autonomous-orchestration.md`). |
| **`.local/WORKSPACE.md`** | Local only | **Company/repo bindings.** Loaded alongside `IMPRINT.md` only when working inside that workspace's repositories. |

### Automatic Wiring Across Your Tools (`imprint link`)

Running `imprint sync`, `imprint validate`, or `imprint link` automatically wires `~/.imprint/IMPRINT.md` and `~/.imprint/.local/WORKSPACE.md` into every detected coding harness on your machine via an idempotent managed block:

- **Claude Code CLI & Claude Desktop (Code / Cowork tabs):** Auto-linked in `~/.claude/CLAUDE.md`.
- **Pi (`pi`) & Oh My Pi (`omp`):** Auto-linked in `~/.pi/agent/AGENTS.md` and `~/.omp/agent/AGENTS.md`.
- **OpenCode (`opencode`):** Auto-linked in `~/.config/opencode/AGENTS.md`.
- **Hermes (`hermes`):** Auto-linked in `~/.hermes/SOUL.md`.
- **Harness (`auggie`) & Cursor:** Reference `~/.imprint/IMPRINT.md` in your global user rules or workspace `AGENTS.md` / `.cursor/rules`.
- **Claude Desktop (Chat / Projects) & ChatGPT:** Paste `~/.imprint/IMPRINT.md` into **Settings -> Personal Preferences / Custom Instructions**, and attach specific `overlays/mode-*.md` files to individual Projects.
- **Custom Agents & Memory Systems:** Read `~/.imprint/imprint.json` programmatically to select rules and condition episodic memory extraction by `facet`, `dimension`, `scope`, `applies_when`, `flips_when`, and `habit_reflection`.

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

### Seeded Facets + Emergent Dimensions & Habit Reflection

`imprint` combines **5 seeded portable facets** with **emergent cluster discovery** (`F6_emergent`) so recurring habits outside the seed list are discovered automatically:

| Facet | Question it answers |
|---|---|
| **`F1_epistemic_posture`** | How does this person decide what is true, well-abstracted, and properly evaluated? |
| **`F2_dialectical_style`** | When should the agent challenge vs. execute, and how are disagreements resolved? |
| **`F3_compression_and_structure`** | Where does this person want extreme compression vs. full analytical depth, and how are documents structured? |
| **`F4_voice_and_visuals`** | What sentence cadence (anchored on `--pre-ai` pure-human writing), vocabulary, visual hierarchy, and diagram medium match their hand? |
| **`F5_delegation_and_risk`** | Which actions are autonomous, which require human sign-off, and how are long-running tasks coordinated? |
| **`F6_emergent`** | What recurring situational corrections or preferences unique to this builder emerged outside `F1`–`F5`? |

Each claim also records **`habit_reflection`** (`observed_keep` vs. `aspirational_shift`), letting you inspect your observed habits and decide whether to preserve them or intentionally steer your agent toward a new habit.

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
go install github.com/sunilgattupalle/imprint/cmd/imprint@latest
```

All commands default to **`~/.imprint`** (no directory or owner arguments needed):

```bash
# 1. Sync local agent sessions, scaffold ~/.imprint, and build .local/corpus/digest.md
imprint sync

# 2. (Optional) Register pre-AI pure-human writings, authored docs/decks, and disallowed company terms
imprint add \
  --pre-ai "~/writings/pre-2023-*.md" \
  --authored "~/docs/*.md" \
  --authored "~/decks/*.pptx" \
  --disallow-term "CompanyName"

# 3. Agent reads ~/.imprint/.local/corpus/digest.md (per SKILL.md) and writes:
#    .local/evidence_ledger.json, .local/WORKSPACE.md, IMPRINT.md, and overlays/

# 4. Enforce Gate 1 & Gate 2, export quote-free imprint.json, auto-link into all harnesses, and score holdout
imprint validate
imprint eval
```

---

## 5. Repository Layout

- [`SKILL.md`](./SKILL.md): Step-by-step agent skill for generating or refreshing an imprint instance.
- [`AGENTS.md`](./AGENTS.md): Operating instructions and invariants for agents working in this repository.
- [`cmd/imprint/`](./cmd/imprint) & [`internal/`](./internal): Go CLI packages (`config`, `collector`, `ingest`, `corpus`, `digest`, `gate`, `eval`).
- [`schemas/`](./schemas): JSON Schemas for local evidence claims (`claim.schema.json`) and exported quote-free imprints (`imprint.schema.json`).
- [`docs/`](./docs): Detailed method specifications (`01-sources.md` through `05-evaluation.md`).
