package driver

import (
	"testing"

	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/state"
)

func TestPlanChecksCopiesConfigInOrder(t *testing.T) {
	got := planChecks([]config.CheckDef{{Name: "a", Paths: []string{"a/**"}, Run: "x"}, {Name: "b", Run: "y"}})
	want := []state.PlanCheck{{Name: "a", Paths: []string{"a/**"}, Run: "x"}, {Name: "b", Paths: []string{}, Run: "y"}}
	if !sameChecks(got, want) {
		t.Errorf("planChecks = %+v, want %+v", got, want)
	}
	if sameChecks(got, want[:1]) || sameChecks(got, []state.PlanCheck{want[1], want[0]}) {
		t.Error("sameChecks ignores a missing check or the order")
	}
}
