package cli

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/mvelosop/vloop/internal/config"
)

const readmeLineCap = 200

// walk collects the path of every runnable leaf command (without the root
// name) and every flag name in the tree.
func walk(c *cobra.Command, path string, cmds, flags map[string]bool) {
	add := func(f *pflag.Flag) { flags[f.Name] = true }
	c.LocalFlags().VisitAll(add)
	c.PersistentFlags().VisitAll(add)
	for _, s := range c.Commands() {
		if s.Name() == "help" || s.Name() == "completion" {
			continue
		}
		p := strings.TrimSpace(path + " " + s.Name())
		if s.HasSubCommands() {
			walk(s, p, cmds, flags)
			continue
		}
		cmds[p] = true
		walk(s, p, cmds, flags)
	}
}

func TestReadmeMatchesCommandTree(t *testing.T) {
	b, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	if n := strings.Count(text, "\n"); n >= readmeLineCap {
		t.Errorf("README.md is %d lines; the cap is under %d", n, readmeLineCap)
	}

	root, _ := NewRoot(Build{})
	cmds, flags := map[string]bool{}, map[string]bool{}
	walk(root, "", cmds, flags)
	delete(flags, "help")

	for c := range cmds {
		if !strings.Contains(text, "vloop "+c) {
			t.Errorf("README does not document vloop %s", c)
		}
	}
	for f := range flags {
		if !strings.Contains(text, "--"+f) {
			t.Errorf("README does not document --%s", f)
		}
	}

	for _, m := range regexp.MustCompile(`--[a-z][a-z-]*`).FindAllString(text, -1) {
		if !flags[strings.TrimPrefix(m, "--")] {
			t.Errorf("README names %s, which no command has", m)
		}
	}

	// Every `vloop <words>` mention must resolve to a command in the tree.
	mention := regexp.MustCompile("(?:`|\\$ )vloop((?: [a-z][a-z-]*)+)")
	for _, m := range mention.FindAllStringSubmatch(text, -1) {
		args := strings.Fields(m[1])
		cmd, _, err := root.Find(args)
		if err != nil || cmd == root {
			t.Errorf("README names vloop%s, which is not a command", m[1])
		}
	}

	keys, envs := map[string]bool{}, map[string]bool{}
	for _, k := range config.Keys {
		keys[k.Name] = true
		envs["VLOOP_"+strings.ToUpper(strings.NewReplacer(".", "_", "-", "_").Replace(k.Name))] = true
		if !strings.Contains(text, "`"+k.Name+"`") {
			t.Errorf("README does not mention config key %s", k.Name)
		}
	}
	for _, m := range regexp.MustCompile(`\b(?:model|effort)\.[a-z]+`).FindAllString(text, -1) {
		if !keys[m] {
			t.Errorf("README names config key %s, which does not exist", m)
		}
	}
	var missing []string
	for _, m := range regexp.MustCompile(`VLOOP_[A-Z_]+`).FindAllString(text, -1) {
		if !envs[m] {
			missing = append(missing, m)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("README names variables vloop does not read: %v", missing)
	}
}
