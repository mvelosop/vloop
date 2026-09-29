package state

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func amendRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Dir(Path(root)), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join("testdata", "plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Path(root), data, 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestAmendLeavesPlanOnChangeError(t *testing.T) {
	root := amendRoot(t)
	before, _ := os.ReadFile(Path(root))
	want := errors.New("no")
	if err := Amend(root, nil, func(*Plan) error { return want }); !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
	after, _ := os.ReadFile(Path(root))
	if string(before) != string(after) {
		t.Fatal("plan changed")
	}
}

func TestAmendRefusesInvalidResult(t *testing.T) {
	root := amendRoot(t)
	before, _ := os.ReadFile(Path(root))
	err := Amend(root, nil, func(p *Plan) error { p.Tasks[0].Status = "paused"; return nil })
	if err == nil {
		t.Fatal("want a refusal")
	}
	after, _ := os.ReadFile(Path(root))
	if string(before) != string(after) {
		t.Fatal("plan changed")
	}
}

func TestDropRefusesDependent(t *testing.T) {
	p := &Plan{Tasks: []Task{{ID: "T1"}, {ID: "T2", DependsOn: []string{"T1"}}}}
	if err := Drop(p, "T1"); err == nil || err.Error() != "T2 depends on T1" {
		t.Fatalf("got %v", err)
	}
	if err := Drop(p, "T2"); err != nil || len(p.Tasks) != 1 {
		t.Fatalf("got %v, %d tasks", err, len(p.Tasks))
	}
}
