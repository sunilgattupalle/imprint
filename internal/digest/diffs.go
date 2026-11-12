package digest

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/sunilgattupalle/imprint/internal/ingest"
)

// VersionDiffSummary captures concrete structural and stylistic changes between
// two sequential revisions of an authored document (e.g. -v1 -> -v2).
type VersionDiffSummary struct {
	FromFile        string   `json:"from_file"`
	ToFile          string   `json:"to_file"`
	BeforeWords     int      `json:"before_words"`
	AfterWords      int      `json:"after_words"`
	BeforeEmDashes  int      `json:"before_em_dashes"`
	AfterEmDashes   int      `json:"after_em_dashes"`
	RemovedHeadings []string `json:"removed_headings,omitempty"`
	AddedHeadings   []string `json:"added_headings,omitempty"`
	SampleRemoved   []string `json:"sample_removed,omitempty"`
	SampleAdded     []string `json:"sample_added,omitempty"`
}

var versionFileRE = regexp.MustCompile(`(?i)^(.*?)[-_]v(\d+)(\.[a-z0-9]+)$`)

type versionedEntry struct {
	stem    string
	ext     string
	version int
	path    string
}

// ExtractVersionDiffs finds versioned sibling pairs (e.g. spec-v1.md -> spec-v2.md)
// among authored paths and computes a compact stylometric and structural delta.
func ExtractVersionDiffs(paths []string) []VersionDiffSummary {
	byGroup := map[string][]versionedEntry{}
	for _, p := range paths {
		base := filepath.Base(p)
		m := versionFileRE.FindStringSubmatch(base)
		if m == nil {
			continue
		}
		var ver int
		_, _ = fmt.Sscanf(m[2], "%d", &ver)
		key := filepath.Join(filepath.Dir(p), strings.ToLower(m[1])+strings.ToLower(m[3]))
		byGroup[key] = append(byGroup[key], versionedEntry{
			stem:    m[1],
			ext:     m[3],
			version: ver,
			path:    p,
		})
	}

	var keys []string
	for k, group := range byGroup {
		if len(group) >= 2 {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var out []VersionDiffSummary
	for _, k := range keys {
		group := byGroup[k]
		sort.Slice(group, func(i, j int) bool { return group[i].version < group[j].version })
		for i := 0; i+1 < len(group); i++ {
			d, ok := diffDocumentPair(group[i].path, group[i+1].path)
			if ok {
				out = append(out, d)
			}
		}
	}
	return out
}

func diffDocumentPair(fromPath, toPath string) (VersionDiffSummary, bool) {
	docA, errA := ingest.ExtractDocument(fromPath)
	docB, errB := ingest.ExtractDocument(toPath)
	if errA != nil || errB != nil {
		return VersionDiffSummary{}, false
	}

	wordsA := len(strings.Fields(docA.Text))
	wordsB := len(strings.Fields(docB.Text))
	emA := strings.Count(docA.Text, "—")
	emB := strings.Count(docB.Text, "—")

	linesA := nonEmptyLines(docA.Text)
	linesB := nonEmptyLines(docB.Text)
	setA := map[string]bool{}
	setB := map[string]bool{}
	for _, l := range linesA {
		setA[l] = true
	}
	for _, l := range linesB {
		setB[l] = true
	}

	var removedHeadings, addedHeadings []string
	var sampleRemoved, sampleAdded []string

	for _, l := range linesA {
		if setB[l] {
			continue
		}
		if strings.HasPrefix(l, "#") {
			if len(removedHeadings) < 5 {
				removedHeadings = append(removedHeadings, l)
			}
		} else if len(l) >= 30 && len(l) <= 220 && len(sampleRemoved) < 4 {
			sampleRemoved = append(sampleRemoved, l)
		}
	}
	for _, l := range linesB {
		if setA[l] {
			continue
		}
		if strings.HasPrefix(l, "#") {
			if len(addedHeadings) < 5 {
				addedHeadings = append(addedHeadings, l)
			}
		} else if len(l) >= 30 && len(l) <= 220 && len(sampleAdded) < 4 {
			sampleAdded = append(sampleAdded, l)
		}
	}

	return VersionDiffSummary{
		FromFile:        filepath.Base(fromPath),
		ToFile:          filepath.Base(toPath),
		BeforeWords:     wordsA,
		AfterWords:      wordsB,
		BeforeEmDashes:  emA,
		AfterEmDashes:   emB,
		RemovedHeadings: removedHeadings,
		AddedHeadings:   addedHeadings,
		SampleRemoved:   sampleRemoved,
		SampleAdded:     sampleAdded,
	}, true
}

func nonEmptyLines(s string) []string {
	raw := strings.Split(s, "\n")
	var out []string
	for _, r := range raw {
		t := strings.TrimSpace(r)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}
