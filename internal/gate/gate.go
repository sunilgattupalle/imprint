package gate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/sunilgattupalle/imprint-model/internal/config"
)

type AdmissionTests struct {
	PortabilityCheck string `json:"portability_check"`
	PeerFork         string `json:"peer_fork"`
}

type EvidenceItem struct {
	SourceType string `json:"source_type"`
	SourceID   string `json:"source_id"`
	Project    string `json:"project"`
	Timestamp  string `json:"timestamp"`
	Ordinal    int    `json:"ordinal,omitempty"`
	Quote      string `json:"quote"`
}

type CounterEvidenceItem struct {
	SourceID    string `json:"source_id"`
	Observation string `json:"observation"`
}

type RawClaim struct {
	ID                string                `json:"id"`
	Facet             string                `json:"facet"`
	Scope             string                `json:"scope"`
	Status            string                `json:"status"`
	Title             string                `json:"title"`
	PortableRule      string                `json:"portable_rule"`
	AdmissionTests    AdmissionTests        `json:"admission_tests"`
	AppliesWhen       string                `json:"applies_when"`
	FlipsWhen         string                `json:"flips_when"`
	WorkspaceBindings map[string]string     `json:"workspace_bindings,omitempty"`
	EvalMatcher       string                `json:"eval_matcher,omitempty"`
	Evidence          []EvidenceItem        `json:"evidence"`
	CounterEvidence   []CounterEvidenceItem `json:"counter_evidence"`
}

type DemotedRule struct {
	ID                        string `json:"id"`
	Layer                     string `json:"layer"`
	Title                     string `json:"title"`
	ReasonExcludedFromImprint string `json:"reason_excluded_from_imprint"`
	EvalMatcher               string `json:"eval_matcher,omitempty"`
}

type ProvenanceCounts struct {
	EvidenceItems        int `json:"evidence_items"`
	DistinctSourceTypes  int `json:"distinct_source_types"`
	DistinctProjects     int `json:"distinct_projects"`
	CounterEvidenceItems int `json:"counter_evidence_items"`
}

type PublicClaim struct {
	ID               string           `json:"id"`
	Facet            string           `json:"facet"`
	Scope            string           `json:"scope"`
	Status           string           `json:"status"`
	Title            string           `json:"title"`
	PortableRule     string           `json:"portable_rule"`
	AdmissionTests   AdmissionTests   `json:"admission_tests"`
	AppliesWhen      string           `json:"applies_when"`
	FlipsWhen        string           `json:"flips_when"`
	ProvenanceCounts ProvenanceCounts `json:"provenance_counts"`
}

type PublicImprint struct {
	SchemaVersion string        `json:"schema_version"`
	GeneratedAt   string        `json:"generated_at"`
	Claims        []PublicClaim `json:"claims"`
}

var validFacets = map[string]bool{
	"F1_epistemic_posture":         true,
	"F2_dialectical_style":         true,
	"F3_compression_and_structure": true,
	"F4_voice_and_visuals":         true,
	"F5_delegation_and_risk":       true,
}

var idPattern = regexp.MustCompile(`^IMP-F[1-5]-\d{2}$`)

// ValidateAndExport enforces Gate 1 (Portability), Gate 2 (Non-Trivial Fork), and Graduation Bar,
// then exports the sanitized quote-free `<instance>/imprint.json`.
func ValidateAndExport(instanceDir string) error {
	absInstance, localDir, err := config.ResolveInstancePaths(instanceDir)
	if err != nil {
		return err
	}
	cfg, err := config.LoadSources(localDir)
	if err != nil {
		return err
	}

	claimsBytes, err := os.ReadFile(filepath.Join(localDir, "evidence_ledger.json"))
	if err != nil {
		return err
	}
	var claims []RawClaim
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return err
	}

	demotedBytes, err := os.ReadFile(filepath.Join(localDir, "demoted_non_imprint.json"))
	var demoted []DemotedRule
	if err == nil {
		_ = json.Unmarshal(demotedBytes, &demoted)
	}

	var termPatterns []*regexp.Regexp
	for _, t := range cfg.WorkspaceTermsDisallowedInPortableRules {
		re, err := regexp.Compile(`(?i)\b` + regexp.QuoteMeta(t) + `\b`)
		if err == nil {
			termPatterns = append(termPatterns, re)
		}
	}

	var errs []string
	facetCounts := map[string]int{}
	var sanitized []PublicClaim

	for _, c := range claims {
		if !idPattern.MatchString(c.ID) {
			errs = append(errs, fmt.Sprintf("%s: id does not match ^IMP-F[1-5]-\\d{2}$", c.ID))
		}
		if !validFacets[c.Facet] {
			errs = append(errs, fmt.Sprintf("%s: invalid facet %q", c.ID, c.Facet))
		} else {
			facetCounts[c.Facet]++
		}

		// Gate 1: Portability check across all exported text fields
		blob := strings.Join([]string{
			c.Title,
			c.PortableRule,
			c.AppliesWhen,
			c.FlipsWhen,
			c.AdmissionTests.PortabilityCheck,
			c.AdmissionTests.PeerFork,
		}, " ")
		for idx, re := range termPatterns {
			if re.MatchString(blob) {
				errs = append(errs, fmt.Sprintf("%s: Gate 1 (Portability) failed — workspace term %q leaked into portable fields",
					c.ID, cfg.WorkspaceTermsDisallowedInPortableRules[idx]))
			}
		}
		if len(strings.TrimSpace(c.AdmissionTests.PortabilityCheck)) < 25 {
			errs = append(errs, fmt.Sprintf("%s: Gate 1 portability_check explanation too short", c.ID))
		}
		if len(strings.TrimSpace(c.AdmissionTests.PeerFork)) < 25 {
			errs = append(errs, fmt.Sprintf("%s: Gate 2 peer_fork explanation too short", c.ID))
		}

		if len(c.Evidence) < 3 {
			errs = append(errs, fmt.Sprintf("%s: fewer than 3 evidence items (%d)", c.ID, len(c.Evidence)))
		}
		srcTypes := map[string]bool{}
		projects := map[string]bool{}
		for _, e := range c.Evidence {
			if e.SourceType != "" {
				srcTypes[e.SourceType] = true
			}
			if e.Project != "" {
				projects[e.Project] = true
			}
		}
		if len(projects) < 2 && len(srcTypes) < 2 {
			errs = append(errs, fmt.Sprintf("%s: evidence does not span >= 2 contexts", c.ID))
		}
		if len(strings.TrimSpace(c.FlipsWhen)) < 15 {
			errs = append(errs, fmt.Sprintf("%s: flips_when boundary condition missing or too short", c.ID))
		}

		sanitized = append(sanitized, PublicClaim{
			ID:             c.ID,
			Facet:          c.Facet,
			Scope:          c.Scope,
			Status:         c.Status,
			Title:          c.Title,
			PortableRule:   c.PortableRule,
			AdmissionTests: c.AdmissionTests,
			AppliesWhen:    c.AppliesWhen,
			FlipsWhen:      c.FlipsWhen,
			ProvenanceCounts: ProvenanceCounts{
				EvidenceItems:        len(c.Evidence),
				DistinctSourceTypes:  len(srcTypes),
				DistinctProjects:     len(projects),
				CounterEvidenceItems: len(c.CounterEvidence),
			},
		})
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}

	pub := PublicImprint{
		SchemaVersion: "1.0.0",
		GeneratedAt:   time.Now().Format("2006-01-02"),
		Claims:        sanitized,
	}
	outBytes, _ := json.MarshalIndent(pub, "", "  ")
	outPath := filepath.Join(absInstance, "imprint.json")
	if err := os.WriteFile(outPath, append(outBytes, '\n'), 0o644); err != nil {
		return err
	}

	fmt.Printf("VALIDATION PASSED for instance '%s'\n", filepath.Base(absInstance))
	fmt.Printf("  Exported quote-free portable summary -> %s\n", outPath)
	fmt.Printf("  Graduated portable imprint claims (L1): %d\n", len(claims))
	var facets []string
	for f := range validFacets {
		facets = append(facets, f)
	}
	sort.Strings(facets)
	for _, f := range facets {
		fmt.Printf("    - %s: %d claims\n", f, facetCounts[f])
	}
	l2, l3 := 0, 0
	for _, d := range demoted {
		if d.Layer == "L2_workspace_binding" {
			l2++
		} else if d.Layer == "L3_agent_baseline" {
			l3++
		}
	}
	fmt.Printf("  Separated local-only rules (.local/): %d (%d L2 workspace bindings, %d L3 agent baseline rules)\n",
		len(demoted), l2, l3)
	return nil
}
