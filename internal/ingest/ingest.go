package ingest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hunterkritik-byte/FlakeHawk/internal/model"
	"github.com/hunterkritik-byte/FlakeHawk/internal/parser/junit"
)

type Metadata struct {
	CommitSHA string
	Branch string
	Runner string
	Attempt int
}

type Record struct {
	RunID string `json:"run_id"`
	Source string `json:"source"`
	Metadata Metadata `json:"metadata"`
	Executions []model.TestExecution `json:"executions"`
}

func ParseAndEnrich(path string, meta Metadata) (Record, error) {
	execs, err := junit.ParsePath(path)
	if err != nil { return Record{}, fmt.Errorf("parse reports: %w", err) }
	for i := range execs {
		execs[i].CommitSHA = meta.CommitSHA
		execs[i].Branch = meta.Branch
		execs[i].Runner = meta.Runner
		execs[i].Attempt = meta.Attempt
	}
	id, err := runID(path, meta)
	if err != nil { return Record{}, err }
	return Record{RunID: id, Source: path, Metadata: meta, Executions: execs}, nil
}

func AppendUnique(path string, record Record) (bool, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil { return false, fmt.Errorf("create ingest directory: %w", err) }
	existing, err := os.Open(path)
	if err == nil {
		defer existing.Close()
		dec := json.NewDecoder(existing)
		for {
			var r Record
			if err := dec.Decode(&r); err != nil { break }
			if r.RunID == record.RunID { return false, nil }
		}
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("read ingest store: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil { return false, fmt.Errorf("open ingest store: %w", err) }
	defer f.Close()
	if err := json.NewEncoder(f).Encode(record); err != nil { return false, fmt.Errorf("write ingest record: %w", err) }
	return true, nil
}

func runID(path string, meta Metadata) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil { return "", fmt.Errorf("read report for run identity: %w", err) }
	h := sha256.New()
	h.Write(data)
	fmt.Fprintf(h, "\x00%s\x00%s\x00%s\x00%d", meta.CommitSHA, meta.Branch, meta.Runner, meta.Attempt)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func MetadataFromEnv() Metadata {
	attempt := 1
	if n := parsePositiveInt(os.Getenv("GITHUB_RUN_ATTEMPT")); n > 0 { attempt = n }
	branch := os.Getenv("GITHUB_REF_NAME")
	if branch == "" { branch = os.Getenv("GITHUB_HEAD_REF") }
	return Metadata{
		CommitSHA: os.Getenv("GITHUB_SHA"),
		Branch: branch,
		Runner: os.Getenv("RUNNER_OS") + "/" + os.Getenv("RUNNER_ARCH"),
		Attempt: attempt,
	}
}

func parsePositiveInt(s string) int {
	n := 0
	for _, r := range strings.TrimSpace(s) {
		if r < '0' || r > '9' { return 0 }
		n = n*10 + int(r-'0')
	}
	return n
}
