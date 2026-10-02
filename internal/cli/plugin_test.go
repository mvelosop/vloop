package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func runPluginCLI(t *testing.T, b Build, dir string, args ...string) (string, string, int) {
	t.Helper()
	var out, errb bytes.Buffer
	code := Execute(b, append([]string{"-C", dir}, args...), &out, &errb)
	return out.String(), errb.String(), code
}

func TestCheckPluginMessages(t *testing.T) {
	d := t.TempDir()
	write := func(sub, body string) {
		p := filepath.Join(d, sub, ".claude-plugin")
		os.MkdirAll(p, 0o755)
		os.WriteFile(filepath.Join(p, "plugin.json"), []byte(body), 0o644)
	}
	write("eq", `{"version":"1.0.0"}`)
	write("old", `{"version":"0.9.0"}`)
	write("bad", `{oops`)
	write("nov", `{}`)
	b := Build{Version: "1.0.0"}
	for _, c := range []struct{ dir, want string }{
		{"eq", ""},
		{"old", "vloop: the vloop plugin is 0.9.0 but the vloop binary is 1.0.0 — update the one that is behind\n"},
		{"bad", "vloop: cannot read the vloop plugin's version in bad\n"},
		{"nov", "vloop: cannot read the vloop plugin's version in nov\n"},
		{"missing", "vloop: cannot read the vloop plugin's version in missing\n"},
	} {
		var out, errb bytes.Buffer
		code := Execute(b, []string{"version", "--check-plugin", filepath.Join(d, c.dir)}, &out, &errb)
		want := strings.ReplaceAll(c.want, " in "+c.dir+"\n", " in "+filepath.Join(d, c.dir)+"\n")
		if code != 0 || out.String() != want || errb.Len() != 0 {
			t.Errorf("%s: code %d out %q err %q, want %q", c.dir, code, out.String(), errb.String(), want)
		}
	}
}

func TestPluginPathIdempotentAndClean(t *testing.T) {
	d := t.TempDir()
	os.Mkdir(filepath.Join(d, ".git"), 0o755)
	b := Build{Version: "9.9.9"}
	out, _, code := runPluginCLI(t, b, d, "plugin", "path")
	if code != 0 || out != ".vloop/tmp/plugin/9.9.9\n" {
		t.Fatalf("code %d out %q", code, out)
	}
	x := filepath.Join(d, ".vloop/tmp/plugin/9.9.9")
	hooks := filepath.Join(x, "hooks/hooks.json")
	old := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	os.Chtimes(hooks, old, old)
	os.WriteFile(filepath.Join(x, "extra.txt"), []byte("x"), 0o644)
	os.MkdirAll(filepath.Join(x, "stray"), 0o755)
	runPluginCLI(t, b, d, "plugin", "path")
	if fi, _ := os.Stat(hooks); !fi.ModTime().Equal(old) {
		t.Error("a matching file was rewritten")
	}
	for _, p := range []string{"extra.txt", "stray"} {
		if _, err := os.Stat(filepath.Join(x, p)); err == nil {
			t.Errorf("%s was not removed", p)
		}
	}
	os.WriteFile(hooks, []byte("broken"), 0o644)
	runPluginCLI(t, b, d, "plugin", "path")
	if got, _ := os.ReadFile(hooks); string(got) == "broken" {
		t.Error("a modified file was not restored")
	}
	if out, _, _ := runPluginCLI(t, b, d, "version", "--check-plugin", x); out != "" {
		t.Errorf("check-plugin on the extracted copy printed %q", out)
	}
}

func TestPluginPathRefusesSymlink(t *testing.T) {
	d, outside := t.TempDir(), t.TempDir()
	os.MkdirAll(filepath.Join(d, ".vloop"), 0o755)
	if err := os.Symlink(outside, filepath.Join(d, ".vloop/tmp")); err != nil {
		t.Skip("symlinks unavailable")
	}
	_, _, code := runPluginCLI(t, Build{Version: "1.0.0"}, d, "plugin", "path")
	if code == 0 {
		t.Error("want a non-zero exit")
	}
	if ents, _ := os.ReadDir(outside); len(ents) != 0 {
		t.Errorf("wrote through the symlink: %v", ents)
	}
}

func TestPluginPathLeavesOutEvals(t *testing.T) {
	d := t.TempDir()
	os.Mkdir(filepath.Join(d, ".git"), 0o755)
	b := Build{Version: "9.9.9"}
	x := filepath.Join(d, ".vloop/tmp/plugin/9.9.9")
	if _, _, code := runPluginCLI(t, b, d, "plugin", "path"); code != 0 {
		t.Fatalf("code %d", code)
	}
	if _, err := os.Stat(filepath.Join(x, "skills/review/SKILL.md")); err != nil {
		t.Errorf("skills were not extracted: %v", err)
	}
	if _, err := os.Stat(filepath.Join(x, "evals")); err == nil {
		t.Error("evals/ was extracted")
	}
	os.MkdirAll(filepath.Join(x, "evals/stale"), 0o755)
	os.WriteFile(filepath.Join(x, "evals/stale/prompt.md"), []byte("x"), 0o644)
	if _, _, code := runPluginCLI(t, b, d, "plugin", "path"); code != 0 {
		t.Fatalf("second run: code %d", code)
	}
	if _, err := os.Stat(filepath.Join(x, "evals")); err == nil {
		t.Error("an evals/ left in the extracted tree was not removed")
	}
}
