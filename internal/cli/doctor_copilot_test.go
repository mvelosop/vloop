package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

const copilotWarning = "agent.review is copilot: experimental — review sessions are not fenced yet (docs/guide/configuration.md)"

// copilot installs a stub copilot beside the stub claude; versionOK false makes
// its --version fail.
func (e *doctorEnv) copilot(versionOK bool) {
	e.t.Helper()
	script := "#!/bin/sh\nexit 0\n"
	if !versionOK {
		script = "#!/bin/sh\nexit 4\n"
	}
	if err := os.WriteFile(filepath.Join(e.stub, "copilot"), []byte(script), 0o755); err != nil {
		e.t.Fatal(err)
	}
	e.gh("echo gho_stubtoken")
}

// gh installs a stub gh running body; copilot() installs one that prints a
// token, so the tests that are not about the login have one.
func (e *doctorEnv) gh(body string) {
	e.t.Helper()
	if err := os.WriteFile(filepath.Join(e.stub, "gh"), []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		e.t.Fatal(err)
	}
}

func (e *doctorEnv) skill(name string) {
	e.t.Helper()
	e.write(".claude/skills/"+name+"/SKILL.md", "---\nname: "+name+"\n---\n")
}

// preflight runs the run preflight and returns its stderr and error.
func (e *doctorEnv) preflight() (string, error) {
	var errOut bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetErr(&errOut)
	err := runPreflight(e.build, e.repo, cmd)
	return errOut.String(), err
}

func TestDoctorCopilotFoundAndExperimental(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", doctorCheckConfig+"[agent]\nreview = \"copilot\"\n[model]\nreview = \"gpt-x\"\n")
	e.copilot(true)
	e.want("copilot", resPass)
	c, code := e.check("agent.review")
	if code != 0 || c.Result != resWarning || "agent.review "+c.Message != copilotWarning {
		t.Errorf("agent.review = %s %q, exit %d", c.Result, c.Message, code)
	}
	out, _ := e.run("doctor")
	if !strings.Contains(out, "! "+copilotWarning+"\n") {
		t.Errorf("doctor output lacks the warning line:\n%s", out)
	}
	errOut, err := e.preflight()
	if err != nil {
		t.Errorf("preflight: %v", err)
	}
	if !strings.Contains(errOut, "! "+copilotWarning+"\n") {
		t.Errorf("preflight lacks the warning:\n%s", errOut)
	}
}

func TestDoctorCopilotMissingOrBroken(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", doctorCheckConfig+"[agent]\nreview = \"copilot\"\n[model]\nreview = \"gpt-x\"\n")
	c, code := e.check("copilot")
	if c.Result != resProblem || code != 1 || !strings.Contains(c.Message, "not on PATH") {
		t.Errorf("missing copilot = %s %q, exit %d", c.Result, c.Message, code)
	}
	if _, err := e.preflight(); err == nil {
		t.Error("preflight with no copilot succeeded")
	}
	e.copilot(false)
	c, code = e.check("copilot")
	if c.Result != resProblem || code != 1 || !strings.Contains(c.Message, "--version failed") {
		t.Errorf("broken copilot = %s %q, exit %d", c.Result, c.Message, code)
	}
	if _, err := e.preflight(); err == nil {
		t.Error("preflight with a broken copilot succeeded")
	}
}

func TestRunPreflightRefusesDefaultReviewModel(t *testing.T) {
	e := newDoctorEnv(t)
	e.copilot(true)
	e.write(".vloop/config.toml", doctorCheckConfig+"[agent]\nreview = \"copilot\"\n")
	_, err := e.preflight()
	want := `agent.review is copilot but model.review is the default "sonnet", a Claude alias: set model.review to a Copilot model id`
	if err == nil || err.Error() != want {
		t.Errorf("preflight error = %v, want %q", err, want)
	}
	e.write(".vloop/config.toml", doctorCheckConfig+"[agent]\nreview = \"copilot\"\n")
	t.Setenv("VLOOP_MODEL_REVIEW", "gpt-x")
	if _, err := e.preflight(); err != nil {
		t.Errorf("a model.review from the environment was refused: %v", err)
	}
}

func TestDoctorCopilotSkillCollisions(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", doctorCheckConfig+"[agent]\nreview = \"copilot\"\n[model]\nreview = \"gpt-x\"\n")
	e.copilot(true)
	e.skill("loop-review")
	e.skill("reviewer")
	out, code := e.run("doctor")
	if code != 0 {
		t.Errorf("doctor exit %d\n%s", code, out)
	}
	if !strings.Contains(out, "! .claude/skills/loop-review may be loaded instead of /vloop:review by a copilot session\n") {
		t.Errorf("no collision warning:\n%s", out)
	}
	if strings.Contains(out, "skills/reviewer") {
		t.Errorf("warned of reviewer:\n%s", out)
	}
	errOut, _ := e.preflight()
	if !strings.Contains(errOut, "! .claude/skills/loop-review may be loaded") {
		t.Errorf("preflight lacks the collision warning:\n%s", errOut)
	}
}

func TestDoctorAllClaudeSaysNothingOfCopilot(t *testing.T) {
	e := newDoctorEnv(t)
	e.skill("loop-review")
	out, code := e.run("doctor")
	if code != 0 || strings.Contains(strings.ToLower(out), "copilot") || strings.Contains(out, "skills/") {
		t.Errorf("doctor exit %d:\n%s", code, out)
	}
	errOut, err := e.preflight()
	if err != nil || strings.Contains(strings.ToLower(errOut), "copilot") || strings.Contains(errOut, "skills/") {
		t.Errorf("preflight: %v\n%s", err, errOut)
	}
}

const loginMsg = "no GitHub login reaches Copilot sessions: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login"

func TestDoctorCopilotLogin(t *testing.T) {
	const secret = "s3cr3t-token"
	cases := []struct {
		name string
		gh   string // body of the stub gh; "" installs none
		env  map[string]string
		ok   bool
	}{
		{"nothing", "", nil, false},
		{"gh exits 1", "exit 1", nil, false},
		{"gh prints nothing", "exit 0", nil, false},
		{"gh prints a token", "echo " + secret, nil, true},
		{"gh prints a token and fails", "echo " + secret + "; exit 1", nil, false},
		{"COPILOT_GITHUB_TOKEN", "exit 1", map[string]string{"COPILOT_GITHUB_TOKEN": secret}, true},
		{"GH_TOKEN", "exit 1", map[string]string{"GH_TOKEN": secret}, true},
		{"GITHUB_TOKEN", "exit 1", map[string]string{"GITHUB_TOKEN": secret}, true},
		{"empty variable", "exit 1", map[string]string{"GH_TOKEN": ""}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := newDoctorEnv(t)
			e.write(".vloop/config.toml", doctorCheckConfig+"[agent]\nreview = \"copilot\"\n[model]\nreview = \"gpt-x\"\n")
			e.copilot(true)
			os.Remove(filepath.Join(e.stub, "gh"))
			if tc.gh != "" {
				e.gh(tc.gh)
			}
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			c, code := e.check("copilot login")
			out, _ := e.run("doctor")
			jsonOut, _ := e.run("doctor", "--json")
			errOut, err := e.preflight()
			for _, o := range []string{out, jsonOut, errOut} {
				if strings.Contains(o, secret) {
					t.Errorf("the token is in the output:\n%s", o)
				}
			}
			if tc.ok {
				if c.Result != resPass || code != 0 || !strings.Contains(out, "✓ copilot login\n") || err != nil {
					t.Errorf("login = %s %q, exit %d, preflight %v", c.Result, c.Message, code, err)
				}
				return
			}
			if c.Result != resProblem || c.Message != loginMsg || code != 1 || !strings.Contains(out, "✗ copilot login "+loginMsg+"\n") {
				t.Errorf("login = %s %q, exit %d\n%s", c.Result, c.Message, code, out)
			}
			want := "agent.review is copilot but no GitHub login reaches its sessions: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login"
			if err == nil || err.Error() != want {
				t.Errorf("preflight error = %v, want %q", err, want)
			}
		})
	}
}

func TestDoctorClaudeReviewHasNoCopilotLogin(t *testing.T) {
	e := newDoctorEnv(t)
	e.gh("exit 1")
	if out, _ := e.run("doctor"); strings.Contains(out, "copilot login") {
		t.Errorf("doctor shows a copilot login row:\n%s", out)
	}
}
