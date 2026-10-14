package driver

import (
	"reflect"
	"testing"
)

func TestDropEnv(t *testing.T) {
	in := []string{"A=1", "VLOOP_GATE_TASK=T1", "B=2", "VLOOP_GATE_TASK_X=keep", "VLOOP_ACTIVE_TASK=T2", "C="}
	got := dropEnv(in, "VLOOP_GATE_TASK", "VLOOP_ACTIVE_TASK")
	want := []string{"A=1", "B=2", "VLOOP_GATE_TASK_X=keep", "C="}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
