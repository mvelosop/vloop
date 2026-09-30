package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CommandsReference renders docs/guide/commands.md from the command tree: every
// command and subcommand (help and completion aside) with its usage line, its
// short description and its flags. The page is generated; nothing edits it.
func CommandsReference() string {
	root, _ := NewRoot(Build{})
	var b strings.Builder
	b.WriteString("# Command reference\n\n")
	b.WriteString("This page is generated from vloop's command tree. Do not edit it: run\n")
	b.WriteString("`go generate ./...` to regenerate it. A test fails when it is out of date.\n\n")
	b.WriteString("## Global flags\n\nThese work on every command.\n\n")
	writeFlags(&b, root.PersistentFlags())
	writeCommand(&b, root, true)
	return b.String()
}

func writeCommand(b *strings.Builder, c *cobra.Command, isRoot bool) {
	if !isRoot {
		fmt.Fprintf(b, "## %s\n\n%s\n\n```\n%s\n```\n\n", c.CommandPath(), c.Short, c.UseLine())
		writeFlags(b, c.LocalFlags())
	}
	for _, s := range c.Commands() {
		if s.Name() == "help" || s.Name() == "completion" || s.Hidden {
			continue
		}
		writeCommand(b, s, false)
	}
}

// writeFlags lists a flag set's flags (help aside) as a bullet each, or
// nothing when there are none.
func writeFlags(b *strings.Builder, fs *pflag.FlagSet) {
	var flags []*pflag.Flag
	fs.VisitAll(func(f *pflag.Flag) {
		if f.Name != "help" {
			flags = append(flags, f)
		}
	})
	if len(flags) == 0 {
		return
	}
	sort.Slice(flags, func(i, j int) bool { return flags[i].Name < flags[j].Name })
	for _, f := range flags {
		name := "--" + f.Name
		if f.Shorthand != "" {
			name = "-" + f.Shorthand + ", " + name
		}
		varname, usage := pflag.UnquoteUsage(f)
		if varname != "" {
			name += " " + varname
		}
		fmt.Fprintf(b, "- `%s`: %s", name, usage)
		if f.DefValue != "" && f.DefValue != "false" && f.DefValue != "[]" {
			fmt.Fprintf(b, " (default %s)", f.DefValue)
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
}
