// Package state loads, saves and renders the plan: .vloop/state/state.json.
// The embedded state/v2 schema decides whether a plan is valid; the structs
// here only carry it, in schema order, so a load-save cycle is byte-stable.
package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mvelosop/vloop/internal/schema"
)

// FilePath is the plan's location relative to the repo root, with "/"
// separators on every OS.
const FilePath = ".vloop/state/state.json"

// SchemaName is the schema the plan is validated against.
const SchemaName = "state/v2"

// SchemaV1 is the plan vloop 1.x wrote. vloop 2 reads it for status and metrics
// and refuses to run it.
const SchemaV1 = "state/v1"

// ErrNoPlan is returned by Load when the plan file does not exist.
var ErrNoPlan = errors.New("no plan — vloop run <brief> makes one")

// Plan is the plan document. Field order is the schema's key order.
type Plan struct {
	Schema    string `json:"schema"`
	RunID     string `json:"run_id"`
	Brief     string `json:"brief"`
	Base      string `json:"base"`
	Branch    string `json:"branch"`
	Status    string `json:"status"`
	Iteration int    `json:"iteration"`
	Created   string `json:"created"`
	Updated   string `json:"updated"`
	Shell     string `json:"shell"`

	Checks      []PlanCheck `json:"checks"`
	GateScratch []string    `json:"gate_scratch"`
	GateReview  GateReview  `json:"gate_review"`
	Tasks       []Task      `json:"tasks"`
}

// PlanCheck is a repo check copied into the plan: a name, the paths it covers and
// the command it runs.
type PlanCheck struct {
	Name  string   `json:"name"`
	Paths []string `json:"paths"`
	Run   string   `json:"run"`
}

// GateReview is where the gate-review of the plan stands.
type GateReview struct {
	Rounds  int    `json:"rounds"`
	Verdict string `json:"verdict"`
}

// Task is one task of the plan. Field order is the schema's key order.
type Task struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Goal        string        `json:"goal"`
	Kind        string        `json:"kind"`
	Area        string        `json:"area,omitempty"`
	Fixtures    string        `json:"fixtures"`
	References  []Reference   `json:"references"`
	DependsOn   []string      `json:"depends_on"`
	Acceptance  []string      `json:"acceptance"`
	Verify      string        `json:"verify"`
	Status      string        `json:"status"`
	Attempts    int           `json:"attempts"`
	Notes       string        `json:"notes"`
	Model       *Sessions     `json:"model,omitempty"`
	Effort      *Sessions     `json:"effort,omitempty"`
	GateHistory []GateReplace `json:"gate_history,omitempty"`
}

// Reference is a file a task is bound by, and why.
type Reference struct {
	Path string `json:"path"`
	Why  string `json:"why"`
}

// Sessions holds a per-session-kind value (model or effort).
type Sessions struct {
	Work   string `json:"work,omitempty"`
	Review string `json:"review,omitempty"`
}

// GateReplace records a replaced verify command.
type GateReplace struct {
	Verify     string `json:"verify"`
	ReplacedAt string `json:"replaced_at"`
	Reason     string `json:"reason"`
	By         string `json:"by"`
	Fixtures   string `json:"fixtures"`
}

// Path returns the plan file's path under root.
func Path(root string) string { return filepath.Join(root, filepath.FromSlash(FilePath)) }

// Load reads the plan under root. It does not validate: a plan that parses
// but fails its schema still loads; use Validate for the verdict. A missing
// plan is ErrNoPlan.
func Load(root string) (*Plan, error) {
	data, err := os.ReadFile(Path(root))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNoPlan
	}
	if err != nil {
		return nil, err
	}
	var p Plan
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("%s: %w", FilePath, err)
	}
	return &p, nil
}

// Validate checks the plan file under root against the state/v2 schema.
func Validate(root string) ([]schema.Violation, error) {
	data, err := os.ReadFile(Path(root))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNoPlan
	}
	if err != nil {
		return nil, err
	}
	return schema.Validate(SchemaName, data)
}

// Marshal returns the canonical bytes of p: 2-space indent, keys in schema
// order, a trailing newline, and <, > and & written literally.
func Marshal(p *Plan) ([]byte, error) {
	q := *p
	if q.Checks == nil {
		q.Checks = []PlanCheck{}
	}
	for i, c := range q.Checks {
		if c.Paths == nil {
			q.Checks[i].Paths = []string{}
		}
	}
	if q.GateScratch == nil {
		q.GateScratch = []string{}
	}
	q.Tasks = make([]Task, len(p.Tasks))
	for i, t := range p.Tasks {
		if t.References == nil {
			t.References = []Reference{}
		}
		if t.DependsOn == nil {
			t.DependsOn = []string{}
		}
		if t.Acceptance == nil {
			t.Acceptance = []string{}
		}
		q.Tasks[i] = t
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(&q); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Save sets Updated to now (RFC 3339, UTC) and writes p under root
// atomically: a temporary file in the plan's directory, renamed over it.
func Save(root string, p *Plan) error {
	p.Updated = time.Now().UTC().Format(time.RFC3339)
	data, err := Marshal(p)
	if err != nil {
		return err
	}
	dest := Path(root)
	tmp, err := os.CreateTemp(filepath.Dir(dest), ".state-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	fail := func(err error) error {
		tmp.Close()
		os.Remove(name)
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		return fail(err)
	}
	if err := tmp.Chmod(0o644); err != nil && !errors.Is(err, errors.ErrUnsupported) {
		return fail(err)
	}
	if err := tmp.Close(); err != nil {
		return fail(err)
	}
	if err := os.Rename(name, dest); err != nil {
		os.Remove(name)
		return err
	}
	return nil
}

// Counts returns how many tasks are done and blocked.
func (p *Plan) Counts() (done, blocked int) {
	for _, t := range p.Tasks {
		switch t.Status {
		case "done":
			done++
		case "blocked":
			blocked++
		}
	}
	return
}

func note(t Task) string {
	switch {
	case t.Status == "blocked":
		return " · **blocked**"
	case t.Attempts > 0:
		return fmt.Sprintf(" · %d attempt(s)", t.Attempts)
	}
	return ""
}

// Markdown renders p for human reading, as .loop/render-plan.sh renders its plan.
func (p *Plan) Markdown() string {
	var b strings.Builder
	w := func(format string, a ...any) { fmt.Fprintf(&b, format+"\n", a...) }
	done, _ := p.Counts()
	w("# Plan — %s", p.RunID)
	w("")
	w("<!-- Rendered from %s by vloop status --markdown. Do NOT edit. -->", FilePath)
	w("")
	w("**Status:** %s · **%d/%d done** · iteration %d", p.Status, done, len(p.Tasks), p.Iteration)
	w("")
	w("**Brief:** `%s` · **Updated:** %s", p.Brief, p.Updated)
	w("")
	w("## Progress")
	w("")
	for _, t := range p.Tasks {
		mark := " "
		if t.Status == "done" {
			mark = "x"
		}
		w("- [%s] **%s** — %s%s", mark, t.ID, t.Title, note(t))
	}
	w("")
	w("## Tasks")
	w("")
	for _, t := range p.Tasks {
		deps := "none"
		if len(t.DependsOn) > 0 {
			deps = strings.Join(t.DependsOn, ", ")
		}
		w("### %s — %s", t.ID, t.Title)
		w("")
		w("`%s`%s · depends on: %s", t.Status, note(t), deps)
		w("")
		if t.Goal != "" {
			w("%s", t.Goal)
		} else {
			w("_no goal recorded_")
		}
		w("")
		w("**Acceptance**")
		w("")
		for _, a := range t.Acceptance {
			w("- %s", a)
		}
		w("")
		if t.Notes != "" {
			w("**From the last attempt:** %s", t.Notes)
			w("")
		}
		w("<details><summary>verify command</summary>")
		w("")
		w("```sh")
		w("%s", t.Verify)
		w("```")
		w("")
		w("</details>")
		w("")
	}
	return b.String()
}
