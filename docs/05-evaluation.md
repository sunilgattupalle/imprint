# 05: Held-Out Evaluation & Living Maintenance

An imprint is useful only if it **reduces corrective turns in unseen sessions** without over-constraining the agent when the context flips.

---

## 1. Held-Out Intervention Coverage (`imprint eval`)
Using the untouched 20% holdout split (`~/.imprint/.local/corpus/holdout_turns.jsonl`):
1. Extract imprint-eligible directive and corrective turns from `holdout_turns.jsonl` (excluding raw JSON/log pastes and routine code-typo fixes).
2. Score what percentage of held-out interventions are anticipated by **L1 Portable Imprint claims** (`evidence_ledger.json` -> `imprint.json`), and report **L2 Workspace** / **L3 Agent Baseline** matches separately.
3. Write `.local/eval/holdout_report.md` and `.local/eval/holdout_results.json`.

---

## 2. Paired Artifact Check
Select 3 prompts from `holdout` (1 architecture spec, 1 executive summary / visual layout, 1 coding/refactoring task):
- Generate Response A **without** `IMPRINT.md`.
- Generate Response B **with** `IMPRINT.md` + the matching `overlays/mode-*.md`.
- Verify that Response B satisfies the user's subsequent correction turn in the held-out transcript on the first shot.

---

## 3. Maintenance Loop
All commands default to `~/.imprint`:
```bash
# Refresh deduplicated corpus, regenerate digest.md, and ensure harnesses are linked
imprint sync

# Re-validate gates, export quote-free imprint.json, and update harness links after editing claims
imprint validate

# Re-run held-out coverage check
imprint eval
```
