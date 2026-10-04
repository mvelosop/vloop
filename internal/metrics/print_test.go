package metrics

import (
	"bytes"
	"strings"
	"testing"
)

// A brief whose plan had a gate review shows its cost, time and models beside
// the other phases; a brief without one keeps the v1 lines.
func TestPrintSummaryGateReview(t *testing.T) {
	r := &Report{}
	r.Cost.Total, r.Cost.Plan, r.Cost.GateReview = 1.5, 1.0, 0.5
	r.Time.GateReviewMS = 120000
	r.Models.GateReview = []string{"claude-sonnet-5-5"}
	var b bytes.Buffer
	PrintSummary(&b, r)
	for _, want := range []string{"· gate review 0.50 ·", "· gate review 2.0 min ·", "· gate review claude-sonnet-5-5\n"} {
		if !strings.Contains(b.String(), want) {
			t.Errorf("summary lacks %q:\n%s", want, b.String())
		}
	}
	b.Reset()
	PrintSummary(&b, &Report{})
	if strings.Contains(b.String(), "gate review") {
		t.Errorf("a brief with no gate review shows one:\n%s", b.String())
	}
}
