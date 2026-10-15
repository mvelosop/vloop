package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSemverOrder(t *testing.T) {
	order := []string{"v0.11.0-rc.1", "v0.11.0-rc.2", "v0.11.0-rc.10", "v0.11.0", "v0.11.1", "v0.12.0", "v1.0.0"}
	for i := range order {
		for j := range order {
			a, _ := parseVersion(order[i])
			b, ok := parseVersion(order[j])
			if !ok {
				t.Fatalf("%s is not parsed", order[j])
			}
			want := sign(i - j)
			if got := a.cmp(b); got != want {
				t.Errorf("cmp(%s, %s) = %d, want %d", order[i], order[j], got, want)
			}
		}
	}
	for _, bad := range []string{"0.12.0", "v0.12", "v0.12.0-", "v01.2.3", "nightly"} {
		if _, ok := parseVersion(bad); ok {
			t.Errorf("%q parsed as a version", bad)
		}
	}
}

// prepared builds a repository on main with an origin, a stub claude on PATH
// and a [[check]] that logs its runs; it returns the clone's dir.
func prepared(t *testing.T) (dir, origin string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("uses sh stubs")
	}
	r := newRepo(t)
	origin = filepath.Join(t.TempDir(), "origin.git")
	if out, err := exec.Command("git", "init", "-q", "--bare", "-b", "main", origin).CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	r.write("go.mod", "module github.com/mvelosop/vloop\n\ngo 1.22\n\nretract v1.0.0\n\nretract [v0.13.0, v0.13.2]\n")
	r.write(".vloop/config.toml", "[[check]]\nname = \"unit\"\npaths = [\"**\"]\nrun = \"echo said-hello; test ! -f ../check-fail\"\n")
	r.write("cmd/x.go", "package main\n")
	r.commit("Initial", "")
	r.git("tag", "-a", "v0.11.0", "-m", "v0.11.0")
	r.git("tag", "nightly")
	r.git("remote", "add", "origin", origin)
	r.git("push", "-q", "origin", "main", "--tags")
	bin := t.TempDir()
	stub := "#!/bin/sh\necho \"$*\" >>\"" + filepath.Join(bin, "claude.log") + "\"\n[ -f \"" + filepath.Join(bin, "fail") + "\" ] && { echo planted >&2; exit 1; }\nexit 0\n"
	if err := os.WriteFile(filepath.Join(bin, "claude"), []byte(stub), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("STUBS", bin)
	return r.dir, origin
}

func state(t *testing.T, dir, origin string) string {
	t.Helper()
	g := func(d string, a ...string) string {
		out, err := exec.Command("git", append([]string{"-C", d}, a...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", a, err, out)
		}
		return string(out)
	}
	return g(origin, "for-each-ref") + g(dir, "for-each-ref", "refs/heads", "refs/tags") + g(dir, "status", "--porcelain")
}

func prep(dir string, args ...string) (int, string, string) {
	var o, e bytes.Buffer
	code := run(append([]string{"prepare"}, args...), dir, &o, &e)
	return code, o.String(), e.String()
}

func TestPrepareRefusals(t *testing.T) {
	dir, origin := prepared(t)
	no := func(wantCode int, want string, args ...string) {
		t.Helper()
		before := state(t, dir, origin)
		code, _, errs := prep(dir, args...)
		if code != wantCode || strings.TrimSpace(errs) != want {
			t.Errorf("prepare %v: exit %d stderr %q; want exit %d %q", args, code, errs, wantCode, want)
		}
		if after := state(t, dir, origin); after != before {
			t.Errorf("prepare %v wrote something:\n%s\n---\n%s", args, before, after)
		}
	}
	no(2, "release: 0.12.0 is not a version: vX.Y.Z or vX.Y.Z-<pre>", "0.12.0")
	no(1, "release: v2.0.0 needs the module path to end in /v2, and github.com/mvelosop/vloop does not", "v2.0.0")
	no(1, "release: v1.0.0 is retracted in go.mod", "v1.0.0")
	no(1, "release: v0.13.1 is retracted in go.mod", "v0.13.1")
	no(1, "release: v0.11.0 is not after the latest release v0.11.0", "v0.11.0")
	no(1, "release: v0.11.0-rc.2 is not after the latest release v0.11.0", "v0.11.0-rc.2")

	// A tag only on origin refuses.
	g := func(a ...string) {
		t.Helper()
		if out, err := exec.Command("git", append([]string{"-C", dir}, a...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", a, err, out)
		}
	}
	g("tag", "v0.14.0")
	g("push", "-q", "origin", "v0.14.0")
	g("tag", "-d", "v0.14.0")
	no(1, "release: tag v0.14.0 already exists", "v0.14.0")

	// The repository's state, in order.
	g("checkout", "-q", "-b", "feature")
	no(1, "release: prepare runs on main, not feature", "v0.12.0")
	g("checkout", "-q", "main")
	if err := os.WriteFile(filepath.Join(dir, "cmd/x.go"), []byte("dirty\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	no(1, "release: the working tree is not clean", "v0.12.0")
	g("checkout", "--", "cmd/x.go")
	if err := os.WriteFile(filepath.Join(dir, "cmd/local.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	g("add", "-A")
	g("commit", "-q", "-m", "Local work")
	no(1, "release: main is not origin/main — pull or push first", "v0.12.0")
	g("push", "-q", "origin", "main")

	// The checks, then claude plugin validate; neither runs after a refusal above.
	if err := os.WriteFile(filepath.Join(filepath.Dir(dir), "check-fail"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, errs := prep(dir, "v0.12.0")
	if code != 1 || !strings.Contains(errs, "said-hello") || !strings.Contains(errs, "release: check unit failed") {
		t.Errorf("failed check: exit %d stderr %q", code, errs)
	}
	if _, err := os.Stat(filepath.Join(os.Getenv("STUBS"), "claude.log")); err == nil {
		t.Error("claude plugin validate ran after a check failed")
	}
	os.Remove(filepath.Join(filepath.Dir(dir), "check-fail"))
	if err := os.WriteFile(filepath.Join(os.Getenv("STUBS"), "fail"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	no(1, "planted\nrelease: claude plugin validate plugin failed", "v0.12.0")
}

func TestPrepareDryRun(t *testing.T) {
	dir, origin := prepared(t)
	before := state(t, dir, origin)
	code, out, errs := prep(dir, "v0.12.0", "--dry-run")
	if code != 0 || out == "" || errs != "" {
		t.Fatalf("dry run: exit %d stdout %q stderr %q", code, out, errs)
	}
	if after := state(t, dir, origin); after != before {
		t.Errorf("--dry-run wrote something:\n%s\n---\n%s", before, after)
	}
	log, _ := os.ReadFile(filepath.Join(os.Getenv("STUBS"), "claude.log"))
	if !strings.HasPrefix(string(log), "plugin validate plugin") {
		t.Errorf("claude was called with %q", log)
	}
}
