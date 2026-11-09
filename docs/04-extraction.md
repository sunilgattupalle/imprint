# 04 — Four-Pass Extraction & Graduation Pipeline

Extraction turns `train_turns.jsonl` (S1), authored documents (S2), and explicit priors (S3) into structured, evidence-backed claims validated against `schemas/claim.schema.json`.

---

## Pass 1: High-Signal Turn Mining (S1)
Scan `train_turns.jsonl` for turns carrying high behavioral signal:
1. **Corrections & pushback:** negations, imperatives, or frustration markers (*"don't just agree"*, *"too many colors"*, *"don't overwrite previous versions"*, *"compressed this too much"*, *"not what exists — focus on this document"*).
2. **Structural & framing directives:** how to start, what primitive to use, how to version (`-v2`, `-v3`), when to plan before coding.
3. **Autonomy & risk gates:** instructions on Git branches, author identity, environments, and what requires human sign-off.

## Pass 2: Ground-Truth Stylometric & Structural Diffing (S2 + S3)
Compare the user's authored artifacts (S2) and curated style guides (S3) against standard LLM prose across:
- **Opening moves:** Does the author start with a thesis/conclusion box, a concrete contrast (*"A developer does not experience X as ... They experience: ..."*), or a construct table?
- **Paragraph & sentence rhythm:** Variation in sentence length, avoidance of repetitive parallel triple-lists or identical sentence openers, inline open questions rather than detached "Open Questions" sections.
- **Lexical signature:** Recurring precision terms (*"mental model"*, *"primitives and constructs"*, *"the hypothesis"*) vs. explicitly rejected terms (*"load-bearing"*, *"the bet"*, *"leverage"* as a verb, overused em-dashes).

## Pass 3: Admission Gate Filtering & Counter-Evidence Search
For every candidate claim:
1. **Apply Gate 1 (Portability):** Strip company/repo-specific nouns out of `portable_rule` into `workspace_bindings`. If the rule is 100% repo-specific, move it to `demoted_non_imprint.json` (`L2_workspace_binding`).
2. **Apply Gate 2 (Non-Trivial Fork):** State what a reasonable peer engineer would prefer (`admission_tests.peer_fork`). If no reasonable engineer would prefer the opposite, move it to `demoted_non_imprint.json` (`L3_agent_baseline`).
3. **Search for Counter-Evidence (`flips_when`):** Search `train_turns.jsonl` and S3 priors for places where the user appears to do the opposite (e.g., asking for SVG diagrams in published docs vs. asking for ASCII diagrams in interactive chat explanations) and synthesize the exact `flips_when` boundary condition.

## Pass 4: Graduation & Compilation
A candidate claim graduates to `status: "graduated"` only when it has:
- **≥ 3 independent evidence items** spanning **≥ 2 distinct projects or source classes**
- Both `admission_tests` (`portability_check` and `peer_fork`) populated and passing `pipeline/validate_instance.py`
- A concrete `flips_when` condition

Graduated `global` claims compile into `<instance>/profile/IMPRINT.md`. `mode:*` claims compile into `<instance>/profile/overlays/mode-*.md`. Separated `L2` workspace bindings compile into `<instance>/profile/WORKSPACE.md`.
