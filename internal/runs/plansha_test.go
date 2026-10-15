package runs

import "testing"

func TestOwnedPlanSHA(t *testing.T) {
	o := &Owned{Plan: Commit{SHA: "plan"}}
	if got := o.PlanSHA(); got != "plan" {
		t.Errorf("no tasks, no run: got %q, want plan", got)
	}
	o.Tasks = []Commit{{SHA: "t1"}, {SHA: "t2"}}
	if got := o.PlanSHA(); got != "t2" {
		t.Errorf("tasks: got %q, want t2", got)
	}
	o.Run = &Commit{SHA: "run"}
	if got := o.PlanSHA(); got != "run" {
		t.Errorf("run commit: got %q, want run", got)
	}
}
