package driver

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/mvelosop/vloop/internal/state"
)

func TestSessionArgs(t *testing.T) {
	s := Spec{Phase: PhaseWork, Iteration: 2, Arg: "T3", Model: "sonnet"}
	want := []string{"-p", "/vloop:work T3", "--model", "sonnet", "--permission-mode", "auto",
		"--setting-sources", "project", "--settings", "f/settings.json", "--plugin-dir", "p",
		"--strict-mcp-config", "--output-format", "json"}
	if got := Args(s, "f/settings.json", "p"); !reflect.DeepEqual(got, want) {
		t.Errorf("without effort:\n got %q\nwant %q", got, want)
	}
	s.Effort = "high"
	got := Args(s, "f/settings.json", "p")
	if got[4] != "--effort" || got[5] != "high" || len(got) != len(want)+2 {
		t.Errorf("with effort: %q", got)
	}
	if p := Args(Spec{Phase: PhasePlan, Arg: "docs/briefs/x.md", Model: "m"}, "f", "p")[1]; p != "/vloop:plan docs/briefs/x.md" {
		t.Errorf("plan prompt %q", p)
	}
	if p := Prompt(PhaseReview, "T1"); p != "/vloop:review T1" {
		t.Errorf("review prompt %q", p)
	}
}

func clearEnv(t *testing.T) {
	for _, k := range []string{"VLOOP_MODEL_PLAN", "VLOOP_MODEL_WORK", "VLOOP_MODEL_REVIEW",
		"VLOOP_EFFORT_PLAN", "VLOOP_EFFORT_WORK", "VLOOP_EFFORT_REVIEW"} {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}
}

func TestSessionReviewModelResolution(t *testing.T) {
	clearEnv(t)
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, ".git"), 0o755)
	os.MkdirAll(filepath.Join(root, ".vloop"), 0o755)
	os.WriteFile(filepath.Join(root, ".vloop", "config.toml"), []byte("[model]\nwork = \"wmodel\"\nreview = \"rmodel\"\n[effort]\nreview = \"low\"\n"), 0o644)

	m, e, err := ModelEffort(root, &state.Task{}, PhaseReview)
	if err != nil || m != "rmodel" || e != "low" {
		t.Errorf("review: %q %q %v", m, e, err)
	}
	m, e, _ = ModelEffort(root, &state.Task{}, PhaseWork)
	if m != "wmodel" || e != "" {
		t.Errorf("work: %q %q", m, e)
	}
	task := &state.Task{Model: &state.Sessions{Work: "twork"}, Effort: &state.Sessions{Review: "max"}}
	if m, _, _ = ModelEffort(root, task, PhaseReview); m != "rmodel" {
		t.Errorf("a work override leaked into the review: %q", m)
	}
	if m, _, _ = ModelEffort(root, task, PhaseWork); m != "twork" {
		t.Errorf("work override: %q", m)
	}
	if _, e, _ = ModelEffort(root, task, PhaseReview); e != "max" {
		t.Errorf("review effort override: %q", e)
	}
	t.Setenv("VLOOP_MODEL_REVIEW", "envmodel")
	if m, _, _ = ModelEffort(root, &state.Task{}, PhaseReview); m != "envmodel" {
		t.Errorf("env: %q", m)
	}
	t.Setenv("VLOOP_MODEL_REVIEW", "")
	os.Unsetenv("VLOOP_MODEL_REVIEW")
	if m, _, _ = ModelEffort(root, nil, PhasePlan); m != "opus" {
		t.Errorf("plan default: %q", m)
	}
}

// fixture builds a repo, a run folder and a stub claude that records its argv
// and environment and prints body on stdout.
func fixture(t *testing.T, body string) (*Runner, string, *bytes.Buffer) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("stub claude is a POSIX script")
	}
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, ".git"), 0o755)
	home := filepath.Join(t.TempDir(), "alice")
	os.Mkdir(home, 0o755)
	stub := filepath.Join(t.TempDir(), "claude")
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"" + root + "/argv\"\nenv >> \"" + root + "/env\"\ncat <<'EOF'\n" + body + "EOF\n"
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	var log bytes.Buffer
	run := filepath.Join(root, "runs", "r")
	return &Runner{Root: root, Version: "1.2.3", RunID: "B1-x", RunDir: run, Log: &log,
		Claude: stub, Home: home, User: "alice",
		Now: func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) },
		Env: []string{"PATH=" + os.Getenv("PATH"), "VLOOP_ACTIVE_TASK=T9", "VLOOP_GATE_TASK=T9", "KEEP=1"},
	}, home, &log
}

func readJSON(t *testing.T, p string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("%s: %v", p, err)
	}
	return m
}

const claudeOut = `{"type":"result","total_cost_usd":0.5,"duration_ms":1234,"num_turns":7,"is_error":false,
"permission_denials":[{"tool_name":"Bash"}],"result":"x",
"modelUsage":{"claude-a":{"inputTokens":1,"outputTokens":2,"cacheReadInputTokens":3,"cacheCreationInputTokens":4,"costUSD":0.25,"webSearchRequests":0},
"claude-b":{"inputTokens":10,"outputTokens":20,"cacheReadInputTokens":30,"cacheCreationInputTokens":40,"costUSD":0.25}}}
`

func TestSessionFieldMapping(t *testing.T) {
	r, _, _ := fixture(t, claudeOut)
	res, err := r.Run(Spec{Phase: PhaseWork, Iteration: 3, Arg: "T2", Model: "sonnet", Effort: "high"})
	if err != nil || !res.Recorded {
		t.Fatalf("%+v %v", res, err)
	}
	if filepath.Base(res.Path) != "001-work.json" {
		t.Errorf("record name %s", res.Path)
	}
	got := readJSON(t, res.Path)
	want := map[string]any{
		"schema": "session/v1", "run_id": "B1-x", "iteration": 3.0, "phase": "work", "task": "T2",
		"model": "sonnet", "effort": "high", "started": "2026-01-02T03:04:05Z",
		"duration_ms": 1234.0, "cost_usd": 0.5, "turns": 7.0, "is_error": false,
		"permission_denials": []any{map[string]any{"tool_name": "Bash"}},
		"models_used": map[string]any{
			"claude-a": map[string]any{"input_tokens": 1.0, "output_tokens": 2.0, "cache_read_input_tokens": 3.0, "cache_creation_input_tokens": 4.0, "cost_usd": 0.25},
			"claude-b": map[string]any{"input_tokens": 10.0, "output_tokens": 20.0, "cache_read_input_tokens": 30.0, "cache_creation_input_tokens": 40.0, "cost_usd": 0.25},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("record\n got %v\nwant %v", got, want)
	}

	// A plan has no task, an unset effort is null, and the counter advances.
	res, err = r.Run(Spec{Phase: PhasePlan, Iteration: 0, Arg: "docs/briefs/b.md", Model: "opus"})
	if err != nil || filepath.Base(res.Path) != "002-plan.json" {
		t.Fatalf("%+v %v", res, err)
	}
	got = readJSON(t, res.Path)
	if _, has := got["task"]; has {
		t.Error("a plan record carries a task")
	}
	if v, has := got["effort"]; !has || v != nil {
		t.Errorf("effort %v, want null", v)
	}
}

func TestSessionInvocationAndEnv(t *testing.T) {
	r, _, _ := fixture(t, claudeOut)
	if _, err := r.Run(Spec{Phase: PhaseReview, Iteration: 1, Arg: "T1", Model: "sonnet"}); err != nil {
		t.Fatal(err)
	}
	argv, _ := os.ReadFile(filepath.Join(r.Root, "argv"))
	for _, w := range []string{"-p /vloop:review T1 --model sonnet --permission-mode auto",
		"--settings .vloop/tmp/fence/1.2.3/settings.json", "--plugin-dir .vloop/tmp/plugin/1.2.3 "} {
		if !strings.Contains(string(argv), w) {
			t.Errorf("argv %q lacks %q", argv, w)
		}
	}
	if strings.Contains(string(argv), "--effort") {
		t.Errorf("--effort passed while unset: %s", argv)
	}
	env, _ := os.ReadFile(filepath.Join(r.Root, "env"))
	if strings.Contains(string(env), "VLOOP_ACTIVE_TASK") || strings.Contains(string(env), "VLOOP_GATE_TASK") || !strings.Contains(string(env), "KEEP=1") {
		t.Errorf("environment: %s", env)
	}
	if _, err := os.Stat(filepath.Join(r.Root, ".vloop/tmp/fence/1.2.3/settings.json")); err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(filepath.Join(r.Root, ".vloop/tmp/plugin/1.2.3/.claude-plugin/plugin.json")); err != nil {
		t.Error(err)
	}
}

func TestSessionMasking(t *testing.T) {
	_, home, log := fixture(t, "")
	body := `{"total_cost_usd":0,"duration_ms":1,"num_turns":1,"is_error":false,"permission_denials":[{"cmd":"cat ` + home + `/f"}],"modelUsage":{"alice-model":{"inputTokens":0,"outputTokens":0,"cacheReadInputTokens":0,"cacheCreationInputTokens":0,"costUSD":0}}}` + "\n"
	r2, _, _ := fixture(t, body)
	r2.Home, r2.User, r2.Log = home, "alice", log
	res, err := r2.Run(Spec{Phase: PhaseWork, Iteration: 1, Arg: "T1", Model: "m"})
	if err != nil || !res.Recorded {
		t.Fatal(res, err)
	}
	b, _ := os.ReadFile(res.Path)
	if strings.Contains(string(b), home) || strings.Contains(string(b), "alice") {
		t.Errorf("record not masked: %s", b)
	}
	if !strings.Contains(string(b), "cat ~/f") || !strings.Contains(string(b), "USER-model") {
		t.Errorf("record: %s", b)
	}
	if got := r2.Mask("in " + home + "/a, user alice"); got != "in ~/a, user USER" {
		t.Errorf("Mask: %q", got)
	}
	r2.Logf("saw %s", home)
	if strings.Contains(log.String(), home) || !strings.Contains(log.String(), "saw ~") {
		t.Errorf("log: %q", log.String())
	}
}

func TestSessionRecordMissing(t *testing.T) {
	r, _, log := fixture(t, "")
	res, err := r.Run(Spec{Phase: PhaseWork, Iteration: 4, Arg: "T7", Model: "m"})
	if err != nil || res.Recorded {
		t.Fatalf("%+v %v", res, err)
	}
	ents, _ := os.ReadDir(filepath.Join(r.RunDir, "sessions"))
	if len(ents) != 0 {
		t.Errorf("a record was written: %v", ents)
	}
	if want := "SESSION RECORD MISSING work T7 (iteration 4)\n"; log.String() != want {
		t.Errorf("log %q, want %q", log.String(), want)
	}
	log.Reset()
	if _, err := r.Run(Spec{Phase: PhasePlan, Iteration: 0, Arg: "b.md", Model: "m"}); err != nil {
		t.Fatal(err)
	}
	if want := "SESSION RECORD MISSING plan - (iteration 0)\n"; log.String() != want {
		t.Errorf("log %q, want %q", log.String(), want)
	}
}

func TestSessionRefusesSymlink(t *testing.T) {
	r, _, _ := fixture(t, claudeOut)
	other := t.TempDir()
	os.MkdirAll(filepath.Join(r.Root, ".vloop", "tmp"), 0o755)
	if err := os.Symlink(other, filepath.Join(r.Root, ".vloop", "tmp", "fence")); err != nil {
		t.Skip(err)
	}
	if _, err := r.Run(Spec{Phase: PhaseWork, Iteration: 1, Arg: "T1", Model: "m"}); err == nil {
		t.Error("ran through a symlinked fence folder")
	}
	if ents, _ := os.ReadDir(other); len(ents) != 0 {
		t.Errorf("wrote through the symlink: %v", ents)
	}
}
