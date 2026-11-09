package collector

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sunilgattupalle/imprint-model/internal/config"
	_ "modernc.org/sqlite"
)

const gitignoreTemplate = `# Raw session data, verbatim quote ledgers, and workspace bindings NEVER enter git
.local/
corpus/
*.jsonl
*.db
*evidence_ledger*
*demoted_non_imprint*
WORKSPACE.md
.DS_Store
`

// InitInstance scaffolds `<instance>` with the two-zone privacy layout, configures agentsview
// (including ~/.augment/sessions-v2 if present), labels Auggie sessions, and auto-discovers S3 priors.
func InitInstance(instanceDir string, ownerID, ownerName string, runSync bool) error {
	absInstance, localDir, err := config.ResolveInstancePaths(instanceDir)
	if err != nil {
		return err
	}

	dirs := []string{
		filepath.Join(absInstance, "overlays"),
		filepath.Join(localDir, "corpus"),
		filepath.Join(localDir, "authored"),
		filepath.Join(localDir, "eval"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}

	giPath := filepath.Join(absInstance, ".gitignore")
	if _, err := os.Stat(giPath); os.IsNotExist(err) {
		if err := os.WriteFile(giPath, []byte(gitignoreTemplate), 0o644); err != nil {
			return err
		}
	}

	// Ensure agentsview config wires ~/.augment/sessions-v2 if present
	if err := ensureAuggiePiDirs(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not update ~/.agentsview/config.toml: %v\n", err)
	}

	if runSync {
		if avBin, err := findAgentsView(); err == nil {
			cmd := exec.Command(avBin, "sync")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			_ = cmd.Run()
		}
	}

	dbPath := config.ExpandPath("~/.agentsview/sessions.db")
	if _, err := os.Stat(dbPath); err == nil {
		_ = labelAuggieSessions(dbPath)
	}

	// Create .local/sources.yaml if it does not exist yet
	sourcesPath := filepath.Join(localDir, "sources.yaml")
	if _, err := os.Stat(sourcesPath); os.IsNotExist(err) {
		if ownerID == "" {
			ownerID = filepath.Base(absInstance)
		}
		if ownerName == "" {
			ownerName = ownerID
		}
		priors := discoverExplicitPriors()
		cfg := &config.SourcesConfig{
			Owner: config.OwnerConfig{
				ID:   ownerID,
				Name: ownerName,
			},
			SessionArchive: config.SessionArchiveConfig{
				SQLitePath: "~/.agentsview/sessions.db",
			},
			ExplicitPriors: config.ExplicitPriorsConfig{
				Files: priors,
			},
			Hygiene: config.HygieneConfig{
				HoldoutRatio: 0.20,
				ExtraRejectPatterns: []config.NamedRegex{
					{Name: "cosmos_events", Regex: `^\s*<cosmos-events\b`},
				},
				ExtraInlineStripRegexes: []string{
					`<cosmos-events\b[^>]*>.*?</cosmos-events>`,
				},
			},
		}
		if err := config.SaveSources(localDir, cfg); err != nil {
			return err
		}
		fmt.Printf("Created %s (discovered %d explicit prior files)\n", sourcesPath, len(priors))
	} else {
		fmt.Printf("Using existing %s\n", sourcesPath)
	}

	fmt.Printf("Initialized two-zone imprint instance at %s\n", absInstance)
	fmt.Printf("  Tracked zone:     %s/{IMPRINT.md,imprint.json,overlays/}\n", absInstance)
	fmt.Printf("  Local-only zone:  %s/ (gitignored)\n", localDir)
	return nil
}

func findAgentsView() (string, error) {
	if p, err := exec.LookPath("agentsview"); err == nil {
		return p, nil
	}
	candidate := config.ExpandPath("~/.local/bin/agentsview")
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}
	return "", fmt.Errorf("agentsview binary not found")
}

func ensureAuggiePiDirs() error {
	auggieDir := config.ExpandPath("~/.augment/sessions-v2")
	if _, err := os.Stat(auggieDir); os.IsNotExist(err) {
		return nil
	}
	cfgPath := config.ExpandPath("~/.agentsview/config.toml")
	b, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil
	}
	content := string(b)
	if strings.Contains(content, ".augment/sessions-v2") {
		return nil
	}
	line := "pi_dirs = [\"~/.pi/agent/sessions\", \"~/.augment/sessions-v2\"]\n"
	return os.WriteFile(cfgPath, []byte(line+content), 0o644)
}

func labelAuggieSessions(dbPath string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(`
		INSERT OR IGNORE INTO session_labels (session_id, label, created_at)
		SELECT id, 'source=auggie', datetime('now')
		FROM sessions
		WHERE file_path LIKE '%/.augment/sessions-v2/%'
	`)
	return err
}

func discoverExplicitPriors() []string {
	candidates := []string{
		"~/.claude/CLAUDE.md",
		"~/personal-ai/writing-style/style.md",
		"~/personal-ai/presentation-style/style.md",
		"~/personal-ai/prototype-style/style.md",
	}
	var found []string
	for _, c := range candidates {
		if _, err := os.Stat(config.ExpandPath(c)); err == nil {
			found = append(found, c)
		}
	}
	return found
}
