package detector

import (
	"math"
	"sort"

	"github.com/hunterkritik-byte/FlakeHawk/internal/model"
)

type Config struct {
	MinRuns int
	MinConfidence float64
}

func Analyze(executions []model.TestExecution, cfg Config) []model.FlakeAssessment {
	groups := make(map[string][]model.TestExecution)
	for _, e := range executions {
		if e.Status != model.StatusSkipped { groups[e.TestID] = append(groups[e.TestID], e) }
	}

	var out []model.FlakeAssessment
	for id, runs := range groups {
		sort.SliceStable(runs, func(i, j int) bool {
			if runs[i].Timestamp.Equal(runs[j].Timestamp) { return runs[i].Attempt < runs[j].Attempt }
			return runs[i].Timestamp.Before(runs[j].Timestamp)
		})
		failures, waste := 0, int64(0)
		for _, r := range runs {
			if r.Status == model.StatusFailed || r.Status == model.StatusError {
				failures++
				waste += r.DurationMS
			}
		}
		retryPasses := countRetryPasses(runs)
		conflicts := countSameSHAConflicts(runs)
		confidence := confidenceScore(len(runs), failures, retryPasses, conflicts)
		signals := []string{}
		if conflicts > 0 { signals = append(signals, "same_commit_disagreement") }
		if retryPasses > 0 { signals = append(signals, "retry_pass") }

		if len(runs) >= cfg.MinRuns && confidence >= cfg.MinConfidence && len(signals) > 0 {
			out = append(out, model.FlakeAssessment{
				TestID: id, Name: runs[0].Name, Runs: len(runs), Failures: failures,
				RetryPasses: retryPasses, SameSHAConflicts: conflicts,
				FailureRate: float64(failures)/float64(len(runs)),
				Confidence: confidence, Signals: signals, EstimatedWasteMS: waste,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Confidence == out[j].Confidence { return out[i].EstimatedWasteMS > out[j].EstimatedWasteMS }
		return out[i].Confidence > out[j].Confidence
	})
	return out
}

func countRetryPasses(runs []model.TestExecution) int {
	count := 0
	for i := 0; i+1 < len(runs); i++ {
		if runs[i].Status == model.StatusFailed && runs[i+1].Attempt > runs[i].Attempt && runs[i+1].Status == model.StatusPassed { count++ }
	}
	return count
}

func countSameSHAConflicts(runs []model.TestExecution) int {
	type state struct{ pass, fail bool }
	bySHA := map[string]state{}
	for _, r := range runs {
		if r.CommitSHA == "" { continue }
		s := bySHA[r.CommitSHA]
		if r.Status == model.StatusPassed { s.pass = true }
		if r.Status == model.StatusFailed || r.Status == model.StatusError { s.fail = true }
		bySHA[r.CommitSHA] = s
	}
	count := 0
	for _, s := range bySHA { if s.pass && s.fail { count++ } }
	return count
}

// confidenceScore is deliberately explainable rather than an opaque ML score.
func confidenceScore(n, failures, retryPasses, conflicts int) float64 {
	if n == 0 { return 0 }
	base := wilsonLowerBound(n, failures)
	if retryPasses > 0 { base += 0.15 * float64(retryPasses)/float64(n) }
	if conflicts > 0 { base += 0.50 }
	return math.Min(0.99, base)
}

func wilsonLowerBound(n, successes int) float64 {
	if n == 0 { return 0 }
	p := float64(successes)/float64(n)
	z := 1.96
	den := 1 + z*z/float64(n)
	center := p + z*z/(2*float64(n))
	margin := z * math.Sqrt((p*(1-p)+z*z/(4*float64(n)))/float64(n))
	return (center-margin)/den
}
