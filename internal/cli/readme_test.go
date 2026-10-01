package cli

import (
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/driver"
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

// TestReadmeShellDefaultMatchesCode checks the README's stated default for the
// `shell` key on each OS against config's (F1).
func TestReadmeShellDefaultMatchesCode(t *testing.T) {
	b, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatal(err)
	}
	var row string
	for _, l := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(l, "| `shell` |") {
			row = l
		}
	}
	if row == "" {
		t.Fatal("README has no shell row in its config table")
	}
	def := strings.TrimSpace(strings.Split(row, "|")[2])
	m := regexp.MustCompile("^`([a-z]+)` \\(`([a-z]+)` on Windows\\)$").FindStringSubmatch(def)
	if m == nil {
		t.Fatalf("README's shell default %q is not of the form `x` (`y` on Windows)", def)
	}
	// The code's default on the other OS is read from defaultShell's source,
	// since config.Keys holds only the running OS's.
	src, err := os.ReadFile("../../internal/config/config.go")
	if err != nil {
		t.Fatal(err)
	}
	fn := regexp.MustCompile(`(?s)func defaultShell\(\) string \{.*?return "([a-z]+)"\n\t\}\n\treturn "([a-z]+)"`).FindStringSubmatch(string(src))
	if fn == nil {
		t.Fatal("cannot read defaultShell in internal/config/config.go")
	}
	if m[2] != fn[1] || m[1] != fn[2] {
		t.Errorf("README says shell defaults to %s (%s on Windows); the code says %s (%s on Windows)", m[1], m[2], fn[2], fn[1])
	}
	for _, k := range config.Keys {
		if k.Name == "shell" {
			want := m[1]
			if runtime.GOOS == "windows" {
				want = m[2]
			}
			if k.Default != want {
				t.Errorf("README says the shell default here is %s; config says %s", want, k.Default)
			}
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
