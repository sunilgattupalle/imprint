# Imprint (`imprint`)

> Blueprint, Go CLI, and Agent Skill to extract a portable imprint of a person and carry their judgment, taste, and voice across every AI tool they use.

---

## 1. Purpose and Core Hypothesis

Most context engineering focuses on the organization or the repository. Even personal "memory" systems miss the person: they take raw session transcripts and distill flat episodic facts, treating every turn as a neutral observation. That strips out how the person actually interacted, what risks they refused to take, where they demanded depth, and why they pushed back.

And when people try to write their own `CLAUDE.md` or `AGENTS.md` by hand, three things go wrong:
1. **Abstract platitudes lose against base-model RLHF reflexes.** Writing *"stress-test my ideas, don't flatter me"* in a prompt file fails the moment you ask a short skeptical question (*"is it really X?"*), because the base model's strongest RLHF prior is to treat a skeptical question as a cue to apologize and reverse course. Rules only hold under pressure when they state the **observable trigger cue**, the **specific base-model failure mode**, the **required counter-move**, and the **flip boundary**.
2. **Three layers get conflated into one file.** Hand-written instruction files mix **who the person is (L1)**, **company-specific repo conventions (L2)**, and **generic workarounds for LLM bugs (L3)**. That file cannot leave the repo it was written in.
3. **AI-coauthored text slowly erodes identity.** Articles and specs drafted with LLMs carry a subtle drift. Read them three months later and you no longer recognize the thought as yours. Without anchoring on pre-AI human writing and in-session rewrites, style profiles just feed AI-tainted prose back into the model.

**The hypothesis:** Operationalized for an AI agent, a person's imprint is the set of situational deltas $(C, F_{\text{base}}) \rightarrow (M, \neg C)$ where the base model's default behavior $F_{\text{base}}$ diverges from the person's judgment, taste, or governance boundary $M$. If we mine those deltas from verbal pushbacks, tool-call trajectory interruptions, and draft-to-final file diffs across all coding agents, anchor voice on pre-AI writing, and separate portable judgment from workspace trivia, the resulting imprint can travel with the person across any tool.

---

## 2. Mental Model

```text
Local Sources (Read-Only)                 Deterministic Engine (imprint CLI)             Portable Output (~/.imprint)
-------------------------                 ----------------------------------             ----------------------------
S1. Agent Sessions                        1. Native + AgentsView Ingest & Dedup          IMPRINT.md (~500-650 tok)
    Claude, Cursor, Pi, OMP,      --->    2. Authorship Hygiene (strip non-human)  --->  overlays/mode-*.md (situational)
    Auggie, OpenCode, Devin               3. Seeded + Emergent Dimension Mining          imprint.json (quote-free L1)
S2A. Pre-AI Human Writing (--pre-ai)      4. Gate 1 (Portability) & Gate 2 (Fork)        .local/WORKSPACE.md (L2 local)
S2B. Specs, Decks, HTML (--authored)      5. 20% Holdout Eval & Harness Auto-Link        .local/evidence_ledger.json
S3. Explicit Priors (CLAUDE.md, etc.)
```

### Three Layers (Only L1 Is the Imprint)

| Layer | Definition | Destination |
|---|---|---|
| **L1. Personal Imprint** | Portable cognitive, dialectical, structural, stylistic, and delegation dimensions of **the person** across any workspace | `~/.imprint/IMPRINT.md`, `imprint.json`, `overlays/` (tracked) |
| **L2. Workspace Bindings** | How that imprint binds to a **specific company, stack, or repo** | `~/.imprint/.local/WORKSPACE.md` (gitignored) |
| **L3. Agent Baseline** | Universal competence **every** engineer expects (e.g., verify before claiming done) | `~/.imprint/.local/demoted_non_imprint.json` (excluded from imprint) |

### Two Admission Gates and the 4-Part Claim Contract

A candidate rule graduates into **L1** only if it passes both admission gates and specifies its 4-part situational contract:

1. **Portability Gate (`portability_check`):** Still holds if the person switches companies or works on an open-source repo tomorrow. Zero company or workspace proper nouns may appear in any exported field.
2. **Non-Trivial Fork Gate (`peer_fork`):** Another strong, reasonable engineer would legitimately prefer the opposite. If no reasonable peer would prefer the opposite, it is **L3 Agent Baseline**, not personal imprint.
3. **The 4-Part Situational Contract (`applies_when` $\rightarrow$ `default_model_failure` $\rightarrow$ `portable_rule` $\rightarrow$ `flips_when`):**
   - `applies_when`: The observable conversational or task cue that activates the rule.
   - `default_model_failure`: What an un-imprinted base model does wrong in this situation (e.g., capitulating and flip-flopping when probed with a skeptical question, or creating unrequested `.md` files when asked a question).
   - `portable_rule`: The required counter-move.
   - `flips_when`: The explicit boundary condition where the person intentionally suspends or inverts the rule.

### Seeded Facets and Emergent Dimensions

`imprint` starts with five seeded facets (`F1`–`F5`) and an open emergent bucket (`F6_emergent`) so recurring interventions outside the seed list are discovered from unclustered corrections rather than forced into fixed boxes:

| Facet | Seeded Dimensions | Question it answers |
|---|---|---|
| **`F1_epistemic_posture`** | `epistemic.abstraction_depth`, `epistemic.single_source_of_truth`, `epistemic.verification_bar`, `epistemic.unit_economics` | How does this person decide what is true, well-abstracted, and properly evaluated? |
| **`F2_dialectical_style`** | `dialectic.challenge_threshold`, `dialectic.plan_before_code`, `dialectic.option_comparison`, `dialectic.lock_in_flip` | When should the agent challenge vs. execute, and how are disagreements resolved? |
| **`F3_compression_and_structure`** | `structure.compression_asymmetry`, `structure.opening_move`, `structure.progressive_depth`, `structure.revision_mechanics` | Where does this person want extreme compression vs. full analytical depth, and how are documents structured? |
| **`F4_voice_and_visuals`** | `voice.prose_cadence`, `voice.lexical_bans`, `visuals.palette_restraint`, `visuals.diagram_medium` | What sentence cadence (anchored on `--pre-ai` human writing), vocabulary, visual hierarchy, and diagram medium match their hand? |
| **`F5_delegation_and_risk`** | `delegation.blast_radius_gate`, `delegation.merge_ownership`, `delegation.git_and_identity`, `delegation.continuation_model` | Which actions are autonomous, which require human sign-off, and how are long-running tasks coordinated? |
| **`F6_emergent`** | `emergent.<custom_dimension>` | What recurring situational corrections or preferences unique to this builder emerged outside `F1`–`F5`? |

### Two-Zone Privacy Boundary

Raw session transcripts, verbatim quote ledgers, session UUIDs, and company bindings stay strictly inside `.local/` (gitignored). The tracked root contains only portable rules and aggregate provenance counts:

```text
~/.imprint/
├── .gitignore                 # Ignores .local/
├── IMPRINT.md                 # Always-on portable imprint (~500-650 tokens)
├── imprint.json               # Validated L1 claims (aggregate counts only, zero quotes/UUIDs)
├── overlays/                  # Situational depth loaded when the task matches
│   ├── mode-architecture-and-specs.md
│   ├── mode-exec-memos-and-decks.md
│   ├── mode-visual-html-and-prototypes.md
│   ├── mode-coding-and-prs.md
│   └── mode-autonomous-orchestration.md
└── .local/                    # Gitignored: never leaves the machine
    ├── sources.yaml           # Pre-AI globs, authored globs, disallowed workspace terms
    ├── authored/              # Remote docs/decks fetched via MCP
    ├── corpus/                # Deduplicated train/holdout JSONL + synthesis digest.md
    ├── evidence_ledger.json   # Full claims with verbatim quotes & session UUIDs
    ├── demoted_non_imprint.json
    ├── WORKSPACE.md           # L2 company/repo bindings
    └── eval/                  # 20% holdout evaluation report
```

---

## 3. What You Do With an Imprint

Install and run `imprint` once; it defaults to `~/.imprint` and links your imprint into your local tools automatically:

```bash
go install github.com/sunilgattupalle/imprint/cmd/imprint@latest
```

```bash
# 1. Sync local agent sessions, scaffold ~/.imprint, build .local/corpus/digest.md, and link harnesses
imprint sync

# 2. Register pre-AI human writings, authored docs/decks, and disallowed company terms
imprint add \
  --pre-ai "~/writings/pre-2023-*.md" \
  --authored "~/docs/*.md" \
  --authored "~/decks/*.pptx" \
  --disallow-term "CompanyName"

# 3. Your agent (via SKILL.md) synthesizes claims and writes:
#    .local/evidence_ledger.json, .local/WORKSPACE.md, IMPRINT.md, and overlays/

# 4. Validate Gate 1 & Gate 2, export quote-free imprint.json, auto-link harnesses, and run holdout blind-spot eval
imprint validate
imprint eval

# 5. Compile a situational operating brief for a specific prompt and working directory
imprint brief --cwd . --prompt "write the architecture spec" --files "README.md"
```

### How It Wires Into Your Tools

Running `imprint sync`, `imprint validate`, or `imprint link` writes an idempotent `@`-include block for `~/.imprint/IMPRINT.md` and `~/.imprint/.local/WORKSPACE.md` into every detected agent harness on your machine:

- **Claude Code CLI & Claude Desktop (Code / Cowork):** Auto-linked in `~/.claude/CLAUDE.md`.
- **Pi (`pi`) & Oh My Pi (`omp`):** Auto-linked in `~/.pi/agent/AGENTS.md` and `~/.omp/agent/AGENTS.md`.
- **OpenCode (`opencode`):** Auto-linked in `~/.config/opencode/AGENTS.md`.
- **Hermes (`hermes`):** Auto-linked in `~/.hermes/SOUL.md`.
- **Harness (`auggie`) & Cursor:** Reference `~/.imprint/IMPRINT.md` in your global user rules or workspace `AGENTS.md` / `.cursor/rules`.
- **Episodic Memory Systems:** Read `~/.imprint/imprint.json` so memory extraction and retrieval are conditioned on the person's active dimensions (`facet`, `dimension`, `applies_when`, `default_model_failure`, `flips_when`) rather than flat transcript summaries.

---

## 4. Capability Tracks

Future work is organized into three capability tracks:
1. **Track 1: Embedded Session Indexing (`sunilgattupalle/agentsview` integration).** Subsume session collection and incremental watching directly under `imprint` so multi-agent transcript indexing, native `auggie` parsing, and friction signal detection run without a separate daemon setup.
2. **Track 2: Paired Behavioral Replay on Held-Out Failure Pairs.** Use `.local/eval/holdout_pairs.jsonl` (`PrevAssistantSnippet -> UserIntervention`) to run automated A/B replay tests verifying that responses generated with `imprint brief` eliminate the user's turn $t+1$ correction.
3. **Track 3: Imprint-Conditioned Salience Filter for Episodic Memory.** Pass `imprint.json` as the write-filter into `agentsview recall` so project memories reflect what the person actually cares about.

---

## 5. Repository Layout

- [`SKILL.md`](./SKILL.md): Agent skill for building, refreshing, and validating an imprint in `~/.imprint`.
- [`AGENTS.md`](./AGENTS.md): Operating instructions and non-negotiable invariants for agents working in this repo.
- [`adrs/`](./adrs): Architecture Decision Records (`ADR-0001` through `ADR-0004`).
- [`cmd/imprint/`](./cmd/imprint) & [`internal/`](./internal): Deterministic Go CLI packages (`config`, `collector`, `ingest`, `corpus`, `digest`, `gate`, `eval`, `link`, `brief`).
- [`schemas/`](./schemas): JSON Schemas for local evidence claims (`claim.schema.json`) and exported quote-free imprints (`imprint.schema.json`).
- [`docs/`](./docs): Method and architecture specifications (`01-sources.md` through `06-architecture-spec.md`).
