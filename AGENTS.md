# AGENTS.md: `imprint`

This repository contains the **person-agnostic model, schemas, Go CLI (`imprint`), and Agent Skill (`SKILL.md`)** for building a portable personal imprint.

## 1. If the User Asks You to Build or Refresh Their Imprint

Follow [`SKILL.md`](./SKILL.md) step by step (all commands default to `~/.imprint` when `[instance_dir]` is omitted):
1. Ask the user for any pre-AI writings (`--pre-ai`), authored documents/decks (`--authored`), and disallowed company/workspace terms (`--disallow-term`), fetch remote pages (Confluence/Notion/GitHub) via MCP into `~/.imprint/.local/authored/`, and register local files via `imprint add`.
2. Run `imprint sync` to initialize `~/.imprint`, sync sessions, deduplicate turns, and produce `~/.imprint/.local/corpus/digest.md`.
3. Synthesize `.local/evidence_ledger.json` (across `F1`–`F5` seeded facets + `F6_emergent` clusters), `.local/demoted_non_imprint.json`, `.local/WORKSPACE.md`, `IMPRINT.md`, and `overlays/`.
4. Run the Habit Reflection check with the user (`observed_keep` vs. `aspirational_shift`), then run `imprint validate` (which also auto-links all local coding harnesses) and `imprint eval`.

**Never write personal or company data into this (`imprint`) repository.** All personal instances live in `~/.imprint` (or another directory outside this repo).

---

## 2. If You Are Modifying `imprint` Itself

### Non-Negotiable Invariants
1. **Zero personal or company leakage:** Never commit a real person's name, session UUID, verbatim transcript quote, or company/repo proper noun into code, tests, docs, or schemas. Use generic placeholders (`alice`, `project_alpha`, `InternalSQL`) in tests and examples.
2. **Hard determinism boundary:** Keep `cmd/imprint` and `internal/*` strictly deterministic and offline (no LLM API keys or network calls inside the Go binary). Synthesis belongs to the agent executing `SKILL.md`.
3. **Two-zone privacy enforcement:** `imprint validate` must always read raw quotes from `<instance>/.local/evidence_ledger.json`, verify zero disallowed workspace terms appear in portable fields, and write only quote-free, UUID-free records to `<instance>/imprint.json`.
4. **Read-only archive access:** `internal/corpus` must open `~/.agentsview/sessions.db` in read-only mode (`?mode=ro`). Only `imprint init` writes `source=auggie` labels into `session_labels`.

### Build & Verification Commands
Run before committing any change:
```bash
go test ./...
go build -o ~/.local/bin/imprint ./cmd/imprint
```
