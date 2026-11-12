package collector

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sunilgattupalle/imprint/internal/config"
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
		autoID, autoName := discoverOwner(absInstance)
		if ownerID == "" {
			ownerID = autoID
		}
		if ownerName == "" {
			ownerName = autoName
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
					{Name: "event_bus_envelope", Regex: `^\s*<[a-z0-9_-]+-events\b`},
				},
				ExtraInlineStripRegexes: []string{
					`<[a-z0-9_-]+-events\b[^>]*>.*?</[a-z0-9_-]+-events>`,
				},
			},
		}
		if err := config.SaveSources(localDir, cfg); err != nil {
			return err
		}
		fmt.Printf("Created %s (owner=%s, discovered %d explicit prior files)\n", sourcesPath, ownerID, len(priors))
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

var piDirsLineRE = regexp.MustCompile(`(?m)^pi_dirs\s*=\s*\[[^\]]*\]`)

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
	line := `pi_dirs = ["~/.pi/agent/sessions", "~/.augment/sessions-v2"]`
	if piDirsLineRE.MatchString(content) {
		updated := piDirsLineRE.ReplaceAllString(content, line)
		return os.WriteFile(cfgPath, []byte(updated), 0o644)
	}
	return os.WriteFile(cfgPath, []byte(line+"\n"+content), 0o644)
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

func discoverOwner(absInstance string) (id string, name string) {
	if u, err := user.Current(); err == nil {
		if u.Username != "" {
			id = strings.ToLower(strings.ReplaceAll(u.Username, ".", "-"))
		}
		if u.Name != "" {
			name = u.Name
		}
	}
	if out, err := exec.Command("git", "config", "user.name").Output(); err == nil {
		if gName := strings.TrimSpace(string(out)); gName != "" {
			name = gName
		}
	}
	if id == "" {
		base := strings.TrimPrefix(filepath.Base(absInstance), ".")
		if base == "" || base == "imprint" {
			id = "builder"
		} else {
			id = base
		}
	}
	if name == "" {
		name = id
	}
	return id, name
}

func discoverExplicitPriors() []string {
	staticCandidates := []string{
		"~/.claude/CLAUDE.md",
		"~/.pi/agent/AGENTS.md",
		"~/.omp/agent/AGENTS.md",
		"~/.config/opencode/AGENTS.md",
	}
	var found []string
	seen := map[string]bool{}
	for _, c := range staticCandidates {
		if _, err := os.Stat(config.ExpandPath(c)); err == nil {
			seen[c] = true
			found = append(found, c)
		}
	}
	styleMatches, _ := filepath.Glob(config.ExpandPath("~/personal-ai/*/style.md"))
	for _, m := range styleMatches {
		tilde := config.ToTildePath(m)
		if !seen[tilde] {
			seen[tilde] = true
			found = append(found, tilde)
		}
	}
	return found
}
