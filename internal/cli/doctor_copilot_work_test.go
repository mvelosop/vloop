package cli

import (
	"strings"
	"testing"
)

const copilotWorkConfig = doctorCheckConfig + "[agent]\nwork = \"copilot\"\n"

func TestDoctorCopilotWorkExperimental(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", copilotWorkConfig+"[model]\nwork = \"gpt-x\"\n")
	e.copilot(true)
	e.want("copilot", resPass)
	e.want("copilot login", resPass)
	const warn = "agent.work is copilot: experimental (docs/guide/configuration.md)"
	out, code := e.run("doctor")
	if code != 0 || !strings.Contains(out, "! "+warn+"\n") {
		t.Errorf("doctor exit %d, no warning line:\n%s", code, out)
	}
	errOut, err := e.preflight()
	if err != nil || !strings.Contains(errOut, "! "+warn+"\n") {
		t.Errorf("preflight = %v, no warning:\n%s", err, errOut)
	}
}

func TestRunPreflightRefusesDefaultWorkModel(t *testing.T) {
	e := newDoctorEnv(t)
	e.copilot(true)
	e.write(".vloop/config.toml", copilotWorkConfig)
	_, err := e.preflight()
	want := `agent.work is copilot but model.work is the default "sonnet", a Claude alias: set model.work to a Copilot model id`
	if err == nil || err.Error() != want {
		t.Errorf("preflight error = %v, want %q", err, want)
	}
	t.Setenv("VLOOP_MODEL_WORK", "gpt-x")
	if _, err := e.preflight(); err != nil {
		t.Errorf("a model.work from the environment was refused: %v", err)
	}
}

func TestCopilotWorkSkillWithPluginNameStopsTheRun(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", copilotWorkConfig+"[model]\nwork = \"gpt-x\"\n")
	e.copilot(true)
	e.skill("work")
	out, code := e.run("doctor")
	if code != 1 || !strings.Contains(out, "✗ .claude/skills/work has the name of vloop's work skill") {
		t.Errorf("doctor exit %d:\n%s", code, out)
	}
	_, err := e.preflight()
	want := ".claude/skills/work has the name of vloop's work skill; a copilot session may load it instead — rename it"
	if err == nil || err.Error() != want {
		t.Errorf("preflight error = %v, want %q", err, want)
	}
}

func TestCopilotWorkLoginRefusal(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", copilotWorkConfig+"[model]\nwork = \"gpt-x\"\n")
	e.copilot(true)
	e.gh("exit 1")
	for _, k := range []string{"COPILOT_GITHUB_TOKEN", "GH_TOKEN", "GITHUB_TOKEN"} {
		t.Setenv(k, "")
	}
	_, err := e.preflight()
	want := "agent.work is copilot but no GitHub login reaches its sessions: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login"
	if err == nil || err.Error() != want {
		t.Errorf("preflight error = %v, want %q", err, want)
	}
	e.want("copilot login", resProblem)
}

func TestClaudeWorkIgnoresPluginNamedSkill(t *testing.T) {
	e := newDoctorEnv(t)
	e.skill("work")
	if errOut, err := e.preflight(); err != nil || strings.Contains(errOut, "skills/") || strings.Contains(errOut, "agent.work") {
		t.Errorf("preflight = %v:\n%s", err, errOut)
	}
}
