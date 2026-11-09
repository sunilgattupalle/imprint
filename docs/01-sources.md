# 01: Source Taxonomy & Provenance

To build an accurate imprint of how a person thinks, writes, and delegates, `imprint` ingests three distinct source classes (`S1`, `S2`, `S3`). Relying on coding-agent chat logs alone biases the profile toward debugging corrections; combining session archives with authored manuscripts and explicit priors separates **how the person corrects agents** from **how the person writes and frames problems when in full control**.

---

## Source Classes

### S1. Interactive & Autonomous Session Archives (`sessions.db`)
- **What it captures:** Real-time steering, pushback, scope corrections, architectural trade-offs, and autonomy boundaries across tools.
- **Collectors:** AgentsView SQLite archive (`~/.agentsview/sessions.db`) + exported web chat JSONs.
- **Normalized agent ecosystems:**
  - IDE & CLI pairs (`cursor-ide` + `cursor`, `claude`, `codex`, `gemini`)
  - Terminal harnesses (`pi`, `omp`, `opencode`, `hermes`, and Harness CLI `~/.augment/sessions-v2` labeled `source=auggie` via `session_labels`)
  - Autonomous background workers (`devin`, subagent forks with `parent_session_id != ''`)
  - Web reasoning threads (`chatgpt` exported JSONs)

### S2. Authored Ground-Truth Artifacts (High-Trust Voice & Structure)
- **What it captures:** The person's actual prose, sentence cadence, vocabulary, visual hierarchy, and argument structure when writing for peers or leadership.
- **Supported formats (extracted natively by `imprint digest`):**
  - Markdown and text drafts (`.md`, `.txt`), especially early `v1`–`v3` manuscripts before heavy LLM expansion
  - Word documents (`.docx`, parsed from `word/document.xml`)
  - Slide decks (`.pptx`, parsed slide-by-slide from `ppt/slides/slide*.xml`)
  - Visual HTML decks and prototypes (`.html`, extracting both text and `:root` `--*` CSS custom properties)
  - Remote wiki pages (Confluence, Notion, Google Docs) fetched by the agent via MCP into `<instance>/.local/authored/`

### S3. Curated Operating Rules (Explicit Self-Model)
- **What it captures:** Rules the user has already codified (`AGENTS.md`, `CLAUDE.md`, `style.md`).
- **Role in extraction:** Treated as **prior hypotheses** that are cross-checked against `S1` session behavior and filtered through the **Portability** and **Non-Trivial Fork** gates.

---

## Configuration Contract (`<instance>/.local/sources.yaml`)

Created by `imprint init <instance>` and updated via `imprint add <instance>`:

```yaml
owner:
  id: "<person>"
  name: "<Full Name>"

session_archive:
  sqlite_path: "~/.agentsview/sessions.db"

chatgpt_exports:
  default_project: "exported_chats"
  globs:
    - "~/exports/chatgpt-*.json"

authored_artifacts:
  globs:
    - "~/projects/*/docs/*.md"
    - "~/decks/*.pptx"

explicit_priors:
  files:
    - "~/.claude/CLAUDE.md"

hygiene:
  holdout_ratio: 0.20
  extra_reject_patterns:
    - name: "cosmos_events"
      regex: "^\\s*<cosmos-events\\b"
  extra_inline_strip_regexes:
    - "<cosmos-events\\b[^>]*>.*?</cosmos-events>"

workspace_terms_disallowed_in_portable_rules:
  - "InternalDSL"
  - "staging-01"
```
