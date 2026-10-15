package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// The driver's git ignores hooks and fsmonitor, and a run halts, committing
// nothing, when a session or gate changes .git/config, the hooks or the refs.

func TestRunHooksPlanted(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), defaultScript+`if [ "$PHASE" = work ]; then printf '#!/bin/sh\ngit branch evil\n' > .git/hooks/post-commit; chmod +x .git/hooks/post-commit; fi
`)
	res := r.runWith(nil, runBrief)
	wantExit(t, res, 9)
	wantIn(t, "stderr", res.err, "vloop: work changed the git hooks — nothing was committed; restore it, then re-run")
	if out := r.git("branch", "--list", "evil"); strings.TrimSpace(out) != "" {
		t.Errorf("the planted post-commit hook ran: %s", out)
	}
	for _, s := range r.subjects("--all") {
		if strings.HasPrefix(s, "[vloop] T1:") {
			t.Errorf("T1 was committed after the hooks changed")
		}
	}
}

func TestRunGitConfigPlanted(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), defaultScript+`if [ "$PHASE" = work ]; then printf '#!/bin/sh\ntouch fsm.ran\n' > fsm.sh; chmod +x fsm.sh; printf '[core]\n\tfsmonitor = ./fsm.sh\n' >> .git/config; fi
`)
	res := r.runWith(nil, runBrief)
	wantExit(t, res, 9)
	wantIn(t, "stderr", res.err, "vloop: work changed .git/config (core.fsmonitor) — nothing was committed; restore it, then re-run")
	if _, err := os.Stat(filepath.Join(r.dir, "fsm.ran")); err == nil {
		t.Errorf("the planted fsmonitor ran inside the driver's git")
	}
}

func TestRunGateMovesRefs(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(planJSON(t, planTask("T1", map[string]any{"verify": "git commit --allow-empty -qm x && true"})), defaultScript)
	res := r.runWith(nil, runBrief)
	wantExit(t, res, 9)
	wantIn(t, "stderr", res.err, "vloop: the gate of T1 moved git refs — nothing was committed; restore them, then re-run")
	for _, s := range r.subjects("--all") {
		if strings.HasPrefix(s, "[vloop] T1:") {
			t.Errorf("T1 was committed after its gate moved refs")
		}
	}
}

// The guard compares .git/config by its keys: an editor's bookkeeping key does
// not stop a run, any other changed key does and is named.
func TestRunGitConfigIgnoredKeyContinues(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), defaultScript+`if [ "$PHASE" = work ]; then printf '[branch "x"]\n\tvscode-merge-base = refs/heads/main\n' >> .git/config; fi
`)
	wantExit(t, r.runWith(nil, runBrief), 0)
}

func TestRunGitConfigIgnoreFromEnvironment(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), defaultScript+`if [ "$PHASE" = work ]; then git config gitlens.x 1; fi
`)
	wantExit(t, r.runWith([]string{"VLOOP_RUN_GIT_CONFIG_IGNORE=gitlens.*"}, runBrief), 0)
}

func TestRunGitConfigProtectedPatternRefused(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), defaultScript)
	res := r.runWith([]string{"VLOOP_RUN_GIT_CONFIG_IGNORE=gitlens.*,core.*"}, runBrief)
	wantExit(t, res, 1)
	wantIn(t, "stderr", res.err, "vloop: run.git-config-ignore: core.* would ignore core.* — the guard must see it")
}

// A halt during acceptance leaves the plan stamped and planning with its run
// folder, and the re-run resumes acceptance: no second plan session, and the
// gate review runs before any work.
func TestRunBaseGateHaltResumesAcceptance(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(planJSON(t, planTask("T1", map[string]any{
		"verify": "if [ -f .git/tamper ]; then rm -f .git/tamper; printf '[core]\\n\\tfsmonitor = false\\n' >> .git/config; fi; test -f T1.out"})), defaultScript)
	cfg := filepath.Join(r.dir, ".git", "config")
	orig, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(r.dir, ".vloop", "tmp"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(r.dir, ".git", "tamper"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	res := r.vloop("run", runBrief)
	wantExit(t, res, 9)
	wantIn(t, "stderr", res.err, "the gate of T1 changed .git/config")
	r.wantStatus("planning")
	if err := os.WriteFile(cfg, orig, 0o644); err != nil {
		t.Fatal(err)
	}
	res = r.vloop("run", runBrief)
	wantExit(t, res, 0)
	wantIn(t, "output", r.outputs(res), "resuming acceptance of "+runID+": round 1")
	if n := r.planSessions(); n != 1 {
		t.Errorf("%d plan sessions after the resume, want 1", n)
	}
	if n := r.gateReviews(); n != 1 {
		t.Errorf("%d gate review sessions, want 1", n)
	}
	argv := strings.Join(r.argv(), "\n")
	if strings.Index(argv, "/vloop:gate-review") > strings.Index(argv, "/vloop:work") {
		t.Error("work started before the gate review passed")
	}
}

func TestRunGateReviewHaltResumesAcceptance(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(oneTask(t), defaultScript+`if [ "$PHASE" = gate-review ] && [ ! -f .git/reviewed ]; then
  : > .git/reviewed
  printf '[core]\n\tfsmonitor = false\n' >> .git/config
fi
`)
	cfg := filepath.Join(r.dir, ".git", "config")
	orig, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	wantExit(t, r.vloop("run", runBrief), 9)
	r.wantStatus("planning")
	if err := os.WriteFile(cfg, orig, 0o644); err != nil {
		t.Fatal(err)
	}
	res := r.vloop("run")
	wantExit(t, res, 0)
	wantIn(t, "output", r.outputs(res), "resuming acceptance of "+runID+": round 1")
	if n := r.planSessions(); n != 1 {
		t.Errorf("%d plan sessions after the resume, want 1", n)
	}
	if n := r.gateReviews(); n != 2 {
		t.Errorf("%d gate review sessions, want 2", n)
	}
	argv := strings.Join(r.argv(), "\n")
	if strings.LastIndex(argv, "/vloop:gate-review") > strings.Index(argv, "/vloop:work") {
		t.Error("work started before the gate review passed")
	}
	r.wantTask("T1", "done", 0)
}

// S1: the driver empties .vloop/tmp/ of everything but .running at run start,
// before every work session and before the final pass, and at no other point:
// a plan round's scratch and plan-problems.md reach the next round, and
// proposal.json, checks.json and a work session's scratch reach its review.
func TestRunClearsTmpAtItsPointsOnly(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.write(".vloop/config.toml", runCheckConfig+`
[[check]]
name = "late"
paths = ["nomatch/**"]
run = "test ! -e .vloop/tmp/late.txt"
`)
	r.commitAll("late check")
	keep := filepath.Join(t.TempDir(), "keep.txt")
	if err := os.WriteFile(keep, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	tmp := filepath.Join(r.dir, ".vloop", "tmp")
	if err := os.MkdirAll(filepath.Join(tmp, "olddir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "old.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(keep, filepath.Join(tmp, "linkfile")); err != nil {
		t.Skipf("no symbolic links: %v", err)
	}
	r.rounds(planJSON(t, planTask("T1", map[string]any{"verify": "true"})),
		planJSON(t, planTask("T1", nil)))
	r.script(strings.Replace(defaultScript, `cp "$dir/plan.json" .vloop/state/state.json`,
		`n=$(cat "$dir/pn" 2>/dev/null || echo 0); n=$((n+1)); echo $n > "$dir/pn"; cp "$dir/plan.$n.json" .vloop/state/state.json`, 1) + `
seen() { [ -e ".vloop/tmp/$1" ] || [ -L ".vloop/tmp/$1" ]; }
note() { echo "$*" >> "$dir/s1"; }
seen .running || note "$PHASE: .running missing"
case "$PHASE" in
  plan)
    n=$(cat "$dir/pn")
    for e in old.txt olddir linkfile; do seen $e && note "plan round $n saw the stale $e"; done
    if [ "$n" = 1 ]; then : > .vloop/tmp/plan-scratch.txt
    else
      seen plan-problems.md || note "plan round $n: plan-problems.md was cleared"
      seen plan-scratch.txt || note "plan round $n: round 1's scratch was cleared"
    fi;;
  work)
    for e in old.txt olddir linkfile plan-scratch.txt plan-problems.md; do seen $e && note "work saw $e"; done
    : > .vloop/tmp/w1.txt; : > .vloop/tmp/late.txt;;
  review)
    for e in proposal.json checks.json w1.txt; do seen $e || note "review found no $e"; done;;
esac
`)
	wantExit(t, r.vloop("run", runBrief), 0)
	if b, err := os.ReadFile(filepath.Join(r.stub, "s1")); err == nil {
		t.Errorf("the scratch was not cleared as S1 requires:\n%s", b)
	}
	if n := r.planSessions(); n != 2 {
		t.Errorf("%d plan sessions, want 2", n)
	}
	r.wantTask("T1", "done", 0)
	if _, err := os.Stat(keep); err != nil {
		t.Errorf("clearing followed a symbolic link out of .vloop/tmp/: %v", err)
	}
	for _, e := range []string{"old.txt", "olddir", "linkfile", "late.txt"} {
		if _, err := os.Lstat(filepath.Join(tmp, e)); err == nil {
			t.Errorf(".vloop/tmp/%s outlived the run", e)
		}
	}
}

// S1: an entry that cannot be removed is a warning and the run goes on.
func TestRunClearsTmpWarnsOnUnremovableEntry(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("directory permissions do not stop removal here")
	}
	r := newRunRepo(t)
	locked := filepath.Join(r.dir, ".vloop", "tmp", "locked")
	if err := os.MkdirAll(locked, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(locked, "f"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(locked, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(locked, 0o755) })
	r.scripted(oneTask(t), defaultScript)
	res := r.vloop("run", runBrief)
	wantExit(t, res, 0)
	wantIn(t, "output", r.outputs(res), "warning: cannot clear .vloop/tmp/locked")
}
