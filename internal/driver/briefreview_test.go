package driver

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const (
	reviewBriefPath = "docs/briefs/B20990101-0000-x.loop-brief.md"
	reviewFixReport = `{"schema":"brief-review/v1","brief":"B20990101-0000-x.loop-brief","verdict":"FIX","findings":[` +
		`{"kind":"runnability","section":"Shape","summary":"the go check is red after task 1"}],"decisions":[]}`
	reviewPassReport = `{"schema":"brief-review/v1","brief":"B20990101-0000-x.loop-brief","verdict":"PASS","findings":[]}`
)

// reviewRepo is a git repository with the brief committed and a stub claude
// that runs the shell body given, then prints a claude result document.
func reviewRepo(t *testing.T, body string) (*BriefReview, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the stub claude is a shell script")
	}
	root := t.TempDir()
	root, _ = filepath.EvalSymlinks(root)
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	git := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q", "-b", "main")
	git("config", "user.name", "t")
	git("config", "user.email", "t@example.com")
	if err := os.MkdirAll(filepath.Join(root, "docs", "briefs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".vloop", "tmp"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, reviewBriefPath), []byte("# brief\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "base")
	stub := filepath.Join(t.TempDir(), "claude")
	script := "#!/bin/sh\necho \"$@\" >" + filepath.Join(filepath.Dir(stub), "argv") + "\n" + body + "\n" +
		`echo '{"type":"result","is_error":false,"total_cost_usd":0.1,"num_turns":2,"duration_ms":10,"permission_denials":[]}'` + "\n" +
		"exit ${STUB_EXIT:-0}\n"
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	return &BriefReview{Root: root, Version: "0.0.0-test", Brief: reviewBriefPath, Out: &out, Err: &errb,
		Timeout: time.Minute, Claude: stub, Home: t.TempDir(), User: "tester",
		Now: func() time.Time { return time.Date(2099, 1, 2, 3, 4, 5, 0, time.UTC) }}, &out, &errb
}

func writeReport(report string) string {
	return "printf '%s\\n' '" + report + "' >.vloop/tmp/brief-review.json"
}

func TestBriefReviewFixKeepsReportAndRecord(t *testing.T) {
	b, out, errb := reviewRepo(t, writeReport(reviewFixReport))
	verdict, err := b.Run()
	if err != nil || verdict != "FIX" {
		t.Fatalf("verdict %q err %v", verdict, err)
	}
	folder := ".vloop/state/brief-reviews/B20990101-0000-x/20990102-030405"
	want := "verdict: FIX — 1 finding(s)\n  runnability   Shape: the go check is red after task 1\nkept: " + folder + "/brief-review.json\n"
	if out.String() != want {
		t.Errorf("stdout:\n%q\nwant\n%q", out.String(), want)
	}
	if !strings.Contains(errb.String(), "reviewing "+reviewBriefPath+" using opus (agent claude)\n") {
		t.Errorf("stderr %q", errb.String())
	}
	if _, err := os.ReadFile(filepath.Join(b.Root, folder, "brief-review.json")); err != nil {
		t.Errorf("report not kept: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(b.Root, folder, "sessions", "001-review-brief.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rec map[string]any
	if err := json.Unmarshal(raw, &rec); err != nil {
		t.Fatal(err)
	}
	if rec["schema"] != "session/v5" || rec["phase"] != "review-brief" || rec["run_id"] != "B20990101-0000-x" || rec["iteration"] != float64(0) {
		t.Errorf("record %v", rec)
	}
	argv, _ := os.ReadFile(filepath.Join(filepath.Dir(b.Claude), "argv"))
	if !strings.Contains(string(argv), "-p /vloop:review-brief "+reviewBriefPath+" --model opus") || !strings.Contains(string(argv), "fence/0.0.0-test/review-brief.json") {
		t.Errorf("argv %q", argv)
	}
	if _, err := os.Stat(filepath.Join(b.Root, ".vloop", "tmp", ".running")); err == nil {
		t.Error("the lock was not released")
	}
}

func TestBriefReviewPass(t *testing.T) {
	b, out, _ := reviewRepo(t, writeReport(reviewPassReport))
	verdict, err := b.Run()
	if err != nil || verdict != "PASS" || !strings.HasPrefix(out.String(), "verdict: PASS — 0 finding(s)\nkept: ") {
		t.Fatalf("verdict %q err %v out %q", verdict, err, out.String())
	}
}

func TestBriefReviewPlantedFailures(t *testing.T) {
	cases := []struct {
		name, body, env string
		code            int
		line            string
	}{
		{"no report", ":", "", ExitSessionError, "the brief review left no .vloop/tmp/brief-review.json — see "},
		{"verdict OK", writeReport(strings.Replace(reviewFixReport, `"FIX"`, `"OK"`, 1)), "", ExitSessionError, "does not follow brief-review/v1 — see "},
		{"FIX without findings", writeReport(strings.Replace(reviewFixReport, `"findings":[`+
			`{"kind":"runnability","section":"Shape","summary":"the go check is red after task 1"}]`, `"findings":[]`, 1)), "", ExitSessionError, "does not follow brief-review/v1 — see "},
		{"another brief", writeReport(strings.Replace(reviewFixReport, "B20990101-0000-x", "B20990101-0001-y", 1)), "", ExitSessionError, "does not follow brief-review/v1 — see "},
		{"exit 3", writeReport(reviewFixReport), "3", ExitSessionError, "brief review session failed (claude exited 3) — see "},
		{"git config", writeReport(reviewFixReport) + "\ngit config gate.planted yes", "", ExitRefsMoved, "the brief review changed .git/config (gate.planted) — nothing was committed; restore it, then re-run"},
		{"moved HEAD", writeReport(reviewFixReport) + "\ngit commit -q --allow-empty -m moved", "", ExitRefsMoved, "REFS MOVED review-brief — the brief review changed git refs; nothing was committed"},
		{"changed brief", writeReport(reviewFixReport) + "\necho more >>" + reviewBriefPath, "", ExitRefsMoved, "the brief review changed " + reviewBriefPath + "; it was restored"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.env != "" {
				t.Setenv("STUB_EXIT", c.env)
			}
			b, out, _ := reviewRepo(t, c.body)
			_, err := b.Run()
			h, ok := err.(*Halt)
			if !ok || h.Code != c.code || !strings.Contains(h.Error(), c.line) {
				t.Fatalf("err %v, want code %d containing %q", err, c.code, c.line)
			}
			if out.Len() != 0 {
				t.Errorf("stdout %q", out.String())
			}
			if got, _ := os.ReadFile(filepath.Join(b.Root, reviewBriefPath)); string(got) != "# brief\n" {
				t.Errorf("brief bytes %q", got)
			}
			kept, _ := filepath.Glob(filepath.Join(b.Root, ".vloop/state/brief-reviews/*/*/brief-review.json"))
			if len(kept) != 0 {
				t.Errorf("a report was kept: %v", kept)
			}
		})
	}
}

func TestBriefReviewTimeout(t *testing.T) {
	b, _, _ := reviewRepo(t, "sleep 30")
	b.Timeout = 300 * time.Millisecond
	_, err := b.Run()
	h, ok := err.(*Halt)
	if !ok || h.Code != ExitSessionError || !strings.Contains(h.Error(), "brief review session timed out after 300ms — see .vloop/state/brief-reviews/B20990101-0000-x/") {
		t.Fatalf("err %v", err)
	}
}

func TestBriefReviewRefusesALiveLoop(t *testing.T) {
	b, _, _ := reviewRepo(t, writeReport(reviewPassReport))
	release, err := AcquireLock(b.Root, "main", "other", time.Now(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	_, err = b.Run()
	h, ok := err.(*Halt)
	if !ok || h.Code != ExitPreflight || !strings.Contains(h.Error(), "a loop is already running in this working tree (pid ") {
		t.Fatalf("err %v", err)
	}
	if _, err := os.Stat(filepath.Join(b.Root, ".vloop", "state", "brief-reviews")); err == nil {
		t.Error("a folder was made")
	}
}
