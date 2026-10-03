package junit

import (
	"strings"
	"testing"

	"github.com/hunterkritik-byte/FlakeHawk/internal/model"
)

func TestParse(t *testing.T) {
	data := `<testsuites><testsuite name="unit" timestamp="2026-10-03T10:00:00Z">
	<testcase classname="Checkout" name="TestTimeout" file="checkout_test.go" time="0.250"><failure message="timeout"/></testcase>
	<testcase classname="Checkout" name="TestOK" time="0.010"/>
</testsuite></testsuites>`
	got, err := Parse(strings.NewReader(data), "fixture.xml")
	if err != nil { t.Fatal(err) }
	if len(got) != 2 { t.Fatalf("got %d executions, want 2", len(got)) }
	if got[0].Status != model.StatusFailed || got[1].Status != model.StatusPassed { t.Fatalf("unexpected statuses: %#v", got) }
	if got[0].DurationMS != 250 { t.Fatalf("duration = %d, want 250ms", got[0].DurationMS) }
}
