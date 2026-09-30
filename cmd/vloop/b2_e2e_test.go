package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

const b2Plan = `{"schema":"state/v1","run_id":"B20260101-0900-a","brief":"docs/briefs/B20260101-0900-a.loop-brief.md",` +
	`"base":"0123456789abcdef0123456789abcdef01234567","branch":"B20260101-0900-a","status":"running","iteration":2,` +
	`"created":"2026-01-01T09:00:00Z","updated":"2026-01-01T09:00:00Z","shell":"sh","tasks":[` +
	`{"id":"T1","title":"Skeleton","goal":"g","kind":"feature","area":"cli","files":[],"references":[],"depends_on":[],"acceptance":["a"],"verify":"test -f skeleton.txt","status":"done","attempts":0,"notes":""},` +
	`{"id":"T2","title":"Config","goal":"g","kind":"feature","area":"config","files":[],"references":[],"depends_on":["T1"],"acceptance":["a"],"verify":"test -f config.txt","status":"pending","attempts":1,"notes":"","model":{"work":"opus"}},` +
	`{"id":"T3","title":"README","goal":"g","kind":"docs","area":"docs","files":[],"references":[],"depends_on":["T2"],"acceptance":["a"],"verify":"test -f README.md","status":"pending","attempts":0,"notes":""}]}`

const planPath = ".vloop/state/state.json"

func b2Scratch(t *testing.T) *scratch {
	t.Helper()
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}
	s := newScratch(t)
	s.write(".vloop/config.toml", "areas = [\"cli\", \"config\", \"docs\"]\n\n[effort]\nreview = \"high\"\n")
	s.write(planPath, b2Plan)
	return s
}

// plant rewrites the fixture plan through mutate.
func (s *scratch) plant(mutate func(m map[string]any)) {
	s.t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(b2Plan), &m); err != nil {
		s.t.Fatal(err)
	}
	mutate(m)
	b, err := json.Marshal(m)
	if err != nil {
		s.t.Fatal(err)
	}
	s.write(planPath, string(b))
}

func b2Task(m map[string]any, i int) map[string]any {
	return m["tasks"].([]any)[i].(map[string]any)
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	return lines[len(lines)-1]
}

func hasLine(out, line string) bool {
	for _, l := range strings.Split(out, "\n") {
		if l == line {
			return true
		}
	}
	return false
}

func TestWorkedExampleB2Commands(t *testing.T) {
	s := b2Scratch(t)

	expect(t, s.run(nil, "status"), 0,
		"B20260101-0900-a — running · 1/3 done · 0 blocked · iteration 2\n"+
			"  T1  done     cli/feature     Skeleton\n"+
			"  T2  pending  config/feature  Config  (1 attempt)\n"+
			"  T3  pending  docs/docs       README\n", "")

	r := s.run(nil, "task", "show", "T2")
	if r.code != 0 || r.err != "" {
		t.Fatalf("task show: %+v", r)
	}
	for _, l := range []string{
		"model   work opus (task) · review sonnet (default)",
		"effort  work - (default) · review high (file)",
	} {
		if !hasLine(r.out, l) {
			t.Errorf("task show T2: no line %q in %q", l, r.out)
		}
	}

	expect(t, s.run(nil, "task", "validate"), 0, "plan ok\n", "")

	r = s.run(nil, "task", "gate", "T2")
	if r.code != 1 || r.err != "" || !regexp.MustCompile(`^gate T2: fail \(exit 1, [^)]+\)$`).MatchString(lastLine(r.out)) {
		t.Fatalf("gate fail: %+v", r)
	}
	s.write("config.txt", "")
	r = s.run(nil, "task", "gate", "T2")
	if r.code != 0 || r.err != "" || !regexp.MustCompile(`^gate T2: pass \([^)]+\)$`).MatchString(lastLine(r.out)) {
		t.Fatalf("gate pass: %+v", r)
	}

	expect(t, s.run(nil, "task", "verify", "T2", "test -f cfg.toml", "--reason", "the file is cfg.toml"), 0, "", "")
	r = s.run(nil, "task", "show", "T2", "--json")
	if r.code != 0 || r.err != "" {
		t.Fatalf("task show --json: %+v", r)
	}
	var shown struct {
		GateHistory []struct {
			Verify, Reason, By string
		} `json:"gate_history"`
	}
	if err := json.Unmarshal([]byte(r.out), &shown); err != nil || len(shown.GateHistory) == 0 {
		t.Fatalf("gate_history: %v %q", err, r.out)
	}
	if g := shown.GateHistory[0]; g.Verify != "test -f config.txt" || g.Reason != "the file is cfg.toml" || g.By != "operator" {
		t.Fatalf("gate_history[0] = %+v", g)
	}

	expect(t, s.run(nil, "task", "verify", "T2", "test -f other.toml"), 2, "", "vloop: required flag(s) \"reason\" not set\n")
	expect(t, s.run(nil, "task", "drop", "T2"), 1, "", "vloop: T3 depends on T2\n")
	expect(t, s.run(nil, "task", "set", "T3", "area", "ops"), 2, "",
		"vloop: invalid value \"ops\" for area: want one of cli, config, docs\n")

	expect(t, s.run(nil, "task", "reset", "T2"), 0, "", "")
	r = s.run(nil, "task", "list")
	if r.code != 0 || !hasLine(r.out, "  T2  pending  config/feature  Config") {
		t.Fatalf("task list after reset: %+v", r)
	}

	expect(t, s.run(nil, "schema", "list"), 0, "defect/v1\nexport/v1\niteration/v1\nmetrics/v1\nproposal/v1\nsession/v1\nstate/v1\nverdict/v1\n", "")
	expect(t, s.run(nil, "schema", "validate", "state/v1", planPath), 0, planPath+": ok\n", "")
}

func TestWorkedExampleB2PlantedFailures(t *testing.T) {
	problem := func(t *testing.T, s *scratch, re string) {
		t.Helper()
		r := s.run(nil, "task", "validate")
		if r.code != 1 || r.err != "" {
			t.Fatalf("task validate: %+v", r)
		}
		if !regexp.MustCompile(`(?m)^ *✗ ` + re + `$`).MatchString(r.out) {
			t.Fatalf("no problem line /%s/ in %q", re, r.out)
		}
	}

	t.Run("duplicate id", func(t *testing.T) {
		s := b2Scratch(t)
		s.plant(func(m map[string]any) { b2Task(m, 2)["id"] = "T2" })
		problem(t, s, `duplicate task id: T2`)
	})
	t.Run("cycle", func(t *testing.T) {
		s := b2Scratch(t)
		s.plant(func(m map[string]any) { b2Task(m, 1)["depends_on"] = []any{"T3"} })
		problem(t, s, `depends_on cycle: T2 -> T3 -> T2`)
	})
	t.Run("area", func(t *testing.T) {
		s := b2Scratch(t)
		s.plant(func(m map[string]any) { b2Task(m, 2)["area"] = "ops" })
		problem(t, s, `T3: area "ops" is not in areas \(cli, config, docs\)`)
	})
	t.Run("schema violation and note refusal", func(t *testing.T) {
		s := b2Scratch(t)
		s.plant(func(m map[string]any) { b2Task(m, 1)["effort"] = map[string]any{"work": "turbo"} })
		problem(t, s, `schema: /tasks/1/effort/work: .+`)
		before := s.read(planPath)
		expect(t, s.run(nil, "task", "note", "T2", "x"), 1, "", "vloop: plan is not valid — run vloop task validate\n")
		if s.read(planPath) != before {
			t.Fatal("task note modified an invalid plan")
		}
	})
	t.Run("no plan", func(t *testing.T) {
		s := b2Scratch(t)
		if err := os.Remove(s.dir + "/" + planPath); err != nil {
			t.Fatal(err)
		}
		expect(t, s.run(nil, "task", "validate"), 1, "", "vloop: no plan: .vloop/state/state.json\n")
	})
	t.Run("shell not on PATH", func(t *testing.T) {
		s := b2Scratch(t)
		s.plant(func(m map[string]any) { m["shell"] = "pwsh" })
		r := s.run([]string{"PATH=" + t.TempDir(), "Path=" + t.TempDir()}, "task", "gate", "T2")
		expect(t, r, 1, "", "vloop: this plan's gates are pwsh commands and pwsh is not on PATH\n")
	})
}
