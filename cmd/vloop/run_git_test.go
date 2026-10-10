package main

import (
	"os"
	"path/filepath"
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
