package cli

import (
	"strings"
	"testing"
)

const copilotGateReviewConfig = doctorCheckConfig + "[agent]\ngate-review = \"copilot\"\n"

func TestDoctorCopilotGateReviewExperimental(t *testing.T) {
	e := newDoctorEnv(t)
	// The gate review has no switch: the key is in use whatever run.verify says.
	e.write(".vloop/config.toml", copilotGateReviewConfig+"[model]\ngate-review = \"gpt-x\"\n[run]\nverify = \"false\"\n")
	e.copilot(true)
	e.want("copilot", resPass)
	e.want("copilot login", resPass)
	const warn = "agent.gate-review is copilot: experimental (docs/guide/configuration.md)"
	out, code := e.run("doctor")
	if code != 0 || !strings.Contains(out, "! "+warn+"\n") {
		t.Errorf("doctor exit %d, no warning line:\n%s", code, out)
	}
	errOut, err := e.preflight()
	if err != nil || !strings.Contains(errOut, "! "+warn+"\n") {
		t.Errorf("preflight = %v, no warning:\n%s", err, errOut)
	}
}

func TestRunPreflightRefusesDefaultGateReviewModel(t *testing.T) {
	e := newDoctorEnv(t)
	e.copilot(true)
	e.write(".vloop/config.toml", copilotGateReviewConfig)
	_, err := e.preflight()
	want := `agent.gate-review is copilot but model.gate-review is the default "sonnet", a Claude alias: set model.gate-review to a Copilot model id`
	if err == nil || err.Error() != want {
		t.Errorf("preflight error = %v, want %q", err, want)
	}
}

func TestCopilotGateReviewSkillWithPluginNameStopsTheRun(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", copilotGateReviewConfig+"[model]\ngate-review = \"gpt-x\"\n")
	e.copilot(true)
	e.skill("gate-review")
	out, code := e.run("doctor")
	if code != 1 || !strings.Contains(out, "✗ .claude/skills/gate-review has the name of vloop's gate-review skill") {
		t.Errorf("doctor exit %d:\n%s", code, out)
	}
	_, err := e.preflight()
	want := ".claude/skills/gate-review has the name of vloop's gate-review skill; a copilot session may load it instead — rename it"
	if err == nil || err.Error() != want {
		t.Errorf("preflight error = %v, want %q", err, want)
	}
}

func TestCopilotGateReviewLoginRefusal(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", copilotGateReviewConfig+"[model]\ngate-review = \"gpt-x\"\n")
	e.copilot(true)
	e.gh("exit 1")
	for _, k := range []string{"COPILOT_GITHUB_TOKEN", "GH_TOKEN", "GITHUB_TOKEN"} {
		t.Setenv(k, "")
	}
	_, err := e.preflight()
	want := "agent.gate-review is copilot but no GitHub login reaches its sessions: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login"
	if err == nil || err.Error() != want {
		t.Errorf("preflight error = %v, want %q", err, want)
	}
	e.want("copilot login", resProblem)
}

func TestGateReviewOnClaudeSaysNothing(t *testing.T) {
	e := newDoctorEnv(t)
	e.copilot(true)
	e.skill("gate-review")
	out, code := e.run("doctor")
	if code != 0 || strings.Contains(out, "agent.gate-review") || strings.Contains(out, "has the name of vloop's") {
		t.Errorf("doctor exit %d:\n%s", code, out)
	}
	if errOut, err := e.preflight(); err != nil || strings.Contains(errOut, "agent.gate-review") {
		t.Errorf("preflight = %v:\n%s", err, errOut)
	}
}
