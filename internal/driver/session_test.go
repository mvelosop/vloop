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
		"--settings .vloop/tmp/fence/1.2.3/review.json", "--plugin-dir .vloop/tmp/plugin/1.2.3 "} {
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
	if _, err := os.Stat(filepath.Join(r.Root, ".vloop/tmp/fence/1.2.3/review.json")); err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(filepath.Join(r.Root, ".vloop/tmp/plugin/1.2.3/.claude-plugin/plugin.json")); err != nil {
		t.Error(err)
	}
}

func TestSessionMasking(t *testing.T) {
	_, home, log := fixture(t, "")
	body := `{"total_cost_usd":0,"duration_ms":1,"num_turns":1,"is_error":false,"permission_denials":[{"tool_name":"Read","tool_input":{"file_path":"` + home + `/f"}}],"modelUsage":{"alice-model":{"inputTokens":0,"outputTokens":0,"cacheReadInputTokens":0,"cacheCreationInputTokens":0,"costUSD":0}}}` + "\n"
	r2, _, _ := fixture(t, body)
	r2.Home, r2.User, r2.Log = home, "alice", log
	res, err := r2.Run(Spec{Phase: PhaseWork, Iteration: 1, Arg: "T1", Model: "m"})
	if err != nil || !res.Recorded {
		t.Fatal(res, err)
	}
	b, _ := os.ReadFile(res.Path)
	if strings.Contains(string(b), home) {
		t.Errorf("record not masked: %s", b)
	}
	if !strings.Contains(string(b), `"~/f"`) || !strings.Contains(string(b), "alice-model") {
		t.Errorf("record: %s", b)
	}
	if got := r2.Mask("in " + home + "/a, user alice, /home/alice/b"); got != "in ~/a, user alice, /home/USER/b" {
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

func TestMaskUserOnlyAsPathComponent(t *testing.T) {
	for _, user := range []string{"us", "ion", "1000", "al"} {
		r := &Runner{Home: "/Users/" + user, User: user}
		for in, want := range map[string]string{
			"cost_usd iteration version 1000 al": "cost_usd iteration version 1000 al",
			"/Users/" + user + "x/f":             "/Users/" + user + "x/f",
			"/Users/x" + user + "/f":             "/Users/x" + user + "/f",
			"/home/" + user + "/f":               "/home/USER/f",
			`C:\Users\` + user + `\f`:            `C:\Users\USER\f`,
			"/Users/" + user + "/f":              "~/f",
			"/Users/other/" + user + "/f":        "/Users/other/" + user + "/f",
		} {
			if got := r.Mask(in); got != want {
				t.Errorf("user %s: Mask(%q) = %q, want %q", user, in, got, want)
			}
		}
	}
	// the user is also masked where the home differs
	r := &Runner{Home: "/elsewhere/me", User: "alice"}
	if got := r.Mask("/Users/alice/x /home/alice /home/alicia"); got != "/Users/USER/x /home/USER /home/alicia" {
		t.Errorf("Mask: %q", got)
	}
}

func TestMaskHomeAtPathBoundary(t *testing.T) {
	r := &Runner{Home: "/home/al", User: "al"}
	for in, want := range map[string]string{
		"/home/al":                 "~",
		"/home/al/x":               "~/x",
		`"/home/al"`:               `"~"`,
		`{"p":"/home/al\\x"}`:      `{"p":"~\\x"}`,
		"/home/alice/x":            "/home/alice/x",
		"/home/al-b/x /home/al.d":  "/home/al-b/x /home/al.d",
		"a /home/al, b /home/al/c": "a ~, b ~/c",
	} {
		if got := r.Mask(in); got != want {
			t.Errorf("Mask(%q) = %q, want %q", in, got, want)
		}
	}
	// a numeric home never corrupts a number
	n := &Runner{Home: "/home/1000", User: "1000"}
	if got := n.Mask("cost 1000 /home/10000/x"); got != "cost 1000 /home/10000/x" {
		t.Errorf("Mask: %q", got)
	}
}

func TestMaskRootHomeUnchanged(t *testing.T) {
	for _, home := range []string{"/", `\`, "//"} {
		r := &Runner{Home: home}
		in := `/usr/bin /tmp/x C:\a \\srv`
		if got := r.Mask(in); got != in {
			t.Errorf("home %q: Mask(%q) = %q", home, in, got)
		}
	}
}

func TestMaskWindowsForms(t *testing.T) {
	r := &Runner{Home: `C:\Users\Alice`, User: "Alice"}
	for in, want := range map[string]string{
		`C:\Users\Alice\f`:      `~\f`,
		`C:/Users/Alice/f`:      `~/f`,
		`c:\users\alice\f`:      `~\f`,
		`C:/users/ALICE`:        `~`,
		`"C:\\Users\\Alice\\f"`: `"~\\f"`,
		`C:\Users\Alicia\f`:     `C:\Users\Alicia\f`,
		`D:\Users\Alice\f`:      `D:\Users\USER\f`,
	} {
		if got := r.Mask(in); got != want {
			t.Errorf("Mask(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRedactSecretEnvValues(t *testing.T) {
	r := &Runner{Home: "/h/alice", User: "alice", Env: []string{
		"FOO_TOKEN=s3cr3t-value", "my_Password=hunter2hunter2", "API_KEY=abc123", "PLAIN=innocent-value",
		"AWS_CREDENTIAL=line\"quoted-secret",
	}}
	got := r.Mask("a s3cr3t-value b hunter2hunter2 c abc123 d innocent-value e line\"quoted-secret")
	want := "a <redacted:FOO_TOKEN> b <redacted:my_Password> c abc123 d innocent-value e <redacted:AWS_CREDENTIAL>"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
	rec, err := r.maskRecord(map[string]any{"x": `say "line\"quoted-secret"`, "y": []any{"s3cr3t-value"}})
	if err != nil || strings.Contains(string(rec), "s3cr3t-value") || strings.Contains(string(rec), "quoted-secret") {
		t.Errorf("record %s %v", rec, err)
	}
}

func TestSessionDenialsKeepToolAndPath(t *testing.T) {
	body := `{"type":"result","total_cost_usd":0.1,"duration_ms":1,"num_turns":1,"is_error":false,
"permission_denials":[{"tool_name":"Write","tool_use_id":"w","tool_input":{"file_path":"notes/a.txt","content":"TOPSECRET"}},
{"tool_name":"Bash","tool_use_id":"b","tool_input":{"command":"echo CMD-SECRET"}}],"modelUsage":{}}
`
	r, _, _ := fixture(t, body)
	res, err := r.Run(Spec{Phase: PhaseWork, Iteration: 1, Arg: "T1", Model: "sonnet"})
	if err != nil || !res.Recorded || res.Denials != 2 {
		t.Fatalf("%+v %v", res, err)
	}
	raw, _ := os.ReadFile(res.Path)
	if strings.Contains(string(raw), "TOPSECRET") || strings.Contains(string(raw), "CMD-SECRET") {
		t.Errorf("record keeps content or command: %s", raw)
	}
	want := []any{map[string]any{"tool_name": "Write", "file_path": "notes/a.txt"}, map[string]any{"tool_name": "Bash"}}
	if got := readJSON(t, res.Path)["permission_denials"]; !reflect.DeepEqual(got, want) {
		t.Errorf("denials %v, want %v", got, want)
	}
}
