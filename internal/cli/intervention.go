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
	cmd.AddCommand(newInterventionAdd(g), newInterventionList(g), newInterventionSet(g), newInterventionMigrate(g), newInterventionShow(g))
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
				return Usage(errors.New("the summary must not be empty"))
			}
			for _, v := range []struct{ f, v string }{{"phase", in.Phase}, {"kind", in.Kind}, {"automatable", in.Automatable}, {"by", in.By}} {
				if v.v == "" {
					return Usage(fmt.Errorf("--%s is required", v.f))
				}
			}
			for _, v := range []struct{ f, v string }{{"phase", in.Phase}, {"kind", in.Kind}, {"automatable", in.Automatable}, {"by", in.By}} {
				if err := intervention.Validate(v.f, v.v); err != nil {
					return err
				}
			}
			in.RecommendedSet = cmd.Flags().Changed("recommended")
			in.DecidedOptionSet = cmd.Flags().Changed("decided-option")
			if err := intervention.CheckOptions(in); err != nil {
				return Usage(err)
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
	f.StringVar(&in.Context, "context", "", "the situation, and the ids it refers to")
	f.StringArrayVar(&in.Options, "option", nil, "an option put to the operator (repeatable, at most three)")
	f.IntVar(&in.Recommended, "recommended", 0, "the recommended option, 1 to the number of options")
	f.StringVar(&in.Why, "why", "", "why that option is recommended")
	f.IntVar(&in.DecidedOption, "decided-option", 0, "the option decided, 1 to the number of options")
	f.BoolVar(&in.DecidedOther, "decided-other", false, "something other than the options was decided")
	f.BoolVar(&in.Adjusted, "adjusted", false, "the decided option was adjusted")
	f.StringVar(&in.Decided, "decided", "", "what was decided")
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
				fmt.Fprintf(out, "%s  %s  %s  %s  %s\n", v.Phase, v.Kind, v.Automatable, v.Agreement, v.ID)
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
				var cu *config.UsageError
				if errors.As(err, &inv) || errors.As(err, &cu) || !contains(intervention.SetFields, args[1]) {
					return Usage(err)
				}
				return Problem(err)
			}
			return nil
		},
	}
}

func newInterventionMigrate(g *Globals) *cobra.Command {
	var dry bool
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Move v1 intervention records to intervention/v2, inserting the no-options frontmatter lines and changing nothing else",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := g.root()
			if err != nil {
				return err
			}
			names, err := intervention.Migrate(root, dry)
			if err != nil {
				return Problem(err)
			}
			out := cmd.OutOrStdout()
			if len(names) == 0 {
				fmt.Fprintln(out, "nothing to migrate")
				return nil
			}
			if dry {
				for _, n := range names {
					fmt.Fprintln(out, n)
				}
				return nil
			}
			fmt.Fprintf(out, "migrated %d record(s)\n", len(names))
			return nil
		},
	}
	cmd.Flags().BoolVar(&dry, "dry-run", false, "list the records that would be migrated and write nothing")
	return cmd
}

func newInterventionShow(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Print an intervention and resolve the ids named in its Context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := g.root()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			v, err := intervention.Get(root, args[0])
			if err != nil {
				var none *intervention.NoInterventionError
				if errors.As(err, &none) {
					return jsonProblem(g, out, err)
				}
				return Problem(err)
			}
			s := intervention.Shown{Intervention: v, Links: intervention.Links(root, v)}
			if g.JSON {
				return json.NewEncoder(out).Encode(s)
			}
			fmt.Fprint(out, s.Text())
			return nil
		},
	}
}
