package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestWorkedExampleB8 plays every line of B8's worked example with the stub
// claude, one subtest per line, each in its own temporary repository. The
// per-feature tests pin the details; this one pins that the example holds as
// the brief wrote it.
func TestWorkedExampleB8(t *testing.T) {
	t.Run("newest draft, older ready: plans the ready one", func(t *testing.T) {
		r := newRunRepo(t)
		r.write("docs/briefs/B20260102-0900-newer.loop-brief.md",
			strings.Replace(runBriefText(), "status: ready", "status: draft", 1))
		r.commitAll("a newer draft")
		r.scripted(oneTask(t), defaultScript)
		wantExit(t, r.vloop("run"), 0)
		wantIn(t, "argv", strings.Join(r.argv(), "\n"), "-p /vloop:plan "+runBrief+" ")
		r.wantTask("T1", "done", 0)
	})

	t.Run("only drafts: no ready brief", func(t *testing.T) {
		r := newRunRepo(t)
		r.write(runBrief, strings.Replace(runBriefText(), "status: ready", "status: draft", 1))
		r.commitAll("all drafts")
		res := r.vloop("run")
		wantExit(t, res, 1)
		wantIn(t, "stderr", res.err, "vloop: no ready brief in docs/briefs/ — name one, or set status: ready on a checked brief")
		if r.sessions() != 0 {
			t.Error("a session ran")
		}
	})

	t.Run("untracked .env: the tree is not clean", func(t *testing.T) {
		r := newRunRepo(t)
		r.write(".env", "API_KEY=x\n")
		res := r.vloop("run", runBrief)
		wantExit(t, res, 1)
		wantIn(t, "stderr", res.err, "vloop: the tree is not clean — commit, ignore or remove these first: .env")
		if r.sessions() != 0 {
			t.Error("a session ran")
		}
	})

	t.Run("verify sleep 1000 & exit 0: the gate returns, no sleep remains", func(t *testing.T) {
		r := newRunRepo(t)
		// A distinctive duration finds this test's sleep among the machine's.
		verify := "sleep 1017 & exit 0"
		r.scripted(planJSON(t, planTask("T1", map[string]any{"verify": verify})), defaultScript)
		start := time.Now()
		wantExit(t, r.vloop("run", runBrief), 0)
		if d := time.Since(start); d > 30*time.Second {
			t.Errorf("the run took %s; the gate did not return", d)
		}
		r.wantTask("T1", "done", 0)
		if out, _ := exec.Command("pgrep", "-f", "sleep 1017").Output(); strings.TrimSpace(string(out)) != "" {
			t.Errorf("a sleep process remains: %s", out)
		}
	})

	t.Run("verify sleeps past the gate timeout: the log ends with the timeout line", func(t *testing.T) {
		t.Parallel() // the shortest timeout is a minute
		r := newRunRepo(t)
		r.scripted(planJSON(t, planTask("T1", map[string]any{"verify": "echo started; sleep 1018"})), defaultScript)
		res := r.runWith([]string{"VLOOP_RUN_GATE_TIMEOUT=1"}, "--max-iterations", "1", runBrief)
		wantExit(t, res, 4)
		b, err := os.ReadFile(filepath.Join(r.runFolder(), "gates", "T1.log"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasSuffix(string(b), "vloop: gate T1 timed out after 1 min\n") {
			t.Errorf("gate log: %q", b)
		}
		r.wantIterations("T1:gate_failed")
		if out, _ := exec.Command("pgrep", "-f", "sleep 1018").Output(); strings.TrimSpace(string(out)) != "" {
			t.Errorf("a sleep process remains: %s", out)
		}
	})

	t.Run("work writes a post-commit hook creating branch evil", func(t *testing.T) {
		r := newRunRepo(t)
		r.scripted(oneTask(t), defaultScript+`if [ "$PHASE" = work ]; then printf '#!/bin/sh\ngit branch evil\n' > .git/hooks/post-commit; chmod +x .git/hooks/post-commit; fi
`)
		res := r.runWith(nil, runBrief)
		wantExit(t, res, 9)
		wantIn(t, "stderr", res.err, "vloop: work changed the git hooks — nothing was committed; restore it, then re-run")
		if out := r.git("branch", "--list", "evil"); strings.TrimSpace(out) != "" {
			t.Errorf("git branch --list evil printed %q", out)
		}
	})

	t.Run("a gate that moves refs", func(t *testing.T) {
		r := newRunRepo(t)
		r.scripted(planJSON(t, planTask("T1", map[string]any{"verify": "git commit --allow-empty -qm x && true"})), defaultScript)
		res := r.runWith(nil, runBrief)
		wantExit(t, res, 9)
		wantIn(t, "stderr", res.err, "vloop: the gate of T1 moved git refs — nothing was committed; restore them, then re-run")
	})

	t.Run("work zeroes earlier cost_usd: cost ceiling reached", func(t *testing.T) {
		r := newRunRepo(t)
		r.scripted(planJSON(t, planTask("T1", nil), planTask("T2", map[string]any{"depends_on": []string{"T1"}}),
			planTask("T3", map[string]any{"depends_on": []string{"T2"}})), defaultScript+`if [ "$PHASE" = work ] && [ "$TASK" != T1 ]; then
  for f in .vloop/state/runs/*/*/sessions/*.json; do sed 's/"cost_usd": [0-9.]*/"cost_usd": 0/' "$f" > "$f.z" && mv "$f.z" "$f"; done
fi
`)
		res := r.runWith(nil, "--cost-ceiling", "0.35", runBrief)
		wantExit(t, res, 6)
		wantIn(t, "outputs", r.outputs(res), "cost ceiling")
		r.wantTask("T3", "pending", 0)
	})

	t.Run("work edits state.json verify and runs vloop task gate", func(t *testing.T) {
		r := newRunRepo(t)
		r.scripted(planJSON(t, planTask("T1", map[string]any{"verify": "test -f T1.out"})), defaultScript+`if [ "$PHASE" = work ] && [ "$ATTEMPT" = 1 ]; then
  sed 's/"verify": "test -f T1.out"/"verify": "true"/' .vloop/state/state.json > st.tmp && mv st.tmp .vloop/state/state.json
  mkdir -p .vloop/tmp
  '`+binPath+`' task gate T1 > .vloop/tmp/inside.out 2> .vloop/tmp/inside.err; echo $? > .vloop/tmp/inside.code
  cp .vloop/tmp/inside.code "$dir/inside.code"; cp .vloop/tmp/inside.err "$dir/inside.err"
fi
`)
		wantExit(t, r.runWith(nil, runBrief), 0)
		code, err := os.ReadFile(filepath.Join(r.stub, "inside.code"))
		if err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(string(code)) != "1" {
			t.Errorf("vloop task gate inside the session exited %s, want 1", code)
		}
		msg, _ := os.ReadFile(filepath.Join(r.stub, "inside.err"))
		if string(msg) != "vloop: the plan was changed during this session — gates run only from the plan the driver holds\n" {
			t.Errorf("stderr inside the session: %q", msg)
		}
		if v := r.task("T1")["verify"]; v != "test -f T1.out" {
			t.Errorf("the authored verify did not survive: %v", v)
		}
	})

	t.Run("review edits src and returns PASS: reverted, FAIL, not done", func(t *testing.T) {
		r := newRunRepo(t)
		r.commitFile("src.txt", "orig\n")
		r.scripted(oneTask(t), defaultScript+`if [ "$PHASE" = review ] && [ "$ATTEMPT" = 1 ]; then echo tampered >> src.txt; fi
`)
		wantExit(t, r.runWith(nil, runBrief), 0)
		r.wantIterations("T1:rejected", "T1:done")
		wantIn(t, "run log", r.runLog(), "vloop: the review session changed src.txt — reverted")
		if got := r.read("src.txt"); got != "orig\n" {
			t.Errorf("src.txt not reverted: %q", got)
		}
		// Not done after the tampered review: the first iteration was rejected.
		if got := r.iterations()[0]["outcome"]; got != "rejected" {
			t.Errorf("first iteration outcome %v, want rejected", got)
		}
	})

	t.Run("runner user us, session costing $1.50: record parses, spend 1.50", func(t *testing.T) {
		r := newRunRepo(t)
		short := filepath.Join(filepath.Dir(r.home), "us")
		if err := os.Rename(r.home, short); err != nil {
			t.Fatal(err)
		}
		r.home = short
		r.scripted(oneTask(t), "STUB_COST=1.50\n"+defaultScript)
		wantExit(t, r.vloop("run", runBrief), 0)
		names := r.sessionFiles()
		if len(names) == 0 {
			t.Fatal("no session records")
		}
		var total float64
		for _, n := range names {
			var rec map[string]any
			sessionRecord(t, filepath.Join(r.runFolder(), "sessions", n), &rec)
			c, ok := rec["cost_usd"].(float64)
			if !ok || c != 1.5 {
				t.Errorf("%s: cost_usd = %v, want 1.5", n, rec["cost_usd"])
			}
			total += c
		}
		if total != 1.5*float64(len(names)) {
			t.Errorf("spend %.2f over %d sessions", total, len(names))
		}
	})

	t.Run("gate env with FOO_TOKEN: the log is redacted", func(t *testing.T) {
		r := newRunRepo(t)
		r.scripted(planJSON(t, planTask("T1", map[string]any{"verify": "env"})), defaultScript)
		wantExit(t, r.runWith([]string{"FOO_TOKEN=s3cr3t-value"}, runBrief), 0)
		log, err := os.ReadFile(filepath.Join(r.runFolder(), "gates", "T1.log"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(log), "<redacted:FOO_TOKEN>") || strings.Contains(string(log), "s3cr3t-value") {
			t.Errorf("gate log:\n%s", log)
		}
	})

	t.Run("two vloop run at once: one runs, the other exits 1 naming the pid", func(t *testing.T) {
		r := newRunRepo(t)
		r.scripted(oneTask(t), defaultScript+`if [ "$PHASE" = work ]; then sleep 3; fi
`)
		first := exec.Command(binPath, "run", runBrief)
		first.Dir = r.dir
		first.Env = r.env()
		if err := first.Start(); err != nil {
			t.Fatal(err)
		}
		waited := false
		defer func() {
			if !waited {
				first.Process.Kill()
				first.Wait()
			}
		}()
		deadline := time.Now().Add(20 * time.Second)
		for !r.has(".vloop/tmp/.running") {
			if time.Now().After(deadline) {
				t.Fatal("the first run never took the lock")
			}
			time.Sleep(20 * time.Millisecond)
		}
		res := r.vloop("run", runBrief)
		wantExit(t, res, 1)
		wantIn(t, "stderr", res.err, "a loop is already running in this working tree", strconv.Itoa(first.Process.Pid))
		waited = true
		if err := first.Wait(); err != nil {
			t.Errorf("the first run: %v", err)
		}
		r.wantTask("T1", "done", 0)
	})
}
