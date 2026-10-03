package detector

import (
	"testing"
	"time"

	"github.com/hunterkritik-byte/FlakeHawk/internal/model"
)

func TestAnalyzeSameSHAAndRetry(t *testing.T) {
	base := time.Date(2026, 10, 3, 10, 0, 0, 0, time.UTC)
	var runs []model.TestExecution
	for i := 0; i < 6; i++ {
		status := model.StatusPassed
		attempt := 1
		if i == 1 { status = model.StatusFailed }
		if i == 2 { status = model.StatusPassed; attempt = 2 }
		sha := "sha-" + string(rune('a'+i))
		if i == 1 || i == 2 { sha = "same-sha" }
		runs = append(runs, model.TestExecution{
			TestID: "checkout::timeout", Name: "TestTimeout", CommitSHA: sha,
			Status: status, Attempt: attempt, Timestamp: base.Add(time.Duration(i)*time.Minute), DurationMS: 1000,
		})
	}
	got := Analyze(runs, Config{MinRuns: 5, MinConfidence: 0.5})
	if len(got) != 1 { t.Fatalf("got %d assessments, want 1", len(got)) }
	if got[0].SameSHAConflicts != 1 { t.Fatalf("same-SHA conflicts = %d, want 1", got[0].SameSHAConflicts) }
	if got[0].RetryPasses != 1 { t.Fatalf("retry passes = %d, want 1", got[0].RetryPasses) }
}
