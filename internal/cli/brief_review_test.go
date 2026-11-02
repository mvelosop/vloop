package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBriefReviewRefusalsRunNoSession(t *testing.T) {
	d := scratchRepo(t)
	// A repository without .vloop is not a vloop repository.
	code, _, errOut := run(t, "-C", d, "brief", "review", "x.loop-brief.md")
	if code != 1 || !strings.Contains(errOut, "vloop: not in a vloop repository — run vloop init first") {
		t.Errorf("no .vloop: exit %d %q", code, errOut)
	}
	if err := os.Mkdir(filepath.Join(d, ".vloop"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "notes.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, errOut = run(t, "-C", d, "brief", "review", "notes.md")
	if code != 2 || !strings.Contains(errOut, "vloop: not a loop brief: notes.md") {
		t.Errorf("notes.md: exit %d %q", code, errOut)
	}
	code, _, errOut = run(t, "-C", d, "brief", "review", "docs/briefs/B1-x.loop-brief.md")
	if code != 1 || !strings.Contains(errOut, "vloop: no such file: docs/briefs/B1-x.loop-brief.md") {
		t.Errorf("missing: exit %d %q", code, errOut)
	}
	if code, _, _ = run(t, "-C", d, "brief", "review"); code != 2 {
		t.Errorf("no argument: exit %d, want 2", code)
	}
	if code, _, _ = run(t, "-C", d, "brief", "review", "a", "b"); code != 2 {
		t.Errorf("two arguments: exit %d, want 2", code)
	}
	if _, err := os.Stat(filepath.Join(d, ".vloop", "state")); err == nil {
		t.Error("a refusal made a state folder")
	}
}

func TestConfigSetAgentReviewBrief(t *testing.T) {
	d := scratchRepo(t)
	if code, _, _ := run(t, "-C", d, "config", "set", "agent.review-brief", "copilot"); code != 0 {
		t.Errorf("copilot: exit %d", code)
	}
	if code, _, _ := run(t, "-C", d, "config", "set", "agent.review-brief", "codex"); code != 2 {
		t.Errorf("codex: exit %d, want 2", code)
	}
}
