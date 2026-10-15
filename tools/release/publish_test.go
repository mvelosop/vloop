package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func pub(dir string, args ...string) (int, string, string) {
	var o, e bytes.Buffer
	code := run(append([]string{"publish"}, args...), dir, &o, &e)
	return code, o.String(), e.String()
}

func TestPublish(t *testing.T) {
	dir, origin := prepared(t)
	stubs := os.Getenv("STUBS")
	gh := "#!/bin/sh\necho \"$*\" >>\"" + filepath.Join(stubs, "gh.log") + "\"\n" +
		"case \"$1 $2\" in \"release view\") exit 1;; esac\necho https://github.example/releases/tag/$3\n"
	if err := os.WriteFile(filepath.Join(stubs, "gh"), []byte(gh), 0o755); err != nil {
		t.Fatal(err)
	}
	g := func(d string, a ...string) string {
		t.Helper()
		out, err := git(d, a...)
		if err != nil {
			t.Fatal(err)
		}
		return strings.TrimSpace(out)
	}
	write := func(rel, content string) {
		t.Helper()
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	push := func(msg string) string {
		t.Helper()
		g(dir, "add", "-A")
		g(dir, "commit", "-q", "-m", msg)
		g(dir, "push", "-q", "origin", "main")
		return g(dir, "rev-parse", "HEAD")
	}
	pj := func(v string) string { return "{\n  \"name\": \"vloop\",\n  \"version\": \"" + v + "\"\n}\n" }
	ghlog := func() string { b, _ := os.ReadFile(filepath.Join(stubs, "gh.log")); return string(b) }

	write("plugin/.claude-plugin/plugin.json", pj("0.11.0"))
	push("Add plugin")

	// Before the merge, and with a decoy whose plugin.json lacks the version:
	// refused, nothing written.
	want := "release: no release commit for v0.12.0 on origin/main — merge the release PR first\n"
	if code, _, errs := pub(dir, "v0.12.0"); code != 1 || errs != want {
		t.Fatalf("before merge: exit %d stderr %q", code, errs)
	}
	write("docs/prep.md", "x\n")
	push("Release v0.12.0 prep (#8)")
	if code, _, errs := pub(dir, "v0.12.0"); code != 1 || errs != want {
		t.Fatalf("decoy: exit %d stderr %q", code, errs)
	}
	if tags := g(origin, "tag", "--list", "v0.12.0"); tags != "" || ghlog() != "" {
		t.Fatalf("a refusal wrote: tag %q gh %q", tags, ghlog())
	}

	// The merge, with a brief, then a later commit: the tag goes on the merge.
	write("docs/briefs/B1-alpha.md", "# Brief — the alpha\n")
	write("plugin/.claude-plugin/plugin.json", pj("0.12.0"))
	g(dir, "add", "-A")
	g(dir, "commit", "-q", "-m", "Release v0.12.0 (#9)", "-m", "Vloop-Brief: B1-alpha")
	g(dir, "push", "-q", "origin", "main")
	rc := g(dir, "rev-parse", "HEAD")
	write("cmd/later.go", "package main\n")
	push("Later fix (#10)")
	main0 := g(origin, "rev-parse", "refs/heads/main")

	code, out, errs := pub(dir, "v0.12.0")
	if code != 0 || errs != "" {
		t.Fatalf("exit %d stderr %q", code, errs)
	}
	if !strings.HasSuffix(out, "published v0.12.0: https://github.example/releases/tag/v0.12.0\n"+
		"install: go install github.com/mvelosop/vloop/cmd/vloop@v0.12.0\n"+
		"then, in each repository: vloop upgrade\n"+
		"and for the plugin: claude plugin marketplace update vloop && claude plugin update vloop@vloop\n") {
		t.Errorf("stdout %q", out)
	}
	if g(origin, "rev-parse", "refs/tags/v0.12.0^{commit}") != rc {
		t.Errorf("tag is not on the release commit")
	}
	if msg := g(origin, "for-each-ref", "--format=%(contents)", "refs/tags/v0.12.0"); msg != "v0.12.0: the alpha" {
		t.Errorf("tag message %q", msg)
	}
	if g(origin, "rev-parse", "refs/heads/main") != main0 {
		t.Errorf("publish moved origin's main")
	}
	if log := ghlog(); !strings.Contains(log, "release create v0.12.0 --verify-tag --title v0.12.0 --notes ") || strings.Contains(log, "--prerelease") {
		t.Errorf("gh args %q", log)
	}

	// A pre-release whose tag exists elsewhere is refused and left as it was;
	// without the tag, it is created on the release commit with --prerelease.
	write("plugin/.claude-plugin/plugin.json", pj("0.12.1-rc.1"))
	g(dir, "add", "-A")
	g(dir, "commit", "-q", "-m", "Release v0.12.1-rc.1 (#11)")
	g(dir, "push", "-q", "origin", "main")
	rc2 := g(dir, "rev-parse", "HEAD")
	g(dir, "tag", "-a", "v0.12.1-rc.1", "-m", "elsewhere", main0)
	g(dir, "push", "-q", "origin", "v0.12.1-rc.1")
	code, _, errs = pub(dir, "v0.12.1-rc.1")
	if code != 1 || !strings.HasPrefix(errs, "release: tag v0.12.1-rc.1 points at "+main0+", not the release commit "+rc2) {
		t.Fatalf("tag elsewhere: exit %d stderr %q", code, errs)
	}
	if g(origin, "rev-parse", "refs/tags/v0.12.1-rc.1^{commit}") != main0 {
		t.Errorf("the refusal moved the tag")
	}
	g(dir, "push", "-q", "origin", ":refs/tags/v0.12.1-rc.1")
	g(dir, "tag", "-d", "v0.12.1-rc.1")
	if code, _, errs := pub(dir, "v0.12.1-rc.1"); code != 0 {
		t.Fatalf("pre-release: exit %d stderr %q", code, errs)
	}
	if g(origin, "rev-parse", "refs/tags/v0.12.1-rc.1^{commit}") != rc2 {
		t.Errorf("pre-release tag is not on the release commit")
	}
	if msg := g(origin, "for-each-ref", "--format=%(contents)", "refs/tags/v0.12.1-rc.1"); msg != "v0.12.1-rc.1" {
		t.Errorf("tag message %q", msg)
	}
	if !strings.Contains(ghlog(), "release create v0.12.1-rc.1 --verify-tag --title v0.12.1-rc.1 --notes ") ||
		!strings.Contains(ghlog(), " --prerelease") {
		t.Errorf("gh args %q", ghlog())
	}
}
