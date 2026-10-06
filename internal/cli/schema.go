package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/schema"
)

func newSchema(g *Globals) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "List, show and validate against the embedded JSON Schemas",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List the schema names",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			names := schema.Names()
			if g.JSON {
				return json.NewEncoder(out).Encode(names)
			}
			for _, n := range names {
				fmt.Fprintln(out, n)
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "show <name>",
		Short: "Show a schema document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := schema.Document(args[0])
			if err != nil {
				return Usage(err)
			}
			_, err = cmd.OutOrStdout().Write(doc)
			return err
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "validate <name> <file>",
		Short: "Validate a JSON file against a schema",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, path := args[0], args[1]
			if _, err := schema.Document(name); err != nil {
				return Usage(err)
			}
			read := path
			if g.Dir != "" && !filepath.IsAbs(read) {
				read = filepath.Join(g.Dir, read)
			}
			data, err := os.ReadFile(read)
			if err != nil {
				return Problem(err)
			}
			vs, err := schema.Validate(name, data)
			if err != nil {
				return Problem(err)
			}
			out := cmd.OutOrStdout()
			if g.JSON {
				if vs == nil {
					vs = []schema.Violation{}
				}
				if err := json.NewEncoder(out).Encode(struct {
					OK     bool               `json:"ok"`
					Errors []schema.Violation `json:"errors"`
				}{len(vs) == 0, vs}); err != nil {
					return err
				}
			} else if len(vs) == 0 {
				fmt.Fprintf(out, "%s: ok\n", path)
			} else {
				for _, v := range vs {
					if v.Pointer == "" {
						fmt.Fprintf(out, "%s: %s\n", path, v.Message)
						continue
					}
					fmt.Fprintf(out, "%s: %s: %s\n", path, v.Pointer, v.Message)
				}
			}
			if len(vs) > 0 {
				return Problem(errors.New(""))
			}
			return nil
		},
	})
	return cmd
}
