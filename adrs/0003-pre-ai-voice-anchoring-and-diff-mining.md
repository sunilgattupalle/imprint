# ADR-0003: Preventing Identity Drift via Pre-AI Voice Anchors and Non-Verbal Diff/Trajectory Mining

- **Status:** Accepted
- **Date:** 2026-11-11

---

## 1. Context and Problem

Two hidden biases corrupt naive personal-style and preference extractors:

1. **AI-Coauthored Identity Drift (`S2` Contamination):**
   Documents drafted with the help of LLMs over the past two years carry subtle phrasing, rhythm, and structural modifications introduced by the model. While the author may accept those drafts in the moment under time pressure, reading them months later feels foreign: the thought and voice are slightly shifted from how the person actually thinks. If those AI-coauthored documents are fed back into `imprint` as ground-truth voice examples, the imprint locks in LLM-flavored drift.
2. **The Verbal-Only Blind Spot (`S1` Under-Sampling):**
   Not every builder articulates their preferences using explicit meta-language (*"use progressive depth"*, *"don't use em-dashes"*). Often, builders express their taste non-verbally:
   - By **rewriting** the draft themselves (`spec-v1.md` $\rightarrow$ `spec-v2.md`), or
   - By **aborting/pushing back** immediately after an agent calls a `write` or `edit` tool prematurely.

---

## 2. Decision

### 1. Split Authored Ground Truth into `S2A` (Pre-AI Gold) and `S2B` (Co-Authored Artifacts)
- **`S2A` (`imprint add --pre-ai <glob>`):** Writings produced before AI tools (or written purely by hand without LLM drafting). Compiled into Section `3A. Pre-AI / Pure-Human Voice Gold` of `.local/corpus/digest.md` and treated as the highest-trust anchor for `F4_voice_and_visuals` (sentence cadence, opening moves, punctuation, and vocabulary).
- **`S2B` (`imprint add --authored <glob>`):** Recent specs, slide decks (`.pptx`), Word docs (`.docx`), and HTML prototypes (`.html`). Used for structural patterns, CSS `:root` visual tokens, and domain framing, but subordinate to `S2A` whenever prose cadence conflicts.

### 2. Mine Versioned Revision Diffs (`S2-Diff`) and Tool-Call Trajectories (`S1-Trajectory`)
- **Versioned Document Diffs (`internal/digest/diffs.go`):** When sequential versions exist (`-v1` $\rightarrow$ `-v2` $\rightarrow$ `-v3`), `imprint digest` computes the word-count compression ratio, em-dash delta, heading replacements, and concrete cut/added lines. The delta between drafts captures what the author removes and tightens.
- **Tool-Call Trajectory Pushbacks (`internal/corpus/signals.go`):** `imprint digest` queries `tool_calls` joined with `messages` in `sessions.db` to extract sequences where an assistant `write`/`edit` tool call was immediately followed by a user rejection or redirection turn, surfacing delegation and planning boundaries even when the user never uses taxonomy keywords.
