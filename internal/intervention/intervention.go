// Package intervention records what the operator does around a run besides
// testing, one file each under .vloop/interventions/.
package intervention

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
	"github.com/mvelosop/vloop/internal/defect"
)

// Dir is the repo-relative folder the intervention files live in.
const Dir = ".vloop/interventions"

// Enum values, in the order errors list them (and, for phases, the lifecycle).
var (
	Phases       = []string{"setup", "design", "run", "halt", "verify", "close", "next"}
	Kinds        = []string{"direction", "decision", "context-supply", "halt", "verification-finding", "repair", "carry-forward", "ceremony"}
	Automatables = []string{"yes", "partly", "no"}
	Bys          = []string{"operator", "assistant", "both"}
)

// SetFields are the fields `intervention set` may change.
var SetFields = []string{"brief", "phase", "kind", "automatable", "by", "occurred"}

// Intervention is the intervention/v1 frontmatter as JSON, plus the body.
type Intervention struct {
	Schema      string `json:"schema"`
	ID          string `json:"id"`
	Brief       string `json:"brief"`
	Phase       string `json:"phase"`
	Kind        string `json:"kind"`
	Automatable string `json:"automatable"`
	By          string `json:"by"`
	Occurred    string `json:"occurred"`
	Recorded    string `json:"recorded"`
	Backfilled  bool   `json:"backfilled,omitempty"`
	Summary     string `json:"summary"`
	Trigger     string `json:"trigger"`
	Done        string `json:"done"`
	Automation  string `json:"automation"`
}

// NewInput is what `intervention add` takes.
type NewInput struct {
	Summary, Brief, Phase, Kind, Automatable, By, Trigger, Done, Automation string
}

// Validate checks one enum value, returning the usage error the CLI reports.
func Validate(field, value string) error {
	var valid []string
	switch field {
	case "phase":
		valid = Phases
	case "kind":
		valid = Kinds
	case "automatable":
		valid = Automatables
	case "by":
		valid = Bys
	case "occurred":
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return &config.InvalidValueError{Key: field, Value: value}
		}
		return nil
	default:
		return nil
	}
	if !slices.Contains(valid, value) {
		return &config.InvalidValueError{Key: field, Value: value, Valid: valid}
	}
	return nil
}

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

// Add writes a new intervention file and returns its repo-relative path.
// Values are validated by the caller (Validate).
func Add(root string, in NewInput, now time.Time) (string, error) {
	summary := oneLine(in.Summary)
	v := Intervention{
		Schema: "intervention/v1", Brief: defect.BriefName(in.Brief), Phase: in.Phase, Kind: in.Kind,
		Automatable: in.Automatable, By: in.By, Occurred: now.Format("2006-01-02"),
		Recorded: now.UTC().Format("2006-01-02T15:04:05Z"), Summary: summary,
		Trigger: oneLine(in.Trigger), Done: oneLine(in.Done), Automation: oneLine(in.Automation),
	}
	dir := filepath.Join(root, filepath.FromSlash(Dir))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	slug := defect.Slug(summary)
	if slug == "defect" && summary == "" {
		slug = "intervention"
	}
	base := "I" + now.Format("20060102-1504") + "-" + slug
	for n := 1; ; n++ {
		id := base
		if n > 1 {
			id = base + "-" + strconv.Itoa(n)
		}
		v.ID = id
		f, err := os.OpenFile(filepath.Join(dir, id+".md"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		_, werr := f.WriteString(render(v))
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

func render(v Intervention) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "id: %s\nbrief: %s\nphase: %s\nkind: %s\nautomatable: %s\nby: %s\n", v.ID, quote(v.Brief), v.Phase, v.Kind, v.Automatable, v.By)
	fmt.Fprintf(&b, "occurred: %s\nrecorded: %s\n", v.Occurred, v.Recorded)
	if v.Backfilled {
		b.WriteString("backfilled: true\n")
	}
	fmt.Fprintf(&b, "---\n%s\n\n**Trigger.** %s\n\n**Done.** %s\n\n**What would automate it.** %s\n", v.Summary, v.Trigger, v.Done, v.Automation)
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

var sectionRe = regexp.MustCompile(`^\*\*(Trigger|Done|What would automate it)\.\*\*\s*`)

func parse(text string) (Intervention, error) {
	v := Intervention{Schema: "intervention/v1"}
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if len(lines) == 0 || lines[0] != "---" {
		return v, errors.New("missing frontmatter")
	}
	end := -1
	for i, l := range lines[1:] {
		if l == "---" {
			end = i + 1
			break
		}
		k, val, ok := strings.Cut(l, ":")
		if !ok {
			continue
		}
		val = unquote(val)
		switch k {
		case "id":
			v.ID = val
		case "brief":
			v.Brief = val
		case "phase":
			v.Phase = val
		case "kind":
			v.Kind = val
		case "automatable":
			v.Automatable = val
		case "by":
			v.By = val
		case "occurred":
			v.Occurred = val
		case "recorded":
			v.Recorded = val
		case "backfilled":
			v.Backfilled = val == "true"
		}
	}
	if end < 0 {
		return v, errors.New("unterminated frontmatter")
	}
	// The summary is the first body line; each section runs to the next marker.
	var cur *string
	var summary []string
	for _, l := range lines[end+1:] {
		if m := sectionRe.FindStringSubmatch(l); m != nil {
			switch m[1] {
			case "Trigger":
				cur = &v.Trigger
			case "Done":
				cur = &v.Done
			default:
				cur = &v.Automation
			}
			*cur = strings.TrimSpace(l[len(m[0]):])
			continue
		}
		if cur != nil {
			*cur = strings.TrimSpace(*cur + " " + strings.TrimSpace(l))
		} else {
			summary = append(summary, l)
		}
	}
	v.Summary = strings.TrimSpace(strings.Join(summary, "\n"))
	return v, nil
}

// List reads every recorded intervention, sorted by phase in lifecycle order,
// then kind, then id, keeping only briefs matching brief when it is not empty.
func List(root, brief string) ([]Intervention, error) {
	dir := filepath.Join(root, filepath.FromSlash(Dir))
	ents, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return []Intervention{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := []Intervention{}
	for _, e := range ents {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "I") || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		v, err := parse(string(b))
		if err != nil {
			return nil, fmt.Errorf("%s/%s: %w", Dir, e.Name(), err)
		}
		if brief != "" && v.Brief != defect.BriefName(brief) {
			continue
		}
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if pa, pb := slices.Index(Phases, a.Phase), slices.Index(Phases, b.Phase); pa != pb {
			return pa < pb
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.ID < b.ID
	})
	return out, nil
}

// Set changes one frontmatter field of a recorded intervention, keeping the
// rest of the file. Validation happens before the file is touched.
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
		return fmt.Errorf("no intervention %s", id)
	}
	if _, err := parse(string(b)); err != nil {
		return err
	}
	line := field + ": " + value
	if field == "brief" {
		line = field + ": " + quote(defect.BriefName(value))
	}
	lines := strings.Split(string(b), "\n")
	for i := 1; i < len(lines) && lines[i] != "---"; i++ {
		if strings.HasPrefix(lines[i], field+":") {
			lines[i] = line
			return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644)
		}
	}
	for i := 1; i < len(lines); i++ { // field absent: insert before the closing fence
		if lines[i] == "---" {
			lines = slices.Insert(lines, i, line)
			return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644)
		}
	}
	return errors.New("unterminated frontmatter")
}
