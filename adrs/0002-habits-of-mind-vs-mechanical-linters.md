# ADR-0002: Separating Deliberate Imprint (Judgment, Taste, Voice) from Involuntary Habits

- **Status:** Accepted
- **Date:** 2026-11-11

---

## 1. Context and Problem

When extracting a personal profile from historical session transcripts (`S1`) and applying it to an agent, it is easy to conflate **Imprint** with **Habits**:
- **On the human side (extraction):** Session logs record everything a person does repeatedly. Some of those recurring patterns reflect **deliberate judgment, taste, and epistemic standards** (how the person thinks and writes at their best). Others are merely **involuntary habits** (shortcuts taken when rushed, or mechanical routines done without thinking). If an extractor blindly promotes every recurring behavior into the imprint, it codifies involuntary habits alongside genuine judgment.
- **On the agent side (execution):** A builder wants the agent to apply their imprint *automatically* (without having to repeat *"stress-test this"* or *"find the primitive first"* on every turn). However, wanting the agent to apply an imprint *without being reminded* does not mean the imprint itself is a list of mechanical habits or regex linters.

---

## 2. Decision: Four Distinct Primitives (Context, Memory, Habit, Imprint)

We keep **Imprint** and **Habit** strictly distinct:

| Primitive | Definition | Role in `imprint` |
|---|---|---|
| **1. Workspace Context** | Current state of the repo and organization. | `.local/WORKSPACE.md` (L2 bindings) |
| **2. Episodic Memory** | History of project events, facts, and decisions. | External session archive (`sessions.db` / `recall_entries`) |
| **3. Habit** | An involuntary recurring routine in human logs, or a mechanical trigger-action reflex in execution. | Filtered during synthesis and review (`habit_reflection`) so unwanted habits do not pollute the imprint |
| **4. Personal Imprint** | **The person's deliberate epistemic posture (`F1`), dialectical style (`F2`), structural taste (`F3`), voice/visual identity (`F4`), and risk governance (`F5`).** | `IMPRINT.md`, `imprint.json`, and `overlays/mode-*.md` (L1) |

### 1. Why `S2` (Authored Writing) Anchors `S1` (Session Logs)
Chat transcripts (`S1`) are noisy and reactive. Authored manuscripts, architecture specs, and decks (`S2A` pre-AI writing + `S2B` authored artifacts) capture how the person frames problems and writes when in full deliberate control. Synthesis always anchors voice and epistemic structure on `S2`, using `S1` to discover where the person had to correct agents that drifted from that standard.

### 2. Why `habit_reflection` Exists (`observed_keep` vs. `aspirational_shift`)
Because recurring patterns in `S1` can be either authentic judgment or involuntary habit, `imprint review` surfaces every graduated claim for explicit confirmation:
- `observed_keep`: The owner confirms this pattern reflects their deliberate judgment, taste, or governance boundary.
- `aspirational_shift`: The owner recognizes an involuntary pattern in their past logs and intentionally overrides or refines the rule toward how they *want* to work.

### 3. Situational Activation Without Reducing Judgment to Linters
An agent applies the person's Imprint without needing constant reminders because **Situation $S$** (task intent + target artifact surface) activates the relevant combination of dimensions via `imprint brief` and the Situational Dispatch Table in `IMPRINT.md`:
- Every claim defines when it activates (`applies_when`), the judgment/taste principle (`portable_rule`), and the exact boundary where the person intentionally inverts it (`flips_when`).
- When the user has to issue a meta-correction in chat (*"why did I need to ask you?"*, *"don't just agree, stress-test this"*), that turn is recorded as a failure of situational activation, refining `applies_when` and `flips_when` on the next `imprint sync`.
