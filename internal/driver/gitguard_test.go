package driver

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func guardRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", root).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	return root
}

func appendConfig(t *testing.T, root, text string) {
	t.Helper()
	f, err := os.OpenFile(filepath.Join(root, ".git", "config"), os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(text); err != nil {
		t.Fatal(err)
	}
}

func TestGitGuardCompareEntries(t *testing.T) {
	t.Parallel()
	root := guardRepo(t)
	g := snapshotGit(root)
	if got := g.changed(); got != "" {
		t.Fatalf("untouched: %q", got)
	}

	appendConfig(t, root, "[branch \"b\"]\n\tvscode-merge-base = refs/heads/main\n")
	if got := g.changed(); got != "" {
		t.Errorf("the built-in pattern did not ignore the editor key: %q", got)
	}

	appendConfig(t, root, "[core]\n\tfsmonitor = false\n")
	if got, want := g.changed(), ".git/config (core.fsmonitor)"; got != want {
		t.Errorf("changed = %q, want %q", got, want)
	}
}

func TestGitGuardNamesThreeKeysThenCount(t *testing.T) {
	t.Parallel()
	root := guardRepo(t)
	g := snapshotGit(root)
	appendConfig(t, root, "[alias]\n\te = x\n\tc = x\n\ta = x\n\td = x\n\tb = x\n")
	if got, want := g.changed(), ".git/config (alias.a, alias.b, alias.c and 2 more)"; got != want {
		t.Errorf("changed = %q, want %q", got, want)
	}
}

func TestGitGuardIgnoresOnlyWhatMatches(t *testing.T) {
	t.Parallel()
	root := guardRepo(t)
	g := snapshotGitWith(root, []string{"gitlens.*"})
	appendConfig(t, root, "[gitlens \"x\"]\n\ty = 1\n")
	if got := g.changed(); got != "" {
		t.Errorf("ignored key halted: %q", got)
	}
	appendConfig(t, root, "[include]\n\tpath = extra.cfg\n")
	if got := g.changed(); !strings.Contains(got, "include.path") {
		t.Errorf("include directive not named: %q", got)
	}
}

func TestGitGuardValueChangeAndRemoval(t *testing.T) {
	t.Parallel()
	root := guardRepo(t)
	appendConfig(t, root, "[user]\n\tname = a\n\temail = a@x\n")
	g := snapshotGit(root)
	if out, err := exec.Command("git", "-C", root, "config", "user.email", "b@x").CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if out, err := exec.Command("git", "-C", root, "config", "--unset", "user.name").CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if got, want := g.changed(), ".git/config (user.email, user.name)"; got != want {
		t.Errorf("changed = %q, want %q", got, want)
	}
}

func rewriteConfig(t *testing.T, root, from, to string) {
	t.Helper()
	p := filepath.Join(root, ".git", "config")
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), from) {
		t.Fatalf("config has no %q", from)
	}
	if err := os.WriteFile(p, []byte(strings.Replace(string(data), from, to, 1)), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestGitGuardValuelessKeyIsNotAnEmptyValue(t *testing.T) {
	t.Parallel()
	root := guardRepo(t)
	appendConfig(t, root, "[core]\n\tfsmonitor\n")
	g := snapshotGit(root)
	rewriteConfig(t, root, "\tfsmonitor\n", "\tfsmonitor =\n")
	if got, want := g.changed(), ".git/config (core.fsmonitor)"; got != want {
		t.Errorf("bare -> empty: changed = %q, want %q", got, want)
	}
}

func TestGitGuardRepeatedKeyOrderMatters(t *testing.T) {
	t.Parallel()
	root := guardRepo(t)
	appendConfig(t, root, "[remote \"o\"]\n\tfetch = a\n\tfetch = b\n")
	g := snapshotGit(root)
	rewriteConfig(t, root, "\tfetch = a\n\tfetch = b\n", "\tfetch = b\n\tfetch = a\n")
	if got, want := g.changed(), ".git/config (remote.o.fetch)"; got != want {
		t.Errorf("reordered: changed = %q, want %q", got, want)
	}
}
