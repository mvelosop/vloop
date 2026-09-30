package cli

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/install"
)

func newUpgrade(b Build, g *Globals) *cobra.Command {
	var dryRun, yes bool
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Refresh vloop's parts of a repository set up by an older vloop",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runUpgrade(b, g, cmd, dryRun, yes, time.Now())
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print what would be updated and write nothing")
	cmd.Flags().BoolVar(&yes, "yes", false, "allow a breaking upgrade")
	return cmd
}

func runUpgrade(b Build, g *Globals, cmd *cobra.Command, dryRun, yes bool, now time.Time) error {
	out := cmd.OutOrStdout()
	root, err := g.root()
	if err != nil {
		return Problem(err)
	}
	st, err := install.Read(root)
	if errors.Is(err, os.ErrNotExist) {
		return Problem(errors.New("not initialized — run vloop init"))
	}
	if err != nil {
		return Problem(err)
	}
	rel, err := install.Compare(st.Version, b.Version)
	if err != nil {
		return Problem(err)
	}
	switch rel {
	case install.Same:
		fmt.Fprintf(out, "already at %s\n", b.Version)
		return nil
	case install.FromNewer:
		return Problem(fmt.Errorf("this repository was set up by vloop %s, newer than this binary (%s) — upgrade the binary, not the repository", st.Version, b.Version))
	}

	type change struct {
		path    string
		content []byte
	}
	var changes []change
	old, exists, err := readRepoFile(root, "CLAUDE.md")
	if err != nil {
		return Problem(err)
	}
	if merged := install.MergeClaudeMD(old, exists, b.Version); !exists || string(merged) != string(old) {
		changes = append(changes, change{"CLAUDE.md", merged})
	}
	old, _, err = readRepoFile(root, ".gitignore")
	if err != nil {
		return Problem(err)
	}
	if merged, changed := install.MergeGitignore(old); changed {
		changes = append(changes, change{".gitignore", merged})
	}
	if err := checkRepoPath(root, install.Path); err != nil {
		return Problem(err)
	}

	breaking := rel != install.Newer && !yes
	if dryRun || breaking {
		for _, c := range changes {
			fmt.Fprintf(out, "would update %s\n", c.path)
		}
		if breaking && !dryRun {
			return Problem(fmt.Errorf("%s → %s may break this repository's setup — review the changes above, then run vloop upgrade --yes", st.Version, b.Version))
		}
		return nil
	}

	for _, c := range changes {
		if err := writeRepoFile(root, c.path, c.content); err != nil {
			return Problem(err)
		}
		fmt.Fprintf(out, "updated %s\n", c.path)
	}
	up := now.UTC().Format("2006-01-02T15:04:05Z")
	st.Version, st.Commit, st.Upgraded = b.Version, b.Commit, &up
	if err := install.Write(root, st); err != nil {
		return Problem(err)
	}
	fmt.Fprintf(out, "updated %s\n", install.Path)
	return nil
}
