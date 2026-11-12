package digest

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/sunilgattupalle/imprint/internal/config"
	"github.com/sunilgattupalle/imprint/internal/corpus"
	"github.com/sunilgattupalle/imprint/internal/ingest"
)

type facetProbe struct {
	id             string
	title          string
	seedDimensions string
	re             *regexp.Regexp
}

// Person-agnostic conversational-act and dimension probes across the 5 seeded facets.
var facetProbes = []facetProbe{
	{
		id:             "F1_epistemic_posture",
		title:          "F1: Epistemic Posture (First Principles, Abstractions, Evidence & Verification Bar, Economics)",
		seedDimensions: "epistemic.abstraction_depth, epistemic.single_source_of_truth, epistemic.verification_bar, epistemic.unit_economics",
		re:             regexp.MustCompile(`(?i)\b(first principles|primitive|construct|source of truth|canonical|ontology|abstraction|schema|invariant|rubric|ground truth|inferred|observed|provenance|benchmark|baseline|metric|unit economics|cost|latency|throughput|verify|root cause|assumption)\b`),
	},
	{
		id:             "F2_dialectical_style",
		title:          "F2: Dialectical Style (Stress-Testing, Anti-Sycophancy, Mental Model Before Code, Trade-off Framing)",
		seedDimensions: "dialectic.challenge_threshold, dialectic.plan_before_code, dialectic.option_comparison, dialectic.lock_in_flip",
		re:             regexp.MustCompile(`(?i)\b(don'?t just agree|stress[- ]test|honest|what do you think|poke holes|devil'?s advocate|critique|disagree|conflating|mental model|before (?:you |we )?(?:implement|write|code)|let'?s (?:first )?(?:discuss|talk|align)|design first|trade-?offs?|options?|over-?complicat|simpler)\b`),
	},
	{
		id:             "F3_compression_and_structure",
		title:          "F3: Compression & Structure (Compression Asymmetry, Document Architecture, Revision Mechanics)",
		seedDimensions: "structure.compression_asymmetry, structure.opening_move, structure.progressive_depth, structure.revision_mechanics",
		re:             regexp.MustCompile(`(?i)\b(concise|terse|brief|too long|too verbose|too concise|compressed .* too much|fluff|bloat|over-?correct|don'?t overwrite|new file|version|-v\d|separate (?:doc|slide|file)|table|matrix|progressive|section|executive summary|one-?pager|deep dive)\b`),
	},
	{
		id:             "F4_voice_and_visuals",
		title:          "F4: Voice & Visual Signature (Prose Cadence, Banned Clichés, Palette Restraint, Diagram Medium)",
		seedDimensions: "voice.prose_cadence, voice.lexical_bans, visuals.palette_restraint, visuals.diagram_medium",
		re:             regexp.MustCompile(`(?i)\b(my tone|my voice|writing style|write it like|rephrase|rewrite|plainly|buzzword|cliche|em-?dash|opening line|contrast|too many colors|glaring|palette|dark mode|light mode|svg|ascii|mermaid|diagram|chart|layout|typography|font|visually)\b`),
	},
	{
		id:             "F5_delegation_and_risk",
		title:          "F5: Delegation & Risk Appetite (Autonomy Gates, Blast Radius, Git & Identity, Continuation Model)",
		seedDimensions: "delegation.blast_radius_gate, delegation.merge_ownership, delegation.git_and_identity, delegation.continuation_model",
		re:             regexp.MustCompile(`(?i)\b(autonom|permission|ask before|confirm before|sole approver|never merge|don'?t merge|don'?t push|don'?t commit|git pull|branch|worktree|author|identity|destructive|production|rollback|event[- ]driven|webhook|background|subagent|parallel)\b`),
	},
}

// Friction / correction detector (ported from agentsview/internal/friction/detect_corrections.go)
// to catch emergent user habits & corrections that don't match the 5 seeded facet probes.
var generalCorrectionRE = regexp.MustCompile(`(?i)(` +
	`^no[,.! ]|` +
	`\bdon'?t\b|\bdo not\b|\bnever\b|\balways\b|\bplease stop\b|` +
	`\bstop (?:doing|replying|posting|answering|sending|using|adding|changing|updating|overwriting)\b|` +
	`\bwrong\b|\bincorrect\b|\bnot (?:that|what|like that|right|the right)\b|` +
	`\bshould(?:n'?t| not| never)\b|\bthat'?s not\b|\binstead of\b|\bwhy did you\b|\bi prefer\b` +
	`)`)

// BuildDigest runs the corpus hygiene pipeline and generates `<instance>/.local/corpus/digest.md`
// combining S1 clustered session signals (seeded facets + emergent corrections), S2 authored docs
// (with Pre-AI Pure-Human Voice Gold separated), and S3 explicit priors.
func BuildDigest(instanceDir string) (string, *corpus.CorpusStats, error) {
	_, localDir, err := config.ResolveInstancePaths(instanceDir)
	if err != nil {
		return "", nil, err
	}
	cfg, err := config.LoadSources(localDir)
	if err != nil {
		return "", nil, err
	}

	stats, err := corpus.Build(instanceDir)
	if err != nil {
		return "", nil, err
	}

	var md strings.Builder
	md.WriteString(fmt.Sprintf("# Imprint Extraction Digest: %s (`%s`)\n\n", cfg.Owner.Name, cfg.Owner.ID))
	md.WriteString("> **Local-only synthesis input (`.local/corpus/digest.md`).** Read this file to synthesize `.local/evidence_ledger.json`, `.local/demoted_non_imprint.json`, `.local/WORKSPACE.md`, `IMPRINT.md`, and `overlays/`.\n\n")

	// 1. Corpus Summary
	md.WriteString("## 1. Corpus & Hygiene Summary (S1)\n\n")
	md.WriteString(fmt.Sprintf("- **Raw sessions in archive:** %d\n", stats.RawSessionsInDB))
	md.WriteString(fmt.Sprintf("- **Canonical deduplicated sessions:** %d (%d duplicate sessions merged/dropped)\n", stats.CanonicalSessionsAfterDedup, stats.DuplicateSessionsDropped))
	md.WriteString(fmt.Sprintf("- **Train / Holdout split:** %d train sessions (%d turns) / %d holdout sessions (%d turns)\n\n",
		stats.TrainSessions, stats.TrainTurns, stats.HoldoutSessions, stats.HoldoutTurns))

	// 2. S3 Explicit Priors
	md.WriteString("## 2. Curated Operating Rules & Style Priors (S3)\n\n")
	for _, p := range cfg.ExplicitPriors.Files {
		doc, err := ingest.ExtractDocument(p)
		if err != nil {
			continue
		}
		md.WriteString(fmt.Sprintf("### Prior: `%s`\n\n```markdown\n%s\n```\n\n", p, clipText(doc.Text, 3500)))
	}

	// 3. S2 Authored Ground-Truth Artifacts (Separated into 3A Pre-AI Gold vs. 3B General Authored)
	md.WriteString("## 3. Authored Ground-Truth Artifacts (S2)\n\n")
	seenPath := map[string]bool{}

	var preAIPaths []string
	for _, g := range cfg.AuthoredArtifacts.PreAIGlobs {
		matches, _ := filepath.Glob(config.ExpandPath(g))
		preAIPaths = append(preAIPaths, matches...)
	}
	sort.Strings(preAIPaths)
	if len(preAIPaths) > 0 {
		md.WriteString("### 3A. Pre-AI / Pure-Human Voice Gold (`--pre-ai`: Highest Priority for F4 Voice & Thought Identity)\n\n")
		for _, p := range preAIPaths {
			if seenPath[p] {
				continue
			}
			seenPath[p] = true
			doc, err := ingest.ExtractDocument(p)
			if err != nil {
				continue
			}
			md.WriteString(fmt.Sprintf("#### [PRE-AI GOLD] (`%s`): `%s`\n\n", doc.Format, filepath.Base(p)))
			md.WriteString(fmt.Sprintf("```text\n%s\n```\n\n", clipTextStratified(doc.Text, 3200)))
		}
	}

	var authoredPaths []string
	for _, g := range cfg.AuthoredArtifacts.Globs {
		matches, _ := filepath.Glob(config.ExpandPath(g))
		authoredPaths = append(authoredPaths, matches...)
	}
	localAuthored, _ := filepath.Glob(filepath.Join(localDir, "authored/*"))
	authoredPaths = append(authoredPaths, localAuthored...)
	sort.Strings(authoredPaths)

	for _, p := range authoredPaths {
		if seenPath[p] {
			continue
		}
		seenPath[p] = true
		doc, err := ingest.ExtractDocument(p)
		if err != nil {
			continue
		}
		md.WriteString(fmt.Sprintf("### Authored Artifact (`%s`): `%s`\n\n", doc.Format, filepath.Base(p)))
		if len(doc.CSSTokens) > 0 {
			md.WriteString("**Extracted CSS Tokens:** `" + strings.Join(doc.CSSTokens, "`, `") + "`\n\n")
		}
		md.WriteString(fmt.Sprintf("```text\n%s\n```\n\n", clipTextStratified(doc.Text, 2600)))
	}

	// 4. Clustered High-Signal Training Turns (S1)
	md.WriteString("## 4. High-Signal Training Turns by Seeded Facet + Emergent Clusters (S1, 80% Train Split Only)\n\n")
	trainPath := filepath.Join(localDir, "corpus/train_turns.jsonl")
	f, err := os.Open(trainPath)
	if err != nil {
		return "", stats, err
	}
	defer f.Close()

	byFacet := map[string][]corpus.TurnRecord{}
	var emergentTurns []corpus.TurnRecord
	seenTurn := map[string]bool{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		var rec corpus.TurnRecord
		if err := json.Unmarshal(sc.Bytes(), &rec); err != nil {
			continue
		}
		if rec.UserChars < 20 || rec.UserChars > 750 {
			continue
		}
		if strings.HasPrefix(rec.UserText, "{") {
			continue
		}
		dedupKey := rec.SessionID + "|" + rec.UserText
		if seenTurn[dedupKey] {
			continue
		}
		matchedSeed := false
		for _, fp := range facetProbes {
			if fp.re.MatchString(rec.UserText) {
				matchedSeed = true
				byFacet[fp.id] = append(byFacet[fp.id], rec)
			}
		}
		if matchedSeed {
			seenTurn[dedupKey] = true
		} else if generalCorrectionRE.MatchString(rec.UserText) && rec.UserChars <= 400 {
			seenTurn[dedupKey] = true
			emergentTurns = append(emergentTurns, rec)
		}
	}

	for _, fp := range facetProbes {
		turns := diversifyTurns(byFacet[fp.id], 35)
		md.WriteString(fmt.Sprintf("### %s (%d candidate turns)\n", fp.title, len(byFacet[fp.id])))
		md.WriteString(fmt.Sprintf("- **Seeded dimensions:** `%s`\n\n", fp.seedDimensions))
		writeTurnBullets(&md, turns)
		md.WriteString("\n")
	}

	// 4.6 Emergent / Unclustered Corrections & Directives
	emergentSample := diversifyTurns(emergentTurns, 30)
	md.WriteString(fmt.Sprintf("### FX: Emergent / Unclustered Corrections & Directives (%d candidate turns)\n", len(emergentTurns)))
	md.WriteString("- **Purpose:** High-signal user corrections and preferences not matched by the 5 seeded probes. Cluster recurring patterns here into either `F1`–`F5` or `F6_emergent` claims.\n\n")
	writeTurnBullets(&md, emergentSample)
	md.WriteString("\n")

	// 5. Multi-Signal Behavioral Evidence (Trajectory Pushbacks, Friction Findings & Version Diffs)
	frictions, trajectories, recalls := corpus.ExtractBehavioralSignals(instanceDir)
	allDocPaths := append(append([]string{}, preAIPaths...), authoredPaths...)
	versionDiffs := ExtractVersionDiffs(allDocPaths)

	if len(trajectories) > 0 || len(frictions) > 0 || len(versionDiffs) > 0 || len(recalls) > 0 {
		md.WriteString("## 5. Non-Verbal & Multi-Signal Behavioral Evidence\n\n")

		if len(trajectories) > 0 {
			md.WriteString(fmt.Sprintf("### 5A. Tool-Call Trajectory Pushbacks (`S1-Trajectory`, %d events)\n", len(trajectories)))
			md.WriteString("- **What this shows:** Turns where the assistant invoked `write`/`edit` and the user immediately stepped in to stop, redirect, or enforce design alignment first.\n\n")
			limit := 15
			if len(trajectories) < limit {
				limit = len(trajectories)
			}
			for i := 0; i < limit; i++ {
				tr := trajectories[i]
				date := tr.Timestamp
				if len(date) >= 10 {
					date = date[:10]
				}
				oneLine := strings.ReplaceAll(tr.UserReaction, "\n", " ")
				target := tr.ToolName
				if tr.FilePath != "" {
					target += " -> " + filepath.Base(tr.FilePath)
				}
				md.WriteString(fmt.Sprintf("- `[%s | %s | %s | #%d]` (%s) after `%s`: %q\n",
					date, tr.Agent, tr.Project, tr.Ordinal, tr.SessionID, target, oneLine))
			}
			md.WriteString("\n")
		}

		if len(versionDiffs) > 0 {
			md.WriteString(fmt.Sprintf("### 5B. Versioned Document Revision Diffs (`S2-Diff`, %d pairs)\n", len(versionDiffs)))
			md.WriteString("- **What this shows:** Concrete structural, word-count, and punctuation deltas across sequential `-v1` -> `-v2` drafts.\n\n")
			for _, d := range versionDiffs {
				md.WriteString(fmt.Sprintf("#### `%s` -> `%s` (words: %d -> %d, em-dashes: %d -> %d)\n",
					d.FromFile, d.ToFile, d.BeforeWords, d.AfterWords, d.BeforeEmDashes, d.AfterEmDashes))
				for _, h := range d.RemovedHeadings {
					md.WriteString(fmt.Sprintf("- *Removed heading:* `%s`\n", h))
				}
				for _, h := range d.AddedHeadings {
					md.WriteString(fmt.Sprintf("- *Added heading:* `%s`\n", h))
				}
				for _, r := range d.SampleRemoved {
					md.WriteString(fmt.Sprintf("- *Cut line:* %q\n", r))
				}
				for _, a := range d.SampleAdded {
					md.WriteString(fmt.Sprintf("- *Added line:* %q\n", a))
				}
				md.WriteString("\n")
			}
		}

		if len(frictions) > 0 {
			md.WriteString(fmt.Sprintf("### 5C. Indexed Friction Findings (`sessions.db`, %d events)\n\n", len(frictions)))
			limit := 15
			if len(frictions) < limit {
				limit = len(frictions)
			}
			for i := 0; i < limit; i++ {
				fr := frictions[i]
				date := fr.OccurredAt
				if len(date) >= 10 {
					date = date[:10]
				}
				oneLine := strings.ReplaceAll(fr.Text, "\n", " ")
				md.WriteString(fmt.Sprintf("- `[%s | %s | %s | %s]` (%s): %q\n",
					date, fr.Agent, fr.Project, fr.Kind, fr.SessionID, oneLine))
			}
			md.WriteString("\n")
		}

		if len(recalls) > 0 {
			md.WriteString(fmt.Sprintf("### 5D. Accepted Episodic Recall Entries (`sessions.db`, %d entries)\n\n", len(recalls)))
			limit := 15
			if len(recalls) < limit {
				limit = len(recalls)
			}
			for i := 0; i < limit; i++ {
				rc := recalls[i]
				md.WriteString(fmt.Sprintf("- `[%s | scope=%s | project=%s | transferable=%v]` **%s**: %s\n",
					rc.Type, rc.Scope, rc.Project, rc.Transferable, rc.Title, rc.Body))
			}
			md.WriteString("\n")
		}
	}

	digestPath := filepath.Join(localDir, "corpus/digest.md")
	if err := os.WriteFile(digestPath, []byte(md.String()), 0o644); err != nil {
		return "", stats, err
	}
	return digestPath, stats, nil
}

func writeTurnBullets(md *strings.Builder, turns []corpus.TurnRecord) {
	for _, t := range turns {
		date := t.Timestamp
		if len(date) >= 10 {
			date = date[:10]
		}
		oneLine := strings.ReplaceAll(t.UserText, "\n", " ")
		if t.PrevAssistantSnippet != "" {
			prev := strings.ReplaceAll(t.PrevAssistantSnippet, "\n", " ")
			if len(prev) > 140 {
				prev = prev[:140] + "..."
			}
			md.WriteString(fmt.Sprintf("- `[%s | %s | %s | #%d]` (%s)\n  - *Context:* %q\n  - *User:* %q\n",
				date, t.Agent, t.Project, t.Ordinal, t.SessionID, prev, oneLine))
		} else {
			md.WriteString(fmt.Sprintf("- `[%s | %s | %s | #%d]` (%s): %q\n",
				date, t.Agent, t.Project, t.Ordinal, t.SessionID, oneLine))
		}
	}
}

// diversifyTurns selects up to limit turns round-robin across distinct projects and sessions
// so a single long session does not crowd out evidence from other projects.
func diversifyTurns(turns []corpus.TurnRecord, limit int) []corpus.TurnRecord {
	if len(turns) <= limit {
		return turns
	}
	byProject := map[string][]corpus.TurnRecord{}
	var projects []string
	for _, t := range turns {
		p := t.Project
		if p == "" {
			p = "unassigned"
		}
		if _, exists := byProject[p]; !exists {
			projects = append(projects, p)
		}
		byProject[p] = append(byProject[p], t)
	}
	sort.Strings(projects)

	var out []corpus.TurnRecord
	idx := 0
	for len(out) < limit {
		added := false
		for _, p := range projects {
			bucket := byProject[p]
			if idx < len(bucket) {
				out = append(out, bucket[idx])
				added = true
				if len(out) >= limit {
					break
				}
			}
		}
		if !added {
			break
		}
		idx++
	}
	return out
}

func clipText(s string, maxChars int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxChars {
		return s
	}
	return s[:maxChars] + "\n... [truncated for digest]"
}

// clipTextStratified preserves the opening, middle, and closing sections of a long document
// so structural endings, conclusions, and slide summaries are not lost to head-only truncation.
func clipTextStratified(s string, maxChars int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxChars {
		return s
	}
	headLen := (maxChars * 6) / 10
	tailLen := maxChars - headLen
	return s[:headLen] + "\n\n... [middle section elided for digest] ...\n\n" + s[len(s)-tailLen:]
}
