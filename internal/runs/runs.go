// Package runs reads what the two loop drivers leave behind — run folders and
// git history — into one model per brief. The shell loop's layout lives under
// .loop/, vloop's own under .vloop/. Nothing here writes to either.
package runs

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Layout names one of the two on-disk conventions.
type Layout struct {
	Name   string // "loop" or "vloop"
	Dir    string // ".loop" or ".vloop"
	Prefix string // commit subject prefix, "[loop]" or "[vloop]"
}

var (
	// ShellLoop is the shell loop's layout.
	ShellLoop = Layout{Name: "loop", Dir: ".loop", Prefix: "[loop]"}
	// Vloop is vloop's own layout.
	Vloop = Layout{Name: "vloop", Dir: ".vloop", Prefix: "[vloop]"}
	// Layouts lists both, shell loop first.
	Layouts = []Layout{ShellLoop, Vloop}
)

// StatePath is the plan's repo-relative path in this layout.
func (l Layout) StatePath() string { return l.Dir + "/state/state.json" }

func (l Layout) runsDir() string { return l.Dir + "/state/runs" }

// Tokens is one model's token use and cost in one session.
type Tokens struct {
	Input         int64
	Output        int64
	CacheRead     int64
	CacheCreation int64
	CostUSD       float64
}

// Session is one claude session's telemetry record.
type Session struct {
	Seq        int // NNN of the file name
	Phase      string
	Iteration  int
	Task       string // from the iteration the session names; empty for plan sessions
	Model      string
	Effort     string
	Started    time.Time
	DurationMS int64
	CostUSD    float64
	Denials    int
	Models     map[string]Tokens
	IsError    bool
}

// Iteration is one line of iterations.jsonl.
type Iteration struct {
	Iteration int
	Task      string
	Attempt   int
	Outcome   string // as written; see Canonical
	GateMS    *int64 // nil when the record carries no gate duration
	ChecksMS  *int64 // sum of the checks' durations; nil when the record lists none
	Flaky     bool   // the gate failed, then passed on an immediate re-run
	Started   time.Time
	Ended     time.Time
}

// Canonical maps the shell loop's outcome names to iteration/v1's.
func (i Iteration) Canonical() string {
	switch i.Outcome {
	case "gate_fail":
		return "gate_failed"
	case "review_fail":
		return "rejected"
	}
	return i.Outcome
}

// Finding is one review finding. The shell loop writes bare strings, so its
// Kind is empty.
type Finding struct {
	Summary string
	Kind    string
}

// Verdict is one review's verdict for one iteration.
type Verdict struct {
	Iteration int // NNN of the file name
	Task      string
	Verdict   string
	Findings  []Finding
}

// Folder is one run folder.
type Folder struct {
	Layout     Layout
	Path       string // repo-relative, "/" separators
	Sessions   []Session
	Iterations []Iteration
	Verdicts   []Verdict
}

// Model is everything read for one brief.
type Model struct {
	RunID     string
	BriefPath string
	Folders   []Folder
	Owned     *Owned // nil when no plan commit exists
}

// RunID is a brief's run id: its file name minus the extension and `.loop-brief`.
func RunID(brief string) string {
	b := path.Base(filepath.ToSlash(brief))
	b = strings.TrimSuffix(b, ".md")
	return strings.TrimSuffix(b, ".loop-brief")
}

// BriefPath resolves a brief given as a path or a bare name to the
// repo-relative path the loop logs. A name is looked up in docs/briefs.
func BriefPath(root, brief string) string {
	b := filepath.ToSlash(brief)
	if filepath.IsAbs(brief) {
		if rel, err := filepath.Rel(root, brief); err == nil {
			b = filepath.ToSlash(rel)
		}
	}
	b = strings.TrimPrefix(b, "./")
	if strings.Contains(b, "/") || strings.HasSuffix(b, ".md") {
		return b
	}
	return "docs/briefs/" + strings.TrimSuffix(b, ".loop-brief") + ".loop-brief.md"
}

// Read collects the run folders and the owned commits of a brief.
func Read(root, brief string) (*Model, error) {
	m, err := Load(root, brief)
	if err != nil {
		return nil, err
	}
	owned, err := OwnedCommits(root, m.RunID)
	if err != nil {
		return nil, err
	}
	m.Owned = owned
	return m, nil
}

// Load collects every run folder that belongs to the brief, in both layouts.
// Folders are ordered shell loop first, then by path.
func Load(root, brief string) (*Model, error) {
	m := &Model{RunID: RunID(brief), BriefPath: BriefPath(root, brief)}
	shell, err := shellFolders(root, m)
	if err != nil {
		return nil, err
	}
	vl, err := vloopFolders(root, m)
	if err != nil {
		return nil, err
	}
	sortFolders(shell)
	sortFolders(vl)
	m.Folders = append(shell, vl...)
	return m, nil
}

// start is the earliest timestamp among a folder's iteration and session
// records; ok is false for a folder with none.
func (f Folder) start() (t time.Time, ok bool) {
	note := func(c time.Time) {
		if !c.IsZero() && (!ok || c.Before(t)) {
			t, ok = c, true
		}
	}
	for _, it := range f.Iterations {
		note(it.Started)
	}
	for _, s := range f.Sessions {
		note(s.Started)
	}
	return t, ok
}

// sortFolders orders folders by the earliest timestamp of their records, which
// are UTC, rather than by name: older folders are named in local time, so a
// DST fall-back sorts a later run first. A folder with no timestamped record
// is ordered by name against the others.
func sortFolders(fs []Folder) {
	sort.SliceStable(fs, func(a, b int) bool {
		ta, oka := fs[a].start()
		tb, okb := fs[b].start()
		if oka && okb && !ta.Equal(tb) {
			return ta.Before(tb)
		}
		return fs[a].Path < fs[b].Path
	})
}

func shellFolders(root string, m *Model) ([]Folder, error) {
	base := filepath.Join(root, filepath.FromSlash(ShellLoop.runsDir()))
	var out []Folder
	groups, err := readDirs(base)
	if err != nil {
		return nil, err
	}
	for _, g := range groups {
		folders, err := readDirs(filepath.Join(base, g))
		if err != nil {
			return nil, err
		}
		for _, f := range folders {
			rel := path.Join(ShellLoop.runsDir(), g, f)
			owns, err := logOwns(filepath.Join(root, filepath.FromSlash(rel), "loop.log"), m)
			if err != nil {
				return nil, err
			}
			if !owns {
				continue
			}
			fo, err := readFolder(root, rel, ShellLoop)
			if err != nil {
				return nil, err
			}
			out = append(out, fo)
		}
	}
	return out, nil
}

func vloopFolders(root string, m *Model) ([]Folder, error) {
	dir := path.Join(Vloop.runsDir(), m.RunID)
	names, err := readDirs(filepath.Join(root, filepath.FromSlash(dir)))
	if err != nil {
		return nil, err
	}
	var out []Folder
	for _, n := range names {
		fo, err := readFolder(root, path.Join(dir, n), Vloop)
		if err != nil {
			return nil, err
		}
		out = append(out, fo)
	}
	return out, nil
}

// readDirs lists the sub-directory names of dir, sorted; a missing dir is empty.
func readDirs(dir string) ([]string, error) {
	ents, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range ents {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out, nil
}

var ansi = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

// logOwns reports whether a shell-loop log says its folder planned from the
// brief or resumed the run. A missing log owns nothing.
func logOwns(logPath string, m *Model) (bool, error) {
	f, err := os.Open(logPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	defer f.Close()
	resuming := "resuming " + m.RunID
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<24)
	for sc.Scan() {
		line := ansi.ReplaceAllString(sc.Text(), "")
		if i := strings.Index(line, "planning from "); i >= 0 {
			rest := strings.TrimPrefix(strings.Fields(line[i+len("planning from "):] + " ")[0], "./")
			if rest == m.BriefPath {
				return true, nil
			}
		}
		if i := strings.Index(line, resuming); i >= 0 {
			after := line[i+len(resuming):]
			if after == "" || after[0] == ' ' || after[0] == '\t' {
				return true, nil
			}
		}
	}
	return false, sc.Err()
}

var (
	sessionName = regexp.MustCompile(`^(\d+)-([a-z-]+)\.json$`)
	verdictName = regexp.MustCompile(`^(\d+)-verdict\.json$`)
)

func readFolder(root, rel string, l Layout) (Folder, error) {
	dir := filepath.Join(root, filepath.FromSlash(rel))
	fo := Folder{Layout: l, Path: rel}
	iters, err := readIterations(filepath.Join(dir, "iterations.jsonl"))
	if err != nil {
		return fo, err
	}
	fo.Iterations = iters
	taskOf := map[int]string{}
	for _, it := range iters {
		taskOf[it.Iteration] = it.Task
	}

	ents, err := os.ReadDir(filepath.Join(dir, "sessions"))
	if err != nil && !os.IsNotExist(err) {
		return fo, err
	}
	for _, e := range ents {
		mm := sessionName.FindStringSubmatch(e.Name())
		if mm == nil {
			continue
		}
		s, err := readSession(filepath.Join(dir, "sessions", e.Name()))
		if err != nil {
			return fo, err
		}
		s.Seq, _ = strconv.Atoi(mm[1])
		if s.Phase == "" {
			s.Phase = mm[2]
		}
		if t, ok := taskOf[s.Iteration]; ok && s.Phase != "plan" {
			s.Task = t
		}
		fo.Sessions = append(fo.Sessions, s)
	}
	sort.SliceStable(fo.Sessions, func(a, b int) bool { return fo.Sessions[a].Seq < fo.Sessions[b].Seq })

	ents, err = os.ReadDir(filepath.Join(dir, "reports"))
	if err != nil && !os.IsNotExist(err) {
		return fo, err
	}
	for _, e := range ents {
		mm := verdictName.FindStringSubmatch(e.Name())
		if mm == nil {
			continue
		}
		v, err := readVerdict(filepath.Join(dir, "reports", e.Name()))
		if err != nil {
			return fo, err
		}
		v.Iteration, _ = strconv.Atoi(mm[1])
		if v.Task == "" {
			v.Task = taskOf[v.Iteration]
		}
		fo.Verdicts = append(fo.Verdicts, v)
	}
	sort.SliceStable(fo.Verdicts, func(a, b int) bool { return fo.Verdicts[a].Iteration < fo.Verdicts[b].Iteration })
	return fo, nil
}

func readJSON(p string, v any) error {
	b, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%s: %w", filepath.Base(p), err)
	}
	return nil
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

// rawSession accepts both the shell loop's claude result and session/v1.
type rawSession struct {
	Phase      string  `json:"phase"`
	Iteration  int     `json:"iteration"`
	Task       string  `json:"task"`
	Model      string  `json:"model"`
	Effort     *string `json:"effort"`
	Started    string  `json:"started"`
	DurationMS int64   `json:"duration_ms"`
	TotalCost  float64 `json:"total_cost_usd"`
	CostUSD    float64 `json:"cost_usd"`
	IsError    bool    `json:"is_error"`
	Denials    []any   `json:"permission_denials"`
	ModelUsage map[string]struct {
		Input         int64   `json:"inputTokens"`
		Output        int64   `json:"outputTokens"`
		CacheRead     int64   `json:"cacheReadInputTokens"`
		CacheCreation int64   `json:"cacheCreationInputTokens"`
		CostUSD       float64 `json:"costUSD"`
	} `json:"modelUsage"`
	ModelsUsed map[string]struct {
		Input         int64   `json:"input_tokens"`
		Output        int64   `json:"output_tokens"`
		CacheRead     int64   `json:"cache_read_input_tokens"`
		CacheCreation int64   `json:"cache_creation_input_tokens"`
		CostUSD       float64 `json:"cost_usd"`
	} `json:"models_used"`
}

func readSession(p string) (Session, error) {
	var r rawSession
	if err := readJSON(p, &r); err != nil {
		return Session{}, err
	}
	s := Session{
		Phase: r.Phase, Iteration: r.Iteration, Task: r.Task, Model: r.Model,
		Started: parseTime(r.Started), DurationMS: r.DurationMS, IsError: r.IsError,
		CostUSD: r.CostUSD, Denials: len(r.Denials), Models: map[string]Tokens{},
	}
	if r.Effort != nil {
		s.Effort = *r.Effort
	}
	if r.TotalCost != 0 {
		s.CostUSD = r.TotalCost
	}
	for name, u := range r.ModelUsage {
		s.Models[name] = Tokens{u.Input, u.Output, u.CacheRead, u.CacheCreation, u.CostUSD}
	}
	for name, u := range r.ModelsUsed {
		s.Models[name] = Tokens{u.Input, u.Output, u.CacheRead, u.CacheCreation, u.CostUSD}
	}
	return s, nil
}

func readIterations(p string) ([]Iteration, error) {
	b, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Iteration
	for n, line := range bytes.Split(b, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var r struct {
			Iteration int    `json:"iteration"`
			Task      string `json:"task"`
			Attempt   int    `json:"attempt"`
			Outcome   string `json:"outcome"`
			Gate      *struct {
				DurationMS *int64 `json:"duration_ms"`
				Flaky      bool   `json:"flaky"`
			} `json:"gate"`
			Checks []struct {
				DurationMS int64 `json:"duration_ms"`
			} `json:"checks"`
			Started string `json:"started"`
			Ended   string `json:"ended"`
		}
		if err := json.Unmarshal(line, &r); err != nil {
			return nil, fmt.Errorf("%s line %d: %w", filepath.Base(p), n+1, err)
		}
		it := Iteration{Iteration: r.Iteration, Task: r.Task, Attempt: r.Attempt, Outcome: r.Outcome,
			Started: parseTime(r.Started), Ended: parseTime(r.Ended)}
		if r.Gate != nil {
			it.GateMS = r.Gate.DurationMS
			it.Flaky = r.Gate.Flaky
		}
		if len(r.Checks) > 0 {
			var ms int64
			for _, c := range r.Checks {
				ms += c.DurationMS
			}
			it.ChecksMS = &ms
		}
		out = append(out, it)
	}
	return out, nil
}

func readVerdict(p string) (Verdict, error) {
	var r struct {
		Task     string            `json:"task"`
		Verdict  string            `json:"verdict"`
		Findings []json.RawMessage `json:"findings"`
	}
	if err := readJSON(p, &r); err != nil {
		return Verdict{}, err
	}
	v := Verdict{Task: r.Task, Verdict: r.Verdict}
	for _, raw := range r.Findings {
		var s string
		if json.Unmarshal(raw, &s) == nil {
			v.Findings = append(v.Findings, Finding{Summary: s})
			continue
		}
		var o struct {
			Summary string `json:"summary"`
			Kind    string `json:"kind"`
		}
		if err := json.Unmarshal(raw, &o); err != nil {
			return v, fmt.Errorf("%s: finding: %w", filepath.Base(p), err)
		}
		v.Findings = append(v.Findings, Finding{Summary: o.Summary, Kind: o.Kind})
	}
	return v, nil
}
