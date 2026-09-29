package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/brief"
)

type listJSON struct {
	Name      string   `json:"name"`
	Path      string   `json:"path"`
	Status    string   `json:"status"`
	Ready     string   `json:"ready"`
	DependsOn []string `json:"depends_on"`
}

func newBriefList(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List briefs in dependency order",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := g.root()
			if err != nil {
				return err
			}
			entries, err := brief.Load(root)
			if err != nil {
				return Problem(err)
			}
			ordered, err := brief.Order(entries)
			if err != nil {
				return Problem(errors.New(err.Error()))
			}
			out := cmd.OutOrStdout()
			if g.JSON {
				docs := make([]listJSON, len(ordered))
				for i, e := range ordered {
					docs[i] = listJSON{e.Name, e.Path, e.Status, e.Ready, e.DependsOn}
				}
				return writeJSON(out, docs)
			}
			for _, e := range ordered {
				deps := "-"
				if len(e.DependsOn) > 0 {
					deps = strings.Join(e.DependsOn, ",")
				}
				fmt.Fprintf(out, "%s  %s  %s  %s\n", e.Name, e.Status, e.Ready, deps)
			}
			return nil
		},
	}
}
