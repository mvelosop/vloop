package cli

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/state"
)

var taskKinds = []string{"feature", "fix", "refactor", "test", "docs", "chore"}

var settableFields = []string{"area", "kind", "model.work", "model.review", "effort.work", "effort.review"}

// amend applies change to the plan, mapping the amend errors: an unknown task
// or an invalid plan or result is a problem (exit 1); anything else passes
// through (usage errors stay exit 2).
func amend(g *Globals, cmd *cobra.Command, change func(*state.Plan) error) error {
	out := cmd.OutOrStdout()
	root, err := g.root()
	if err != nil {
		return err
	}
	areas, err := config.Get(root, "areas")
	if err != nil {
		return configErr(g, out, err)
	}
	err = state.Amend(root, areas.List, func(p *state.Plan) error {
		err := change(p)
		var inv *config.InvalidValueError
		if err != nil && !errors.As(err, &inv) {
			err = Problem(err)
		}
		return err
	})
	if err == nil {
		return nil
	}
	var pe *ProblemError
	var inv *config.InvalidValueError
	if errors.As(err, &pe) || errors.As(err, &inv) {
		return err
	}
	return jsonProblem(g, out, err)
}

func newTaskAmend(g *Globals) []*cobra.Command {
	reset := &cobra.Command{
		Use:   "reset <id>",
		Short: "Set a task back to pending with no attempts",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return amend(g, cmd, func(p *state.Plan) error { return state.Reset(p, args[0]) })
		},
	}
	note := &cobra.Command{
		Use:   "note <id> <text>",
		Short: "Replace a task's notes",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return amend(g, cmd, func(p *state.Plan) error { return state.SetNote(p, args[0], args[1]) })
		},
	}
	drop := &cobra.Command{
		Use:   "drop <id>",
		Short: "Remove a task nothing depends on",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return amend(g, cmd, func(p *state.Plan) error { return state.Drop(p, args[0]) })
		},
	}
	set := &cobra.Command{
		Use:   "set <id> <field> <value>",
		Short: "Set or clear one field of a task: " + strings.Join(settableFields, ", "),
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, field, value := args[0], args[1], args[2]
			if !slices.Contains(settableFields, field) {
				return Usage(fmt.Errorf("cannot set %q: want one of %s", field, strings.Join(settableFields, ", ")))
			}
			root, err := g.root()
			if err != nil {
				return err
			}
			return amend(g, cmd, func(p *state.Plan) error {
				if p.Find(id) == nil {
					return &state.NoTaskError{ID: id}
				}
				if err := checkFieldValue(root, field, value); err != nil {
					return err
				}
				return state.SetField(p, id, field, value)
			})
		},
	}
	return []*cobra.Command{reset, note, drop, set}
}

// checkFieldValue validates a non-empty value for a settable field.
func checkFieldValue(root, field, value string) error {
	if value == "" {
		return nil
	}
	switch field {
	case "kind":
		if !slices.Contains(taskKinds, value) {
			return &config.InvalidValueError{Key: field, Value: value, Valid: taskKinds}
		}
	case "effort.work", "effort.review":
		k, err := config.Lookup(field)
		if err != nil {
			return err
		}
		return k.Validate(value)
	case "area":
		areas, err := config.Get(root, "areas")
		if err != nil {
			return err
		}
		if len(areas.List) > 0 && !slices.Contains(areas.List, value) {
			return &config.InvalidValueError{Key: field, Value: value, Valid: areas.List}
		}
	}
	return nil
}
