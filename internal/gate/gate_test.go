package gate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateRejectsLeakedWorkspaceTerm(t *testing.T) {
	tmp := t.TempDir()
	localDir := filepath.Join(tmp, ".local")
	if err := os.MkdirAll(localDir, 0o755); err != nil {
		t.Fatal(err)
	}
	sourcesYAML := `owner:
  id: "alice"
  name: "Alice"
session_archive:
  sqlite_path: "~/.agentsview/sessions.db"
hygiene:
  holdout_ratio: 0.20
workspace_terms_disallowed_in_portable_rules:
  - "InternalSQL"
`
	if err := os.WriteFile(filepath.Join(localDir, "sources.yaml"), []byte(sourcesYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	ledgerJSON := `[
  {
    "id": "IMP-F1-01",
    "facet": "F1_epistemic_posture",
    "scope": "global",
    "status": "graduated",
    "title": "Always use InternalSQL for queries",
    "portable_rule": "Never bypass InternalSQL with raw queries.",
    "admission_tests": {
      "portability_check": "Applies across all services and data stores in any stack.",
      "peer_fork": "Other engineers prefer writing raw SQL queries directly."
    },
    "applies_when": "Writing data access code.",
    "flips_when": "When running one-off migration scripts.",
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
	_ = os.WriteFile(filepath.Join(localDir, "demoted_non_imprint.json"), []byte(`[]`), 0o644)

	err := ValidateAndExport(tmp)
	if err == nil || !strings.Contains(err.Error(), "InternalSQL") {
		t.Fatalf("expected Gate 1 Portability failure for leaked term InternalSQL, got: %v", err)
	}
}

func TestValidateRejectsNonWordDisallowedTermAndAcceptsEmergentClaim(t *testing.T) {
	tmp := t.TempDir()
	localDir := filepath.Join(tmp, ".local")
	if err := os.MkdirAll(localDir, 0o755); err != nil {
		t.Fatal(err)
	}
	sourcesYAML := `owner:
  id: "alice"
  name: "Alice"
session_archive:
  sqlite_path: "~/.agentsview/sessions.db"
hygiene:
  holdout_ratio: 0.20
workspace_terms_disallowed_in_portable_rules:
  - "@acme/core"
`
	if err := os.WriteFile(filepath.Join(localDir, "sources.yaml"), []byte(sourcesYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	// 1. Verify non-word prefix term (@acme/core) is caught by Gate 1
	leakedJSON := `[
  {
    "id": "IMP-E-01",
    "facet": "F6_emergent",
    "dimension": "emergent.package_hygiene",
    "scope": "global",
    "status": "graduated",
    "title": "Import from @acme/core instead of subpackages",
    "default_model_failure": "Importing internal subpackages directly instead of the public package entrypoint.",
    "portable_rule": "Always import shared primitives from @acme/core.",
    "admission_tests": {
      "portability_check": "Applies across all TypeScript monorepos regardless of domain.",
      "peer_fork": "Other engineers prefer deep relative imports to minimize barrel size."
    },
    "applies_when": "Writing package imports.",
    "flips_when": "When building standalone zero-dependency scripts.",
    "evidence": [
      {"source_type": "session", "source_id": "s1", "project": "p1", "timestamp": "2026-01-01", "quote": "q1"},
      {"source_type": "session", "source_id": "s2", "project": "p2", "timestamp": "2026-01-02", "quote": "q2"},
      {"source_type": "authored_doc", "source_id": "d1", "project": "p1", "timestamp": "2026-01-03", "quote": "q3"}
    ],
    "counter_evidence": []
  }
]`
	if err := os.WriteFile(filepath.Join(localDir, "evidence_ledger.json"), []byte(leakedJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ValidateAndExport(tmp); err == nil || !strings.Contains(err.Error(), "@acme/core") {
		t.Fatalf("expected Gate 1 failure for @acme/core, got: %v", err)
	}

	// 2. Clean portable rule with F6_emergent + habit_reflection passes
	cleanJSON := strings.ReplaceAll(leakedJSON, "@acme/core", "the root barrel package")
	if err := os.WriteFile(filepath.Join(localDir, "evidence_ledger.json"), []byte(cleanJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ValidateAndExport(tmp); err != nil {
		t.Fatalf("expected clean emergent claim to validate, got: %v", err)
	}
}
