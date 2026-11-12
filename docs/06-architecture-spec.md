# 06: Situational Activation, Multi-Signal Extraction & Conditioned Memory Specification

## 1. Purpose and Core Hypothesis

Static instruction files (`CLAUDE.md`, `AGENTS.md`) and flat episodic memory stores fail for opposite reasons:
- A static instruction file applies every rule to every prompt. Rules written to keep chat answers terse compress the technical mechanics out of architecture specs; rules written for deep specs turn a five-line bug fix into an essay.
- Flat episodic memory systems (`recall_entries`, vector stores over transcripts) record passive observations (*"Project X uses SQLite"*) and discard the person's judgment: what risks they refused, why they rejected an abstraction, or when they aborted an agent that started coding before aligning on the design.

**The hypothesis:** A builder's imprint is a set of behavioral dimensions that activate in combinations based on the current situation. If we model each rule and project memory as a **situational policy** $(S, W) \rightarrow D \oplus B(W, S)$, and mine those policies from three independent signals (verbal pushback, tool-call trajectory interruptions, and before/after document diffs), the system can both discover a person's real habits without keyword heuristics and activate the exact slice needed for each task.

### Evaluation Rubric
1. **Deterministic situational routing (<5ms, zero LLM calls):** Given `--cwd`, `--prompt`, and optional target file paths, `imprint brief` selects the matching mode overlay and project-scoped memories with zero manual `@`-mentions.
2. **Non-verbal signal extraction:** `imprint sync` surfaces (a) tool-call trajectory interruptions (e.g., user aborts or corrects an assistant that jumped into `write`/`edit` before design alignment) and (b) concrete prose/structure diffs between `-v1` and `-v2` drafts (or assistant tool-writes vs. on-disk user files).
3. **Habit reflection control:** `imprint review` lists every graduated L1 claim with its dimension, provenance counts, and current `habit_reflection` state (`observed_keep`, `aspirational_shift`, `unreviewed`), allowing one-command updates.
4. **Held-out intervention coverage:** Validated L1 claims + L2 project memories cover $\ge 60\%$ of held-out directive/corrective turns without leaking any workspace proper nouns into `imprint.json`.

---

## 2. Mental Model & Construct

### The Computational Primitive: Conditioned Behavioral Policy

Every portable claim (L1) and workspace memory (L2) is an instance of the same primitive:

$$\text{Policy } p : (\text{Situation } S, \text{Workspace } W) \xrightarrow{\text{activates}} (\text{Dimension } d, \text{Directive } r, \text{Inverter } \neg S)$$

| Component | Definition | Determinism & Scrutiny Class |
|---|---|---|
| **Situation $S$** | Tuple of `(intent, surface, tempo)` inferred from prompt text and target file paths (`architecture-and-specs`, `exec-memos-and-decks`, `visual-html-and-prototypes`, `coding-and-prs`, `autonomous-orchestration`). | **Deterministic classifier** (`internal/brief`). Fast lexical + file-glob matching; no network call. |
| **Workspace $W$** | Resolved repository name and working directory (`--cwd`). | **Deterministic path resolution.** |
| **Dimension $d$** | Seeded axis (`epistemic.*`, `dialectic.*`, `structure.*`, `voice.*`, `visuals.*`, `delegation.*`) or emergent axis (`emergent.*`). | **Human-reviewed schema (`claim.schema.json`).** |
| **Inverter $\neg S$ (`flips_when`)** | Explicit boundary condition where the owner suspends or inverts $r$ (for example, tempo shifts from deliberate design to rapid one-off script). | **Enforced at validation time (`internal/gate`).** |

### Determinism Boundary

```text
+-------------------------------------------------------------------------+
| Deterministic Offline Plane (`imprint` Go Binary)                       |
|                                                                         |
|  [sessions.db + JSONL + S2 Docs]                                        |
|         |                                                               |
|         +--> 1. Multi-Signal Miner (`internal/corpus`, `internal/digest`)|
|         |       - Verbal:     friction_findings + conversational triples |
|         |       - Trajectory: tool_calls write/edit before alignment     |
|         |       - Diff:       -v1 -> -v2 & assistant-write vs. disk file |
|         |                                                               |
|         +--> 2. Gate Validator & Habit Reviewer (`gate`, `review`)      |
|         |       - Enforces Gate 1 (Portability) & Gate 2 (Peer Fork)    |
|         |       - Manages `habit_reflection` & `.local/memories.json`   |
|         |                                                               |
|         +--> 3. Runtime Situational Router (`internal/brief`)           |
|                 - Resolves (S, W) -> IMPRINT.md + active mode-*.md      |
|                   + repo-filtered .local/WORKSPACE.md & memories.json   |
+-------------------------------------------------------------------------+
                                    ^
                                    | Reads digest.md; writes ledger & overlays
+-----------------------------------+-------------------------------------+
| Synthesis Plane (Coding Agent executing `SKILL.md`)                     |
|  - Clusters candidate turns, trajectory aborts, and draft diffs         |
|  - Factors portable L1 claims vs. L2 project memories vs. L3 baseline   |
+-------------------------------------------------------------------------+
```

---

## 3. Multi-Signal Extraction (`internal/corpus` & `internal/digest`)

Relying only on what a user types in prompts misses what they do when they edit files or stop a runaway agent. `imprint digest` mines three distinct signal streams into `.local/corpus/digest.md`:

### Signal Stream 1: Verbal Corrections & `friction_findings` (`S1-Verbal`)
- **Source (`observed`):** Reads `friction_findings` (`kind IN ('correction', 'frustration', 'interruption')`) from `~/.agentsview/sessions.db` when present, combined with the conversational triple scanner over `messages`.
- **Output:** Turns grouped into the 5 seeded facets (`F1`–`F5`) plus `FX: Emergent / Unclustered Corrections`, round-robin sampled across projects so a single noisy repository never dominates the digest.

### Signal Stream 2: Tool-Call Trajectory & Premature-Action Signals (`S1-Trajectory`)
- **Source (`observed`):** Queries `tool_calls` joined with `messages` in `sessions.db` (and native JSONL tool blocks).
- **What it detects:**
  1. **Premature write/edit before alignment:** Assistant invokes a `write` or `edit` tool call on its first turn of a session or immediately before a user correction/rejection turn.
  2. **Aborted or errored command loops:** Sequences where `observed_outcome = 'error'` or `sequence_ending` is followed by a user interventionredirecting the approach.
- **Why this matters:** Captures `F2_dialectical_style` (`dialectic.plan_before_code`) and `F5_delegation_and_risk` (`delegation.blast_radius_gate`) even if the user never says the words *"mental model"*.

### Signal Stream 3: Draft-to-Final & Versioned Document Diffs (`S2-Diff`)
- **Source (`observed`):**
  1. **Versioned sibling pairs** among authored artifacts (e.g., `foo-v1.md` $\rightarrow$ `foo-v2.md`, `deck-v1.html` $\rightarrow$ `deck-v2.html`).
  2. **Assistant-draft vs. on-disk file diffs:** When `tool_calls` records the content written by an assistant `write` tool call to an authored `.md`/`.html` path that still exists on disk and was subsequently modified, compute a compact structural/line diff summary (lines removed, headings changed, banned phrases stripped, length delta).
- **Why this matters:** Solves the AI-tainted prose problem. The delta between what the LLM wrote and what the human kept is pure ground truth for `F3_compression_and_structure` and `F4_voice_and_visuals`.

---

## 4. Episodic Project Memory & Habit Reflection

### Project Memory Ledger (`.local/memories.json` + `recall_entries` Bridge)
While `.local/WORKSPACE.md` holds static workspace bindings, `.local/memories.json` stores structured episodic project memories (seeded from `recall_entries` in `sessions.db` where `status = 'accepted'` and from agent synthesis):

```json
[
  {
    "id": "MEM-01",
    "project": "imprint",
    "type": "decision",
    "dimension": "epistemic.single_source_of_truth",
    "trigger": "Modifying schemas or CLI flags",
    "rule": "Keep cmd/imprint and internal/* strictly deterministic and offline; never add LLM API calls inside the Go binary.",
    "source_sessions": ["claude:7100df1b"]
  }
]
```

### Habit Reflection CLI (`imprint review`)
- `imprint review`: Displays a table of all L1 claims (`id`, `facet`, `dimension`, `habit_reflection`, `provenance_counts`, `title`) and L2 project memories.
- `imprint review --set <claim_id>=<observed_keep|aspirational_shift>`: Updates `habit_reflection` in `.local/evidence_ledger.json` and re-runs `ValidateAndExport` so `imprint.json` stays in sync.

---

## 5. Runtime Situational Router (`imprint brief`)

### Self-Routing `IMPRINT.md` Header (Zero-Hook Path)
Every generated `IMPRINT.md` includes a compact **Situational Dispatch Table** at the top of the file mapping task types and target file patterns to the exact overlay in `~/.imprint/overlays/mode-*.md`. Even in tools that only load `CLAUDE.md` / `AGENTS.md` statically, the agent is instructed to `read` the matching overlay before acting whenever the task matches.

### Deterministic CLI Router (`imprint brief`)
```bash
imprint brief [--cwd <dir>] [--prompt "<user prompt>"] [--files "README.md,docs/06.md"] [--format markdown|json]
```
1. **Mode classification:** Scores the 5 modes against `--prompt` and `--files`:
   - `mode-architecture-and-specs.md`: `README.md`, `AGENTS.md`, `SKILL.md`, `docs/**`, `*spec*.md`, `schemas/**`, or prompt terms (`architecture`, `spec`, `design`, `mental model`, `construct`, `ontology`, `evaluate`, `compare`, `think through`, `readme`).
   - `mode-exec-memos-and-decks.md`: `*.pptx`, `*deck*`, `*memo*`, `*slides*`, or prompt terms (`deck`, `slide`, `board`, `executive`, `memo`, `one-pager`, `presentation`).
   - `mode-visual-html-and-prototypes.md`: `*.html`, `*.css`, `*.svelte`, `*.tsx`, or prompt terms (`html`, `prototype`, `ui`, `visual`, `diagram`, `svg`, `palette`, `layout`, `css`).
   - `mode-autonomous-orchestration.md`: prompt terms (`subagent`, `parallel`, `background`, `worktree`, `autonomous`, `batch`, `fan out`, `webhook`).
   - `mode-coding-and-prs.md`: `*.go`, `*.py`, `*.ts`, `*.java`, `*.rs`, or prompt terms (`implement`, `fix`, `test`, `commit`, `pr`, `pull request`, `refactor`, `build`).
2. **Workspace & memory resolution:** Infers the active project from `--cwd` (git repo root or directory basename), includes `.local/WORKSPACE.md` if present, and filters `.local/memories.json` (plus `recall_entries` in `sessions.db`) to entries matching `project` or `scope == "global"`.
3. **Output:** Emits the assembled, token-budgeted situational brief ready for direct agent consumption or hook injection.
