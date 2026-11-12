package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sunilgattupalle/imprint/internal/brief"
	"github.com/sunilgattupalle/imprint/internal/collector"
	"github.com/sunilgattupalle/imprint/internal/config"
	"github.com/sunilgattupalle/imprint/internal/digest"
	"github.com/sunilgattupalle/imprint/internal/eval"
	"github.com/sunilgattupalle/imprint/internal/gate"
	"github.com/sunilgattupalle/imprint/internal/ingest"
	"github.com/sunilgattupalle/imprint/internal/link"
	"github.com/sunilgattupalle/imprint/internal/review"
)

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, `imprint: First-class personal cognitive, behavioral, and voice imprint CLI (default home: ~/.imprint)

Usage:
  imprint sync     [instance_dir] [--owner-id <id>] [--owner-name <name>] [--link]
  imprint brief    [instance_dir] [--cwd <dir>] [--prompt <text>] [--files <csv>] [--json]
  imprint review   [instance_dir] [--set <CLAIM_ID>=observed_keep|aspirational_shift]
  imprint add      [instance_dir] [--pre-ai <glob/path>] [--authored <glob/path>] [--chats <glob/path>] [--prior <path>] [--disallow-term <term>]
  imprint link     [instance_dir] [--dry-run]
  imprint validate [instance_dir]
  imprint eval     [instance_dir]
  imprint init     [instance_dir] [--owner-id <id>] [--owner-name <name>] [--sync]
  imprint digest   [instance_dir]
`)
	os.Exit(2)
}

// parseFlagsWithPositional allows flags to appear either before or after the optional positional instance_dir.
func parseFlagsWithPositional(fs *flag.FlagSet, args []string, instanceFlag *string) string {
	var flagArgs []string
	var positional string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-") {
			flagArgs = append(flagArgs, a)
			name := strings.TrimLeft(a, "-")
			if !strings.Contains(name, "=") {
				if f := fs.Lookup(name); f != nil {
					if bf, ok := f.Value.(interface{ IsBoolFlag() bool }); !ok || !bf.IsBoolFlag() {
						if i+1 < len(args) {
							i++
							flagArgs = append(flagArgs, args[i])
						}
					}
				}
			}
		} else if positional == "" {
			positional = a
		} else {
			flagArgs = append(flagArgs, a)
		}
	}
	_ = fs.Parse(flagArgs)
	if *instanceFlag != "" {
		return *instanceFlag
	}
	if positional != "" {
		return positional
	}
	if fs.NArg() > 0 {
		return fs.Arg(0)
	}
	return config.DefaultInstanceDir
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	sub := os.Args[1]
	switch sub {
	case "sync":
		fs := flag.NewFlagSet("sync", flag.ExitOnError)
		ownerID := fs.String("owner-id", "", "Short owner identifier (auto-discovered if omitted)")
		ownerName := fs.String("owner-name", "", "Full owner display name (auto-discovered if omitted)")
		doLink := fs.Bool("link", true, "Automatically link IMPRINT.md into detected coding harnesses if present")
		instanceFlag := fs.String("instance", "", "Instance directory path (defaults to ~/.imprint)")
		inst := parseFlagsWithPositional(fs, os.Args[2:], instanceFlag)
		if err := collector.InitInstance(inst, *ownerID, *ownerName, true); err != nil {
			fatal(err)
		}
		digestPath, stats, err := digest.BuildDigest(inst)
		if err != nil {
			fatal(err)
		}
		b, _ := json.MarshalIndent(stats, "", "  ")
		fmt.Println(string(b))
		fmt.Printf("Wrote synthesis digest -> %s\n", digestPath)
		if *doLink {
			if linked, err := link.LinkInstance(inst, "", false); err == nil && len(linked) > 0 {
				for _, r := range linked {
					fmt.Printf("  Linked [%s] %s (%s)\n", r.Harness, r.Path, r.Action)
				}
			}
		}

	case "init":
		fs := flag.NewFlagSet("init", flag.ExitOnError)
		ownerID := fs.String("owner-id", "", "Short owner identifier")
		ownerName := fs.String("owner-name", "", "Full owner display name")
		syncDB := fs.Bool("sync", false, "Run agentsview sync before initializing")
		instanceFlag := fs.String("instance", "", "Instance directory path (defaults to ~/.imprint)")
		inst := parseFlagsWithPositional(fs, os.Args[2:], instanceFlag)
		if err := collector.InitInstance(inst, *ownerID, *ownerName, *syncDB); err != nil {
			fatal(err)
		}

	case "add":
		fs := flag.NewFlagSet("add", flag.ExitOnError)
		var preAI, authored, chats, priors, disallow multiFlag
		fs.Var(&preAI, "pre-ai", "Pre-AI / pure-human authored gold manuscript path or glob (.md, .txt, .docx, .pptx, .html)")
		fs.Var(&authored, "authored", "Authored document/deck path or glob (.md, .txt, .docx, .pptx, .html)")
		fs.Var(&chats, "chats", "Exported web chat JSON path or glob")
		fs.Var(&priors, "prior", "Explicit prior file (CLAUDE.md, AGENTS.md, style.md)")
		fs.Var(&disallow, "disallow-term", "Workspace/company term disallowed in portable imprint rules")
		instanceFlag := fs.String("instance", "", "Instance directory path (defaults to ~/.imprint)")
		inst := parseFlagsWithPositional(fs, os.Args[2:], instanceFlag)
		if err := ingest.AddSources(inst, preAI, authored, chats, priors, disallow); err != nil {
			fatal(err)
		}
		absInst, _, _ := config.ResolveInstancePaths(inst)
		fmt.Printf("Updated %s/.local/sources.yaml\n", config.ToTildePath(absInst))

	case "digest":
		fs := flag.NewFlagSet("digest", flag.ExitOnError)
		instanceFlag := fs.String("instance", "", "Instance directory path (defaults to ~/.imprint)")
		inst := parseFlagsWithPositional(fs, os.Args[2:], instanceFlag)
		digestPath, stats, err := digest.BuildDigest(inst)
		if err != nil {
			fatal(err)
		}
		b, _ := json.MarshalIndent(stats, "", "  ")
		fmt.Println(string(b))
		fmt.Printf("Wrote synthesis digest -> %s\n", digestPath)

	case "link":
		fs := flag.NewFlagSet("link", flag.ExitOnError)
		dryRun := fs.Bool("dry-run", false, "Preview harness files that would be linked without writing")
		instanceFlag := fs.String("instance", "", "Instance directory path (defaults to ~/.imprint)")
		inst := parseFlagsWithPositional(fs, os.Args[2:], instanceFlag)
		results, err := link.LinkInstance(inst, "", *dryRun)
		if err != nil {
			fatal(err)
		}
		for _, r := range results {
			fmt.Printf("Linked [%s] %s (%s)\n", r.Harness, r.Path, r.Action)
		}

	case "validate":
		fs := flag.NewFlagSet("validate", flag.ExitOnError)
		doLink := fs.Bool("link", true, "Automatically link IMPRINT.md into detected coding harnesses after validation")
		instanceFlag := fs.String("instance", "", "Instance directory path (defaults to ~/.imprint)")
		inst := parseFlagsWithPositional(fs, os.Args[2:], instanceFlag)
		if err := gate.ValidateAndExport(inst); err != nil {
			fatal(err)
		}
		if *doLink {
			if results, err := link.LinkInstance(inst, "", false); err == nil {
				for _, r := range results {
					fmt.Printf("  Linked [%s] %s (%s)\n", r.Harness, r.Path, r.Action)
				}
			}
		}

	case "brief":
		fs := flag.NewFlagSet("brief", flag.ExitOnError)
		cwd := fs.String("cwd", "", "Working directory to resolve active project (defaults to current directory)")
		prompt := fs.String("prompt", "", "Active user prompt or task description for situational mode routing")
		filesCSV := fs.String("files", "", "Comma-separated target file paths for situational mode routing")
		asJSON := fs.Bool("json", false, "Output structured JSON instead of compiled Markdown")
		instanceFlag := fs.String("instance", "", "Instance directory path (defaults to ~/.imprint)")
		inst := parseFlagsWithPositional(fs, os.Args[2:], instanceFlag)
		var files []string
		if *filesCSV != "" {
			files = strings.Split(*filesCSV, ",")
		}
		res, err := brief.BuildBrief(inst, *cwd, *prompt, files)
		if err != nil {
			fatal(err)
		}
		if *asJSON {
			b, _ := json.MarshalIndent(res, "", "  ")
			fmt.Println(string(b))
		} else {
			fmt.Print(res.Markdown)
		}

	case "review":
		fs := flag.NewFlagSet("review", flag.ExitOnError)
		var setPairs multiFlag
		fs.Var(&setPairs, "set", "Set habit_reflection (<CLAIM_ID>=observed_keep|aspirational_shift|unreviewed)")
		instanceFlag := fs.String("instance", "", "Instance directory path (defaults to ~/.imprint)")
		inst := parseFlagsWithPositional(fs, os.Args[2:], instanceFlag)
		if err := review.RunReview(inst, setPairs); err != nil {
			fatal(err)
		}

	case "eval":
		fs := flag.NewFlagSet("eval", flag.ExitOnError)
		instanceFlag := fs.String("instance", "", "Instance directory path (defaults to ~/.imprint)")
		inst := parseFlagsWithPositional(fs, os.Args[2:], instanceFlag)
		if _, err := eval.RunHoldoutEval(inst); err != nil {
			fatal(err)
		}

	default:
		usage()
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
