package defect

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mvelosop/vloop/internal/config"
)

func repo(t *testing.T, branch string) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("GIT_CONFIG_GLOBAL", "/dev/null")
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	g(t, dir, "init", "-q", "-b", branch, ".")
	g(t, dir, "config", "user.name", "t")
	g(t, dir, "config", "user.email", "t@example.com")
	return dir
}

func g(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func write(t *testing.T, dir, rel, text string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

const briefPath = "docs/briefs/B1.loop-brief.md"

func brief(status string) string {
	return "---\nname: B1.loop-brief\nstatus: " + status + "\n---\n# B\n"
}

var at = time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

func TestAddWritesFileAndSlug(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, briefPath, brief("consumed"))
	p, err := Add(dir, NewInput{Summary: "Util: Drops -- a LINE!! in main.go", FoundBy: "user", Brief: "B1.loop-brief"}, at)
	if err != nil {
		t.Fatal(err)
	}
	want := ".vloop/defects/D" + at.Format("20060102-1504") + "-util-drops-a-line-in-main-go.md"
	if p != want {
		t.Fatalf("path %q, want %q", p, want)
	}
	b, _ := os.ReadFile(filepath.Join(dir, p))
	if !strings.HasPrefix(string(b), "---\nid: D") || !strings.HasSuffix(string(b), "\n---\nUtil: Drops -- a LINE!! in main.go\n") {
		t.Fatalf("file:\n%s", b)
	}
	ds, err := List(dir, "")
	if err != nil || len(ds) != 1 || ds[0].Origin != "work" || ds[0].Kind != "bug" || ds[0].Severity != "medium" || ds[0].Status != "open" || ds[0].Summary != "Util: Drops -- a LINE!! in main.go" {
		t.Fatalf("list: %+v %v", ds, err)
	}
}

func TestAddSlugCap(t *testing.T) {
	if got := Slug("Parsing everything always fails whenever something happens"); got != "parsing-everything-always-fails-whenever" {
		t.Fatal(got)
	}
	if got := Slug("a b c d e f g h i j k l m n o p q r s t u v w x y z"); len(got) > 40 || strings.HasSuffix(got, "-") {
		t.Fatalf("%q", got)
	}
}

func TestAddCollisionNeverOverwrites(t *testing.T) {
	dir := t.TempDir()
	in := NewInput{Summary: "twice over", FoundBy: "operator", Brief: "B1.loop-brief"}
	p1, _ := Add(dir, in, at)
	p2, _ := Add(dir, in, at)
	p3, _ := Add(dir, in, at)
	if p2 != strings.TrimSuffix(p1, ".md")+"-2.md" || p3 != strings.TrimSuffix(p1, ".md")+"-3.md" {
		t.Fatalf("%s %s %s", p1, p2, p3)
	}
}

func TestAddValidation(t *testing.T) {
	err := Validate("found-by", "martian")
	var inv *config.InvalidValueError
	if !errors.As(err, &inv) || err.Error() != `invalid value "martian" for found-by: want one of gate, review, operator, user` {
		t.Fatalf("%v", err)
	}
	for _, f := range [][2]string{{"origin", "x"}, {"kind", "x"}, {"severity", "x"}, {"status", "x"}} {
		if Validate(f[0], f[1]) == nil {
			t.Errorf("%v accepted", f)
		}
	}
	dir := t.TempDir()
	if CheckBrief(dir, "B1.loop-brief") == nil || CheckBrief(dir, "notes") == nil {
		t.Fatal("unknown and non-loop briefs must be refused")
	}
	write(t, dir, briefPath, brief("ready"))
	if err := CheckBrief(dir, "B1.loop-brief"); err != nil {
		t.Fatal(err)
	}
}

func TestAddSetKeepsFileOnRefusal(t *testing.T) {
	dir := t.TempDir()
	p, _ := Add(dir, NewInput{Summary: "x", FoundBy: "user", Brief: "B1.loop-brief"}, at)
	id := strings.TrimSuffix(filepath.Base(p), ".md")
	before, _ := os.ReadFile(filepath.Join(dir, p))
	if err := Set(dir, id, "status", "bogus"); err == nil {
		t.Fatal("bogus status accepted")
	}
	if err := Set(dir, id, "summary", "y"); err == nil {
		t.Fatal("summary is not settable")
	}
	after, _ := os.ReadFile(filepath.Join(dir, p))
	if string(before) != string(after) {
		t.Fatal("refused set changed the file")
	}
	for _, kv := range [][2]string{{"status", "fixed"}, {"fixed-by", "B2.loop-brief"}, {"case", "a_test.go: TestX"}, {"severity", "high"}} {
		if err := Set(dir, id, kv[0], kv[1]); err != nil {
			t.Fatal(err)
		}
	}
	ds, _ := List(dir, "B1.loop-brief")
	if len(ds) != 1 || ds[0].Status != "fixed" || ds[0].FixedBy != "B2.loop-brief" || ds[0].Case != "a_test.go: TestX" || ds[0].Severity != "high" || ds[0].Summary != "x" {
		t.Fatalf("%+v", ds)
	}
}

// blameRepo builds a main with one squash-like commit that consumes the brief.
func blameRepo(t *testing.T, branch string, trailer bool) (string, string) {
	dir := repo(t, branch)
	write(t, dir, briefPath, brief("ready"))
	write(t, dir, "notes.txt", "scratch\n")
	g(t, dir, "add", "-A")
	g(t, dir, "commit", "-q", "-m", "init")
	write(t, dir, briefPath, brief("consumed"))
	write(t, dir, "util.go", "package main\n")
	g(t, dir, "add", "-A")
	if trailer {
		g(t, dir, "commit", "-q", "-m", "B1", "-m", "Vloop-Brief: B1.loop-brief")
	} else {
		g(t, dir, "commit", "-q", "-m", "B1")
	}
	return dir, g(t, dir, "rev-parse", "--short=7", "HEAD")
}

func TestBlameTrailerAndConsumed(t *testing.T) {
	dir, s7 := blameRepo(t, "main", true)
	a, err := Blame(dir, "util.go", 1)
	if err != nil || a.String() != "attributed to B1.loop-brief (trailer on "+s7+")" {
		t.Fatalf("%v %v", a, err)
	}
	dir, s7 = blameRepo(t, "main", false)
	a, err = Blame(dir, "util.go", 1)
	if err != nil || a.String() != "attributed to B1.loop-brief (consumed in "+s7+")" {
		t.Fatalf("%v %v", a, err)
	}
}

func TestBlameUnattributable(t *testing.T) {
	dir, _ := blameRepo(t, "main", true)
	_, err := Blame(dir, "notes.txt", 1)
	var un *ErrUnattributable
	if !errors.As(err, &un) || err.Error() != "cannot attribute notes.txt:1 to a loop brief — pass --brief" {
		t.Fatalf("%v", err)
	}
	if _, err := Blame(dir, "missing.go", 1); !errors.As(err, &un) {
		t.Fatalf("%v", err)
	}
}

func TestBlameReadsDefaultBranchNotHead(t *testing.T) {
	dir, s7 := blameRepo(t, "main", true)
	g(t, dir, "checkout", "-q", "-b", "work")
	write(t, dir, "util.go", "package changed\n")
	g(t, dir, "commit", "-q", "-am", "edit on work")
	a, err := Blame(dir, "util.go", 1)
	if err != nil || a.SHA7 != s7 {
		t.Fatalf("%v %v", a, err)
	}
}

func TestBlameMasterAndOriginHeadFallbacks(t *testing.T) {
	dir, s7 := blameRepo(t, "master", true)
	if a, err := Blame(dir, "util.go", 1); err != nil || a.SHA7 != s7 {
		t.Fatalf("master: %v %v", a, err)
	}
	dir, s7 = blameRepo(t, "main", true)
	g(t, dir, "update-ref", "refs/remotes/origin/trunk", "HEAD")
	g(t, dir, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/trunk")
	g(t, dir, "checkout", "-q", "-b", "work")
	g(t, dir, "branch", "-q", "-D", "main")
	if a, err := Blame(dir, "util.go", 1); err != nil || a.SHA7 != s7 {
		t.Fatalf("origin/HEAD: %v %v", a, err)
	}
}
