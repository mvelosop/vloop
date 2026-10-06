package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/driver"
)

const readmeLineCap = 200

// claudeFlags are flags of the claude CLI that the README may name.
var claudeFlags = map[string]bool{"--plugin-dir": true}

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

// readmeProblems returns what is wrong with a README: a flag, command, config
// key or variable it names that does not exist, or a guide it does not link.
// Completeness (every name documented) is the guides' check, not the README's.
func readmeProblems(text string, guides []string) []string {
	var out []string
	if n := strings.Count(text, "\n"); n >= readmeLineCap {
		out = append(out, fmt.Sprintf("README.md is %d lines; the cap is under %d", n, readmeLineCap))
	}
	for _, g := range guides {
		if !strings.Contains(text, "("+g+")") {
			out = append(out, "README does not link "+g)
		}
	}

	root, _ := NewRoot(Build{})
	cmds, flags := map[string]bool{}, map[string]bool{}
	walk(root, "", cmds, flags)
	delete(flags, "help")

	for _, m := range regexp.MustCompile(`--[a-z][a-z-]*`).FindAllString(text, -1) {
		if !flags[strings.TrimPrefix(m, "--")] && !claudeFlags[m] {
			out = append(out, "README names "+m+", which no command has")
		}
	}

	// Every `vloop <words>` mention must resolve to a command in the tree.
	mention := regexp.MustCompile("(?:`|\\$ )vloop((?: [a-z][a-z-]*)+)")
	for _, m := range mention.FindAllStringSubmatch(text, -1) {
		args := strings.Fields(m[1])
		cmd, _, err := root.Find(args)
		if err != nil || cmd == root {
			out = append(out, "README names vloop"+m[1]+", which is not a command")
		}
	}

	keys, envs := map[string]bool{}, map[string]bool{}
	for _, k := range config.Keys {
		keys[k.Name] = true
		envs["VLOOP_"+strings.ToUpper(strings.NewReplacer(".", "_", "-", "_").Replace(k.Name))] = true
	}
	for _, m := range regexp.MustCompile(`\b(?:model|effort)\.[a-z]+`).FindAllString(text, -1) {
		if !keys[m] {
			out = append(out, "README names config key "+m+", which does not exist")
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
		out = append(out, fmt.Sprintf("README names variables vloop does not read: %v", missing))
	}
	return out
}

func guideLinks(t *testing.T) []string {
	t.Helper()
	files, err := filepath.Glob("../../docs/guide/*.md")
	if err != nil || len(files) == 0 {
		t.Fatalf("no guides found: %v", err)
	}
	var out []string
	for _, f := range files {
		out = append(out, "docs/guide/"+filepath.Base(f))
	}
	return out
}

func TestReadmeMatchesCommandTree(t *testing.T) {
	b, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range readmeProblems(string(b), guideLinks(t)) {
		t.Error(p)
	}
}

// TestReadmeRejectsUnknownNames shows the README check fails on a name that
// does not exist.
func TestReadmeRejectsUnknownNames(t *testing.T) {
	text := "Run `vloop status --frobnicate`.\n"
	var found bool
	for _, p := range readmeProblems(text, nil) {
		if strings.Contains(p, "--frobnicate") {
			found = true
		}
	}
	if !found {
		t.Error("a README naming --frobnicate passed the check")
	}
}

// TestReadmeNamesTheSkills checks the README has a Skills section naming the
// four plugin skills.
func TestReadmeNamesTheSkills(t *testing.T) {
	b, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	if !regexp.MustCompile(`(?m)^#+ .*Skills`).MatchString(text) {
		t.Error("README has no heading with \"Skills\"")
	}
	for _, k := range []string{"plan", "work", "review", "gate-review", "operate"} {
		if !strings.Contains(text, "/vloop:"+k) {
			t.Errorf("README does not name /vloop:%s", k)
		}
	}
}

// TestGuideExitCodesMatchDriver checks that the guide's exit-code section lists
// exactly the codes the driver can return.
func TestGuideExitCodesMatchDriver(t *testing.T) {
	b, err := os.ReadFile("../../docs/guide/concepts.md")
	if err != nil {
		t.Fatal(err)
	}
	var sec []string
	in := false
	for _, l := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(l, "## ") {
			in = l == "## Exit codes"
			continue
		}
		if in {
			sec = append(sec, l)
		}
	}
	listed := map[int]bool{}
	row := regexp.MustCompile("^\\| *`?([0-9]+)`? *\\|")
	for _, l := range sec {
		if m := row.FindStringSubmatch(l); m != nil {
			n := 0
			for _, c := range m[1] {
				n = n*10 + int(c-'0')
			}
			listed[n] = true
		}
	}
	can := map[int]bool{}
	for _, n := range []int{0, ExitProblems, ExitUsage,
		driver.ExitBlocked, driver.ExitStalled, driver.ExitMaxIter,
		driver.ExitNotConverging, driver.ExitCostCeiling, driver.ExitSessionError,
		driver.ExitRepeatBlocked, driver.ExitRefsMoved} {
		can[n] = true
	}
	for n := range can {
		if !listed[n] {
			t.Errorf("the guide's exit codes do not list %d, which the driver can return", n)
		}
	}
	for n := range listed {
		if !can[n] {
			t.Errorf("the guide's exit codes list %d, which the driver cannot return", n)
		}
	}
}
