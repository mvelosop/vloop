package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

var (
	buildOnce sync.Once
	vloopBin  string
	buildErr  string
)

// vloop builds the binary from this tree, once.
func vloop(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		dir, err := os.MkdirTemp("", "kit-vloop")
		if err != nil {
			buildErr = err.Error()
			return
		}
		vloopBin = filepath.Join(dir, "vloop")
		if runtime.GOOS == "windows" {
			vloopBin += ".exe"
		}
		cmd := exec.Command("go", "build", "-buildvcs=false", "-o", vloopBin, "../../../cmd/vloop")
		if out, err := cmd.CombinedOutput(); err != nil {
			buildErr = err.Error() + ": " + string(out)
		}
	})
	if buildErr != "" {
		t.Fatalf("go build ./cmd/vloop: %s", buildErr)
	}
	return vloopBin
}

func TestMain(m *testing.M) {
	code := m.Run()
	if vloopBin != "" {
		os.RemoveAll(filepath.Dir(vloopBin))
	}
	os.Exit(code)
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@example.com"}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
}

// fixtureRepo copies testdata/reports/<name> into a new git repository.
func fixtureRepo(t *testing.T, name string) string {
	t.Helper()
	repo := t.TempDir()
	src := filepath.Join("testdata", "reports", name)
	err := filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		if info.IsDir() {
			return os.MkdirAll(filepath.Join(repo, rel), 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(repo, rel), b, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	git(t, repo, "init", "-q", "-b", "main")
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-q", "-m", "fixtures")
	return repo
}

// metricsOut is `vloop metrics <id> --by session` in repo: its stdout and exit code.
func metricsOut(t *testing.T, repo, id string) (string, int) {
	t.Helper()
	cmd := exec.Command(vloop(t), "metrics", id, "--by", "session")
	cmd.Dir = repo
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if ee, ok := err.(*exec.ExitError); ok {
		return out.String(), ee.ExitCode()
	}
	if err != nil {
		t.Fatal(err)
	}
	return out.String(), 0
}

// block is one expected Sessions block: the id, the minimum rows its metrics
// call must print, and the lines after the fence, written literally.
type block struct {
	id    string
	rows  int
	after string
}

func (b block) text(t *testing.T, repo string) string {
	out, code := metricsOut(t, repo, b.id)
	if code != 0 {
		t.Fatalf("vloop metrics %s --by session exited %d", b.id, code)
	}
	if rows := len(strings.Split(strings.TrimSpace(out), "\n")) - 1; rows < b.rows {
		t.Fatalf("vloop metrics %s --by session printed %d rows, want at least %d:\n%s", b.id, rows, b.rows, out)
	}
	return "Sessions (`vloop metrics " + b.id + " --by session`):\n\n```\n" + out + "```\n\n" + b.after + "\n"
}

// replaceTable replaces the golden's first table (its header row through the
// blank line after it) with the Sessions text.
func replaceTable(t *testing.T, golden, sessions string) string {
	t.Helper()
	lines := strings.SplitAfter(golden, "\n")
	start := -1
	for i, l := range lines {
		if strings.HasPrefix(l, "| record |") || strings.HasPrefix(l, "| round | record |") {
			start = i
			break
		}
	}
	if start < 0 {
		t.Fatal("the golden has no session table")
	}
	end := start
	for end < len(lines) && lines[end] != "\n" {
		end++
	}
	return strings.Join(lines[:start], "") + sessions + strings.Join(lines[end+1:], "")
}

func runKit(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	var so, se bytes.Buffer
	code := run(args, &so, &se)
	return so.String(), se.String(), code
}

func golden(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "golden", "report", name+".md"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

const (
	r1ka = "B20260201-0000-ka"
	r1kb = "B20260203-0000-kb"
	r2ks = "B20260301-0000-ks"
)

const r1kaAfter = `Session records: 7, schema-valid 5 (session/v3 1, session/v4 5, no schema 1).
- .vloop/state/runs/B20260201-0000-ka/20260201-100000/sessions/005-work.json: session/v4 INVALID
- .vloop/state/runs/B20260201-0000-ka/20260201-100000/sessions/006-review.json: no schema

Permission denials:

- 20260201-100000/001 plan Bash: git push origin main
- 20260201-100000/004 review Write: src/é.go
- 20260201-100000/006 review Bash: rm -rf	x
- 20260202-100000/001 verify Write: a/b
- 20260202-100000/001 verify Bash: denied <&> é 😀
`

const r1kbAfter = "Session records: 2, schema-valid 2 (session/v4 2).\n"

const r2ksAfter = `Session records: 5, schema-valid 5 (session/v3 1, session/v4 4).

Permission denials:

- 20260301-100000/004 second-review shell: blocked by the fence
`

func TestReports(t *testing.T) {
	r1 := []block{{r1ka, 2, r1kaAfter}, {r1kb, 2, r1kbAfter}}
	cases := []struct {
		kind, fixture, code string
		blocks              []block
	}{
		{"run", "r1", "0", r1},
		{"work", "r1", "1", r1},
		{"verify", "r1", "0", r1},
		{"gate-review", "r1", "2", r1},
		{"plan", "r1", "0", r1},
		{"second", "r2", "0", []block{{r2ks, 2, r2ksAfter}}},
	}
	for _, c := range cases {
		t.Run(c.kind+"-"+c.fixture, func(t *testing.T) {
			repo := fixtureRepo(t, c.fixture)
			var sessions string
			for _, b := range c.blocks {
				sessions += b.text(t, repo)
			}
			want := replaceTable(t, golden(t, c.kind+"-"+c.fixture), sessions)
			got, serr, code := runKit(t, "report", c.kind, repo, vloop(t), c.code)
			if code != 0 || serr != "" {
				t.Fatalf("exit %d: %s", code, serr)
			}
			if got != want {
				t.Errorf("report %s on %s differs from the golden with the Sessions blocks:\n--- want\n%s\n--- got\n%s", c.kind, c.fixture, want, got)
			}
		})
	}
}

// vloop metrics exits 1 on r3: the fence holds its error, the rest follows.
func TestReportMetricsFails(t *testing.T) {
	repo := fixtureRepo(t, "r3")
	_, code := metricsOut(t, repo, "B20260401-0000-kbad")
	if code != 1 {
		t.Fatalf("vloop metrics on r3 exited %d, want 1", code)
	}
	sessions := "Sessions (`vloop metrics B20260401-0000-kbad --by session`):\n\n```\n" +
		"vloop metrics exited 1: vloop: iterations.jsonl line 2: invalid character 'o' in literal null (expecting 'u')\n" +
		"```\n\nSession records: 2, schema-valid 2 (session/v4 2).\n\nPermission denials:\n\n- 20260401-100000/002 work Bash: curl x\n\n"
	want := replaceTable(t, golden(t, "work-r3"), sessions)
	got, serr, rc := runKit(t, "report", "work", repo, vloop(t), "0")
	if rc != 0 || serr != "" {
		t.Fatalf("exit %d: %s", rc, serr)
	}
	if got != want {
		t.Errorf("--- want\n%s\n--- got\n%s", want, got)
	}
}

// No run folder has a brief: the table is replaced by one line.
func TestReportNoBriefWithRuns(t *testing.T) {
	repo := fixtureRepo(t, "r4")
	want := replaceTable(t, golden(t, "run-r4"), "Sessions: no brief with runs.\n\n")
	got, serr, rc := runKit(t, "report", "run", repo, vloop(t), "0")
	if rc != 0 || serr != "" {
		t.Fatalf("exit %d: %s", rc, serr)
	}
	if got != want {
		t.Errorf("--- want\n%s\n--- got\n%s", want, got)
	}
}

// A session file that is not JSON counts in the total and is listed.
func TestReportUnreadableRecord(t *testing.T) {
	repo := fixtureRepo(t, "r1")
	rel := ".vloop/state/runs/B20260203-0000-kb/20260203-100000/sessions/003-work.json"
	if err := os.WriteFile(filepath.Join(repo, rel), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, serr, rc := runKit(t, "report", "run", repo, vloop(t), "0")
	if rc != 0 || serr != "" {
		t.Fatalf("exit %d: %s", rc, serr)
	}
	for _, line := range []string{"Session records: 3, schema-valid 2 (session/v4 2).", "- " + rel + ": unreadable"} {
		if !strings.Contains("\n"+got, "\n"+line+"\n") {
			t.Errorf("no line %q in:\n%s", line, got)
		}
	}
}

func TestReportRefusals(t *testing.T) {
	for _, args := range [][]string{
		nil,
		{"bogus"},
		{"report"},
		{"report", "bogus", ".", "vloop", "0"},
		{"report", "run", "."},
		{"report", "run", ".", "vloop"},
	} {
		so, se, code := runKit(t, args...)
		if code != 1 || so != "" || se == "" || strings.Count(strings.TrimSuffix(se, "\n"), "\n") != 0 {
			t.Errorf("kit %v: exit %d, stdout %q, stderr %q; want exit 1, empty stdout, a one-line message", args, code, so, se)
		}
	}
}

func TestPyRendering(t *testing.T) {
	v, err := parseJSON(`{"b": 1.0, "a": [1e16, 1e-5, 0.1, -0, "\u00e9\ud83d\ude00", null, true], "c": {"x": "it's"}}`)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := dumps(v, false), `{"b": 1.0, "a": [1e+16, 1e-05, 0.1, 0, "\u00e9\ud83d\ude00", null, true], "c": {"x": "it's"}}`; got != want {
		t.Errorf("dumps = %s, want %s", got, want)
	}
	if got, want := dumps(v, true), `{"a": [1e+16, 1e-05, 0.1, 0, "\u00e9\ud83d\ude00", null, true], "b": 1.0, "c": {"x": "it's"}}`; got != want {
		t.Errorf("sorted dumps = %s, want %s", got, want)
	}
	if got, want := pyStr(v), `{'b': 1.0, 'a': [1e+16, 1e-05, 0.1, 0, 'é😀', None, True], 'c': {'x': "it's"}}`; got != want {
		t.Errorf("str = %s, want %s", got, want)
	}
	for _, bad := range []string{``, `{`, `[1,]`, `1 2`, `NaN`, "\"a\nb\"", `01`} {
		if _, err := parseJSON(bad); err == nil {
			t.Errorf("parseJSON(%q) succeeded", bad)
		}
	}
}
