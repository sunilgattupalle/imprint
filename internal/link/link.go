package link

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sunilgattupalle/imprint/internal/config"
)

const (
	markerStart = "<!-- imprint:start -->"
	markerEnd   = "<!-- imprint:end -->"
)

var managedBlockRE = regexp.MustCompile(`(?s)<!-- imprint:start -->.*?<!-- imprint:end -->\n?`)

type TargetResult struct {
	Harness string `json:"harness"`
	Path    string `json:"path"`
	Action  string `json:"action"`
}

type targetSpec struct {
	harness   string
	detectDir string
	filePath  string
}

func defaultTargets(homeDir string) []targetSpec {
	return []targetSpec{
		{
			harness:   "claude",
			detectDir: filepath.Join(homeDir, ".claude"),
			filePath:  filepath.Join(homeDir, ".claude", "CLAUDE.md"),
		},
		{
			harness:   "pi",
			detectDir: filepath.Join(homeDir, ".pi"),
			filePath:  filepath.Join(homeDir, ".pi", "agent", "AGENTS.md"),
		},
		{
			harness:   "omp",
			detectDir: filepath.Join(homeDir, ".omp"),
			filePath:  filepath.Join(homeDir, ".omp", "agent", "AGENTS.md"),
		},
		{
			harness:   "opencode",
			detectDir: filepath.Join(homeDir, ".config", "opencode"),
			filePath:  filepath.Join(homeDir, ".config", "opencode", "AGENTS.md"),
		},
		{
			harness:   "hermes",
			detectDir: filepath.Join(homeDir, ".hermes"),
			filePath:  filepath.Join(homeDir, ".hermes", "SOUL.md"),
		},
	}
}

// RenderManagedBlock builds the idempotent @-include block for IMPRINT.md and WORKSPACE.md.
func RenderManagedBlock(instanceDir string) (string, error) {
	absInstance, localDir, err := config.ResolveInstancePaths(instanceDir)
	if err != nil {
		return "", err
	}
	imprintRef := config.ToTildePath(filepath.Join(absInstance, "IMPRINT.md"))
	workspaceAbs := filepath.Join(localDir, "WORKSPACE.md")
	workspaceRef := config.ToTildePath(workspaceAbs)

	var b strings.Builder
	b.WriteString(markerStart + "\n")
	b.WriteString("# Personal Imprint\n")
	b.WriteString("@" + imprintRef + "\n")
	if _, err := os.Stat(workspaceAbs); err == nil {
		b.WriteString("@" + workspaceRef + "\n")
	}
	b.WriteString(markerEnd + "\n")
	return b.String(), nil
}

// UpsertManagedBlock inserts or replaces the managed imprint block in existing content.
func UpsertManagedBlock(existing, block string) (string, bool) {
	if managedBlockRE.MatchString(existing) {
		updated := managedBlockRE.ReplaceAllString(existing, block)
		return updated, updated != existing
	}
	trimmed := strings.TrimRight(existing, "\n")
	if trimmed == "" {
		return block, true
	}
	return trimmed + "\n\n" + block, true
}

// LinkInstance wires the imprint instance into all detected coding-agent harnesses under homeDir.
// If homeDir is empty, os.UserHomeDir() is used.
func LinkInstance(instanceDir string, homeDir string, dryRun bool) ([]TargetResult, error) {
	if homeDir == "" {
		var err error
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return nil, err
		}
	}
	block, err := RenderManagedBlock(instanceDir)
	if err != nil {
		return nil, err
	}

	var results []TargetResult
	for _, t := range defaultTargets(homeDir) {
		if info, err := os.Stat(t.detectDir); err != nil || !info.IsDir() {
			continue
		}
		existingBytes, readErr := os.ReadFile(t.filePath)
		existing := ""
		if readErr == nil {
			existing = string(existingBytes)
		}
		updated, changed := UpsertManagedBlock(existing, block)
		action := "unchanged"
		if readErr != nil && os.IsNotExist(readErr) {
			action = "created"
		} else if changed {
			action = "updated"
		}
		if !dryRun && (action == "created" || action == "updated") {
			if err := os.MkdirAll(filepath.Dir(t.filePath), 0o755); err != nil {
				return nil, fmt.Errorf("mkdir for %s: %w", t.filePath, err)
			}
			if err := os.WriteFile(t.filePath, []byte(updated), 0o644); err != nil {
				return nil, fmt.Errorf("write %s: %w", t.filePath, err)
			}
		}
		results = append(results, TargetResult{
			Harness: t.harness,
			Path:    config.ToTildePath(t.filePath),
			Action:  action,
		})
	}
	return results, nil
}
