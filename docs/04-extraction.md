# 04 — Four-Pass Synthesis & Graduation Pipeline

Synthesis turns `.local/corpus/digest.md` (`S1` clustered training turns, `S2` authored manuscripts, and `S3` explicit priors) into structured claims validated by `imprint validate <instance>`.

---

## Pass 1: High-Signal Turn Mining (`S1`)
`imprint digest` scans `.local/corpus/train_turns.jsonl` and clusters turns carrying behavioral signal across `F1`–`F5`:
1. **Corrections & pushback:** negations, imperatives, and anti-sycophancy prompts (*"don't just agree"*, *"too many colors"*, *"don't overwrite previous versions"*, *"compressed this too much"*).
2. **Structural & framing directives:** what primitive to lead with, how to version (`-v2`, `-v3`), when to align on the mental model before writing code.
3. **Autonomy & risk gates:** instructions on Git branches, author identity, deployment tiers, and what requires human sign-off.

## Pass 2: Ground-Truth Stylometric & Structural Diffing (`S2` + `S3`)
Compare extracted text and CSS tokens from the person's authored artifacts (`S2`) and explicit style guides (`S3`) against default LLM output:
- **Opening moves:** leading with a thesis/conclusion box, a concrete contrast pair, or a construct table rather than background throat-clearing.
- **Sentence rhythm & punctuation:** varied sentence length, inline open questions rather than detached meta-commentary, restraint on em-dashes and formulaic transitions.
- **Visual tokens:** `:root` CSS palette restraint (warm neutral surfaces, semantic status colors) and diagram medium selection (SVG vs. ASCII).

## Pass 3: Admission Gate Filtering & Counter-Evidence Search
For every candidate claim:
1. **Gate 1 (Portability):** Strip company/repo-specific nouns out of `portable_rule` into `workspace_bindings`. If a rule is 100% repo-specific, record it in `.local/demoted_non_imprint.json` (`L2_workspace_binding`) and `.local/WORKSPACE.md`.
2. **Gate 2 (Non-Trivial Fork):** State what a strong peer engineer would legitimately prefer (`admission_tests.peer_fork`). If no reasonable peer would prefer the opposite, move it to `.local/demoted_non_imprint.json` (`L3_agent_baseline`).
3. **Counter-Evidence Boundary (`flips_when`):** Identify where the person intentionally inverts the rule (e.g., SVG diagrams in published docs vs. ASCII diagrams in interactive chat) and record the exact boundary condition in `flips_when`.

## Pass 4: Graduation & Export
`imprint validate <instance>` enforces that every claim in `.local/evidence_ledger.json` has:
- **≥ 3 independent evidence items** spanning **≥ 2 distinct projects or source classes**
- Passing `portability_check` (zero disallowed workspace terms in any exported field) and `peer_fork`
- A concrete `flips_when` boundary condition

Passing claims are exported without verbatim quotes or session UUIDs to `<instance>/imprint.json`. `global` claims compile into `<instance>/IMPRINT.md`; `mode:*` claims compile into `<instance>/overlays/mode-*.md`.
