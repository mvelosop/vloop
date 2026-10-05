package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

var binPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "vloop-e2e-")
	if err != nil {
		panic(err)
	}
	binPath = filepath.Join(dir, "vloop")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	if out, err := exec.Command("go", "build", "-o", binPath, ".").CombinedOutput(); err != nil {
		os.RemoveAll(dir)
		panic("build failed: " + err.Error() + "\n" + string(out))
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// scratch is a fresh git repository the binary runs in.
type scratch struct {
	t    *testing.T
	dir  string
	home string
}

func newScratch(t *testing.T) *scratch {
	t.Helper()
	s := &scratch{t: t, dir: t.TempDir(), home: t.TempDir()}
	if out, err := exec.Command("git", "init", "-q", s.dir).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	s.write("docs/x.md", "x\n")
	return s
}

func (s *scratch) write(rel, content string) {
	s.t.Helper()
	p := filepath.Join(s.dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		s.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		s.t.Fatal(err)
	}
}

func (s *scratch) read(rel string) string {
	s.t.Helper()
	b, err := os.ReadFile(filepath.Join(s.dir, filepath.FromSlash(rel)))
	if err != nil {
		s.t.Fatal(err)
	}
	return string(b)
}

type result struct {
	out, err string
	code     int
}

func (s *scratch) run(env []string, args ...string) result {
	s.t.Helper()
	cmd := exec.Command(binPath, args...)
	cmd.Dir = s.dir
	cmd.Env = append(cleanEnv(s.home), env...)
	var o, e bytes.Buffer
	cmd.Stdout, cmd.Stderr = &o, &e
	code := 0
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			s.t.Fatalf("run %v: %v", args, err)
		}
		code = ee.ExitCode()
	}
	return result{o.String(), e.String(), code}
}

// cleanEnv drops every vloop-relevant variable and points home at a temp dir.
func cleanEnv(home string) []string {
	var env []string
	for _, kv := range os.Environ() {
		k := strings.ToUpper(strings.SplitN(kv, "=", 2)[0])
		if strings.HasPrefix(k, "VLOOP_") || k == "NO_COLOR" || k == "HOME" || k == "USERPROFILE" ||
			k == "XDG_CONFIG_HOME" || k == "APPDATA" {
			continue
		}
		env = append(env, kv)
	}
	return append(env, "HOME="+home, "USERPROFILE="+home, "XDG_CONFIG_HOME="+filepath.Join(home, ".config"), "APPDATA="+home)
}

func expect(t *testing.T, r result, code int, out, errOut string) {
	t.Helper()
	if r.code != code || r.out != out || r.err != errOut {
		t.Fatalf("got code=%d out=%q err=%q\nwant code=%d out=%q err=%q", r.code, r.out, r.err, code, out, errOut)
	}
	if strings.Contains(r.out+r.err, "\x1b") {
		t.Fatalf("colour escapes in output: %q %q", r.out, r.err)
	}
}

func TestWorkedExampleEnglishSession(t *testing.T) {
	s := newScratch(t)

	r := s.run(nil, "version", "--json")
	if r.code != 0 || r.err != "" {
		t.Fatalf("version: %+v", r)
	}
	want := `{"version":"2.0.0-beta.2","commit":"unknown","plugin":"2.0.0-beta.2","go":"` + runtime.Version() +
		`","os":"` + runtime.GOOS + `","arch":"` + runtime.GOARCH + `"}` + "\n"
	if r.out != want {
		t.Fatalf("version --json = %q, want %q", r.out, want)
	}

	expect(t, s.run(nil, "config", "list"), 0,
		"language=en (default)\nmodel.plan=opus (default)\nmodel.work=sonnet (default)\nmodel.review=sonnet (default)\nmodel.gate-review=sonnet (default)\n"+
			"effort.plan= (default)\neffort.work= (default)\neffort.review= (default)\neffort.gate-review= (default)\n"+
			"shell="+defaultShell()+" (default)\nareas= (default)\n"+
			"metrics.stacks= (default)\nmetrics.code= (default)\nmetrics.test= (default)\nmetrics.docs= (default)\nmetrics.excluded= (default)\n"+
			"run.max-iterations=30 (default)\nrun.cost-ceiling=40 (default)\nrun.max-attempts=3 (default)\n"+
			"run.stall-limit=2 (default)\nrun.convergence-max=3.0 (default)\nrun.convergence-min=6 (default)\nrun.gate-timeout=15 (default)\nrun.gate-scratch= (default)\nrun.session-timeout=60 (default)\nrun.keep-awake=on (default)\n", "")
	expect(t, s.run(nil, "config", "set", "language", "es"), 0, "", "")
	expect(t, s.run(nil, "config", "set", "effort.review", "high"), 0, "", "")
	expect(t, s.run([]string{"VLOOP_MODEL_WORK=opus"}, "config", "get", "model.work", "--json"), 0,
		`{"key":"model.work","value":"opus","source":"env"}`+"\n", "")
	expect(t, s.run(nil, "config", "get", "language"), 0, "es\n", "")
	expect(t, s.run(nil, "config", "set", "language", "fr"), 2, "",
		"vloop: invalid value \"fr\" for language: want one of en, es\n")
	expect(t, s.run(nil, "config", "get", "colour"), 2, "", "vloop: unknown config key \"colour\"\n")

	r = s.run(nil, "brief", "new", "pagos-base")
	if r.code != 0 || r.err != "" {
		t.Fatalf("brief new: %+v", r)
	}
	path := strings.TrimSuffix(r.out, "\n")
	if !regexp.MustCompile(`^docs/briefs/B\d{8}-\d{4}-pagos-base\.loop-brief\.md$`).MatchString(path) {
		t.Fatalf("brief new printed %q", path)
	}
	expect(t, s.run(nil, "brief", "new", "pagos-base"), 1, "", "vloop: brief exists: "+path+"\n")
	expect(t, s.run(nil, "brief", "check", path), 0, path+"\n  - skipped: status is draft\nbriefs ok\n", "")
	expect(t, s.run(nil, "brief", "check", "docs/notes.md"), 2, "", "vloop: not a loop brief: docs/notes.md\n")

	// vloop wrote only under .vloop/ and docs/briefs/, and nothing in home.
	filepath.WalkDir(s.dir, func(p string, d os.DirEntry, err error) error {
		rel, _ := filepath.Rel(s.dir, p)
		rel = filepath.ToSlash(rel)
		if d.IsDir() && rel == ".git" {
			return filepath.SkipDir
		}
		if !d.IsDir() && rel != "docs/x.md" && !strings.HasPrefix(rel, ".vloop/") && !strings.HasPrefix(rel, "docs/briefs/") {
			t.Errorf("unexpected file %s", rel)
		}
		return nil
	})
	if ents, _ := os.ReadDir(s.home); len(ents) != 0 {
		t.Errorf("home was written: %v", ents)
	}
}

const (
	nameA = "B20260101-0900-a.loop-brief"
	nameB = "B20260101-0800-b.loop-brief"
)

func esBrief(name, status, deps string) string {
	return strings.Join([]string{
		"---", "name: " + name, "description: Construir a", "kind: brief", "status: " + status,
		"created: 2026-01-01", "seeds: Una ejecución que construye a", "depends-on: " + deps, "---",
		"# Brief", "", "## Qué es", "", "Una cosa.", "",
		"## Por qué esta forma, y qué se descartó", "", "Porque sí.", "",
		"## Referencias vinculantes", "", "- `docs/x.md` — el contrato de errores", "",
		"## Contrato de comportamiento", "", "Un id desconocido termina con código de salida 1.", "",
		"## Ejemplo trabajado", "", "```", "$ a", "ok", "```", "",
		"## Fuera de alcance", "", "- uno", "- dos", "",
		"## Restricciones", "", "- Go.", "",
		"## Forma", "", "6 a 9 tareas.", "",
	}, "\n")
}

func esFixture(t *testing.T) *scratch {
	s := newScratch(t)
	expect(t, s.run(nil, "config", "set", "language", "es"), 0, "", "")
	s.write("docs/briefs/"+nameA+".md", esBrief(nameA, "ready", "["+nameB+"]"))
	s.write("docs/briefs/"+nameB+".md", esBrief(nameB, "consumed", "[]"))
	return s
}

func TestWorkedExampleSpanishBrief(t *testing.T) {
	s := esFixture(t)
	pa := "docs/briefs/" + nameA + ".md"
	r := s.run(nil, "brief", "check", pa)
	if r.code != 0 || r.err != "" || !strings.Contains(r.out, "briefs ok") || !regexp.MustCompile(`(?m)^ *ok, 0 warning\(s\)$`).MatchString(r.out) {
		t.Fatalf("check: %+v", r)
	}
	expect(t, s.run(nil, "brief", "list"), 0,
		nameB+"  consumed  -  -\n"+nameA+"  ready  ready  "+nameB+"\n", "")
}

func TestWorkedExamplePlantedFailures(t *testing.T) {
	pa := "docs/briefs/" + nameA + ".md"
	pb := "docs/briefs/" + nameB + ".md"
	cases := []struct {
		name   string
		plant  func(s *scratch)
		problm string
	}{
		{"wrong-language heading", func(s *scratch) {
			s.write(pa, strings.Replace(s.read(pa), "## Ejemplo trabajado", "## Worked example", 1))
		}, "no worked example section — nothing arbitrates a disagreement"},
		{"missing binding ref", func(s *scratch) { os.Remove(filepath.Join(s.dir, "docs", "x.md")) },
			"binding reference does not resolve: docs/x.md"},
		{"binding ref without reason", func(s *scratch) {
			s.write(pa, strings.Replace(s.read(pa), " — el contrato de errores", "", 1))
		}, "binding reference has no reason: docs/x.md"},
		{"cycle", func(s *scratch) {
			s.write(pb, strings.Replace(s.read(pb), "depends-on: []", "depends-on: ["+nameA+"]", 1))
		}, "depends-on cycle: " + nameB + " -> " + nameA + " -> " + nameB},
		{"already run", func(s *scratch) {
			s.write(".vloop/state/journals/B20260101-0900-a.md", "")
		}, "already run — .vloop/state/journals/B20260101-0900-a.md exists"},
		{"name mismatch", func(s *scratch) {
			s.write(pa, strings.Replace(s.read(pa), "name: "+nameA, "name: B20260101-0901-a.loop-brief", 1))
		}, "name B20260101-0901-a.loop-brief does not match filename " + nameA},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := esFixture(t)
			c.plant(s)
			r := s.run(nil, "brief", "check", pa)
			line := regexp.MustCompile(`(?m)^ *✗ ` + regexp.QuoteMeta(c.problm) + `(; .*)?$`)
			if r.code != 1 || !line.MatchString(r.out) || !strings.Contains(r.out, "1 brief(s) need work") ||
				strings.Contains(r.out+r.err, "\x1b") {
				t.Fatalf("check: %+v", r)
			}
			if c.name == "cycle" {
				l := s.run(nil, "brief", "list")
				if l.code != 1 || l.out != "" || !strings.Contains(l.err, "depends-on cycle: ") {
					t.Fatalf("list: %+v", l)
				}
			}
		})
	}
}

func defaultShell() string {
	if runtime.GOOS == "windows" {
		return "pwsh"
	}
	return "sh"
}
