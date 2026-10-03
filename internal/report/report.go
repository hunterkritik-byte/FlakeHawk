package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/hunterkritik-byte/FlakeHawk/internal/model"
)

func Text(w io.Writer, assessments []model.FlakeAssessment) error {
	fmt.Fprintln(w, "FlakeHawk")
	fmt.Fprintln(w, "────────────────────────────────────────────")
	if len(assessments) == 0 {
		_, err := fmt.Fprintln(w, "No flaky tests detected.")
		return err
	}
	fmt.Fprintln(w, "Potentially flaky tests")
	fmt.Fprintln(w, "────────────────────────────────────────────")
	for _, a := range assessments {
		fmt.Fprintf(w, "\n%.0f%%  %s\n", a.Confidence*100, a.Name)
		fmt.Fprintf(w, "     %d runs · %d failures · %d retry passes · %d same-SHA conflicts\n",
			a.Runs, a.Failures, a.RetryPasses, a.SameSHAConflicts)
		fmt.Fprintf(w, "     signals: %s\n", strings.Join(a.Signals, ", "))
	}
	_, err := fmt.Fprintln(w, "────────────────────────────────────────────")
	return err
}

func JSON(w io.Writer, assessments []model.FlakeAssessment) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(assessments)
}
