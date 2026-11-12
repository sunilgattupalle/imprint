package link

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLinkInstanceIdempotent(t *testing.T) {
	tmpHome := t.TempDir()
	instanceDir := filepath.Join(tmpHome, ".imprint")
	if err := os.MkdirAll(filepath.Join(instanceDir, ".local"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(instanceDir, "IMPRINT.md"), []byte("# Imprint\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(instanceDir, ".local", "WORKSPACE.md"), []byte("# Workspace\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Simulate installed harnesses (.claude and .pi)
	claudeDir := filepath.Join(tmpHome, ".claude")
	piDir := filepath.Join(tmpHome, ".pi")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(piDir, 0o755); err != nil {
		t.Fatal(err)
	}
	claudeFile := filepath.Join(claudeDir, "CLAUDE.md")
	if err := os.WriteFile(claudeFile, []byte("# Existing Rules\nKeep output concise.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res1, err := LinkInstance(instanceDir, tmpHome, false)
	if err != nil {
		t.Fatalf("first LinkInstance failed: %v", err)
	}
	if len(res1) != 2 {
		t.Fatalf("expected 2 linked harnesses, got %d (%v)", len(res1), res1)
	}

	data, err := os.ReadFile(claudeFile)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "# Existing Rules") || !strings.Contains(content, "IMPRINT.md") || !strings.Contains(content, "WORKSPACE.md") {
		t.Fatalf("unexpected linked content:\n%s", content)
	}

	// Second call must be idempotent ("unchanged")
	res2, err := LinkInstance(instanceDir, tmpHome, false)
	if err != nil {
		t.Fatalf("second LinkInstance failed: %v", err)
	}
	for _, r := range res2 {
		if r.Action != "unchanged" {
			t.Fatalf("expected unchanged on second run for %s, got %s", r.Harness, r.Action)
		}
	}
}
