package metrics

import (
	"fmt"
	"time"

	"github.com/mvelosop/vloop/internal/defect"
	"github.com/mvelosop/vloop/internal/runs"
)

// Derived is a defect the loop itself caught, computed from a run and never
// stored.
type Derived struct {
	Task      string
	Iteration int
	Origin    string
	Kind      string
	FoundBy   string
	Summary   string
}

// Derive computes the defects of a brief's runs. plan is the plan at the last
// run commit and supplies gate_history; it may be nil. A gate failure that
// happened before an operator's gate_history entry on its task is the plan's
// fault (origin plan, kind gate), not the work's. Blocked iterations are not
// defects.
func Derive(m *runs.Model, plan *runs.PlanDoc) []Derived {
	replaced := map[string]time.Time{} // task -> latest operator replacement
	if plan != nil {
		for _, pt := range plan.Tasks {
			for _, e := range pt.GateHistory {
				t, err := time.Parse(time.RFC3339, e.ReplacedAt)
				if e.By == "operator" && err == nil && t.After(replaced[pt.ID]) {
					replaced[pt.ID] = t
				}
			}
		}
	}
	// The kth gate failure of a task is stamped by the kth gate_failed task
	// commit, since iteration records may carry no times.
	var commitTimes map[string][]time.Time
	if m.Owned != nil {
		commitTimes = map[string][]time.Time{}
		for _, c := range m.Owned.Tasks {
			if c.Outcome == "gate_failed" || c.Outcome == "gate_fail" {
				commitTimes[c.Task] = append(commitTimes[c.Task], c.Time)
			}
		}
	}
	seen := map[string]int{}
	var out []Derived
	for _, f := range m.Folders {
		verdicts := map[int]runs.Verdict{}
		for _, v := range f.Verdicts {
			verdicts[v.Iteration] = v
		}
		for _, it := range f.Iterations {
			// A flaky gate is the environment's, not the work's.
			if it.Flaky {
				out = append(out, Derived{Task: it.Task, Iteration: it.Iteration, Origin: "env", Kind: "bug", FoundBy: "gate", Summary: "flaky gate"})
			}
			switch it.Canonical() {
			case "gate_failed":
				k := seen[it.Task]
				seen[it.Task]++
				d := Derived{Task: it.Task, Iteration: it.Iteration, Origin: "work", Kind: "bug", FoundBy: "gate", Summary: "gate failed"}
				var at time.Time
				if ts := commitTimes[it.Task]; k < len(ts) {
					at = ts[k]
				} else {
					at = it.Ended
				}
				if r, ok := replaced[it.Task]; ok && !at.IsZero() && at.Before(r) {
					d.Origin, d.Kind = "plan", "gate"
				}
				out = append(out, d)
			case "rejected":
				fs := verdicts[it.Iteration].Findings
				if len(fs) == 0 {
					out = append(out, Derived{Task: it.Task, Iteration: it.Iteration, Origin: "work", Kind: "bug", FoundBy: "review", Summary: "review rejected"})
				}
				for _, fi := range fs {
					d := Derived{Task: it.Task, Iteration: it.Iteration, Origin: "work", Kind: fi.Kind, FoundBy: "review", Summary: fi.Summary}
					switch fi.Kind {
					case "":
						d.Kind = "bug"
					case "spec-gap":
						d.Origin = "brief"
					case "gate-gap":
						d.Origin = "plan"
					}
					out = append(out, d)
				}
			}
		}
	}
	return out
}

// Matrix counts defects by origin (rows) and catcher (columns), in the order of
// defect.Origins and defect.FoundBys.
type Matrix [4][4]int

func index(list []string, s string) int {
	for i, x := range list {
		if x == s {
			return i
		}
	}
	return -1
}

// Add counts one defect; values outside the enums are ignored.
func (x *Matrix) Add(origin, foundBy string) {
	if o, f := index(defect.Origins, origin), index(defect.FoundBys, foundBy); o >= 0 && f >= 0 {
		x[o][f]++
	}
}

// AddDerived counts derived defects.
func (x *Matrix) AddDerived(ds []Derived) {
	for _, d := range ds {
		x.Add(d.Origin, d.FoundBy)
	}
}

// AddRecorded counts recorded defects.
func (x *Matrix) AddRecorded(ds []defect.Defect) {
	for _, d := range ds {
		x.Add(d.Origin, d.FoundBy)
	}
}

// String is the matrix in the layout `defect list --matrix` prints.
func (x Matrix) String() string {
	s := "       gate  review  operator  user\n"
	for i, o := range defect.Origins {
		s += fmt.Sprintf("%-5s  %4d  %6d  %8d  %4d\n", o, x[i][0], x[i][1], x[i][2], x[i][3])
	}
	return s
}

// DefectCounts are the counting rule's three groups.
type DefectCounts struct {
	InLoop   int // found by gate or review
	Operator int
	Escaped  int // found by user
}

// Counts groups the matrix by catcher.
func (x Matrix) Counts() DefectCounts {
	var c DefectCounts
	for i := range x {
		c.InLoop += x[i][0] + x[i][1]
		c.Operator += x[i][2]
		c.Escaped += x[i][3]
	}
	return c
}

// All is every defect counted.
func (c DefectCounts) All() int { return c.InLoop + c.Operator + c.Escaped }

// RemovalEfficiency is (in-loop + operator) / all as a whole percentage; nil
// (printed `n/a`) when there are no defects.
func (c DefectCounts) RemovalEfficiency() *int {
	if c.All() == 0 {
		return nil
	}
	p := (100*(c.InLoop+c.Operator) + c.All()/2) / c.All()
	return &p
}

// DerivedIDs is the id of each derived defect, in the order given:
// `<run id>/i<iteration>-gate` for a gate failure, `<run id>/i<iteration>-review-<n>`
// for a review finding, n counting from 1 within the iteration.
func DerivedIDs(runID string, ds []Derived) []string {
	reviews := map[int]int{}
	ids := make([]string, len(ds))
	for i, d := range ds {
		if d.FoundBy == "gate" {
			ids[i] = fmt.Sprintf("%s/i%d-gate", runID, d.Iteration)
			continue
		}
		reviews[d.Iteration]++
		ids[i] = fmt.Sprintf("%s/i%d-review-%d", runID, d.Iteration, reviews[d.Iteration])
	}
	return ids
}

// DeriveBrief computes the derived defects of a brief's runs, exactly as Build
// counts them. It returns nil when the brief has no runs.
func DeriveBrief(root, brief string) ([]Derived, error) {
	m, err := runs.Read(root, brief)
	if err != nil {
		return nil, err
	}
	if len(m.Folders) == 0 {
		return nil, nil
	}
	var plan *runs.PlanDoc
	if o := m.Owned; o != nil {
		if plan, err = runs.PlanAt(root, o.Layout, planSHA(o)); err != nil {
			return nil, err
		}
	}
	return Derive(m, plan), nil
}

// PlanOf reads a brief's plan as committed at its last run commit. It returns
// nil when the brief has no owned run.
func PlanOf(root, brief string) (*runs.PlanDoc, error) {
	m, err := runs.Read(root, brief)
	if err != nil {
		return nil, err
	}
	if m.Owned == nil {
		return nil, nil
	}
	return runs.PlanAt(root, m.Owned.Layout, planSHA(m.Owned))
}
