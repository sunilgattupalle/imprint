# 03 — The Five Imprint Facets & Two Admission Gates

To prevent a personal imprint from turning into a dumping ground of company trivia or generic "don't write bugs" rules, every claim must pass **two admission gates** and map to **one of five portable facets**.

---

## 1. The Two Admission Gates

### Gate 1: The Portability Gate
> *Would this still be true of the user if they switched companies or worked on an unrelated open-source project tomorrow?*

- **Pass:** *"Never bypass the canonical domain query or schema abstraction with raw ad-hoc queries."*
- **Fail (Demote to L2 Workspace Binding):** *"Use DomainQL instead of raw SQL in `query_service`."*
- **Resolution:** Factor the portable principle into `portable_rule` and record the company-specific instantiation under `workspace_bindings` (compiled separately into `WORKSPACE.md`).

### Gate 2: The Non-Trivial Fork Gate
> *Would another strong, reasonable engineer legitimately prefer the opposite?*

- **Pass:** *"Write major document or slide revisions as `-v2` / `-v3` or a separate slide instead of overwriting in place."* (A peer engineer may prefer in-place overwrites + `git diff` to avoid file clutter.)
- **Pass:** *"Human owns the merge button; the agent never merges."* (A peer engineer may prefer autonomous merge once CI passes.)
- **Fail (Demote to L3 Agent Baseline):** *"Test upload and download before claiming the feature works"* or *"Do not delete the document's top-level title when editing."* (No engineer wants broken features or deleted titles; these belong in the agent runtime baseline, not the person's imprint.)

---

## 2. The Five Portable Facets

| Facet | What it captures |
|---|---|
| **`F1_epistemic_posture`** | How the user decides what is true, well-abstracted, and properly evaluated (computational primitives first, single source of truth, explicit epistemic status, rubric-before-benchmark, independent evaluation before org fit, conservation-law math). |
| **`F2_dialectical_style`** | How the user wants the agent to collaborate and argue (stress-test vs. flatter, lock-in flip once an axiom is set, mental model before code/formatting, cross-model adversarial review). |
| **`F3_compression_and_structure`** | How the user structures density and revisions across genres (Compression Asymmetry between terse chat/summaries and deep technical evaluations, companion-pair pattern, progressive depth, capability tracks vs. timelines, versioned `-v2` iteration). |
| **`F4_voice_and_visuals`** | How the user's prose and visual taste read (architect/CTO tone, contrast pairs, banned AI clichés, restrained palette, surface-matched diagrams: ASCII in chat vs. embedded SVG/PNG in published deliverables). |
| **`F5_delegation_and_risk`** | Where the user sets the human/agent boundary (risk-tiered autonomy gates, human-owned merge, event-driven/webhook continuation over polling, Git & author identity discipline). |

---

## 3. Scope Hierarchy (`global` and `mode:*`)

Because company/repo rules are factored out into `workspace_bindings`, an Imprint claim only has two valid scopes:
1. `global` — Always active in `IMPRINT.md` (~500–650 tokens total).
2. `mode:<task_type>` — Loaded on demand from `profile/overlays/mode-<task_type>.md` (`mode:architecture-and-specs`, `mode:exec-memos-and-decks`, `mode:coding-and-prs`, `mode:autonomous-orchestration`).
