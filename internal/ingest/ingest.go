package ingest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
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
			if err := dec.Decode(&r); err != nil {
				if err != io.EOF { return false, fmt.Errorf("decode ingest store: %w", err) }
				break
			}
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
	h := sha256.New()
	info, err := os.Stat(path)
	if err != nil { return "", fmt.Errorf("stat report path: %w", err) }

	if info.IsDir() {
		var files []string
		err = filepath.Walk(path, func(p string, info os.FileInfo, walkErr error) error {
			if walkErr != nil { return walkErr }
			if info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".xml") { return nil }
			files = append(files, p)
			return nil
		})
		if err != nil { return "", fmt.Errorf("enumerate reports: %w", err) }
		sort.Strings(files)
		for _, p := range files {
			if err := hashFile(h, p); err != nil { return "", err }
		}
	} else {
		if err := hashFile(h, path); err != nil { return "", err }
	}

	fmt.Fprintf(h, "\x00%s\x00%s\x00%s\x00%d", meta.CommitSHA, meta.Branch, meta.Runner, meta.Attempt)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func hashFile(h io.Writer, path string) error {
	f, err := os.Open(path)
	if err != nil { return fmt.Errorf("open report for run identity: %w", err) }
	defer f.Close()
	if _, err := io.Copy(h, f); err != nil { return fmt.Errorf("hash report: %w", err) }
	return nil
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
