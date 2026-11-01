package cli

import (
	"strings"
	"testing"
)

const copilotPlanConfig = doctorCheckConfig + "[agent]\nplan = \"copilot\"\n"

func TestDoctorCopilotPlanExperimental(t *testing.T) {
	e := newDoctorEnv(t)
	// Planning has no switch: the key is in use whatever run.verify says, and
	// the preflight does not look at whether a plan exists (a resume).
	e.write(".vloop/config.toml", copilotPlanConfig+"[model]\nplan = \"gpt-x\"\n[run]\nverify = \"false\"\n")
	e.copilot(true)
	e.want("copilot", resPass)
	e.want("copilot login", resPass)
	const warn = "agent.plan is copilot: experimental (docs/guide/configuration.md)"
	out, code := e.run("doctor")
	if code != 0 || !strings.Contains(out, "! "+warn+"\n") {
		t.Errorf("doctor exit %d, no warning line:\n%s", code, out)
	}
	errOut, err := e.preflight()
	if err != nil || !strings.Contains(errOut, "! "+warn+"\n") {
		t.Errorf("preflight = %v, no warning:\n%s", err, errOut)
	}
}

func TestRunPreflightRefusesDefaultPlanModel(t *testing.T) {
	e := newDoctorEnv(t)
	e.copilot(true)
	e.write(".vloop/config.toml", copilotPlanConfig)
	_, err := e.preflight()
	want := `agent.plan is copilot but model.plan is the default "opus", a Claude alias: set model.plan to a Copilot model id`
	if err == nil || err.Error() != want {
		t.Errorf("preflight error = %v, want %q", err, want)
	}
}

func TestCopilotPlanSkillWithPluginNameStopsTheRun(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", copilotPlanConfig+"[model]\nplan = \"gpt-x\"\n")
	e.copilot(true)
	e.skill("plan")
	out, code := e.run("doctor")
	if code != 1 || !strings.Contains(out, "✗ .claude/skills/plan has the name of vloop's plan skill") {
		t.Errorf("doctor exit %d:\n%s", code, out)
	}
	_, err := e.preflight()
	want := ".claude/skills/plan has the name of vloop's plan skill; a copilot session may load it instead — rename it"
	if err == nil || err.Error() != want {
		t.Errorf("preflight error = %v, want %q", err, want)
	}
}

func TestCopilotPlanLoginRefusal(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", copilotPlanConfig+"[model]\nplan = \"gpt-x\"\n")
	e.copilot(true)
	e.gh("exit 1")
	for _, k := range []string{"COPILOT_GITHUB_TOKEN", "GH_TOKEN", "GITHUB_TOKEN"} {
		t.Setenv(k, "")
	}
	_, err := e.preflight()
	want := "agent.plan is copilot but no GitHub login reaches its sessions: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login"
	if err == nil || err.Error() != want {
		t.Errorf("preflight error = %v, want %q", err, want)
	}
	e.want("copilot login", resProblem)
}

func TestPlanOnClaudeSaysNothing(t *testing.T) {
	e := newDoctorEnv(t)
	e.copilot(true)
	e.skill("plan")
	out, code := e.run("doctor")
	if code != 0 || strings.Contains(out, "agent.plan") || strings.Contains(out, "has the name of vloop's") {
		t.Errorf("doctor exit %d:\n%s", code, out)
	}
	if errOut, err := e.preflight(); err != nil || strings.Contains(errOut, "agent.plan") {
		t.Errorf("preflight = %v:\n%s", err, errOut)
	}
}
