package driver

import (
	"encoding/json"
	"strings"
	"testing"
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
