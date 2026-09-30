package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/classify"
	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/defect"
	"github.com/mvelosop/vloop/internal/metrics"
)

func newMetrics(g *Globals) *cobra.Command {
	var by, workspace string
	cmd := &cobra.Command{
		Use:   "metrics [<brief>…]",
		Short: "Summarise what a brief cost and delivered, from its runs and commits",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if by != "" && by != "task" {
				return fmt.Errorf("unknown --by %q: want task", by)
			}
			out := cmd.OutOrStdout()
			if workspace != "" {
				return workspaceMetrics(g, out, cmd.ErrOrStderr(), workspace, args)
			}
			root, err := g.root()
			if err != nil {
				return err
			}
			c, err := newClassifier(g, out, root)
			if err != nil {
				return err
			}
			var reports []*metrics.Report
			for _, b := range args {
				r, err := metrics.Build(root, b, c)
				if err != nil {
					return Problem(err)
				}
				if r == nil {
					return Problem(fmt.Errorf("no runs for %s", defect.BriefName(b)))
				}
				reports = append(reports, r)
			}
			if len(args) == 0 {
				if reports, err = allReports(root, c); err != nil {
					return Problem(err)
				}
			}
			if g.JSON {
				if len(args) == 1 {
					return json.NewEncoder(out).Encode(reports[0])
				}
				if reports == nil {
					reports = []*metrics.Report{}
				}
				return json.NewEncoder(out).Encode(reports)
			}
			switch {
			case len(args) == 0:
				metrics.PrintBriefTable(out, reports)
			case by == "task":
				for i, r := range reports {
					if i > 0 {
						fmt.Fprintln(out)
					}
					metrics.PrintByTask(out, r)
				}
			default:
				for i, r := range reports {
					if i > 0 {
						fmt.Fprintln(out)
					}
					metrics.PrintSummary(out, r)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&by, "by", "", "break the summary down by `task`")
	cmd.Flags().StringVar(&workspace, "workspace", "", "show every repository the workspace `file` lists, with a repo column")
	cmd.AddCommand(newMetricsStacks(g), newMetricsClassify(g), newMetricsExport(g))
	return cmd
}

// allReports reports every brief in docs/briefs that has runs, in name order.
func allReports(root string, c *classify.Classifier) ([]*metrics.Report, error) {
	ents, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(defect.BriefsDir)))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	var out []*metrics.Report
	for _, e := range ents {
		n := e.Name()
		if e.IsDir() || !strings.HasSuffix(n, ".loop-brief.md") {
			continue
		}
		r, err := metrics.Build(root, defect.BriefsDir+"/"+n, c)
		if err != nil {
			return nil, err
		}
		if r != nil {
			out = append(out, r)
		}
	}
	return out, nil
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

// newClassifier builds the classifier from the repo's metrics.* config keys.
func newClassifier(g *Globals, out io.Writer, root string) (*classify.Classifier, error) {
	get := func(k string) ([]string, error) {
		v, err := config.Get(root, k)
		return v.List, err
	}
	var repo classify.Preset
	for _, r := range []struct {
		key string
		dst *[]string
	}{{"metrics.excluded", &repo.Excluded}, {"metrics.test", &repo.Test}, {"metrics.docs", &repo.Docs}, {"metrics.code", &repo.Code}} {
		var err error
		if *r.dst, err = get(r.key); err != nil {
			return nil, configErr(g, out, err)
		}
	}
	stacks, err := get("metrics.stacks")
	if err != nil {
		return nil, configErr(g, out, err)
	}
	return classify.New(repo, stacks), nil
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
			c, err := newClassifier(g, out, root)
			if err != nil {
				return err
			}
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
