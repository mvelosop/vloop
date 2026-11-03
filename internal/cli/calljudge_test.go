package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestCallJudge pins the driver's re-judge to the hook's own reading and
// rule: string toolArgs, the empty-command rule, a file tool's path.
func TestCallJudge(t *testing.T) {
	judge := CallJudge(t.TempDir())
	cases := []struct {
		name, tool, args string
		allow            bool
		reason           string // substring
		path, command    string
	}{
		{"object command denied", "bash", `{"command":"git commit -m x"}`, false, "", "", "git commit -m x"},
		{"string toolArgs", "bash", `"{\"command\":\"git commit -m x\"}"`, false, "", "", "git commit -m x"},
		{"allowed ls", "bash", `{"command":"ls"}`, true, "", "", "ls"},
		{"empty bash command", "bash", `{"command":"  "}`, false, "the bash call has no command the hook can read", "", "  "},
		{"no args bash", "bash", `{}`, false, "the bash call has no command the hook can read", "", ""},
		{"empty powershell", "powershell", `{}`, false, "the powershell call has no command the hook can read", "", ""},
		{"file path", "create", `{"path":".vloop/state/state.json"}`, false, "", ".vloop/state/state.json", ""},
		{"skill name", "skill", `{"name":"x"}`, false, "", "x", ""},
	}
	for _, c := range cases {
		v := judge("work", c.tool, json.RawMessage(c.args))
		if v.Allow != c.allow {
			t.Errorf("%s: allow = %v (%s), want %v", c.name, v.Allow, v.Reason, c.allow)
		}
		if !strings.Contains(v.Reason, c.reason) {
			t.Errorf("%s: reason %q lacks %q", c.name, v.Reason, c.reason)
		}
		if !c.allow && v.Reason == "" {
			t.Errorf("%s: denied without a reason", c.name)
		}
		if v.Path != c.path || v.Command != c.command {
			t.Errorf("%s: path/command = %q/%q, want %q/%q", c.name, v.Path, v.Command, c.path, c.command)
		}
	}
	if v := judge("nosuchphase", "bash", json.RawMessage(`{"command":"ls"}`)); v.Allow || v.Reason == "" {
		t.Errorf("unknown phase must deny with a reason: %+v", v)
	}
}
