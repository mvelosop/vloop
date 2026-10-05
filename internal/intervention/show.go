package intervention

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mvelosop/vloop/internal/defect"
	"github.com/mvelosop/vloop/internal/runs"
)

// Link is an id written in a record's Context and what it resolves to. A ref
// that resolves to nothing has an empty Title.
type Link struct {
	Ref   string `json:"ref"`
	Kind  string `json:"kind"`
	Title string `json:"title"`
}

// NotFound is the title text printed for a ref that resolves to nothing.
const NotFound = "(not found)"

// NoInterventionError is returned by Get for an id with no record.
type NoInterventionError struct{ ID string }

func (e *NoInterventionError) Error() string { return "no intervention " + e.ID }

// Get reads one recorded intervention by id.
func Get(root, id string) (Intervention, error) {
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(Dir), filepath.Base(id)+".md"))
	if errors.Is(err, os.ErrNotExist) || id == "" {
		return Intervention{}, &NoInterventionError{id}
	}
	if err != nil {
		return Intervention{}, err
	}
	v, err := parse(string(b))
	if err != nil {
		return v, fmt.Errorf("%s/%s.md: %w", Dir, id, err)
	}
	return v, nil
}

var (
	wordRe   = regexp.MustCompile(`[A-Za-z0-9-]+`)
	taskRe   = regexp.MustCompile(`^T[0-9]+$`)
	runIDRe  = regexp.MustCompile(`^B[0-9]{8}-[0-9]{4}-[a-z0-9-]+$`)
	defectRe = regexp.MustCompile(`^D[0-9]{8}-[0-9]{4}-[a-z0-9-]+$`)
	intervRe = regexp.MustCompile(`^I[0-9]{8}-[0-9]{4}-[a-z0-9-]+$`)
	shaRe    = regexp.MustCompile(`^[0-9a-f]{7,40}$`)
)

// linkKind says what kind of id a word is, or "" when it is not an id.
func linkKind(w string) string {
	switch {
	case taskRe.MatchString(w):
		return "task"
	case runIDRe.MatchString(w):
		return "run"
	case defectRe.MatchString(w):
		return "defect"
	case intervRe.MatchString(w):
		return "intervention"
	case shaRe.MatchString(w):
		return "commit"
	}
	return ""
}

// Links resolves the ids named in v's Context, one per id, in order of first
// appearance. Only the Context section is scanned.
func Links(root string, v Intervention) []Link {
	out := []Link{}
	seen := map[string]bool{}
	for _, w := range wordRe.FindAllString(v.Context, -1) {
		k := linkKind(w)
		if k == "" || seen[w] {
			continue
		}
		seen[w] = true
		out = append(out, Link{Ref: w, Kind: k, Title: resolve(root, v, k, w)})
	}
	return out
}

func resolve(root string, v Intervention, kind, ref string) string {
	switch kind {
	case "task":
		return taskTitle(root, v.Brief, ref)
	case "run":
		return runFolders(root, ref)
	case "defect":
		ds, err := defect.List(root, "")
		if err == nil {
			for _, d := range ds {
				if d.ID == ref {
					return d.Summary
				}
			}
		}
	case "intervention":
		if o, err := Get(root, ref); err == nil {
			return o.Summary
		}
	case "commit":
		return commitSubject(root, ref)
	}
	return ""
}

func gitOut(root string, args ...string) (string, bool) {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if cmd.Run() != nil {
		return "", false
	}
	return strings.TrimSpace(out.String()), true
}

func commitSubject(root, sha string) string {
	full, ok := gitOut(root, "rev-parse", "--verify", "-q", sha+"^{commit}")
	if !ok || full == "" {
		return ""
	}
	s, _ := gitOut(root, "log", "-1", "--format=%s", full)
	return s
}

// taskTitle reads the task's title from the plan of the latest "[vloop] plan"
// commit of the record's brief.
func taskTitle(root, brief, task string) string {
	runID := strings.TrimSuffix(brief, ".loop-brief")
	if runID == "" {
		return ""
	}
	for _, l := range runs.Layouts {
		sha, ok := gitOut(root, "log", "-1", "--format=%H", "-E", "--grep", "^"+regexp.QuoteMeta(l.Prefix+" plan "+runID)+"$")
		if !ok || sha == "" {
			continue
		}
		p, err := runs.PlanAt(root, l, sha)
		if err != nil {
			continue
		}
		for _, t := range p.Tasks {
			if t.ID == task {
				return t.Title
			}
		}
	}
	return ""
}

// runFolders names the run folders recorded for a run id, repo-relative.
func runFolders(root, runID string) string {
	var names []string
	for _, l := range runs.Layouts {
		rel := l.Dir + "/state/runs/" + runID
		ents, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			continue
		}
		for _, e := range ents {
			if e.IsDir() {
				names = append(names, rel+"/"+e.Name())
			}
		}
	}
	return strings.Join(names, ", ")
}

// Shown is a record as `intervention show --json` prints it.
type Shown struct {
	Intervention
	Links []Link `json:"links"`
}

// MarshalJSON keeps the embedded record's fields inline.
func (s Shown) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(s.Intervention)
	if err != nil {
		return nil, err
	}
	l, err := json.Marshal(s.Links)
	if err != nil {
		return nil, err
	}
	return append(b[:len(b)-1], []byte(`,"links":`+string(l)+"}")...), nil
}

// Text renders the record for people: id and summary, the frontmatter fields,
// each section under its name, then the links.
func (s Shown) Text() string {
	v := s.Intervention
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n%s\n\n", v.ID, v.Summary)
	field := func(k, val string) { fmt.Fprintf(&b, "%s: %s\n", k, val) }
	field("brief", v.Brief)
	field("phase", v.Phase)
	field("kind", v.Kind)
	field("automatable", v.Automatable)
	field("by", v.By)
	field("occurred", v.Occurred)
	field("recorded", v.Recorded)
	field("agreement", v.Agreement)
	if len(v.Options) > 0 {
		field("recommended", fmt.Sprint(v.Recommended.Option))
		decided := "other"
		if v.Decided.Option > 0 {
			decided = fmt.Sprint(v.Decided.Option)
		}
		field("decided", decided)
		field("adjusted", fmt.Sprint(v.Decided.Adjusted))
	}
	section := func(name, text string) {
		if text != "" {
			fmt.Fprintf(&b, "\n%s\n%s\n", name, text)
		}
	}
	section("Trigger", v.Trigger)
	section("Done", v.Done)
	section("Context", v.Context)
	if len(v.Options) > 0 {
		b.WriteString("\nOptions\n")
		for i, o := range v.Options {
			fmt.Fprintf(&b, "%d. %s\n", i+1, o)
		}
	}
	section("Recommended", v.Recommended.Why)
	section("Suggested", v.Suggested)
	section("Decided", v.Decided.Text)
	section("What would automate it", v.Automation)
	if len(s.Links) > 0 {
		b.WriteString("\nlinks:\n")
		for _, l := range s.Links {
			t := l.Title
			if t == "" {
				t = NotFound
			}
			fmt.Fprintf(&b, "  %s  %s\n", l.Ref, t)
		}
	}
	return b.String()
}
