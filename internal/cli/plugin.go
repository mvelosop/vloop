package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/driver"
)

func newPlugin(b Build, g *Globals) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Work with the plugin embedded in the binary",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Extract the embedded plugin under .vloop/tmp/plugin/ and print its path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := g.root()
			if err != nil {
				return Problem(err)
			}
			rel, err := driver.ExtractPlugin(root, b.Version)
			if err != nil {
				return Problem(err)
			}
			out := cmd.OutOrStdout()
			if g.JSON {
				return json.NewEncoder(out).Encode(struct {
					Path string `json:"path"`
				}{rel})
			}
			_, err = fmt.Fprintln(out, rel)
			return err
		},
	})
	return cmd
}
