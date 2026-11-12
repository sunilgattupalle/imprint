# 02: Corpus Hygiene, Deduplication & Authorship Purification

Raw `role = 'user'` rows in coding-agent stores contain cross-store duplicates and machine-injected envelopes. Before synthesis runs, `imprint sync` (or `imprint digest`) executes four deterministic hygiene stages in `internal/corpus`.

---

## Stage 1: Cross-Store Session Deduplication & Metadata Merging

When sessions are indexed from both an IDE state database (for example, Cursor's `state.vscdb` under `agent = 'cursor-ide'`) and on-disk project transcripts (`~/.cursor/projects/.../agent-transcripts/` under `agent = 'cursor'`), **90–95%+ of sessions share the same UUID**:
- `cursor-ide` holds fuller message history but often lacks project metadata (`project = 'unassigned'`).
- `cursor` carries the resolved `project`, `cwd`, and `git_branch` from the workspace path.

**Rule:** Group all sessions across `sessions.db` and native JSONL stores by normalized `raw_uuid` (stripping the `<agent>:` prefix). When the same UUID appears in both `cursor-ide` and `cursor`:
1. Merge `project`, `cwd`, and `git_branch` from `cursor` onto the `cursor-ide` record.
2. Emit **one** canonical session so every human turn is counted once.

---

## Stage 2: Authorship Purification (Stripping Non-Human `user` Turns)

Agent harnesses frequently inject machine payloads into `role = 'user'` messages. A turn is rejected if it matches any of the built-in or instance-configured reject rules:

1. **Subagent / worker reports:** starts with `<worker`, `<teammate-message`, `<task-notification`, or `[Subagent Context]`.
2. **Webhook / event bus envelopes:** starts with `<*-events` or custom tags configured in `hygiene.extra_reject_patterns`.
3. **CLI harness / slash scaffolding:** starts with `<local-command-caveat`, `<command-name`, `<command-message`, `/login`, `/clear`, `/compact`, `/mcp`, `/model`, or `Use ultraplan`.
4. **System reminders:** inline `<system-reminder>...</system-reminder>` blocks are stripped before length and content checks.
5. **Pure log / stack-trace dumps:** turns with ≥6 non-empty lines where >75% of lines are stack frames or timestamped log lines with no human directive.

---

## Stage 3: Conversational Triple Construction

Every surviving user turn is paired with its surrounding assistant context and written to `.local/corpus/{train_turns,holdout_turns}.jsonl`:

```json
{
  "session_id": "claude:7100df1b-...",
  "agent": "claude",
  "project": "project_alpha",
  "ordinal": 63,
  "prev_assistant_snippet": "...assistant text that triggered the user...",
  "user_text": "Stop using consultant buzzwords. State the hypothesis plainly...",
  "next_assistant_snippet": "...how the assistant adapted...",
  "split": "train"
}
```

Keeping `prev_assistant_snippet` attached to `user_text` is critical: it preserves the **situational trigger** (what the agent did that caused the user to say *"go fast"* in one session vs. *"stop and align on the mental model"* in another).

---

## Stage 4: Deterministic 80/20 Session-Level Holdout Split

Sessions are split at the **canonical `raw_uuid` level** using `SHA256(raw_uuid)[:4] % 1000 < holdout_ratio * 1000`:
- **`.local/corpus/train_turns.jsonl` (80% of sessions):** clustered into `.local/corpus/digest.md` to synthesize and graduate Imprint claims.
- **`.local/corpus/holdout_turns.jsonl` (20% of sessions):** untouched during synthesis; scored by `imprint eval` to measure held-out coverage.
