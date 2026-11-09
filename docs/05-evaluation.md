# 05 — Held-Out Evaluation & Living Maintenance

An Imprint is only useful if it **reduces corrective turns in future sessions** without over-constraining the agent when the context flips.

---

## 1. Held-Out Intervention Prediction (`pipeline/evaluate_holdout.py`)
Using the untouched 20% holdout split (`holdout_turns.jsonl`):
1. Extract all human corrective/directive turns from `holdout_turns.jsonl`.
2. Measure what percentage of held-out user interventions are directly anticipated by **L1 Portable Imprint claims** (`imprint_claims.json`), and track **L2 Workspace** / **L3 Agent Baseline** matches separately.
3. **Target bar:** ≥ 70% of held-out style, epistemic, and workflow corrections are covered by L1 portable claims; uncovered holdout turns are inspected to either refine a `flips_when` boundary or promote a missing claim.

---

## 2. Blind Paired Artifact Check
Pick 3 prompts from `holdout` (1 architecture spec section, 1 executive summary / visual layout, 1 coding/refactoring plan):
- Generate Response A **without** `IMPRINT.md`.
- Generate Response B **with** `IMPRINT.md` + the matching `mode:*` overlay.
- Verify that Response B satisfies the user's actual subsequent correction turn in the held-out transcript on the **first shot**.

---

## 3. Living Maintenance Loop
- **Periodic corpus refresh:** run `pipeline/build_corpus.py --instance <dir>` to ingest new sessions from `~/.agentsview/sessions.db`.
- **Gate enforcement:** run `pipeline/validate_instance.py --instance <dir>` on every claim edit so workspace nouns and generic agent-competence patches never leak into `IMPRINT.md`.
