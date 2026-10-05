package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mvelosop/vloop/internal/state"
)

// amendRepo is a taskRepo whose plan is rewritten in the canonical format, so
// a write can be compared with it.
func amendRepo(t *testing.T, cfg string) string {
	t.Helper()
	t.Setenv("VLOOP_AREAS", "")
	root := taskRepo(t, cfg)
	p, err := state.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	b, err := state.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(state.Path(root), b, 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func planBytes(t *testing.T, root string) string {
	t.Helper()
	b, err := os.ReadFile(state.Path(root))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// refuse runs args and wants the given exit code and stderr, an empty stdout
// and the plan untouched.
func refuse(t *testing.T, root string, code int, stderr string, args ...string) {
	t.Helper()
	before := planBytes(t, root)
	gotCode, out, errOut := run(t, append([]string{"-C", root}, args...)...)
	if gotCode != code || out != "" || !strings.HasPrefix(errOut, stderr) {
		t.Fatalf("%v: code %d out %q err %q, want code %d err %q", args, gotCode, out, errOut, code, stderr)
	}
	if planBytes(t, root) != before {
		t.Fatalf("%v: the plan was modified", args)
	}
	entries, _ := os.ReadDir(filepath.Dir(state.Path(root)))
	if len(entries) != 1 {
		t.Fatalf("%v: files left in .vloop/state: %v", args, entries)
	}
}

// changed runs args, wants success, and returns the plan with only mutate
// applied (and updated refreshed) compared to before.
func changed(t *testing.T, root string, mutate func(*state.Plan), args ...string) {
	t.Helper()
	p, _ := state.Load(root)
	code, out, errOut := run(t, append([]string{"-C", root}, args...)...)
	if code != 0 || out != "" || errOut != "" {
		t.Fatalf("%v: code %d out %q err %q", args, code, out, errOut)
	}
	got, err := state.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.Updated == "2026-01-01T09:00:00Z" {
		t.Fatalf("%v: updated was not refreshed", args)
	}
	mutate(p)
	p.Updated = got.Updated
	want, _ := state.Marshal(p)
	if planBytes(t, root) != string(want) {
		t.Fatalf("%v: plan is not the prior plan with only the change:\n%s", args, planBytes(t, root))
	}
	entries, _ := os.ReadDir(filepath.Dir(state.Path(root)))
	if len(entries) != 1 {
		t.Fatalf("%v: files left in .vloop/state: %v", args, entries)
	}
}

func TestTaskResetChanges(t *testing.T) {
	root := amendRepo(t, "")
	changed(t, root, func(p *state.Plan) { p.Tasks[2].Status, p.Tasks[2].Attempts = "pending", 0 }, "task", "reset", "T3")
}

func TestTaskResetRefusals(t *testing.T) {
	root := amendRepo(t, "")
	refuse(t, root, 1, "vloop: no task T9 — vloop task list shows the plan's tasks\n", "task", "reset", "T9")
	empty := statusRepo(t, "")
	if code, _, errOut := run(t, "-C", empty, "task", "reset", "T1"); code != 1 || errOut != "vloop: no plan — vloop run <brief> makes one\n" {
		t.Fatalf("code %d err %q", code, errOut)
	}
}

func TestTaskNoteChanges(t *testing.T) {
	root := amendRepo(t, "")
	changed(t, root, func(p *state.Plan) { p.Tasks[1].Notes = "use <b> & c" }, "task", "note", "T2", "use <b> & c")
	if !strings.Contains(planBytes(t, root), `"notes": "use <b> & c"`) {
		t.Fatal("note was escaped")
	}
}

func TestTaskNoteUnknownTask(t *testing.T) {
	root := amendRepo(t, "")
	refuse(t, root, 1, "vloop: no task T9 — vloop task list shows the plan's tasks\n", "task", "note", "T9", "x")
}

func TestTaskDropChanges(t *testing.T) {
	root := amendRepo(t, "")
	changed(t, root, func(p *state.Plan) { p.Tasks = p.Tasks[:2] }, "task", "drop", "T3")
}

func TestTaskDropRefusals(t *testing.T) {
	root := amendRepo(t, "")
	refuse(t, root, 1, "vloop: T2 depends on T1\n", "task", "drop", "T1")
	refuse(t, root, 1, "vloop: no task T9 — vloop task list shows the plan's tasks\n", "task", "drop", "T9")
}

func TestTaskSetChanges(t *testing.T) {
	root := amendRepo(t, "areas = [\"cli\", \"config\", \"docs\"]\n")
	changed(t, root, func(p *state.Plan) { p.Tasks[2].Area = "docs" }, "task", "set", "T3", "area", "docs")
	changed(t, root, func(p *state.Plan) { p.Tasks[2].Kind = "fix" }, "task", "set", "T3", "kind", "fix")
	changed(t, root, func(p *state.Plan) { p.Tasks[2].Model = &state.Sessions{Work: "opus"} }, "task", "set", "T3", "model.work", "opus")
	changed(t, root, func(p *state.Plan) { p.Tasks[2].Model.Review = "fable" }, "task", "set", "T3", "model.review", "fable")
	changed(t, root, func(p *state.Plan) { p.Tasks[2].Effort = &state.Sessions{Work: "max"} }, "task", "set", "T3", "effort.work", "max")
	changed(t, root, func(p *state.Plan) { p.Tasks[2].Effort.Review = "low" }, "task", "set", "T3", "effort.review", "low")
}

func TestTaskSetClears(t *testing.T) {
	root := amendRepo(t, "")
	changed(t, root, func(p *state.Plan) { p.Tasks[0].Area = "" }, "task", "set", "T1", "area", "")
	if strings.Contains(planBytes(t, root), `"area": "cli"`) {
		t.Fatal("area key was not removed")
	}
	changed(t, root, func(p *state.Plan) { p.Tasks[2].Model = &state.Sessions{Work: "opus"} }, "task", "set", "T3", "model.work", "opus")
	changed(t, root, func(p *state.Plan) { p.Tasks[2].Model = nil }, "task", "set", "T3", "model.work", "")
	var m map[string]any
	_ = json.Unmarshal([]byte(planBytes(t, root)), &m)
	if _, ok := m["tasks"].([]any)[2].(map[string]any)["model"]; ok {
		t.Fatal("empty model object left behind")
	}
}

func TestTaskSetRefusals(t *testing.T) {
	root := amendRepo(t, "areas = [\"cli\", \"config\"]\n")
	refuse(t, root, 2, `vloop: invalid value "ops" for area: want one of cli, config`+"\n", "task", "set", "T2", "area", "ops")
	refuse(t, root, 2, `vloop: invalid value "bug" for kind: want one of feature, fix, refactor, test, docs, chore`+"\n", "task", "set", "T2", "kind", "bug")
	refuse(t, root, 2, `vloop: invalid value "turbo" for effort.work: want one of low, medium, high, xhigh, max`+"\n", "task", "set", "T2", "effort.work", "turbo")
	refuse(t, root, 2, "vloop: cannot set \"title\"", "task", "set", "T2", "title", "x")
	refuse(t, root, 1, "vloop: no task T9 — vloop task list shows the plan's tasks\n", "task", "set", "T9", "kind", "fix")
	// T3 has no area and areas is set, so clearing T2's area would fail validate
	// only for the cleared task; the plan already has that problem, so any write is refused.
	refuse(t, root, 1, "vloop: refusing to write", "task", "set", "T2", "area", "")
}

func TestTaskSetClearAreaRefusedWhenAreasSet(t *testing.T) {
	root := amendRepo(t, "areas = [\"cli\", \"config\", \"docs\"]\n")
	changed(t, root, func(p *state.Plan) { p.Tasks[2].Area = "docs" }, "task", "set", "T3", "area", "docs")
	refuse(t, root, 1, "vloop: refusing to write", "task", "set", "T3", "area", "")
}

func TestTaskAmendInvalidPlan(t *testing.T) {
	root := amendRepo(t, "")
	p := planBytes(t, root)
	bad := strings.Replace(p, `"status": "pending"`, `"status": "paused"`, 1)
	if bad == p {
		t.Fatal("fixture not changed")
	}
	if err := os.WriteFile(state.Path(root), []byte(bad), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"task", "reset", "T2"},
		{"task", "note", "T2", "x"},
		{"task", "drop", "T3"},
		{"task", "set", "T3", "kind", "fix"},
	} {
		refuse(t, root, 1, "vloop: plan is not valid — run vloop task validate\n", args...)
	}
}
