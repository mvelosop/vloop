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

	var hphase, hroot string
	hook := &cobra.Command{
		Use:   "hook --phase <review|second-review> --root <dir>",
		Short: "Answer one Copilot preToolUse payload from stdin with an allow or deny line",
		Long: `The command Copilot runs before every tool call of a review session. It reads
one payload ({"toolName", "toolArgs", "cwd"}) on stdin and prints one line of
JSON: {"permissionDecision":"allow"}, or {"permissionDecision":"deny",
"permissionDecisionReason":"vloop fence: ..."}, and exits 0. toolArgs may be an
object or a JSON string holding one. A payload it cannot read, an unknown phase
or any internal error prints the deny line and exits 2: it never allows on an
error. It reads nothing but the payload and writes nothing.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			reply := func(allow bool, reason string) {
				m := map[string]string{"permissionDecision": "deny", "permissionDecisionReason": "vloop fence: " + reason}
				if allow {
					m = map[string]string{"permissionDecision": "allow"}
				}
				b, _ := json.Marshal(m)
				fmt.Fprintln(out, string(b))
			}
			failed := func(reason string) error {
				reply(false, reason)
				return &ExitError{Code: ExitUsage, Err: errors.New("")}
			}
			home, _ := os.UserHomeDir()
			abs, err := filepath.Abs(hroot)
			if hroot == "" || err != nil {
				return failed("--root is required")
			}
			j, err := fence.New(hphase, abs, home)
			if err != nil {
				return failed(err.Error())
			}
			raw, err := io.ReadAll(io.LimitReader(cmd.InOrStdin(), 8<<20))
			if err != nil {
				return failed("cannot read the payload: " + err.Error())
			}
			tool, path, command, err := parseHookPayload(raw)
			if err != nil {
				return failed(err.Error())
			}
			if t := strings.ToLower(tool); (t == "bash" || t == "powershell") && strings.TrimSpace(command) == "" {
				reply(false, "the "+tool+" call has no command the hook can read")
				return nil
			}
			d := j.Decide(tool, path, command)
			reply(d.Allow, d.Reason)
			return nil
		},
	}
	hook.Flags().StringVar(&hphase, "phase", "", "the session phase: review or second-review")
	hook.Flags().StringVar(&hroot, "root", "", "the repository root relative paths resolve against")
	cmd.AddCommand(hook)
	return cmd
}

// hookArgs are the toolArgs keys the hook reads.
type hookArgs struct {
	Command string `json:"command"`
	Path    string `json:"path"`
	Skill   string `json:"skill"`
	Name    string `json:"name"`
	Input   string `json:"input"`
	Patch   string `json:"patch"`
}

// parseHookPayload reads a Copilot preToolUse payload into the judge's
// arguments. toolArgs is an object or a JSON string holding one. A call whose
// arguments the hook cannot read comes back with an empty path or command, so
// the judge denies it (a file tool without a path, an empty patch); a shell
// tool with none is denied by the hook itself.
func parseHookPayload(raw []byte) (tool, path, command string, err error) {
	var p struct {
		ToolName string          `json:"toolName"`
		ToolArgs json.RawMessage `json:"toolArgs"`
	}
	if err = json.Unmarshal(raw, &p); err != nil {
		return "", "", "", errors.New("the payload is not JSON")
	}
	if p.ToolName == "" {
		return "", "", "", errors.New("the payload has no toolName")
	}
	tool = p.ToolName
	args := p.ToolArgs
	var str string
	if json.Unmarshal(args, &str) == nil { // a JSON string: an embedded object, or a raw patch
		inner := json.RawMessage(str)
		var m map[string]any
		if json.Unmarshal(inner, &m) == nil {
			args = inner
		} else if strings.EqualFold(tool, "apply_patch") {
			return tool, "", str, nil
		} else {
			args = nil
		}
	}
	var a hookArgs
	if len(args) > 0 && json.Unmarshal(args, &a) != nil {
		a = hookArgs{}
	}
	path = a.Path
	if strings.EqualFold(tool, "skill") {
		path = a.Skill
		if path == "" {
			path = a.Name
		}
	}
	command = a.Command
	if strings.EqualFold(tool, "apply_patch") {
		command = a.Input
		if command == "" {
			command = a.Patch
		}
		if command == "" {
			command = a.Command
		}
	}
	return tool, path, command, nil
}
