package driver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func copilotCommand(t *testing.T, root string, s Spec, env []string) (args []string, cmdEnv []string, dir string, p Prepared) {
	t.Helper()
	a := copilotAdapter{}
	p, err := a.Prepare(root, "9.9.9", PhaseReview)
	if err != nil {
		t.Fatal(err)
	}
	cmd := a.Command(s, root, env, p)
	return cmd.Args[1:], cmd.Env, cmd.Dir, p
}

func envValue(env []string, name string) (string, bool) {
	for _, kv := range env {
		if v, ok := strings.CutPrefix(kv, name+"="); ok {
			return v, true
		}
	}
	return "", false
}

func TestCopilotArgv(t *testing.T) {
	root := t.TempDir()
	s := Spec{Phase: PhaseReview, Arg: "T3", Model: "m1", Agent: AgentCopilot}
	args, _, dir, p := copilotCommand(t, root, s, nil)
	if dir != root {
		t.Errorf("dir %q, want the root", dir)
	}
	line := " " + strings.Join(args, " ") + " "
	for _, want := range []string{"-p /vloop:review T3 ", " --model m1 ", " --plugin-dir " + p.Plugin + " ", " --no-ask-user ", " --output-format json ", " --allow-all-tools ", " --disable-builtin-mcps "} {
		if !strings.Contains(line, want) {
			t.Errorf("argv lacks %q: %v", want, args)
		}
	}
	if strings.Contains(line, "--effort") {
		t.Errorf("--effort without an effort: %v", args)
	}
	for _, bad := range []string{"--allow-all-paths", "--add-dir", "--sandbox", "--no-sandbox", "--allow-all ", "--yolo"} {
		if strings.Contains(line, bad) {
			t.Errorf("argv holds %q: %v", bad, args)
		}
	}
	for _, d := range copilotDenials {
		if !strings.Contains(line, " --deny-tool="+d+" ") {
			t.Errorf("no denial %q", d)
		}
	}
	s.Effort = "high"
	args, _, _, _ = copilotCommand(t, root, s, nil)
	if !strings.Contains(" "+strings.Join(args, " ")+" ", " --effort high ") {
		t.Errorf("no --effort high: %v", args)
	}
}

func TestCopilotEnvAndHome(t *testing.T) {
	root := t.TempDir()
	base := []string{"VLOOP_ACTIVE_TASK=T1", "VLOOP_GATE_TASK=T1", PlanHashEnv + "=stale", "GH_TOKEN=tok", "COPILOT_HOME=/elsewhere"}
	s := Spec{Phase: PhaseReview, Arg: "T1", Model: "m", PlanSHA: "fresh"}
	_, env, _, p := copilotCommand(t, root, s, base)
	for _, n := range []string{"VLOOP_ACTIVE_TASK", "VLOOP_GATE_TASK"} {
		if _, ok := envValue(env, n); ok {
			t.Errorf("%s passed", n)
		}
	}
	if v, _ := envValue(env, PlanHashEnv); v != "fresh" {
		t.Errorf("plan hash %q", v)
	}
	if v, _ := envValue(env, "GH_TOKEN"); v != "tok" {
		t.Errorf("GH_TOKEN %q", v)
	}
	home, _ := envValue(env, "COPILOT_HOME")
	if home != p.Home || !strings.HasPrefix(home, filepath.Join(root, ".vloop", "tmp")+string(filepath.Separator)) {
		t.Errorf("COPILOT_HOME %q", home)
	}
	if ents, err := os.ReadDir(home); err != nil || len(ents) != 0 {
		t.Errorf("COPILOT_HOME not an empty directory: %v %v", ents, err)
	}
	_, env2, _, p2 := copilotCommand(t, root, s, base)
	if p2.Home == p.Home {
		t.Error("COPILOT_HOME reused across sessions")
	}
	_ = env2
	if _, ok := envValue(env, "COPILOT_CUSTOM_INSTRUCTIONS_DIRS"); ok {
		t.Error("guidelines directory without guidelines")
	}
}

func TestCopilotGuidelines(t *testing.T) {
	root := t.TempDir()
	g := filepath.Join(root, GuidelinesFile)
	if err := os.MkdirAll(filepath.Dir(g), 0o755); err != nil {
		t.Fatal(err)
	}
	s := Spec{Phase: PhaseReview, Arg: "T1", Model: "m"}
	if err := os.WriteFile(g, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, env, _, _ := copilotCommand(t, root, s, nil); func() bool { _, ok := envValue(env, "COPILOT_CUSTOM_INSTRUCTIONS_DIRS"); return ok }() {
		t.Error("an empty guidelines file produced a directory")
	}
	if err := os.WriteFile(g, []byte("- a lesson\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, env, _, p := copilotCommand(t, root, s, []string{"COPILOT_CUSTOM_INSTRUCTIONS_DIRS=/other"})
	dir, _ := envValue(env, "COPILOT_CUSTOM_INSTRUCTIONS_DIRS")
	if dir != p.Guidelines || !strings.HasPrefix(dir, filepath.Join(root, ".vloop", "tmp")+string(filepath.Separator)) {
		t.Fatalf("COPILOT_CUSTOM_INSTRUCTIONS_DIRS %q", dir)
	}
	b, err := os.ReadFile(filepath.Join(dir, copilotGuidelinesName))
	if err != nil || string(b) != "- a lesson\n" {
		t.Errorf("guidelines not in the directory: %q %v", b, err)
	}
}

func TestCopilotSupportsReviewOnly(t *testing.T) {
	a, err := adapterFor(AgentCopilot, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, ph := range []string{PhasePlan, PhaseWork, PhaseGateReview, PhaseVerify} {
		if a.Supports(ph) {
			t.Errorf("copilot supports %s", ph)
		}
	}
	if !a.Supports(PhaseReview) {
		t.Error("copilot does not support review")
	}
}
