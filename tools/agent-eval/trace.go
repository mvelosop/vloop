package main

import (
	"bytes"
	"encoding/json"
	"strings"
)

// event is one entry of a run's trace: kind "text" or "tool".
type event struct {
	Kind   string `json:"kind"`
	Text   string `json:"text,omitempty"`
	Tool   string `json:"tool,omitempty"`
	Input  string `json:"input,omitempty"`
	Output string `json:"output,omitempty"`
}

// MarshalJSON writes an event with exactly its kind's keys, and no HTML escaping.
func (e event) MarshalJSON() ([]byte, error) {
	var v any
	if e.Kind == "text" {
		v = struct {
			Kind string `json:"kind"`
			Text string `json:"text"`
		}{e.Kind, e.Text}
	} else {
		v = struct {
			Kind   string `json:"kind"`
			Tool   string `json:"tool"`
			Input  string `json:"input"`
			Output string `json:"output"`
		}{e.Kind, e.Tool, e.Input, e.Output}
	}
	return marshalPlain(v)
}

// marshalPlain is json.Marshal without HTML escaping.
func marshalPlain(v any) ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimRight(b.Bytes(), "\n"), nil
}

// copilotTools maps Copilot's tool names onto Claude's; any other is unchanged.
var copilotTools = map[string]string{
	"bash": "Bash", "powershell": "Bash",
	"create": "Write",
	"edit":   "Edit", "apply_patch": "Edit",
	"view": "Read", "read_file": "Read",
	"grep": "Grep", "rg": "Grep", "grep_search": "Grep",
	"glob": "Glob", "file_search": "Glob",
	"skill": "Skill",
}

func toolName(n string) string {
	if m, ok := copilotTools[n]; ok {
		return m
	}
	return n
}

// compactInput is a call's arguments as compact JSON, `<`, `>` and `&` as written.
func compactInput(raw json.RawMessage) string {
	if len(bytes.TrimSpace(raw)) == 0 {
		return ""
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if dec.Decode(&v) != nil {
		return strings.TrimSpace(string(raw))
	}
	b, err := marshalPlain(v)
	if err != nil {
		return strings.TrimSpace(string(raw))
	}
	return string(b)
}

// resultText is a result's content: a string as it is, anything else as compact JSON.
func resultText(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	if len(bytes.TrimSpace(raw)) == 0 || string(bytes.TrimSpace(raw)) == "null" {
		return ""
	}
	return compactInput(raw)
}

// copilotTrace reads Copilot's JSON Lines into events, in order: an
// assistant.message with content is text; a tool.execution_start is a tool
// call, paired by toolCallId with its tool.execution_complete.
func copilotTrace(transcript []byte) []event {
	var events []event
	at := map[string]int{} // toolCallId -> index in events
	for _, line := range bytes.Split(transcript, []byte("\n")) {
		var ev struct {
			Type string `json:"type"`
			Data struct {
				Content    json.RawMessage `json:"content"`
				ToolCallID string          `json:"toolCallId"`
				ToolName   string          `json:"toolName"`
				Arguments  json.RawMessage `json:"arguments"`
				Result     struct {
					Content json.RawMessage `json:"content"`
				} `json:"result"`
			} `json:"data"`
		}
		if json.Unmarshal(bytes.TrimSpace(line), &ev) != nil {
			continue
		}
		switch ev.Type {
		case "assistant.message":
			var s string
			if json.Unmarshal(ev.Data.Content, &s) == nil && s != "" {
				events = append(events, event{Kind: "text", Text: s})
			}
		case "tool.execution_start":
			at[ev.Data.ToolCallID] = len(events)
			events = append(events, event{Kind: "tool", Tool: toolName(ev.Data.ToolName), Input: compactInput(ev.Data.Arguments)})
		case "tool.execution_complete":
			if i, ok := at[ev.Data.ToolCallID]; ok {
				events[i].Output = resultText(ev.Data.Result.Content)
			}
		}
	}
	return events
}

// traceText is each event on its own line: a text event's text; a tool event's
// `<tool> <input>` and, on the next line, its output.
func traceText(events []event) string {
	var lines []string
	for _, e := range events {
		if e.Kind == "text" {
			lines = append(lines, e.Text)
		} else {
			lines = append(lines, e.Tool+" "+e.Input, e.Output)
		}
	}
	return strings.Join(lines, "\n")
}

// claudeTrace reads Claude's stream-json lines into events, in order: a text
// item of an assistant line is text; a tool_use item is a tool call, paired by
// id with the tool_result of a user line. Thinking items and the system,
// rate_limit_event and result lines are not events.
func claudeTrace(transcript []byte) []event {
	var events []event
	at := map[string]int{} // tool_use id -> index in events
	for _, line := range bytes.Split(transcript, []byte("\n")) {
		var ev struct {
			Type    string `json:"type"`
			Message struct {
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal(bytes.TrimSpace(line), &ev) != nil || (ev.Type != "assistant" && ev.Type != "user") {
			continue
		}
		var items []struct {
			Type      string          `json:"type"`
			Text      string          `json:"text"`
			ID        string          `json:"id"`
			Name      string          `json:"name"`
			Input     json.RawMessage `json:"input"`
			ToolUseID string          `json:"tool_use_id"`
			Content   json.RawMessage `json:"content"`
		}
		if json.Unmarshal(ev.Message.Content, &items) != nil {
			continue
		}
		for _, it := range items {
			switch {
			case ev.Type == "assistant" && it.Type == "text":
				if it.Text != "" {
					events = append(events, event{Kind: "text", Text: it.Text})
				}
			case ev.Type == "assistant" && it.Type == "tool_use":
				at[it.ID] = len(events)
				events = append(events, event{Kind: "tool", Tool: it.Name, Input: compactInput(it.Input)})
			case ev.Type == "user" && it.Type == "tool_result":
				if i, ok := at[it.ToolUseID]; ok {
					events[i].Output = claudeResultText(it.Content)
				}
			}
		}
	}
	return events
}

// claudeResultText is a tool_result's content: a string as it is; a list of
// {type: text, text} items as their texts joined by newlines.
func claudeResultText(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var items []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &items) == nil {
		var parts []string
		for _, it := range items {
			if it.Type == "text" {
				parts = append(parts, it.Text)
			}
		}
		return strings.Join(parts, "\n")
	}
	return resultText(raw)
}
