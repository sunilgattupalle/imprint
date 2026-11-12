package brief

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sunilgattupalle/imprint/internal/config"
	"github.com/sunilgattupalle/imprint/internal/corpus"
	"github.com/sunilgattupalle/imprint/internal/review"
)

type ModeRule struct {
	ModeID      string
	OverlayFile string
	IntentRE    *regexp.Regexp
	FileRE      *regexp.Regexp
}

var modeRules = []ModeRule{
	{
		ModeID:      "architecture-and-specs",
		OverlayFile: "mode-architecture-and-specs.md",
		IntentRE:    regexp.MustCompile(`(?i)\b(architect|spec|specs|specification|design|rfc|readme|ontology|mental model|construct|primitive|evaluat|compare|comparison|rubric|think through|diligence)\b`),
		FileRE:      regexp.MustCompile(`(?i)(^|/)(README\.md|AGENTS\.md|SKILL\.md|CLAUDE\.md|docs/.*\.md|.*spec.*\.md|.*design.*\.md|schemas/.*\.json)$`),
	},
	{
		ModeID:      "exec-memos-and-decks",
		OverlayFile: "mode-exec-memos-and-decks.md",
		IntentRE:    regexp.MustCompile(`(?i)\b(deck|slides?|pptx|memo|executive|board|one-?pager|leadership|presentation|narrative)\b`),
		FileRE:      regexp.MustCompile(`(?i)\.(pptx|key|marp\.md)$|(?i)(deck|slides|memo|one-pager)`),
	},
	{
		ModeID:      "visual-html-and-prototypes",
		OverlayFile: "mode-visual-html-and-prototypes.md",
		IntentRE:    regexp.MustCompile(`(?i)\b(html|prototype|ui|visual|diagram|svg|palette|color|layout|css|chart|typography|component)\b`),
		FileRE:      regexp.MustCompile(`(?i)\.(html|htm|css|svg|svelte|tsx|jsx)$`),
	},
	{
		ModeID:      "autonomous-orchestration",
		OverlayFile: "mode-autonomous-orchestration.md",
		IntentRE:    regexp.MustCompile(`(?i)\b(subagent|parallel|background|worktree|autonom|orchestrat|batch|fan[- ]out|webhook|event[- ]driven)\b`),
		FileRE:      regexp.MustCompile(`(?i)(workflow|pipeline|orchestrat)`),
	},
	{
		ModeID:      "coding-and-prs",
		OverlayFile: "mode-coding-and-prs.md",
		IntentRE:    regexp.MustCompile(`(?i)\b(implement|fix|bug|test|unit test|commit|push|pull request|\bpr\b|refactor|build|compile)\b`),
		FileRE:      regexp.MustCompile(`(?i)\.(go|py|ts|js|java|rs|c|cc|cpp|h|rb|sh|sql)$`),
	},
}

type BriefResult struct {
	InstanceDir     string   `json:"instance_dir"`
	Project         string   `json:"project"`
	MatchedModes    []string `json:"matched_modes"`
	IncludedFiles   []string `json:"included_files"`
	MatchedMemories int      `json:"matched_memories"`
	Markdown        string   `json:"markdown"`
}

// ClassifyModes deterministically selects up to 2 mode overlays matching the prompt and target files.
func ClassifyModes(prompt string, files []string) []ModeRule {
	type scored struct {
		rule  ModeRule
		score int
	}
	var matches []scored
	for _, r := range modeRules {
		s := 0
		if prompt != "" && r.IntentRE.MatchString(prompt) {
			s += 2
		}
		for _, f := range files {
			f = strings.TrimSpace(f)
			if f != "" && r.FileRE.MatchString(filepath.ToSlash(f)) {
				s += 3
			}
		}
		if s > 0 {
			matches = append(matches, scored{rule: r, score: s})
		}
	}
	// Sort highest score first, preserving declaration order on ties
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].score > matches[i].score {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}
	var out []ModeRule
	for i := 0; i < len(matches) && i < 2; i++ {
		out = append(out, matches[i].rule)
	}
	return out
}

func resolveProjectName(cwd string) string {
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	if cwd == "" {
		return ""
	}
	cmd := exec.Command("git", "-C", cwd, "rev-parse", "--show-toplevel")
	if out, err := cmd.Output(); err == nil {
		top := strings.TrimSpace(string(out))
		if top != "" {
			return filepath.Base(top)
		}
	}
	return filepath.Base(filepath.Clean(cwd))
}

// BuildBrief compiles a situational operating brief combining IMPRINT.md,
// the matched overlays/mode-*.md, and repo-matched L2 workspace bindings/memories.
func BuildBrief(instanceDir, cwd, prompt string, files []string) (*BriefResult, error) {
	absInstance, localDir, err := config.ResolveInstancePaths(instanceDir)
	if err != nil {
		return nil, err
	}
	proj := resolveProjectName(cwd)
	matchedRules := ClassifyModes(prompt, files)

	var md strings.Builder
	var included []string
	var modeIDs []string

	imprintPath := filepath.Join(absInstance, "IMPRINT.md")
	if b, err := os.ReadFile(imprintPath); err == nil {
		md.WriteString(strings.TrimSpace(string(b)) + "\n\n")
		included = append(included, config.ToTildePath(imprintPath))
	}

	for _, r := range matchedRules {
		modeIDs = append(modeIDs, r.ModeID)
		ovPath := filepath.Join(absInstance, "overlays", r.OverlayFile)
		if b, err := os.ReadFile(ovPath); err == nil {
			md.WriteString("---\n\n")
			md.WriteString(strings.TrimSpace(string(b)) + "\n\n")
			included = append(included, config.ToTildePath(ovPath))
		}
	}

	wsPath := filepath.Join(localDir, "WORKSPACE.md")
	if b, err := os.ReadFile(wsPath); err == nil {
		md.WriteString("---\n\n")
		md.WriteString(strings.TrimSpace(string(b)) + "\n\n")
		included = append(included, config.ToTildePath(wsPath))
	}

	memories, _ := review.LoadProjectMemories(localDir)
	_, _, recalls := corpus.ExtractBehavioralSignals(instanceDir)

	var matchedMemLines []string
	for _, m := range memories {
		if m.Project == "" || m.Project == "*" || strings.EqualFold(m.Project, proj) {
			matchedMemLines = append(matchedMemLines, fmt.Sprintf("- **[%s | %s]** (*%s*): %s", m.ID, m.Dimension, m.Trigger, m.Rule))
		}
	}
	for _, r := range recalls {
		if r.Scope == "global" || strings.EqualFold(r.Project, proj) {
			matchedMemLines = append(matchedMemLines, fmt.Sprintf("- **[recall:%s | %s]** **%s**: %s", r.ID, r.Type, r.Title, r.Body))
		}
	}
	if len(matchedMemLines) > 0 {
		md.WriteString("---\n\n")
		md.WriteString(fmt.Sprintf("## Active Project Episodic Memories (`%s`)\n\n", proj))
		for _, line := range matchedMemLines {
			md.WriteString(line + "\n")
		}
	}

	return &BriefResult{
		InstanceDir:     config.ToTildePath(absInstance),
		Project:         proj,
		MatchedModes:    modeIDs,
		IncludedFiles:   included,
		MatchedMemories: len(matchedMemLines),
		Markdown:        strings.TrimSpace(md.String()) + "\n",
	}, nil
}
