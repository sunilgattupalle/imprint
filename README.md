# Imprint Model (`imprint-model`)

**Imprint** is a local-first model and pipeline for distilling how a specific builder thinks, argues, writes, structures information, and delegates risk from their raw coding-agent sessions and authored manuscripts.

This repository contains **only the person-agnostic model, schemas, and pipeline**. It contains zero personal or company data. A person's instance lives in a separate directory (e.g., `<person>-imprint/`) and is compiled, validated, and evaluated by `pipeline/`.

---

## 1. The Three-Layer Separation

Most "memory" or `AGENTS.md` / `CLAUDE.md` files degrade because they mix three distinct layers into a single file:

| Layer | What it is | Lives in | Example |
|---|---|---|---|
| **L1. Personal Imprint** | Portable cognitive, dialectical, stylistic, and delegation traits of **the person** across any company or repo | `<instance>/claims/imprint_claims.json` → `<instance>/profile/IMPRINT.md` | *"Map a problem to its computational primitive first so the API falls out of the construct."* |
| **L2. Workspace Bindings** | How that person's imprint binds to a **specific company, stack, or repo** | `<instance>/claims/demoted_non_imprint.json` → `<instance>/profile/WORKSPACE.md` (or repo `AGENTS.md`) | *"Use the internal DSL rather than raw SQL; validate in `staging-01` first; Part II = organization fit."* |
| **L3. Agent Baseline** | Universal agent competence that **every** engineer expects (workarounds for model bugs) | Agent system prompt / runtime (stripped from the person's imprint) | *"Test upload/download before claiming done"*, *"Don't delete the top title when editing"*, *"Don't take 20 tool calls to read comments."* |

---

## 2. The Two Admission Gates

Before any pattern observed in session history can enter a person's **Imprint (L1)**, it must pass two explicit tests recorded on the claim object:

1. **Gate 1 — The Portability Test (`portability_check`):**
   *If this person moved to a different company or worked on an open-source repo tomorrow, would this rule still hold?*
   - Any company, environment, or repo nouns must be factored out into `workspace_bindings` so the core `portable_rule` is 100% workspace-agnostic.
2. **Gate 2 — The Non-Trivial Fork Test (`peer_fork`):**
   *Would another strong, reasonable engineer legitimately prefer the opposite?*
   - **Passes:** *"Human owns the merge button; the agent never merges"* (another strong engineer prefers auto-merge when CI is green).
   - **Passes:** *"Version major revisions as `-v2`/`-v3` rather than overwriting in place"* (another engineer prefers in-place edits and `git diff`).
   - **Fails (Demoted to L3 Agent Baseline):** *"Verify the code builds and works before claiming completion"* (no engineer prefers broken code; this is baseline agent competence, not a personal imprint trait).

---

## 3. The Five Imprint Facets

Every graduated claim belongs to one of five portable facets:

1. **`F1_epistemic_posture`** — How the person establishes truth, abstractions, and evaluation rubrics (primitives first, single source of truth, provenance labeling, rubric-before-benchmark, physical conservation laws in modeling).
2. **`F2_dialectical_style`** — How the person wants an agent to reason *with* them (adversarial stress-testing vs. agreement, when to stop debating and execute a locked axiom, mental-model-before-code/formatting, cross-model red-teaming).
3. **`F3_compression_and_structure`** — How the person structures density across genres (the Compression Asymmetry between terse chat/one-pagers and deep technical evaluations, companion-pair documents, progressive depth from What/Why to How, capability tracks vs. timelines, versioned `-v2` iteration).
4. **`F4_voice_and_visuals`** — How the person's prose and visual taste read (sentence rhythm, contrast pairs, banned AI clichés, restrained color palette, surface-matched diagrams: ASCII in chat vs. embedded SVG/PNG in published artifacts).
5. **`F5_delegation_and_risk`** — Where the person draws the human/agent autonomy boundary (what the agent drives autonomously, what triggers a mandatory human gate, who merges, and how the agent transitions from active driver to event-driven observer).

---

## 4. Running the Model Pipeline on an Instance

```bash
# 1. Build clean, deduplicated train/holdout corpus from <instance>/sources.yaml
python3 pipeline/build_corpus.py --instance <path-to-instance>

# 2. Validate instance claims against schema + Portability & Fork gates
python3 pipeline/validate_instance.py --instance <path-to-instance>

# 3. Evaluate portable imprint claims against untouched 20% holdout sessions
python3 pipeline/evaluate_holdout.py --instance <path-to-instance>
```
