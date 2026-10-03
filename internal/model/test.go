package model

import "time"

type Status string

const (
	StatusPassed Status = "passed"
	StatusFailed Status = "failed"
	StatusSkipped Status = "skipped"
	StatusError Status = "error"
)

type TestExecution struct {
	TestID string `json:"test_id"`
	Name string `json:"name"`
	ClassName string `json:"class_name,omitempty"`
	File string `json:"file,omitempty"`
	CommitSHA string `json:"commit_sha,omitempty"`
	Branch string `json:"branch,omitempty"`
	Runner string `json:"runner,omitempty"`
	Status Status `json:"status"`
	DurationMS int64 `json:"duration_ms"`
	Attempt int `json:"attempt"`
	Timestamp time.Time `json:"timestamp"`
}

type FlakeAssessment struct {
	TestID string `json:"test_id"`
	Name string `json:"name"`
	Runs int `json:"runs"`
	Failures int `json:"failures"`
	RetryPasses int `json:"retry_passes"`
	SameSHAConflicts int `json:"same_sha_conflicts"`
	FailureRate float64 `json:"failure_rate"`
	Confidence float64 `json:"confidence"`
	Signals []string `json:"signals"`
	EstimatedWasteMS int64 `json:"estimated_waste_ms"`
}
