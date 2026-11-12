package corpus

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseAuggieV2AndLegacyV1(t *testing.T) {
	tmp := t.TempDir()

	// 1. Test Auggie v2 (Pi JSONL format in ~/.augment/sessions-v2)
	v2Path := filepath.Join(tmp, "session-v2.jsonl")
	v2Content := `{"type":"session","id":"uuid-v2","cwd":"/work/project_alpha","timestamp":"2026-01-01T10:00:00Z"}
{"type":"message","timestamp":"2026-01-01T10:01:00Z","message":{"role":"user","content":"Align on the mental model before writing code."}}
{"type":"message","timestamp":"2026-01-01T10:02:00Z","message":{"role":"assistant","content":"Here is the core construct."}}
`
	if err := os.WriteFile(v2Path, []byte(v2Content), 0o644); err != nil {
		t.Fatal(err)
	}
	nsV2, ok := parsePiFamilyFile(v2Path, "auggie")
	if !ok {
		t.Fatalf("expected parsePiFamilyFile to succeed for auggie v2")
	}
	if nsV2.row.effectiveAgent != "auggie" || nsV2.row.project != "project_alpha" || len(nsV2.messages) != 2 {
		t.Fatalf("unexpected auggie v2 session: %+v, msgs=%d", nsV2.row, len(nsV2.messages))
	}

	// 2. Test Auggie v1 Legacy JSON format (~/.augment/sessions/*.json)
	v1Path := filepath.Join(tmp, "session-v1.json")
	v1Content := `{
  "sessionId": "uuid-v1",
  "created": "2025-11-01T09:00:00Z",
  "chatHistory": [
    {
      "finishedAt": "2025-11-01T09:01:00Z",
      "isHistorySummary": false,
      "exchange": {
        "request_message": "Stress-test this design first.",
        "response_text": "Here are the two main failure modes."
      }
    }
  ]
}`
	if err := os.WriteFile(v1Path, []byte(v1Content), 0o644); err != nil {
		t.Fatal(err)
	}
	nsV1, ok := parseAuggieLegacyJSONFile(v1Path)
	if !ok {
		t.Fatalf("expected parseAuggieLegacyJSONFile to succeed for auggie v1")
	}
	if nsV1.row.id != "auggie:uuid-v1" || len(nsV1.messages) != 2 {
		t.Fatalf("unexpected auggie v1 session: %+v, msgs=%d", nsV1.row, len(nsV1.messages))
	}
}
