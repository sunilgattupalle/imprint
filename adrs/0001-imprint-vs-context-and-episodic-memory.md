# ADR-0001: Separating Personal Imprint from Workspace Context and Episodic Memory

- **Status:** Accepted
- **Date:** 2026-11-11

---

## 1. Context and Problem

Builders working across multiple AI tools and repositories face three distinct continuity problems that are routinely conflated into a single `CLAUDE.md` file or "agent memory" database:
1. Knowing what exists in the organization or repository right now.
2. Remembering what happened in past sessions on a project (which architecture decision was made, which bug was fixed, which migration failed).
3. Carrying **who the builder is** (how they judge abstractions, when they want pushback vs. execution, where they draw risk boundaries, and how they write) across any tool or project.

Existing "agent memory" systems take raw session transcripts and distill flat episodic or semantic facts (`"Service X uses SQLite"`, `"PR #42 fixed a race condition"`). That treats every session as a passive observation of the world and discards the human operator: why they rejected a proposal, what risks they refused to take, and how they framed the problem.

Conversely, if we store project-level episodic facts inside `imprint` (for example, `.local/memories.json` holding repo-specific facts), `imprint` degenerates into another project-memory database and mixes workspace state with personal identity.

---

## 2. Decision: The Three-Construct Boundary

We treat an agent working with a person over time as a system with three distinct constructs that must never be stored in the same layer:

| Construct | Mathematical Role | Question It Answers | Owner & Lifespan | Where It Lives |
|---|---|---|---|---|
| **1. Workspace Context** | Current state $x_t$ | *What exists in this repo/org right now?* | Owned by the **repo/org**; shared by all engineers on the team | Repo `AGENTS.md`, code, schemas, and local `.local/WORKSPACE.md` bindings |
| **2. Episodic & Semantic Memory** | State history $H_{1:t} = \{(x_i, a_i, o_i)\}$ | *What events, decisions, and facts occurred in this project over time?* | Owned by the **project/workspace**; decays as code and architecture evolve | Session archive (`sessions.db` / `recall_entries`) |
| **3. Personal Imprint** | Operator policy & salience prior $\theta_P$ | *Given any state $x_t$, what do I **involuntarily attend to**, how do I **judge** it, and what is my **default move**?* | Owned by the **person**; portable across companies, domains, and tools | `~/.imprint/{IMPRINT.md, imprint.json, overlays/}` |

### The Falsification Test (Memory vs. Imprint)
To decide whether a candidate rule belongs in **Project Memory** or **Personal Imprint**, apply two tests:
1. **Peer Replacement Test (`peer_fork`):** If another strong engineer took over this exact repository tomorrow, would they need this fact to avoid breaking the system?
   - If **yes** (*"SQLite must be opened `?mode=ro` in `internal/corpus`"*), it is **Project Context / Project Memory**, not personal imprint.
   - If **no** (another strong engineer could legitimately prefer the opposite, such as in-place file overwrites instead of `-v2` drafts, or autonomous PR merges once CI is green), it is a candidate for **Personal Imprint**.
2. **Company Switch Test (`portability_check`):** If the person switched companies tomorrow, would this still describe how they think, argue, write, and delegate? If no, it is at most a **Workspace Binding (L2)** of a deeper portable principle, never L1 Imprint.

---

## 3. How Imprint and Memory Interact Without Conflation

Although Imprint ($\theta_P$) and Episodic Memory ($H_{1:t}$) are separate constructs, they connect through two well-defined interfaces:

```text
                       1. Distillation (Factor out project nouns & singletons)
Episodic Memory (H_1:t) -----------------------------------------------------> Personal Imprint (theta_P)
(sessions.db / recall)  <----------------------------------------------------- (~/.imprint/imprint.json)
                       2. Salience Conditioning (Write-filter & ranking lens)
```

### Interface 1: Episodic Memory $\xrightarrow{\text{distills}}$ Imprint (Habit Consolidation)
In human cognition, you do not retain verbatim memories of the 50 individual meetings where a muddled design failed. The individual episodes decay; what survives is the **cross-episode invariant**: the automatic habit of asking *"what is the underlying primitive?"* before accepting an interface.
- `imprint` reads raw episodes from `sessions.db`, strips project-specific nouns, and graduates a claim into `imprint.json` only when the same human intervention recurs across $\ge 3$ evidence items and $\ge 2$ distinct contexts.
- Once graduated, the habit replaces the need to retrieve those historical episodes.

### Interface 2: Imprint $\xrightarrow{\text{conditions}}$ Episodic Memory (The Salience Filter)
Two engineers observing the same codebase and meetings do not form the same memories. Memory is selective compression, and **the person's imprint is the write-filter**:
- A risk-averse engineer and an aggressive engineer, or a primitives-first systems thinker and a UI-first prototyper, notice and record different events as "important."
- When an episodic memory system (such as `agentsview recall`) extracts or ranks memories for a user, it should read `~/.imprint/imprint.json` as its **salience prior** so the memories it builds reflect what *that person* cares about (`F1`–`F6`), without `imprint` itself becoming the storage table for project facts.

---

## 4. Consequences

1. **Removed `.local/memories.json` from `imprint`:** `imprint` does not store or manage project-level episodic facts. It owns L1 (`IMPRINT.md`, `imprint.json`, `overlays/mode-*.md`) and the L2 workspace binding file (`.local/WORKSPACE.md`) that maps L1 rules to local tools.
2. **Read-only consumption of `sessions.db`:** `imprint` consumes `messages`, `tool_calls`, and `friction_findings` from `sessions.db` strictly to mine cross-project human invariants, leaving episodic search and recall storage inside `agentsview`.
