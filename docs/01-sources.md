# 01 — Source Taxonomy & Provenance

To build an accurate imprint of how a person thinks, writes, and delegates to agents, the pipeline ingests three distinct source classes (`S1`, `S2`, `S3`). Relying on coding-agent chat logs alone biases the profile toward debugging frustration; combining chat logs with authored writing and explicit rules separates **how the person corrects agents** from **how the person thinks and writes when in full control**.

---

## Source Classes

### S1. Interactive & Autonomous Session Archives (`sessions.db`)
- **What it captures:** Real-time steering, pushback, scope corrections, architectural push-and-pull, and autonomy boundaries across tools.
- **Supported collectors:** AgentsView SQLite archive (`~/.agentsview/sessions.db`) + exported web chat JSONs.
- **Agent ecosystems normalized:**
  - IDE & CLI pairs (`cursor-ide` + `cursor`, `claude`, `codex`, `gemini`)
  - Terminal harnesses (`pi`, `omp`, and Pi-compatible JSONL session stores labeled via `session_labels`)
  - Autonomous background workers (`devin`, subagent forks with `parent_session_id IS NOT NULL`)
  - Web reasoning threads (`chatgpt` shared/exported JSONs)

### S2. Authored Ground-Truth Artifacts (High-Trust Voice & Structure)
- **What it captures:** The user's actual prose, sentence cadence, vocabulary, visual layout, and argument structure when writing for peers, executives, or open-source readers.
- **Artifacts:**
  - Early user-written drafts (`v1`–`v3` manuscripts before heavy LLM expansion)
  - Published technical specs, architecture docs, and wiki pages authored by the user
  - Slide decks (`.pptx`, `.html`, Marp `.md`) and executive memos

### S3. Curated Operating Rules (Explicit Self-Model)
- **What it captures:** Rules the user has already codified after repeated agent failures (`AGENTS.md`, `CLAUDE.md`, `writing-style/style.md`, `presentation-style/style.md`, `prototype-style/style.md`, `SKILL.md`).
- **Role in extraction:** Treated as **prior hypotheses** that are cross-checked against `S1` session behavior (to verify whether the user actually enforces them or overrides them under deadline pressure) and filtered through the **Portability** and **Non-Trivial Fork** gates.

---

## Configuration Contract (`<instance>/sources.yaml`)

Each personal instance defines a `sources.yaml` declaring its local paths and instance-specific hygiene rules:

```yaml
owner:
  id: "<person>"
  name: "<Full Name>"

session_archive:
  sqlite_path: "~/.agentsview/sessions.db"

chatgpt_exports:
  globs:
    - "~/exports/chatgpt-*.json"

authored_artifacts:
  globs:
    - "~/projects/*/docs/*.md"

explicit_priors:
  files:
    - "~/personal-ai/writing-style/style.md"
    - "~/personal-ai/presentation-style/style.md"

hygiene:
  holdout_ratio: 0.20
  extra_reject_regexes:
    - "^\\s*<custom-event-envelope>"

workspace_terms_disallowed_in_portable_rules:
  - "InternalDSL"
  - "staging-01"
```
