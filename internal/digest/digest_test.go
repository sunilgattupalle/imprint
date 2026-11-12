package digest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sunilgattupalle/imprint/internal/corpus"
)

func TestDiversifyTurnsAndEmergentClustering(t *testing.T) {
	turns := []corpus.TurnRecord{
		{SessionID: "s1", Project: "proj_a", UserText: "t1"},
		{SessionID: "s1", Project: "proj_a", UserText: "t2"},
		{SessionID: "s1", Project: "proj_a", UserText: "t3"},
		{SessionID: "s2", Project: "proj_b", UserText: "t4"},
		{SessionID: "s3", Project: "proj_c", UserText: "t5"},
	}
	sampled := diversifyTurns(turns, 3)
	if len(sampled) != 3 {
		t.Fatalf("expected 3 sampled turns, got %d", len(sampled))
	}
	projects := map[string]bool{}
	for _, s := range sampled {
		projects[s.Project] = true
	}
	if len(projects) != 3 {
		t.Fatalf("expected round-robin diversity across all 3 projects, got %v", projects)
	}

	// Verify generalCorrectionRE catches person-agnostic corrections not in seeded probes
	correction := "No, please stop adding getter boilerplate to value structs."
	if !generalCorrectionRE.MatchString(correction) {
		t.Fatalf("expected generalCorrectionRE to match %q", correction)
	}

	// Verify stratified clipping keeps both head and tail
	longText := strings.Repeat("HEAD_MARKER ", 100) + strings.Repeat("MID ", 200) + strings.Repeat("TAIL_MARKER ", 100)
	clipped := clipTextStratified(longText, 500)
	if !strings.Contains(clipped, "HEAD_MARKER") || !strings.Contains(clipped, "TAIL_MARKER") {
		t.Fatalf("expected stratified clip to preserve both head and tail, got: %q", clipped)
	}

	// Verify ExtractVersionDiffs detects -v1 -> -v2 deltas
	tmp := t.TempDir()
	v1 := filepath.Join(tmp, "spec-v1.md")
	v2 := filepath.Join(tmp, "spec-v2.md")
	_ = os.WriteFile(v1, []byte("# Old Heading\nThis draft — has em-dashes — and extra filler text.\n"), 0o644)
	_ = os.WriteFile(v2, []byte("# New Heading\nThis draft states the primitive directly.\n"), 0o644)
	diffs := ExtractVersionDiffs([]string{v1, v2})
	if len(diffs) != 1 || diffs[0].BeforeEmDashes != 2 || diffs[0].AfterEmDashes != 0 {
		t.Fatalf("unexpected version diff result: %+v", diffs)
	}
}

func TestBuildDigestWithPreAIAndNativeSession(t *testing.T) {
	tmp := t.TempDir()
	localDir := filepath.Join(tmp, ".local")
	if err := os.MkdirAll(localDir, 0o755); err != nil {
		t.Fatal(err)
	}
	preAIFile := filepath.Join(tmp, "essay_2021.md")
	if err := os.WriteFile(preAIFile, []byte("# Essay\nClear thinking requires explicit primitives.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sourcesYAML := `owner:
  id: "alice"
  name: "Alice"
session_archive:
  sqlite_path: "` + filepath.Join(tmp, "nonexistent.db") + `"
authored_artifacts:
  pre_ai_globs:
    - "` + preAIFile + `"
hygiene:
  holdout_ratio: 0.20
`
	if err := os.WriteFile(filepath.Join(localDir, "sources.yaml"), []byte(sourcesYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	digestPath, _, err := BuildDigest(tmp)
	if err != nil {
		t.Fatalf("BuildDigest failed: %v", err)
	}
	b, err := os.ReadFile(digestPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(b)
	if !strings.Contains(content, "[PRE-AI GOLD]") || !strings.Contains(content, "essay_2021.md") {
		t.Fatalf("expected Pre-AI Gold section in digest.md, got:\n%s", content)
	}
}
