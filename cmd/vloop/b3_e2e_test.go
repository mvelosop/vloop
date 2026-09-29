package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

const b3Brief = "B20260101-0900-a"

// b3Opts plants one failure in the worked example's fixture repository.
type b3Opts struct {
	noTrailer   bool   // squash commit without the Vloop-Brief trailer
	gateHistory bool   // T2 gets an operator gate_history entry before its done commit
	logNames    string // the brief loop.log names, when not the brief
}

func b3Git(t *testing.T, dir, date string, args ...string) string {
	t.Helper()
	c := exec.Command("git", append([]string{"-C", dir}, args...)...)
	c.Env = append(cleanEnv(t.TempDir()), "GIT_AUTHOR_DATE="+date, "GIT_COMMITTER_DATE="+date,
		"GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func b3Lines(n, blank int) string {
	return strings.Repeat("x\n", n) + strings.Repeat(" \n", blank)
}

// b3Fixture builds the brief's worked example and returns the scratch and the
// squash commit's short SHA.
func b3Fixture(t *testing.T, o b3Opts) (*scratch, string) {
	t.Helper()
	s := &scratch{t: t, dir: t.TempDir(), home: t.TempDir()}
	at := func(date string, args ...string) string { return b3Git(t, s.dir, date, args...) }
	commit := func(date, msg string, extra ...string) {
		at(date, "add", "-A")
		args := []string{"commit", "-q", "--allow-empty", "-m", msg}
		for _, e := range extra {
			args = append(args, "-m", e)
		}
		at(date, args...)
	}
	briefFile := "docs/briefs/" + b3Brief + ".loop-brief.md"
	brief := func(status string) string {
		return "---\nname: " + b3Brief + ".loop-brief\nstatus: " + status + "\n---\n## Shape\n\n2 to 3 tasks.\n"
	}
	at("2026-01-01T08:00:00Z", "init", "-q", "-b", "main")
	s.write(briefFile, brief("ready"))
	s.write(".vloop/config.toml", "[metrics]\nstacks = [\"go\"]\n")
	commit("2026-01-01T08:00:00Z", "init")
	at("2026-01-01T08:00:00Z", "checkout", "-q", "-b", b3Brief)

	plan := func(a, b, status string, history bool) {
		t2 := fmt.Sprintf(`{"id":"T2","status":%q`, b)
		if history {
			t2 += `,"gate_history":[{"verify":"old","reason":"the gate was wrong","by":"operator","replaced_at":"2026-01-01T09:05:00Z"}]`
		}
		s.write(".loop/state/state.json", fmt.Sprintf(
			`{"run_id":%q,"status":%q,"tasks":[{"id":"T1","status":%q},%s}]}`, b3Brief, status, a, t2))
	}
	plan("pending", "pending", "running", false)
	commit("2026-01-01T09:00:00Z", "[loop] plan "+b3Brief)
	s.write("main.go", b3Lines(10, 2))
	s.write("main_test.go", b3Lines(6, 0))
	s.write("README.md", b3Lines(3, 0))
	plan("done", "pending", "running", false)
	commit("2026-01-01T09:02:00Z", "[loop] T1: done")
	s.write("util.go", "a\nb\nc\nd\n")
	commit("2026-01-01T09:04:00Z", "[loop] T2: gate_fail")
	s.write("util.go", "a\nb\nv1\nv2\nv3\n")
	plan("done", "done", "running", o.gateHistory)
	commit("2026-01-01T09:06:00Z", "[loop] T2: done")

	folder := ".loop/state/runs/" + b3Brief + "/20260101-090000/"
	named := b3Brief
	if o.logNames != "" {
		named = o.logNames
	}
	s.write(folder+"loop.log", "\x1b[36m[loop]\x1b[0m planning from docs/briefs/"+named+".loop-brief.md using opus\n")
	s.write(folder+"iterations.jsonl",
		`{"iteration":1,"task":"T1","outcome":"done"}`+"\n"+`{"iteration":2,"task":"T2","outcome":"gate_fail"}`+"\n"+`{"iteration":3,"task":"T2","outcome":"done"}`+"\n")
	for _, x := range []struct {
		file, phase, model string
		iter, ms           int
		cost               float64
	}{
		{"001-plan", "plan", "claude-opus-5-5", 0, 120000, 1.00},
		{"002-work", "work", "claude-sonnet-5-5", 1, 60000, 0.40},
		{"003-review", "review", "claude-sonnet-5-5", 1, 20000, 0.10},
		{"004-work", "work", "claude-sonnet-5-5", 2, 40000, 0.30},
		{"005-work", "work", "claude-sonnet-5-5", 3, 30000, 0.20},
		{"006-review", "review", "claude-sonnet-5-5", 3, 10000, 0.10},
	} {
		s.write(folder+"sessions/"+x.file+".json", fmt.Sprintf(
			`{"duration_ms":%d,"total_cost_usd":%v,"phase":%q,"iteration":%d,"permission_denials":[],"modelUsage":{%q:{"inputTokens":10,"outputTokens":100,"cacheReadInputTokens":1000,"cacheCreationInputTokens":50,"costUSD":%v}}}`,
			x.ms, x.cost, x.phase, x.iter, x.model, x.cost))
	}
	s.write(folder+"reports/001-verdict.json", `{"task":"T1","verdict":"PASS","findings":[]}`)
	s.write(folder+"reports/003-verdict.json", `{"task":"T2","verdict":"PASS","findings":[]}`)
	plan("done", "done", "complete", o.gateHistory)
	commit("2026-01-01T09:10:00Z", "[loop] run "+b3Brief+"/20260101-090000: complete")

	at("2026-01-01T09:20:00Z", "checkout", "-q", "main")
	at("2026-01-01T09:20:00Z", "merge", "-q", "--squash", b3Brief)
	s.write(briefFile, brief("consumed"))
	at("2026-01-01T09:20:00Z", "add", "-A")
	if o.noTrailer {
		at("2026-01-01T09:20:00Z", "commit", "-q", "-m", "squash of "+b3Brief)
	} else {
		at("2026-01-01T09:20:00Z", "commit", "-q", "-m", "squash of "+b3Brief, "-m", "Vloop-Brief: "+b3Brief+".loop-brief")
	}
	sha := at("2026-01-01T09:20:00Z", "rev-parse", "--short", "HEAD")
	// A later commit that is neither trailered nor a consumption.
	s.write("notes.txt", "note\n")
	commit("2026-01-01T09:30:00Z", "notes")
	return s, sha
}

func b3Collapse(s string) string {
	var lines []string
	for _, l := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		lines = append(lines, strings.Join(strings.Fields(l), " "))
	}
	return strings.Join(lines, "\n")
}

func b3Expect(t *testing.T, r result, code int, want string) {
	t.Helper()
	if r.code != code || b3Collapse(r.out) != want || r.err != "" {
		t.Fatalf("code=%d err=%q out:\n%s\nwant code=%d, out:\n%s", r.code, r.err, r.out, code, want)
	}
}

func b3Summary(sha string) []string {
	return []string{
		b3Brief + " consumed · merged " + sha,
		"tasks 2 planned (brief said 2–3) · 2 done · 0 blocked · first-pass 1/2",
		"size delivered code 15 · test 6 · docs 3 · test:code 0.40",
		"churn code 17 · test 6 · docs 3 · rework 1.13",
		"time agent 2.7 min (work 2.2 · review 0.5) · plan 2.0 min · gates n/a · wall 10.0 min",
		"rate 5.6 code lines/min · 7.9 incl. tests",
		"cost $2.10 · plan 1.00 · work 0.90 · review 0.20 · $140.00 per 1,000 code lines",
		"models plan claude-opus-5-5 · work claude-sonnet-5-5 · review claude-sonnet-5-5",
		"defects in-loop 1 · operator 0 · escaped 0 · removal efficiency 100%",
	}
}

func TestWorkedExampleB3Commands(t *testing.T) {
	s, sha := b3Fixture(t, b3Opts{})
	brief := b3Brief + ".loop-brief"

	b3Expect(t, s.run(nil, "metrics", brief), 0, strings.Join(b3Summary(sha), "\n"))

	b3Expect(t, s.run(nil, "metrics", brief, "--by", "task"), 0,
		"id area kind att code+ test+ docs+ other+ agent cost model\n"+
			"T1 - - 1 10 6 3 0 1m20s $0.50 claude-sonnet-5-5\n"+
			"T2 - - 2 7 0 0 0 1m20s $0.60 claude-sonnet-5-5")

	b3Expect(t, s.run(nil, "metrics", "classify", "main.go", "main_test.go", "README.md", "go.sum", ".loop/x", "notes.txt"), 0,
		"main.go code (go: **/*.go)\n"+
			"main_test.go test (go: **/*_test.go)\n"+
			"README.md docs (go: **/*.md)\n"+
			"go.sum excluded (go: go.sum)\n"+
			".loop/x excluded (always: .loop/**)\n"+
			"notes.txt other (none)")

	r := s.run(nil, "defect", "add", "util drops a line", "--found-by", "user", "--blame", "util.go:1")
	if r.code != 0 || !regexp.MustCompile(`^\.vloop/defects/D\d{8}-\d{4}-util-drops-a-line\.md\n$`).MatchString(r.out) {
		t.Fatalf("defect add: %+v", r)
	}
	if want := "attributed to " + brief + " (trailer on " + sha + ")\n"; r.err != want {
		t.Fatalf("defect add stderr = %q, want %q", r.err, want)
	}

	r = s.run(nil, "metrics", brief)
	want := " defects in-loop 1 · operator 0 · escaped 1 · removal efficiency 50%"
	if r.code != 0 || !hasLine(b3Collapse(r.out), strings.TrimSpace(want)) {
		t.Fatalf("metrics after defect add: %+v", r)
	}

	b3Expect(t, s.run(nil, "defect", "list", "--matrix"), 0,
		"gate review operator user\nbrief 0 0 0 0\nplan 0 0 0 0\nwork 1 0 0 1\nenv 0 0 0 0")
}

func TestWorkedExampleB3PlantedFailures(t *testing.T) {
	brief := b3Brief + ".loop-brief"

	t.Run("trailer removed", func(t *testing.T) {
		s, sha := b3Fixture(t, b3Opts{noTrailer: true})
		r := s.run(nil, "defect", "add", "util drops a line", "--found-by", "user", "--blame", "util.go:1")
		if want := "attributed to " + brief + " (consumed in " + sha + ")\n"; r.code != 0 || r.err != want {
			t.Fatalf("got %+v, want stderr %q", r, want)
		}
	})
	t.Run("unattributable blame", func(t *testing.T) {
		s, _ := b3Fixture(t, b3Opts{})
		before := s.run(nil, "defect", "list")
		r := s.run(nil, "defect", "add", "x", "--found-by", "user", "--blame", "notes.txt:1")
		if r.code != 1 || r.out != "" || r.err != "vloop: cannot attribute notes.txt:1 to a loop brief — pass --brief\n" {
			t.Fatalf("got %+v", r)
		}
		if after := s.run(nil, "defect", "list"); after != before {
			t.Fatalf("nothing should be written: %+v -> %+v", before, after)
		}
	})
	t.Run("missing review record", func(t *testing.T) {
		s, sha := b3Fixture(t, b3Opts{})
		if err := os.Remove(s.dir + "/.loop/state/runs/" + b3Brief + "/20260101-090000/sessions/006-review.json"); err != nil {
			t.Fatal(err)
		}
		r := s.run(nil, "metrics", brief)
		c := b3Collapse(r.out)
		if r.code != 0 || r.err != "" ||
			!hasLine(c, "records 1 session record(s) missing: T2 review (iteration 3) — their time and cost are not counted") ||
			!strings.Contains(c, "· plan 1.00 · work 0.90 · review 0.10 · ") || !strings.Contains(c, "cost $2.00 · ") ||
			!strings.HasPrefix(c, b3Brief+" consumed · merged "+sha+"\n") {
			t.Fatalf("got %+v", r)
		}
	})
	t.Run("loop.log names another brief", func(t *testing.T) {
		s, _ := b3Fixture(t, b3Opts{logNames: "B20260101-1000-other"})
		r := s.run(nil, "metrics", brief)
		if r.code != 1 || r.out != "" || r.err != "vloop: no runs for "+brief+"\n" {
			t.Fatalf("got %+v", r)
		}
	})
	t.Run("gate_history reclassifies the gate failure", func(t *testing.T) {
		s, _ := b3Fixture(t, b3Opts{gateHistory: true})
		b3Expect(t, s.run(nil, "defect", "list", "--matrix"), 0,
			"gate review operator user\nbrief 0 0 0 0\nplan 1 0 0 0\nwork 0 0 0 0\nenv 0 0 0 0")
	})
	t.Run("unknown stack", func(t *testing.T) {
		s, _ := b3Fixture(t, b3Opts{})
		expect(t, s.run(nil, "metrics", "stacks", "cobol"), 2, "", "vloop: unknown stack \"cobol\"\n")
	})
}
