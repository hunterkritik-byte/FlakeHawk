package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hunterkritik-byte/FlakeHawk/internal/detector"
	"github.com/hunterkritik-byte/FlakeHawk/internal/ingest"
	"github.com/hunterkritik-byte/FlakeHawk/internal/parser/junit"
	"github.com/hunterkritik-byte/FlakeHawk/internal/report"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 { usage(); os.Exit(2) }
	switch os.Args[1] {
	case "report": reportCmd(os.Args[2:])
	case "ingest": ingestCmd(os.Args[2:])
	case "version": fmt.Println("flakehawk", version)
	case "help", "-h", "--help": usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q
", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func reportCmd(args []string) {
	fs := flag.NewFlagSet("report", flag.ExitOnError)
	format := fs.String("format", "text", "output format: text or json")
	minRuns := fs.Int("min-runs", 5, "minimum executions required")
	_ = fs.Parse(args)
	if fs.NArg() != 1 { fmt.Fprintln(os.Stderr, "usage: flakehawk report [flags] <file-or-directory>"); os.Exit(2) }
	executions, err := junit.ParsePath(fs.Arg(0))
	if err != nil { fmt.Fprintln(os.Stderr, "error:", err); os.Exit(1) }
	assessments := detector.Analyze(executions, detector.Config{MinRuns: *minRuns, MinConfidence: 0.50})
	var outputErr error
	switch *format {
	case "json": outputErr = report.JSON(os.Stdout, assessments)
	case "text": outputErr = report.Text(os.Stdout, assessments)
	default: fmt.Fprintln(os.Stderr, "error: unsupported format:", *format); os.Exit(2)
	}
	if outputErr != nil { fmt.Fprintln(os.Stderr, "error:", outputErr); os.Exit(1) }
}

func ingestCmd(args []string) {
	fs := flag.NewFlagSet("ingest", flag.ExitOnError)
	output := fs.String("output", ".flakehawk/runs.jsonl", "append-only JSONL ingestion store")
	commit := fs.String("commit", "", "commit SHA; defaults to GITHUB_SHA")
	branch := fs.String("branch", "", "branch/ref; defaults to GITHUB_REF_NAME or GITHUB_HEAD_REF")
	runner := fs.String("runner", "", "runner identity; defaults to RUNNER_OS/RUNNER_ARCH")
	attempt := fs.Int("attempt", 0, "CI attempt; defaults to GITHUB_RUN_ATTEMPT or 1")
	_ = fs.Parse(args)
	if fs.NArg() != 1 { fmt.Fprintln(os.Stderr, "usage: flakehawk ingest [flags] <junit-file-or-directory>"); os.Exit(2) }

	meta := ingest.MetadataFromEnv()
	if *commit != "" { meta.CommitSHA = *commit }
	if *branch != "" { meta.Branch = *branch }
	if *runner != "" { meta.Runner = *runner }
	if *attempt > 0 { meta.Attempt = *attempt }

	record, err := ingest.ParseAndEnrich(fs.Arg(0), meta)
	if err != nil { fmt.Fprintln(os.Stderr, "error:", err); os.Exit(1) }
	added, err := ingest.AppendUnique(*output, record)
	if err != nil { fmt.Fprintln(os.Stderr, "error:", err); os.Exit(1) }
	if added {
		fmt.Printf("ingested run %s (%d executions)
", record.RunID[:12], len(record.Executions))
	} else {
		fmt.Printf("duplicate run ignored: %s
", record.RunID[:12])
	}
}

func usage() {
	fmt.Println(strings.TrimSpace("FlakeHawk - explainable flaky test detection

Usage:
  flakehawk report [flags] <file-or-directory>
  flakehawk ingest [flags] <junit-file-or-directory>
  flakehawk version

Ingest flags:
  -commit   Commit SHA
  -branch   Branch/ref name
  -runner   Runner identity
  -attempt  CI attempt number
  -output   JSONL ingestion store (default .flakehawk/runs.jsonl)"))
}
