package driver

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestBypassedCalls(t *testing.T) {
	long := strings.Repeat("é", 250)
	lines := []string{
		`{"type":"tool.execution_start","data":{"toolCallId":"a","toolName":"bash","arguments":{"command":"nope"}}}`, // no completion: ran
		`{"type":"tool.execution_start","data":{"toolCallId":"b","toolName":"bash","arguments":{"command":"nope"}}}`,
		`{"type":"tool.execution_complete","data":{"toolCallId":"b","error":{"code":"denied"}}}`, // denied: did not run
		`{"type":"tool.execution_start","data":{"toolCallId":"c","toolName":"create","arguments":{"path":"x/y"}}}`,
		`{"type":"tool.execution_complete","data":{"toolCallId":"c","success":true}}`,
		`{"type":"tool.execution_start","data":{"toolCallId":"d","toolName":"bash","arguments":{"command":"` + long + `\nsecond"}}}`,
		`{"type":"tool.execution_complete","data":{"toolCallId":"d","success":true}}`,
		`{"type":"tool.execution_start","data":{"toolCallId":"e","toolName":"frob","arguments":{}}}`,
		`{"type":"tool.execution_start","data":{"toolCallId":"f","toolName":"bash","arguments":{"command":"ok"}}}`,
	}
	judge := func(phase, tool string, args json.RawMessage) CallVerdict {
		var a struct{ Command, Path string }
		_ = json.Unmarshal(args, &a)
		return CallVerdict{Allow: a.Command == "ok", Reason: phase + " denies", Path: a.Path, Command: a.Command}
	}
	got := bypassedCalls([]byte(strings.Join(lines, "\n")), "work", judge)
	want := []BypassedCall{
		{"bash", "nope", "work denies"},
		{"create", "x/y", "work denies"},
		{"bash", strings.Repeat("é", 200) + "…", "work denies"},
		{"frob", "", "work denies"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %+v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("call %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

// stubCopilot writes an executable `copilot` running body into a new directory
// that is put first on PATH, and returns a Runner with the denying judge.
func stubCopilot(t *testing.T, body string) *Runner {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the stub copilot is a POSIX shell script")
	}
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "copilot"), []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	judge := func(phase, tool string, args json.RawMessage) CallVerdict {
		var a struct{ Command string }
		_ = json.Unmarshal(args, &a)
		return CallVerdict{Allow: a.Command == "ls", Reason: "no " + phase, Command: a.Command}
	}
	return &Runner{Root: t.TempDir(), Version: "9.9.9", RunID: "r", RunDir: t.TempDir(), Judge: judge,
		Env: []string{"PATH=" + os.Getenv("PATH")}}
}

const startCommit = `printf '%s\n' '{"type":"tool.execution_start","data":{"toolCallId":"c1","toolName":"bash","arguments":{"command":"git commit -m x"}}}'
`

func warnInto(lines *[]string) func(string, ...any) {
	return func(f string, a ...any) { *lines = append(*lines, fmt.Sprintf(f, a...)) }
}

// F2: the binary the hook runs changes between the check's before and after.
func TestBackstopHookChanged(t *testing.T) {
	hook := filepath.Join(t.TempDir(), "vloop")
	if err := os.WriteFile(hook, []byte("one"), 0o755); err != nil {
		t.Fatal(err)
	}
	r := stubCopilot(t, "echo two >> '"+hook+"'\n")
	r.HookBinary = hook
	res, err := r.Run(Spec{Agent: AgentCopilot, Phase: PhaseWork, Arg: "T1", Model: "m", Iteration: 1})
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	h := res.backstop(PhaseWork, warnInto(&lines))
	want := "HOOK CHANGED work — the binary the fence hook runs changed during the session: " + r.Mask(hook) + "; nothing from it was used"
	if h == nil || h.Code != ExitRefsMoved || h.Error() != want {
		t.Errorf("halt = %v, want exit 9 %q", h, want)
	}

	// Unchanged, the same session passes; gone, it halts.
	r = stubCopilot(t, "")
	r.HookBinary = hook
	if res, err = r.Run(Spec{Agent: AgentCopilot, Phase: PhaseWork, Arg: "T1", Model: "m", Iteration: 1}); err != nil || res.backstop(PhaseWork, warnInto(&lines)) != nil {
		t.Errorf("an unchanged binary halted: %v", err)
	}
	r = stubCopilot(t, "rm '"+hook+"'\n")
	r.HookBinary = hook
	if res, err = r.Run(Spec{Agent: AgentCopilot, Phase: PhaseWork, Arg: "T1", Model: "m", Iteration: 1}); err != nil || !res.HookChanged {
		t.Errorf("a removed binary was not noticed: %v", err)
	}
}

// A session killed at its timeout is still re-judged; a call with no
// completion counts as run.
func TestBackstopTimedOutSession(t *testing.T) {
	hook := filepath.Join(t.TempDir(), "vloop")
	if err := os.WriteFile(hook, []byte("one"), 0o755); err != nil {
		t.Fatal(err)
	}
	r := stubCopilot(t, startCommit+"sleep 30\n")
	r.HookBinary, r.Timeout = hook, time.Second
	res, err := r.Run(Spec{Agent: AgentCopilot, Phase: PhaseWork, Arg: "T1", Model: "m", Iteration: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !res.TimedOut {
		t.Fatalf("the session was not killed: %+v", res)
	}
	var lines []string
	h := res.backstop(PhaseWork, warnInto(&lines))
	if h == nil || h.Code != ExitRefsMoved || !strings.HasPrefix(h.Error(), "FENCE BYPASSED work — the session ran 1 call(s)") {
		t.Errorf("halt = %v", h)
	}
	if len(lines) == 0 || lines[0] != "vloop: bash git commit -m x — no work" {
		t.Errorf("lines = %q", lines)
	}
}
