package runs

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func relRepo(t *testing.T, branch string) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("GIT_CONFIG_GLOBAL", "/dev/null")
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	relGit(t, dir, "init", "-q", "-b", branch, ".")
	relGit(t, dir, "config", "user.name", "t")
	relGit(t, dir, "config", "user.email", "t@example.com")
	return dir
}

func relGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func relBrief(t *testing.T, dir, status, msg string) {
	t.Helper()
	p := filepath.Join(dir, "docs/briefs/B1.loop-brief.md")
	os.MkdirAll(filepath.Dir(p), 0o755)
	os.WriteFile(p, []byte("---\nname: B1.loop-brief\nstatus: "+status+"\n---\n# B\n"), 0o644)
	relGit(t, dir, "add", "-A")
	relGit(t, dir, "commit", "-q", "-m", msg)
}

func TestReleaseFirstConsumedOnFirstParent(t *testing.T) {
	dir := relRepo(t, "main")
	relBrief(t, dir, "ready", "init")
	relGit(t, dir, "checkout", "-q", "-b", "side")
	relBrief(t, dir, "consumed", "side consumes")
	relGit(t, dir, "checkout", "-q", "main")
	if got, err := Release(dir, "docs/briefs/B1.loop-brief.md"); err != nil || got != "" {
		t.Fatalf("unmerged brief: got %q, %v", got, err)
	}
	relGit(t, dir, "merge", "-q", "--no-ff", "-m", "merge", "side")
	merge := relGit(t, dir, "rev-parse", "HEAD")
	relBrief(t, dir, "draft", "later edit")
	got, err := Release(dir, "docs/briefs/B1.loop-brief.md")
	if err != nil || got != merge {
		t.Fatalf("got %q, %v; want the merge %s", got, err, merge)
	}
}

func TestReleaseDefaultBranchFallbacks(t *testing.T) {
	dir := relRepo(t, "master")
	relBrief(t, dir, "consumed", "init")
	sha := relGit(t, dir, "rev-parse", "HEAD")
	if b, err := DefaultBranch(dir); err != nil || b != "master" {
		t.Fatalf("master fallback: %q, %v", b, err)
	}
	if got, _ := Release(dir, "docs/briefs/B1.loop-brief.md"); got != sha {
		t.Fatalf("release on master: %q", got)
	}
	relGit(t, dir, "branch", "main")
	if b, _ := DefaultBranch(dir); b != "main" {
		t.Fatalf("main beats master: %q", b)
	}
	relGit(t, dir, "update-ref", "refs/remotes/origin/trunk", "HEAD")
	relGit(t, dir, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/trunk")
	if b, _ := DefaultBranch(dir); b != "refs/remotes/origin/trunk" {
		t.Fatalf("origin/HEAD wins: %q", b)
	}
}

func TestReleaseNoDefaultBranch(t *testing.T) {
	dir := relRepo(t, "dev")
	relBrief(t, dir, "consumed", "init")
	if _, err := Release(dir, "docs/briefs/B1.loop-brief.md"); err == nil {
		t.Fatal("want an error without a default branch")
	}
}
