package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// The snapshot each iteration leaves and the summary every run ends with.

func (r *runRepo) snapshotAt(rev string) map[string]any {
	r.t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(r.git("show", rev+":.vloop/state/metrics/"+runID+".json")), &m); err != nil {
		r.t.Fatalf("snapshot at %s: %v", rev, err)
	}
	return m
}

func tasksDone(m map[string]any) float64 { return m["tasks"].(map[string]any)["done"].(float64) }

func TestRunSnapshot(t *testing.T) {
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript)
	wantExit(t, r.vloop("run", runBrief), 0)

	// The snapshot written after T1's commit rides in T2's commit.
	var t2 string
	for _, l := range strings.Split(strings.TrimSpace(r.git("log", "--format=%H %s")), "\n") {
		if strings.HasSuffix(l, "] T2: done") {
			t2 = strings.Fields(l)[0]
		}
	}
	if t2 == "" {
		t.Fatalf("no T2 commit in:\n%s", r.git("log", "--format=%s"))
	}
	if got := tasksDone(r.snapshotAt(t2)); got != 1 {
		t.Errorf("the snapshot in T2's commit has tasks.done %v, want 1 (as of T1's commit)", got)
	}
	if !strings.Contains(r.git("show", "--name-only", "--format=", t2), ".vloop/state/metrics/"+runID+".json") {
		t.Errorf("T2's commit does not include the snapshot")
	}
	final := r.snapshotAt("HEAD")
	if tasksDone(final) != 2 || final["schema"] != "metrics/v1" || final["run_id"] != runID {
		t.Errorf("the closing snapshot = %v", final)
	}
	r.wantClean()
}

func TestRunSummary(t *testing.T) {
	r := newRunRepo(t)
	r.scripted(twoTasks(t), defaultScript)
	res := r.vloop("run", runBrief)
	wantExit(t, res, 0)
	wantIn(t, "stdout", res.out, "· 2 done · 0 blocked")

	t.Run("a halted run prints it too", func(t *testing.T) {
		r := newRunRepo(t)
		r.scripted(twoTasks(t), defaultScript)
		res := r.vloop("run", "--max-iterations", "1", runBrief)
		wantExit(t, res, 4)
		wantIn(t, "stdout", res.out, "· 1 done · 0 blocked")
		r.wantClean()
	})
}

// A flaky gate costs the environment a defect, not the work.
func TestRunSnapshotFlakyDefect(t *testing.T) {
	r := newRunRepo(t)
	seen := r.stub + "/flaky.seen"
	verify := `test -f T1.out && { [ -f "` + seen + `" ] || { : > "` + seen + `"; exit 1; }; }`
	r.scripted(planJSON(t, planTask("T1", map[string]any{"verify": verify})), defaultScript)
	wantExit(t, r.vloop("run", runBrief), 0)
	res := r.vloop("defect", "list", "--matrix", "--json", "--brief", runBriefName)
	wantExit(t, res, 0)
	var m [][]int
	if err := json.Unmarshal([]byte(res.out), &m); err != nil {
		t.Fatal(err)
	}
	if m[3][0] != 1 || m[2][0] != 0 {
		t.Errorf("matrix %v, want one env defect found by the gate and no work defect", m)
	}
}
