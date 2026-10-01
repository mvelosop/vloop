package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const b4Name = b3Brief + ".loop-brief"
const b4File = "docs/briefs/" + b4Name + ".md"

// b4Fixture is B3's worked example before the merge (HEAD on the work branch,
// the run complete, brief ready), with an origin carrying credentials and a
// git identity so `close` can commit.
func b4Fixture(t *testing.T, o b3Opts) *scratch {
	t.Helper()
	s, _, _ := b3PreMerge(t, o)
	at := func(args ...string) string { return b3Git(t, s.dir, "2026-01-01T09:15:00Z", args...) }
	at("remote", "add", "origin", "https://user:tok@example.com/acme/shop.git")
	at("config", "user.name", "t")
	at("config", "user.email", "t@example.com")
	at("config", "commit.gpgsign", "false")
	return s
}

func b4Porcelain(t *testing.T, s *scratch) string {
	return b3Git(t, s.dir, "2026-01-01T09:15:00Z", "status", "--porcelain", "--untracked-files=all")
}

func b4Refs(t *testing.T, s *scratch) string {
	return b3Git(t, s.dir, "2026-01-01T09:15:00Z", "for-each-ref", "--format=%(refname) %(objectname)")
}

func b4Names(t *testing.T, s *scratch, args ...string) []string {
	out := b3Git(t, s.dir, "2026-01-01T09:15:00Z", args...)
	l := strings.Split(out, "\n")
	for i := range l {
		l[i] = strings.TrimSpace(l[i])
	}
	return l
}

func b4Err(t *testing.T, r result, code int, err string) {
	t.Helper()
	if r.code != code || r.out != "" || r.err != err {
		t.Fatalf("got code=%d out=%q err=%q, want code=%d err=%q", r.code, r.out, r.err, code, err)
	}
}

func TestWorkedExampleB4Close(t *testing.T) {
	s := b4Fixture(t, b3Opts{})
	before := b4Refs(t, s)

	b4Err(t, s.run(nil, "brief", "close", b4Name), 2,
		"vloop: say what you found before merge: --finding \"<summary>\" (repeatable) or --no-findings\n")

	r := s.run(nil, "brief", "close", b4Name, "--finding", "README omits the flag", "--dry-run")
	c := b3Collapse(r.out)
	stamp := regexp.MustCompile(`would record \.vloop/defects/(D\d{8}-\d{4}-readme-omits-the-flag)\.md`).FindStringSubmatch(c)
	if r.code != 0 || r.err != "" || stamp == nil {
		t.Fatalf("dry run: %+v", r)
	}
	defect := stamp[1]
	for _, w := range []string{
		b3Brief + " consumed · not merged",
		"would record .vloop/defects/" + defect + ".md",
		"would write .vloop/state/metrics/" + b3Brief + ".json",
		"would update " + b4File,
		"would commit [vloop] close " + b3Brief,
	} {
		if !hasLine(c, w) {
			t.Fatalf("dry run lacks %q:\n%s", w, c)
		}
	}
	if p := b4Porcelain(t, s); p != "" || b4Refs(t, s) != before {
		t.Fatalf("dry run changed the repository: %q", p)
	}

	r = s.run(nil, "brief", "close", b4Name, "--finding", "README omits the flag")
	if r.code != 0 || r.err != "" {
		t.Fatalf("close: %+v", r)
	}
	c = b3Collapse(r.out)
	lines := strings.Split(c, "\n")
	for _, w := range []string{
		b3Brief + " consumed · not merged",
		"tasks 2 planned (brief said 2–3) · 2 done · 0 blocked · first-pass 1/2",
		"defects in-loop 1 · operator 1 · escaped 0 · removal efficiency 100%",
		"recorded .vloop/defects/" + defect + ".md",
		"wrote .vloop/state/metrics/" + b3Brief + ".json",
		"updated " + b4File,
	} {
		if !hasLine(c, w) {
			t.Fatalf("close lacks %q:\n%s", w, c)
		}
	}
	sha := b3Git(t, s.dir, "2026-01-01T09:15:00Z", "rev-parse", "--short", "HEAD")
	if !hasLine(c, "committed "+sha+" [vloop] close "+b3Brief) {
		t.Fatalf("close lacks the committed line for %s:\n%s", sha, c)
	}
	if last := lines[len(lines)-1]; last != "squash-merge with the trailer: Vloop-Brief: "+b4Name {
		t.Fatalf("last line = %q", last)
	}
	if got := b4Names(t, s, "log", "-1", "--format=%B")[0]; got != "[vloop] close "+b3Brief {
		t.Fatalf("subject = %q", got)
	}
	body := b3Git(t, s.dir, "2026-01-01T09:15:00Z", "log", "-1", "--format=%B")
	if !strings.HasSuffix(body, "\n\nVloop-Brief: "+b4Name) {
		t.Fatalf("message %q lacks the trailer after a blank line", body)
	}
	if got := strings.Join(b4Names(t, s, "show", "--name-only", "--format=", "HEAD"), "\n"); got !=
		".vloop/defects/"+defect+".md\n.vloop/state/metrics/"+b3Brief+".json\n"+b4File {
		t.Fatalf("commit files:\n%s", got)
	}
	if p := b4Porcelain(t, s); p != "" {
		t.Fatalf("tree not clean after close: %q", p)
	}
	if n := strings.Count(b4Refs(t, s), "\n"); n != strings.Count(before, "\n") {
		t.Fatalf("refs changed: %q -> %q", before, b4Refs(t, s))
	}
	if br := b3Git(t, s.dir, "2026-01-01T09:15:00Z", "branch", "--show-current"); br != b3Brief {
		t.Fatalf("branch = %q", br)
	}

	snap := ".vloop/state/metrics/" + b3Brief + ".json"
	expect(t, s.run(nil, "schema", "validate", "metrics/v1", snap), 0, snap+": ok\n", "")
	if raw := s.read(snap); !strings.HasSuffix(raw, "}\n") || !strings.Contains(raw, "\n  \"") {
		t.Fatalf("snapshot is not 2-space indented with a trailing newline: %q", raw)
	}
	text := s.read(b4File)
	if !strings.Contains(text, "status: consumed\n") ||
		!strings.Contains(text, "**Status:** consumed — closed ") ||
		!strings.Contains(text, " as run "+b3Brief+". **Do not re-plan from this brief.**") ||
		!strings.Contains(text, "<!-- vloop:run-record:begin -->\n## Run record\n") ||
		!strings.Contains(text, "Recompute with: vloop metrics "+b4Name) ||
		!strings.Contains(text, "- "+defect+" — bug, work, found by operator: README omits the flag") {
		t.Fatalf("brief after close:\n%s", text)
	}

	b4Err(t, s.run(nil, "brief", "close", b4Name, "--no-findings"), 1, "vloop: "+b4Name+" is already consumed\n")

	// The export: five records.
	r = s.run(nil, "metrics", "export")
	if r.code != 0 || r.err != "" {
		t.Fatalf("export: %+v", r)
	}
	var got []string
	var first map[string]any
	for i, l := range strings.Split(strings.TrimRight(r.out, "\n"), "\n") {
		var m map[string]any
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			t.Fatalf("export line %d: %v", i, err)
		}
		if i == 0 {
			first = m
		}
		id := m["id"]
		if m["type"] == "brief" {
			id = m["run_id"]
		}
		got = append(got, m["type"].(string)+" "+id.(string))
		if m["schema"] != "export/v1" {
			t.Fatalf("line %d schema = %v", i, m["schema"])
		}
	}
	want := []string{"brief " + b3Brief, "task T1", "task T2", "defect " + b3Brief + "/i2-gate", "defect " + defect}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("export records %v, want %v", got, want)
	}
	repo, _ := json.Marshal(first["repo"])
	if string(repo) != `{"name":"shop","remote":"https://example.com/acme/shop.git","stacks":["go"]}` {
		t.Fatalf("repo = %s", repo)
	}
	if strings.Contains(r.out, "user:") || strings.Contains(r.out, "tok@") || strings.Contains(r.out, s.dir) {
		t.Fatalf("export leaks a credential or a path: %s", r.out)
	}
	if p := b4Porcelain(t, s); p != "" {
		t.Fatalf("export wrote: %q", p)
	}
}

func TestWorkedExampleB4Workspace(t *testing.T) {
	shop := b4Fixture(t, b3Opts{})
	// The second repository has a brief of its own, from B3's builder.
	api, _, _ := b3PreMerge(t, b3Opts{})
	root := t.TempDir()
	for name, s := range map[string]*scratch{"shop": shop, "api": api} {
		if err := os.Rename(s.dir, filepath.Join(root, name)); err != nil {
			t.Skipf("cannot move a fixture: %v", err)
		}
		s.dir = filepath.Join(root, name)
	}
	ws := filepath.Join(root, "portfolio")
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ws, "workspace.toml"),
		[]byte("[[repo]]\npath = \"../shop\"\n[[repo]]\npath = \"../api\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	wsFile := filepath.Join(ws, "workspace.toml")

	own := func(s *scratch) []string {
		r := s.run(nil, "metrics")
		if r.code != 0 || r.err != "" {
			t.Fatalf("metrics in %s: %+v", s.dir, r)
		}
		return strings.Split(b3Collapse(r.out), "\n")
	}
	r := shop.run(nil, "metrics", "--workspace", wsFile)
	if r.code != 0 || r.err != "" {
		t.Fatalf("workspace: %+v", r)
	}
	got := strings.Split(b3Collapse(r.out), "\n")
	want := []string{"repo " + own(shop)[0]}
	for _, n := range []struct {
		name string
		s    *scratch
	}{{"shop", shop}, {"api", api}} {
		for _, l := range own(n.s)[1:] {
			want = append(want, n.name+" "+l)
		}
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("workspace table:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if len(got) != 3 || !strings.HasPrefix(got[1], "shop "+b3Brief+" ") {
		t.Fatalf("workspace table has the wrong rows: %v", got)
	}

	// A missing repository: the others still print, exit 1.
	if err := os.WriteFile(wsFile,
		[]byte("[[repo]]\npath = \"../shop\"\n[[repo]]\npath = \"../missing\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r = shop.run(nil, "metrics", "--workspace", wsFile)
	if r.code != 1 || r.err != "vloop: workspace repo not found: ../missing\n" ||
		!strings.Contains(b3Collapse(r.out), "shop "+b3Brief+" ") {
		t.Fatalf("missing repo: %+v", r)
	}
	r = shop.run(nil, "metrics", "export", "--workspace", wsFile)
	if r.code != 1 || r.err != "vloop: workspace repo not found: ../missing\n" || !strings.Contains(r.out, `"type":"brief"`) {
		t.Fatalf("missing repo, export: %+v", r)
	}
}

func TestWorkedExampleB4PlantedFailures(t *testing.T) {
	t.Run("HEAD on main", func(t *testing.T) {
		s := b4Fixture(t, b3Opts{})
		b3Git(t, s.dir, "2026-01-01T09:15:00Z", "checkout", "-q", "main")
		b3Git(t, s.dir, "2026-01-01T09:15:00Z", "merge", "-q", "--ff-only", b3Brief)
		before := b4Refs(t, s)
		b4Err(t, s.run(nil, "brief", "close", b4Name, "--no-findings"), 1, "vloop: close on the brief's work branch, not main\n")
		if b4Refs(t, s) != before || b4Porcelain(t, s) != "" {
			t.Fatal("a refusal wrote something")
		}
	})
	t.Run("untracked file", func(t *testing.T) {
		s := b4Fixture(t, b3Opts{})
		s.write("scratch.txt", "x\n")
		before := b4Refs(t, s)
		b4Err(t, s.run(nil, "brief", "close", b4Name, "--no-findings"), 1,
			"vloop: commit or stash your changes first — close makes one commit of its own\n")
		if b4Refs(t, s) != before || b4Porcelain(t, s) != "?? scratch.txt" {
			t.Fatal("a refusal wrote something")
		}
	})
	t.Run("stalled run", func(t *testing.T) {
		s := b4Fixture(t, b3Opts{stalled: true})
		b4Err(t, s.run(nil, "brief", "close", b4Name, "--no-findings"), 1,
			"vloop: the plan is not complete (1/2 done) — finish it, or pass --abandon \"<reason>\"\n")
		s.write("docs/briefs/B20260101-1000-b.loop-brief.md",
			"---\nname: B20260101-1000-b.loop-brief\nstatus: ready\ndepends-on: "+b4Name+"\n---\n# B\n")
		b3Git(t, s.dir, "2026-01-01T09:16:00Z", "add", "-A")
		b3Git(t, s.dir, "2026-01-01T09:16:00Z", "commit", "-q", "-m", "add dependent")

		r := s.run(nil, "brief", "close", b4Name, "--abandon", "superseded", "--no-findings")
		if r.code != 0 || r.err != "" {
			t.Fatalf("abandon: %+v", r)
		}
		if !hasLine(b3Collapse(r.out), b3Brief+" abandoned · not merged") {
			t.Fatalf("abandon summary:\n%s", r.out)
		}
		text := s.read(b4File)
		if !strings.Contains(text, "status: abandoned\n") || !strings.Contains(text, "**Status:** abandoned — superseded\n") {
			t.Fatalf("brief after abandon:\n%s", text)
		}
		r = s.run(nil, "brief", "check", b4File)
		if r.code != 0 || !strings.Contains(r.out, "  - skipped: status is abandoned\n") {
			t.Fatalf("brief check: %+v", r)
		}
		r = s.run(nil, "brief", "list")
		var dep string
		for _, l := range strings.Split(r.out, "\n") {
			if strings.HasPrefix(l, "B20260101-1000-b") {
				dep = l
			}
		}
		if r.code != 0 || !strings.Contains(dep, "blocked") {
			t.Fatalf("brief list should show the dependent as blocked:\n%s", r.out)
		}
	})
	t.Run("hand-written paragraph after the markers", func(t *testing.T) {
		s := b4Fixture(t, b3Opts{})
		if r := s.run(nil, "brief", "close", b4Name, "--no-findings"); r.code != 0 {
			t.Fatalf("close: %+v", r)
		}
		para := "\nA hand-written paragraph, kept byte for byte.\n\n- with a list\n"
		s.write(b4File, s.read(b4File)+para)
		b3Git(t, s.dir, "2026-01-01T09:40:00Z", "commit", "-q", "-am", "add a paragraph")
		// A re-render: put the brief back to ready, so close will render again.
		text := strings.Replace(s.read(b4File), "status: consumed", "status: ready", 1)
		s.write(b4File, text)
		b3Git(t, s.dir, "2026-01-01T09:41:00Z", "commit", "-q", "-am", "reopen")
		if r := s.run(nil, "brief", "close", b4Name, "--no-findings"); r.code != 0 {
			t.Fatalf("re-render: %+v", r)
		}
		if got := s.read(b4File); !strings.HasSuffix(got, "<!-- vloop:run-record:end -->\n"+para) {
			t.Fatalf("the paragraph after the markers changed:\n%s", got)
		}
	})
}
