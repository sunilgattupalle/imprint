# 01: Source Taxonomy & Provenance

To build an accurate imprint of how a person thinks, writes, and delegates, `imprint` ingests three distinct source classes (`S1`, `S2`, `S3`). Relying on coding-agent chat logs alone biases the profile toward debugging corrections; combining session archives with pre-AI human writing and authored specs separates **how the person corrects agents** from **how the person writes and frames problems when in full control**.

---

## Source Classes

### S1. Interactive & Autonomous Session Archives
- **What it captures:** Real-time steering, pushback, scope corrections, architectural trade-offs, and autonomy boundaries across tools, paired with the preceding assistant snippet that triggered the intervention.
- **Collectors:** Native JSONL session scanner (`internal/corpus/native_scan.go`) + AgentsView SQLite archive (`~/.agentsview/sessions.db`) + exported web chat JSONs.
- **Normalized agent harnesses:**
  - IDE & CLI pairs (`cursor-ide` + `cursor`, `claude`, `codex`, `gemini`)
  - Terminal harnesses (`pi`, `omp`, `auggie` at `~/.augment/sessions-v2`, `opencode`, `hermes`)
  - Autonomous background workers (`devin`, subagent forks with `parent_session_id != ''`)
  - Web reasoning threads (`chatgpt` exported JSONs)

### S2. Authored Ground-Truth Artifacts (Split into S2A Pre-AI Gold & S2B Co-Authored)
 Documents drafted with LLMs carry subtle phrasing and structural drift that the author often stops recognizing as their own after a few months. `imprint` therefore separates authored inputs into two tiers:
- **S2A. Pre-AI / Pure-Human Voice Gold (`--pre-ai`):** Essays, specs, or notes written before AI tools or purely by hand. Used as the primary anchor for sentence cadence, opening moves, and vocabulary in `F4_voice_and_visuals`.
- **S2B. Authored & Co-Authored Specs, Decks, and Prototypes (`--authored`):**
  - Markdown and text drafts (`.md`, `.txt`), especially early `v1`–`v3` manuscripts
  - Word documents (`.docx`, parsed from `word/document.xml`)
  - Slide decks (`.pptx`, parsed slide-by-slide from `ppt/slides/slide*.xml`)
  - Visual HTML decks and prototypes (`.html`, extracting both text and `:root` `--*` CSS custom properties)
  - Remote wiki pages (Confluence, Notion, Google Docs) fetched by the agent via MCP into `~/.imprint/.local/authored/`

### S3. Curated Operating Rules (Explicit Self-Model)
- **What it captures:** Rules the user has already written down (`AGENTS.md`, `CLAUDE.md`, `style.md`).
- **Role in extraction:** Treated as **prior hypotheses** that are tested against `S1` session behavior and filtered through the **Portability** and **Non-Trivial Fork** gates.

---

## Configuration Contract (`~/.imprint/.local/sources.yaml`)

Created automatically by `imprint sync` (or `imprint init`) and updated via `imprint add`:

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
  pre_ai_globs:
    - "~/writings/pre-2023-*.md"
  globs:
    - "~/projects/*/docs/*.md"
    - "~/decks/*.pptx"

explicit_priors:
  files:
    - "~/.claude/CLAUDE.md"

hygiene:
  holdout_ratio: 0.20
  extra_reject_patterns:
    - name: "event_bus_envelope"
      regex: "^\\s*<[a-z0-9_-]+-events\\b"
  extra_inline_strip_regexes:
    - "<[a-z0-9_-]+-events\\b[^>]*>.*?</[a-z0-9_-]+-events>"

workspace_terms_disallowed_in_portable_rules:
  - "InternalDSL"
  - "staging-01"
```
