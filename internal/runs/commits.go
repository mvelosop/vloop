package runs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Commit is one commit, as far as metrics need it.
type Commit struct {
	SHA     string
	Parents []string
	Time    time.Time
	Subject string
	Task    string // task commits only
	Outcome string // task commits: the outcome; run commits: the status
}

// Owned is the set of commits a brief owns: its latest plan commit, the last
// run commit after it on the same history, and the task commits between them.
type Owned struct {
	Layout Layout
	Plan   Commit
	Run    *Commit // nil while no run commit follows the plan
	Tasks  []Commit
}

// PlanTask is a task as the plan at a commit records it. The shell loop's
// plans predate state/v1, so only the fields both share are read.
type PlanTask struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Area     string `json:"area"`
	Kind     string `json:"kind"`
	Status   string `json:"status"`
	Attempts int    `json:"attempts"`
}

// PlanDoc is the plan at a commit.
type PlanDoc struct {
	RunID string     `json:"run_id"`
	Brief string     `json:"brief"`
	Tasks []PlanTask `json:"tasks"`
}

const fieldSep = "\x1f"

func git(root string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(errb.String()))
	}
	return out.Bytes(), nil
}

func parseLog(b []byte) []Commit {
	var out []Commit
	for _, line := range strings.Split(string(b), "\n") {
		f := strings.SplitN(line, fieldSep, 4)
		if len(f) != 4 {
			continue
		}
		c := Commit{SHA: f[0], Subject: f[3]}
		if f[1] != "" {
			c.Parents = strings.Fields(f[1])
		}
		if sec, err := strconv.ParseInt(f[2], 10, 64); err == nil {
			c.Time = time.Unix(sec, 0).UTC()
		}
		out = append(out, c)
	}
	return out
}

const logFormat = "--format=%H" + fieldSep + "%P" + fieldSep + "%ct" + fieldSep + "%s"

var taskSubject = regexp.MustCompile(`^\[v?loop\] (\S+): (\S+)$`)

// OwnedCommits finds the commits the run id owns, across every local ref and
// both layouts. It returns nil when no plan commit exists for the run id.
func OwnedCommits(root, runID string) (*Owned, error) {
	all, err := git(root, "log", "--all", "--topo-order", logFormat)
	if err != nil {
		return nil, err
	}
	var best *Commit
	var layout Layout
	for _, c := range parseLog(all) {
		for _, l := range Layouts {
			if c.Subject != l.Prefix+" plan "+runID {
				continue
			}
			// topo-order lists descendants first, so a strict > keeps the
			// later commit when times tie.
			if best == nil || c.Time.After(best.Time) {
				c := c
				best, layout = &c, l
			}
		}
	}
	if best == nil {
		return nil, nil
	}
	o := &Owned{Layout: layout, Plan: *best}

	after, err := git(root, "log", "--all", "--topo-order", "--reverse", "--ancestry-path", logFormat, "^"+best.SHA)
	if err != nil {
		return nil, err
	}
	runPrefix := layout.Prefix + " run " + runID + "/"
	descendants := parseLog(after)
	for _, c := range descendants {
		if strings.HasPrefix(c.Subject, runPrefix) && (o.Run == nil || !c.Time.Before(o.Run.Time)) {
			c := c
			c.Outcome = c.Subject[strings.LastIndex(c.Subject, ": ")+2:]
			o.Run = &c
		}
	}
	between := descendants
	if o.Run != nil {
		b, err := git(root, "log", "--topo-order", "--reverse", "--ancestry-path", logFormat, best.SHA+".."+o.Run.SHA)
		if err != nil {
			return nil, err
		}
		between = parseLog(b)
	}
	for _, c := range between {
		m := taskSubject.FindStringSubmatch(c.Subject)
		if m == nil || !strings.HasPrefix(c.Subject, layout.Prefix+" ") {
			continue
		}
		c.Task, c.Outcome = m[1], m[2]
		o.Tasks = append(o.Tasks, c)
	}
	return o, nil
}

// PlanAt reads the plan of the given layout as committed at sha. The working
// tree is not consulted.
func PlanAt(root string, l Layout, sha string) (*PlanDoc, error) {
	b, err := git(root, "show", sha+":"+l.StatePath())
	if err != nil {
		return nil, err
	}
	var p PlanDoc
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("%s at %s: %w", l.StatePath(), sha, err)
	}
	return &p, nil
}
