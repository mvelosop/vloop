package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/config"
)

// root resolves the repo root from -C or the working directory.
func (g *Globals) root() (string, error) {
	dir := g.Dir
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		dir = wd
	}
	return config.Root(dir), nil
}

// configErr maps config errors: bad source values are problems (exit 1), and
// under --json stdout still carries one JSON document.
func configErr(g *Globals, out io.Writer, err error) error {
	var se *config.SourceError
	if errors.As(err, &se) {
		if g.JSON {
			_ = json.NewEncoder(out).Encode(struct {
				Error string `json:"error"`
			}{err.Error()})
		}
		return Problem(err)
	}
	return err
}

func jsonValue(v config.Value) any {
	if k, err := config.Lookup(v.Key); err == nil && k.List {
		return v.List
	}
	if !v.Set {
		return nil
	}
	return v.Value
}

func newConfig(g *Globals) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Read and write repo-local configuration",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "Print every key with its value and source",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			root, err := g.root()
			if err != nil {
				return err
			}
			vals, err := config.List(root)
			if err != nil {
				return configErr(g, out, err)
			}
			if g.JSON {
				return writeOrdered(out, vals, func(v config.Value) (string, any) {
					return v.Key, struct {
						Value  any    `json:"value"`
						Source string `json:"source"`
					}{jsonValue(v), v.Source}
				})
			}
			for _, v := range vals {
				fmt.Fprintf(out, "%s=%s (%s)\n", v.Key, v.Value, v.Source)
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "get <key>",
		Short: "Print the resolved value of a key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			root, err := g.root()
			if err != nil {
				return err
			}
			v, err := config.Get(root, args[0])
			if err != nil {
				return configErr(g, out, err)
			}
			if g.JSON {
				return json.NewEncoder(out).Encode(struct {
					Key    string `json:"key"`
					Value  any    `json:"value"`
					Source string `json:"source"`
				}{v.Key, jsonValue(v), v.Source})
			}
			fmt.Fprintln(out, v.Value)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "set <key> <value>",
		Short: "Write a key to .vloop/config.toml ('' removes it)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := g.root()
			if err != nil {
				return err
			}
			return configErr(g, cmd.OutOrStdout(), config.Set(root, args[0], args[1]))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Print the config file path, relative to the repo root",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), config.FilePath)
			return nil
		},
	})
	return cmd
}

// writeOrdered emits a JSON object whose keys keep the order of vals.
func writeOrdered[T any](w io.Writer, vals []T, kv func(T) (string, any)) error {
	buf := []byte{'{'}
	for i, v := range vals {
		k, x := kv(v)
		kb, _ := json.Marshal(k)
		xb, err := json.Marshal(x)
		if err != nil {
			return err
		}
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = append(append(append(buf, kb...), ':'), xb...)
	}
	buf = append(buf, '}', '\n')
	_, err := w.Write(buf)
	return err
}
