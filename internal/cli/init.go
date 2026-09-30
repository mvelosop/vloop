package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/brief"
	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/detect"
	"github.com/mvelosop/vloop/internal/install"
)

func newInit(b Build, g *Globals) *cobra.Command {
	var language, stacks string
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Set a repository up for vloop",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runInit(b, g, cmd, initOpts{
				language: language, languageSet: cmd.Flags().Changed("language"),
				stacks: stacks, stacksSet: cmd.Flags().Changed("stacks"), dryRun: dryRun,
			}, time.Now())
		},
	}
	cmd.Flags().StringVar(&language, "language", "en", "language of briefs and guidance: en or es")
	cmd.Flags().StringVar(&stacks, "stacks", "", "metrics.stacks to write, instead of the detected ones")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print what would be written and write nothing")
	return cmd
}

type initOpts struct {
	language, stacks       string
	languageSet, stacksSet bool
	dryRun                 bool
}

// initStep is one decided write: its path, what happens and the bytes to store.
type initStep struct {
	path, verb string // verb: wrote, updated or kept
	content    []byte
	write      func() error // set when the step is not a plain file write
}

func runInit(b Build, g *Globals, cmd *cobra.Command, o initOpts, now time.Time) error {
	out := cmd.OutOrStdout()
	root, err := g.root()
	if err != nil {
		return Problem(err)
	}
	if _, err := os.Lstat(filepath.Join(root, ".git")); err != nil {
		return Problem(errors.New("not a git repository — run git init first"))
	}
	switch st, err := install.Read(root); {
	case err == nil:
		return Problem(fmt.Errorf("already initialized by vloop %s — run vloop upgrade", st.Version))
	case !errors.Is(err, os.ErrNotExist):
		return Problem(err)
	}

	lang := "en"
	if o.languageSet {
		lang = o.language
		if err := config.Check(root, "language", lang); err != nil {
			return err
		}
	}
	if o.stacksSet {
		if err := config.Check(root, "metrics.stacks", o.stacks); err != nil {
			return err
		}
	}
	_, cfgErr := os.Lstat(filepath.Join(root, filepath.FromSlash(config.FilePath)))
	haveConfig := cfgErr == nil
	if haveConfig {
		v, err := config.Get(root, "language")
		if err != nil {
			return configErr(g, out, err)
		}
		if !o.languageSet {
			lang = v.Value
		}
	}

	detected, err := detect.Stacks(root)
	if err != nil {
		return Problem(err)
	}
	if len(detected) == 0 {
		fmt.Fprintln(out, "detected stacks: none")
	} else {
		fmt.Fprintln(out, "detected stacks: "+strings.Join(detected, ", "))
	}
	chosen := strings.Join(detected, ",")
	if o.stacksSet {
		chosen = o.stacks
	}

	var steps []initStep
	// 1. config
	if haveConfig {
		steps = append(steps, initStep{path: config.FilePath, verb: "kept"})
		if cur, err := config.Get(root, "metrics.stacks"); err == nil {
			var missing []string
			for _, d := range detected {
				found := false
				for _, c := range cur.List {
					found = found || c == d
				}
				if !found {
					missing = append(missing, d)
				}
			}
			if len(missing) > 0 {
				fmt.Fprintln(out, "suggest metrics.stacks: "+strings.Join(missing, ","))
			}
		}
	} else {
		steps = append(steps, initStep{path: config.FilePath, verb: "wrote", write: func() error {
			if err := config.Set(root, "language", lang); err != nil {
				return err
			}
			if chosen == "" {
				return nil
			}
			return config.Set(root, "metrics.stacks", chosen)
		}})
	}
	// 2. stamp
	stamp := install.Stamp{
		Version: b.Version, Commit: b.Commit,
		Initialized: now.UTC().Format("2006-01-02T15:04:05Z"),
	}
	steps = append(steps, initStep{path: install.Path, verb: "wrote", write: func() error { return install.Write(root, stamp) }})
	// 3. starter brief
	hasBrief := false
	if ents, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(brief.Dir))); err == nil {
		for _, e := range ents {
			hasBrief = hasBrief || strings.HasSuffix(e.Name(), brief.Suffix)
		}
	}
	if hasBrief {
		steps = append(steps, initStep{path: brief.Dir + "/", verb: "kept"})
	} else {
		slug := "getting-started"
		if lang == "es" {
			slug = "primeros-pasos"
		}
		text, err := brief.Render(lang, slug, now)
		if err != nil {
			return Problem(err)
		}
		steps = append(steps, initStep{path: brief.NewPath(slug, now), verb: "wrote", content: []byte(text)})
	}
	// 4. .gitignore
	old, exists, err := readRepoFile(root, ".gitignore")
	if err != nil {
		return Problem(err)
	}
	if merged, changed := install.MergeGitignore(old); !changed {
		steps = append(steps, initStep{path: ".gitignore", verb: "kept"})
	} else if exists {
		steps = append(steps, initStep{path: ".gitignore", verb: "updated", content: merged})
	} else {
		steps = append(steps, initStep{path: ".gitignore", verb: "wrote", content: merged})
	}
	// 5. CLAUDE.md
	old, exists, err = readRepoFile(root, "CLAUDE.md")
	if err != nil {
		return Problem(err)
	}
	merged := install.MergeClaudeMD(old, exists, b.Version)
	switch {
	case exists && string(merged) == string(old):
		steps = append(steps, initStep{path: "CLAUDE.md", verb: "kept"})
	case exists:
		steps = append(steps, initStep{path: "CLAUDE.md", verb: "updated", content: merged})
	default:
		steps = append(steps, initStep{path: "CLAUDE.md", verb: "wrote", content: merged})
	}

	// Every target is checked for a symlink before anything is written.
	for _, s := range steps {
		if s.verb != "kept" {
			if err := checkRepoPath(root, s.path); err != nil {
				return Problem(err)
			}
		}
	}
	for _, s := range steps {
		if !o.dryRun && s.verb != "kept" {
			if s.write != nil {
				if err := s.write(); err != nil {
					return Problem(err)
				}
			} else if err := writeRepoFile(root, s.path, s.content); err != nil {
				return Problem(err)
			}
		}
		verb := s.verb
		if o.dryRun {
			verb = map[string]string{"wrote": "would write", "updated": "would update", "kept": "would keep"}[verb]
		}
		fmt.Fprintf(out, "%s %s\n", verb, s.path)
	}
	if !o.dryRun {
		fmt.Fprintln(out, "next:")
		fmt.Fprintln(out, "  claude plugin marketplace add mvelosop/vloop")
		fmt.Fprintln(out, "  claude plugin install vloop@vloop")
		fmt.Fprintln(out, "  vloop doctor")
	}
	return nil
}

// checkRepoPath refuses a repo-relative path that passes through a symlink or
// through something that is not a directory.
func checkRepoPath(root, rel string) error {
	cur := root
	parts := strings.Split(strings.TrimSuffix(rel, "/"), "/")
	for i, p := range parts {
		cur = filepath.Join(cur, p)
		fi, err := os.Lstat(cur)
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		if fi.Mode()&os.ModeSymlink != 0 || (i < len(parts)-1 && !fi.IsDir()) {
			return fmt.Errorf("%s is a symlink or not a directory; vloop does not write through it", strings.Join(parts[:i+1], "/"))
		}
	}
	return nil
}

// readRepoFile reads a file at a repo-relative path, refusing a symlink.
func readRepoFile(root, rel string) ([]byte, bool, error) {
	if err := checkRepoPath(root, rel); err != nil {
		return nil, false, err
	}
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	return data, err == nil, err
}

func writeRepoFile(root, rel string, content []byte) error {
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	return os.WriteFile(abs, content, 0o644)
}
