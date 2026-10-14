package metrics

import (
	"testing"
	"time"

	"github.com/mvelosop/vloop/internal/defect"
	"github.com/mvelosop/vloop/internal/intervention"
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

func TestDerivedFlakyGate(t *testing.T) {
	m := &runs.Model{Folders: []runs.Folder{{
		Iterations: []runs.Iteration{
			{Iteration: 1, Task: "T1", Outcome: "done", Flaky: true},
			{Iteration: 2, Task: "T2", Outcome: "done"},
		},
	}}}
	got := Derive(m, nil)
	if len(got) != 1 {
		t.Fatalf("%d defects, want 1: %+v", len(got), got)
	}
	if d := got[0]; d.Task != "T1" || d.Origin != "env" || d.Kind != "bug" || d.FoundBy != "gate" {
		t.Errorf("flaky defect = %+v", d)
	}
	var x Matrix
	x.AddDerived(got)
	if x[3][0] != 1 || x[2][0] != 0 {
		t.Errorf("matrix %v: want env/gate 1 and work/gate 0", x)
	}
}

func TestDerivedCheckFailed(t *testing.T) {
	m := &runs.Model{Folders: []runs.Folder{{
		Iterations: []runs.Iteration{
			{Iteration: 1, Task: "T1", Outcome: "check_failed"},
			{Iteration: 2, Task: "T1", Outcome: "done"},
		},
	}}}
	got := Derive(m, nil)
	if len(got) != 1 {
		t.Fatalf("%d defects, want 1: %+v", len(got), got)
	}
	if d := got[0]; d.Task != "T1" || d.Iteration != 1 || d.Origin != "work" || d.Kind != "regression" || d.FoundBy != "gate" {
		t.Errorf("check-failure defect = %+v", d)
	}
	if ids := DerivedIDs("R", got); ids[0] != "R/i1-gate" {
		t.Errorf("id = %v", ids)
	}
}

func TestReclassifyByReleaseTime(t *testing.T) {
	rel := at(10, 0)
	mk := func(by string, d time.Duration) defect.Defect {
		return defect.Defect{FoundBy: by, Created: rel.Add(d).Format("2006-01-02T15:04:05Z")}
	}
	rec := []defect.Defect{
		mk("operator", -time.Hour), mk("operator", 0), mk("operator", time.Second),
		mk("review", time.Minute), mk("gate", -time.Minute), mk("user", -time.Minute),
	}
	var x Matrix
	for _, d := range rec {
		x.Add("work", d.FoundBy)
	}
	c := x.Counts()
	c.Reclassify(rec, rel)
	if c.InLoop != 1 || c.Operator != 2 || c.Escaped != 3 {
		t.Errorf("counts = %+v, want in-loop 1, operator 2, escaped 3", c)
	}
	u := x.Counts()
	u.Reclassify(rec, time.Time{})
	if u != x.Counts() {
		t.Errorf("unmerged brief reclassified: %+v", u)
	}
}

func TestAddFindingsCountsOnlyFindingsWithOrigin(t *testing.T) {
	var c DefectCounts
	c.Operator = 1 // an operator defect that did not escape
	c.AddFindings([]intervention.Intervention{
		{Kind: "verification-finding", Origin: "plan"},
		{Kind: "verification-finding"},
		{Kind: "repair", Origin: "work"},
	})
	if c.Operator != 2 || c.All() != 2 {
		t.Fatalf("operator %d, all %d, want 2 and 2", c.Operator, c.All())
	}
}
