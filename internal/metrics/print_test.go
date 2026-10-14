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

// A brief that delivered no line classified as code or test shows n/a in the
// table and says how many lines matched no stack; one with code keeps its
// numbers and has no such line.
func TestPrintNotMeasuredWhereNothingClassified(t *testing.T) {
	none := &Report{RunID: "B1-none"}
	none.Size.Delivered = JSONLines{Docs: 12, Other: 3487}
	code := &Report{RunID: "B2-code"}
	code.Size.Delivered = JSONLines{Code: 40, Test: 10, Other: 5}

	var b bytes.Buffer
	PrintBriefTable(&b, []*Report{none, code})
	rows := strings.Split(strings.TrimSpace(b.String()), "\n")
	if f := strings.Fields(rows[1]); f[3] != "n/a" || f[4] != "n/a" {
		t.Errorf("code and test of a brief with none classified = %q %q, want n/a n/a", f[3], f[4])
	}
	if f := strings.Fields(rows[2]); f[3] != "40" || f[4] != "10" {
		t.Errorf("code and test of a brief with code = %q %q, want 40 10", f[3], f[4])
	}

	b.Reset()
	PrintSummary(&b, none)
	want := "docs 12 · test:code n/a\n           3,487 delivered lines match no stack in metrics.stacks\n"
	if !strings.Contains(b.String(), want) {
		t.Errorf("summary lacks %q:\n%s", want, b.String())
	}
	b.Reset()
	PrintSummary(&b, code)
	if strings.Contains(b.String(), "match no stack") {
		t.Errorf("a brief with code shows the unclassified line:\n%s", b.String())
	}
	b.Reset()
	none.Size.Delivered.Other = 0
	PrintSummary(&b, none)
	if strings.Contains(b.String(), "match no stack") {
		t.Errorf("the line is printed with zero unclassified lines:\n%s", b.String())
	}
}

func TestPrintByTaskRatioColumns(t *testing.T) {
	ratio, per1k := 1.5, 2.5
	r := &Report{ByTask: []ReportTask{
		{ID: "T1", Churn: JSONLines{Code: 200, Test: 300}, CostUSD: 0.5, TestCodeRatio: &ratio, Per1000CodeUSD: &per1k},
		{ID: "T2"},
	}}
	var b strings.Builder
	PrintByTask(&b, r)
	got := strings.Join(strings.Fields(b.String()), " ")
	want := "id area kind att code+ test+ docs+ other+ test:code $/kloc agent cost model " +
		"T1 - - 0 200 300 0 0 1.50 2.50 0m00s $0.50 - T2 - - 0 0 0 0 0 - - 0m00s $0.00 -"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}
