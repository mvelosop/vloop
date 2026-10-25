package driver

import (
	"strings"
	"testing"
)

const copilotLines = `{"type":"session.start","data":{"sessionId":"s"}}
{"type":"assistant.turn_start","data":{"turnId":"1"}}
{"type":"session.usage_checkpoint","data":{"totalNanoAiu":5100000000}}
{"type":"assistant.turn_start","data":{"turnId":"2"}}
{"type":"session.usage_checkpoint","data":{"totalNanoAiu":14200000000}}
{"type":"result","exitCode":0,"usage":{"sessionDurationMs":18000,"codeChanges":{"filesModified":["/abs/path/x.out"]}}}
`

func parseCopilot(t *testing.T, out string) (Output, bool) {
	t.Helper()
	return copilotAdapter{model: "m1", creditUSD: 0.01}.Parse([]byte(out), func(s string) string { return s })
}

// The last checkpoint's credits count, at the configured price; turns, the
// duration and the exit code come from the events, and codeChanges goes nowhere.
func TestCopilotParseRecord(t *testing.T) {
	o, ok := parseCopilot(t, copilotLines)
	if !ok {
		t.Fatal("no record")
	}
	if o.Credits == nil || *o.Credits < 14.1999 || *o.Credits > 14.2001 {
		t.Errorf("credits %v, want 14.2", o.Credits)
	}
	if want := 0.142; o.Cost < want-1e-9 || o.Cost > want+1e-9 || o.CostSource != CostEstimated {
		t.Errorf("cost %v (%s), want %v estimated", o.Cost, o.CostSource, want)
	}
	if o.Turns != 2 || o.AgentMS != 18000 || o.IsError || len(o.Denials) != 0 {
		t.Errorf("output %+v", o)
	}
	if u, ok := o.ModelsUsed["m1"]; !ok || len(o.ModelsUsed) != 1 || u.Input != 0 || u.Cost != o.Cost {
		t.Errorf("models_used %+v", o.ModelsUsed)
	}
}

func TestCopilotParseNoResultNoRecord(t *testing.T) {
	cut := copilotLines[:strings.Index(copilotLines, `{"type":"result"`)]
	if _, ok := parseCopilot(t, cut); ok {
		t.Error("a record without a result event")
	}
	if _, ok := parseCopilot(t, ""); ok {
		t.Error("a record from nothing")
	}
}

func TestCopilotParseNoCheckpoint(t *testing.T) {
	var keep []string
	for _, l := range strings.Split(copilotLines, "\n") {
		if !strings.Contains(l, "usage_checkpoint") {
			keep = append(keep, l)
		}
	}
	o, ok := parseCopilot(t, strings.Join(keep, "\n"))
	if !ok || o.Credits == nil || *o.Credits != 0 || o.Cost != 0 || o.CostSource != CostNone {
		t.Errorf("no checkpoint: %+v %v", o, ok)
	}
}

func TestCopilotParseExitCode(t *testing.T) {
	o, ok := parseCopilot(t, strings.Replace(copilotLines, `"exitCode":0`, `"exitCode":1`, 1))
	if !ok || !o.IsError {
		t.Errorf("exitCode 1: %+v %v", o, ok)
	}
}
