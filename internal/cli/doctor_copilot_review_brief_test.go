package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCopilotReviewSkillNamedReviewBriefClashes(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", doctorCheckConfig+"[agent]\nreview = \"copilot\"\n[model]\nreview = \"gpt-x\"\n")
	e.copilot(true)
	e.skill("review-brief")
	out, code := e.run("doctor")
	want := "✗ .claude/skills/review-brief has the name of vloop's review-brief skill; a copilot session may load it instead — rename it"
	if code != 1 || !strings.Contains(out, want) {
		t.Errorf("doctor exit %d, want line %q:\n%s", code, want, out)
	}
}

const copilotReviewBriefConfig = doctorCheckConfig + "[agent]\nreview-brief = \"copilot\"\n"

func TestDoctorCopilotReviewBriefRows(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", copilotReviewBriefConfig+"[model]\nreview-brief = \"gpt-x\"\n")
	e.copilot(true)
	e.skill("review-brief")
	out, code := e.run("doctor")
	for _, want := range []string{
		"! agent.review-brief is copilot: experimental (docs/guide/configuration.md)\n",
		"✓ copilot\n", "✓ copilot login\n",
		"✗ .claude/skills/review-brief has the name of vloop's review-brief skill; a copilot session may load it instead — rename it\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("doctor exit %d lacks %q:\n%s", code, want, out)
		}
	}
}

func TestRunPreflightIgnoresReviewBriefCopilot(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", copilotReviewBriefConfig) // default model, no copilot, a clash
	e.skill("review-brief")
	errOut, err := e.preflight()
	if err != nil || strings.Contains(errOut, "copilot") || strings.Contains(errOut, "review-brief") {
		t.Errorf("preflight = %v:\n%s", err, errOut)
	}
}

func briefReviewPreflightErr(e *doctorEnv) (string, error) {
	var errOut bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetErr(&errOut)
	err := briefReviewPreflight(e.repo, cmd)
	return errOut.String(), err
}

func TestBriefReviewCopilotPreflight(t *testing.T) {
	good := copilotReviewBriefConfig + "[model]\nreview-brief = \"gpt-x\"\n"
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", copilotReviewBriefConfig)
	e.copilot(true)
	_, err := briefReviewPreflightErr(e)
	if err == nil || err.Error() != `agent.review-brief is copilot but model.review-brief is the default "opus", a Claude alias: set model.review-brief to a Copilot model id` {
		t.Errorf("default model: %v", err)
	}
	e.write(".vloop/config.toml", good)
	e.gh("exit 1")
	for _, k := range []string{"COPILOT_GITHUB_TOKEN", "GH_TOKEN", "GITHUB_TOKEN"} {
		t.Setenv(k, "")
	}
	if _, err = briefReviewPreflightErr(e); err == nil || !strings.Contains(err.Error(), "agent.review-brief is copilot but no GitHub login reaches its sessions") {
		t.Errorf("no login: %v", err)
	}
	e.copilot(true)
	e.skill("review-brief")
	if _, err = briefReviewPreflightErr(e); err == nil || err.Error() != ".claude/skills/review-brief has the name of vloop's review-brief skill; a copilot session may load it instead — rename it" {
		t.Errorf("clash: %v", err)
	}
	if err := os.RemoveAll(filepath.Join(e.repo, ".claude", "skills")); err != nil {
		t.Fatal(err)
	}
	errOut, err := briefReviewPreflightErr(e)
	if err != nil || !strings.Contains(errOut, "  ! agent.review-brief is copilot: experimental (docs/guide/configuration.md)\n") {
		t.Errorf("good: %v\n%s", err, errOut)
	}
	if err := os.Remove(filepath.Join(e.stub, "copilot")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", e.stub)
	errOut, err = briefReviewPreflightErr(e)
	if err == nil || err.Error() != "preflight failed — 1 problem(s), nothing has run" || !strings.Contains(errOut, "  ✗ copilot ") {
		t.Errorf("missing: %v\n%s", err, errOut)
	}
}
