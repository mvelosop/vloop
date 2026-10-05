package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// taskRepo is a scratch repo with statusPlan, no VLOOP_* overrides, and a
// config file with the given body (none when empty).
func taskRepo(t *testing.T, cfg string) string {
	t.Helper()
	for _, v := range []string{"MODEL_WORK", "MODEL_REVIEW", "EFFORT_WORK", "EFFORT_REVIEW"} {
		t.Setenv("VLOOP_"+v, "")
	}
	root := statusRepo(t, statusPlan)
	if cfg != "" {
		if err := os.WriteFile(filepath.Join(root, ".vloop", "config.toml"), []byte(cfg), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func setTask(t *testing.T, root, id, field, value string) {
	t.Helper()
	p := filepath.Join(root, ".vloop", "state", "state.json")
	b, _ := os.ReadFile(p)
	old := `{"id":"` + id + `",`
	if !strings.Contains(string(b), old) {
		t.Fatal("no task " + id)
	}
	b = []byte(strings.Replace(string(b), old, old+field+":"+value+",", 1))
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestTaskListText(t *testing.T) {
	root := taskRepo(t, "")
	want := "  T1  done     cli/feature     Skeleton\n" +
		"  T2  pending  config/feature  Config  (1 attempt)\n" +
		"  T3  blocked  -/docs          README  (2 attempts)\n"
	code, out, errOut := run(t, "-C", root, "task", "list")
	if code != 0 || out != want || errOut != "" {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
}

func TestTaskListJSONMatchesStatus(t *testing.T) {
	root := taskRepo(t, "")
	code, out, _ := run(t, "-C", root, "task", "list", "--json")
	_, st, _ := run(t, "-C", root, "status", "--json")
	var s struct {
		Tasks json.RawMessage `json:"tasks"`
	}
	if err := json.Unmarshal([]byte(st), &s); err != nil {
		t.Fatal(err)
	}
	if code != 0 || strings.TrimSpace(out) != string(s.Tasks) {
		t.Fatalf("task list --json = %q, status tasks = %s", out, s.Tasks)
	}
}

func TestTaskListNoPlan(t *testing.T) {
	root := statusRepo(t, "")
	code, out, errOut := run(t, "-C", root, "task", "list")
	if code != 1 || out != "" || errOut != "vloop: no plan — vloop run <brief> makes one\n" {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
}

func TestTaskShowTextSources(t *testing.T) {
	root := taskRepo(t, "[effort]\nreview = \"high\"\n")
	setTask(t, root, "T2", `"model"`, `{"work":"opus"}`)
	code, out, errOut := run(t, "-C", root, "task", "show", "T2")
	if code != 0 || errOut != "" {
		t.Fatalf("code %d err %q", code, errOut)
	}
	for _, l := range []string{"T2  Config", "goal    Add the config.", "notes   watch the path",
		"model   work opus (task) · review sonnet (default)",
		"effort  work - (default) · review high (file)"} {
		if !strings.Contains(out, l+"\n") {
			t.Errorf("missing %q in %q", l, out)
		}
	}
}

func TestTaskShowEnvAndFileAndTask(t *testing.T) {
	root := taskRepo(t, "[model]\nwork = \"fable\"\n")
	t.Setenv("VLOOP_MODEL_REVIEW", "haiku")
	t.Setenv("VLOOP_EFFORT_REVIEW", "low")
	setTask(t, root, "T3", `"effort"`, `{"review":"max"}`)
	_, out, _ := run(t, "-C", root, "task", "show", "T3")
	for _, l := range []string{"model   work fable (file) · review haiku (env)",
		"effort  work - (default) · review max (task)"} {
		if !strings.Contains(out, l+"\n") {
			t.Errorf("missing %q in %q", l, out)
		}
	}
	_, out, _ = run(t, "-C", root, "task", "show", "T2")
	if !strings.Contains(out, "effort  work - (default) · review low (env)\n") {
		t.Errorf("env effort: %q", out)
	}
}

func TestTaskShowJSON(t *testing.T) {
	root := taskRepo(t, "[effort]\nreview = \"high\"\n")
	setTask(t, root, "T2", `"model"`, `{"work":"opus"}`)
	code, out, _ := run(t, "-C", root, "task", "show", "T2", "--json")
	if code != 0 {
		t.Fatalf("code %d", code)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	for k, want := range map[string]any{"id": "T2", "title": "Config", "kind": "feature", "area": "config",
		"verify": "true", "status": "pending", "attempts": float64(1), "notes": "watch the path"} {
		if got[k] != want {
			t.Errorf("%s = %v, want %v", k, got[k], want)
		}
	}
	res, _ := json.Marshal(got["resolved"])
	want := `{"review":{"effort":"high","effort_source":"file","model":"sonnet","model_source":"default"},` +
		`"work":{"effort":null,"effort_source":"default","model":"opus","model_source":"task"}}`
	if string(res) != want {
		t.Errorf("resolved = %s", res)
	}
}

func TestTaskShowUnknownAndNoPlan(t *testing.T) {
	root := taskRepo(t, "")
	code, out, errOut := run(t, "-C", root, "task", "show", "T9")
	if code != 1 || out != "" || errOut != "vloop: no task T9 — vloop task list shows the plan's tasks\n" {
		t.Errorf("unknown: %d %q %q", code, out, errOut)
	}
	code, out, errOut = run(t, "-C", statusRepo(t, ""), "task", "show", "T1")
	if code != 1 || out != "" || errOut != "vloop: no plan — vloop run <brief> makes one\n" {
		t.Errorf("no plan: %d %q %q", code, out, errOut)
	}
}
