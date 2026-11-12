# 03: The Five Imprint Facets & Two Admission Gates

To prevent a personal imprint from turning into a dumping ground of company trivia or generic "don't write bugs" rules, every claim must pass **two admission gates** and map to **one of five portable facets**.

---

## 1. The Two Admission Gates

### Gate 1: The Portability Gate
> *Would this still be true of the person if they switched companies or worked on an unrelated open-source project tomorrow?*

- **Pass:** *"Never bypass the canonical domain query or schema abstraction with raw ad-hoc queries."*
- **Fail (Demote to L2 Workspace Binding):** *"Use DomainQL instead of raw SQL in `query_service`."*
- **Resolution:** Factor the portable principle into `portable_rule` and record the company-specific instantiation under `workspace_bindings` (compiled into `.local/WORKSPACE.md`).

### Gate 2: The Non-Trivial Fork Gate
> *Would another strong, reasonable engineer legitimately prefer the opposite?*

- **Pass:** *"Write major document or slide revisions as `-v2` / `-v3` or a separate slide instead of overwriting in place."* (A peer engineer may prefer in-place overwrites + `git diff` to avoid file proliferation.)
- **Pass:** *"Human owns the merge button; the agent never merges."* (A peer engineer may prefer autonomous merge once CI passes.)
- **Fail (Demote to L3 Agent Baseline):** *"Verify both upload and download before claiming the feature works."* (No engineer wants unverified features; this belongs in the agent baseline, not the person's imprint.)

---

## 2. Five Seeded Facets + Emergent Dimensions (`F6_emergent`)

| Facet | Seeded Dimensions | What it captures |
|---|---|---|
| **`F1_epistemic_posture`** | `epistemic.abstraction_depth`, `epistemic.single_source_of_truth`, `epistemic.verification_bar`, `epistemic.unit_economics` | How the person decides what is true, well-abstracted, and properly evaluated (computational primitives first, single source of truth, explicit epistemic status, rubric-before-benchmark, conservation-law unit economics). |
| **`F2_dialectical_style`** | `dialectic.challenge_threshold`, `dialectic.plan_before_code`, `dialectic.option_comparison`, `dialectic.lock_in_flip` | How the person wants the agent to collaborate and argue (adversarial stress-testing vs. flattery, lock-in flip once an axiom is set, mental model alignment before code/formatting). |
| **`F3_compression_and_structure`** | `structure.compression_asymmetry`, `structure.opening_move`, `structure.progressive_depth`, `structure.revision_mechanics` | How the person structures density and revisions across genres (Compression Asymmetry between terse chat/one-pagers and deep evaluations, companion-pair docs, progressive depth, versioned `-v2` iteration). |
| **`F4_voice_and_visuals`** | `voice.prose_cadence`, `voice.lexical_bans`, `visuals.palette_restraint`, `visuals.diagram_medium` | How the person's prose (anchored on `--pre-ai` pure-human writing) and visual taste read (sentence cadence, contrast pairs, banned AI clichés, restrained palette, surface-matched diagrams). |
| **`F5_delegation_and_risk`** | `delegation.blast_radius_gate`, `delegation.merge_ownership`, `delegation.git_and_identity`, `delegation.continuation_model` | Where the person sets the human/agent boundary (risk-tiered autonomy gates, human-owned merge, event-driven/webhook continuation over polling, Git & author identity discipline). |
| **`F6_emergent`** | `emergent.<custom_dimension>` | Recurring situational corrections or preferences discovered in the user's sessions (`FX` unclustered corrections in `digest.md`) that fall outside `F1`–`F5`. |

---

## 3. Habit Reflection (`observed_keep` vs. `aspirational_shift`)

Because habits are often automatic rather than deliberate, each claim supports a `habit_reflection` status:
- `observed_keep`: Confirmed by the owner as their authentic, intended working style.
- `aspirational_shift`: An observed pattern that the owner intentionally chooses to modify or invert going forward.

---

## 3. Scope Hierarchy (`global` and `mode:*`)

Because company/repo rules are separated into `.local/WORKSPACE.md`, an L1 Imprint claim has two valid scopes:
1. `global`: Compiled into `<instance>/IMPRINT.md` (~500–650 tokens total).
2. `mode:<task_type>`: Compiled into `<instance>/overlays/mode-<task_type>.md` (`mode:architecture-and-specs`, `mode:exec-memos-and-decks`, `mode:visual-html-and-prototypes`, `mode:coding-and-prs`, `mode:autonomous-orchestration`).
