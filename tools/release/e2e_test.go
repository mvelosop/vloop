package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestReleaseSequence plays the brief's worked example: prepare's refusals,
// prepare, publish before and after the operator's squash-merge, the re-run
// and the mismatched install — with stubs on PATH and a local bare origin.
func TestReleaseSequence(t *testing.T) {
	proxyInterval = 10 * time.Millisecond
	dir, origin := prepared(t)
	stubs := os.Getenv("STUBS")
	t.Setenv("HOME", t.TempDir())
	goStub(t, stubs)
	released := filepath.Join(stubs, "released")
	gh := "#!/bin/sh\necho \"$*\" >>\"" + filepath.Join(stubs, "gh.log") + "\"\n" +
		"case \"$1 $2\" in \"pr create\") echo https://github.example/pull/7; exit 0;; " +
		"\"release view\") [ -f \"" + released + "\" ] && { echo https://github.example/releases/tag/v0.12.0; exit 0; }; exit 1;; " +
		"\"release create\") touch \"" + released + "\"; echo https://github.example/releases/tag/v0.12.0; exit 0;; esac\n"
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
	pj := "{\n  \"name\": \"vloop\",\n  \"version\": \"0.11.0\"\n}\n"
	p := filepath.Join(dir, "plugin/.claude-plugin/plugin.json")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(pj), 0o644); err != nil {
		t.Fatal(err)
	}
	g(dir, "add", "-A")
	g(dir, "commit", "-q", "-m", "Add plugin")
	g(dir, "push", "-q", "origin", "main")
	main0 := g(origin, "rev-parse", "refs/heads/main")

	refused := func(code int, want string, cmd func() (int, string, string)) {
		t.Helper()
		before := state(t, dir, origin)
		c, _, errs := cmd()
		if c != code || strings.TrimSpace(errs) != want {
			t.Errorf("exit %d stderr %q; want exit %d %q", c, errs, code, want)
		}
		if state(t, dir, origin) != before {
			t.Errorf("a refusal wrote something")
		}
	}
	prepare := func(v string) func() (int, string, string) {
		return func() (int, string, string) { return prep(dir, v) }
	}
	refused(2, "release: 0.12.0 is not a version: vX.Y.Z or vX.Y.Z-<pre>", prepare("0.12.0"))
	refused(1, "release: v2.0.0 needs the module path to end in /v2, and github.com/mvelosop/vloop does not", prepare("v2.0.0"))
	refused(1, "release: v1.0.0 is retracted in go.mod", prepare("v1.0.0"))
	refused(1, "release: v0.11.0 is not after the latest release v0.11.0", prepare("v0.11.0"))
	g(dir, "checkout", "-q", "-b", "topic")
	refused(1, "release: prepare runs on main, not topic", prepare("v0.12.0"))
	g(dir, "checkout", "-q", "main")

	code, out, errs := prep(dir, "v0.12.0")
	if code != 0 || errs != "" || out != "prepared release/v0.12.0: https://github.example/pull/7\n"+
		"squash-merge it, then: go run ./tools/release publish v0.12.0\n" {
		t.Fatalf("prepare: exit %d stdout %q stderr %q", code, out, errs)
	}
	rb := g(origin, "rev-parse", "refs/heads/release/v0.12.0")
	if g(origin, "rev-list", "--count", main0+".."+rb) != "1" || g(origin, "log", "-1", "--format=%s", rb) != "Release v0.12.0" ||
		g(origin, "diff", "--name-only", main0, rb) != "plugin/.claude-plugin/plugin.json" {
		t.Errorf("the release branch is not one plugin.json commit 'Release v0.12.0'")
	}

	// Before the merge: refused, nothing written.
	os.WriteFile(filepath.Join(stubs, "proxy-v0.12.0"), nil, 0o644)
	refused(1, "release: no release commit for v0.12.0 on origin/main — merge the release PR first",
		func() (int, string, string) { return pub(dir, "v0.12.0") })
	if g(origin, "tag", "--list", "v0.12.0") != "" {
		t.Fatalf("a refused publish tagged")
	}

	// The operator squash-merges the pushed branch.
	g(dir, "fetch", "-q", "origin")
	g(dir, "merge", "-q", "--squash", "origin/release/v0.12.0")
	g(dir, "commit", "-q", "-m", "Release v0.12.0 (#7)")
	g(dir, "push", "-q", "origin", "main")
	rc := g(dir, "rev-parse", "HEAD")

	wantTail := "published v0.12.0: https://github.example/releases/tag/v0.12.0\n" +
		"install: go install github.com/mvelosop/vloop/cmd/vloop@v0.12.0\n" +
		"then, in each repository: vloop upgrade\n" +
		"and for the plugin: claude plugin marketplace update vloop && claude plugin update vloop@vloop\n"
	code, out, errs = pub(dir, "v0.12.0")
	if code != 0 || !strings.HasSuffix(out, wantTail) {
		t.Fatalf("publish: exit %d stdout %q stderr %q", code, out, errs)
	}
	if g(origin, "cat-file", "-t", "refs/tags/v0.12.0") != "tag" || g(origin, "rev-parse", "refs/tags/v0.12.0^{commit}") != rc {
		t.Errorf("v0.12.0 is not an annotated tag on the release commit")
	}

	code, out, errs = pub(dir, "v0.12.0")
	if code != 0 || !strings.Contains(out, "tag: already done\n") || !strings.Contains(out, "release: already done\n") ||
		!strings.HasSuffix(out, wantTail) {
		t.Fatalf("re-run: exit %d stdout %q stderr %q", code, out, errs)
	}

	// An install that reports another version is refused.
	os.WriteFile(filepath.Join(stubs, "report"), []byte("0.11.0\nv0.11.0\n"), 0o644)
	code, _, errs = pub(dir, "v0.12.0")
	if code != 1 || !strings.Contains(errs, "release: go install github.com/mvelosop/vloop/cmd/vloop@v0.12.0 reports version 0.11.0, module v0.11.0") {
		t.Fatalf("mismatch: exit %d stderr %q", code, errs)
	}
}
