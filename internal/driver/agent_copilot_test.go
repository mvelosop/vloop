package driver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	if ents, err := os.ReadDir(home); err != nil || len(ents) != 1 || ents[0].Name() != "hooks" {
		t.Errorf("COPILOT_HOME holds more than the hooks directory (H2 adds the fence hook): %v %v", ents, err)
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

func TestCopilotSupportsWorkReviewAndVerify(t *testing.T) {
	a, err := adapterFor(AgentCopilot, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, ph := range []string{PhasePlan, PhaseWork, PhaseReview, PhaseSecondReview, PhaseVerify, PhaseGateReview, PhaseReviewBrief} {
		if !a.Supports(ph) {
			t.Errorf("copilot does not support %s", ph)
		}
	}
}

func TestCopilotWorkArgvAndHooks(t *testing.T) {
	root := t.TempDir()
	a := copilotAdapter{}
	p, err := a.Prepare(root, "9.9.9", PhaseWork)
	if err != nil {
		t.Fatal(err)
	}
	args := a.Command(Spec{Phase: PhaseWork, Arg: "T3", Model: "m1", Agent: AgentCopilot}, root, nil, p).Args[1:]
	line := " " + strings.Join(args, " ") + " "
	if !strings.Contains(line, "-p /vloop:work T3 ") {
		t.Errorf("argv lacks the work prompt: %v", args)
	}
	if strings.Contains(line, " --deny-tool=shell(rm:*) ") {
		t.Errorf("work argv holds shell(rm:*): %v", args)
	}
	for _, d := range copilotDenials {
		if d != "shell(rm:*)" && !strings.Contains(line, " --deny-tool="+d+" ") {
			t.Errorf("no denial %q", d)
		}
	}
	data, err := os.ReadFile(filepath.Join(p.Home, "hooks", "vloop-fence.json"))
	if err != nil || !strings.Contains(string(data), " fence hook --phase work --root ") || strings.Contains(string(data), "--phase review") {
		t.Errorf("hooks file %s: %v", data, err)
	}
}

func TestCopilotVerifyArgvAndHooks(t *testing.T) {
	root := t.TempDir()
	a := copilotAdapter{}
	p, err := a.Prepare(root, "9.9.9", PhaseVerify)
	if err != nil {
		t.Fatal(err)
	}
	args := a.Command(Spec{Phase: PhaseVerify, Model: "m1", Agent: AgentCopilot}, root, nil, p).Args[1:]
	line := " " + strings.Join(args, " ") + " "
	if !strings.Contains(line, "-p /vloop:verify ") {
		t.Errorf("argv lacks the verify prompt: %v", args)
	}
	for _, d := range copilotDenials {
		if !strings.Contains(line, " --deny-tool="+d+" ") {
			t.Errorf("no denial %q", d)
		}
	}
	if !strings.Contains(line, " --deny-tool=shell(rm:*) ") {
		t.Errorf("verify argv lacks shell(rm:*): %v", args)
	}
	data, err := os.ReadFile(filepath.Join(p.Home, "hooks", "vloop-fence.json"))
	if err != nil || !strings.Contains(string(data), " fence hook --phase verify --root ") || strings.Contains(string(data), "--phase review") {
		t.Errorf("hooks file %s: %v", data, err)
	}
}

func TestCopilotVerifyRecordIsSessionV4(t *testing.T) {
	r := &Runner{}
	for _, c := range []struct{ phase, agent, want string }{
		{PhaseVerify, AgentCopilot, "session/v4"},
		{PhaseVerify, "claude", "session/v3"},
		{PhaseReview, AgentCopilot, "session/v3"},
		{PhaseGateReview, AgentCopilot, "session/v4"},
		{PhaseGateReview, "claude", "session/v3"},
		{PhasePlan, AgentCopilot, "session/v4"},
		{PhasePlan, "claude", "session/v3"},
		{PhaseReviewBrief, AgentCopilot, "session/v5"},
		{PhaseReviewBrief, "claude", "session/v5"},
	} {
		rec := r.record(Spec{Phase: c.phase, Model: "m"}, c.agent, "", time.Now(), time.Now(), Output{})
		if rec["schema"] != c.want || rec["phase"] != c.phase || rec["agent"] != c.agent {
			t.Errorf("%s/%s: schema %v, want %s", c.phase, c.agent, rec["schema"], c.want)
		}
	}
}

func TestCopilotGateReviewArgvAndHooks(t *testing.T) {
	root := t.TempDir()
	a := copilotAdapter{}
	p, err := a.Prepare(root, "9.9.9", PhaseGateReview)
	if err != nil {
		t.Fatal(err)
	}
	args := a.Command(Spec{Phase: PhaseGateReview, Model: "m1", Agent: AgentCopilot}, root, nil, p).Args[1:]
	line := " " + strings.Join(args, " ") + " "
	if !strings.Contains(line, "-p /vloop:gate-review ") {
		t.Errorf("argv lacks the gate-review prompt: %v", args)
	}
	for _, d := range copilotDenials {
		if !strings.Contains(line, " --deny-tool="+d+" ") {
			t.Errorf("no denial %q", d)
		}
	}
	if !strings.Contains(line, " --deny-tool=shell(rm:*) ") || !strings.Contains(line, " --deny-tool=shell(git push:*) ") {
		t.Errorf("gate-review argv lacks shell(rm:*) or shell(git push:*): %v", args)
	}
	data, err := os.ReadFile(filepath.Join(p.Home, "hooks", "vloop-fence.json"))
	if err != nil || !strings.Contains(string(data), " fence hook --phase gate-review --root ") || strings.Contains(string(data), "--phase review") {
		t.Errorf("hooks file %s: %v", data, err)
	}
}

func TestCopilotPlanArgvAndHooks(t *testing.T) {
	root := t.TempDir()
	a := copilotAdapter{}
	p, err := a.Prepare(root, "9.9.9", PhasePlan)
	if err != nil {
		t.Fatal(err)
	}
	args := a.Command(Spec{Phase: PhasePlan, Arg: "docs/briefs/b.md", Model: "m1", Agent: AgentCopilot}, root, nil, p).Args[1:]
	line := " " + strings.Join(args, " ") + " "
	if !strings.Contains(line, " /vloop:plan docs/briefs/b.md ") {
		t.Errorf("argv lacks the plan prompt: %v", args)
	}
	for _, d := range copilotDenials {
		if !strings.Contains(line, " --deny-tool="+d+" ") {
			t.Errorf("no denial %q", d)
		}
	}
	data, err := os.ReadFile(filepath.Join(p.Home, "hooks", "vloop-fence.json"))
	if err != nil || !strings.Contains(string(data), " fence hook --phase plan --root ") {
		t.Errorf("hooks file %s: %v", data, err)
	}
}

func TestCopilotReviewBriefArgvAndHooks(t *testing.T) {
	root := t.TempDir()
	a := copilotAdapter{}
	p, err := a.Prepare(root, "9.9.9", PhaseReviewBrief)
	if err != nil {
		t.Fatal(err)
	}
	args := a.Command(Spec{Phase: PhaseReviewBrief, Arg: "docs/briefs/b.md", Model: "m1", Agent: AgentCopilot}, root, nil, p).Args[1:]
	line := " " + strings.Join(args, " ") + " "
	if !strings.Contains(line, "-p /vloop:review-brief docs/briefs/b.md ") {
		t.Errorf("argv lacks the review-brief prompt: %v", args)
	}
	for _, d := range copilotDenials {
		if !strings.Contains(line, " --deny-tool="+d+" ") {
			t.Errorf("no denial %q", d)
		}
	}
	data, err := os.ReadFile(filepath.Join(p.Home, "hooks", "vloop-fence.json"))
	if err != nil || !strings.Contains(string(data), " fence hook --phase review-brief --root ") {
		t.Errorf("hooks file %s: %v", data, err)
	}
}

func TestCopilotHookBinaryOption(t *testing.T) {
	root := t.TempDir()
	a := copilotAdapter{hook: "/opt/it's/vloop"}
	p, err := a.Prepare(root, "9.9.9", PhaseReview)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(p.Home, "hooks", "vloop-fence.json"))
	if err != nil || !strings.Contains(string(data), `'/opt/it'\\''s/vloop' fence hook --phase review --root `) {
		t.Errorf("the hook does not run the named binary: %s %v", data, err)
	}
	exe, _ := os.Executable()
	p, err = copilotAdapter{}.Prepare(t.TempDir(), "9.9.9", PhaseReview)
	if err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(filepath.Join(p.Home, "hooks", "vloop-fence.json"))
	if !strings.Contains(string(data), "'"+exe+"' fence hook") {
		t.Errorf("by default the hook runs os.Executable(): %s", data)
	}
}

func TestRunnerKeepsTranscriptOnlyWhenAsked(t *testing.T) {
	bin := t.TempDir()
	fake := filepath.Join(bin, "claude")
	script := "#!/bin/sh\nprintf '{\"type\":\"x\"}\\n{\"total_cost_usd\":0.5,\"num_turns\":1}\\n'\n"
	if err := os.WriteFile(fake, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, keep := range []bool{false, true} {
		root, dir := t.TempDir(), t.TempDir()
		r := &Runner{Root: root, Version: "9.9.9", RunID: "r", RunDir: dir, Claude: fake, Env: []string{"PATH=" + os.Getenv("PATH")}}
		tr := filepath.Join(dir, "transcript.jsonl")
		if keep {
			r.Transcript = tr
		}
		if _, err := r.Run(Spec{Phase: PhaseReview, Arg: "T1", Model: "m", Iteration: 1}); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(tr)
		if keep && (err != nil || string(data) != "{\"type\":\"x\"}\n{\"total_cost_usd\":0.5,\"num_turns\":1}\n") {
			t.Errorf("transcript = %q, %v", data, err)
		}
		if !keep && err == nil {
			t.Errorf("a transcript was kept without being asked")
		}
	}
}
