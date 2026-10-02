package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/intervention"
)

func newIntervention(g *Globals) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "intervention",
		Short: "Record, list and update operator interventions",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	cmd.AddCommand(newInterventionAdd(g), newInterventionList(g), newInterventionSet(g))
	return cmd
}

func newInterventionAdd(g *Globals) *cobra.Command {
	var in intervention.NewInput
	cmd := &cobra.Command{
		Use:   `add "<summary>"`,
		Short: "Record an intervention as .vloop/interventions/I<stamp>-<slug>.md and print its path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			in.Summary = args[0]
			if strings.TrimSpace(in.Summary) == "" {
				return errors.New("the summary must not be empty")
			}
			for _, v := range []struct{ f, v string }{{"phase", in.Phase}, {"kind", in.Kind}, {"automatable", in.Automatable}, {"by", in.By}} {
				if v.v == "" {
					return fmt.Errorf("--%s is required", v.f)
				}
			}
			for _, v := range []struct{ f, v string }{{"phase", in.Phase}, {"kind", in.Kind}, {"automatable", in.Automatable}, {"by", in.By}} {
				if err := intervention.Validate(v.f, v.v); err != nil {
					return err
				}
			}
			root, err := g.root()
			if err != nil {
				return err
			}
			p, err := intervention.Add(root, in, time.Now())
			if err != nil {
				return Problem(err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), p)
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&in.Phase, "phase", "", "lifecycle phase: "+strings.Join(intervention.Phases, ", "))
	f.StringVar(&in.Kind, "kind", "", "kind: "+strings.Join(intervention.Kinds, ", "))
	f.StringVar(&in.Automatable, "automatable", "", "could a driver do it: "+strings.Join(intervention.Automatables, ", "))
	f.StringVar(&in.By, "by", "", "who did it: "+strings.Join(intervention.Bys, ", "))
	f.StringVar(&in.Brief, "brief", "", "the loop brief it belongs to (none: series-level)")
	f.StringVar(&in.Trigger, "trigger", "", "what made it necessary")
	f.StringVar(&in.Done, "done", "", "what was done")
	f.StringVar(&in.Automation, "automation", "", "what would automate it")
	return cmd
}

func newInterventionList(g *Globals) *cobra.Command {
	var brief string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Print the recorded interventions by phase, kind and id",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := g.root()
			if err != nil {
				return err
			}
			vs, err := intervention.List(root, brief)
			if err != nil {
				return Problem(err)
			}
			out := cmd.OutOrStdout()
			if g.JSON {
				return json.NewEncoder(out).Encode(vs)
			}
			for _, v := range vs {
				fmt.Fprintf(out, "%s  %s  %s  %s\n", v.Phase, v.Kind, v.Automatable, v.ID)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&brief, "brief", "", "only interventions of this loop brief")
	return cmd
}

func newInterventionSet(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "set <id> <field> <value>",
		Short: "Set " + strings.Join(intervention.SetFields, ", ") + " of an intervention",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := g.root()
			if err != nil {
				return err
			}
			if err := intervention.Set(root, args[0], args[1], args[2]); err != nil {
				var inv *config.InvalidValueError
				if errors.As(err, &inv) || !contains(intervention.SetFields, args[1]) {
					return err
				}
				return Problem(err)
			}
			return nil
		},
	}
}
