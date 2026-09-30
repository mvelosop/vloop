package metrics

import (
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mvelosop/vloop/internal/runs"
)

const briefRel = "docs/briefs/B20260101-0900-a.loop-brief.md"

// fixture writes one shell-loop run folder under a temp repo.
type fixture struct {
	root string
	t    *testing.T
}

func newFixture(t *testing.T) *fixture { return &fixture{root: t.TempDir(), t: t} }

func (f *fixture) write(rel, body string) {
	f.t.Helper()
	p := filepath.Join(f.root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		f.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		f.t.Fatal(err)
	}
}

func (f *fixture) folder(name, log string, iters string, sessions map[string]string) {
	base := ".loop/state/runs/B20260101-0900-a/" + name + "/"
	f.write(base+"loop.log", log+"\n")
	if iters != "" {
		f.write(base+"iterations.jsonl", iters)
	}
	for n, s := range sessions {
		f.write(base+"sessions/"+n, s)
	}
}

func sess(phase string, it int, ms int64, cost float64, model string, in, out, cr, cc int64) string {
	return strings.NewReplacer("PHASE", phase, "ITER", itoa(it), "MS", itoa64(ms), "COST", ftoa(cost),
		"MODEL", model, "IN", itoa64(in), "OUT", itoa64(out), "CR", itoa64(cr), "CC", itoa64(cc)).Replace(
		`{"phase":"PHASE","iteration":ITER,"duration_ms":MS,"total_cost_usd":COST,"modelUsage":{"MODEL":{"inputTokens":IN,"outputTokens":OUT,"cacheReadInputTokens":CR,"cacheCreationInputTokens":CC,"costUSD":COST}}}`)
}

func itoa(i int) string     { return strconv.Itoa(i) }
func itoa64(i int64) string { return strconv.FormatInt(i, 10) }
func ftoa(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }

// workedRun plays the brief's worked example: plan in its own folder, T1 done,
// T2 gate_fail then done.
func workedRun(t *testing.T) *runs.Model {
	f := newFixture(t)
	f.folder("20260101-090000", "planning from "+briefRel, "", map[string]string{
		"001-plan.json": sess("plan", 0, 120000, 1.00, "claude-opus-5-5", 100, 50, 0, 0),
	})
	f.folder("20260101-090100", "resuming B20260101-0900-a",
		`{"iteration":1,"task":"T1","outcome":"done"}
{"iteration":2,"task":"T2","outcome":"gate_fail"}
{"iteration":3,"task":"T2","outcome":"done"}
`, map[string]string{
			"001-work.json":   sess("work", 1, 60000, 0.40, "claude-sonnet-5-5", 10, 20, 300, 100),
			"002-review.json": sess("review", 1, 20000, 0.10, "claude-sonnet-5-5", 5, 5, 100, 0),
			"003-work.json":   sess("work", 2, 40000, 0.30, "claude-sonnet-5-5", 10, 10, 0, 0),
			"004-work.json":   sess("work", 3, 30000, 0.20, "claude-sonnet-5-5", 10, 10, 0, 0),
			"005-review.json": sess("review", 3, 10000, 0.10, "claude-sonnet-5-5", 5, 5, 0, 0),
		})
	m, err := runs.Load(f.root, "B20260101-0900-a.loop-brief")
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Folders) != 2 {
		t.Fatalf("folders = %d, want 2", len(m.Folders))
	}
	base := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	m.Owned = &runs.Owned{
		Plan: runs.Commit{Time: base},
		Run:  &runs.Commit{Time: base.Add(10 * time.Minute)},
	}
	return m
}

func near(t *testing.T, name string, got *float64, want float64) {
	t.Helper()
	if got == nil || math.Abs(*got-want) > 1e-9 {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

func TestTimeCostWorkedExample(t *testing.T) {
	m := workedRun(t)
	u := Aggregate(m, Counts{Code: 15, Test: 6, Docs: 3})

	if u.Time.WorkMS != 130000 || u.Time.ReviewMS != 30000 || u.Time.AgentMS != 160000 {
		t.Errorf("work/review/agent = %d/%d/%d", u.Time.WorkMS, u.Time.ReviewMS, u.Time.AgentMS)
	}
	if u.Time.PlanMS != 120000 {
		t.Errorf("plan ms = %d, plan lives in another folder and is reported apart", u.Time.PlanMS)
	}
	if u.Time.GateMS != nil {
		t.Errorf("gate ms = %d, want unknown for the shell loop", *u.Time.GateMS)
	}
	if u.Time.WallMS == nil || *u.Time.WallMS != 600000 {
		t.Errorf("wall = %v, want 600000", u.Time.WallMS)
	}
	near(t, "code/min", u.Rate.CodePerMin, 15/(160000.0/60000))
	near(t, "code+test/min", u.Rate.CodeTestPerMin, 21/(160000.0/60000))

	near(t, "total", &u.Cost.Total, 2.10)
	near(t, "plan", &u.Cost.Plan, 1.00)
	near(t, "work", &u.Cost.Work, 0.90)
	near(t, "review", &u.Cost.Review, 0.20)
	near(t, "per 1k", u.Cost.Per1kCode, 140)

	if u.Tokens.Input != 140 || u.Tokens.Output != 100 || u.Tokens.CacheRead != 400 || u.Tokens.CacheCreation != 100 {
		t.Errorf("tokens = %+v", u.Tokens)
	}
	near(t, "cache hit", u.Tokens.CacheHit, 400.0/640)

	want := map[string]string{"plan": "claude-opus-5-5", "work": "claude-sonnet-5-5", "review": "claude-sonnet-5-5"}
	if !reflect.DeepEqual(u.Models, want) {
		t.Errorf("models = %v", u.Models)
	}
	if len(u.Effort) != 0 {
		t.Errorf("effort = %v, shell-loop records carry none", u.Effort)
	}
	if len(u.Missing) != 0 {
		t.Errorf("missing = %v", u.Missing)
	}
}

func TestTimeCostUnknownsAreNil(t *testing.T) {
	m := &runs.Model{}
	u := Aggregate(m, Counts{})
	if u.Rate.CodePerMin != nil || u.Cost.Per1kCode != nil || u.Tokens.CacheHit != nil || u.Time.WallMS != nil {
		t.Errorf("unknowns should be nil: %+v", u)
	}
}

func TestTimeCostSumsGateDurationsAndModelsAndEffort(t *testing.T) {
	g1, g2 := int64(1500), int64(2500)
	m := &runs.Model{Folders: []runs.Folder{{
		Iterations: []runs.Iteration{{Iteration: 1, Task: "T1", Outcome: "done", GateMS: &g1}, {Iteration: 2, Task: "T2", Outcome: "blocked", GateMS: &g2}},
		Sessions: []runs.Session{
			{Phase: "work", Iteration: 1, DurationMS: 1, CostUSD: 0.1, Effort: "high", Models: map[string]runs.Tokens{"b": {}, "a": {}}},
			{Phase: "work", Iteration: 2, DurationMS: 1, CostUSD: 0.2, Effort: "low", Models: map[string]runs.Tokens{"a": {}}},
			{Phase: "review", Iteration: 1, DurationMS: 1, CostUSD: 0.3, Model: "m"},
		},
	}}}
	u := Aggregate(m, Counts{})
	if u.Time.GateMS == nil || *u.Time.GateMS != 4000 {
		t.Errorf("gate = %v, want 4000", u.Time.GateMS)
	}
	if u.Models["work"] != "a,b" || u.Models["review"] != "m" {
		t.Errorf("models = %v", u.Models)
	}
	if u.Effort["work"] != "high,low" {
		t.Errorf("effort = %v", u.Effort)
	}
}

func TestTasksWorkedExample(t *testing.T) {
	m := workedRun(t)
	plan := &runs.PlanDoc{Tasks: []runs.PlanTask{{ID: "T1", Status: "done"}, {ID: "T2", Status: "done"}}}
	got := CountTasks(m, plan, "## Shape\n\n2 to 3 tasks\n")
	if got.Planned != 2 || got.Done != 2 || got.Blocked != 0 || got.FirstPass != 1 {
		t.Errorf("tasks = %+v", got)
	}
	if got.Estimate == nil || *got.Estimate != (Estimate{2, 3}) {
		t.Errorf("estimate = %v", got.Estimate)
	}
	near(t, "iterations per closed", got.IterationsPerClosed, 1.5)
}

func TestTasksBlockedThenDoneIsNotFirstPass(t *testing.T) {
	m := &runs.Model{Folders: []runs.Folder{{Iterations: []runs.Iteration{
		{Iteration: 1, Task: "T1", Outcome: "blocked"},
		{Iteration: 2, Task: "T1", Outcome: "done"},
		{Iteration: 3, Task: "T2", Outcome: "done"},
		{Iteration: 4, Task: "T3", Outcome: "blocked"},
	}}}}
	plan := &runs.PlanDoc{Tasks: []runs.PlanTask{{ID: "T1", Status: "done"}, {ID: "T2", Status: "done"}, {ID: "T3", Status: "blocked"}}}
	got := CountTasks(m, plan, "no estimate here")
	if got.Planned != 3 || got.Done != 2 || got.Blocked != 1 || got.FirstPass != 1 {
		t.Errorf("tasks = %+v", got)
	}
	if got.Estimate != nil {
		t.Errorf("estimate = %v, want none", got.Estimate)
	}
	near(t, "iterations per closed", got.IterationsPerClosed, 2)
}

func TestTasksNoPlanOrNoneDone(t *testing.T) {
	got := CountTasks(&runs.Model{}, nil, "")
	if got.Planned != 0 || got.IterationsPerClosed != nil {
		t.Errorf("tasks = %+v", got)
	}
}

func TestMissingRecordsReportedAndNotEstimated(t *testing.T) {
	f := newFixture(t)
	f.folder("20260101-090000", "planning from "+briefRel,
		`{"iteration":1,"task":"T1","outcome":"done"}
{"iteration":2,"task":"T2","outcome":"gate_fail"}
{"iteration":3,"task":"T2","outcome":"review_fail"}
{"iteration":4,"task":"T3","outcome":"blocked"}
{"iteration":5,"task":"T4","outcome":"rejected"}
`, map[string]string{
			"001-work.json":   sess("work", 1, 60000, 0.40, "claude-sonnet-5-5", 1, 1, 0, 0),
			"002-work.json":   sess("work", 2, 1000, 0.05, "claude-sonnet-5-5", 1, 1, 0, 0),
			"003-work.json":   sess("work", 3, 1000, 0.05, "claude-sonnet-5-5", 1, 1, 0, 0),
			"004-review.json": sess("review", 3, 1000, 0.05, "claude-sonnet-5-5", 1, 1, 0, 0),
			"005-work.json":   sess("work", 4, 1000, 0.05, "claude-sonnet-5-5", 1, 1, 0, 0),
			"006-review.json": sess("review", 5, 1000, 0.05, "claude-sonnet-5-5", 1, 1, 0, 0),
		})
	m, err := runs.Load(f.root, "B20260101-0900-a.loop-brief")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"T1 review (iteration 1)", "T4 work (iteration 5)"}
	if got := MissingRecords(m); !reflect.DeepEqual(got, want) {
		t.Errorf("missing = %v, want %v", got, want)
	}
	u := Aggregate(m, Counts{})
	if !reflect.DeepEqual(u.Missing, want) {
		t.Errorf("usage missing = %v", u.Missing)
	}
	// Only the records that exist are summed; nothing stands in for the rest.
	if u.Time.ReviewMS != 2000 || math.Abs(u.Cost.Review-0.10) > 1e-9 {
		t.Errorf("review = %d ms, $%v", u.Time.ReviewMS, u.Cost.Review)
	}
}

func TestMissingRecordsNoneWhenComplete(t *testing.T) {
	if got := MissingRecords(workedRun(t)); len(got) != 0 {
		t.Errorf("missing = %v", got)
	}
}

// The estimate is the Shape section's, not the first "N to M tasks" anywhere:
// B3's own worked example describes a fixture with "2 to 3 tasks" long before
// its Shape says "9 to 11", and the summary reported "brief said 2–3".
func TestBriefEstimateReadsOnlyTheShapeSection(t *testing.T) {
	cases := []struct {
		name, text string
		want       *Estimate
	}{
		{"shape after an earlier phrase", "## Worked example\n\nA fixture with Shape `2 to 3 tasks`.\n\n## Shape\n\n9 to 11 tasks, each verifiable.\n", &Estimate{9, 11}},
		{"spanish forma", "## Ejemplo trabajado\n\n2 a 3 tareas en el fixture.\n\n## Forma\n\n6 a 9 tareas, cada una verificable.\n", &Estimate{6, 9}},
		{"heading with trailing text", "## Shape — the order\n\n4 to 5 tasks.\n", &Estimate{4, 5}},
		{"phrase only outside shape", "## Worked example\n\n2 to 3 tasks\n\n## Shape\n\nA handful of tasks.\n", nil},
		{"no shape section", "8 to 10 tasks\n", nil},
		{"next section ends shape", "## Shape\n\nSmall.\n\n## Out of scope\n\n- 2 to 3 tasks of polish\n", nil},
	}
	for _, c := range cases {
		got := BriefEstimate(c.text)
		switch {
		case c.want == nil && got != nil:
			t.Errorf("%s: estimate = %v, want none", c.name, *got)
		case c.want != nil && (got == nil || *got != *c.want):
			t.Errorf("%s: estimate = %v, want %v", c.name, got, *c.want)
		}
	}
}
