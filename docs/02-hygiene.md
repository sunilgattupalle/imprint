# 02 — Corpus Hygiene, Deduplication & Authorship Purification

Raw `role = 'user'` rows in coding-agent databases are heavily contaminated. Before any extraction runs, `pipeline/build_corpus.py` applies four deterministic hygiene stages.

---

## Stage 1: Cross-Store Session Deduplication & Metadata Merging

When a collector indexes both an IDE state database (e.g., Cursor's `state.vscdb` under `agent = 'cursor-ide'`) and on-disk project transcripts (`~/.cursor/projects/.../agent-transcripts/` under `agent = 'cursor'`), **90–95%+ of sessions share the exact same UUID**.

- `cursor-ide` often has slightly fuller message history (including terminal/tool outputs) but lacks the project name (`project = 'unassigned'`).
- `cursor` has the resolved `project`, `cwd`, and `git_branch` from the project directory path.

**Rule:** Group all sessions by normalized `raw_uuid` (stripping the `<agent>:` prefix). When the same UUID appears in both `cursor-ide` and `cursor`:
1. Merge `project`, `cwd`, and `git_branch` from `cursor` onto the richer transcript.
2. Keep **one** canonical session record so every human turn is counted once.

---

## Stage 2: Authorship Purification (Stripping Non-Human `user` Turns)

Agent harnesses frequently inject machine payloads into `role = 'user'` messages. A turn is **dropped** if it matches any of the following (plus any instance-configured `extra_reject_regexes`):

1. **Subagent / worker reports:** starts with `<worker>`, `<teammate-message`, `<task-notification>`, or `[Subagent Context]`.
2. **Webhook / event bus payloads:** starts with event envelope tags or raw webhook JSON payloads.
3. **CLI harness / slash scaffolding:** starts with `<local-command-caveat>`, `<command-name>`, `<command-message>`, `/login`, `/clear`, `/compact`, `/mcp`, `/model`, or `Use ultraplan`.
4. **System reminders:** inline `<system-reminder>...</system-reminder>` blocks are stripped before length and content checks.
5. **Pure log / stack-trace dumps:** turns where >75% of lines are stack frames, timestamps, or raw JSON/log pastes with no human directive sentence.

---

## Stage 3: Conversational Triple Construction

A user turn like *"no, don't do that — keep it simple"* is meaningless without what the assistant just did. Every surviving user turn is emitted as a **conversational triple**:

```json
{
  "session_id": "claude:7100df1b-...",
  "agent": "claude",
  "project": "project_alpha",
  "ordinal": 63,
  "prev_assistant_snippet": "...assistant text/action that triggered the user...",
  "user_text": "'load-bearing' is such an overused AI term. think honestly and genuinely...",
  "next_assistant_snippet": "...how the assistant adapted..."
}
```

---

## Stage 4: Deterministic 80/20 Session-Level Holdout Split

To prevent overfitting, sessions are split at the **canonical `raw_uuid` level** using `SHA256(raw_uuid) % 100 < 20`:
- **`train_turns.jsonl` (80% of sessions):** used to discover, cluster, and graduate Imprint claims.
- **`holdout_turns.jsonl` (20% of sessions):** untouched during extraction; used in `05-evaluation.md` to test whether the graduated Imprint predicts the user's actual interventions in unseen sessions.
