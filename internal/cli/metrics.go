package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/classify"
	"github.com/mvelosop/vloop/internal/config"
)

func newMetrics(g *Globals) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Line classification and brief metrics",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	cmd.AddCommand(newMetricsStacks(g), newMetricsClassify(g))
	return cmd
}

func newMetricsStacks(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "stacks [name]",
		Short: "Print the built-in stack presets",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			names := classify.Names()
			if len(args) == 1 {
				if _, ok := classify.Lookup(args[0]); !ok {
					return fmt.Errorf("unknown stack %q", args[0])
				}
				names = args
			}
			if g.JSON {
				all := map[string]classify.Preset{}
				for _, n := range names {
					all[n], _ = classify.Lookup(n)
				}
				return json.NewEncoder(out).Encode(all)
			}
			if len(args) == 0 {
				fmt.Fprintln(out, strings.Join(names, "\n"))
				return nil
			}
			p, _ := classify.Lookup(args[0])
			for _, sec := range []struct {
				name  string
				globs []string
			}{{"code", p.Code}, {"test", p.Test}, {"docs", p.Docs}, {"excluded", p.Excluded}} {
				fmt.Fprintf(out, "%s:\n", sec.name)
				for _, gl := range sec.globs {
					fmt.Fprintf(out, "  %s\n", gl)
				}
			}
			return nil
		},
	}
}

type classifyLine struct {
	Path     string  `json:"path"`
	Category string  `json:"category"`
	Layer    *string `json:"layer"`
	Glob     *string `json:"glob"`
}

func newMetricsClassify(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "classify <path>…",
		Short: "Show how paths are classified and by which glob",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			root, err := g.root()
			if err != nil {
				return err
			}
			get := func(k string) ([]string, error) {
				v, err := config.Get(root, k)
				return v.List, err
			}
			var repo classify.Preset
			for _, r := range []struct {
				key string
				dst *[]string
			}{{"metrics.excluded", &repo.Excluded}, {"metrics.test", &repo.Test}, {"metrics.docs", &repo.Docs}, {"metrics.code", &repo.Code}} {
				if *r.dst, err = get(r.key); err != nil {
					return configErr(g, out, err)
				}
			}
			stacks, err := get("metrics.stacks")
			if err != nil {
				return configErr(g, out, err)
			}
			c := classify.New(repo, stacks)
			var lines []classifyLine
			for _, p := range args {
				r := c.Classify(strings.ReplaceAll(p, "\\", "/"))
				l := classifyLine{Path: p, Category: r.Category}
				if r.Layer != "" {
					l.Layer, l.Glob = &r.Layer, &r.Glob
				}
				lines = append(lines, l)
				if !g.JSON {
					if r.Layer == "" {
						fmt.Fprintf(out, "%s  %s  (none)\n", p, r.Category)
					} else {
						fmt.Fprintf(out, "%s  %s  (%s: %s)\n", p, r.Category, r.Layer, r.Glob)
					}
				}
			}
			if g.JSON {
				return json.NewEncoder(out).Encode(lines)
			}
			return nil
		},
	}
}
