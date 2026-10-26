package driver

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// CopilotDenyStems are the shell stems of the deny list in the fence's Bash
// matching form (`git push:*`), for the judge behind the Copilot hook.
func CopilotDenyStems() []string {
	var out []string
	for _, d := range copilotDenials {
		if s, ok := strings.CutPrefix(d, "shell("); ok {
			out = append(out, strings.TrimSuffix(s, ")"))
		}
	}
	return out
}

// copilotGuidelinesName is the file Copilot reads from a custom instructions
// directory.
const copilotGuidelinesName = "AGENTS.md"

type copilotAdapter struct {
	bin       string
	model     string  // the session's model, the one entry of models_used
	creditUSD float64 // dollars per credit, for the estimated cost
}

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

// Parse reads Copilot's JSON Lines. There is a record exactly when a result
// event was printed. Credits are the last usage checkpoint's totalNanoAiu / 1e9
// and the cost is those credits at the configured price, so cost_source is
// estimated; with no checkpoint both are zero and the source is none. Nothing
// of the result's codeChanges is kept: it holds absolute paths.
func (a copilotAdapter) Parse(stdout []byte, mask func(string) string) (Output, bool) {
	var (
		result     bool
		o          Output
		nano       float64
		checkpoint bool
	)
	for _, line := range bytes.Split(stdout, []byte("\n")) {
		var ev struct {
			Type     string `json:"type"`
			ExitCode int    `json:"exitCode"`
			Data     struct {
				NanoAIU *float64 `json:"totalNanoAiu"`
			} `json:"data"`
			Usage struct {
				SessionDurationMS float64 `json:"sessionDurationMs"`
			} `json:"usage"`
		}
		if json.Unmarshal(bytes.TrimSpace(line), &ev) != nil {
			continue
		}
		switch ev.Type {
		case "assistant.turn_start":
			o.Turns++
		case "session.usage_checkpoint":
			if ev.Data.NanoAIU != nil {
				nano, checkpoint = *ev.Data.NanoAIU, true
			}
		case "result":
			result = true
			o.IsError = ev.ExitCode != 0
			o.AgentMS = int64(ev.Usage.SessionDurationMS)
		}
	}
	if !result {
		return Output{}, false
	}
	credits := nano / 1e9
	o.Credits = &credits
	o.CostSource = CostNone
	if checkpoint {
		o.CostSource = CostEstimated
		o.Cost = credits * a.creditUSD
	}
	o.Denials = []json.RawMessage{}
	o.ModelsUsed = map[string]usage{a.model: {Cost: o.Cost}}
	return o, true
}
