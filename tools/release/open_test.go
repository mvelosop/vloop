package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrepareOpensRelease(t *testing.T) {
	dir, origin := prepared(t)
	stubs := os.Getenv("STUBS")
	gh := "#!/bin/sh\necho \"$*\" >>\"" + filepath.Join(stubs, "gh.log") + "\"\n" +
		"while [ $# -gt 0 ]; do case \"$1\" in --title) printf %s \"$2\" >\"" + filepath.Join(stubs, "title") +
		"\";; --body) printf %s \"$2\" >\"" + filepath.Join(stubs, "body") + "\";; esac; shift; done\n" +
		"echo https://github.example/pull/7\n"
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
	if err := os.MkdirAll(filepath.Join(dir, "plugin/.claude-plugin"), 0o755); err != nil {
		t.Fatal(err)
	}
	pj := "{\n  \"name\": \"vloop\",\n  \"version\": \"0.11.0\",\n  \"description\": \"x\"\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "plugin/.claude-plugin/plugin.json"), []byte(pj), 0o644); err != nil {
		t.Fatal(err)
	}
	g(dir, "add", "-A")
	g(dir, "commit", "-q", "-m", "Add plugin")
	g(dir, "push", "-q", "origin", "main")
	main0 := g(dir, "rev-parse", "main")

	code, out, errs := prep(dir, "v0.12.0")
	if code != 0 || errs != "" {
		t.Fatalf("exit %d stderr %q", code, errs)
	}
	want := "prepared release/v0.12.0: https://github.example/pull/7\nsquash-merge it, then: go run ./tools/release publish v0.12.0\n"
	if out != want {
		t.Errorf("stdout %q, want %q", out, want)
	}
	rb := g(origin, "rev-parse", "refs/heads/release/v0.12.0")
	if g(origin, "rev-parse", rb+"^") != main0 || g(origin, "rev-list", "--count", main0+".."+rb) != "1" {
		t.Errorf("release branch is not one commit on main")
	}
	if s := g(origin, "log", "-1", "--format=%s", rb); s != "Release v0.12.0" {
		t.Errorf("subject %q", s)
	}
	if n := g(origin, "diff", "--numstat", main0, rb); n != "1\t1\tplugin/.claude-plugin/plugin.json" {
		t.Errorf("numstat %q", n)
	}
	if g(origin, "show", rb+":plugin/.claude-plugin/plugin.json") != strings.TrimSpace(strings.Replace(pj, "0.11.0", "0.12.0", 1)) {
		t.Errorf("plugin.json not updated by one line")
	}
	if g(dir, "rev-parse", "main") != main0 || g(dir, "rev-parse", "--abbrev-ref", "HEAD") != "main" {
		t.Errorf("main moved or not checked out")
	}
	if g(origin, "rev-parse", "refs/heads/main") != main0 {
		t.Errorf("origin main moved")
	}
	title, _ := os.ReadFile(filepath.Join(stubs, "title"))
	body, _ := os.ReadFile(filepath.Join(stubs, "body"))
	wantBody, err := notes(dir, "v0.11.0", "main")
	if err != nil {
		t.Fatal(err)
	}
	if string(title) != "Release v0.12.0" || string(body) != wantBody {
		t.Errorf("PR title %q body %q, want body %q", title, body, wantBody)
	}
}
