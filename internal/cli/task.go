package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/state"
)

// Sources of a resolved model or effort that come from the task itself.
const sourceTask = "task"

// resolvedValue is what a session kind would run with, and where it came from.
type resolvedValue struct {
	value, source string
}

// resolveSession resolves the model or effort ("model", "effort") for a
// session kind ("work", "review"): the task's own value, else the environment,
// else the config file, else the default.
func resolveSession(root, field, kind string, taskValue string) (resolvedValue, error) {
	if taskValue != "" {
		return resolvedValue{taskValue, sourceTask}, nil
	}
	v, err := config.Get(root, field+"."+kind)
	if err != nil {
		return resolvedValue{}, err
	}
	if !v.Set {
		return resolvedValue{"", v.Source}, nil
	}
	return resolvedValue{v.Value, v.Source}, nil
}

type resolvedSession struct {
	Model        string  `json:"model"`
	ModelSource  string  `json:"model_source"`
	Effort       *string `json:"effort"`
	EffortSource string  `json:"effort_source"`
}

func sessionValue(s *state.Sessions, kind string) string {
	if s == nil {
		return ""
	}
	if kind == "work" {
		return s.Work
	}
	return s.Review
}

func resolveKind(root string, t *state.Task, kind string) (resolvedSession, error) {
	m, err := resolveSession(root, "model", kind, sessionValue(t.Model, kind))
	if err != nil {
		return resolvedSession{}, err
	}
	e, err := resolveSession(root, "effort", kind, sessionValue(t.Effort, kind))
	if err != nil {
		return resolvedSession{}, err
	}
	r := resolvedSession{Model: m.value, ModelSource: m.source, EffortSource: e.source}
	if e.value != "" {
		r.Effort = &e.value
	}
	return r, nil
}

func effortText(r resolvedSession) string {
	if r.Effort == nil {
		return "-"
	}
	return *r.Effort
}

func newTask(g *Globals) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Inspect the plan's tasks",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "Print one line per task",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			p, err := loadPlan(g, out)
			if err != nil {
				return err
			}
			if g.JSON {
				return json.NewEncoder(out).Encode(taskLines(p))
			}
			writeTaskLines(out, taskLines(p))
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "show <id>",
		Short: "Print a task with the model and effort its sessions resolve to",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			p, err := loadPlan(g, out)
			if err != nil {
				return err
			}
			var t *state.Task
			for i := range p.Tasks {
				if p.Tasks[i].ID == args[0] {
					t = &p.Tasks[i]
				}
			}
			if t == nil {
				return jsonProblem(g, out, fmt.Errorf("no task %s", args[0]))
			}
			root, err := g.root()
			if err != nil {
				return err
			}
			work, err := resolveKind(root, t, "work")
			if err == nil {
				var review resolvedSession
				if review, err = resolveKind(root, t, "review"); err == nil {
					return showTask(g, out, t, work, review)
				}
			}
			return configErr(g, out, err)
		},
	})
	cmd.AddCommand(newTaskValidate(g))
	return cmd
}

// jsonProblem is a problem (exit 1); under --json stdout still carries one
// JSON document.
func jsonProblem(g *Globals, out io.Writer, err error) error {
	if g.JSON {
		_ = json.NewEncoder(out).Encode(struct {
			Error string `json:"error"`
		}{err.Error()})
	}
	return Problem(err)
}

func showTask(g *Globals, out io.Writer, t *state.Task, work, review resolvedSession) error {
	if g.JSON {
		// The task object as the plan carries it, plus the resolution.
		b, err := json.Marshal(t)
		if err != nil {
			return err
		}
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(b, &obj); err != nil {
			return err
		}
		res, err := json.Marshal(struct {
			Work   resolvedSession `json:"work"`
			Review resolvedSession `json:"review"`
		}{work, review})
		if err != nil {
			return err
		}
		obj["resolved"] = res
		enc := json.NewEncoder(out)
		enc.SetEscapeHTML(false)
		return enc.Encode(obj)
	}
	area := t.Area
	if area == "" {
		area = "-"
	}
	fmt.Fprintf(out, "%s  %s\n", t.ID, t.Title)
	fmt.Fprintf(out, "status  %s · %d attempt(s)\n", t.Status, t.Attempts)
	fmt.Fprintf(out, "kind    %s\n", t.Kind)
	fmt.Fprintf(out, "area    %s\n", area)
	fmt.Fprintf(out, "depends %s\n", dash(strings.Join(t.DependsOn, ", ")))
	fmt.Fprintf(out, "goal    %s\n", t.Goal)
	writeList(out, "files", t.Files)
	writeList(out, "acceptance", t.Acceptance)
	refs := make([]string, len(t.References))
	for i, r := range t.References {
		refs[i] = r.Path + " — " + r.Why
	}
	writeList(out, "references", refs)
	fmt.Fprintf(out, "verify  %s\n", t.Verify)
	if t.Notes != "" {
		fmt.Fprintf(out, "notes   %s\n", t.Notes)
	}
	fmt.Fprintf(out, "model   work %s (%s) · review %s (%s)\n", work.Model, work.ModelSource, review.Model, review.ModelSource)
	fmt.Fprintf(out, "effort  work %s (%s) · review %s (%s)\n", effortText(work), work.EffortSource, effortText(review), review.EffortSource)
	return nil
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func writeList(out io.Writer, label string, items []string) {
	if len(items) == 0 {
		fmt.Fprintf(out, "%s: -\n", label)
		return
	}
	fmt.Fprintf(out, "%s:\n", label)
	for _, it := range items {
		fmt.Fprintf(out, "  - %s\n", it)
	}
}
