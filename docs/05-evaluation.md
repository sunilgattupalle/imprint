# 05: Held-Out Evaluation & Living Maintenance

An imprint is useful only if it **reduces corrective turns in unseen sessions** without over-constraining the agent when the context flips.

---

## 1. Held-Out Intervention Coverage (`imprint eval <instance>`)
Using the untouched 20% holdout split (`.local/corpus/holdout_turns.jsonl`):
1. Extract human corrective and directive turns from `holdout_turns.jsonl` (excluding raw JSON/log pastes).
2. Score what percentage of held-out interventions are anticipated by **L1 Portable Imprint claims** (`evidence_ledger.json` -> `imprint.json`), and report **L2 Workspace** / **L3 Agent Baseline** matches separately.
3. Write `.local/eval/holdout_report.md` and `.local/eval/holdout_results.json`.
4. **Target bar:** ≥ 70% of held-out epistemic, structural, stylistic, and delegation interventions covered by L1 portable claims.

---

## 2. Paired Artifact Check
Select 3 prompts from `holdout` (1 architecture spec, 1 executive summary / visual layout, 1 coding/refactoring task):
- Generate Response A **without** `IMPRINT.md`.
- Generate Response B **with** `IMPRINT.md` + the matching `overlays/mode-*.md`.
- Verify that Response B satisfies the user's subsequent correction turn in the held-out transcript on the first shot.

---

## 3. Maintenance Loop
```bash
# Refresh deduplicated corpus and digest after new sessions accumulate
imprint digest <instance_dir>

# Re-validate gates and re-export quote-free imprint.json after editing claims
imprint validate <instance_dir>

# Re-run held-out coverage check
imprint eval <instance_dir>
```
