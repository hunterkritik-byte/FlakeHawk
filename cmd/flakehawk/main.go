package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hunterkritik-byte/FlakeHawk/internal/detector"
	"github.com/hunterkritik-byte/FlakeHawk/internal/parser/junit"
	"github.com/hunterkritik-byte/FlakeHawk/internal/report"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 { usage(); os.Exit(2) }
	switch os.Args[1] {
	case "report": reportCmd(os.Args[2:])
	case "version": fmt.Println("flakehawk", version)
	case "help", "-h", "--help": usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1:])
		usage()
		os.Exit(2)
	}
}

func reportCmd(args []string) {
	fs := flag.NewFlagSet("report", flag.ExitOnError)
	format := fs.String("format", "text", "output format: text or json")
	minRuns := fs.Int("min-runs", 5, "minimum executions required")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: flakehawk report [flags] <file-or-directory>")
		os.Exit(2)
	}
	executions, err := junit.ParsePath(fs.Arg(0))
	if err != nil { fmt.Fprintln(os.Stderr, "error:", err); os.Exit(1) }
	assessments := detector.Analyze(executions, detector.Config{MinRuns: *minRuns, MinConfidence: 0.50})
	var outputErr error
	if *format == "json" { outputErr = report.JSON(os.Stdout, assessments) } else { outputErr = report.Text(os.Stdout, assessments) }
	if outputErr != nil { fmt.Fprintln(os.Stderr, "error:", outputErr); os.Exit(1) }
}

func usage() {
	fmt.Println("FlakeHawk - explainable flaky test detection\n\nUsage:\n  flakehawk report [flags] <file-or-directory>\n  flakehawk version")
}
