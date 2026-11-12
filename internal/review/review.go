package review

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sunilgattupalle/imprint/internal/config"
	"github.com/sunilgattupalle/imprint/internal/corpus"
	"github.com/sunilgattupalle/imprint/internal/gate"
)

// ProjectMemory represents a project-scoped episodic decision, warning, or workflow rule
// conditioned on an Imprint dimension.
type ProjectMemory struct {
	ID             string   `json:"id"`
	Project        string   `json:"project"`
	Type           string   `json:"type"`
	Dimension      string   `json:"dimension,omitempty"`
	Trigger        string   `json:"trigger"`
	Rule           string   `json:"rule"`
	SourceSessions []string `json:"source_sessions,omitempty"`
}

func LoadProjectMemories(localDir string) ([]ProjectMemory, error) {
	p := filepath.Join(localDir, "memories.json")
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []ProjectMemory
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// RunReview updates any requested habit_reflection states (e.g. IMP-F1-01=observed_keep)
// and prints a summary of all L1 claims and L2 project memories.
func RunReview(instanceDir string, setPairs []string) error {
	_, localDir, err := config.ResolveInstancePaths(instanceDir)
	if err != nil {
		return err
	}
	ledgerPath := filepath.Join(localDir, "evidence_ledger.json")
	b, err := os.ReadFile(ledgerPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", ledgerPath, err)
	}
	var claims []gate.RawClaim
	if err := json.Unmarshal(b, &claims); err != nil {
		return err
	}

	if len(setPairs) > 0 {
		updates := map[string]string{}
		for _, pair := range setPairs {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid --set %q (expected <CLAIM_ID>=observed_keep|aspirational_shift|unreviewed)", pair)
			}
			id := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if val != "observed_keep" && val != "aspirational_shift" && val != "unreviewed" {
				return fmt.Errorf("invalid habit_reflection value %q for %s (must be observed_keep, aspirational_shift, or unreviewed)", val, id)
			}
			updates[id] = val
		}
		for i := range claims {
			if newState, ok := updates[claims[i].ID]; ok {
				claims[i].HabitReflection = newState
				delete(updates, claims[i].ID)
			}
		}
		if len(updates) > 0 {
			var unknown []string
			for k := range updates {
				unknown = append(unknown, k)
			}
			return fmt.Errorf("unknown claim ID(s): %s", strings.Join(unknown, ", "))
		}
		outBytes, err := json.MarshalIndent(claims, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(ledgerPath, append(outBytes, '\n'), 0o644); err != nil {
			return err
		}
		if err := gate.ValidateAndExport(instanceDir); err != nil {
			return err
		}
	}

	fmt.Printf("L1 Portable Imprint Claims (%d):\n", len(claims))
	fmt.Printf("%-10s | %-28s | %-18s | %-5s | %s\n", "ID", "FACET", "HABIT REFLECTION", "EVID", "TITLE")
	fmt.Println(strings.Repeat("-", 105))
	for _, c := range claims {
		hr := c.HabitReflection
		if hr == "" {
			hr = "unreviewed"
		}
		fmt.Printf("%-10s | %-28s | %-18s | %-5d | %s\n", c.ID, c.Facet, hr, len(c.Evidence), c.Title)
	}

	memories, _ := LoadProjectMemories(localDir)
	_, _, recalls := corpus.ExtractBehavioralSignals(instanceDir)
	if len(memories) > 0 || len(recalls) > 0 {
		fmt.Printf("\nL2 Project Memories (%d in .local/memories.json, %d in recall_entries):\n", len(memories), len(recalls))
		for _, m := range memories {
			fmt.Printf("  - [%s | project=%s | dim=%s] %s: %s\n", m.ID, m.Project, m.Dimension, m.Trigger, m.Rule)
		}
		for _, r := range recalls {
			fmt.Printf("  - [recall:%s | project=%s | type=%s] %s: %s\n", r.ID, r.Project, r.Type, r.Title, r.Body)
		}
	}
	return nil
}
