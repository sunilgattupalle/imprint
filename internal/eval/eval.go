package eval

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sunilgattupalle/imprint-model/internal/config"
	"github.com/sunilgattupalle/imprint-model/internal/corpus"
	"github.com/sunilgattupalle/imprint-model/internal/gate"
)

var signalRE = regexp.MustCompile(`(?i)(` +
	`\bdon't just\b|\bdon't make it\b|\bdon't overwrite\b|\bdo not use\b|\bstop updating\b|\bwhat is wrong\b|\bwhat the hell\b|` +
	`\bconcise\b|\bprecise\b|\bno ai fluff\b|\bover-?complicat|\bover-?correct|\btoo concise\b|\btoo big\b|` +
	`\bmental model\b|\bprimitive\b|\bconstruct\b|\bfirst principles\b|\bsource of truth\b|\brubric\b|\bground truth\b|\bindependent doc\b|` +
	`\bhonest\b|\bstress test\b|\bwhat do you think\b|\bask opus\b|\bopus 5\b|\bam i conflating\b|\breview this thoroughly\b|` +
	`\bwrite -?v\d\b|\bevolution till v\d\b|\bseparate slide\b|` +
	`\buse svg\b|\bfix all the ascii\b|\buse ascii diagrams\b|\bvisually\b|\bvisual report\b|\btoo many colors\b|\btoo much text\b|` +
	`\bautonomy\b|\buse my identity\b|\bworkload identity\b|\btrust .*-issued identity\b|\bdispatches an event\b|\bcome back to resume\b|` +
	`\buse my writing style\b|\brephrase the whole thing\b|\bopening line\b` +
	`)`)

type EvalResults struct {
	Instance                  string         `json:"instance"`
	HoldoutSignalTurns        int            `json:"holdout_signal_turns"`
	CoveredByL1Portable       int            `json:"covered_by_L1_portable_imprint"`
	L1CoverageRate            float64        `json:"L1_coverage_rate"`
	CoveredByL1OrL2L3         int            `json:"covered_by_L1_or_L2_L3"`
	TotalCoverageRate         float64        `json:"total_coverage_rate"`
	HitsPerL1Claim            map[string]int `json:"hits_per_L1_claim"`
	HitsPerDemotedRule        map[string]int `json:"hits_per_demoted_rule"`
}

// RunHoldoutEval evaluates L1 portable claims against untouched holdout_turns.jsonl.
func RunHoldoutEval(instanceDir string) (*EvalResults, error) {
	absInstance, localDir, err := config.ResolveInstancePaths(instanceDir)
	if err != nil {
		return nil, err
	}
	instanceName := filepath.Base(absInstance)

	claimsBytes, err := os.ReadFile(filepath.Join(localDir, "evidence_ledger.json"))
	if err != nil {
		return nil, err
	}
	var claims []gate.RawClaim
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, err
	}

	demotedBytes, _ := os.ReadFile(filepath.Join(localDir, "demoted_non_imprint.json"))
	var demoted []gate.DemotedRule
	_ = json.Unmarshal(demotedBytes, &demoted)

	holdoutFile, err := os.Open(filepath.Join(localDir, "corpus/holdout_turns.jsonl"))
	if err != nil {
		return nil, err
	}
	defer holdoutFile.Close()

	seen := map[string]bool{}
	var signalTurns []corpus.TurnRecord
	sc := bufio.NewScanner(holdoutFile)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		var r corpus.TurnRecord
		if err := json.Unmarshal(sc.Bytes(), &r); err != nil {
			continue
		}
		txt := strings.TrimSpace(r.UserText)
		if len(txt) < 20 || len(txt) > 1200 {
			continue
		}
		if strings.HasPrefix(txt, "{") || strings.Contains(txt, `"commits": [`) {
			continue
		}
		if !signalRE.MatchString(txt) {
			continue
		}
		key := r.SessionID + "|" + txt
		if seen[key] {
			continue
		}
		seen[key] = true
		signalTurns = append(signalTurns, r)
	}

	type matcher struct {
		id string
		re *regexp.Regexp
	}
	var l1Matchers []matcher
	l1Hits := map[string][]corpus.TurnRecord{}
	for _, c := range claims {
		l1Hits[c.ID] = nil
		if c.EvalMatcher != "" {
			if re, err := regexp.Compile("(?i)" + c.EvalMatcher); err == nil {
				l1Matchers = append(l1Matchers, matcher{c.ID, re})
			}
		}
	}

	var demMatchers []matcher
	demHits := map[string][]corpus.TurnRecord{}
	for _, d := range demoted {
		demHits[d.ID] = nil
		if d.EvalMatcher != "" {
			if re, err := regexp.Compile("(?i)" + d.EvalMatcher); err == nil {
				demMatchers = append(demMatchers, matcher{d.ID, re})
			}
		}
	}

	coveredL1 := 0
	coveredAny := 0
	for _, r := range signalTurns {
		matchedL1 := false
		for _, m := range l1Matchers {
			if m.re.MatchString(r.UserText) {
				matchedL1 = true
				l1Hits[m.id] = append(l1Hits[m.id], r)
			}
		}
		matchedDem := false
		for _, m := range demMatchers {
			if m.re.MatchString(r.UserText) {
				matchedDem = true
				demHits[m.id] = append(demHits[m.id], r)
			}
		}
		if matchedL1 {
			coveredL1++
		}
		if matchedL1 || matchedDem {
			coveredAny++
		}
	}

	total := len(signalTurns)
	l1Rate := 0.0
	anyRate := 0.0
	if total > 0 {
		l1Rate = float64(coveredL1) / float64(total)
		anyRate = float64(coveredAny) / float64(total)
	}

	res := &EvalResults{
		Instance:            instanceName,
		HoldoutSignalTurns:  total,
		CoveredByL1Portable: coveredL1,
		L1CoverageRate:      float64(int(l1Rate*1000+0.5)) / 1000,
		CoveredByL1OrL2L3:   coveredAny,
		TotalCoverageRate:   float64(int(anyRate*1000+0.5)) / 1000,
		HitsPerL1Claim:      map[string]int{},
		HitsPerDemotedRule:  map[string]int{},
	}
	for k, v := range l1Hits {
		res.HitsPerL1Claim[k] = len(v)
	}
	for k, v := range demHits {
		res.HitsPerDemotedRule[k] = len(v)
	}

	evalDir := filepath.Join(localDir, "eval")
	_ = os.MkdirAll(evalDir, 0o755)
	resBytes, _ := json.MarshalIndent(res, "", "  ")
	_ = os.WriteFile(filepath.Join(evalDir, "holdout_results.json"), append(resBytes, '\n'), 0o644)

	var md strings.Builder
	md.WriteString(fmt.Sprintf("# Held-Out Evaluation Report (`%s/.local/eval`)\n\n", instanceName))
	md.WriteString(fmt.Sprintf("- **High-signal directive/corrective turns in holdout (excluding JSON pastes):** %d\n", total))
	md.WriteString(fmt.Sprintf("- **Covered by L1 Portable Imprint (`evidence_ledger.json` -> `imprint.json`):** %d / %d (**%.1f%%**)\n", coveredL1, total, 100*l1Rate))
	md.WriteString(fmt.Sprintf("- **Covered including separated L2 Workspace & L3 Agent Baseline (`demoted_non_imprint.json`):** %d / %d (**%.1f%%**)\n\n", coveredAny, total, 100*anyRate))
	md.WriteString("## L1 Portable Imprint Claims — Holdout Verification\n\n")
	md.WriteString("| Claim ID | Facet | Portable Claim | Holdout Hits | Sample Held-Out Evidence |\n")
	md.WriteString("|---|---|---|---:|---|\n")
	for _, c := range claims {
		hits := l1Hits[c.ID]
		sample := "—"
		if len(hits) > 0 {
			ts := hits[0].Timestamp
			if len(ts) >= 10 {
				ts = ts[:10]
			}
			snippet := strings.ReplaceAll(hits[0].UserText, "\n", " ")
			if len(snippet) > 105 {
				snippet = snippet[:105]
			}
			sample = fmt.Sprintf("`%s` (%s): *%q...*", ts, hits[0].Agent, snippet)
		}
		md.WriteString(fmt.Sprintf("| `%s` | `%s` | %s | %d | %s |\n", c.ID, c.Facet, c.Title, len(hits), sample))
	}
	md.WriteString("\n## Separated L2 Workspace & L3 Agent Baseline Rules\n\n")
	md.WriteString("| ID | Layer | Rule Summary | Why Excluded from L1 Imprint | Holdout Hits |\n")
	md.WriteString("|---|---|---|---|---:|\n")
	for _, d := range demoted {
		md.WriteString(fmt.Sprintf("| `%s` | `%s` | %s | %s | %d |\n",
			d.ID, d.Layer, d.Title, d.ReasonExcludedFromImprint, len(demHits[d.ID])))
	}
	_ = os.WriteFile(filepath.Join(evalDir, "holdout_report.md"), []byte(md.String()), 0o644)

	fmt.Printf("Holdout L1 portable coverage: %d/%d (%.1f%%)\n", coveredL1, total, 100*l1Rate)
	fmt.Printf("Holdout L1+L2+L3 coverage:    %d/%d (%.1f%%)\n", coveredAny, total, 100*anyRate)
	return res, nil
}
