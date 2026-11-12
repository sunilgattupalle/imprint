package eval

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunHoldoutEvalExportsPairsAndBlindSpots(t *testing.T) {
	tmp := t.TempDir()
	localDir := filepath.Join(tmp, ".local")
	if err := os.MkdirAll(filepath.Join(localDir, "corpus"), 0o755); err != nil {
		t.Fatal(err)
	}
	ledger := `[
  {
    "id": "IMP-F2-01",
    "facet": "F2_dialectical_style",
    "scope": "global",
    "status": "graduated",
    "title": "Stress-test claims and hold ground on skeptical probes",
    "portable_rule": "Do not agree just to agree.",
    "admission_tests": {
      "portability_check": "Applies across all repositories and technical discussions.",
      "peer_fork": "Other engineers prefer rapid agreement."
    },
    "applies_when": "Reviewing ideas or answering skeptical questions.",
    "flips_when": "Once an axiom is locked.",
    "eval_matcher": "stress[- ]test|don't just agree",
    "evidence": []
  }
]`
	_ = os.WriteFile(filepath.Join(localDir, "evidence_ledger.json"), []byte(ledger), 0o644)
	_ = os.WriteFile(filepath.Join(localDir, "demoted_non_imprint.json"), []byte(`[]`), 0o644)

	holdoutLines := strings.Join([]string{
		`{"session_id":"s1","agent":"claude","project":"p1","ordinal":2,"timestamp":"2026-01-01","prev_assistant_snippet":"You are completely right!","user_text":"Don't just agree with me, stress-test this architecture.","split":"holdout"}`,
		`{"session_id":"s2","agent":"pi","project":"p2","ordinal":4,"timestamp":"2026-01-02","prev_assistant_snippet":"I created 3 new markdown files.","user_text":"Stop creating unnecessary docs when I ask a question.","split":"holdout"}`,
	}, "\n") + "\n"
	_ = os.WriteFile(filepath.Join(localDir, "corpus", "holdout_turns.jsonl"), []byte(holdoutLines), 0o644)

	res, err := RunHoldoutEval(tmp)
	if err != nil {
		t.Fatalf("RunHoldoutEval failed: %v", err)
	}
	if res.HoldoutSignalTurns != 2 || res.CoveredByL1Portable != 1 || res.UncoveredBlindSpots != 1 || res.ExportedFailurePairs != 2 {
		t.Fatalf("unexpected EvalResults: %+v", res)
	}
	reportBytes, err := os.ReadFile(filepath.Join(localDir, "eval", "holdout_report.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(reportBytes), "Uncovered Held-Out Interventions") || !strings.Contains(string(reportBytes), "Stop creating unnecessary docs") {
		t.Fatalf("expected uncovered blind spot in holdout_report.md, got:\n%s", string(reportBytes))
	}
}
