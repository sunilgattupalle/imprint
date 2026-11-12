package review

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sunilgattupalle/imprint/internal/config"
	"github.com/sunilgattupalle/imprint/internal/gate"
)

// RunReview updates any requested habit_reflection states (e.g. IMP-F1-01=observed_keep)
// and prints a summary of all L1 portable imprint claims and their situational triggers.
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
	fmt.Printf("%-10s | %-28s | %-18s | %-5s | %s\n", "ID", "FACET", "REFLECTION", "EVID", "TITLE")
	fmt.Println(strings.Repeat("-", 105))
	for _, c := range claims {
		hr := c.HabitReflection
		if hr == "" {
			hr = "unreviewed"
		}
		fmt.Printf("%-10s | %-28s | %-18s | %-5d | %s\n", c.ID, c.Facet, hr, len(c.Evidence), c.Title)
	}
	return nil
}
