package corpus

import (
	"bufio"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sunilgattupalle/imprint/internal/config"
)

type nativeSession struct {
	row      sessionRow
	messages []msgRow
}

// scanNativeStores scans local Pi, OMP, Auggie, and Claude JSONL stores directly
// so imprint works out of the box even before agentsview is synced or installed.
func scanNativeStores() map[string]nativeSession {
	out := map[string]nativeSession{}

	piFamilyStores := []struct {
		agent string
		dir   string
	}{
		{"auggie", config.ExpandPath("~/.augment/sessions-v2")},
		{"pi", config.ExpandPath("~/.pi/agent/sessions")},
		{"omp", config.ExpandPath("~/.omp/agent/sessions")},
	}

	for _, store := range piFamilyStores {
		if info, err := os.Stat(store.dir); err != nil || !info.IsDir() {
			continue
		}
		_ = filepath.WalkDir(store.dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".jsonl") {
				return nil
			}
			ns, ok := parsePiFamilyFile(path, store.agent)
			if !ok || len(ns.messages) == 0 {
				return nil
			}
			if prev, exists := out[ns.row.rawUUID]; !exists || len(ns.messages) > len(prev.messages) {
				out[ns.row.rawUUID] = ns
			}
			return nil
		})
	}

	claudeDir := config.ExpandPath("~/.claude/projects")
	if info, err := os.Stat(claudeDir); err == nil && info.IsDir() {
		_ = filepath.WalkDir(claudeDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".jsonl") {
				return nil
			}
			ns, ok := parseClaudeJSONLFile(path)
			if !ok || len(ns.messages) == 0 {
				return nil
			}
			if prev, exists := out[ns.row.rawUUID]; !exists || len(ns.messages) > len(prev.messages) {
				out[ns.row.rawUUID] = ns
			}
			return nil
		})
	}

	return out
}

func parsePiFamilyFile(path string, agent string) (nativeSession, bool) {
	f, err := os.Open(path)
	if err != nil {
		return nativeSession{}, false
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)

	var (
		sessionID string
		cwd       string
		startedAt string
		parentID  string
		msgs      []msgRow
		ordinal   int
	)

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var raw map[string]any
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}
		typ, _ := raw["type"].(string)
		switch typ {
		case "session":
			sessionID, _ = raw["id"].(string)
			cwd, _ = raw["cwd"].(string)
			startedAt, _ = raw["timestamp"].(string)
			parentID, _ = raw["parentSession"].(string)
		case "message":
			ts, _ := raw["timestamp"].(string)
			msgObj, ok := raw["message"].(map[string]any)
			if !ok {
				continue
			}
			role, _ := msgObj["role"].(string)
			if role != "user" && role != "assistant" {
				continue
			}
			txt := extractMessageContentText(msgObj["content"])
			if strings.TrimSpace(txt) == "" {
				continue
			}
			msgs = append(msgs, msgRow{
				ordinal:   ordinal,
				role:      role,
				content:   txt,
				timestamp: ts,
			})
			ordinal++
		}
	}

	if sessionID == "" {
		base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		sessionID = base
	}
	proj := projectFromCwd(cwd)
	return nativeSession{
		row: sessionRow{
			id:              agent + ":" + sessionID,
			rawUUID:         sessionID,
			project:         proj,
			agent:           agent,
			effectiveAgent:  agent,
			startedAt:       startedAt,
			messageCount:    len(msgs),
			parentSessionID: parentID,
			cwd:             cwd,
		},
		messages: msgs,
	}, true
}

func parseClaudeJSONLFile(path string) (nativeSession, bool) {
	f, err := os.Open(path)
	if err != nil {
		return nativeSession{}, false
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)

	var (
		sessionID string
		cwd       string
		gitBranch string
		startedAt string
		msgs      []msgRow
		ordinal   int
	)

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var raw map[string]any
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}
		if sid, _ := raw["sessionId"].(string); sid != "" && sessionID == "" {
			sessionID = sid
		}
		if c, _ := raw["cwd"].(string); c != "" && cwd == "" {
			cwd = c
		}
		if gb, _ := raw["gitBranch"].(string); gb != "" && gitBranch == "" {
			gitBranch = gb
		}
		ts, _ := raw["timestamp"].(string)
		if ts != "" && startedAt == "" {
			startedAt = ts
		}
		typ, _ := raw["type"].(string)
		if typ != "user" && typ != "assistant" {
			continue
		}
		msgObj, ok := raw["message"].(map[string]any)
		if !ok {
			continue
		}
		role, _ := msgObj["role"].(string)
		if role == "" {
			role = typ
		}
		txt := extractMessageContentText(msgObj["content"])
		if strings.TrimSpace(txt) == "" {
			continue
		}
		msgs = append(msgs, msgRow{
			ordinal:   ordinal,
			role:      role,
			content:   txt,
			timestamp: ts,
		})
		ordinal++
	}

	if sessionID == "" {
		sessionID = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	return nativeSession{
		row: sessionRow{
			id:             "claude:" + sessionID,
			rawUUID:        sessionID,
			project:        projectFromCwd(cwd),
			agent:          "claude",
			effectiveAgent: "claude",
			startedAt:      startedAt,
			messageCount:   len(msgs),
			cwd:            cwd,
			gitBranch:      gitBranch,
		},
		messages: msgs,
	}, true
}

func extractMessageContentText(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []any:
		var parts []string
		for _, item := range v {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			typ, _ := m["type"].(string)
			if typ == "tool_result" || typ == "toolResult" || typ == "thinking" {
				continue
			}
			if txt, ok := m["text"].(string); ok && strings.TrimSpace(txt) != "" {
				parts = append(parts, txt)
			}
		}
		return strings.Join(parts, "\n")
	default:
		return ""
	}
}

func projectFromCwd(cwd string) string {
	cwd = strings.TrimSpace(cwd)
	if cwd == "" {
		return "unassigned"
	}
	base := filepath.Base(filepath.Clean(cwd))
	if base == "" || base == "." || base == "/" {
		return "unassigned"
	}
	return base
}
