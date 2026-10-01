package driver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Session kinds.
const (
	PhasePlan   = "plan"
	PhaseWork   = "work"
	PhaseReview = "review"
)

// Runner starts the sessions of one run. The zero value of the optional fields
// is the real thing: `claude` on PATH, the real clock, the real home.
type Runner struct {
	Root    string // repository root; sessions start here
	Version string // the binary's version, naming the plugin and fence folders
	RunID   string
	RunDir  string    // the run folder (absolute); records go in its sessions/
	Log     io.Writer // run log lines, masked; may be nil

	Claude string           // command to run; "claude" when empty
	Now    func() time.Time // the driver's clock; time.Now when nil
	Home   string           // masked to "~"; os.UserHomeDir when empty
	User   string           // masked to "USER"; the home's base name when empty
	Env    []string         // base environment; os.Environ when nil

	n int // sessions started in this run folder
}

// Spec is one session to start. Arg is the brief path for a plan and the task
// id for work and review. Effort is "" when unset.
type Spec struct {
	Phase     string
	Iteration int
	Arg       string
	Model     string
	Effort    string
}

// Result is how a session ended. Recorded is false when it printed nothing and
// so left no record.
type Result struct {
	ExitCode int
	Recorded bool
	Path     string // the record file, absolute
	Cost     float64
	Turns    int
	IsError  bool
	Denials  int
}

// Prompt is the slash command a session of the phase starts with.
func Prompt(phase, arg string) string {
	return "/vloop:" + phase + " " + arg
}

// Args is the claude argv for a session, with the fence and plugin paths as
// given.
func Args(s Spec, fence, plugin string) []string {
	a := []string{"-p", Prompt(s.Phase, s.Arg), "--model", s.Model}
	if s.Effort != "" {
		a = append(a, "--effort", s.Effort)
	}
	return append(a,
		"--permission-mode", "auto",
		"--setting-sources", "project",
		"--settings", fence,
		"--plugin-dir", plugin,
		"--strict-mcp-config",
		"--output-format", "json")
}

func (r *Runner) now() time.Time {
	if r.Now != nil {
		return r.Now()
	}
	return time.Now()
}

func (r *Runner) home() (home, user string) {
	home = r.Home
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	user = r.User
	if user == "" && home != "" {
		user = filepath.Base(home)
	}
	return home, user
}

// Mask replaces $HOME with ~ and the user name with USER, as the shell
// driver's mask() does. A path is masked in its JSON-escaped form too.
func (r *Runner) Mask(s string) string {
	home, user := r.home()
	if home != "" && home != "/" {
		s = strings.ReplaceAll(s, home, "~")
		if esc := strings.ReplaceAll(home, `\`, `\\`); esc != home {
			s = strings.ReplaceAll(s, esc, "~")
		}
	}
	if user != "" {
		s = strings.ReplaceAll(s, user, "USER")
	}
	return s
}

// Logf writes one masked line to the run log.
func (r *Runner) Logf(format string, a ...any) {
	if r.Log != nil {
		fmt.Fprintln(r.Log, r.Mask(fmt.Sprintf(format, a...)))
	}
}

// next returns the next per-run-folder session number, continuing after any
// record already in the folder so a resumed run never overwrites one.
func (r *Runner) next() int {
	if r.n == 0 {
		ents, _ := os.ReadDir(filepath.Join(r.RunDir, "sessions"))
		for _, e := range ents {
			if k, err := strconv.Atoi(strings.SplitN(e.Name(), "-", 2)[0]); err == nil && k > r.n {
				r.n = k
			}
		}
	}
	r.n++
	return r.n
}

// Run starts one session and, if it printed JSON, writes its session/v1 record.
// A non-zero exit of claude is a Result, not an error; an error means the
// session could not be started or its files written.
func (r *Runner) Run(s Spec) (Result, error) {
	plugin, err := ExtractPlugin(r.Root, r.Version)
	if err != nil {
		return Result{}, err
	}
	fence, err := ExtractFence(r.Root, r.Version)
	if err != nil {
		return Result{}, err
	}
	task := ""
	if s.Phase != PhasePlan {
		task = s.Arg
	}
	if err := os.MkdirAll(filepath.Join(r.RunDir, "sessions"), 0o755); err != nil {
		return Result{}, err
	}
	n := r.next()

	bin := r.Claude
	if bin == "" {
		bin = "claude"
	}
	cmd := exec.Command(bin, Args(s, fence, plugin)...)
	cmd.Dir = r.Root
	base := r.Env
	if base == nil {
		base = os.Environ()
	}
	for _, kv := range base {
		if strings.HasPrefix(kv, "VLOOP_ACTIVE_TASK=") || strings.HasPrefix(kv, "VLOOP_GATE_TASK=") {
			continue
		}
		cmd.Env = append(cmd.Env, kv)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	started := r.now().UTC()
	runErr := cmd.Run()
	res := Result{}
	if runErr != nil {
		ee, ok := runErr.(*exec.ExitError)
		if !ok {
			return Result{}, fmt.Errorf("cannot start %s: %w", bin, runErr)
		}
		res.ExitCode = ee.ExitCode()
	}
	errName := filepath.Join(r.RunDir, fmt.Sprintf("%s-%d.stderr", s.Phase, s.Iteration))
	if err := os.WriteFile(errName, []byte(r.Mask(stderr.String())), 0o644); err != nil {
		return res, err
	}

	var raw map[string]json.RawMessage
	if len(bytes.TrimSpace(stdout.Bytes())) == 0 || json.Unmarshal(stdout.Bytes(), &raw) != nil {
		r.Logf("SESSION RECORD MISSING %s %s (iteration %d)", s.Phase, orDash(task), s.Iteration)
		return res, nil
	}
	rec, out := r.record(s, task, started, raw)
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return res, err
	}
	res.Path = filepath.Join(r.RunDir, "sessions", fmt.Sprintf("%03d-%s.json", n, s.Phase))
	if err := os.WriteFile(res.Path, []byte(r.Mask(string(data))+"\n"), 0o644); err != nil {
		return res, err
	}
	res.Recorded = true
	res.Cost, res.Turns, res.IsError, res.Denials = out.Cost, out.Turns, out.IsError, len(out.Denials)
	return res, nil
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

type usage struct {
	Input         int64   `json:"input_tokens"`
	Output        int64   `json:"output_tokens"`
	CacheRead     int64   `json:"cache_read_input_tokens"`
	CacheCreation int64   `json:"cache_creation_input_tokens"`
	Cost          float64 `json:"cost_usd"`
}

type output struct {
	Cost    float64
	Turns   int
	IsError bool
	Denials []json.RawMessage
}

// record maps claude's JSON onto session/v1. Fields claude did not print are
// zero values of the same type; none is made up beyond that.
func (r *Runner) record(s Spec, task string, started time.Time, raw map[string]json.RawMessage) (map[string]any, output) {
	var o output
	get := func(k string, v any) { _ = json.Unmarshal(raw[k], v) }
	var dur float64
	var turns float64
	get("total_cost_usd", &o.Cost)
	get("duration_ms", &dur)
	get("num_turns", &turns)
	get("is_error", &o.IsError)
	get("permission_denials", &o.Denials)
	o.Turns = int(turns)
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
	used := map[string]usage{}
	names := make([]string, 0, len(mu))
	for k := range mu {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		u := mu[k]
		used[k] = usage{u.Input, u.Output, u.CacheRead, u.CacheCreation, u.Cost}
	}

	var effort any
	if s.Effort != "" {
		effort = s.Effort
	}
	rec := map[string]any{
		"schema":             "session/v1",
		"run_id":             r.RunID,
		"iteration":          s.Iteration,
		"phase":              s.Phase,
		"model":              s.Model,
		"effort":             effort,
		"models_used":        used,
		"started":            started.Format(time.RFC3339),
		"duration_ms":        int64(dur),
		"cost_usd":           o.Cost,
		"turns":              o.Turns,
		"is_error":           o.IsError,
		"permission_denials": o.Denials,
	}
	if task != "" {
		rec["task"] = task
	}
	return rec, o
}
