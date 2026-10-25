package driver

import (
	"os"
	"os/exec"
	"path/filepath"
)

// AgentCopilot is the GitHub Copilot CLI, which can run review sessions only.
const AgentCopilot = "copilot"

// copilotDenials are the stems a copilot session may not run: Copilot's
// Kind(argument) patterns, shell(<stem>:*) matching the stem alone or followed
// by a space. It is a denylist until the fence of B20 replaces it.
var copilotDenials = []string{
	"shell(gh:*)",
	"shell(git push:*)", "shell(git remote:*)",
	"shell(git config --global:*)", "shell(git config --system:*)", "shell(git config --file:*)",
	"shell(curl:*)", "shell(wget:*)", "shell(ssh:*)", "shell(scp:*)", "shell(nc:*)",
	"shell(rm:*)",
	"shell(git commit:*)", "shell(git reset:*)", "shell(git checkout:*)", "shell(git switch:*)",
	"shell(git stash:*)", "shell(git branch:*)", "shell(git tag:*)",
	"url",
}

// copilotGuidelinesName is the file Copilot reads from a custom instructions
// directory.
const copilotGuidelinesName = "AGENTS.md"

type copilotAdapter struct{ bin string }

func (copilotAdapter) Name() string { return AgentCopilot }

func (copilotAdapter) Supports(phase string) bool { return phase == PhaseReview }

func (a copilotAdapter) command() string {
	if a.bin != "" {
		return a.bin
	}
	return "copilot"
}

func (a copilotAdapter) Preflight() (bool, string) {
	if _, err := exec.LookPath(a.command()); err != nil {
		return false, "copilot is not on PATH"
	}
	if err := exec.Command(a.command(), "--version").Run(); err != nil {
		return true, "copilot --version failed: " + err.Error()
	}
	return true, ""
}

// Prepare extracts the plugin and makes, for this session alone, an empty
// Copilot home and, when the repository has guidelines, the directory holding
// them; both live under .vloop/tmp/, which is emptied between sessions.
func (copilotAdapter) Prepare(root, version, phase string) (Prepared, error) {
	plugin, err := ExtractPlugin(root, version)
	if err != nil {
		return Prepared{}, err
	}
	tmp := filepath.Join(root, filepath.FromSlash(tmpDir))
	if err := os.MkdirAll(tmp, 0o755); err != nil {
		return Prepared{}, err
	}
	home, err := os.MkdirTemp(tmp, "copilot-home-")
	if err != nil {
		return Prepared{}, err
	}
	p := Prepared{Plugin: plugin, Home: home}
	src := filepath.Join(root, filepath.FromSlash(GuidelinesFile))
	if fi, err := os.Stat(src); err != nil || !fi.Mode().IsRegular() || fi.Size() == 0 {
		return p, nil
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return Prepared{}, err
	}
	dir, err := os.MkdirTemp(tmp, "copilot-guidelines-")
	if err != nil {
		return Prepared{}, err
	}
	if err := os.WriteFile(filepath.Join(dir, copilotGuidelinesName), data, 0o644); err != nil {
		return Prepared{}, err
	}
	p.Guidelines = dir
	return p, nil
}

func (a copilotAdapter) Command(s Spec, root string, env []string, p Prepared) *exec.Cmd {
	args := []string{"-p", Prompt(s.Phase, s.Arg), "--model", s.Model}
	if s.Effort != "" {
		args = append(args, "--effort", s.Effort)
	}
	args = append(args,
		"--plugin-dir", p.Plugin,
		"--no-ask-user",
		"--output-format", "json",
		"--allow-all-tools")
	for _, d := range copilotDenials {
		args = append(args, "--deny-tool="+d)
	}
	args = append(args, "--disable-builtin-mcps")
	cmd := exec.Command(a.command(), args...)
	cmd.Dir = root
	drop := []string{"VLOOP_ACTIVE_TASK", "VLOOP_GATE_TASK", PlanHashEnv, "COPILOT_HOME"}
	if p.Guidelines != "" {
		drop = append(drop, "COPILOT_CUSTOM_INSTRUCTIONS_DIRS")
	}
	cmd.Env = dropEnv(env, drop...)
	if s.PlanSHA != "" {
		cmd.Env = append(cmd.Env, PlanHashEnv+"="+s.PlanSHA)
	}
	cmd.Env = append(cmd.Env, "COPILOT_HOME="+p.Home)
	if p.Guidelines != "" {
		cmd.Env = append(cmd.Env, "COPILOT_CUSTOM_INSTRUCTIONS_DIRS="+p.Guidelines)
	}
	return cmd
}

// Parse leaves the record missing; reading Copilot's JSON Lines is a later step.
func (copilotAdapter) Parse(stdout []byte, mask func(string) string) (Output, bool) {
	return Output{}, false
}
