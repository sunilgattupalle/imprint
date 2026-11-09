# Imprint Model (`imprint-model`)

**Imprint** is a local-first model, Go CLI (`imprint`), and Agent Skill (`SKILL.md`) for distilling how a specific builder thinks, argues, writes, structures information, and delegates risk from their coding-agent sessions (via [AgentsView](https://github.com/wesm/agentsview)) and authored manuscripts.

This repository contains **only the person-agnostic model, schemas, Go CLI, and `SKILL.md`**. It contains zero personal or company data. A person's instance lives in a separate directory with a hard two-zone privacy boundary (`.local/` for raw data and verbatim quotes; root for the sanitized, portable imprint).

---

## 1. Install the `imprint` Go CLI

```bash
go install github.com/sunilgattupalle/imprint-model/cmd/imprint@latest
# or from a local clone:
go build -o ~/.local/bin/imprint ./cmd/imprint
```

---

## 2. How the CLI (`imprint`) and Agent Skill (`SKILL.md`) Work Together

Building an imprint splits cleanly across a **deterministic data plane** (`imprint` Go CLI) and a **synthesis plane** (`SKILL.md` executed by the user's own coding agent):

| Step | Command / Actor | What it does |
|---|---|---|
| **1. Initialize** | `imprint init <dir> [--sync]` | Creates `<dir>/.local/` (gitignored), wires `~/.agentsview/config.toml` (including `~/.augment/sessions-v2` if present), labels `source=auggie` sessions in SQLite, and auto-discovers local `CLAUDE.md`/`style.md` priors. |
| **2. Add Sources** | `imprint add <dir> --authored ... --chats ... --disallow-term ...` | Registers user-authored ground-truth artifacts (`.md`, `.txt`, `.docx`, `.pptx`, `.html`), exported web chat `.json` files, and workspace terms that must not leak into the portable imprint. Remote Confluence/Notion docs are fetched by the agent via MCP into `<dir>/.local/authored/`. |
| **3. Build Digest** | `imprint digest <dir>` | Deduplicates cross-store session UUIDs (e.g., Cursor `state.vscdb` vs. JSONL transcripts), strips `<worker>` / event-envelope payloads from `role=user`, splits sessions 80/20 into `train` and `holdout`, extracts text/CSS tokens from `.docx`/`.pptx`/`.html`/`.md`, and writes a token-budgeted `<dir>/.local/corpus/digest.md`. |
| **4. Synthesize** | Agent (`SKILL.md`) | Reads `.local/corpus/digest.md` and writes `.local/evidence_ledger.json` (L1 portable claims), `.local/demoted_non_imprint.json` + `.local/WORKSPACE.md` (L2 workspace bindings & L3 agent baseline), `IMPRINT.md`, and `overlays/mode-*.md`. |
| **5. Validate & Eval** | `imprint validate <dir>`<br>`imprint eval <dir>` | Enforces **Gate 1 (Portability)** and **Gate 2 (Non-Trivial Fork)**, exports the quote-free `<dir>/imprint.json`, and evaluates coverage against the untouched 20% holdout split (`.local/eval/holdout_report.md`). |

---

## 3. The Three-Layer Separation & Two Admission Gates

| Layer | What it is | Lives in |
|---|---|---|
| **L1. Personal Imprint** | Portable cognitive, dialectical, stylistic, and delegation traits of **the person** across any company or repo | `<instance>/IMPRINT.md`, `<instance>/imprint.json`, `<instance>/overlays/` (safe to commit) |
| **L2. Workspace Bindings** | How that person's imprint binds to a **specific company, stack, or repo** | `<instance>/.local/WORKSPACE.md` & `.local/demoted_non_imprint.json` (gitignored) |
| **L3. Agent Baseline** | Universal agent competence that **every** engineer expects (workarounds for model bugs) | Stripped from the person's imprint into `.local/demoted_non_imprint.json` |

Every graduated L1 claim in `imprint.json` must pass:
1. **Gate 1 — Portability Check (`portability_check`):** zero disallowed workspace/company terms in any exported field (`imprint validate` fails if a workspace term leaks).
2. **Gate 2 — Non-Trivial Fork Check (`peer_fork`):** states what another strong, reasonable engineer would legitimately prefer instead.
3. **Counter-Evidence Boundary (`flips_when`):** states the exact condition where the user intentionally suspends or inverts the rule.

---

## 4. The Five Portable Facets

1. **`F1_epistemic_posture`** — Computational primitives first, single source of truth, provenance labeling (`observed / interview / inferred / unknown`), rubric-before-benchmark, conservation-law unit economics.
2. **`F2_dialectical_style`** — Adversarial stress-testing vs. flattery, lock-in flip once an axiom is set, mental model before code/formatting, cross-model red-teaming.
3. **`F3_compression_and_structure`** — The Compression Asymmetry (terse chat/one-pagers vs. never over-compressing deep evaluations), companion-pair docs, progressive depth, capability tracks over timelines, `-v2` versioned iteration.
4. **`F4_voice_and_visuals`** — Sentence cadence, punctuation discipline, banned AI clichés, neutral structure / semantic state colors, surface-matched diagrams (SVG in published docs vs. ASCII in chat).
5. **`F5_delegation_and_risk`** — Risk-tiered autonomy gates, human-owned merge, author identity preservation, event-driven/webhook continuation over polling.
