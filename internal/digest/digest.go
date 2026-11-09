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

	"github.com/sunilgattupalle/imprint-model/internal/config"
	"github.com/sunilgattupalle/imprint-model/internal/corpus"
	"github.com/sunilgattupalle/imprint-model/internal/ingest"
)

type facetprobe struct {
	id    string
	title string
	re    *regexp.Regexp
}

var facetProbes = []facetprobe{
	{
		"F1_epistemic_posture",
		"F1 — Epistemic Posture (Primitives, Single Source of Truth, Provenance, Rubrics, Economics)",
		regexp.MustCompile(`(?i)\b(primitive|construct|first principles|source of truth|ontology|abstraction|rubric|ground truth|inferred|observed|provenance|independent doc|tok/s|cohorts?)\b`),
	},
	{
		"F2_dialectical_style",
		"F2 — Dialectical Style (Stress-Testing, Anti-Sycophancy, Mental Model Before Code, Cross-Model Critique)",
		regexp.MustCompile(`(?i)\b(don't just agree|stress test|honest|what do you think|am i conflating|mental model|let's first talk|before you implement|ask opus|opus 5|overcomplicat)\b`),
	},
	{
		"F3_compression_and_structure",
		"F3 — Compression & Structure (Compression Asymmetry, Companion Docs, Progressive Depth, -v2 Iteration)",
		regexp.MustCompile(`(?i)\b(concise|no ai fluff|too concise|compressed this too much|over-?correct|write -?v\d|don't overwrite|separate slide|no timelines|different tracks|what's next)\b`),
	},
	{
		"F4_voice_and_visuals",
		"F4 — Voice & Visual Signature (CTO Prose, Banned Phrases, Color Restraint, SVG vs. ASCII)",
		regexp.MustCompile(`(?i)\b(load-bearing|the hypothesis|write it like i would|my tone|my writing style|too many colors|glaring colors|use svg|fix all the ascii|use ascii diagrams|visually)\b`),
	},
	{
		"F5_delegation_and_risk",
		"F5 — Delegation & Risk Appetite (Governed Autonomy, Author Identity, Git Pull Main, Event-Driven Loops)",
		regexp.MustCompile(`(?i)\b(autonomy|git pull on main|use my identity|sole approver|event driven|webhook|dispatches an event|come back to resume)\b`),
	},
}

// BuildDigest runs the corpus hygiene pipeline and generates `<instance>/.local/corpus/digest.md`
// (~15-25k tokens) combining S1 clustered session signals, S2 authored docs, and S3 explicit priors.
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
	md.WriteString(fmt.Sprintf("# Imprint Extraction Digest — %s (`%s`)\n\n", cfg.Owner.Name, cfg.Owner.ID))
	md.WriteString("> **Local-only synthesis input (`.local/corpus/digest.md`).** Read this file to synthesize `.local/evidence_ledger.json`, `.local/demoted_non_imprint.json`, `IMPRINT.md`, and `overlays/`.\n\n")

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

	// 3. S2 Authored Ground-Truth Artifacts
	md.WriteString("## 3. Authored Ground-Truth Artifacts (S2)\n\n")
	var authoredPaths []string
	for _, g := range cfg.AuthoredArtifacts.Globs {
		matches, _ := filepath.Glob(config.ExpandPath(g))
		authoredPaths = append(authoredPaths, matches...)
	}
	localAuthored, _ := filepath.Glob(filepath.Join(localDir, "authored/*"))
	authoredPaths = append(authoredPaths, localAuthored...)
	sort.Strings(authoredPaths)

	seenPath := map[string]bool{}
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
		md.WriteString(fmt.Sprintf("```text\n%s\n```\n\n", clipText(doc.Text, 2200)))
	}

	// 4. Clustered High-Signal Training Turns (S1)
	md.WriteString("## 4. High-Signal Training Turns Clustered by Facet (S1 — 80% Train Split Only)\n\n")
	trainPath := filepath.Join(localDir, "corpus/train_turns.jsonl")
	f, err := os.Open(trainPath)
	if err != nil {
		return "", stats, err
	}
	defer f.Close()

	byFacet := map[string][]corpus.TurnRecord{}
	seenTurn := map[string]bool{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		var rec corpus.TurnRecord
		if err := json.Unmarshal(sc.Bytes(), &rec); err != nil {
			continue
		}
		if rec.UserChars < 20 || rec.UserChars > 650 {
			continue
		}
		if strings.HasPrefix(rec.UserText, "{") {
			continue
		}
		dedupKey := rec.SessionID + "|" + rec.UserText
		if seenTurn[dedupKey] {
			continue
		}
		for _, fp := range facetProbes {
			if fp.re.MatchString(rec.UserText) {
				seenTurn[dedupKey] = true
				byFacet[fp.id] = append(byFacet[fp.id], rec)
			}
		}
	}

	for _, fp := range facetProbes {
		turns := byFacet[fp.id]
		md.WriteString(fmt.Sprintf("### %s (%d candidate turns)\n\n", fp.title, len(turns)))
		limit := 35
		if len(turns) < limit {
			limit = len(turns)
		}
		for i := 0; i < limit; i++ {
			t := turns[i]
			date := t.Timestamp
			if len(date) >= 10 {
				date = date[:10]
			}
			oneLine := strings.ReplaceAll(t.UserText, "\n", " ")
			md.WriteString(fmt.Sprintf("- `[%s | %s | %s | #%d]` (%s): %q\n",
				date, t.Agent, t.Project, t.Ordinal, t.SessionID, oneLine))
		}
		md.WriteString("\n")
	}

	digestPath := filepath.Join(localDir, "corpus/digest.md")
	if err := os.WriteFile(digestPath, []byte(md.String()), 0o644); err != nil {
		return "", stats, err
	}
	return digestPath, stats, nil
}

func clipText(s string, maxChars int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxChars {
		return s
	}
	return s[:maxChars] + "\n... [truncated for digest]"
}
