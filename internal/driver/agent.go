package driver

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/mvelosop/vloop/internal/config"
)

// AgentClaude is the agent every phase resolves to unless configured otherwise.
const AgentClaude = "claude"

// Cost sources of a session record: the agent reported a cost, the driver
// estimated it from credits, or the agent printed none.
const (
	CostReported  = "reported"
	CostEstimated = "estimated"
	CostNone      = "none"
)

// Adapter is everything agent-specific about starting a session and reading
// what it left: which phases the agent supports and whether it can run at all,
// the work to prepare once per session, the command and environment, whether
// stdout is a record, and how it maps onto one. Everything else — gates,
// guards, the lock, budgets, masking, stderr capture, timeouts, record
// numbering — is the driver's, whatever the agent.
type Adapter interface {
	Name() string
	// Supports reports whether the agent can run a session of the phase.
	Supports(phase string) bool
	// Preflight reports whether the agent's command is on PATH and runs; when
	// it does not, problem says why.
	Preflight() (found bool, problem string)
	// Prepare readies what one session of the phase needs (extracted plugin and
	// fence folders, say) and returns it for Command.
	Prepare(root, version, phase string) (Prepared, error)
	// Command is the process of a session: argv, working directory, environment.
	Command(s Spec, root string, env []string, p Prepared) *exec.Cmd
	// Parse decides whether stdout is a record and maps it onto one; mask is
	// the driver's, for text that reaches the record.
	Parse(stdout []byte, mask func(string) string) (Output, bool)
}

// Prepared is what an adapter's Prepare returns for its own Command.
type Prepared struct{ Fence, Plugin, Home, Guidelines string }

// Output is a session's result as the agent reported it.
type Output struct {
	Cost       float64
	Credits    *float64 // Copilot credits; nil for an agent that reports none
	CostSource string
	AgentMS    int64 // the agent's own clock for the session
	Turns      int
	IsError    bool
	Denials    []json.RawMessage
	ModelsUsed map[string]usage
}

// withSession gives an adapter what only the session knows and its Parse needs:
// the model, and the price of a Copilot credit.
func (r *Runner) withSession(ad Adapter, s Spec) (Adapter, error) {
	c, ok := ad.(copilotAdapter)
	if !ok {
		return ad, nil
	}
	got, err := config.Get(r.Root, "copilot.credit-usd")
	if err != nil {
		return nil, err
	}
	price, err := strconv.ParseFloat(got.Value, 64)
	if err != nil {
		return nil, fmt.Errorf("copilot.credit-usd %q: %w", got.Value, err)
	}
	c.model, c.creditUSD = s.Model, price
	return c, nil
}

// adapterFor is the adapter of an agent name, claude run as the given command
// ("claude" when empty); copilot is run as "copilot".
func adapterFor(name, bin string) (Adapter, error) {
	if name == "" || name == AgentClaude {
		return claudeAdapter{bin: bin}, nil
	}
	if name == AgentCopilot {
		return copilotAdapter{}, nil
	}
	return nil, fmt.Errorf("unknown agent %q", name)
}

// AgentPreflight is the preflight of an agent: whether it is found and what is
// wrong with it.
func AgentPreflight(name string) (found bool, problem string) {
	a, err := adapterFor(name, "")
	if err != nil {
		return false, err.Error()
	}
	return a.Preflight()
}

type claudeAdapter struct{ bin string }

func (claudeAdapter) Name() string { return AgentClaude }

func (claudeAdapter) Supports(phase string) bool {
	switch phase {
	case PhasePlan, PhaseWork, PhaseReview, PhaseGateReview, PhaseVerify:
		return true
	}
	return false
}

func (claudeAdapter) Preflight() (bool, string) {
	if _, err := exec.LookPath("claude"); err != nil {
		return false, "claude is not on PATH"
	}
	if err := exec.Command("claude", "--version").Run(); err != nil {
		return true, "claude --version failed: " + err.Error()
	}
	return true, ""
}

func (claudeAdapter) Prepare(root, version, phase string) (Prepared, error) {
	plugin, err := ExtractPlugin(root, version)
	if err != nil {
		return Prepared{}, err
	}
	fence, err := ExtractFence(root, version, phase)
	if err != nil {
		return Prepared{}, err
	}
	return Prepared{Fence: fence, Plugin: plugin}, nil
}

func (a claudeAdapter) Command(s Spec, root string, env []string, p Prepared) *exec.Cmd {
	bin := a.bin
	if bin == "" {
		bin = "claude"
	}
	cmd := exec.Command(bin, WithGuidelines(Args(s, p.Fence, p.Plugin), root)...)
	cmd.Dir = root
	cmd.Env = dropEnv(env, "VLOOP_ACTIVE_TASK", "VLOOP_GATE_TASK", PlanHashEnv)
	if s.PlanSHA != "" {
		cmd.Env = append(cmd.Env, PlanHashEnv+"="+s.PlanSHA)
	}
	return cmd
}

// Parse maps claude's one JSON document onto an Output; anything else is not a
// record. Fields claude did not print are zero values of the same type.
func (claudeAdapter) Parse(stdout []byte, mask func(string) string) (Output, bool) {
	var raw map[string]json.RawMessage
	if len(strings.TrimSpace(string(stdout))) == 0 || json.Unmarshal(stdout, &raw) != nil {
		return Output{}, false
	}
	var o Output
	get := func(k string, v any) { _ = json.Unmarshal(raw[k], v) }
	var dur, turns float64
	get("total_cost_usd", &o.Cost)
	get("duration_ms", &dur)
	get("num_turns", &turns)
	get("is_error", &o.IsError)
	o.AgentMS, o.Turns = int64(dur), int(turns)
	o.CostSource = CostNone
	if _, ok := raw["total_cost_usd"]; ok {
		o.CostSource = CostReported
	}
	var denials []struct {
		Tool  string `json:"tool_name"`
		Input struct {
			Path     string `json:"file_path"`
			NotePath string `json:"notebook_path"`
			Command  string `json:"command"`
		} `json:"tool_input"`
	}
	get("permission_denials", &denials)
	for _, d := range denials {
		kept := map[string]string{"tool_name": d.Tool}
		if p := d.Input.Path + d.Input.NotePath; p != "" {
			kept["file_path"] = d.Input.Path
			if d.Input.Path == "" {
				kept["file_path"] = d.Input.NotePath
			}
		}
		if d.Tool == "Bash" && d.Input.Command != "" {
			kept["command"] = cutCommand(mask, d.Input.Command)
		}
		b, _ := json.Marshal(kept)
		o.Denials = append(o.Denials, b)
	}
	if o.Denials == nil {
		o.Denials = []json.RawMessage{}
	}

	var mu map[string]struct {
		Input         int64   `json:"inputTokens"`
		Output        int64   `json:"outputTokens"`
		CacheRead     int64   `json:"cacheReadInputTokens"`
		CacheCreation int64   `json:"cacheCreationInputTokens"`
		Cost          float64 `json:"costUSD"`
	}
	get("modelUsage", &mu)
	o.ModelsUsed = map[string]usage{}
	names := make([]string, 0, len(mu))
	for k := range mu {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		u := mu[k]
		o.ModelsUsed[k] = usage{u.Input, u.Output, u.CacheRead, u.CacheCreation, u.Cost}
	}
	return o, true
}
