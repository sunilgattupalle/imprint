package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type OwnerConfig struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	GithubLogin string `yaml:"github_login,omitempty"`
}

type SessionArchiveConfig struct {
	SQLitePath string `yaml:"sqlite_path"`
}

type ChatGPTExportsConfig struct {
	DefaultProject string   `yaml:"default_project,omitempty"`
	Globs          []string `yaml:"globs,omitempty"`
}

type ConfluencePageRef struct {
	ID    string `yaml:"id"`
	Title string `yaml:"title"`
}

type AuthoredArtifactsConfig struct {
	Globs           []string            `yaml:"globs,omitempty"`
	ConfluencePages []ConfluencePageRef `yaml:"confluence_pages,omitempty"`
}

type ExplicitPriorsConfig struct {
	Files []string `yaml:"files,omitempty"`
}

type NamedRegex struct {
	Name  string `yaml:"name"`
	Regex string `yaml:"regex"`
}

type HygieneConfig struct {
	HoldoutRatio            float64      `yaml:"holdout_ratio"`
	ExtraRejectPatterns     []NamedRegex `yaml:"extra_reject_patterns,omitempty"`
	ExtraInlineStripRegexes []string     `yaml:"extra_inline_strip_regexes,omitempty"`
}

type SourcesConfig struct {
	Owner                                   OwnerConfig             `yaml:"owner"`
	SessionArchive                          SessionArchiveConfig    `yaml:"session_archive"`
	ChatGPTExports                          ChatGPTExportsConfig    `yaml:"chatgpt_exports,omitempty"`
	AuthoredArtifacts                       AuthoredArtifactsConfig `yaml:"authored_artifacts,omitempty"`
	ExplicitPriors                          ExplicitPriorsConfig    `yaml:"explicit_priors,omitempty"`
	Hygiene                                 HygieneConfig           `yaml:"hygiene"`
	WorkspaceTermsDisallowedInPortableRules []string                `yaml:"workspace_terms_disallowed_in_portable_rules,omitempty"`
}

func ExpandPath(p string) string {
	if p == "" {
		return ""
	}
	if strings.HasPrefix(p, "~/") || p == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			if p == "~" {
				return home
			}
			return filepath.Join(home, p[2:])
		}
	}
	return p
}

func ResolveInstancePaths(instanceDir string) (absInstance string, localDir string, err error) {
	absInstance, err = filepath.Abs(ExpandPath(instanceDir))
	if err != nil {
		return "", "", err
	}
	localCandidate := filepath.Join(absInstance, ".local")
	if info, statErr := os.Stat(localCandidate); statErr == nil && info.IsDir() {
		return absInstance, localCandidate, nil
	}
	return absInstance, localCandidate, nil
}

func LoadSources(localDir string) (*SourcesConfig, error) {
	p := filepath.Join(localDir, "sources.yaml")
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var cfg SourcesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Hygiene.HoldoutRatio <= 0 {
		cfg.Hygiene.HoldoutRatio = 0.20
	}
	if cfg.SessionArchive.SQLitePath == "" {
		cfg.SessionArchive.SQLitePath = "~/.agentsview/sessions.db"
	}
	return &cfg, nil
}

func SaveSources(localDir string, cfg *SourcesConfig) error {
	if err := os.MkdirAll(localDir, 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(localDir, "sources.yaml"), data, 0o644)
}

func AppendUnique(slice []string, items ...string) []string {
	seen := make(map[string]bool, len(slice))
	for _, s := range slice {
		seen[s] = true
	}
	for _, it := range items {
		if it != "" && !seen[it] {
			seen[it] = true
			slice = append(slice, it)
		}
	}
	return slice
}
