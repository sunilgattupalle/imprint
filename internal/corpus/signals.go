package corpus

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sunilgattupalle/imprint/internal/config"
	_ "modernc.org/sqlite"
)

// FrictionRecord captures a turn-level friction or interruption event from sessions.db.
type FrictionRecord struct {
	SessionID  string `json:"session_id"`
	Agent      string `json:"agent"`
	Project    string `json:"project"`
	Kind       string `json:"kind"`
	Detector   string `json:"detector"`
	ToolName   string `json:"tool_name,omitempty"`
	Title      string `json:"title"`
	Text       string `json:"text"`
	OccurredAt string `json:"occurred_at"`
}

// TrajectoryRecord captures a tool-call sequence (e.g. write/edit before alignment)
// that immediately triggered a user correction or redirection.
type TrajectoryRecord struct {
	SessionID    string `json:"session_id"`
	Agent        string `json:"agent"`
	Project      string `json:"project"`
	ToolName     string `json:"tool_name"`
	Category     string `json:"category"`
	FilePath     string `json:"file_path,omitempty"`
	Ordinal      int    `json:"ordinal"`
	UserReaction string `json:"user_reaction"`
	Timestamp    string `json:"timestamp"`
}

// RecallRecord captures an accepted episodic memory entry from sessions.db (recall_entries).
type RecallRecord struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	Scope           string `json:"scope"`
	Title           string `json:"title"`
	Body            string `json:"body"`
	Trigger         string `json:"trigger,omitempty"`
	Project         string `json:"project"`
	Agent           string `json:"agent"`
	SourceSessionID string `json:"source_session_id"`
	Transferable    bool   `json:"transferable"`
}

var trajectoryPushbackRE = regexp.MustCompile(`(?i)(` +
	`^no[,.! ]|\bdon'?t\b|\bdo not\b|\bstop\b|\bwait\b|\bwhy did you\b|` +
	`\bbefore (?:you |we )?(?:implement|write|code|edit)\b|\bmental model\b|` +
	`\bwrong\b|\brevert\b|\bundo\b|\bdon'?t overwrite\b|\bask first\b` +
	`)`)

// ExtractBehavioralSignals reads friction_findings, tool_calls trajectories, and recall_entries
// from sessions.db (for non-holdout sessions) when those tables exist.
func ExtractBehavioralSignals(instanceDir string) ([]FrictionRecord, []TrajectoryRecord, []RecallRecord) {
	_, localDir, err := config.ResolveInstancePaths(instanceDir)
	if err != nil {
		return nil, nil, nil
	}
	cfg, err := config.LoadSources(localDir)
	if err != nil {
		return nil, nil, nil
	}
	dbPath := config.ExpandPath(cfg.SessionArchive.SQLitePath)
	if _, err := os.Stat(dbPath); err != nil {
		return nil, nil, nil
	}
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=ro", dbPath))
	if err != nil {
		return nil, nil, nil
	}
	defer db.Close()

	frictions := queryFrictionFindings(db, cfg.Hygiene.HoldoutRatio)
	trajectories := queryToolTrajectoryPushbacks(db, cfg.Hygiene.HoldoutRatio)
	recalls := queryRecallEntries(db)
	return frictions, trajectories, recalls
}

func queryFrictionFindings(db *sql.DB, holdoutRatio float64) []FrictionRecord {
	rows, err := db.Query(`
		SELECT f.session_id, COALESCE(s.agent, ''), COALESCE(s.project, ''),
		       COALESCE(f.kind, ''), COALESCE(f.detector, ''), COALESCE(f.tool_name, ''),
		       COALESCE(f.title, ''), COALESCE(f.text, ''), COALESCE(f.occurred_at, '')
		FROM friction_findings f
		JOIN sessions s ON s.id = f.session_id
		WHERE f.kind IN ('correction', 'frustration', 'interruption')
		ORDER BY f.occurred_at DESC
		LIMIT 250
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var out []FrictionRecord
	seen := map[string]bool{}
	for rows.Next() {
		var r FrictionRecord
		if err := rows.Scan(&r.SessionID, &r.Agent, &r.Project, &r.Kind, &r.Detector, &r.ToolName, &r.Title, &r.Text, &r.OccurredAt); err != nil {
			continue
		}
		rawUUID := extractRawUUID(r.SessionID)
		if isHoldout(rawUUID, holdoutRatio) {
			continue
		}
		r.Text = strings.TrimSpace(r.Text)
		if len(r.Text) < 15 || len(r.Text) > 600 {
			continue
		}
		key := rawUUID + "|" + r.Text
		if seen[key] {
			continue
		}
		seen[key] = true
		if r.Project == "" {
			r.Project = "unassigned"
		}
		out = append(out, r)
	}
	return out
}

func queryToolTrajectoryPushbacks(db *sql.DB, holdoutRatio float64) []TrajectoryRecord {
	rows, err := db.Query(`
		SELECT tc.session_id, COALESCE(s.agent, ''), COALESCE(s.project, ''),
		       COALESCE(tc.tool_name, ''), COALESCE(tc.category, ''), COALESCE(tc.file_path, ''),
		       m_asst.ordinal, COALESCE(m_user.content, ''), COALESCE(m_user.timestamp, COALESCE(s.started_at, ''))
		FROM tool_calls tc
		JOIN messages m_asst ON m_asst.id = tc.message_id
		JOIN sessions s ON s.id = tc.session_id
		JOIN messages m_user ON m_user.session_id = tc.session_id
		     AND m_user.ordinal = m_asst.ordinal + 1
		     AND m_user.role = 'user'
		WHERE (tc.category IN ('write', 'edit') OR tc.tool_name IN ('Write', 'Edit', 'MultiEdit', 'write', 'edit'))
		ORDER BY m_asst.id DESC
		LIMIT 400
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var out []TrajectoryRecord
	seen := map[string]bool{}
	for rows.Next() {
		var r TrajectoryRecord
		if err := rows.Scan(&r.SessionID, &r.Agent, &r.Project, &r.ToolName, &r.Category, &r.FilePath, &r.Ordinal, &r.UserReaction, &r.Timestamp); err != nil {
			continue
		}
		rawUUID := extractRawUUID(r.SessionID)
		if isHoldout(rawUUID, holdoutRatio) {
			continue
		}
		r.UserReaction = strings.TrimSpace(r.UserReaction)
		if len(r.UserReaction) < 15 || len(r.UserReaction) > 550 {
			continue
		}
		if strings.HasPrefix(r.UserReaction, "<") || strings.HasPrefix(r.UserReaction, "{") {
			continue
		}
		if !trajectoryPushbackRE.MatchString(r.UserReaction) {
			continue
		}
		key := fmt.Sprintf("%s|%d", rawUUID, r.Ordinal)
		if seen[key] {
			continue
		}
		seen[key] = true
		if r.Project == "" {
			r.Project = "unassigned"
		}
		out = append(out, r)
		if len(out) >= 30 {
			break
		}
	}
	return out
}

func queryRecallEntries(db *sql.DB) []RecallRecord {
	rows, err := db.Query(`
		SELECT id, COALESCE(type, ''), COALESCE(scope, ''), COALESCE(title, ''),
		       COALESCE(body, ''), COALESCE(trigger, ''), COALESCE(project, ''),
		       COALESCE(agent, ''), COALESCE(source_session_id, ''), COALESCE(transferable, 0)
		FROM recall_entries
		WHERE status = 'accepted' AND review_state != 'human_rejected'
		ORDER BY updated_at DESC
		LIMIT 100
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var out []RecallRecord
	for rows.Next() {
		var r RecallRecord
		var tr int
		if err := rows.Scan(&r.ID, &r.Type, &r.Scope, &r.Title, &r.Body, &r.Trigger, &r.Project, &r.Agent, &r.SourceSessionID, &tr); err != nil {
			continue
		}
		r.Transferable = tr != 0
		out = append(out, r)
	}
	return out
}
