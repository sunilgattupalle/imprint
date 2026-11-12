package corpus

import (
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/sunilgattupalle/imprint/internal/config"
	_ "modernc.org/sqlite"
)

type TurnRecord struct {
	SessionID            string `json:"session_id"`
	RawUUID              string `json:"raw_uuid"`
	Agent                string `json:"agent"`
	Project              string `json:"project"`
	Cwd                  string `json:"cwd"`
	GitBranch            string `json:"git_branch"`
	IsSubagent           bool   `json:"is_subagent"`
	Ordinal              int    `json:"ordinal"`
	Timestamp            string `json:"timestamp"`
	UserText             string `json:"user_text"`
	UserChars            int    `json:"user_chars"`
	PrevAssistantSnippet string `json:"prev_assistant_snippet,omitempty"`
	NextAssistantSnippet string `json:"next_assistant_snippet,omitempty"`
	Split                string `json:"split"`
}

type CorpusStats struct {
	RawSessionsInDB                   int            `json:"raw_sessions_in_db"`
	CanonicalSessionsAfterDedup       int            `json:"canonical_sessions_after_dedup"`
	DuplicateSessionsDropped          int            `json:"duplicate_sessions_dropped"`
	CursorSessionsEnrichedWithProject int            `json:"cursor_sessions_enriched_with_project"`
	RejectedTurnsByReason             map[string]int `json:"rejected_turns_by_reason"`
	KeptTurnsByAgent                  map[string]int `json:"kept_turns_by_agent"`
	SessionsByAgent                   map[string]int `json:"sessions_by_agent"`
	TrainSessions                     int            `json:"train_sessions"`
	HoldoutSessions                   int            `json:"holdout_sessions"`
	TrainTurns                        int            `json:"train_turns"`
	HoldoutTurns                      int            `json:"holdout_turns"`
}

type compiledRule struct {
	name string
	re   *regexp.Regexp
}

var defaultRejectRules = []compiledRule{
	{"worker_tag", regexp.MustCompile(`^\s*<worker\b`)},
	{"teammate_msg", regexp.MustCompile(`^\s*<teammate-message\b`)},
	{"local_command", regexp.MustCompile(`^\s*<(?:local-command-caveat|command-name|command-message)\b`)},
	{"task_notification", regexp.MustCompile(`^\s*<task-notification\b`)},
	{"system_reminder", regexp.MustCompile(`^\s*<system-reminder\b`)},
	{"subagent_header", regexp.MustCompile(`^\s*\[Subagent Context\]`)},
	{"slash_command", regexp.MustCompile(`^\s*/(?:login|logout|clear|compact|mcp|model|help|status|resume|init|cost|doctor|review|pr_comments)\b`)},
	{"ultraplan_scaffold", regexp.MustCompile(`(?i)^\s*Use ultraplan\b`)},
}

var defaultInlineStripRE = []*regexp.Regexp{
	regexp.MustCompile(`(?s)<system-reminder\b[^>]*>.*?</system-reminder>`),
	regexp.MustCompile(`(?s)<local-command-caveat\b[^>]*>.*?</local-command-caveat>`),
}

var timestampLineRE = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}`)

type sessionRow struct {
	id              string
	rawUUID         string
	project         string
	agent           string
	effectiveAgent  string
	startedAt       string
	messageCount    int
	parentSessionID string
	cwd             string
	gitBranch       string
}

type msgRow struct {
	ordinal   int
	role      string
	content   string
	timestamp string
}

func extractRawUUID(id string) string {
	if idx := strings.IndexByte(id, ':'); idx >= 0 {
		return id[idx+1:]
	}
	return id
}

func isHoldout(rawUUID string, ratio float64) bool {
	sum := sha256.Sum256([]byte(rawUUID))
	val := binary.BigEndian.Uint32(sum[:4])
	return int(val%1000) < int(ratio*1000)
}

func cleanUserText(raw string, rejectRules []compiledRule, inlineStrips []*regexp.Regexp) (string, string) {
	if strings.TrimSpace(raw) == "" {
		return "", "empty"
	}
	for _, r := range rejectRules {
		if r.re.MatchString(raw) {
			return "", r.name
		}
	}
	cleaned := raw
	for _, reStrip := range inlineStrips {
		cleaned = reStrip.ReplaceAllString(cleaned, "")
	}
	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "" {
		return "", "empty_after_strip"
	}
	rawLines := strings.Split(cleaned, "\n")
	var nonEmpty []string
	for _, ln := range rawLines {
		t := strings.TrimSpace(ln)
		if t != "" {
			nonEmpty = append(nonEmpty, t)
		}
	}
	if len(nonEmpty) >= 6 {
		stackCount := 0
		for _, ln := range nonEmpty {
			if strings.HasPrefix(ln, "at ") ||
				strings.HasPrefix(ln, "File \"") ||
				strings.HasPrefix(ln, "Traceback (most recent") ||
				strings.HasPrefix(ln, "Caused by:") ||
				strings.HasPrefix(ln, "E   ") ||
				strings.HasPrefix(ln, "STDERR:") ||
				timestampLineRE.MatchString(ln) {
				stackCount++
			}
		}
		if float64(stackCount)/float64(len(nonEmpty)) > 0.75 {
			return "", "pure_log_or_stacktrace"
		}
	}
	return cleaned, ""
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max]
}

// Build executes the 4-stage corpus pipeline and writes train_turns.jsonl, holdout_turns.jsonl, and stats.json.
func Build(instanceDir string) (*CorpusStats, error) {
	_, localDir, err := config.ResolveInstancePaths(instanceDir)
	if err != nil {
		return nil, err
	}
	cfg, err := config.LoadSources(localDir)
	if err != nil {
		return nil, err
	}

	rejectRules := append([]compiledRule{}, defaultRejectRules...)
	for _, extra := range cfg.Hygiene.ExtraRejectPatterns {
		re, err := regexp.Compile(extra.Regex)
		if err != nil {
			return nil, fmt.Errorf("invalid reject regex %q: %w", extra.Name, err)
		}
		rejectRules = append(rejectRules, compiledRule{name: extra.Name, re: re})
	}
	inlineStrips := append([]*regexp.Regexp{}, defaultInlineStripRE...)
	for _, rx := range cfg.Hygiene.ExtraInlineStripRegexes {
		re, err := regexp.Compile("(?s)" + rx)
		if err != nil {
			return nil, fmt.Errorf("invalid inline strip regex %q: %w", rx, err)
		}
		inlineStrips = append(inlineStrips, re)
	}

	dbPath := config.ExpandPath(cfg.SessionArchive.SQLitePath)
	var db *sql.DB
	if _, statErr := os.Stat(dbPath); statErr == nil {
		db, err = sql.Open("sqlite", fmt.Sprintf("file:%s?mode=ro", dbPath))
		if err != nil {
			db = nil
		} else {
			defer db.Close()
		}
	}

	var allSessions []sessionRow
	byUUID := map[string][]sessionRow{}

	if db != nil {
		labelOverrides := map[string]string{}
		lRows, err := db.Query(`SELECT session_id, label FROM session_labels WHERE label LIKE 'source=%'`)
		if err == nil {
			for lRows.Next() {
				var sid, lbl string
				if err := lRows.Scan(&sid, &lbl); err == nil {
					if parts := strings.SplitN(lbl, "=", 2); len(parts) == 2 {
						labelOverrides[sid] = parts[1]
					}
				}
			}
			lRows.Close()
		}

		sRows, err := db.Query(`
			SELECT id, COALESCE(project, ''), COALESCE(agent, ''), COALESCE(started_at, ''),
			       COALESCE(message_count, 0), COALESCE(parent_session_id, ''),
			       COALESCE(cwd, ''), COALESCE(git_branch, '')
			FROM sessions
			WHERE COALESCE(is_automated, 0) = 0
		`)
		if err == nil {
			for sRows.Next() {
				var s sessionRow
				if err := sRows.Scan(&s.id, &s.project, &s.agent, &s.startedAt, &s.messageCount, &s.parentSessionID, &s.cwd, &s.gitBranch); err != nil {
					continue
				}
				s.rawUUID = extractRawUUID(s.id)
				if ov, ok := labelOverrides[s.id]; ok {
					s.effectiveAgent = ov
				} else {
					s.effectiveAgent = s.agent
				}
				allSessions = append(allSessions, s)
				byUUID[s.rawUUID] = append(byUUID[s.rawUUID], s)
			}
			sRows.Close()
		}
	}

	// Merge any native on-disk sessions (Auggie, Pi, OMP, Claude) not yet in sessions.db
	var nativeByUUID map[string]nativeSession
	if cfg.SessionArchive.SQLitePath == "~/.agentsview/sessions.db" || cfg.SessionArchive.SQLitePath == "" {
		nativeByUUID = scanNativeStores()
		for rawUUID, ns := range nativeByUUID {
			if _, inDB := byUUID[rawUUID]; !inDB {
				allSessions = append(allSessions, ns.row)
				byUUID[rawUUID] = append(byUUID[rawUUID], ns.row)
			}
		}
	}

	var canonical []sessionRow
	dedupDropped := 0
	cursorMetaMerged := 0
	var uuids []string
	for u := range byUUID {
		uuids = append(uuids, u)
	}
	sort.Strings(uuids)

	for _, u := range uuids {
		group := byUUID[u]
		if len(group) == 1 {
			canonical = append(canonical, group[0])
			continue
		}
		byAgent := map[string]sessionRow{}
		for _, g := range group {
			byAgent[g.effectiveAgent] = g
		}
		p, hasIDE := byAgent["cursor-ide"]
		sec, hasCLI := byAgent["cursor"]
		if hasIDE && hasCLI {
			if (p.project == "" || p.project == "unassigned") && sec.project != "" {
				p.project = sec.project
				cursorMetaMerged++
			}
			if p.cwd == "" && sec.cwd != "" {
				p.cwd = sec.cwd
			}
			if p.gitBranch == "" && sec.gitBranch != "" {
				p.gitBranch = sec.gitBranch
			}
			canonical = append(canonical, p)
			dedupDropped += len(group) - 1
		} else {
			sort.Slice(group, func(i, j int) bool { return group[i].messageCount > group[j].messageCount })
			canonical = append(canonical, group[0])
			dedupDropped += len(group) - 1
		}
	}

	outDir := filepath.Join(localDir, "corpus")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}
	trainF, err := os.Create(filepath.Join(outDir, "train_turns.jsonl"))
	if err != nil {
		return nil, err
	}
	defer trainF.Close()
	holdoutF, err := os.Create(filepath.Join(outDir, "holdout_turns.jsonl"))
	if err != nil {
		return nil, err
	}
	defer holdoutF.Close()

	stats := &CorpusStats{
		RawSessionsInDB:                   len(allSessions),
		CanonicalSessionsAfterDedup:       len(canonical),
		DuplicateSessionsDropped:          dedupDropped,
		CursorSessionsEnrichedWithProject: cursorMetaMerged,
		RejectedTurnsByReason:             map[string]int{},
		KeptTurnsByAgent:                  map[string]int{},
		SessionsByAgent:                   map[string]int{},
	}

	for _, s := range canonical {
		holdout := isHoldout(s.rawUUID, cfg.Hygiene.HoldoutRatio)
		var msgs []msgRow
		if ns, ok := nativeByUUID[s.rawUUID]; ok && db == nil {
			msgs = ns.messages
		} else if db != nil {
			mRows, err := db.Query(`
				SELECT ordinal, COALESCE(role, ''), COALESCE(content, ''), COALESCE(timestamp, '')
				FROM messages WHERE session_id = ? ORDER BY ordinal ASC
			`, s.id)
			if err == nil {
				for mRows.Next() {
					var m msgRow
					if err := mRows.Scan(&m.ordinal, &m.role, &m.content, &m.timestamp); err == nil {
						msgs = append(msgs, m)
					}
				}
				mRows.Close()
			}
			if len(msgs) == 0 {
				if ns, ok := nativeByUUID[s.rawUUID]; ok {
					msgs = ns.messages
				}
			}
		}

		keptInSession := 0
		for i, m := range msgs {
			if m.role != "user" {
				continue
			}
			cleaned, reason := cleanUserText(m.content, rejectRules, inlineStrips)
			if reason != "" {
				stats.RejectedTurnsByReason[reason]++
				continue
			}
			var prevA, nextA string
			for j := i - 1; j >= 0; j-- {
				if msgs[j].role == "assistant" && strings.TrimSpace(msgs[j].content) != "" {
					prevA = truncate(msgs[j].content, 600)
					break
				}
			}
			for j := i + 1; j < len(msgs); j++ {
				if msgs[j].role == "assistant" && strings.TrimSpace(msgs[j].content) != "" {
					nextA = truncate(msgs[j].content, 600)
					break
				}
			}
			proj := s.project
			if proj == "" {
				proj = "unassigned"
			}
			ts := m.timestamp
			if ts == "" {
				ts = s.startedAt
			}
			split := "train"
			if holdout {
				split = "holdout"
			}
			rec := TurnRecord{
				SessionID:            s.id,
				RawUUID:              s.rawUUID,
				Agent:                s.effectiveAgent,
				Project:              proj,
				Cwd:                  s.cwd,
				GitBranch:            s.gitBranch,
				IsSubagent:           s.parentSessionID != "",
				Ordinal:              m.ordinal,
				Timestamp:            ts,
				UserText:             cleaned,
				UserChars:            len(cleaned),
				PrevAssistantSnippet: prevA,
				NextAssistantSnippet: nextA,
				Split:                split,
			}
			b, _ := json.Marshal(rec)
			if holdout {
				holdoutF.Write(append(b, '\n'))
				stats.HoldoutTurns++
			} else {
				trainF.Write(append(b, '\n'))
				stats.TrainTurns++
			}
			stats.KeptTurnsByAgent[s.effectiveAgent]++
			keptInSession++
		}
		if keptInSession > 0 {
			stats.SessionsByAgent[s.effectiveAgent]++
			if holdout {
				stats.HoldoutSessions++
			} else {
				stats.TrainSessions++
			}
		}
	}

	// Ingest ChatGPT JSON exports if configured
	defaultProj := cfg.ChatGPTExports.DefaultProject
	if defaultProj == "" {
		defaultProj = "exported_chats"
	}
	for _, pattern := range cfg.ChatGPTExports.Globs {
		matches, _ := filepath.Glob(config.ExpandPath(pattern))
		sort.Strings(matches)
		for _, p := range matches {
			rawBytes, err := os.ReadFile(p)
			if err != nil {
				continue
			}
			var payload struct {
				UpdatedAt string `json:"updated_at"`
				Messages  []struct {
					Role    string `json:"role"`
					Content string `json:"content"`
					Text    string `json:"text"`
				} `json:"messages"`
			}
			if err := json.Unmarshal(rawBytes, &payload); err != nil {
				continue
			}
			base := strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))
			sid := "chatgpt:" + base
			holdout := isHoldout(sid, cfg.Hygiene.HoldoutRatio)
			kept := 0
			msgText := func(idx int) string {
				if payload.Messages[idx].Content != "" {
					return payload.Messages[idx].Content
				}
				return payload.Messages[idx].Text
			}
			for idx, m := range payload.Messages {
				if m.Role != "user" {
					continue
				}
				cleaned, reason := cleanUserText(msgText(idx), rejectRules, inlineStrips)
				if reason != "" {
					stats.RejectedTurnsByReason[reason]++
					continue
				}
				var prevA, nextA string
				if idx > 0 && payload.Messages[idx-1].Role == "assistant" {
					prevA = truncate(msgText(idx-1), 600)
				}
				if idx+1 < len(payload.Messages) && payload.Messages[idx+1].Role == "assistant" {
					nextA = truncate(msgText(idx+1), 600)
				}
				split := "train"
				if holdout {
					split = "holdout"
				}
				rec := TurnRecord{
					SessionID:            sid,
					RawUUID:              sid,
					Agent:                "chatgpt",
					Project:              defaultProj,
					Cwd:                  filepath.Dir(p),
					Ordinal:              idx,
					Timestamp:            payload.UpdatedAt,
					UserText:             cleaned,
					UserChars:            len(cleaned),
					PrevAssistantSnippet: prevA,
					NextAssistantSnippet: nextA,
					Split:                split,
				}
				b, _ := json.Marshal(rec)
				if holdout {
					holdoutF.Write(append(b, '\n'))
					stats.HoldoutTurns++
				} else {
					trainF.Write(append(b, '\n'))
					stats.TrainTurns++
				}
				stats.KeptTurnsByAgent["chatgpt"]++
				kept++
			}
			if kept > 0 {
				stats.SessionsByAgent["chatgpt"]++
				if holdout {
					stats.HoldoutSessions++
				} else {
					stats.TrainSessions++
				}
			}
		}
	}

	statsBytes, _ := json.MarshalIndent(stats, "", "  ")
	if err := os.WriteFile(filepath.Join(outDir, "stats.json"), statsBytes, 0o644); err != nil {
		return nil, err
	}
	return stats, nil
}
