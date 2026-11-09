package ingest

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/sunilgattupalle/imprint-model/internal/config"
)

type ExtractedDoc struct {
	Path      string
	Format    string
	Text      string
	CSSTokens []string
}

var (
	styleBlockRE = regexp.MustCompile(`(?is)<style[^>]*>(.*?)</style>`)
	scriptRE     = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	htmlTagRE    = regexp.MustCompile(`(?s)<[^>]+>`)
	cssVarRE     = regexp.MustCompile(`--[a-zA-Z0-9_-]+\s*:\s*[^;\}]+`)
	multiSpaceRE = regexp.MustCompile(`[ \t]{2,}`)
	multiLineRE  = regexp.MustCompile(`\n{3,}`)
)

// ExtractDocument reads .md, .txt, .docx, .pptx, or .html and returns clean text + CSS tokens.
func ExtractDocument(path string) (*ExtractedDoc, error) {
	abs := config.ExpandPath(path)
	ext := strings.ToLower(filepath.Ext(abs))
	switch ext {
	case ".md", ".txt", ".mdc", ".cursorrules":
		b, err := os.ReadFile(abs)
		if err != nil {
			return nil, err
		}
		return &ExtractedDoc{Path: abs, Format: ext, Text: strings.TrimSpace(string(b))}, nil
	case ".docx":
		txt, err := extractZipXMLText(abs, func(name string) bool {
			return name == "word/document.xml"
		})
		if err != nil {
			return nil, err
		}
		return &ExtractedDoc{Path: abs, Format: ".docx", Text: txt}, nil
	case ".pptx":
		txt, err := extractZipXMLText(abs, func(name string) bool {
			return strings.HasPrefix(name, "ppt/slides/slide") && strings.HasSuffix(name, ".xml")
		})
		if err != nil {
			return nil, err
		}
		return &ExtractedDoc{Path: abs, Format: ".pptx", Text: txt}, nil
	case ".html", ".htm":
		b, err := os.ReadFile(abs)
		if err != nil {
			return nil, err
		}
		raw := string(b)
		var tokens []string
		seen := map[string]bool{}
		for _, m := range styleBlockRE.FindAllStringSubmatch(raw, -1) {
			for _, tok := range cssVarRE.FindAllString(m[1], -1) {
				tok = strings.TrimSpace(tok)
				if !seen[tok] {
					seen[tok] = true
					tokens = append(tokens, tok)
				}
			}
		}
		noStyle := styleBlockRE.ReplaceAllString(raw, "\n")
		noScript := scriptRE.ReplaceAllString(noStyle, "\n")
		plain := htmlTagRE.ReplaceAllString(noScript, " ")
		plain = multiSpaceRE.ReplaceAllString(plain, " ")
		plain = multiLineRE.ReplaceAllString(plain, "\n\n")
		return &ExtractedDoc{
			Path:      abs,
			Format:    ".html",
			Text:      strings.TrimSpace(plain),
			CSSTokens: tokens,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported document extension %q for %s", ext, abs)
	}
}

func extractZipXMLText(zipPath string, matchEntry func(string) bool) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	var files []*zip.File
	for _, f := range r.File {
		if matchEntry(f.Name) {
			files = append(files, f)
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })

	var out bytes.Buffer
	for _, f := range files {
		rc, err := f.Open()
		if err != nil {
			continue
		}
		dec := xml.NewDecoder(rc)
		var para strings.Builder
		for {
			tok, err := dec.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}
			switch t := tok.(type) {
			case xml.CharData:
				s := strings.TrimSpace(string(t))
				if s != "" {
					if para.Len() > 0 {
						para.WriteString(" ")
					}
					para.WriteString(s)
				}
			case xml.EndElement:
				if t.Name.Local == "p" || t.Name.Local == "tr" || t.Name.Local == "sld" {
					if para.Len() > 0 {
						out.WriteString(para.String())
						out.WriteString("\n")
						para.Reset()
					}
				}
			}
		}
		if para.Len() > 0 {
			out.WriteString(para.String())
			out.WriteString("\n")
		}
		rc.Close()
	}
	return strings.TrimSpace(out.String()), nil
}

// AddSources updates `<instance>/.local/sources.yaml` with new authored globs, chat exports, priors, or disallowed terms.
func AddSources(instanceDir string, authored, chats, priors, disallowTerms []string) error {
	_, localDir, err := config.ResolveInstancePaths(instanceDir)
	if err != nil {
		return err
	}
	cfg, err := config.LoadSources(localDir)
	if err != nil {
		return fmt.Errorf("failed to load %s/sources.yaml (run `imprint init` first): %w", localDir, err)
	}
	cfg.AuthoredArtifacts.Globs = config.AppendUnique(cfg.AuthoredArtifacts.Globs, authored...)
	cfg.ChatGPTExports.Globs = config.AppendUnique(cfg.ChatGPTExports.Globs, chats...)
	cfg.ExplicitPriors.Files = config.AppendUnique(cfg.ExplicitPriors.Files, priors...)
	cfg.WorkspaceTermsDisallowedInPortableRules = config.AppendUnique(
		cfg.WorkspaceTermsDisallowedInPortableRules,
		disallowTerms...,
	)
	return config.SaveSources(localDir, cfg)
}
