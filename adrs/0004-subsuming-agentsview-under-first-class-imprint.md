# ADR-0004: First-Class `imprint` (`~/.imprint`) Subsuming Session Collection and `agentsview`

- **Status:** Accepted
- **Date:** 2026-11-11

---

## 1. Context and Problem

In its initial prototype (`imprint-model`), building an imprint required the user to:
1. Install and configure an external tool (`agentsview`) and manually edit `~/.agentsview/config.toml` to include Auggie sessions (`pi_dirs`),
2. Run a multi-step CLI sequence (`imprint init <dir> --owner-id ... --sync`, `imprint add <dir> ...`, `imprint digest <dir>`, `imprint validate <dir>`, `imprint eval <dir>`), and
3. Manually copy `@`-references into each coding agent's configuration file.

Meanwhile, inspecting the `agentsview` codebase (`go.kenn.io/agentsview`) showed that `agentsview` is a 70-package observability monorepo pulling in CGO (`mattn/go-sqlite3`, `sqlite-vec`, `duckdb`), ClickHouse, Postgres, AWS S3, Docker, and a Svelte frontend. Merging all 70 packages into `imprint` would bloat `imprint` with heavy infrastructure dependencies, whereas keeping `agentsview` as a manual prerequisite created unacceptable onboarding friction.

---

## 2. Decision

1. **`imprint` Is the First-Class Product and CLI (`github.com/sunilgattupalle/imprint`):**
   - All commands default to `~/.imprint` with zero positional arguments required, and auto-discover owner identity from OS/Git configuration.
   - `imprint sync` collapses initialization, session syncing, deduplication, multi-signal digest generation, and harness linking into one command.
   - `imprint link` automatically wires `@~/.imprint/IMPRINT.md` and `@~/.imprint/.local/WORKSPACE.md` into all detected local agent harnesses (`~/.claude/CLAUDE.md`, `~/.pi/agent/AGENTS.md`, `~/.omp/agent/AGENTS.md`, `~/.config/opencode/AGENTS.md`, `~/.hermes/SOUL.md`) inside an idempotent managed block.

2. **Built-In Pure-Go Session Scanner + Read-Only `sessions.db` Integration:**
   - `imprint` remains a fast, pure-Go (`modernc.org/sqlite`, zero CGO) binary.
   - `internal/corpus/native_scan.go` natively scans local JSONL and JSON session stores (Auggie v2 `~/.augment/sessions-v2`, Auggie v1 `~/.augment/sessions`, Pi `~/.pi/agent/sessions`, Oh My Pi `~/.omp/agent/sessions`, and Claude `~/.claude/projects`), while reading `~/.agentsview/sessions.db` (`?mode=ro`) when present and deduplicating across both by canonical session UUID.

3. **Targeted Fork of `sunilgattupalle/agentsview` as the Background Sensor & Recall Substrate:**
   - In `github.com/sunilgattupalle/agentsview`, we added native `AgentAuggie` (`auggie`, `.augment/sessions-v2`) to `internal/parser/` so `config.toml` and `session_labels` workarounds are no longer needed, while keeping the fork cleanly rebaseable against upstream parser updates.
