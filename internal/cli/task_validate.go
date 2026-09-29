package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/state"
)

func newTaskValidate(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Check the plan's structure",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			root, err := g.root()
			if err != nil {
				return err
			}
			areas, err := config.Get(root, "areas")
			if err != nil {
				return configErr(g, out, err)
			}
			rep, err := state.Check(root, areas.List)
			if err != nil {
				return jsonProblem(g, out, err)
			}
			return writeReport(g, out, rep)
		},
	}
}

func writeReport(g *Globals, out io.Writer, rep *state.Report) error {
	n := len(rep.Problems)
	if g.JSON {
		enc := json.NewEncoder(out)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(struct {
			OK       bool     `json:"ok"`
			Problems []string `json:"problems"`
			Warnings []string `json:"warnings"`
		}{n == 0, rep.Problems, rep.Warnings}); err != nil {
			return err
		}
	} else {
		for _, p := range rep.Problems {
			fmt.Fprintf(out, "✗ %s\n", p)
		}
		for _, w := range rep.Warnings {
			fmt.Fprintf(out, "! %s\n", w)
		}
		if n == 0 {
			fmt.Fprintln(out, "plan ok")
		} else {
			fmt.Fprintf(out, "%d problem(s)\n", n)
		}
	}
	if n > 0 {
		return Problem(errors.New(""))
	}
	return nil
}
