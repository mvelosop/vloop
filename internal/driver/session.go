package driver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mvelosop/vloop/internal/state"
)

// Session kinds.
const (
	PhasePlan       = "plan"
	PhaseWork       = "work"
	PhaseReview     = "review"
	PhaseGateReview = "gate-review"
	PhaseVerify     = "verify"
)

// Runner starts the sessions of one run. The zero value of the optional fields
// is the real thing: `claude` on PATH, the real clock, the real home.
type Runner struct {
	Root    string // repository root; sessions start here
	Version string // the binary's version, naming the plugin and fence folders
	RunID   string
	RunDir  string    // the run folder (absolute); records go in its sessions/
	Log     io.Writer // run log lines, masked; may be nil

	Claude  string           // the claude adapter's command; "claude" when empty
	Now     func() time.Time // the driver's clock; time.Now when nil
	Home    string           // masked to "~"; os.UserHomeDir when empty
	User    string           // masked to "USER"; the home's base name when empty
	TempDir string           // masked to "$TMPDIR"; os.TempDir when empty
	Env     []string         // base environment; os.Environ when nil

	Timeout time.Duration // a session still running after this is killed with its group; none when zero

	n int // sessions started in this run folder
}

// Spec is one session to start. Arg is the brief path for a plan and the task
// id for work and review; a gate review has none. Effort is "" when unset.
type Spec struct {
	Phase     string
	Iteration int
	Arg       string
	Model     string
	Effort    string
	Agent     string // the phase's resolved agent; claude when empty
	PlanSHA   string // the plan's hash handed to the session; "" for none
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
	TimedOut bool          // the session was killed for running past the timeout
	Wall     time.Duration // how long the session ran, by the driver's clock
}

// Prompt is the slash command a session of the phase starts with.
func Prompt(phase, arg string) string {
	if arg == "" {
		return "/vloop:" + phase
	}
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

// GuidelinesFile is the repository's lessons, handed to every session.
const GuidelinesFile = ".vloop/repo-guidelines.md"

// WithGuidelines adds the guidelines file to a session's arguments as one
// contiguous pair when the repository root has it and it is not empty; the
// arguments are returned as given otherwise. The root only, never HOME.
func WithGuidelines(args []string, root string) []string {
	if fi, err := os.Stat(filepath.Join(root, GuidelinesFile)); err != nil || !fi.Mode().IsRegular() || fi.Size() == 0 {
		return args
	}
	return append(args, "--append-system-prompt-file", GuidelinesFile)
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
		h := trimSeps(home)
		user = h[strings.LastIndexAny(h, `/\`)+1:]
	}
	return home, user
}

// pathEnd is what may follow a masked path: the end, a separator, a quote,
// white space or a list delimiter. A letter, digit, dot or dash continues a
// name, so it is not the end of the path.
const pathEnd = `(\z|[/\\"'\s:,;)\]])`

const pathSep = `[/\\]{1,2}`

// windowsPath reports whether p is a drive or UNC path, whose separators and
// letter case are not significant.
func windowsPath(p string) bool {
	drive := len(p) >= 3 && p[1] == ':' && (p[2] == '/' || p[2] == '\\') && (p[0]|0x20 >= 'a' && p[0]|0x20 <= 'z')
	return drive || strings.HasPrefix(p, `\\`)
}

func trimSeps(p string) string { return strings.TrimRight(p, `/\`) }

// homePattern matches the home in the forms it can take: as given, JSON
// escaped, and for a Windows home with either separator in any letter case.
func homePattern(home string) *regexp.Regexp {
	var body string
	if windowsPath(home) {
		var parts []string
		for _, c := range regexp.MustCompile(`[/\\]+`).Split(home, -1) {
			parts = append(parts, regexp.QuoteMeta(c))
		}
		return regexp.MustCompile(`(?i)(?:` + strings.Join(parts, pathSep) + `)` + pathEnd)
	} else {
		body = regexp.QuoteMeta(home)
		if esc := strings.ReplaceAll(home, `\`, `\\`); esc != home {
			body += `|` + regexp.QuoteMeta(esc)
		}
	}
	return regexp.MustCompile(`(?:` + body + `)` + pathEnd)
}

// secretName matches the environment variables whose values are redacted.
var secretName = regexp.MustCompile(`(?i)KEY|TOKEN|SECRET|PASSWORD|CREDENTIAL`)

// minSecret is the shortest value that is redacted; shorter ones are too
// likely to be ordinary words.
const minSecret = 8

// redactSecrets replaces every occurrence of the value of a secret-named
// environment variable with <redacted:NAME>, longest values first. It keys on
// names, not on what a value looks like.
func redactSecrets(env []string, s string) string {
	type secret struct{ name, value string }
	var secrets []secret
	for _, kv := range env {
		name, value, ok := strings.Cut(kv, "=")
		if !ok || len(value) < minSecret || !secretName.MatchString(name) {
			continue
		}
		secrets = append(secrets, secret{name, value})
	}
	sort.SliceStable(secrets, func(i, j int) bool { return len(secrets[i].value) > len(secrets[j].value) })
	for _, c := range secrets {
		s = strings.ReplaceAll(s, c.value, "<redacted:"+c.name+">")
		// A JSON document holds the value escaped.
		if q, err := json.Marshal(c.value); err == nil {
			if esc := string(q[1 : len(q)-1]); esc != c.value {
				s = strings.ReplaceAll(s, esc, "<redacted:"+c.name+">")
			}
		}
	}
	return s
}

func (r *Runner) env() []string {
	if r.Env != nil {
		return r.Env
	}
	return os.Environ()
}

// keptTemp are the shared temporary directories that briefs and gates name;
// they are never masked, whether or not the system temp directory is one.
var keptTemp = []string{"/tmp", "/var/tmp", "/private/tmp", "/private/var/tmp"}

// tempDirs is the system temp directory as given and, when it is a symbolic
// link, resolved — longest first, without /tmp and /var/tmp.
func (r *Runner) tempDirs() []string {
	t := r.TempDir
	if t == "" {
		t = os.TempDir()
	}
	var dirs []string
	for _, d := range []string{t, ""} {
		if d == "" {
			if res, err := filepath.EvalSymlinks(t); err == nil {
				d = res
			}
		}
		d = trimSeps(d)
		if d == "" || slices.Contains(keptTemp, d) || slices.Contains(dirs, d) {
			continue
		}
		dirs = append(dirs, d)
	}
	sort.SliceStable(dirs, func(i, j int) bool { return len(dirs[i]) > len(dirs[j]) })
	return dirs
}

// Mask redacts the values of secret-named environment variables, replaces the
// home with ~ where it ends at a path boundary, then the system temp directory
// with $TMPDIR, and the user
// name with USER where it is a whole path component directly under the users
// directory (Users, home, or the one that holds the home) — never as free
// text. It works on text; a JSON document is masked through its string values
// (maskValue), never as serialized text.
func (r *Runner) Mask(s string) string {
	s = redactSecrets(r.env(), s)
	home, user := r.home()
	if home = trimSeps(home); home != "" {
		s = homePattern(home).ReplaceAllString(s, "~${1}")
	}
	for _, d := range r.tempDirs() {
		s = homePattern(d).ReplaceAllString(s, "$$TMPDIR${1}")
	}
	if user = trimSeps(user); user != "" {
		dirs := []string{"Users", "home"}
		if i := strings.LastIndexAny(trimSeps(home), `/\`); i > 0 {
			if parent := trimSeps(home[:i]); parent != "" {
				dirs = append(dirs, filepath.Base(strings.ReplaceAll(parent, `\`, "/")))
			}
		}
		for i, d := range dirs {
			dirs[i] = regexp.QuoteMeta(d)
		}
		flags := ""
		if windowsPath(home) {
			flags = `(?i)`
		}
		re := regexp.MustCompile(flags + `(` + pathSep + `(?:` + strings.Join(dirs, "|") + `)` + pathSep + `)` + regexp.QuoteMeta(user) + pathEnd)
		s = re.ReplaceAllString(s, "${1}USER${2}")
		// Claude Code's project folders spell a path with dashes.
		dashed := regexp.MustCompile(`(-(?:Users|home)-)` + regexp.QuoteMeta(user) + `-`)
		s = dashed.ReplaceAllString(s, "${1}USER-")
	}
	return s
}

// maxCommand is how many characters of a denied Bash command a record keeps.
const maxCommand = 200

// cutCommand masks a command, then keeps its first maxCommand characters, with
// … appended when it was cut. The masking comes first so the cut never splits
// a secret or a path into something the masks no longer recognize.
func cutCommand(mask func(string) string, c string) string {
	c = mask(c)
	if n := 0; len(c) > maxCommand {
		for i := range c {
			if n == maxCommand {
				return c[:i] + "…"
			}
			n++
		}
	}
	return c
}

func (r *Runner) cutCommand(c string) string { return cutCommand(r.Mask, c) }

// maskValue masks every string, and every object key, of a decoded JSON value.
func (r *Runner) maskValue(v any) any {
	switch x := v.(type) {
	case string:
		return r.Mask(x)
	case []any:
		for i := range x {
			x[i] = r.maskValue(x[i])
		}
		return x
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, e := range x {
			out[r.Mask(k)] = r.maskValue(e)
		}
		return out
	}
	return v
}

// maskRecord masks the strings of a record and returns it as indented JSON.
// Numbers pass through untouched.
func (r *Runner) maskRecord(rec map[string]any) ([]byte, error) {
	raw, err := json.Marshal(rec)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	return json.MarshalIndent(r.maskValue(v), "", "  ")
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

// Run starts one session and, if it printed JSON, writes its session/v3 record.
// A non-zero exit of claude is a Result, not an error; an error means the
// session could not be started or its files written.
func (r *Runner) Run(s Spec) (Result, error) {
	ad, err := adapterFor(s.Agent, r.Claude)
	if err != nil {
		return Result{}, err
	}
	if !ad.Supports(s.Phase) {
		return Result{}, fmt.Errorf("agent %s does not support the %s phase", ad.Name(), s.Phase)
	}
	if ad, err = r.withSession(ad, s); err != nil {
		return Result{}, err
	}
	prep, err := ad.Prepare(r.Root, r.Version, s.Phase)
	if err != nil {
		return Result{}, err
	}
	task := ""
	if s.Phase == PhaseWork || s.Phase == PhaseReview {
		task = s.Arg
	}
	if err := os.MkdirAll(filepath.Join(r.RunDir, "sessions"), 0o755); err != nil {
		return Result{}, err
	}
	n := r.next()

	cmd := ad.Command(s, r.Root, r.env(), prep)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	// Wall clock, not monotonic: the measurement includes any time the machine slept.
	started := r.now().Round(0).UTC()
	spent := spendIn(r.RunDir)
	updateLock(r.Root, func(l *lockRecord) {
		l.Phase, l.Task, l.SessionStarted, l.Spent = s.Phase, task, started.Format(time.RFC3339), spent
	})
	timedOut, runErr := state.RunGroup(cmd, r.Timeout)
	ended := r.now().Round(0).UTC()
	// However it ended, the session is over: record the spend, clear the session.
	defer func() {
		spent := spendIn(r.RunDir)
		updateLock(r.Root, func(l *lockRecord) { l.Phase, l.Task, l.SessionStarted, l.Spent = "", "", "", spent })
	}()
	res := Result{TimedOut: timedOut, Wall: max(ended.Sub(started), 0)}
	if runErr != nil {
		ee, ok := runErr.(*exec.ExitError)
		if !ok {
			return Result{}, fmt.Errorf("cannot start %s: %w", cmd.Path, runErr)
		}
		res.ExitCode = ee.ExitCode()
	}
	if timedOut {
		if res.ExitCode == 0 {
			res.ExitCode = -1
		}
		r.Logf("SESSION TIMED OUT %s %s (iteration %d) after %s", s.Phase, orDash(task), s.Iteration, r.Timeout)
	}
	errName := filepath.Join(r.RunDir, fmt.Sprintf("%s-%d.stderr", s.Phase, s.Iteration))
	if err := os.WriteFile(errName, []byte(r.Mask(stderr.String())), 0o644); err != nil {
		return res, err
	}

	out, ok := ad.Parse(stdout.Bytes(), r.Mask)
	if !ok {
		r.Logf("SESSION RECORD MISSING %s %s (iteration %d)", s.Phase, orDash(task), s.Iteration)
		return res, nil
	}
	rec := r.record(s, ad.Name(), task, started, ended, out)
	data, err := r.maskRecord(rec)
	if err != nil {
		return res, err
	}
	res.Path = filepath.Join(r.RunDir, "sessions", fmt.Sprintf("%03d-%s.json", n, s.Phase))
	if err := os.WriteFile(res.Path, append(data, '\n'), 0o644); err != nil {
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

// record maps an agent's output onto session/v3. Fields the agent did not
// report are zero values of the same type; none is made up beyond that.
func (r *Runner) record(s Spec, agent, task string, started, ended time.Time, o Output) map[string]any {
	var effort any
	if s.Effort != "" {
		effort = s.Effort
	}
	rec := map[string]any{
		"schema":             "session/v3",
		"run_id":             r.RunID,
		"iteration":          s.Iteration,
		"phase":              s.Phase,
		"agent":              agent,
		"model":              s.Model,
		"effort":             effort,
		"models_used":        o.ModelsUsed,
		"started":            started.Format(time.RFC3339),
		"ended":              ended.Format(time.RFC3339),
		"duration_ms":        o.AgentMS,
		"wall_ms":            max(ended.Sub(started).Milliseconds(), 0),
		"cost_usd":           o.Cost,
		"cost_source":        o.CostSource,
		"turns":              o.Turns,
		"is_error":           o.IsError,
		"permission_denials": o.Denials,
	}
	if o.Credits != nil {
		rec["credits"] = *o.Credits
	}
	if task != "" {
		rec["task"] = task
	}
	return rec
}
