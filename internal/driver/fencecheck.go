package driver

import (
	"bytes"
	"encoding/json"
	"strings"
)

// CallVerdict is a fence judge's answer for one tool call of a session: the
// path and command it read from the call's arguments, and whether it allows
// the call and why not.
type CallVerdict struct {
	Allow   bool
	Reason  string
	Path    string
	Command string
}

// CallJudge judges one tool call, given as the session printed it, by the
// fence of the phase. The driver cannot import the fence (which imports the
// driver), so the CLI hands the judge in: the very one the Copilot hook uses.
type CallJudge func(phase, tool string, args json.RawMessage) CallVerdict

// BypassedCall is a call that ran although the fence of its phase denies it.
type BypassedCall struct {
	Tool   string
	What   string // the call's path, else the first line of its command cut to 200 runes, else empty
	Reason string
}

// bypassedCalls reads a Copilot session's JSON Lines and judges every call
// that ran: a tool.execution_start whose completion of the same toolCallId
// does not carry the error code denied, or that has none.
func bypassedCalls(stdout []byte, phase string, judge CallJudge) []BypassedCall {
	type call struct {
		tool string
		args json.RawMessage
	}
	var order []string
	calls := map[string]call{}
	denied := map[string]bool{}
	for _, line := range bytes.Split(stdout, []byte("\n")) {
		var ev struct {
			Type string `json:"type"`
			Data struct {
				ToolCallID string          `json:"toolCallId"`
				ToolName   string          `json:"toolName"`
				Arguments  json.RawMessage `json:"arguments"`
				Error      *struct {
					Code string `json:"code"`
				} `json:"error"`
			} `json:"data"`
		}
		if json.Unmarshal(bytes.TrimSpace(line), &ev) != nil {
			continue
		}
		switch ev.Type {
		case "tool.execution_start":
			id := ev.Data.ToolCallID
			if _, seen := calls[id]; !seen {
				order = append(order, id)
			}
			calls[id] = call{ev.Data.ToolName, ev.Data.Arguments}
			delete(denied, id)
		case "tool.execution_complete":
			if e := ev.Data.Error; e != nil && e.Code == "denied" {
				denied[ev.Data.ToolCallID] = true
			}
		}
	}
	var out []BypassedCall
	for _, id := range order {
		if denied[id] {
			continue
		}
		c := calls[id]
		v := judge(phase, c.tool, c.args)
		if v.Allow {
			continue
		}
		what := v.Path
		if what == "" {
			first, _, _ := strings.Cut(v.Command, "\n")
			first = strings.TrimRight(first, "\r")
			if r := []rune(first); len(r) > maxCommand {
				first = string(r[:maxCommand]) + "…"
			}
			what = first
		}
		out = append(out, BypassedCall{Tool: c.tool, What: what, Reason: v.Reason})
	}
	return out
}
