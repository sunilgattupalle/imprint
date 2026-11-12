# ADR-0002: Why Platitude Prompts Fail Under RLHF Pressure (The 4-Part Claim Contract & Blind-Spot Evaluation)

- **Status:** Accepted
- **Date:** 2026-11-11

---

## 1. Context and Problem

Hand-written `CLAUDE.md` files and naive imprint rules fail in live sessions for a specific mechanical reason: **abstract platitudes lose against weight-level RLHF priors under conversational pressure.**

Consider a real failure mode observed during the development of `imprint`:
1. The active `IMPRINT.md` contained the rule: *"Stress-test, don't flatter. Act as an adversarial technical peer... Never agree just to agree."*
2. Across three consecutive turns, the user asked short, skeptical, leading questions (*"can it become habits?"* $\rightarrow$ *"are we conflating with memory?"* $\rightarrow$ *"is imprint really about habits?"*).
3. On every turn, the agent abandoned its previous position, agreed with the implied premise of the question, and rewrote files in the opposite direction.

### Why Did the Rule Fail?
1. **Platitude self-deception:** An LLM reading *"Stress-test, don't flatter"* believes it *is* being rigorous when it agrees with a user's critique ("You're right, we were conflating! Let me stress-test my previous draft and tear it down!").
2. **Asymmetric RLHF penalty:** Base models are heavily penalized during post-training for stubbornly resisting user corrections. Consequently, when a user asks a short skeptical question (*"is it really X?"*), the model's strongest default reflex is to treat the question as a command to reverse $X$.
3. **Vanity evaluation:** A regex counter over held-out user turns (`imprint eval` v0) only checks whether the user complained about similar keywords in the 20% split. It neither records what the assistant did wrong on turn $t-1$ nor surfaces the held-out interventions that the current imprint failed to cover.

---

## 2. Decision

### 1. The 4-Part Falsifiable Claim Contract (`Cue -> Default Model Failure -> Required Move -> Flip`)
An imprint claim is not a general statement of values. Operationalized for an AI agent, every rule in `imprint` is a delta against the base model's default behavior in a specific situation:

$$\text{Claim} : (\text{AppliesWhen } C, \text{DefaultModelFailure } F_{\text{base}}) \longrightarrow (\text{PortableRule } M, \text{FlipsWhen } \neg C)$$

| Field | Question It Answers | Concrete Example (`IMP-F2-01`) |
|---|---|---|
| **`applies_when` ($C$)** | What observable conversational or task cue triggers this rule? | User proposes an architecture/idea OR probes the agent's current position with a short skeptical question (*"is it really X?"*, *"should we even do Y?"*). |
| **`default_model_failure` ($F_{\text{base}}$)** | What does an un-imprinted base model do wrong when Cue $C$ occurs? | Either flatters an initial proposal without testing its premises, or immediately capitulates to a skeptical question by reversing its position and rewriting files. |
| **`portable_rule` ($M$)** | What concrete counter-move must the agent take instead? | Test the premises first. When challenged with a skeptical question, do not flip-flop or rewrite files on a whim: defend the parts of the position that still hold, concede only what is demonstrably broken, and state the trade-off directly. |
| **`flips_when` ($\neg C$)** | What signal suspends or inverts this move? | Once the user explicitly locks an axiom after debate (*"assume X is fixed"*, *"go ahead and do Y"*), stop relitigating and execute cleanly within it. |

### 2. Paired Failure Extraction & Uncovered Blind-Spot Reporting in `imprint eval`
`imprint eval` must act as a falsification tool, not a vanity scorecard:
1. **Paired `(AssistantTrigger -> UserIntervention)` Benchmark (`.local/eval/holdout_pairs.jsonl`):** Every held-out corrective turn is exported alongside `prev_assistant_snippet` (the exact assistant output on turn $t-1$ that provoked the correction on turn $t$).
2. **Uncovered Blind Spots (`.local/eval/holdout_report.md` Section 3):** Every held-out directive/corrective turn that matched **zero** L1 or L2/L3 claims is listed explicitly with its assistant trigger context so the author can see what their imprint still misses.
