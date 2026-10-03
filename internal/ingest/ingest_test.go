package ingest

import (
	"path/filepath"
	"testing"
)

func TestAppendUniqueRejectsDuplicate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runs.jsonl")
	record := Record{RunID: "same-run", Source: "fixture"}
	added, err := AppendUnique(path, record)
	if err != nil || !added { t.Fatalf("first append: added=%v err=%v", added, err) }
	added, err = AppendUnique(path, record)
	if err != nil { t.Fatalf("second append: %v", err) }
	if added { t.Fatal("duplicate record was appended") }
}

func TestMetadataFromEnv(t *testing.T) {
	t.Setenv("GITHUB_SHA", "abc123")
	t.Setenv("GITHUB_REF_NAME", "feature/test")
	t.Setenv("RUNNER_OS", "Linux")
	t.Setenv("RUNNER_ARCH", "X64")
	t.Setenv("GITHUB_RUN_ATTEMPT", "3")
	m := MetadataFromEnv()
	if m.CommitSHA != "abc123" || m.Branch != "feature/test" || m.Runner != "Linux/X64" || m.Attempt != 3 {
		t.Fatalf("unexpected metadata: %+v", m)
	}
}
