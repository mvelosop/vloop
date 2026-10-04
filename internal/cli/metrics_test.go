package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mvelosop/vloop/internal/schema"
)

const metricsBrief = "B20260101-0900-a"

// collapse squeezes runs of spaces so tables compare independent of padding.
func collapse(s string) string {
	var lines []string
	for _, l := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		lines = append(lines, strings.Join(strings.Fields(l), " "))
	}
	return strings.Join(lines, "\n")
}

func nonBlankLines(n, blank int) string {
	return strings.Repeat("x\n", n) + strings.Repeat(" \n", blank)
}

// metricsRepo builds the brief's worked example: a plan, T1 done, T2 failing
// its gate then done, a run commit, and a squash merge onto main. It returns
// the repo and the merge commit's short SHA.
func metricsRepo(t *testing.T) (string, string) {
	dir := t.TempDir()
	at := func(date string, args ...string) { gitIn(t, dir, date, args...) }
	commit := func(date, msg string) {
		at(date, "add", "-A")
		at(date, "commit", "-q", "--allow-empty", "-m", msg)
	}
	at("2026-01-01T08:00:00Z", "init", "-q", "-b", "main")
	write(t, dir, "docs/briefs/"+metricsBrief+".loop-brief.md",
		"---\nname: "+metricsBrief+".loop-brief\nstatus: ready\n---\n## Shape\n\n2 to 3 tasks.\n")
	write(t, dir, ".vloop/config.toml", "[metrics]\nstacks = [\"go\"]\n")
	commit("2026-01-01T08:00:00Z", "init")
	at("2026-01-01T08:00:00Z", "checkout", "-q", "-b", metricsBrief)

	plan := func(a, b, status string) {
		write(t, dir, ".loop/state/state.json", fmt.Sprintf(
			`{"run_id":%q,"tasks":[{"id":"T1","status":%q},{"id":"T2","status":%q}]}`, metricsBrief, a, b))
	}
	plan("pending", "pending", "running")
	commit("2026-01-01T09:00:00Z", "[loop] plan "+metricsBrief)
	write(t, dir, "main.go", nonBlankLines(10, 2))
	write(t, dir, "main_test.go", nonBlankLines(6, 0))
	write(t, dir, "README.md", nonBlankLines(3, 0))
	plan("done", "pending", "running")
	commit("2026-01-01T09:02:00Z", "[loop] T1: done")
	write(t, dir, "util.go", "a\nb\nc\nd\n")
	commit("2026-01-01T09:04:00Z", "[loop] T2: gate_fail")
	write(t, dir, "util.go", "a\nb\nv1\nv2\nv3\n")
	plan("done", "done", "running")
	commit("2026-01-01T09:06:00Z", "[loop] T2: done")

	folder := ".loop/state/runs/" + metricsBrief + "/20260101-090000/"
	write(t, dir, folder+"loop.log", "\x1b[36m[loop]\x1b[0m planning from docs/briefs/"+metricsBrief+".loop-brief.md using opus\n")
	write(t, dir, folder+"iterations.jsonl",
		`{"iteration":1,"task":"T1","outcome":"done"}`+"\n"+`{"iteration":2,"task":"T2","outcome":"gate_fail"}`+"\n"+`{"iteration":3,"task":"T2","outcome":"done"}`+"\n")
	for _, s := range []struct {
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
		write(t, dir, folder+"sessions/"+s.file+".json", fmt.Sprintf(
			`{"duration_ms":%d,"total_cost_usd":%v,"phase":%q,"iteration":%d,"permission_denials":[],"modelUsage":{%q:{"inputTokens":10,"outputTokens":100,"cacheReadInputTokens":1000,"cacheCreationInputTokens":50,"costUSD":%v}}}`,
			s.ms, s.cost, s.phase, s.iter, s.model, s.cost))
	}
	write(t, dir, folder+"reports/001-verdict.json", `{"task":"T1","verdict":"PASS","findings":[]}`)
	write(t, dir, folder+"reports/003-verdict.json", `{"task":"T2","verdict":"PASS","findings":[]}`)
	plan("done", "done", "complete")
	commit("2026-01-01T09:10:00Z", "[loop] run "+metricsBrief+"/20260101-090000: complete")

	at("2026-01-01T09:20:00Z", "checkout", "-q", "main")
	at("2026-01-01T09:20:00Z", "merge", "-q", "--squash", metricsBrief)
	write(t, dir, "docs/briefs/"+metricsBrief+".loop-brief.md",
		"---\nname: "+metricsBrief+".loop-brief\nstatus: consumed\n---\n## Shape\n\n2 to 3 tasks.\n")
	at("2026-01-01T09:20:00Z", "add", "-A")
	at("2026-01-01T09:20:00Z", "commit", "-q", "-m", "vloop B-a (#1)", "-m", "Vloop-Brief: "+metricsBrief+".loop-brief")
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	return dir, strings.TrimSpace(string(out))
}

func TestMetricsCommandSummary(t *testing.T) {
	dir, sha := metricsRepo(t)
	want := strings.Join([]string{
		metricsBrief + " consumed · merged " + sha,
		"tasks 2 planned (brief said 2–3) · 2 done · 0 blocked · first-pass 1/2",
		"size delivered code 15 · test 6 · docs 3 · test:code 0.40",
		"churn code 17 · test 6 · docs 3 · rework 1.13",
		"time agent 2.7 min (work 2.2 · review 0.5) · plan 2.0 min · gates n/a · wall 10.0 min",
		"rate 5.6 code lines/min · 7.9 incl. tests",
		"cost $2.10 · plan 1.00 · work 0.90 · review 0.20 · $140.00 per 1,000 code lines",
		"models plan claude-opus-5-5 · work claude-sonnet-5-5 · review claude-sonnet-5-5",
		"defects in-loop 1 · operator 0 · escaped 0 · removal efficiency 100%",
	}, "\n")
	for _, arg := range []string{metricsBrief + ".loop-brief", "docs/briefs/" + metricsBrief + ".loop-brief.md"} {
		code, out, errs := runCLI(t, dir, "metrics", arg)
		if code != 0 || collapse(out) != want {
			t.Fatalf("%s: %d\n%s\n%s\nwant\n%s", arg, code, out, errs, want)
		}
	}
	// A recorded escaped defect changes the counts; a deleted record is reported.
	if code, _, e := runCLI(t, dir, "defect", "add", "util drops a line", "--found-by", "user", "--blame", "util.go:1"); code != 0 {
		t.Fatal(e)
	}
	_, out, _ := runCLI(t, dir, "metrics", metricsBrief+".loop-brief")
	if !strings.Contains(collapse(out), "defects in-loop 1 · operator 0 · escaped 1 · removal efficiency 50%") {
		t.Errorf("escaped defect: %s", out)
	}
	if err := os.Remove(filepath.Join(dir, ".loop/state/runs/"+metricsBrief+"/20260101-090000/sessions/006-review.json")); err != nil {
		t.Fatal(err)
	}
	_, out, _ = runCLI(t, dir, "metrics", metricsBrief+".loop-brief")
	c := collapse(out)
	if !strings.Contains(c, "\nrecords 1 session record(s) missing: T2 review (iteration 3) — their time and cost are not counted") ||
		!strings.Contains(c, "· plan 1.00 · work 0.90 · review 0.10 · ") {
		t.Errorf("missing record: %s", out)
	}
}

func TestMetricsCommandByTask(t *testing.T) {
	dir, _ := metricsRepo(t)
	code, out, errs := runCLI(t, dir, "metrics", metricsBrief+".loop-brief", "--by", "task")
	want := "id area kind att code+ test+ docs+ other+ agent cost model\n" +
		"T1 - - 1 10 6 3 0 1m20s $0.50 claude-sonnet-5-5\n" +
		"T2 - - 2 7 0 0 0 1m20s $0.60 claude-sonnet-5-5"
	if code != 0 || collapse(out) != want {
		t.Fatalf("%d\n%s\n%s", code, out, errs)
	}
	if code, _, _ := runCLI(t, dir, "metrics", metricsBrief+".loop-brief", "--by", "area"); code != 2 {
		t.Errorf("--by area: exit %d, want 2", code)
	}
}

func TestMetricsCommandCrossBrief(t *testing.T) {
	dir, _ := metricsRepo(t)
	write(t, dir, "docs/briefs/B20260103-0900-c.loop-brief.md", "---\nname: B20260103-0900-c.loop-brief\nstatus: ready\n---\n")
	code, out, errs := runCLI(t, dir, "metrics")
	want := "brief tasks first-pass code test t:c agent $/1k in-loop operator escaped efficiency\n" +
		metricsBrief + " 2 1/2 15 6 0.40 2.7 140.00 1 0 0 100%"
	if code != 0 || collapse(out) != want {
		t.Fatalf("%d\n%s\n%s", code, out, errs)
	}
}

func TestMetricsCommandNoRuns(t *testing.T) {
	dir, _ := metricsRepo(t)
	log := filepath.Join(dir, ".loop/state/runs/"+metricsBrief+"/20260101-090000/loop.log")
	if err := os.WriteFile(log, []byte("planning from docs/briefs/B20260101-0900-z.loop-brief.md\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, errs := runCLI(t, dir, "metrics", metricsBrief+".loop-brief")
	if code != 1 || out != "" || errs != "vloop: no runs for "+metricsBrief+".loop-brief\n" {
		t.Errorf("%d %q %q", code, out, errs)
	}
}

func TestMetricsCommandJSON(t *testing.T) {
	dir, sha := metricsRepo(t)
	code, out, errs := runCLI(t, dir, "--json", "metrics", metricsBrief+".loop-brief")
	if code != 0 {
		t.Fatalf("%d %s", code, errs)
	}
	if v, err := schema.Validate("metrics/v2", []byte(out)); err != nil || len(v) > 0 {
		t.Fatalf("does not validate: %v %v\n%s", v, err, out)
	}
	var doc struct {
		RunID  string `json:"run_id"`
		Merged string `json:"merged"`
		Time   struct {
			AgentMS int64  `json:"agent_ms"`
			GatesMS *int64 `json:"gates_ms"`
			WallMS  int64  `json:"wall_ms"`
		} `json:"time"`
		Cost struct {
			Total float64 `json:"total_usd"`
		} `json:"cost"`
		IterationsPerClosed float64 `json:"iterations_per_closed"`
	}
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.RunID != metricsBrief || !strings.HasPrefix(doc.Merged, sha) || doc.Time.AgentMS != 160000 ||
		doc.Time.GatesMS != nil || doc.Time.WallMS != 600000 || doc.IterationsPerClosed != 1.5 ||
		doc.Cost.Total < 2.0999 || doc.Cost.Total > 2.1001 {
		t.Errorf("document: %+v", doc)
	}
	// Several briefs print an array.
	_, out, _ = runCLI(t, dir, "--json", "metrics")
	if !strings.HasPrefix(out, "[{") {
		t.Errorf("cross-brief --json: %q", out)
	}
}
