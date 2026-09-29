package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func validateRepo(t *testing.T, cfg string) string {
	t.Helper()
	root := taskRepo(t, cfg)
	t.Setenv("VLOOP_AREAS", "")
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestTaskValidateOK(t *testing.T) {
	root := validateRepo(t, "")
	code, out, errOut := run(t, "-C", root, "task", "validate")
	if code != 0 || out != "plan ok\n" || errOut != "" {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
}

func TestTaskValidateProblems(t *testing.T) {
	root := validateRepo(t, "areas = [\"cli\", \"docs\"]\n")
	before, _ := os.ReadFile(filepath.Join(root, ".vloop", "state", "state.json"))
	code, out, errOut := run(t, "-C", root, "task", "validate")
	want := "✗ T2: area \"config\" is not in areas (cli, docs)\n" +
		"✗ T3: no area — areas is set\n" +
		"2 problem(s)\n"
	if code != 1 || out != want || errOut != "" {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
	after, _ := os.ReadFile(filepath.Join(root, ".vloop", "state", "state.json"))
	if string(before) != string(after) {
		t.Fatal("validate modified the plan")
	}
}

func TestTaskValidateJSON(t *testing.T) {
	root := validateRepo(t, "areas = [\"cli\", \"docs\"]\n")
	code, out, _ := run(t, "-C", root, "task", "validate", "--json")
	var got struct {
		OK       bool
		Problems []string
		Warnings []string
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil || code != 1 || got.OK || len(got.Problems) != 2 {
		t.Fatalf("code %d out %q err %v", code, out, err)
	}
	root = validateRepo(t, "")
	code, out, _ = run(t, "-C", root, "task", "validate", "--json")
	if err := json.Unmarshal([]byte(out), &got); err != nil || code != 0 || !got.OK || len(got.Problems) != 0 {
		t.Fatalf("code %d out %q err %v", code, out, err)
	}
}

func TestTaskValidateNoPlan(t *testing.T) {
	root := statusRepo(t, "")
	code, out, errOut := run(t, "-C", root, "task", "validate")
	if code != 1 || out != "" || errOut != "vloop: no plan: .vloop/state/state.json\n" {
		t.Fatalf("code %d out %q err %q", code, out, errOut)
	}
	code, out, _ = run(t, "-C", root, "task", "validate", "--json")
	if code != 1 || !strings.Contains(out, `"error"`) {
		t.Fatalf("code %d out %q", code, out)
	}
}
