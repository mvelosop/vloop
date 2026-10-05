package cli

import (
	"strings"
	"testing"
)

// TestExitCodesOptInUsage pins brief Q2: exit 2 is for usage only, every other
// failure exits 1.
func TestExitCodesOptInUsage(t *testing.T) {
	d := scratchRepo(t)
	code, out, e := run(t, "-C", d, "intervention", "add", "two options", "--phase", "run", "--kind", "repair",
		"--automatable", "no", "--by", "operator", "--option", "a", "--option", "b", "--recommended", "1", "--why", "w", "--decided-option", "1")
	if code != 0 {
		t.Fatalf("intervention add: %d %q", code, e)
	}
	id := strings.TrimSuffix(out[strings.LastIndex(out, "/")+1:], ".md\n")

	cases := []struct {
		name string
		want int
		msg  string // the whole stderr line, when not empty
		args []string
	}{
		{"run missing brief", 1, "vloop: brief not found: docs/briefs/nope.md\n", []string{"run", "docs/briefs/nope.md"}},
		{"brief check missing file", 1, "vloop: no such file: missing.md\n", []string{"brief", "check", "missing.md"}},
		{"decided out of range", 2, "", []string{"intervention", "set", id, "decided", "5"}},
		{"recommended out of range", 2, "", []string{"intervention", "set", id, "recommended", "9"}},
		{"unknown flag", 2, "", []string{"status", "--bogus"}},
		{"unknown command", 2, "", []string{"nosuchcommand"}},
		{"unknown subcommand", 2, "", []string{"task", "nosuch"}},
		{"argument count", 2, "", []string{"task", "show"}},
		{"extra argument", 2, "", []string{"status", "extra"}},
		{"invalid flag value", 2, "", []string{"run", "--max-attempts", "0"}},
		{"invalid bool flag value", 2, "", []string{"run", "--plan-only=maybe"}},
		{"invalid config value", 2, "", []string{"config", "set", "language", "5x"}},
		{"unknown config key", 2, "", []string{"config", "get", "nope"}},
		{"invalid intervention field", 2, "", []string{"intervention", "set", id, "phase", "bogus"}},
		{"invalid --by", 2, "", []string{"intervention", "add", "x", "--phase", "run", "--kind", "repair", "--automatable", "no", "--by", "bogus"}},
		{"unknown schema", 2, "", []string{"schema", "show", "nope/v1"}},
		{"missing record", 1, "", []string{"intervention", "set", "I20990101-0000-nope", "phase", "run"}},
		{"missing intervention", 1, "", []string{"intervention", "show", "I20990101-0000-nope"}},
		{"missing brief file", 1, "", []string{"brief", "check", "docs/no-such.loop-brief.md"}},
		{"no plan", 1, "", []string{"task", "show", "T1"}},
	}
	for _, c := range cases {
		code, out, e := run(t, append([]string{"-C", d}, c.args...)...)
		if code != c.want {
			t.Errorf("%s: vloop %v exited %d, want %d (stdout %q stderr %q)", c.name, c.args, code, c.want, out, e)
			continue
		}
		if !strings.HasPrefix(e, "vloop: ") {
			t.Errorf("%s: stderr %q has no vloop: line", c.name, e)
		}
		if c.msg != "" && e != c.msg {
			t.Errorf("%s: stderr %q, want %q", c.name, e, c.msg)
		}
	}
}
