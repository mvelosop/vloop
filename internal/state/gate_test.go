package state

import (
	"reflect"
	"testing"
	"time"
)

func TestGateArgs(t *testing.T) {
	for shell, want := range map[string][]string{
		"sh":         {"-c", "x"},
		"bash":       {"-c", "x"},
		"pwsh":       {"-NoProfile", "-NonInteractive", "-Command", "x"},
		"powershell": {"-NoProfile", "-NonInteractive", "-Command", "x"},
		"cmd":        {"/C", "x"},
	} {
		if got := GateArgs(shell, "x"); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %v want %v", shell, got, want)
		}
	}
}

func TestReplaceGate(t *testing.T) {
	p := &Plan{Tasks: []Task{{ID: "T1", Verify: "old"}}}
	at := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := ReplaceGate(p, "T1", "new", "why", at); err != nil {
		t.Fatal(err)
	}
	want := []GateReplace{{Verify: "old", ReplacedAt: "2026-01-02T03:04:05Z", Reason: "why", By: "operator"}}
	if p.Tasks[0].Verify != "new" || !reflect.DeepEqual(p.Tasks[0].GateHistory, want) {
		t.Fatalf("got %+v", p.Tasks[0])
	}
	if err := ReplaceGate(p, "T1", "new", "again", at); err == nil || err.Error() != "T1 already has that verify command" {
		t.Fatalf("same command: %v", err)
	}
	if err := ReplaceGate(p, "T1", "", "r", at); err == nil {
		t.Fatal("empty command accepted")
	}
	if err := ReplaceGate(p, "T9", "x", "r", at); err == nil || err.Error() != "no task T9" {
		t.Fatalf("unknown task: %v", err)
	}
	if len(p.Tasks[0].GateHistory) != 1 {
		t.Fatal("a refused replacement was recorded")
	}
}
