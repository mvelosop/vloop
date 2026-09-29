// Package metrics turns a brief's commits and run records into the numbers
// `vloop metrics` reports.
package metrics

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/mvelosop/vloop/internal/classify"
	"github.com/mvelosop/vloop/internal/runs"
)

// emptyTree is git's well-known empty tree, the parent of a root commit.
const emptyTree = "4b825dc642cb6eb9a060e54bf8d69288fbd2f904"

// Counts is a number of lines per category. Excluded paths are never counted.
type Counts struct {
	Code  int `json:"code"`
	Test  int `json:"test"`
	Docs  int `json:"docs"`
	Other int `json:"other"`
}

func (c *Counts) add(cat string, n int) {
	switch cat {
	case classify.Code:
		c.Code += n
	case classify.Test:
		c.Test += n
	case classify.Docs:
		c.Docs += n
	case classify.Other:
		c.Other += n
	}
}

func (c Counts) plus(o Counts) Counts {
	return Counts{c.Code + o.Code, c.Test + o.Test, c.Docs + o.Docs, c.Other + o.Other}
}

// Lines is the added non-blank lines of a diff, and its deleted non-blank
// lines apart.
type Lines struct {
	Added   Counts `json:"added"`
	Deleted Counts `json:"deleted"`
}

// Plus sums two line counts.
func (l Lines) Plus(o Lines) Lines {
	return Lines{l.Added.plus(o.Added), l.Deleted.plus(o.Deleted)}
}

func gitOut(root string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(errb.String()))
	}
	return out.Bytes(), nil
}

// DiffLines counts the lines of the diff from one tree-ish to another. Blank
// lines (whitespace only), binary files and excluded paths are not counted,
// and a rename counts only its content change.
func DiffLines(root, from, to string, c *classify.Classifier) (Lines, error) {
	b, err := gitOut(root, "-c", "core.quotepath=false", "diff", "-M", "-U0", "--no-color",
		"--no-ext-diff", "--no-textconv", "--src-prefix=a/", "--dst-prefix=b/", from, to)
	if err != nil {
		return Lines{}, err
	}
	return countPatch(string(b), c), nil
}

func patchPath(s string) string {
	s = strings.TrimSuffix(s, "\t")
	if strings.HasPrefix(s, `"`) {
		if u, err := strconv.Unquote(s); err == nil {
			s = u
		}
	}
	return s
}

func countPatch(patch string, c *classify.Classifier) Lines {
	var l Lines
	var oldPath, newPath string
	inHunk := false
	cat := func() string {
		p := newPath
		if p == "" {
			p = oldPath
		}
		return c.Classify(p).Category
	}
	for _, line := range strings.Split(patch, "\n") {
		switch {
		case strings.HasPrefix(line, "diff --git "):
			inHunk, oldPath, newPath = false, "", ""
		case !inHunk && strings.HasPrefix(line, "--- "):
			if p := patchPath(line[4:]); p != "/dev/null" {
				oldPath = strings.TrimPrefix(p, "a/")
			}
		case !inHunk && strings.HasPrefix(line, "+++ "):
			if p := patchPath(line[4:]); p != "/dev/null" {
				newPath = strings.TrimPrefix(p, "b/")
			}
		case strings.HasPrefix(line, "@@"):
			inHunk = true
		case inHunk && (strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-")):
			if strings.TrimSpace(line[1:]) == "" {
				continue
			}
			if line[0] == '+' {
				l.Added.add(cat(), 1)
			} else {
				l.Deleted.add(cat(), 1)
			}
		}
	}
	return l
}

// TaskChurn is the lines the commits of one task changed, whatever their
// outcome.
type TaskChurn struct {
	Task    string
	Commits int
	Lines   Lines
}

// Size is what a brief delivered and what it took to get there.
type Size struct {
	Delivered Lines
	Churn     Lines
	ByTask    []TaskChurn // in order of first commit
}

// Measure counts the delivered diff (plan commit's parent to the last run
// commit; the last task commit, then the plan, while there is no run commit)
// and the churn of every task commit.
func Measure(root string, o *runs.Owned, c *classify.Classifier) (*Size, error) {
	parent := func(cm runs.Commit) string {
		if len(cm.Parents) == 0 {
			return emptyTree
		}
		return cm.Parents[0]
	}
	head := o.Plan
	if n := len(o.Tasks); n > 0 {
		head = o.Tasks[n-1]
	}
	if o.Run != nil {
		head = *o.Run
	}
	var s Size
	var err error
	if s.Delivered, err = DiffLines(root, parent(o.Plan), head.SHA, c); err != nil {
		return nil, err
	}
	idx := map[string]int{}
	for _, t := range o.Tasks {
		l, err := DiffLines(root, parent(t), t.SHA, c)
		if err != nil {
			return nil, err
		}
		i, ok := idx[t.Task]
		if !ok {
			i = len(s.ByTask)
			idx[t.Task] = i
			s.ByTask = append(s.ByTask, TaskChurn{Task: t.Task})
		}
		s.ByTask[i].Commits++
		s.ByTask[i].Lines = s.ByTask[i].Lines.Plus(l)
		s.Churn = s.Churn.Plus(l)
	}
	return &s, nil
}

func ratio(num, den int) *float64 {
	if den == 0 {
		return nil
	}
	r := float64(num) / float64(den)
	return &r
}

// Rework is churn code over delivered code; nil when nothing was delivered.
func (s Size) Rework() *float64 { return ratio(s.Churn.Added.Code, s.Delivered.Added.Code) }

// TestCode is delivered test lines over delivered code lines; nil when no code
// was delivered.
func (s Size) TestCode() *float64 { return ratio(s.Delivered.Added.Test, s.Delivered.Added.Code) }
