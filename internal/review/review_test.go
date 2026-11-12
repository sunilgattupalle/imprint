package review

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunReviewUpdatesHabitReflection(t *testing.T) {
	tmp := t.TempDir()
	localDir := filepath.Join(tmp, ".local")
	if err := os.MkdirAll(localDir, 0o755); err != nil {
		t.Fatal(err)
	}
	sourcesYAML := `owner:
  id: "alice"
  name: "Alice"
session_archive:
  sqlite_path: "` + filepath.Join(tmp, "none.db") + `"
hygiene:
  holdout_ratio: 0.20
`
	if err := os.WriteFile(filepath.Join(localDir, "sources.yaml"), []byte(sourcesYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	ledgerJSON := `[
  {
    "id": "IMP-F1-01",
    "facet": "F1_epistemic_posture",
    "dimension": "epistemic.abstraction_depth",
    "scope": "global",
    "status": "graduated",
    "title": "Define primitives before interfaces",
    "portable_rule": "Identify the underlying construct before designing APIs.",
    "admission_tests": {
      "portability_check": "Applies across all programming languages and domains.",
      "peer_fork": "Other engineers prefer starting from user-facing API mocks."
    },
    "applies_when": "Designing a new system or service.",
    "flips_when": "When patching a localized bug.",
    "evidence": [
      {"source_type": "session", "source_id": "s1", "project": "p1", "timestamp": "2026-01-01", "quote": "q1"},
      {"source_type": "session", "source_id": "s2", "project": "p2", "timestamp": "2026-01-02", "quote": "q2"},
      {"source_type": "authored_doc", "source_id": "d1", "project": "p1", "timestamp": "2026-01-03", "quote": "q3"}
    ],
    "counter_evidence": []
  }
]`
	if err := os.WriteFile(filepath.Join(localDir, "evidence_ledger.json"), []byte(ledgerJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RunReview(tmp, []string{"IMP-F1-01=aspirational_shift"}); err != nil {
		t.Fatalf("RunReview failed: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(tmp, "imprint.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"habit_reflection": "aspirational_shift"`) {
		t.Fatalf("expected imprint.json to contain updated habit_reflection, got:\n%s", string(b))
	}
}
