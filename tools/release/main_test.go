package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type repo struct {
	t   *testing.T
	dir string
}

func newRepo(t *testing.T) *repo {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	r := &repo{t: t, dir: t.TempDir()}
	r.git("init", "-q", "-b", "main")
	r.git("config", "user.name", "Test")
	r.git("config", "user.email", "test@example.invalid")
	r.git("config", "commit.gpgsign", "false")
	r.git("config", "tag.gpgSign", "false")
	return r
}

func (r *repo) git(args ...string) string {
	r.t.Helper()
	out, err := exec.Command("git", append([]string{"-C", r.dir}, args...)...).CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func (r *repo) write(rel, content string) {
	r.t.Helper()
	p := filepath.Join(r.dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		r.t.Fatal(err)
	}
}

// commit commits everything; brief, when non-empty, becomes the trailer.
func (r *repo) commit(subject, brief string) {
	r.t.Helper()
	r.git("add", "-A")
	if brief == "" {
		r.git("commit", "-q", "-m", subject)
		return
	}
	r.git("commit", "-q", "-m", subject, "-m", "Vloop-Brief: "+brief+"\nCo-Authored-By: T <t@example.invalid>")
}

func (r *repo) brief(name, desc, heading string) {
	s := "---\nname: " + name + "\ndescription: " + desc + "\nkind: brief\n---\n"
	if heading != "" {
		s += heading + "\n\n## Executive summary\n"
	}
	r.write("docs/briefs/"+name+".md", s+"\nBody.\n")
}

func (r *repo) notes(from, to string) (int, string, string) {
	var o, e bytes.Buffer
	code := run([]string{"notes", from, to}, r.dir, &o, &e)
	return code, o.String(), e.String()
}

func TestNotes(t *testing.T) {
	r := newRepo(t)
	r.write("cmd/x.go", "package main\n")
	r.commit("Initial", "")
	r.git("tag", "v1")
	r.brief("alpha", "Alpha first", "# Brief — vloop A1: the alpha")
	r.commit("Alpha drafted", "")
	r.write("cmd/a.go", "package main\n")
	r.commit("A1: the alpha (#2)", "alpha")
	r.write("README.md", "x\n")
	r.commit("Fix the README (#3)", "")
	r.write("docs/n.md", "x\n")
	r.write(".vloop/c.toml", "x\n")
	r.write("CLAUDE.md", "x\n")
	r.commit("Notes only (#4)", "")
	r.write("plugin.json", "{}\n")
	r.commit("Release v1.1 (#5)", "")
	r.brief("beta", "Beta described", "")
	r.write("tools/b.sh", "b\n")
	r.commit("B2 (#6)", "beta")
	r.write("cmd/a2.go", "package main\n")
	r.commit("Alpha again (#7)", "alpha")
	r.brief("alpha", "Alpha final", "# Brief — vloop A1: the alpha")
	r.brief("gamma", "Gamma described", "# Brief — Gamma lands")
	r.write("cmd/g.go", "package main\n")
	r.commit("Gamma (#8)", "gamma")
	r.git("tag", "v2")

	want := "## Briefs\n\n" +
		"- **A1: the alpha** — Alpha final\n" +
		"- **beta** — Beta described\n" +
		"- **Gamma lands** — Gamma described\n" +
		"\n## Other changes\n\n" +
		"- Fix the README (#3)\n"
	code, out, errs := r.notes("v1", "v2")
	if code != 0 || out != want || errs != "" {
		t.Fatalf("notes v1 v2: exit %d\n%s\n--- stderr\n%s\nwant\n%s", code, out, errs, want)
	}

	// Up to and including <to>; the description is read at <to>.
	r.git("tag", "mid", r.git("rev-parse", "v2~2"))
	code, out, _ = r.notes("v1", "mid")
	want = "## Briefs\n\n- **A1: the alpha** — Alpha first\n- **beta** — Beta described\n\n## Other changes\n\n- Fix the README (#3)\n"
	if code != 0 || out != want {
		t.Fatalf("notes v1 mid: exit %d\n%s\nwant\n%s", code, out, want)
	}
}

func TestNotesEmptySections(t *testing.T) {
	r := newRepo(t)
	r.write("cmd/x.go", "package main\n")
	r.commit("Initial", "")
	r.git("tag", "a")
	r.write("docs/n.md", "x\n")
	r.commit("Docs only", "")
	r.write("plugin.json", "{}\n")
	r.commit("Release v9 (#1)", "")
	r.git("tag", "b")
	if code, out, _ := r.notes("a", "b"); code != 0 || out != "No changes.\n" {
		t.Fatalf("no changes: exit %d %q", code, out)
	}
	if code, out, _ := r.notes("b", "b"); code != 0 || out != "No changes.\n" {
		t.Fatalf("empty range: exit %d %q", code, out)
	}
	r.write("cmd/y.go", "package main\n")
	r.commit("Only other (#2)", "")
	if code, out, _ := r.notes("b", "HEAD"); code != 0 || out != "## Other changes\n\n- Only other (#2)\n" {
		t.Fatalf("other only: exit %d %q", code, out)
	}
	r.brief("solo", "Solo", "# Brief — vloop S: solo")
	r.write("cmd/z.go", "package main\n")
	r.commit("Solo (#3)", "solo")
	if code, out, _ := r.notes("HEAD~1", "HEAD"); code != 0 || out != "## Briefs\n\n- **S: solo** — Solo\n" {
		t.Fatalf("briefs only: exit %d %q", code, out)
	}
}

func TestNotesRefusals(t *testing.T) {
	r := newRepo(t)
	r.write("a", "a\n")
	r.commit("Initial", "")
	var o, e bytes.Buffer
	if code := run([]string{"notes", "HEAD"}, r.dir, &o, &e); code != 2 {
		t.Fatalf("one argument: exit %d, want 2", code)
	}
	if code := run([]string{"frobnicate"}, r.dir, &o, &e); code != 2 {
		t.Fatalf("unknown subcommand: exit %d, want 2", code)
	}
	code, out, errs := r.notes("HEAD", "no-such-revision")
	if code != 1 || out != "" || !strings.HasPrefix(errs, "release: ") || strings.Count(errs, "\n") != 1 {
		t.Fatalf("unknown revision: exit %d stdout %q stderr %q", code, out, errs)
	}
}
