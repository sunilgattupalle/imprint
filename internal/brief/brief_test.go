package brief

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClassifyModesAndBuildBrief(t *testing.T) {
	// 1. Prompt + file matching for architecture-and-specs
	modes := ClassifyModes("have we updated the readme and specs to reflect the mental model?", []string{"README.md"})
	if len(modes) == 0 || modes[0].ModeID != "architecture-and-specs" {
		t.Fatalf("expected architecture-and-specs mode to rank first, got %+v", modes)
	}

	// 2. HTML prototype matching
	modesHTML := ClassifyModes("build a visual prototype with clean svg diagrams", []string{"index.html"})
	if len(modesHTML) == 0 || modesHTML[0].ModeID != "visual-html-and-prototypes" {
		t.Fatalf("expected visual-html-and-prototypes mode, got %+v", modesHTML)
	}

	// 3. BuildBrief end-to-end assembly with project memories
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "overlays"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, ".local"), 0o755); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(tmp, "IMPRINT.md"), []byte("# Core Imprint\nFind the primitive first.\n"), 0o644)
	_ = os.WriteFile(filepath.Join(tmp, "overlays", "mode-architecture-and-specs.md"), []byte("# Spec Mode\nProgressive depth.\n"), 0o644)
	_ = os.WriteFile(filepath.Join(tmp, ".local", "memories.json"), []byte(`[
  {
    "id": "MEM-01",
    "project": "project_alpha",
    "type": "warning",
    "dimension": "voice.prose_cadence",
    "trigger": "Editing docs",
    "rule": "Never introduce em-dashes in markdown specs."
  }
]`), 0o644)

	projDir := filepath.Join(tmp, "project_alpha")
	_ = os.MkdirAll(projDir, 0o755)

	res, err := BuildBrief(tmp, projDir, "write the architecture spec", []string{"docs/01.md"})
	if err != nil {
		t.Fatalf("BuildBrief failed: %v", err)
	}
	if len(res.MatchedModes) != 1 || res.MatchedModes[0] != "architecture-and-specs" {
		t.Fatalf("unexpected MatchedModes: %v", res.MatchedModes)
	}
	if res.MatchedMemories != 1 {
		t.Fatalf("expected 1 matched project memory, got %d", res.MatchedMemories)
	}
	if !strings.Contains(res.Markdown, "# Spec Mode") || !strings.Contains(res.Markdown, "MEM-01") {
		t.Fatalf("unexpected brief markdown:\n%s", res.Markdown)
	}
}
