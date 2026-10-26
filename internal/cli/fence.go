package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/fence"
)

func newFence(g *Globals) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fence",
		Short: "Judge a tool call against a session phase's fence",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}

	var phase, root, tool, path string
	check := &cobra.Command{
		Use:   "check --phase <review|second-review> --root <dir> --tool <name> [--path <p>] [\"<command>\"]",
		Short: "Print allow or deny: <reason> for one tool call; exit 1 on deny",
		Long: `Judge one tool call the way the Copilot hook does, without running anything.
A shell tool's command string is split at ; && || | & and newlines outside
quotes, looked through env, command, exec, nohup, time, nice, timeout, stdbuf,
xargs and sh -c, and matched against the review fence's Bash deny rules and the
Copilot adapter's deny list. A file tool's path, and a skill's name, are given
with --path. Exits 0 and prints allow, or exits 1 and prints deny: <reason>;
bad arguments exit 2.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if tool == "" {
				return Usage(errors.New("--tool is required"))
			}
			if root == "" {
				return Usage(errors.New("--root is required"))
			}
			abs, err := filepath.Abs(root)
			if err != nil {
				return Usage(err)
			}
			home, _ := os.UserHomeDir()
			j, err := fence.New(phase, abs, home)
			if err != nil {
				return Usage(err)
			}
			var c string
			if len(args) == 1 {
				c = args[0]
			}
			d := j.Decide(tool, path, c)
			out := cmd.OutOrStdout()
			if d.Allow {
				fmt.Fprintln(out, "allow")
				return nil
			}
			fmt.Fprintln(out, "deny: "+d.Reason)
			return &ExitError{Code: ExitProblems, Err: errors.New("")}
		},
	}
	check.Flags().StringVar(&phase, "phase", "", "the session phase: review or second-review")
	check.Flags().StringVar(&root, "root", "", "the repository root relative paths resolve against")
	check.Flags().StringVar(&tool, "tool", "", "the tool's name, as Copilot spells it")
	check.Flags().StringVar(&path, "path", "", "the file path a file tool names, or the skill's name")
	cmd.AddCommand(check)
	return cmd
}
