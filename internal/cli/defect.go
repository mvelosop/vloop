package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/defect"
	"github.com/mvelosop/vloop/internal/metrics"
	"github.com/mvelosop/vloop/internal/runs"
)

func newDefect(g *Globals) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "defect",
		Short: "Record, list and update defects the loop cannot see",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	cmd.AddCommand(newDefectAdd(g), newDefectList(g), newDefectSet(g))
	return cmd
}

// defectErr maps a defect error: invalid values are usage errors (exit 2),
// everything else is a problem (exit 1).
func defectErr(err error) error {
	var inv *config.InvalidValueError
	if errors.As(err, &inv) {
		return err
	}
	return Problem(err)
}

func newDefectAdd(g *Globals) *cobra.Command {
	var in defect.NewInput
	var blame string
	cmd := &cobra.Command{
		Use:   `add "<summary>"`,
		Short: "Record a defect as .vloop/defects/D<stamp>-<slug>.md and print its path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			in.Summary = args[0]
			if strings.TrimSpace(in.Summary) == "" {
				return errors.New("the summary must not be empty")
			}
			if in.FoundBy == "" {
				return errors.New("--found-by is required")
			}
			for _, v := range []struct{ f, v string }{{"found-by", in.FoundBy}, {"origin", in.Origin}, {"kind", in.Kind}, {"severity", in.Severity}} {
				if v.v == "" {
					continue
				}
				if err := defect.Validate(v.f, v.v); err != nil {
					return err
				}
			}
			if in.Brief == "" && blame == "" {
				return errors.New("pass --brief or --blame <file>:<line>")
			}
			root, err := g.root()
			if err != nil {
				return err
			}
			if in.Brief != "" {
				in.Brief = defect.BriefName(in.Brief)
				if err := defect.CheckBrief(root, in.Brief); err != nil {
					return Problem(err)
				}
			} else {
				i := strings.LastIndex(blame, ":")
				n, aerr := strconv.Atoi(blame[i+1:])
				if i <= 0 || aerr != nil || n < 1 {
					return fmt.Errorf("invalid value %q for blame: want <file>:<line>", blame)
				}
				a, err := defect.Blame(root, blame[:i], n)
				if err != nil {
					return Problem(err)
				}
				in.Brief = a.Brief
				fmt.Fprintln(cmd.ErrOrStderr(), a)
			}
			if in.Task != "" {
				if err := defect.CheckTask(root, in.Brief, in.Task); err != nil {
					return Problem(err)
				}
			}
			p, err := defect.Add(root, in, time.Now())
			if err != nil {
				return Problem(err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), p)
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&in.FoundBy, "found-by", "", "who caught it: "+strings.Join(defect.FoundBys, ", "))
	f.StringVar(&in.Brief, "brief", "", "the loop brief that introduced it")
	f.StringVar(&in.Task, "task", "", "the task in that brief's plan")
	f.StringVar(&in.Origin, "origin", "work", "where it came from: "+strings.Join(defect.Origins, ", "))
	f.StringVar(&in.Kind, "kind", "bug", "kind: "+strings.Join(defect.Kinds, ", "))
	f.StringVar(&in.Severity, "severity", "medium", "severity: "+strings.Join(defect.Severities, ", "))
	f.StringVar(&in.Case, "case", "", "the failing test written first")
	f.StringVar(&blame, "blame", "", "attribute to the brief that wrote `file:line` when --brief is absent")
	return cmd
}

func newDefectList(g *Globals) *cobra.Command {
	var brief string
	var matrix bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Print the recorded defects, or with --matrix the origin × catcher counts",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := g.root()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if matrix {
				x, err := defectMatrix(root, brief)
				if err != nil {
					return Problem(err)
				}
				if g.JSON {
					return json.NewEncoder(out).Encode(x)
				}
				fmt.Fprint(out, x.String())
				return nil
			}
			ds, err := defect.List(root, brief)
			if err != nil {
				return Problem(err)
			}
			if g.JSON {
				return json.NewEncoder(out).Encode(ds)
			}
			for _, d := range ds {
				task := d.Task
				if task == "" {
					task = "-"
				}
				fmt.Fprintf(out, "%s  %s  %s  %s  %s  %s  %s  %s\n", d.ID, d.Brief, task, d.Kind, d.Origin, d.FoundBy, d.Status, d.Summary)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&brief, "brief", "", "only defects of this loop brief")
	cmd.Flags().BoolVar(&matrix, "matrix", false, "print origin × catcher counts, derived defects included")
	return cmd
}

// defectMatrix counts the derived and recorded defects of one brief, or of
// every brief with runs when brief is empty.
func defectMatrix(root, brief string) (metrics.Matrix, error) {
	var x metrics.Matrix
	var briefs []string
	if brief != "" {
		briefs = []string{defect.BriefsDir + "/" + defect.BriefName(brief) + ".md"}
	} else {
		ents, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(defect.BriefsDir)))
		if err != nil && !os.IsNotExist(err) {
			return x, err
		}
		for _, e := range ents {
			if n := e.Name(); !e.IsDir() && strings.HasSuffix(n, ".loop-brief.md") {
				briefs = append(briefs, defect.BriefsDir+"/"+n)
			}
		}
	}
	for _, b := range briefs {
		m, err := runs.Read(root, b)
		if err != nil {
			return x, err
		}
		if len(m.Folders) == 0 {
			continue
		}
		var plan *runs.PlanDoc
		if o := m.Owned; o != nil {
			sha := o.Plan.SHA
			if o.Run != nil {
				sha = o.Run.SHA
			} else if n := len(o.Tasks); n > 0 {
				sha = o.Tasks[n-1].SHA
			}
			if plan, err = runs.PlanAt(root, o.Layout, sha); err != nil {
				return x, err
			}
		}
		x.AddDerived(metrics.Derive(m, plan))
	}
	ds, err := defect.List(root, brief)
	if err != nil {
		return x, err
	}
	x.AddRecorded(ds)
	return x, nil
}

func newDefectSet(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "set <id> <field> <value>",
		Short: "Set " + strings.Join(defect.SetFields, ", ") + " of a defect",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := g.root()
			if err != nil {
				return err
			}
			if err := defect.Set(root, args[0], args[1], args[2]); err != nil {
				var inv *config.InvalidValueError
				if errors.As(err, &inv) || !contains(defect.SetFields, args[1]) {
					return err
				}
				return Problem(err)
			}
			return nil
		},
	}
}

func contains(l []string, s string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}
