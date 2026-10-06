package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunSymlinkedProposal: a proposal.json that is a symlink is read as
// missing, and what it points at never reaches .vloop/state/.
func TestRunSymlinkedProposal(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript+`if [ "$PHASE" = work ]; then
  printf '{"schema":"proposal/v1","task":"%s","outcome":"done","summary":"SECRET: made","files":[],"verified":"ok","notes":"none"}\n' "$TASK" > "$dir/outside.json"
  rm -f .vloop/tmp/proposal.json; ln -s "$dir/outside.json" .vloop/tmp/proposal.json
fi
`)
	res := r.vloop("run", "--max-iterations", "1", runBrief)
	wantIn(t, "output", r.outputs(res), "no valid proposal")
	r.wantTask("T1", "pending", 1)
	err := filepath.WalkDir(filepath.Join(r.dir, ".vloop", "state"), func(p string, d os.DirEntry, err error) error {
		if err == nil && d.Type().IsRegular() {
			if b, _ := os.ReadFile(p); strings.Contains(string(b), "SECRET") {
				t.Errorf("SECRET reached %s", p)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestRunBadRunID: a committed plan whose run_id is not a plain name is refused
// before anything is written.
func TestRunBadRunID(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript)
	wantExit(t, r.vloop("run", "--plan-only", runBrief), 0)
	plan := strings.Replace(r.read(".vloop/state/state.json"), `"run_id": "`+runID+`"`, `"run_id": "../x"`, 1)
	if !strings.Contains(plan, `"../x"`) {
		t.Fatalf("fixture: run_id not replaced in\n%s", plan)
	}
	r.write(".vloop/state/state.json", plan)
	r.commitAll("bad run id")
	head, n := r.head(), r.sessions()
	res := r.vloop("run")
	wantExit(t, res, 1)
	wantIn(t, "stderr", res.err, "vloop: ")
	if r.head() != head || r.sessions() != n {
		t.Error("the refused run committed or ran a session")
	}
	if r.has("../x") {
		t.Error("something was written outside the repository")
	}
	if r.has(".vloop/tmp/.running") {
		t.Error("a lock was left behind")
	}
}
