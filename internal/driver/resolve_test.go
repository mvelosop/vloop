package driver

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveAgent(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".vloop"), 0o755); err != nil {
		t.Fatal(err)
	}
	r, err := ResolveRun(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{PhaseWork, PhaseReview, PhaseVerify} {
		if got := r.Agent(p); got != "claude" {
			t.Errorf("run agent %s: %q", p, got)
		}
	}
	for _, p := range []string{PhasePlan, PhaseGateReview} {
		if got, err := ResolveAgent(root, p); err != nil || got != "claude" {
			t.Errorf("session agent %s: %q %v", p, got, err)
		}
	}
	t.Setenv("VLOOP_AGENT_REVIEW", "copilot")
	if _, err := ResolveRun(root); err == nil {
		t.Error("ResolveRun accepted an invalid agent")
	}
	if _, err := ResolveAgent(root, PhaseReview); err == nil {
		t.Error("ResolveAgent accepted an invalid agent")
	}
}
