package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sunilgattupalle/imprint-model/internal/collector"
	"github.com/sunilgattupalle/imprint-model/internal/digest"
	"github.com/sunilgattupalle/imprint-model/internal/eval"
	"github.com/sunilgattupalle/imprint-model/internal/gate"
	"github.com/sunilgattupalle/imprint-model/internal/ingest"
)

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, `imprint: Local-first personal imprint collector, distiller, and validator

Usage:
  imprint init     <instance_dir> [--owner-id <id>] [--owner-name <name>] [--sync]
  imprint add      <instance_dir> [--authored <glob/path>] [--chats <glob/path>] [--prior <path>] [--disallow-term <term>]
  imprint digest   <instance_dir>
  imprint validate <instance_dir>
  imprint eval     <instance_dir>
`)
	os.Exit(2)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	sub := os.Args[1]
	switch sub {
	case "init":
		fs := flag.NewFlagSet("init", flag.ExitOnError)
		ownerID := fs.String("owner-id", "", "Short owner identifier")
		ownerName := fs.String("owner-name", "", "Full owner display name")
		syncDB := fs.Bool("sync", false, "Run agentsview sync before initializing")
		instanceFlag := fs.String("instance", "", "Instance directory path")
		_ = fs.Parse(os.Args[2:])
		inst := *instanceFlag
		if inst == "" && fs.NArg() > 0 {
			inst = fs.Arg(0)
		}
		if inst == "" {
			usage()
		}
		if err := collector.InitInstance(inst, *ownerID, *ownerName, *syncDB); err != nil {
			fatal(err)
		}

	case "add":
		fs := flag.NewFlagSet("add", flag.ExitOnError)
		var authored, chats, priors, disallow multiFlag
		fs.Var(&authored, "authored", "Authored document/deck path or glob (.md, .txt, .docx, .pptx, .html)")
		fs.Var(&chats, "chats", "Exported web chat JSON path or glob")
		fs.Var(&priors, "prior", "Explicit prior file (CLAUDE.md, AGENTS.md, style.md)")
		fs.Var(&disallow, "disallow-term", "Workspace/company term disallowed in portable imprint rules")
		instanceFlag := fs.String("instance", "", "Instance directory path")
		_ = fs.Parse(os.Args[2:])
		inst := *instanceFlag
		if inst == "" && fs.NArg() > 0 {
			inst = fs.Arg(0)
		}
		if inst == "" {
			usage()
		}
		if err := ingest.AddSources(inst, authored, chats, priors, disallow); err != nil {
			fatal(err)
		}
		fmt.Printf("Updated %s/.local/sources.yaml\n", inst)

	case "digest":
		fs := flag.NewFlagSet("digest", flag.ExitOnError)
		instanceFlag := fs.String("instance", "", "Instance directory path")
		_ = fs.Parse(os.Args[2:])
		inst := *instanceFlag
		if inst == "" && fs.NArg() > 0 {
			inst = fs.Arg(0)
		}
		if inst == "" {
			usage()
		}
		digestPath, stats, err := digest.BuildDigest(inst)
		if err != nil {
			fatal(err)
		}
		b, _ := json.MarshalIndent(stats, "", "  ")
		fmt.Println(string(b))
		fmt.Printf("Wrote synthesis digest -> %s\n", digestPath)

	case "validate":
		fs := flag.NewFlagSet("validate", flag.ExitOnError)
		instanceFlag := fs.String("instance", "", "Instance directory path")
		_ = fs.Parse(os.Args[2:])
		inst := *instanceFlag
		if inst == "" && fs.NArg() > 0 {
			inst = fs.Arg(0)
		}
		if inst == "" {
			usage()
		}
		if err := gate.ValidateAndExport(inst); err != nil {
			fatal(err)
		}

	case "eval":
		fs := flag.NewFlagSet("eval", flag.ExitOnError)
		instanceFlag := fs.String("instance", "", "Instance directory path")
		_ = fs.Parse(os.Args[2:])
		inst := *instanceFlag
		if inst == "" && fs.NArg() > 0 {
			inst = fs.Arg(0)
		}
		if inst == "" {
			usage()
		}
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
