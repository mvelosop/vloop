package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/brief"
	"github.com/mvelosop/vloop/internal/config"
)

func newBriefNew(g *Globals) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "new <slug>",
		Short: "Write a draft brief from the template",
		Args:  cobra.ExactArgs(1),
		RunE:  func(cmd *cobra.Command, args []string) error { return runBriefNew(g, cmd, args[0], dryRun, time.Now()) },
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the path and write nothing")
	return cmd
}

func runBriefNew(g *Globals, cmd *cobra.Command, slug string, dryRun bool, now time.Time) error {
	out := cmd.OutOrStdout()
	if !brief.ValidSlug(slug) {
		return fmt.Errorf("invalid slug %q: want lower-case letters, digits and single hyphens", slug)
	}
	root, err := g.root()
	if err != nil {
		return err
	}
	lang, err := config.Get(root, "language")
	if err != nil {
		return configErr(g, out, err)
	}
	rel := brief.NewPath(slug, now)
	report := func(created bool) error {
		if g.JSON {
			return writeJSON(out, struct {
				Path    string `json:"path"`
				Created bool   `json:"created"`
			}{rel, created})
		}
		_, err := fmt.Fprintln(out, rel)
		return err
	}
	if dryRun {
		return report(false)
	}
	text, err := brief.Render(lang.Value, slug, now)
	if err != nil {
		return Problem(err)
	}
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return Problem(err)
	}
	f, err := os.OpenFile(abs, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			msg := "brief exists: " + rel
			if g.JSON {
				_ = writeJSON(out, struct {
					Error string `json:"error"`
				}{msg})
			}
			return Problem(errors.New(msg))
		}
		return Problem(err)
	}
	if _, err := f.WriteString(text); err != nil {
		f.Close()
		return Problem(err)
	}
	if err := f.Close(); err != nil {
		return Problem(err)
	}
	return report(true)
}
