package cli

import (
	"errors"
	"fmt"
	"io"
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

	lang, haveConfig, err := initLanguage(g, out, root, o)
	if err != nil {
		return err
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

	starter := detect.StarterChecks(strings.Split(chosen, ","))
	var steps []initStep
	// 1. config
	steps = append(steps, configStep(out, root, haveConfig, detected, chosen, lang, starter))
	// 2. stamp
	stamp := install.Stamp{
		Version: b.Version, Commit: b.Commit,
		Initialized: now.UTC().Format("2006-01-02T15:04:05Z"),
	}
	steps = append(steps, initStep{path: install.Path, verb: "wrote", write: func() error { return install.Write(root, stamp) }})
	// 2b. repository guidelines: written once, never touched or reported when present
	if _, exists, err := readRepoFile(root, guidelinesPath); err != nil {
		return Problem(err)
	} else if !exists {
		steps = append(steps, initStep{path: guidelinesPath, verb: "wrote", content: []byte(guidelinesTemplate(lang))})
	}
	// 3. starter brief
	briefStep, err := starterBriefStep(root, lang, now)
	if err != nil {
		return Problem(err)
	}
	steps = append(steps, briefStep)
	// 4. .gitignore
	old, exists, err := readRepoFile(root, ".gitignore")
	if err != nil {
		return Problem(err)
	}
	steps = append(steps, mergedStep(".gitignore", old, exists, func() ([]byte, bool) { return install.MergeGitignore(old) }))
	// 5. CLAUDE.md
	old, exists, err = readRepoFile(root, "CLAUDE.md")
	if err != nil {
		return Problem(err)
	}
	claude := install.MergeClaudeMD(old, exists, b.Version)
	steps = append(steps, mergedStep("CLAUDE.md", old, exists, func() ([]byte, bool) { return claude, string(claude) != string(old) }))

	// Every target is checked for a symlink before anything is written.
	for _, s := range steps {
		if s.verb != "kept" {
			if err := checkRepoPath(root, s.path); err != nil {
				return Problem(err)
			}
		}
	}
	if err := applySteps(out, root, steps, o.dryRun); err != nil {
		return err
	}
	if !o.dryRun {
		printInitNext(out)
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

// guidelinesPath is the repository's lessons file; every loop session receives it.
const guidelinesPath = ".vloop/repo-guidelines.md"

// guidelinesTemplate is the empty guidelines file. The two headings are the
// same in every language because vloop brief check reads them; the examples
// sit inside HTML comments so they are not rules.
func guidelinesTemplate(lang string) string {
	if lang == "es" {
		return `# Lecciones de este repositorio

Este archivo guarda las lecciones de este repositorio y todas las sesiones del bucle lo reciben.

## Pinning tests

Las pruebas que fijan un comportamiento: si un contrato menciona la frase, las restricciones deben nombrarlas.

<!-- - when ` + "`<frase>`" + `: ` + "`<prueba>`" + `, ` + "`<prueba>`" + ` -->

## External formats

Los formatos externos: si un contrato menciona la frase, las referencias vinculantes deben citar su fuente.

<!-- - ` + "`<frase>`" + `: ` + "`<ruta>`" + ` -->
`
	}
	return `# Lessons of this repository

This file holds this repository's lessons, and every loop session receives it.

## Pinning tests

Tests that pin a behaviour: when a contract mentions the phrase, Constraints must name them.

<!-- - when ` + "`<phrase>`" + `: ` + "`<name>`" + `, ` + "`<name>`" + ` -->

## External formats

External formats: when a contract mentions the phrase, Binding references must cite its source.

<!-- - ` + "`<phrase>`" + `: ` + "`<path>`" + ` -->
`
}

// initLanguage validates the language and stacks flags and returns the language
// to write in: the flag, else the existing config's, else English.
func initLanguage(g *Globals, out io.Writer, root string, o initOpts) (lang string, haveConfig bool, err error) {
	lang = "en"
	if o.languageSet {
		lang = o.language
		if err := config.Check(root, "language", lang); err != nil {
			return "", false, err
		}
	}
	if o.stacksSet {
		if err := config.Check(root, "metrics.stacks", o.stacks); err != nil {
			return "", false, err
		}
	}
	_, cfgErr := os.Lstat(filepath.Join(root, filepath.FromSlash(config.FilePath)))
	haveConfig = cfgErr == nil
	if haveConfig {
		v, err := config.Get(root, "language")
		if err != nil {
			return "", false, configErr(g, out, err)
		}
		if !o.languageSet {
			lang = v.Value
		}
	}
	return lang, haveConfig, nil
}

// configStep reports on, or builds the write of, .vloop/config.toml, printing
// the suggestions and starter checks that go with it.
func configStep(out io.Writer, root string, haveConfig bool, detected []string, chosen, lang string, starter []config.CheckDef) initStep {
	if !haveConfig {
		for _, c := range starter {
			fmt.Fprintln(out, "check "+detect.FormatCheck(c))
		}
		if len(starter) > 0 {
			fmt.Fprintln(out, "each check is a starting point — edit it in "+config.FilePath)
		}
		return initStep{path: config.FilePath, verb: "wrote", write: func() error {
			if err := config.Set(root, "language", lang); err != nil {
				return err
			}
			if chosen != "" {
				if err := config.Set(root, "metrics.stacks", chosen); err != nil {
					return err
				}
			}
			return config.AddChecks(root, starter)
		}}
	}
	if cur, err := config.Checks(root); err == nil && len(cur) == 0 {
		for _, c := range starter {
			fmt.Fprintln(out, "suggest check: "+detect.FormatCheck(c))
		}
	}
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
	return initStep{path: config.FilePath, verb: "kept"}
}

// starterBriefStep keeps the briefs folder when it holds a brief, else writes
// the getting-started brief in the language.
func starterBriefStep(root, lang string, now time.Time) (initStep, error) {
	hasBrief := false
	if ents, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(brief.Dir))); err == nil {
		for _, e := range ents {
			hasBrief = hasBrief || strings.HasSuffix(e.Name(), brief.Suffix)
		}
	}
	if hasBrief {
		return initStep{path: brief.Dir + "/", verb: "kept"}, nil
	}
	slug := "getting-started"
	if lang == "es" {
		slug = "primeros-pasos"
	}
	text, err := brief.Render(lang, slug, now)
	if err != nil {
		return initStep{}, err
	}
	return initStep{path: brief.NewPath(slug, now), verb: "wrote", content: []byte(text)}, nil
}

// mergedStep is the step for a file vloop merges its own part into: merge says
// what the merged content is and whether it differs from the file.
func mergedStep(path string, old []byte, exists bool, merge func() ([]byte, bool)) initStep {
	merged, changed := merge()
	switch {
	case !changed:
		return initStep{path: path, verb: "kept"}
	case exists:
		return initStep{path: path, verb: "updated", content: merged}
	default:
		return initStep{path: path, verb: "wrote", content: merged}
	}
}

// applySteps writes every step that is not kept, or with dryRun only says what
// it would do.
func applySteps(out io.Writer, root string, steps []initStep, dryRun bool) error {
	for _, s := range steps {
		if !dryRun && s.verb != "kept" {
			if s.write != nil {
				if err := s.write(); err != nil {
					return Problem(err)
				}
			} else if err := writeRepoFile(root, s.path, s.content); err != nil {
				return Problem(err)
			}
		}
		verb := s.verb
		if dryRun {
			verb = map[string]string{"wrote": "would write", "updated": "would update", "kept": "would keep"}[verb]
		}
		fmt.Fprintf(out, "%s %s\n", verb, s.path)
	}
	return nil
}

// printInitNext says what to do after a real init.
func printInitNext(out io.Writer) {
	fmt.Fprintln(out, "next:")
	fmt.Fprintln(out, "  vloop doctor")
	fmt.Fprintln(out, "  get /vloop:operate for your Claude Code sessions, once:")
	fmt.Fprintln(out, "    claude plugin marketplace add mvelosop/vloop")
	fmt.Fprintln(out, "    claude plugin install vloop@vloop")
	fmt.Fprintln(out, `  or, for one session: claude --plugin-dir "$(vloop plugin path)"`)
	fmt.Fprintln(out, `  then ask Claude Code for what you want ("run the brief docs/briefs/<name>"); /vloop:operate drives vloop`)
}
