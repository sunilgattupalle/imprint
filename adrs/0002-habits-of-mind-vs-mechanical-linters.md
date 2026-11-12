# ADR-0002: Modeling Agent Habits as Situational Reflexes (`Cue -> Routine -> Flip`), Not Regex Linters

- **Status:** Accepted
- **Date:** 2026-11-11

---

## 1. Context and Problem

Human working habits are neither impulsive reactions (System 1) nor slow, deliberate checklist reasoning (System 2). As described in behavioral psychology (*Atomic Habits*), a habit is an automatic loop:

$$\text{Cue (Situation)} \longrightarrow \text{Routine (Default Move)} \longrightarrow \text{Reward / Error ($\delta$)}$$

When building an AI agent that carries a person's imprint, two failure modes appear at opposite extremes:

1. **Failure Mode 1: Treating habits as static system-prompt prose.**
   If all of a person's habits are dumped into one static prompt (`CLAUDE.md`), the model must spend deliberate attention on every turn deciding which rules apply. As session context grows to 50k+ tokens of code and tool output, attention on static instructions decays and the agent reverts to its base RLHF training (agreeing sycophantically, jumping straight into code, or writing generic AI prose).
2. **Failure Mode 2: Reducing habits to mechanical pre-tool regex linters.**
   It is tempting to call a `PreToolUse` hook that bans words like `"load-bearing"` or blocks `gh pr merge` an "agent habit." While pre-tool guardrails are useful for a handful of hard boundaries, calling a regex linter a "habit" conflates syntax checking with judgment. A regex hook cannot identify a computational primitive, separate independent technical evaluation from org fit, or judge whether a revision compressed away the core mechanics.

---

## 2. Decision: Habits of Mind as Situational Perceptual & Action Reflexes

We define an **Imprint Habit** as a **Conditioned Perceptual and Action Policy** that fires on Turn 1 when a situational cue appears, shaping what the agent **attends to, challenges, or holds back from doing** without the user having to ask.

### Anatomy of an Imprint Habit

Every graduated claim in `imprint.json` and mode overlay in `overlays/mode-*.md` encodes four fields:

| Field | Role in the Habit Loop | Example (`IMP-F2-02`: Mental Model Before Code) |
|---|---|---|
| **`applies_when` (Cue $C$)** | The situational trigger where this dimension activates automatically. | User asks to build, redesign, or spec a non-trivial capability or multi-component change. |
| **`portable_rule` (Routine $R$)** | The involuntary Turn-1 cognitive move and tool posture. | Stop before calling `write`/`edit`; lay out the core primitive, data model, component boundaries, and reuse check first. |
| **`flips_when` (Flip Cue $\neg C$)** | The explicit signal that suspends or inverts the routine. | Once the user locks the mental model (*"go ahead"*, *"build it"*) or when fixing a localized bug/typo. |
| **`habit_reflection` (Self-Audit)** | Whether the owner confirmed the observed habit (`observed_keep`) or overrode it toward an aspirational habit (`aspirational_shift`). | `observed_keep` |

### Why Multiple Dimensions Activate Together (The Activation Function View)

A builder does not have 5 isolated personas; they have ~15–20 underlying dimensions (`epistemic.abstraction_depth`, `dialectic.challenge_threshold`, `dialectic.plan_before_code`, `structure.compression_asymmetry`, `voice.prose_cadence`, `visuals.palette_restraint`, `delegation.blast_radius_gate`, plus `emergent.*` axes).

A **Situation $S$** acts as an activation function over those dimensions:
- **Situation A: User asks `"What do you think of this repo / idea?"`**
  - Activates: `epistemic.abstraction_depth` + `dialectic.challenge_threshold` + `structure.compression_asymmetry (chat=concise)`.
  - Suppresses: file-writing and code generation.
  - Default Turn-1 move: identify the primitive, point out where concepts are conflated or overfitted, and answer directly in chat without touching files.
- **Situation B: User says `"Let's write down the spec / ADR"`**
  - Activates: `epistemic.abstraction_depth` + `structure.progressive_depth` + `voice.prose_cadence`.
  - Default Turn-1 move: read `overlays/mode-architecture-and-specs.md`, lead with purpose and hypothesis, map the construct, weave open questions inline, and write in direct engineer-to-engineer prose.
- **Situation C: User says `"Go ahead and build it"` (Flip Cue after alignment)**
  - Flips `dialectic.challenge_threshold` from debate to execution; activates `delegation.git_and_identity` + `epistemic.verification_bar`.

---

## 3. Closed-Loop Self-Learning via Meta-Correction Prediction Error ($\delta$)

How does an imprint learn and sharpen over time without manual prompt tweaking?

Whenever the user has to issue a **meta-correction** in a session:
- *"Why did I need to ask you?"*
- *"Don't make changes, I asked you a question."*
- *"Don't just agree with me, stress-test this."*
- *"You compressed this too much and lost the mechanics."*

That turn is a **Habit Prediction Error ($\delta_t \neq 0$)**: the agent encountered Cue $C_t$ and failed to fire Routine $R_t$ (or failed to flip on $\neg C_t$).

`imprint sync` mines these prediction errors from three signals:
1. **Verbal meta-corrections** (`S1-Verbal` + `FX` unclustered corrections).
2. **Tool-call trajectory pushbacks** (`S1-Trajectory`: assistant called `write`/`edit` and the user immediately aborted or scolded it on turn $t+1$).
3. **Silent draft rewrites** (`S2-Diff`: user edited the draft or created `-v2` / `-v3`).

 Because human habits form involuntarily, observed patterns are never silently locked into the active imprint without review. They surface in `imprint review`, where the user decides whether the observed pattern is who they want to be (`observed_keep`) or a habit they want to change (`aspirational_shift`).
