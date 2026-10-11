package driver

import (
	"reflect"
	"testing"
	"time"
)

func TestSessionLineEachPhase(t *testing.T) {
	cases := []struct {
		spec Spec
		want string
	}{
		{Spec{Phase: PhasePlan, Arg: "docs/briefs/x.loop-brief.md"}, "   plan · 0m09s · $0.50 · run $0.50 / $40.00"},
		{Spec{Phase: PhaseGateReview}, "   gate review · 1m05s · $0.50 · run $0.50 / $40.00"},
		{Spec{Phase: PhaseWork, Arg: "T3"}, "   work T3 · 6m12s · $0.50 · run $0.50 / $40.00"},
		{Spec{Phase: PhaseReview, Arg: "T3"}, "   review T3 · 0m00s · $0.50 · run $0.50 / $40.00"},
	}
	walls := []time.Duration{9 * time.Second, 65 * time.Second, 6*time.Minute + 12*time.Second, 0}
	for i, c := range cases {
		if got := sessionLine(c.spec, walls[i], 0.5, 0.5, 40); got != c.want {
			t.Errorf("%s: got %q, want %q", c.spec.Phase, got, c.want)
		}
	}
	if got, want := sessionLine(Spec{Phase: PhaseWork, Arg: "T3"}, 6*time.Minute+12*time.Second, 1.84, 9.4, 40),
		"   work T3 · 6m12s · $1.84 · run $9.40 / $40.00"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCostWarningsOnceEach(t *testing.T) {
	w50 := "   cost: $2.20 spent, 50% of the $4.00 ceiling"
	w80 := "   cost: $3.30 spent, 80% of the $4.00 ceiling"
	steps := []struct {
		before, after float64
		want          []string
	}{
		{0, 1.1, nil},
		{1.1, 2.2, []string{w50}},
		{2.2, 3.3, []string{w80}},
		{3.3, 4.4, nil},
	}
	for _, s := range steps {
		if got := costWarnings(s.before, s.after, 4); !reflect.DeepEqual(got, s.want) {
			t.Errorf("%.2f→%.2f: got %q, want %q", s.before, s.after, got, s.want)
		}
	}
	both := costWarnings(0, 3.5, 4)
	if len(both) != 2 || both[0][len("   cost: $3.50 spent, "):][:3] != "50%" {
		t.Errorf("a session crossing both thresholds warns 50 then 80, got %q", both)
	}
	if got := costWarnings(0, 5, 0); got != nil {
		t.Errorf("no ceiling, no warnings, got %q", got)
	}
}
