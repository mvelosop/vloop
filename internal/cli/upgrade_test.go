package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func upgradeFixture(t *testing.T, from string) string {
	t.Helper()
	d := initRepo(t, map[string]string{"go.mod": "module a\n", ".gitignore": "x/\n"})
	if _, _, code := runPluginCLI(t, Build{Version: from, Commit: "c" + from}, d, "init"); code != 0 {
		t.Fatalf("init exit %d", code)
	}
	return d
}

func TestUpgradeNotInitialized(t *testing.T) {
	d := initRepo(t, map[string]string{"go.mod": "module a\n"})
	out, errs, code := runPluginCLI(t, Build{Version: "0.1.1"}, d, "upgrade")
	if code != 1 || out != "" || errs != "vloop: not initialized — run vloop init\n" {
		t.Errorf("code %d out %q err %q", code, out, errs)
	}
}

func TestUpgradeSameWritesNothing(t *testing.T) {
	d := upgradeFixture(t, "0.1.0")
	before := readFile(t, d, ".vloop/install.json")
	out, _, code := runPluginCLI(t, Build{Version: "0.1.0", Commit: "x"}, d, "upgrade")
	if code != 0 || out != "already at 0.1.0\n" || readFile(t, d, ".vloop/install.json") != before {
		t.Errorf("code %d out %q", code, out)
	}
}

func TestUpgradeRefusesDowngradeNumerically(t *testing.T) {
	d := upgradeFixture(t, "0.1.10")
	before := readFile(t, d, "CLAUDE.md")
	for _, args := range [][]string{{"upgrade"}, {"upgrade", "--yes"}} {
		out, errs, code := runPluginCLI(t, Build{Version: "0.1.9"}, d, args...)
		want := "vloop: this repository was set up by vloop 0.1.10, newer than this binary (0.1.9) — upgrade the binary, not the repository\n"
		if code != 1 || out != "" || errs != want {
			t.Errorf("%v: code %d out %q err %q", args, code, out, errs)
		}
	}
	if readFile(t, d, "CLAUDE.md") != before {
		t.Error("CLAUDE.md changed")
	}
}

func TestUpgradeApplies(t *testing.T) {
	d := upgradeFixture(t, "0.1.0")
	out, errs, code := runPluginCLI(t, Build{Version: "0.1.1", Commit: "c011"}, d, "upgrade")
	if code != 0 || errs != "" || out != "updated CLAUDE.md\nupdated .vloop/install.json\n" {
		t.Fatalf("code %d out %q err %q", code, out, errs)
	}
	st := readFile(t, d, ".vloop/install.json")
	if !strings.Contains(st, `"version":"0.1.1"`) || !strings.Contains(st, `"commit":"c011"`) || strings.Contains(st, `"upgraded":null`) {
		t.Errorf("stamp = %s", st)
	}
	if cm := readFile(t, d, "CLAUDE.md"); !strings.Contains(cm, "vloop 0.1.1") || strings.Contains(cm, "vloop 0.1.0") {
		t.Errorf("CLAUDE.md = %q", cm)
	}
}

func TestUpgradeBreakingNeedsYes(t *testing.T) {
	d := upgradeFixture(t, "0.1.1")
	b := Build{Version: "0.2.0", Commit: "c020"}
	before := readFile(t, d, "CLAUDE.md")
	out, errs, code := runPluginCLI(t, b, d, "upgrade")
	want := "vloop: 0.1.1 → 0.2.0 may break this repository's setup — review the changes above, then run vloop upgrade --yes\n"
	if code != 1 || out != "would update CLAUDE.md\n" || errs != want {
		t.Fatalf("code %d out %q err %q", code, out, errs)
	}
	if readFile(t, d, "CLAUDE.md") != before || strings.Contains(readFile(t, d, ".vloop/install.json"), "0.2.0") {
		t.Error("a refused upgrade wrote")
	}
	if _, _, code := runPluginCLI(t, b, d, "upgrade", "--yes"); code != 0 {
		t.Errorf("--yes exit %d", code)
	}
	if !strings.Contains(readFile(t, d, ".vloop/install.json"), `"version":"0.2.0"`) {
		t.Error("stamp not rewritten")
	}
}

func TestUpgradePreReleaseNeedsYes(t *testing.T) {
	d := upgradeFixture(t, "0.0.0-dev")
	_, errs, code := runPluginCLI(t, Build{Version: "0.1.0"}, d, "upgrade")
	if code != 1 || !strings.Contains(errs, "may break this repository's setup") {
		t.Errorf("code %d err %q", code, errs)
	}
	if _, _, code := runPluginCLI(t, Build{Version: "0.1.0"}, d, "upgrade", "--yes"); code != 0 {
		t.Errorf("--yes exit %d", code)
	}
}

func TestUpgradeMarkerOnlyAndGitignore(t *testing.T) {
	d := upgradeFixture(t, "0.1.0")
	cm := "TOP\n\n" + readFile(t, d, "CLAUDE.md") + "\nbelow  \nno newline"
	os.WriteFile(filepath.Join(d, "CLAUDE.md"), []byte(cm), 0o644)
	os.WriteFile(filepath.Join(d, ".gitignore"), []byte("x/"), 0o644)
	cfg := readFile(t, d, ".vloop/config.toml")
	out, _, code := runPluginCLI(t, Build{Version: "0.1.1"}, d, "upgrade")
	if code != 0 || !strings.Contains(out, "updated .gitignore\n") {
		t.Fatalf("code %d out %q", code, out)
	}
	got := readFile(t, d, "CLAUDE.md")
	if !strings.HasPrefix(got, "TOP\n\n<!-- vloop:begin -->\n") || !strings.HasSuffix(got, "<!-- vloop:end -->\n\nbelow  \nno newline") || !strings.Contains(got, "vloop 0.1.1") {
		t.Errorf("CLAUDE.md = %q", got)
	}
	if readFile(t, d, ".gitignore") != "x/\n.vloop/tmp/\n" || readFile(t, d, ".vloop/config.toml") != cfg {
		t.Error("gitignore or config wrong")
	}
}

func TestUpgradeDryRun(t *testing.T) {
	d := upgradeFixture(t, "0.1.0")
	before := readFile(t, d, ".vloop/install.json")
	out, _, code := runPluginCLI(t, Build{Version: "0.1.1"}, d, "upgrade", "--dry-run")
	if code != 0 || out != "would update CLAUDE.md\n" || readFile(t, d, ".vloop/install.json") != before {
		t.Errorf("code %d out %q", code, out)
	}
	if strings.Contains(readFile(t, d, "CLAUDE.md"), "0.1.1") {
		t.Error("dry-run wrote")
	}
}

func TestUpgradeRefusesSymlink(t *testing.T) {
	d := upgradeFixture(t, "0.1.0")
	out := t.TempDir()
	os.WriteFile(filepath.Join(out, "CLAUDE.md"), []byte("outside\n"), 0o644)
	os.Remove(filepath.Join(d, "CLAUDE.md"))
	if err := os.Symlink(filepath.Join(out, "CLAUDE.md"), filepath.Join(d, "CLAUDE.md")); err != nil {
		t.Skip(err)
	}
	if _, _, code := runPluginCLI(t, Build{Version: "0.1.1"}, d, "upgrade"); code != 1 {
		t.Errorf("code %d", code)
	}
	if b, _ := os.ReadFile(filepath.Join(out, "CLAUDE.md")); string(b) != "outside\n" {
		t.Error("wrote through a symlink")
	}
}
