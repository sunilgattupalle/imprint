package ingest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractHTMLDocument(t *testing.T) {
	tmp := t.TempDir()
	htmlPath := filepath.Join(tmp, "sample.html")
	content := `<!DOCTYPE html>
<html>
<head>
<style>
  :root {
    --surface-page: #f6f5f2;
    --accent: #2a78d6;
  }
</style>
</head>
<body>
  <h1>Architecture One-Pager</h1>
  <p>Once you see the construct, the API falls out of it.</p>
</body>
</html>`
	if err := os.WriteFile(htmlPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := ExtractDocument(htmlPath)
	if err != nil {
		t.Fatalf("ExtractDocument failed: %v", err)
	}
	if len(doc.CSSTokens) != 2 {
		t.Fatalf("expected 2 CSS tokens, got %d (%v)", len(doc.CSSTokens), doc.CSSTokens)
	}
	if !strings.Contains(doc.Text, "Once you see the construct") {
		t.Fatalf("expected body text in extracted doc, got %q", doc.Text)
	}
}
