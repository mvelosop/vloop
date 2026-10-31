package cli

import (
	"strings"
	"testing"
)

const copilotVerifyConfig = doctorCheckConfig + "[agent]\nverify = \"copilot\"\n"

func TestDoctorCopilotVerifyExperimental(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", copilotVerifyConfig+"[model]\nverify = \"gpt-x\"\n")
	e.copilot(true)
	e.want("copilot", resPass)
	e.want("copilot login", resPass)
	const warn = "agent.verify is copilot: experimental (docs/guide/configuration.md)"
	out, code := e.run("doctor")
	if code != 0 || !strings.Contains(out, "! "+warn+"\n") {
		t.Errorf("doctor exit %d, no warning line:\n%s", code, out)
	}
	errOut, err := e.preflight()
	if err != nil || !strings.Contains(errOut, "! "+warn+"\n") {
		t.Errorf("preflight = %v, no warning:\n%s", err, errOut)
	}
}

func TestRunPreflightRefusesDefaultVerifyModel(t *testing.T) {
	e := newDoctorEnv(t)
	e.copilot(true)
	e.write(".vloop/config.toml", copilotVerifyConfig)
	_, err := e.preflight()
	want := `agent.verify is copilot but model.verify is the default "sonnet", a Claude alias: set model.verify to a Copilot model id`
	if err == nil || err.Error() != want {
		t.Errorf("preflight error = %v, want %q", err, want)
	}
}

func TestCopilotVerifySkillWithPluginNameStopsTheRun(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", copilotVerifyConfig+"[model]\nverify = \"gpt-x\"\n")
	e.copilot(true)
	e.skill("verify")
	out, code := e.run("doctor")
	if code != 1 || !strings.Contains(out, "✗ .claude/skills/verify has the name of vloop's verify skill") {
		t.Errorf("doctor exit %d:\n%s", code, out)
	}
	_, err := e.preflight()
	want := ".claude/skills/verify has the name of vloop's verify skill; a copilot session may load it instead — rename it"
	if err == nil || err.Error() != want {
		t.Errorf("preflight error = %v, want %q", err, want)
	}
}

func TestCopilotVerifyLoginRefusal(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", copilotVerifyConfig+"[model]\nverify = \"gpt-x\"\n")
	e.copilot(true)
	e.gh("exit 1")
	for _, k := range []string{"COPILOT_GITHUB_TOKEN", "GH_TOKEN", "GITHUB_TOKEN"} {
		t.Setenv(k, "")
	}
	_, err := e.preflight()
	want := "agent.verify is copilot but no GitHub login reaches its sessions: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login"
	if err == nil || err.Error() != want {
		t.Errorf("preflight error = %v, want %q", err, want)
	}
	e.want("copilot login", resProblem)
}

func TestCopilotVerifyOffWithRunVerifyFalse(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", copilotVerifyConfig+"[run]\nverify = \"false\"\n")
	e.copilot(true)
	e.skill("verify")
	out, code := e.run("doctor")
	if code != 0 || strings.Contains(out, "agent.verify") || strings.Contains(out, "has the name of vloop's") {
		t.Errorf("doctor exit %d:\n%s", code, out)
	}
	if errOut, err := e.preflight(); err != nil || strings.Contains(errOut, "agent.verify") {
		t.Errorf("preflight = %v:\n%s", err, errOut)
	}
}
