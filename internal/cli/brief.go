package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/brief"
	"github.com/mvelosop/vloop/internal/config"
)

func newBrief(g *Globals) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "brief",
		Short: "Check loop briefs",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "check <path>...",
		Short: "Check that briefs are fit to plan",
		Args:  cobra.MinimumNArgs(1),
		RunE:  func(cmd *cobra.Command, args []string) error { return runBriefCheck(g, cmd, args) },
	})
	return cmd
}

type briefJSON struct {
	Path     string   `json:"path"`
	Result   string   `json:"result"`
	Problems []string `json:"problems"`
	Warnings []string `json:"warnings"`
}

func runBriefCheck(g *Globals, cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	root, err := g.root()
	if err != nil {
		return err
	}
	start := g.Dir
	if start == "" {
		if start, err = os.Getwd(); err != nil {
			return err
		}
	}
	// Every path is validated before any brief is checked.
	rels := make([]string, len(args))
	for i, a := range args {
		abs := a
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(start, abs)
		}
		rel, err := filepath.Rel(root, abs)
		if err != nil {
			rel = abs
		}
		rels[i] = filepath.ToSlash(rel)
		if !strings.HasSuffix(rels[i], brief.Suffix) {
			return fmt.Errorf("not a loop brief: %s", rels[i])
		}
	}
	lang, err := config.Get(root, "language")
	if err != nil {
		return configErr(g, out, err)
	}
	set := brief.SetFor(lang.Value)

	failed := 0
	docs := []briefJSON{}
	color := g.Color(out)
	paint := func(code, s string) string {
		if !color {
			return s
		}
		return "\033[" + code + "m" + s + "\033[0m"
	}
	for _, rel := range rels {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return Problem(err)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "vloop: no such brief: %s\n", rel)
			failed++
			docs = append(docs, briefJSON{rel, "problems", []string{"no such brief: " + rel}, []string{}})
			continue
		}
		res := brief.Check(root, brief.Parse(rel, string(data)), set)
		j := briefJSON{rel, "ok", res.Problems(), res.Warnings()}
		switch {
		case res.Skipped:
			j.Result = "skipped"
		case res.Failed():
			j.Result = "problems"
			failed++
		}
		docs = append(docs, j)
		if g.JSON {
			continue
		}
		fmt.Fprintln(out, paint("1", rel))
		if res.Skipped {
			fmt.Fprintf(out, "  %s skipped: status is %s\n", paint("36", "-"), res.Status)
			continue
		}
		for _, l := range res.Lines {
			code := map[brief.Marker]string{brief.Pass: "32", brief.Warning: "33", brief.Problem: "31"}[l.Marker]
			fmt.Fprintf(out, "  %s %s\n", paint(code, string(l.Marker)), l.Text)
		}
		fmt.Fprintf(out, "  %s\n", res.Summary())
	}

	if g.JSON {
		if err := writeJSON(out, struct {
			OK     bool        `json:"ok"`
			Briefs []briefJSON `json:"briefs"`
		}{failed == 0, docs}); err != nil {
			return err
		}
	} else if failed == 0 {
		fmt.Fprintln(out, "briefs ok")
	} else {
		fmt.Fprintf(out, "%d brief(s) need work\n", failed)
	}
	if failed > 0 {
		return Problem(errors.New("")) // already reported above
	}
	return nil
}

func writeJSON(w io.Writer, v any) error { return json.NewEncoder(w).Encode(v) }
