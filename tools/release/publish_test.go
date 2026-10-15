package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	goStub(t, stubs)
	for _, v := range []string{"v0.12.0", "v0.12.1-rc.1"} {
		if err := os.WriteFile(filepath.Join(stubs, "proxy-"+v), nil, 0o644); err != nil {
			t.Fatal(err)
		}
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

// goStub puts a go on PATH (stubs is already first) that logs its argv, serves
// <module>@<v> once stubs/proxy-<v> exists, and installs a vloop reporting <v>
// (or the two lines of stubs/report) into $GOBIN.
func goStub(t *testing.T, stubs string) {
	t.Helper()
	script := `#!/bin/sh
d="` + stubs + `"
echo "$*" >>"$d/go.log"
mv=${*##* }
v=${mv##*@}
[ -f "$d/proxy-$v" ] || { echo "unknown revision $v" >&2; exit 1; }
[ "$1" = list ] && exit 0
echo "GOBIN=$GOBIN" >>"$d/go.log"
rv=${v#v}; rm=$v
if [ -f "$d/report" ]; then rv=$(sed -n 1p "$d/report"); rm=$(sed -n 2p "$d/report"); fi
mkdir -p "$GOBIN"
printf '#!/bin/sh\necho '"'"'{"version":"%s","module":"%s"}'"'"'\n' "$rv" "$rm" >"$GOBIN/vloop"
chmod +x "$GOBIN/vloop"
`
	if err := os.WriteFile(filepath.Join(stubs, "go"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestPublishProxyAndRerun(t *testing.T) {
	proxyInterval = 10 * time.Millisecond
	dir, origin := prepared(t)
	stubs := os.Getenv("STUBS")
	goStub(t, stubs)
	gh := "#!/bin/sh\necho \"$*\" >>\"" + filepath.Join(stubs, "gh.log") + "\"\n" +
		"case \"$1 $2\" in \"release view\") [ -f \"" + filepath.Join(stubs, "released") + "\" ] && echo https://github.example/r && exit 0; exit 1;; " +
		"\"release create\") touch \"" + filepath.Join(stubs, "released") + "\";; esac\necho https://github.example/r\n"
	if err := os.WriteFile(filepath.Join(stubs, "gh"), []byte(gh), 0o755); err != nil {
		t.Fatal(err)
	}
	g := func(a ...string) string {
		t.Helper()
		out, err := git(dir, a...)
		if err != nil {
			t.Fatal(err)
		}
		return strings.TrimSpace(out)
	}
	pj := "{\n  \"name\": \"vloop\",\n  \"version\": \"0.12.0\"\n}\n"
	p := filepath.Join(dir, "plugin/.claude-plugin/plugin.json")
	os.MkdirAll(filepath.Dir(p), 0o755)
	os.WriteFile(p, []byte(pj), 0o644)
	g("add", "-A")
	g("commit", "-q", "-m", "Release v0.12.0 (#1)")
	g("push", "-q", "origin", "main")
	file := func(n string) string { b, _ := os.ReadFile(filepath.Join(stubs, n)); return string(b) }

	// --dry-run writes no tag and no release, prints its steps.
	code, out, errs := pub(dir, "v0.12.0", "--dry-run")
	if code != 0 || errs != "" || !strings.Contains(out, "tag: would tag") || strings.Contains(file("go.log"), "install") {
		t.Fatalf("dry run: exit %d out %q err %q", code, out, errs)
	}
	if g("ls-remote", "--tags", "origin", "v0.12.0") != "" || strings.Contains(file("gh.log"), "release create") {
		t.Fatalf("dry run wrote")
	}

	// The proxy does not serve the version: refused, tag and release kept.
	code, _, errs = pub(dir, "v0.12.0", "--wait", "50ms")
	if code != 1 || errs != "release: the module proxy does not serve v0.12.0 after 50ms\n" {
		t.Fatalf("proxy: exit %d stderr %q", code, errs)
	}
	if !strings.Contains(file("go.log"), "list -m github.com/mvelosop/vloop@v0.12.0") || strings.Contains(file("go.log"), "install") {
		t.Fatalf("go.log %q", file("go.log"))
	}
	if g("ls-remote", "--tags", "origin", "v0.12.0") == "" {
		t.Fatalf("tag not pushed before the install check")
	}

	// Served, but the binary reports another version.
	os.WriteFile(filepath.Join(stubs, "proxy-v0.12.0"), nil, 0o644)
	os.WriteFile(filepath.Join(stubs, "report"), []byte("0.11.0\nv0.11.0\n"), 0o644)
	code, out, errs = pub(dir, "v0.12.0")
	if code != 1 || !strings.Contains(errs, "release: go install github.com/mvelosop/vloop/cmd/vloop@v0.12.0 reports version 0.11.0, module v0.11.0") {
		t.Fatalf("mismatch: exit %d stderr %q", code, errs)
	}
	if !strings.Contains(out, "tag: already done") || !strings.Contains(out, "release: already done") {
		t.Errorf("re-run stdout %q", out)
	}
	gobin := ""
	for _, l := range strings.Split(file("go.log"), "\n") {
		if strings.HasPrefix(l, "GOBIN=") {
			gobin = strings.TrimPrefix(l, "GOBIN=")
		}
	}
	if gobin == "" || strings.HasPrefix(gobin, dir) {
		t.Errorf("GOBIN %q is empty or inside the repository", gobin)
	}

	// Fixed, then fully done: only the install check runs again.
	os.Remove(filepath.Join(stubs, "report"))
	for i := 0; i < 2; i++ {
		code, out, errs = pub(dir, "v0.12.0")
		if code != 0 || errs != "" || !strings.Contains(out, "published v0.12.0: https://github.example/r\n") ||
			!strings.Contains(out, "tag: already done") || !strings.Contains(out, "release: already done") {
			t.Fatalf("run %d: exit %d out %q err %q", i, code, out, errs)
		}
	}
	if n := strings.Count(file("gh.log"), "release create"); n != 1 {
		t.Errorf("release created %d times", n)
	}
	_ = origin
}
