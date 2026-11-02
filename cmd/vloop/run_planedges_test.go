package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const readmeBase = "the fixture's readme\n"

// planEdgeRepo is a run repository with a README.md committed.
func planEdgeRepo(t *testing.T) *runRepo {
	t.Helper()
	r := newRunRepo(t)
	r.write("README.md", readmeBase)
	r.commitAll("harness: readme")
	return r
}

func (r *runRepo) planCommitFiles() string {
	r.t.Helper()
	for _, l := range strings.Split(strings.TrimSpace(r.git("log", "--format=%H %s")), "\n") {
		if h, s, ok := strings.Cut(l, " "); ok && strings.HasPrefix(s, "[vloop] plan ") {
			return r.git("show", "--name-only", "--format=", h)
		}
	}
	r.t.Fatal("no plan commit")
	return ""
}

func TestPlanPhaseRevertsStrayChanges(t *testing.T) {
	t.Parallel()
	r := planEdgeRepo(t)
	r.scripted(oneTask(t), defaultScript+`if [ "$PHASE" = plan ]; then echo plan-line >> README.md; echo notes > notes.txt; fi
if [ "$PHASE" = gate-review ]; then echo scratch > scratch.txt; echo gr-line >> README.md; fi
`)
	res := r.vloop("run", runBrief)
	wantExit(t, res, 0)
	want := []string{
		"vloop: the plan session changed README.md — reverted",
		"vloop: the plan session changed notes.txt — reverted",
		"vloop: the gate-review session changed README.md — reverted",
		"vloop: the gate-review session changed scratch.txt — reverted",
	}
	var got []string
	for _, l := range strings.Split(res.err, "\n") {
		if strings.HasPrefix(l, "vloop: the plan session changed ") || strings.HasPrefix(l, "vloop: the gate-review session changed ") {
			got = append(got, l)
		}
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("warnings on stderr:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	log, _ := os.ReadFile(filepath.Join(r.runFolder(), "run.log"))
	for _, w := range want {
		wantIn(t, "run.log", string(log), w)
	}
	files := r.planCommitFiles()
	wantIn(t, "plan commit", files, ".vloop/state/state.json")
	wantNotIn(t, "plan commit", files, "README.md", "notes.txt", "scratch.txt")
	if r.read("README.md") != readmeBase {
		t.Errorf("README.md is %q, want the original", r.read("README.md"))
	}
	if r.has("notes.txt") || r.has("scratch.txt") {
		t.Error("notes.txt or scratch.txt survived")
	}
}

func TestPlanPhaseRevertsBeforeFailureHalt(t *testing.T) {
	t.Parallel()
	r := planEdgeRepo(t)
	r.scripted(oneTask(t), defaultScript+`if [ "$PHASE" = plan ]; then echo plan-line >> README.md; STUB_EXIT=3; fi
`)
	res := r.vloop("run", runBrief)
	wantExit(t, res, 1)
	w := strings.Index(res.err, "vloop: the plan session changed README.md — reverted")
	h := strings.Index(res.err, "planning session failed (claude exited 3)")
	if w < 0 || h < 0 || w > h {
		t.Errorf("want the revert warning before the halt, got:\n%s", res.err)
	}
	if r.read("README.md") != readmeBase {
		t.Errorf("README.md is %q, want the original", r.read("README.md"))
	}
	for _, l := range strings.Split(r.git("status", "--porcelain", "--untracked-files=all"), "\n") {
		if l != "" && !strings.Contains(l, " .vloop/") {
			t.Errorf("git status lists %q outside .vloop/", l)
		}
	}
}

func TestPlanPhaseRevertsBeforeBaseGates(t *testing.T) {
	t.Parallel()
	r := planEdgeRepo(t)
	r.scripted(planJSON(t, planTask("T1", map[string]any{"verify": "grep -q plan-line README.md"})),
		defaultScript+`if [ "$PHASE" = plan ]; then echo plan-line >> README.md; fi
if [ "$PHASE" = work ]; then echo plan-line >> README.md; fi
`)
	res := r.vloop("run", runBrief)
	wantExit(t, res, 0)
	if n := r.planSessions(); n != 1 {
		t.Errorf("%d plan sessions, want 1", n)
	}
	wantIn(t, "stderr", res.err, "vloop: the plan session changed README.md — reverted")
	wantNotIn(t, "plan commit", r.planCommitFiles(), "README.md")
	if n := strings.Count(r.read("README.md"), "plan-line"); n != 1 {
		t.Errorf("README.md holds plan-line %d times, want 1", n)
	}
}

func TestPlanPhaseCleanSessionNoWarning(t *testing.T) {
	t.Parallel()
	r := planEdgeRepo(t)
	r.scripted(oneTask(t), defaultScript+`if [ "$PHASE" = plan ]; then mkdir -p .vloop/state/gates/T1 .vloop/tmp; echo 'test -f T1.out' > .vloop/state/gates/T1/gate.sh; fi
`)
	res := r.vloop("run", runBrief)
	wantExit(t, res, 0)
	wantNotIn(t, "output", res.out+res.err, "session changed")
	files := r.planCommitFiles()
	wantIn(t, "plan commit", files, ".vloop/state/state.json", ".vloop/state/gates/T1/gate.sh")
}
