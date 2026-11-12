# 04: Synthesis, Habit Reflection & Graduation Pipeline

Synthesis turns `~/.imprint/.local/corpus/digest.md` (`S1` clustered training turns, `S2A` pre-AI human gold, `S2B` authored artifacts, and `S3` explicit priors) into structured claims validated by `imprint validate`.

---

## Pass 1: Situational Turn Mining (`S1`)
`imprint sync` (or `imprint digest`) scans `.local/corpus/train_turns.jsonl`, samples round-robin across projects so no single session dominates, and pairs each user turn with its preceding assistant snippet across `F1`–`F5` plus `FX` (unclustered corrections):
1. **Corrections & pushback:** negations, imperatives, and anti-sycophancy prompts (*"don't just agree"*, *"too many colors"*, *"don't overwrite previous versions"*, *"compressed this too much"*).
2. **Structural & framing directives:** what primitive to lead with, how to version (`-v2`, `-v3`), when to align on the mental model before writing code.
3. **Autonomy & risk gates:** instructions on Git branches, author identity, deployment tiers, and what requires human sign-off.
4. **Emergent clusters (`FX`):** recurring user interventions not matched by the 5 seeded facet probes, eligible to graduate as `F6_emergent` (`IMP-E-xx`).

## Pass 2: Pre-AI Voice & Structural Diffing (`S2A` + `S2B` + `S3`)
Compare extracted text and CSS tokens from the person's pre-AI writing (`S2A`), recent specs/decks (`S2B`), and explicit priors (`S3`) against default LLM output:
- **Opening moves:** leading with a thesis/conclusion box, a concrete contrast pair, or a construct table rather than background throat-clearing.
- **Sentence rhythm & punctuation:** varied sentence length, inline open questions rather than a detached "Open Questions" section, restraint on em-dashes and formulaic transitions.
- **Visual tokens:** `:root` CSS palette restraint (neutral surfaces, semantic status colors) and diagram medium selection (SVG vs. ASCII).

## Pass 3: Admission Gate Filtering, Situational Flip & Habit Reflection
For every candidate claim:
1. **Gate 1 (Portability):** Strip company/repo-specific nouns out of `portable_rule` into `workspace_bindings`. If a rule is 100% repo-specific, record it in `.local/demoted_non_imprint.json` (`L2_workspace_binding`) and `.local/WORKSPACE.md`.
2. **Gate 2 (Non-Trivial Fork):** State what a strong peer engineer would legitimately prefer (`admission_tests.peer_fork`). If no reasonable peer would prefer the opposite, move it to `.local/demoted_non_imprint.json` (`L3_agent_baseline`).
3. **Situational Boundary (`applies_when` & `flips_when`):** Identify when the dimension activates and where the person intentionally inverts it (for example, SVG diagrams in published docs vs. ASCII diagrams in interactive chat).
4. **Habit Reflection (`habit_reflection`):** Review the discovered habit with the user and mark `"observed_keep"` or `"aspirational_shift"`.

## Pass 4: Graduation, Export & Auto-Linking
`imprint validate` enforces that every claim in `.local/evidence_ledger.json` has:
- **≥ 3 independent evidence items** spanning **≥ 2 distinct projects or source classes**
- Passing `portability_check` (zero disallowed workspace terms in any exported field) and `peer_fork`
- A concrete `flips_when` boundary condition

Passing claims are exported without verbatim quotes or session UUIDs to `~/.imprint/imprint.json`, and `imprint validate` automatically links `~/.imprint/IMPRINT.md` and `~/.imprint/.local/WORKSPACE.md` into all detected coding harnesses.
