package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// b13Stub makes the stub claude print the given `plugin list --json` document.
func (r *b5Repo) b13Stub(list string) {
	r.t.Helper()
	script := "#!/bin/sh\ncase \"$1\" in\n  --version) echo '2.0.0 (Claude Code)';;\n  plugin) echo '" + list + "';;\nesac\n"
	if err := os.WriteFile(filepath.Join(r.stub, "claude"), []byte(script), 0o755); err != nil {
		r.t.Fatal(err)
	}
}

// b13Settings replaces the settings files the fixture reads: the project's
// .claude/settings.json and .claude/settings.local.json, and the temporary
// HOME's .claude/settings.json. An empty content leaves that file absent.
func (r *b5Repo) b13Settings(project, local, home string) {
	r.t.Helper()
	for _, f := range []struct{ path, content string }{
		{filepath.Join(r.dir, ".claude", "settings.json"), project},
		{filepath.Join(r.dir, ".claude", "settings.local.json"), local},
		{filepath.Join(r.home, ".claude", "settings.json"), home},
	} {
		if err := os.RemoveAll(f.path); err != nil {
			r.t.Fatal(err)
		}
		if f.content == "" {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(f.path), 0o755); err != nil {
			r.t.Fatal(err)
		}
		if err := os.WriteFile(f.path, []byte(f.content), 0o644); err != nil {
			r.t.Fatal(err)
		}
	}
}

// b13After returns the line after the "plugin" check's line (whatever its mark).
func b13After(t *testing.T, out string) string {
	t.Helper()
	lines := b5Lines(out)
	for i, l := range lines {
		if mark, rest, ok := strings.Cut(l, " "); ok && mark != "" && strings.HasPrefix(rest, "plugin") && i+1 < len(lines) {
			return lines[i+1]
		}
	}
	t.Fatalf("no plugin line in %q", out)
	return ""
}

// TestWorkedExampleB13 runs the brief's worked example: init's closing lines,
// the CLAUDE.md sentence, and doctor's plugin and gate edits checks.
func TestWorkedExampleB13(t *testing.T) {
	t.Parallel()
	const version = "0.9.0"
	const pluginWarn = `! plugin /vloop:operate is not available in your interactive sessions — claude plugin install vloop@vloop (after claude plugin marketplace add mvelosop/vloop), or claude --plugin-dir "$(vloop plugin path)" for one session`
	const gateWarn = `! gate edits no allow rule lets an interactive session edit .vloop/state/gates/ — add "Edit(./.vloop/state/gates/**)" to permissions.allow in .claude/settings.json`
	r := b5New(t, false, false)

	init := r.run(version, "init")
	b5Code(t, init, 0)
	r.write(".vloop/config.toml", r.read(".vloop/config.toml")+"\n[[check]]\nname = \"all\"\npaths = [\"**\"]\nrun = \"true\"\n")
	lines := b5Lines(init.out)
	wantNext := []string{"next:", "vloop doctor",
		"get /vloop:operate for your Claude Code sessions, once:",
		"claude plugin marketplace add mvelosop/vloop",
		"claude plugin install vloop@vloop",
		`or, for one session: claude --plugin-dir "$(vloop plugin path)"`,
		`then ask Claude Code for what you want ("run the brief docs/briefs/<name>"); /vloop:operate drives vloop`}
	if len(lines) < len(wantNext) || strings.Join(lines[len(lines)-len(wantNext):], "\n") != strings.Join(wantNext, "\n") {
		t.Fatalf("init does not end with D3's seven lines: %q", lines)
	}
	if !strings.Contains(r.read("CLAUDE.md"), "If /vloop:operate is not available, the vloop plugin is not loaded") {
		t.Fatalf("CLAUDE.md lacks D3's sentence:\n%s", r.read("CLAUDE.md"))
	}

	// No plugin listed, no settings files: both warnings, exit 0.
	r.b13Stub("[]")
	r.b13Settings("", "", "")
	d := r.run(version, "doctor")
	b5Code(t, d, 0)
	b5Has(t, d, pluginWarn)
	if got := b13After(t, d.out); got != gateWarn {
		t.Fatalf("after the plugin line: %q", got)
	}
	b5Has(t, d, "doctor: 0 problem(s), 2 warning(s)")

	// The plugin at the binary's version, enabled: the plugin check passes.
	r.b13Stub(`[{"id":"vloop@vloop","version":"` + version + `","enabled":true}]`)
	d = r.run(version, "doctor")
	b5Code(t, d, 0)
	b5Has(t, d, "✓ plugin")

	for _, c := range []struct {
		name                 string
		project, local, home string
		want                 string
	}{
		{"project rule", `{"permissions":{"allow":["Edit(./.vloop/state/gates/**)"]}}`, "", "", "✓ gate edits"},
		{"local rule", "", `{"permissions":{"allow":["Edit(.vloop/**)"]}}`, "", "✓ gate edits"},
		{"home bare Edit", "", "", `{"permissions":{"allow":["Edit"]}}`, "✓ gate edits"},
		{"unrelated rules", `{"permissions":{"allow":["Edit(./src/**)","Bash(go:*)"]}}`, "", "", gateWarn},
	} {
		r.b13Settings(c.project, c.local, c.home)
		d = r.run(version, "doctor")
		b5Code(t, d, 0)
		if got := b13After(t, d.out); got != c.want {
			t.Fatalf("%s: after the plugin line %q, want %q", c.name, got, c.want)
		}
	}

	// An unreadable settings file warns and names the file.
	r.b13Settings(`{"permissions":`+"\n", "", "")
	d = r.run(version, "doctor")
	b5Code(t, d, 0)
	if got := b13After(t, d.out); !strings.HasPrefix(got, "! gate edits cannot read .claude/settings.json: ") || len(got) <= len("! gate edits cannot read .claude/settings.json: ") {
		t.Fatalf("unreadable file: %q", got)
	}

	// No claude on PATH: both checks are not applicable.
	r.b13Settings("", "", "")
	sys := "/usr/bin" + string(os.PathListSeparator) + "/bin"
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("no git")
	}
	cmd := exec.Command(b5Binary(t, version), "doctor")
	cmd.Dir = r.dir
	cmd.Env = append(cleanEnv(r.home), "PATH="+sys, "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")
	raw, _ := cmd.CombinedOutput() // exit 1: a missing claude is a problem; only the two n/a lines matter here
	d = result{out: string(raw)}
	b5Has(t, d, "- plugin")
	if got := b13After(t, d.out); got != "- gate edits" {
		t.Fatalf("no claude: after plugin %q", got)
	}

	// --json carries a check named "gate edits" with a result.
	r.b13Settings("", "", "")
	r.b13Stub("[]")
	j := r.run(version, "doctor", "--json")
	var doc struct {
		Checks []struct{ Name, Result, Message string } `json:"checks"`
	}
	if err := json.Unmarshal([]byte(j.out), &doc); err != nil {
		t.Fatalf("doctor --json: %v\n%s", err, j.out)
	}
	n := 0
	for _, c := range doc.Checks {
		if c.Name == "gate edits" && c.Result != "" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("--json has %d 'gate edits' checks: %s", n, j.out)
	}
}
