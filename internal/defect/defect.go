// Package defect records the defects the loop cannot see, one file each under
// .vloop/defects/, and attributes them to the brief that introduced the line.
package defect

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/runs"
)

// Dir is the repo-relative folder the defect files live in.
const Dir = ".vloop/defects"

// BriefsDir is where loop briefs live.
const BriefsDir = "docs/briefs"

// Enum values, in the order errors list them.
var (
	Origins    = []string{"brief", "plan", "work", "env"}
	FoundBys   = []string{"gate", "review", "operator", "user"}
	Kinds      = []string{"bug", "spec-gap", "gate", "gate-gap", "regression"}
	Severities = []string{"low", "medium", "high", "critical"}
	Statuses   = []string{"open", "fixed", "wontfix"}
)

// SetFields are the fields `defect set` may change.
var SetFields = []string{"status", "fixed-by", "case", "severity", "origin", "kind", "task"}

// Defect is the defect/v1 frontmatter as JSON, plus the description.
type Defect struct {
	Schema   string `json:"schema"`
	ID       string `json:"id"`
	Brief    string `json:"brief"`
	Task     string `json:"task,omitempty"`
	Origin   string `json:"origin"`
	FoundBy  string `json:"found-by"`
	Kind     string `json:"kind"`
	Severity string `json:"severity"`
	Status   string `json:"status"`
	FixedBy  string `json:"fixed-by"`
	Case     string `json:"case"`
	Created  string `json:"created"`
	Summary  string `json:"summary"`
}

// NewInput is what `defect add` takes.
type NewInput struct {
	Summary, FoundBy, Brief, Task, Origin, Kind, Severity, Case string
}

// Validate checks one enum value, returning the usage error the CLI reports.
func Validate(field, value string) error {
	var valid []string
	switch field {
	case "origin":
		valid = Origins
	case "found-by":
		valid = FoundBys
	case "kind":
		valid = Kinds
	case "severity":
		valid = Severities
	case "status":
		valid = Statuses
	default:
		return nil
	}
	if !slices.Contains(valid, value) {
		return &config.InvalidValueError{Key: field, Value: value, Valid: valid}
	}
	return nil
}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// Slug is the summary lower-cased, each run of non-alphanumerics as "-", at
// most 40 characters.
func Slug(summary string) string {
	s := strings.Trim(nonAlnum.ReplaceAllString(strings.ToLower(summary), "-"), "-")
	if len(s) > 40 {
		s = strings.Trim(s[:40], "-")
	}
	if s == "" {
		s = "defect"
	}
	return s
}

// BriefName normalises a brief given as a name or a file name to the name.
func BriefName(brief string) string {
	if brief == "" {
		return ""
	}
	b := filepath.Base(filepath.ToSlash(brief))
	return strings.TrimSuffix(b, ".md")
}

// CheckBrief errors unless brief is a *.loop-brief.md in the briefs directory.
func CheckBrief(root, brief string) error {
	if !strings.HasSuffix(brief, ".loop-brief") {
		return fmt.Errorf("%q is not a loop brief", brief)
	}
	if st, err := os.Stat(filepath.Join(root, BriefsDir, brief+".md")); err != nil || st.IsDir() {
		return fmt.Errorf("no loop brief %s in %s", brief, BriefsDir)
	}
	return nil
}

// CheckTask errors unless task is in the brief's plan at its last run commit.
func CheckTask(root, brief, task string) error {
	o, err := runs.OwnedCommits(root, runs.RunID(brief))
	if err != nil {
		return err
	}
	if o == nil {
		return fmt.Errorf("no plan found for %s, so task %s cannot be checked", brief, task)
	}
	sha := o.Plan.SHA
	if o.Run != nil {
		sha = o.Run.SHA
	} else if n := len(o.Tasks); n > 0 {
		sha = o.Tasks[n-1].SHA
	}
	plan, err := runs.PlanAt(root, o.Layout, sha)
	if err != nil {
		return err
	}
	for _, t := range plan.Tasks {
		if t.ID == task {
			return nil
		}
	}
	return fmt.Errorf("no task %s in the plan of %s", task, brief)
}

// Add writes a new defect file and returns its repo-relative path. Values are
// validated by the caller (Validate, CheckBrief, CheckTask).
func Add(root string, in NewInput, now time.Time) (string, error) {
	if in.Origin == "" {
		in.Origin = "work"
	}
	if in.Kind == "" {
		in.Kind = "bug"
	}
	if in.Severity == "" {
		in.Severity = "medium"
	}
	summary := strings.Join(strings.Fields(in.Summary), " ")
	d := Defect{
		Schema: "defect/v1", Brief: in.Brief, Task: in.Task, Origin: in.Origin,
		FoundBy: in.FoundBy, Kind: in.Kind, Severity: in.Severity, Status: "open",
		Case: in.Case, Created: now.UTC().Format("2006-01-02T15:04:05Z"), Summary: summary,
	}
	dir := filepath.Join(root, filepath.FromSlash(Dir))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	base := "D" + now.Format("20060102-1504") + "-" + Slug(summary)
	for n := 1; ; n++ {
		id := base
		if n > 1 {
			id = base + "-" + strconv.Itoa(n)
		}
		d.ID = id
		f, err := os.OpenFile(filepath.Join(dir, id+".md"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		_, werr := f.WriteString(render(d))
		if cerr := f.Close(); werr == nil {
			werr = cerr
		}
		return Dir + "/" + id + ".md", werr
	}
}

func quote(s string) string {
	if s == "" || strings.ContainsAny(s, ":#\"'") || s != strings.TrimSpace(s) {
		return strconv.Quote(s)
	}
	return s
}

func render(d Defect) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "id: %s\nbrief: %s\n", d.ID, d.Brief)
	if d.Task != "" {
		fmt.Fprintf(&b, "task: %s\n", d.Task)
	}
	fmt.Fprintf(&b, "origin: %s\nfound-by: %s\nkind: %s\nseverity: %s\nstatus: %s\n", d.Origin, d.FoundBy, d.Kind, d.Severity, d.Status)
	fmt.Fprintf(&b, "fixed-by: %s\ncase: %s\ncreated: %s\n---\n%s\n", quote(d.FixedBy), quote(d.Case), d.Created, d.Summary)
	return b.String()
}

func unquote(v string) string {
	v = strings.TrimSpace(v)
	if strings.HasPrefix(v, `"`) {
		if u, err := strconv.Unquote(v); err == nil {
			return u
		}
	}
	return v
}

func parse(text string) (Defect, error) {
	d := Defect{Schema: "defect/v1"}
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if len(lines) == 0 || lines[0] != "---" {
		return d, errors.New("missing frontmatter")
	}
	end := -1
	for i, l := range lines[1:] {
		if l == "---" {
			end = i + 1
			break
		}
		k, v, ok := strings.Cut(l, ":")
		if !ok {
			continue
		}
		v = unquote(v)
		switch k {
		case "id":
			d.ID = v
		case "brief":
			d.Brief = v
		case "task":
			d.Task = v
		case "origin":
			d.Origin = v
		case "found-by":
			d.FoundBy = v
		case "kind":
			d.Kind = v
		case "severity":
			d.Severity = v
		case "status":
			d.Status = v
		case "fixed-by":
			d.FixedBy = v
		case "case":
			d.Case = v
		case "created":
			d.Created = v
		}
	}
	if end < 0 {
		return d, errors.New("unterminated frontmatter")
	}
	d.Summary = strings.TrimRight(strings.Join(lines[end+1:], "\n"), "\n")
	return d, nil
}

// List reads every recorded defect, sorted by id, keeping only briefs matching
// brief when it is not empty.
func List(root, brief string) ([]Defect, error) {
	dir := filepath.Join(root, filepath.FromSlash(Dir))
	ents, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return []Defect{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := []Defect{}
	for _, e := range ents {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "D") || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		d, err := parse(string(b))
		if err != nil {
			return nil, fmt.Errorf("%s/%s: %w", Dir, e.Name(), err)
		}
		if brief != "" && d.Brief != BriefName(brief) {
			continue
		}
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// Set changes one field of a recorded defect. Validation happens before the
// file is touched, so a refused set leaves it byte-identical.
func Set(root, id, field, value string) error {
	if !slices.Contains(SetFields, field) {
		return fmt.Errorf("cannot set %q: want one of %s", field, strings.Join(SetFields, ", "))
	}
	if err := Validate(field, value); err != nil {
		return err
	}
	if strings.ContainsAny(value, "\r\n") {
		return &config.InvalidValueError{Key: field, Value: value}
	}
	path := filepath.Join(root, filepath.FromSlash(Dir), filepath.Base(id)+".md")
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("no defect %s", id)
	}
	d, err := parse(string(b))
	if err != nil {
		return err
	}
	if field == "task" && value != "" {
		if err := CheckTask(root, d.Brief, value); err != nil {
			return err
		}
	}
	line := field + ": " + quote(value)
	if field == "task" && value == "" {
		line = ""
	} else if field != "fixed-by" && field != "case" {
		line = field + ": " + value
	}
	lines := strings.Split(string(b), "\n")
	end := 0
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			end = i
			break
		}
	}
	found := -1
	for i := 1; i < end; i++ {
		if strings.HasPrefix(lines[i], field+":") {
			found = i
			break
		}
	}
	switch {
	case found >= 0 && line == "":
		lines = slices.Delete(lines, found, found+1)
	case found >= 0:
		lines[found] = line
	case line != "": // task was absent: it follows brief
		for i := 1; i < end; i++ {
			if strings.HasPrefix(lines[i], "brief:") {
				lines = slices.Insert(lines, i+1, line)
				break
			}
		}
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644)
}
