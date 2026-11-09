# AGENTS.md — `imprint-model`

This repository contains the **person-agnostic model, schemas, Go CLI (`imprint`), and Agent Skill (`SKILL.md`)** for building a portable personal imprint.

## 1. If the User Asks You to Build or Refresh Their Imprint

Follow [`SKILL.md`](./SKILL.md) step by step:
1. Run `imprint init <instance_dir> --owner-id <id> --owner-name "<Name>" --sync`.
2. Ask the user which authored documents/decks (`S2`) and disallowed company/workspace terms to register, fetch any remote pages (Confluence/Notion/GitHub) via MCP into `<instance_dir>/.local/authored/`, and register local files via `imprint add`.
3. Run `imprint digest <instance_dir>` to produce `<instance_dir>/.local/corpus/digest.md`.
4. Synthesize `.local/evidence_ledger.json`, `.local/demoted_non_imprint.json`, `.local/WORKSPACE.md`, `IMPRINT.md`, and `overlays/`.
5. Run `imprint validate <instance_dir>` and `imprint eval <instance_dir>`.

**Never write personal or company data into this (`imprint-model`) repository.** All personal instances live in their own directory outside this repo.

---

## 2. If You Are Modifying `imprint-model` Itself

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
