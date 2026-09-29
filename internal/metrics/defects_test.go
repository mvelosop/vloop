package metrics

import (
	"testing"
	"time"

	"github.com/mvelosop/vloop/internal/defect"
	"github.com/mvelosop/vloop/internal/runs"
)

func at(h, m int) time.Time { return time.Date(2026, 1, 1, h, m, 0, 0, time.UTC) }

func TestDerivedGateAndReview(t *testing.T) {
	m := &runs.Model{Folders: []runs.Folder{{
		Iterations: []runs.Iteration{
			{Iteration: 1, Task: "T1", Outcome: "gate_fail"},
			{Iteration: 2, Task: "T1", Outcome: "review_fail"},
			{Iteration: 3, Task: "T1", Outcome: "rejected"},
			{Iteration: 4, Task: "T1", Outcome: "rejected"},
			{Iteration: 5, Task: "T1", Outcome: "blocked"},
			{Iteration: 6, Task: "T1", Outcome: "done"},
		},
		Verdicts: []runs.Verdict{
			{Iteration: 2, Findings: []runs.Finding{{Summary: "string finding"}}},
			{Iteration: 3, Findings: []runs.Finding{{Summary: "a", Kind: "spec-gap"}, {Summary: "b", Kind: "gate-gap"}, {Summary: "c", Kind: "regression"}}},
		},
	}}}
	got := Derive(m, nil)
	want := []struct{ origin, kind, by string }{
		{"work", "bug", "gate"},
		{"work", "bug", "review"},
		{"brief", "spec-gap", "review"},
		{"plan", "gate-gap", "review"},
		{"work", "regression", "review"},
		{"work", "bug", "review"}, // iteration 4: empty list, one defect
	}
	if len(got) != len(want) {
		t.Fatalf("%d defects, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].Origin != w.origin || got[i].Kind != w.kind || got[i].FoundBy != w.by {
			t.Errorf("defect %d = %+v, want %+v", i, got[i], w)
		}
	}
}

func TestDerivedGateHistoryReclassifies(t *testing.T) {
	m := &runs.Model{
		Folders: []runs.Folder{{Iterations: []runs.Iteration{
			{Iteration: 1, Task: "T2", Outcome: "gate_failed"},
			{Iteration: 2, Task: "T2", Outcome: "gate_failed"},
			{Iteration: 3, Task: "T1", Outcome: "gate_failed"},
		}}},
		Owned: &runs.Owned{Tasks: []runs.Commit{
			{Task: "T2", Outcome: "gate_failed", Time: at(9, 4)},
			{Task: "T2", Outcome: "gate_failed", Time: at(9, 8)},
			{Task: "T1", Outcome: "gate_failed", Time: at(9, 4)},
		}},
	}
	hist := func(by string) *runs.PlanDoc {
		return &runs.PlanDoc{Tasks: []runs.PlanTask{{ID: "T1"}, {ID: "T2", GateHistory: []runs.GateEntry{{ReplacedAt: "2026-01-01T09:05:00Z", By: by}}}}}
	}
	got := Derive(m, hist("operator"))
	if got[0].Origin != "plan" || got[0].Kind != "gate" {
		t.Errorf("failure before the entry = %+v, want plan/gate", got[0])
	}
	if got[1].Origin != "work" {
		t.Errorf("failure after the entry = %+v, want work", got[1])
	}
	if got[2].Origin != "work" {
		t.Errorf("failure on another task = %+v, want work", got[2])
	}
	for _, d := range Derive(m, hist("planner")) {
		if d.Origin != "work" {
			t.Errorf("planner entry reclassified %+v", d)
		}
	}
}

func TestMatrixCountsAndLayout(t *testing.T) {
	var x Matrix
	x.AddDerived([]Derived{{Origin: "work", FoundBy: "gate"}, {Origin: "plan", FoundBy: "review"}})
	x.AddRecorded([]defect.Defect{{Origin: "work", FoundBy: "user"}, {Origin: "env", FoundBy: "operator"}})
	x.Add("bogus", "gate")
	want := "       gate  review  operator  user\n" +
		"brief     0       0         0     0\n" +
		"plan      0       1         0     0\n" +
		"work      1       0         0     1\n" +
		"env       0       0         1     0\n"
	if x.String() != want {
		t.Fatalf("got\n%swant\n%s", x, want)
	}
	if c := x.Counts(); c != (DefectCounts{InLoop: 2, Operator: 1, Escaped: 1}) {
		t.Fatalf("counts %+v", c)
	}
}

func TestRemovalEfficiency(t *testing.T) {
	if (DefectCounts{}).RemovalEfficiency() != nil {
		t.Error("no defects should be n/a")
	}
	for _, c := range []struct {
		c    DefectCounts
		want int
	}{
		{DefectCounts{InLoop: 1}, 100},
		{DefectCounts{InLoop: 1, Escaped: 1}, 50},
		{DefectCounts{InLoop: 1, Operator: 1, Escaped: 1}, 67},
		{DefectCounts{Escaped: 2}, 0},
	} {
		if p := c.c.RemovalEfficiency(); p == nil || *p != c.want {
			t.Errorf("%+v = %v, want %d", c.c, p, c.want)
		}
	}
}
