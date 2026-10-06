package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"

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
				enc := json.NewEncoder(out)
				enc.SetEscapeHTML(false)
				return enc.Encode(struct {
					RunID     string     `json:"run_id"`
					Status    string     `json:"status"`
					Iteration int        `json:"iteration"`
					Done      int        `json:"done"`
					Total     int        `json:"total"`
					Blocked   int        `json:"blocked"`
					Tasks     []taskLine `json:"tasks"`
				}{p.RunID, p.Status, p.Iteration, done, len(p.Tasks), blocked, taskLines(p)})
			}
			fmt.Fprintf(out, "%s — %s · %d/%d done · %d blocked · iteration %d\n",
				p.RunID, p.Status, done, len(p.Tasks), blocked, p.Iteration)
			writeTaskLines(out, taskLines(p))
			return nil
		},
	}
	cmd.Flags().BoolVar(&markdown, "markdown", false, "print the plan as markdown")
	return cmd
}
