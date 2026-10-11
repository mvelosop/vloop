package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/driver"
	"github.com/mvelosop/vloop/internal/schema"
	"github.com/mvelosop/vloop/internal/state"
)

// loadPlan loads the plan under the repo root. A missing plan is a problem
// (exit 1); under --json stdout still carries one JSON document.
func loadPlan(g *Globals, out io.Writer) (*state.Plan, error) {
	root, err := g.root()
	if err != nil {
		return nil, err
	}
	p, err := state.Load(root)
	if err == nil {
		var vs []schema.Violation
		if vs, err = state.Validate(root); err == nil && len(vs) > 0 {
			err = state.ErrNotAPlan
		}
	}
	if err != nil {
		if g.JSON {
			_ = json.NewEncoder(out).Encode(struct {
				Error string `json:"error"`
			}{err.Error()})
		}
		return nil, Problem(err)
	}
	return p, nil
}

type taskLine struct {
	ID       string  `json:"id"`
	Status   string  `json:"status"`
	Area     *string `json:"area"`
	Kind     string  `json:"kind"`
	Title    string  `json:"title"`
	Attempts int     `json:"attempts"`
}

func taskLines(p *state.Plan) []taskLine {
	lines := make([]taskLine, len(p.Tasks))
	for i, t := range p.Tasks {
		l := taskLine{ID: t.ID, Status: t.Status, Kind: t.Kind, Title: t.Title, Attempts: t.Attempts}
		if t.Area != "" {
			a := t.Area
			l.Area = &a
		}
		lines[i] = l
	}
	return lines
}

func pad(s string, width int) string {
	if n := utf8.RuneCountInString(s); n < width {
		return s + strings.Repeat(" ", width-n)
	}
	return s
}

// writeTaskLines prints one line per task, columns left-aligned.
func writeTaskLines(out io.Writer, lines []taskLine) {
	var wID, wStatus, wAK int
	ak := make([]string, len(lines))
	for i, l := range lines {
		area := "-"
		if l.Area != nil {
			area = *l.Area
		}
		ak[i] = area + "/" + l.Kind
		wID = max(wID, utf8.RuneCountInString(l.ID))
		wStatus = max(wStatus, utf8.RuneCountInString(l.Status))
		wAK = max(wAK, utf8.RuneCountInString(ak[i]))
	}
	for i, l := range lines {
		s := "  " + pad(l.ID, wID) + "  " + pad(l.Status, wStatus) + "  " + pad(ak[i], wAK) + "  " + l.Title
		switch {
		case l.Attempts == 1:
			s += "  (1 attempt)"
		case l.Attempts > 1:
			s += fmt.Sprintf("  (%d attempts)", l.Attempts)
		}
		fmt.Fprintln(out, s)
	}
}

func encodeJSON(out io.Writer, v any) error {
	enc := json.NewEncoder(out)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

type runningJSON struct {
	Phase          string  `json:"phase"`
	Task           *string `json:"task,omitempty"`
	SessionStarted string  `json:"session_started"`
	Spent          float64 `json:"spent"`
	Ceiling        float64 `json:"ceiling"`
}

// costCeiling is the run's ceiling as the config states it; 0 when unreadable.
func costCeiling(root string) float64 {
	b, _ := driver.ResolveBudgets(root, nil)
	return b.CostCeiling
}

func newRunningJSON(root string, r driver.Running) *runningJSON {
	j := &runningJSON{Phase: r.Phase, SessionStarted: r.SessionStarted, Spent: r.Spent, Ceiling: costCeiling(root)}
	if r.Task != "" {
		t := r.Task
		j.Task = &t
	}
	return j
}

// runningLine is the line status prints for a live run: the session in
// flight and for how long, and the spend against the ceiling.
func runningLine(root string, r driver.Running) string {
	who := r.Phase
	switch {
	case who == "":
		who = "between sessions"
	case who == driver.PhaseGateReview:
		who = "gate review"
	}
	if r.Task != "" {
		who += " " + r.Task
	}
	line := "running: " + who
	if t, err := time.Parse(time.RFC3339, r.SessionStarted); err == nil {
		secs := max(int(time.Since(t).Round(time.Second)/time.Second), 0)
		line += fmt.Sprintf(" for %dm%02ds", secs/60, secs%60)
	}
	return fmt.Sprintf("%s · $%.2f of $%.2f", line, r.Spent, costCeiling(root))
}

func newStatus(g *Globals) *cobra.Command {
	var markdown bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the plan's progress",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			if markdown && g.JSON {
				return Usage(errors.New("--json and --markdown are mutually exclusive"))
			}
			var running *driver.Running
			if root, err := g.root(); err == nil {
				if r, ok := driver.ReadRunning(root); ok {
					running = &r
					// A run still planning has no plan yet; that is not a failure.
					if _, lerr := state.Load(root); errors.Is(lerr, state.ErrNoPlan) && !markdown {
						if g.JSON {
							return encodeJSON(out, struct {
								Running *runningJSON `json:"running"`
							}{newRunningJSON(root, r)})
						}
						fmt.Fprintln(out, runningLine(root, r))
						return nil
					}
				}
			}
			p, err := loadPlan(g, out)
			if err != nil {
				return err
			}
			done, blocked := p.Counts()
			switch {
			case markdown:
				_, err = io.WriteString(out, p.Markdown())
				return err
			case g.JSON:
				var rj *runningJSON
				if running != nil {
					root, _ := g.root()
					rj = newRunningJSON(root, *running)
				}
				return encodeJSON(out, struct {
					RunID     string       `json:"run_id"`
					Status    string       `json:"status"`
					Iteration int          `json:"iteration"`
					Done      int          `json:"done"`
					Total     int          `json:"total"`
					Blocked   int          `json:"blocked"`
					Tasks     []taskLine   `json:"tasks"`
					Running   *runningJSON `json:"running,omitempty"`
				}{p.RunID, p.Status, p.Iteration, done, len(p.Tasks), blocked, taskLines(p), rj})
			}
			fmt.Fprintf(out, "%s — %s · %d/%d done · %d blocked · iteration %d\n",
				p.RunID, p.Status, done, len(p.Tasks), blocked, p.Iteration)
			if running != nil {
				root, _ := g.root()
				fmt.Fprintln(out, runningLine(root, *running))
			}
			writeTaskLines(out, taskLines(p))
			return nil
		},
	}
	cmd.Flags().BoolVar(&markdown, "markdown", false, "print the plan as markdown")
	return cmd
}
