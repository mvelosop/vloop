package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/mvelosop/vloop/internal/cli"
)

// TestWorkedExampleB12 plays the field-reliability brief's worked example line
// for line against the built binary and the suite's stub claude, one temporary
// repository per scenario, addressed with git -C. Values the binary computes
// (ids, paths, costs, counts) are read back from its records, not hardcoded;
// no scenario sleeps longer than 3 seconds.

// b12Records reads the session records of the brief's run folders whose file
// name ends -<phase>.json, in run-folder then sequence order.
func (r *runRepo) b12Records(phase string) []map[string]any {
	r.t.Helper()
	paths, _ := filepath.Glob(filepath.Join(r.dir, ".vloop", "state", "runs", runID, "*", "sessions", "*-"+phase+".json"))
	var out []map[string]any
	for _, p := range paths {
		raw, err := os.ReadFile(p)
		if err != nil {
			r.t.Fatal(err)
		}
		var rec map[string]any
		if err := json.Unmarshal(raw, &rec); err != nil {
			r.t.Fatalf("%s: %v", p, err)
		}
		out = append(out, rec)
	}
	return out
}

func (r *runRepo) b12Metrics(args ...string) map[string]any {
	r.t.Helper()
	res := r.vloop(append([]string{"metrics", "--json", runID}, args...)...)
	wantExit(r.t, res, 0)
	var m map[string]any
	if err := json.Unmarshal([]byte(res.out), &m); err != nil {
		r.t.Fatalf("metrics --json: %v\n%s", err, res.out)
	}
	return m
}

func b12Num(t *testing.T, m map[string]any, path ...string) float64 {
	t.Helper()
	var v any = m
	for _, k := range path {
		mm, ok := v.(map[string]any)
		if !ok {
			t.Fatalf("no %v in metrics", path)
		}
		v = mm[k]
	}
	f, ok := v.(float64)
	if !ok {
		t.Fatalf("metrics %v = %v, not a number", path, v)
	}
	return f
}

func TestWorkedExampleB12(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("the stub claude is a POSIX shell script and cannot run on windows")
	}

	t.Run("guard keys", func(t *testing.T) {
		t.Parallel()
		// An editor's key appended to .git/config does not stop the run.
		r := newRunRepo(t)
		r.scripted(oneTask(t), defaultScript+`if [ "$PHASE" = work ]; then printf '[branch "B1"]\n\tvscode-merge-base = refs/heads/main\n' >> .git/config; fi
`)
		res := r.runWith(nil, runBrief)
		wantExit(t, res, 0)

		// Any other key stops it and is named.
		r = newRunRepo(t)
		r.scripted(oneTask(t), defaultScript+`if [ "$PHASE" = work ]; then printf '[core]\n\tfsmonitor = true\n' >> .git/config; fi
`)
		res = r.runWith(nil, runBrief)
		wantExit(t, res, 9)
		wantIn(t, "stderr", res.err, "vloop: work changed .git/config (core.fsmonitor) — nothing was committed; restore it, then re-run")

		// A configured pattern is ignored.
		r = newRunRepo(t)
		r.write(".vloop/config.toml", "run.git-config-ignore = [\"gitlens.*\"]\n"+runCheckConfig)
		r.commitAll("config")
		r.scripted(oneTask(t), defaultScript+`if [ "$PHASE" = work ]; then git config gitlens.x 1; fi
`)
		wantExit(t, r.runWith(nil, runBrief), 0)

		// A pattern that would ignore a protected key is refused at set time.
		before := r.read(".vloop/config.toml")
		res = r.vloop("config", "set", "run.git-config-ignore", "core.*")
		wantExit(t, res, 2)
		wantIn(t, "stderr", res.err, "vloop: run.git-config-ignore: core.* would ignore core.* — the guard must see it")
		if after := r.read(".vloop/config.toml"); after != before {
			t.Errorf("the refused config set changed the config:\n%s", after)
		}
	})

	t.Run("halt in acceptance resumes", func(t *testing.T) {
		t.Parallel()
		r := newRunRepo(t)
		r.scripted(planJSON(t, planTask("T1", map[string]any{
			"verify": "if [ -f .git/tamper ]; then rm -f .git/tamper; printf '[core]\\n\\tfsmonitor = true\\n' >> .git/config; fi; test -f T1.out"})),
			defaultScript+`if [ "$PHASE" = plan ]; then STUB_COST=0.37; fi
if [ "$PHASE" = gate-review ]; then STUB_COST=0.21; fi
`)
		cfg := filepath.Join(r.dir, ".git", "config")
		orig, err := os.ReadFile(cfg)
		if err != nil {
			t.Fatal(err)
		}
		r.write(".git/tamper", "")
		res := r.vloop("run", runBrief)
		wantExit(t, res, 9)
		wantIn(t, "stderr", res.err, "the gate of T1 changed .git/config (core.fsmonitor)")
		raw := r.read(".vloop/state/state.json")
		var fields map[string]any
		if err := json.Unmarshal([]byte(raw), &fields); err != nil {
			t.Fatal(err)
		}
		for _, k := range []string{"run_id", "brief", "branch", "status"} {
			if s, _ := fields[k].(string); s == "" {
				t.Errorf("state.json after the halt has no %s", k)
			}
		}
		if fields["status"] != "planning" {
			t.Errorf("status %v, want planning", fields["status"])
		}
		firstFolder := r.runDirs()
		if len(firstFolder) != 1 {
			t.Fatalf("run folders after the halt: %v", firstFolder)
		}
		if err := os.WriteFile(cfg, orig, 0o644); err != nil {
			t.Fatal(err)
		}
		time.Sleep(1100 * time.Millisecond) // the next run folder is named by the second

		res = r.vloop("run", runBrief)
		wantExit(t, res, 0)
		wantIn(t, "output", r.outputs(res), "resuming acceptance of "+runID+": round 1")
		if n := r.planSessions(); n != 1 {
			t.Errorf("%d plan sessions after the resume, want 1", n)
		}
		if n := r.gateReviews(); n != 1 {
			t.Errorf("%d gate review sessions, want 1", n)
		}
		dirs := r.runDirs()
		if len(dirs) != 2 {
			t.Fatalf("run folders after the resume: %v", dirs)
		}
		// The plan commit carries both run folders.
		var planSHA string
		for _, l := range strings.Split(strings.TrimSpace(r.git("log", "--format=%H %s")), "\n") {
			if f := strings.SplitN(l, " ", 2); len(f) == 2 && strings.HasPrefix(f[1], "[vloop] plan ") {
				planSHA = f[0]
			}
		}
		if planSHA == "" {
			t.Fatal("no plan commit")
		}
		files := r.git("show", "--name-only", "--format=", planSHA)
		for _, d := range dirs {
			rel := ".vloop/state/runs/" + runID + "/" + filepath.Base(d) + "/"
			if !strings.Contains(files, rel) {
				t.Errorf("the plan commit lacks run folder %s:\n%s", rel, files)
			}
		}
		// The metrics count the plan session of the first folder and the gate
		// review of the second.
		m := r.b12Metrics()
		var planUSD, reviewUSD float64
		for _, rec := range r.b12Records("plan") {
			planUSD += rec["cost_usd"].(float64)
		}
		for _, rec := range r.b12Records("gate-review") {
			reviewUSD += rec["cost_usd"].(float64)
		}
		if got := b12Num(t, m, "cost", "plan_usd"); got < planUSD-1e-9 || planUSD == 0 {
			t.Errorf("plan_usd %v does not count the first folder's plan session (%v)", got, planUSD)
		}
		if got := b12Num(t, m, "cost", "gate_review_usd"); got < reviewUSD-1e-9 || reviewUSD == 0 {
			t.Errorf("gate_review_usd %v does not count the second folder's gate review (%v)", got, reviewUSD)
		}
	})

	t.Run("record and masks", func(t *testing.T) {
		t.Parallel()
		r := newRunRepo(t)
		user := filepath.Base(r.home)
		long := strings.Repeat("x", 300)
		denials := fmt.Sprintf(`[{"tool_name":"Bash","tool_input":{"command":"rm -rf $X/tmp"}},`+
			`{"tool_name":"Bash","tool_input":{"command":"ls /Users/%s/repo"}},`+
			`{"tool_name":"Bash","tool_input":{"command":"%s"}},`+
			`{"tool_name":"Read","tool_input":{"file_path":"~/.claude/projects/-Users-%s-repo/x"}}]`, user, long, user)
		r.write(".vloop/config.toml", "[metrics]\nstacks = [\"go\"]\n\n"+runCheckConfig)
		r.commitAll("config")
		r.scripted(oneTask(t), defaultScript+`if [ "$PHASE" = plan ]; then STUB_SLEEP=2; STUB_DURATION=20; fi
if [ "$PHASE" = work ]; then
  STUB_COST=1.84
  STUB_DENIALS='`+denials+`'
  printf '#!/bin/sh\necho one\necho two\n' > T1.sh; printf 'a\nb\nc\n' > notes.md
fi
`)
		res := r.runWith(nil, runBrief)
		wantExit(t, res, 0)

		// The per-session line.
		line := regexp.MustCompile(`(?m)^   work T1 · [0-9]+m[0-5][0-9]s · \$1\.84 · run \$([0-9]+\.[0-9]{2}) / \$40\.00$`)
		if !line.MatchString(res.out + res.err) {
			t.Errorf("no per-session line for work T1 at $1.84:\n%s", r.outputs(res))
		}

		// wall_ms against duration_ms.
		plans := r.b12Records("plan")
		if len(plans) != 1 {
			t.Fatalf("%d plan records", len(plans))
		}
		wall, _ := plans[0]["wall_ms"].(float64)
		dur, _ := plans[0]["duration_ms"].(float64)
		if wall < 2000 || dur != 20 {
			t.Errorf("plan record wall_ms %v duration_ms %v, want wall_ms >= 2000 and 20", wall, dur)
		}
		if _, ok := plans[0]["ended"].(string); !ok {
			t.Errorf("plan record has no ended: %v", plans[0])
		}
		m := r.b12Metrics()
		if got := b12Num(t, m, "time", "plan_ms"); got != wall {
			t.Errorf("metrics plan time %v, want the record's wall_ms %v", got, wall)
		}

		// The denied commands.
		work := r.b12Records("work")
		if len(work) != 1 {
			t.Fatalf("%d work records", len(work))
		}
		den, _ := work[0]["permission_denials"].([]any)
		if len(den) != 4 {
			t.Fatalf("permission_denials: %v", den)
		}
		cmdOf := func(i int) string {
			d, _ := den[i].(map[string]any)
			s, _ := d["command"].(string)
			return s
		}
		if first, _ := den[0].(map[string]any); first["tool_name"] != "Bash" || cmdOf(0) != "rm -rf $X/tmp" {
			t.Errorf("first denial: %v", den[0])
		}
		if c := cmdOf(1); strings.Contains(c, user) || !strings.HasPrefix(c, "ls ") {
			t.Errorf("a command with the user's name was kept as %q", c)
		}
		if c := cmdOf(2); c != long[:200]+"…" {
			t.Errorf("a 300-character command was kept as %d characters: %q", len([]rune(c)), c)
		}
		if p := fmt.Sprint(den[3]); !strings.Contains(p, "-Users-USER-repo") || strings.Contains(p, user) {
			t.Errorf("the dashed home was not masked: %s", p)
		}

		// n/a where nothing was classified.
		text := r.vloop("metrics", runID)
		wantExit(t, text, 0)
		wantIn(t, "metrics", text.out, "n/a", "delivered lines match no stack in metrics.stacks")
	})

	t.Run("cost warnings", func(t *testing.T) {
		t.Parallel()
		r := newRunRepo(t)
		r.write(".vloop/config.toml", "run.cost-ceiling = 4\n"+runCheckConfig)
		r.commitAll("config")
		r.scripted(oneTask(t), defaultScript+`STUB_COST=1.10
`)
		res := r.runWith(nil, runBrief)
		out := r.outputs(res)
		for _, w := range []string{"   cost: $2.20 spent, 50% of the $4.00 ceiling\n", "   cost: $3.30 spent, 80% of the $4.00 ceiling\n"} {
			if n := strings.Count(res.out+res.err, w); n != 1 {
				t.Errorf("warning %q printed %d times, want once:\n%s", w, n, out)
			}
		}
	})

	t.Run("status during a work session", func(t *testing.T) {
		t.Parallel()
		r := newRunRepo(t)
		r.scripted(oneTask(t), defaultScript+`if [ "$PHASE" = work ]; then STUB_SLEEP=3; fi
`)
		cmd := exec.Command(binPath, "run", runBrief)
		cmd.Dir = r.dir
		cmd.Env = r.env()
		var o bytes.Buffer
		cmd.Stdout, cmd.Stderr = &o, &o
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		running := regexp.MustCompile(`running: work T1 for 0m0[0-9]s · \$[0-9]+\.[0-9]{2} of \$40\.00`)
		var seen string
		deadline := time.Now().Add(10 * time.Second)
		for seen == "" && time.Now().Before(deadline) {
			select {
			case err := <-done:
				t.Fatalf("the run ended before status saw the work session: %v\n%s", err, o.String())
			case <-time.After(150 * time.Millisecond):
			}
			if res := r.vloop("status"); res.code == 0 {
				seen = running.FindString(res.out)
			}
		}
		if seen == "" {
			t.Errorf("status never showed the running work session")
		}
		if err := <-done; err != nil {
			t.Fatalf("run: %v\n%s", err, o.String())
		}
	})

	t.Run("version", func(t *testing.T) {
		t.Parallel()
		// A commit stamp reads as it did.
		dir := t.TempDir()
		bin := filepath.Join(dir, "vloop-stamped")
		build := exec.Command("go", "build", "-buildvcs=false", "-ldflags", "-X main.commit=abc1234", "-o", bin, ".")
		if out, err := build.CombinedOutput(); err != nil {
			t.Fatalf("go build: %v\n%s", err, out)
		}
		out, err := exec.Command(bin, "version").Output()
		if err != nil {
			t.Fatal(err)
		}
		if !regexp.MustCompile(`^vloop \S+ \(commit abc1234, plugin \S+, go\S+ \S+/\S+\)\n$`).Match(out) {
			t.Errorf("stamped version line: %q", out)
		}

		// No stamp and a module version through the seam: the module line.
		var o, e bytes.Buffer
		b := cli.Build{Version: "0.8.0", Commit: "unknown", Module: "v0.8.0"}
		if code := cli.Execute(b, []string{"version"}, &o, &e); code != 0 {
			t.Fatalf("version: %d %s", code, e.String())
		}
		want := regexp.MustCompile(`^vloop 0\.8\.0 \(module v0\.8\.0, plugin \S+, go\S+ ` + regexp.QuoteMeta(runtime.GOOS+"/"+runtime.GOARCH) + `\)\n$`)
		if !want.MatchString(o.String()) {
			t.Errorf("module version line: %q", o.String())
		}
		o.Reset()
		if code := cli.Execute(b, []string{"version", "--json"}, &o, &e); code != 0 {
			t.Fatalf("version --json: %d %s", code, e.String())
		}
		var v map[string]any
		if err := json.Unmarshal(o.Bytes(), &v); err != nil || v["module"] != "v0.8.0" {
			t.Errorf("version --json: %s (%v)", o.String(), err)
		}
	})
}
