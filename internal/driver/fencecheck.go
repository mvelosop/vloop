package driver

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
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

// hookBinaryPath is the binary a Copilot session's fence hook runs: the one
// writeCopilotHooks writes into the hooks file.
func hookBinaryPath(configured string) string {
	if configured != "" {
		return configured
	}
	exe, _ := os.Executable()
	return exe
}

// fileHash is the SHA-256 of a file, "" when it cannot be read.
func fileHash(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}

// backstop is the after-session check every phase of a Copilot session shares:
// the hook's binary first (F2), then the calls the fence denies that ran (F1).
// It writes its lines through warn, which masks and logs, and returns the exit
// 9 halt, nil when the session passes.
func (res Result) backstop(phase string, warn func(format string, a ...any)) *Halt {
	if res.HookChanged {
		h := halt(ExitRefsMoved, "HOOK CHANGED %s — the binary the fence hook runs changed during the session: %s; nothing from it was used", phase, res.HookPath)
		warn("%v", h)
		return h
	}
	if len(res.Bypassed) == 0 {
		return nil
	}
	for _, c := range res.Bypassed {
		if c.What == "" {
			warn("vloop: %s — %s", c.Tool, c.Reason)
		} else {
			warn("vloop: %s %s — %s", c.Tool, c.What, c.Reason)
		}
	}
	h := halt(ExitRefsMoved, "FENCE BYPASSED %s — the session ran %d call(s) its fence denies; nothing from it was used", phase, len(res.Bypassed))
	warn("%v", h)
	return h
}
