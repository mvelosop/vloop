package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

const statusPlan = `{
  "schema": "state/v2",
  "run_id": "B20260101-0900-a",
  "brief": "docs/briefs/b.loop-brief.md",
  "base": "0123456789abcdef0123456789abcdef01234567",
  "branch": "B20260101-0900-a",
  "status": "running",
  "iteration": 2,
  "created": "2026-01-01T09:00:00Z",
  "updated": "2026-01-01T09:00:00Z",
  "shell": "sh",
  "checks": [],
  "gate_scratch": [],
  "gate_review": {"rounds": 0, "verdict": ""},
  "tasks": [
    {"id":"T1","title":"Skeleton","goal":"g","kind":"feature","area":"cli","fixtures":"","references":[],"depends_on":[],"acceptance":["a"],"verify":"true","status":"done","attempts":0,"notes":""},
    {"id":"T2","title":"Config","goal":"Add the config.","kind":"feature","area":"config","fixtures":"","references":[],"depends_on":["T1"],"acceptance":["b exists"],"verify":"true","status":"pending","attempts":1,"notes":"watch the path"},
    {"id":"T3","title":"README","goal":"g","kind":"docs","fixtures":"","references":[],"depends_on":["T2"],"acceptance":["c"],"verify":"true","status":"blocked","attempts":2,"notes":""}
  ]
}`

func statusRepo(t *testing.T, plan string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".vloop", "state"), 0o755); err != nil {
		t.Fatal(err)
	}
	if plan != "" {
		if err := os.WriteFile(filepath.Join(root, ".vloop", "state", "state.json"), []byte(plan), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestStatusText(t *testing.T) {
	root := statusRepo(t, statusPlan)
	want := "B20260101-0900-a — running · 1/3 done · 1 blocked · iteration 2\n" +
		"  T1  done     cli/feature     Skeleton\n" +
		"  T2  pending  config/feature  Config  (1 attempt)\n" +
		"  T3  blocked  -/docs          README  (2 attempts)\n"
	code, out, errOut := run(t, "-C", root, "status")
	if code != 0 || out != want || errOut != "" {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
}

func TestStatusFromSubdirectory(t *testing.T) {
	root := statusRepo(t, statusPlan)
	sub := filepath.Join(root, "docs", "deep")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	code, out, _ := run(t, "-C", sub, "status")
	if code != 0 || !strings.Contains(out, "  T2  pending") {
		t.Fatalf("code %d out %q", code, out)
	}
}

func TestStatusJSON(t *testing.T) {
	root := statusRepo(t, statusPlan)
	code, out, _ := run(t, "-C", root, "status", "--json")
	want := `{"run_id":"B20260101-0900-a","status":"running","iteration":2,"done":1,"total":3,"blocked":1,"tasks":[` +
		`{"id":"T1","status":"done","area":"cli","kind":"feature","title":"Skeleton","attempts":0},` +
		`{"id":"T2","status":"pending","area":"config","kind":"feature","title":"Config","attempts":1},` +
		`{"id":"T3","status":"blocked","area":null,"kind":"docs","title":"README","attempts":2}]}` + "\n"
	if code != 0 || out != want {
		t.Fatalf("code %d out %q", code, out)
	}
}

func TestStatusMarkdown(t *testing.T) {
	root := statusRepo(t, statusPlan)
	code, out, errOut := run(t, "-C", root, "status", "--markdown")
	if code != 0 || errOut != "" {
		t.Fatalf("code %d err %q", code, errOut)
	}
	for _, want := range []string{
		"# Plan — B20260101-0900-a\n",
		"- [x] **T1** — Skeleton\n",
		"- [ ] **T2** — Config · 1 attempt(s)\n",
		"- [ ] **T3** — README · **blocked**\n",
		"### T2 — Config\n",
		"`pending` · 1 attempt(s) · depends on: T1\n",
		"`blocked` · **blocked** · depends on: T2\n",
		"- b exists\n",
		"**From the last attempt:** watch the path\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("markdown lacks %q", want)
		}
	}
	ents, _ := os.ReadDir(filepath.Join(root, ".vloop", "state"))
	if len(ents) != 1 {
		t.Errorf("status wrote files: %v", ents)
	}
}

func TestStatusMissingPlan(t *testing.T) {
	root := statusRepo(t, "")
	code, out, errOut := run(t, "-C", root, "status")
	if code != 1 || out != "" || errOut != "vloop: no plan — vloop run <brief> makes one\n" {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
	code, out, _ = run(t, "-C", root, "status", "--json")
	var v map[string]any
	if code != 1 || json.Unmarshal([]byte(out), &v) != nil {
		t.Fatalf("--json: code %d out %q", code, out)
	}
}

func TestStatusSchemaInvalidPlanRefuses(t *testing.T) {
	root := statusRepo(t, strings.Replace(statusPlan, `"notes":"watch the path"`, `"notes":"watch the path","effort":{"work":"turbo"}`, 1))
	code, out, errOut := run(t, "-C", root, "status")
	if code != 1 || out != "" || errOut != "vloop: .vloop/state/state.json is not a valid plan — vloop task validate lists the problems\n" {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
}

func TestStatusJSONAndMarkdownConflict(t *testing.T) {
	root := statusRepo(t, statusPlan)
	if code, _, _ := run(t, "-C", root, "status", "--json", "--markdown"); code != 2 {
		t.Fatalf("code %d", code)
	}
}

// plantLock writes a run lock held by pid, the way the driver records a
// session in flight.
func plantLock(t *testing.T, root string, pid int, fields string) {
	t.Helper()
	dir := filepath.Join(root, ".vloop", "tmp")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	rec := `{"pid":` + strconv.Itoa(pid) + `,"branch":"b","started":"2026-01-01T09:00:00Z","run":"B20260101-0900-a",` + fields + `}`
	if err := os.WriteFile(filepath.Join(dir, ".running"), []byte(rec+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestStatusRunningLine(t *testing.T) {
	root := statusRepo(t, statusPlan)
	started := time.Now().Add(-90 * time.Second).UTC().Format(time.RFC3339)
	plantLock(t, root, os.Getpid(), `"phase":"work","task":"T2","session_started":"`+started+`","spent":1.5`)
	code, out, _ := run(t, "-C", root, "status")
	lines := strings.Split(out, "\n")
	if code != 0 || !strings.HasPrefix(lines[0], "B20260101-0900-a — ") ||
		!regexp.MustCompile(`^running: work T2 for 1m(29|30|31|32)s · \$1\.50 of \$\d+\.\d\d$`).MatchString(lines[1]) {
		t.Fatalf("code %d out %q", code, out)
	}
	code, out, _ = run(t, "-C", root, "status", "--json")
	var got struct {
		Running struct {
			Phase          string  `json:"phase"`
			Task           string  `json:"task"`
			SessionStarted string  `json:"session_started"`
			Spent          float64 `json:"spent"`
			Ceiling        float64 `json:"ceiling"`
		} `json:"running"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil || code != 0 ||
		got.Running.Phase != "work" || got.Running.Task != "T2" || got.Running.SessionStarted != started ||
		got.Running.Spent != 1.5 || got.Running.Ceiling <= 0 {
		t.Fatalf("code %d out %q err %v", code, out, err)
	}
}

func TestStatusRunningWhilePlanning(t *testing.T) {
	root := statusRepo(t, "")
	started := time.Now().UTC().Format(time.RFC3339)
	plantLock(t, root, os.Getpid(), `"phase":"plan","session_started":"`+started+`","spent":0`)
	code, out, errOut := run(t, "-C", root, "status")
	if code != 0 || errOut != "" || !regexp.MustCompile(`^running: plan for 0m0[0-2]s · \$0\.00 of \$\d+\.\d\d\n$`).MatchString(out) {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
	code, out, _ = run(t, "-C", root, "status", "--json")
	var got map[string]json.RawMessage
	if err := json.Unmarshal([]byte(out), &got); err != nil || code != 0 || got["running"] == nil {
		t.Fatalf("code %d out %q", code, out)
	}
}

func TestStatusNoRunningWithoutLiveRun(t *testing.T) {
	root := statusRepo(t, statusPlan)
	// No lock at all, then a lock whose process is gone.
	for _, pid := range []int{0, 2147483646} {
		if pid != 0 {
			plantLock(t, root, pid, `"phase":"work","task":"T2","session_started":"2026-01-01T09:00:00Z","spent":1`)
		}
		_, out, _ := run(t, "-C", root, "status")
		if strings.Contains(out, "\nrunning:") || strings.HasPrefix(out, "running:") {
			t.Fatalf("pid %d: running line without a live run: %q", pid, out)
		}
		_, out, _ = run(t, "-C", root, "status", "--json")
		if strings.Contains(out, `"running":`) {
			t.Fatalf("pid %d: running key without a live run: %q", pid, out)
		}
	}
	// A stale lock and no plan is still the missing-plan problem.
	empty := statusRepo(t, "")
	plantLock(t, empty, 2147483646, `"phase":"plan","session_started":"2026-01-01T09:00:00Z","spent":0`)
	if code, _, _ := run(t, "-C", empty, "status"); code != 1 {
		t.Fatalf("stale lock, no plan: code %d, want 1", code)
	}
}
