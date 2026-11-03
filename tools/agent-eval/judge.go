package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// judgeVotes is how many times the judge is asked; the majority decides.
const judgeVotes = 3

// judgePrompt is the question put to the judge: the grader's body and its focus.
func judgePrompt(body, focus string) string {
	return strings.TrimSpace(body) + "\n\n--- FOCUS ---\n" + focus + "\n--- END OF FOCUS ---\n\n" +
		"Judge the focus above against the criterion. Answer briefly, and make the last line of your answer exactly PASS or FAIL.\n"
}

// askJudge runs one judge call in an empty temporary directory and returns its
// answer and cost. Any failure to get a JSON result is an error, never a vote.
func askJudge(model, prompt string) (answer string, cost float64, err error) {
	dir, err := os.MkdirTemp("", "agent-eval-judge-")
	if err != nil {
		return "", 0, err
	}
	defer os.RemoveAll(dir)
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("claude", "-p", "--model", model, "--output-format", "json",
		"--strict-mcp-config", "--setting-sources", "project")
	cmd.Dir, cmd.Stdin, cmd.Stdout, cmd.Stderr = dir, strings.NewReader(prompt), &stdout, &stderr
	if err := cmd.Run(); err != nil {
		msg := lastLine(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", 0, fmt.Errorf("%s", oneLine(msg))
	}
	var out struct {
		Result *string  `json:"result"`
		Cost   *float64 `json:"total_cost_usd"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &out); err != nil || out.Result == nil {
		return "", 0, fmt.Errorf("no JSON result")
	}
	if out.Cost != nil {
		cost = *out.Cost
	}
	return *out.Result, cost, nil
}

// vote reads an answer's last line: PASS is true; FAIL or anything else false.
func vote(answer string) bool {
	return strings.TrimSpace(lastLine(answer)) == "PASS"
}

// judge asks the judge model judgeVotes times and returns whether the majority
// passed, and the calls' total cost.
func judge(model, body, focus string) (bool, float64, error) {
	prompt := judgePrompt(body, focus)
	var pass int
	var total float64
	for i := 0; i < judgeVotes; i++ {
		a, c, err := askJudge(model, prompt)
		if err != nil {
			return false, total, err
		}
		total += c
		if vote(a) {
			pass++
		}
	}
	return pass*2 > judgeVotes, total, nil
}

// scoreLLM scores an llm grader: it picks the focus, then lets the judge vote.
func scoreLLM(g grader, r graderResult, model, workspace string, events []event, text string) (graderResult, error) {
	focus := ""
	switch {
	case g.fm.values["focus"] == "trace":
		focus = text
	case g.fm.maps["focus"] != nil:
		f := g.fm.maps["focus"]
		if f["source"] != "file" {
			return fail(r, "unsupported focus"), nil
		}
		data, err := os.ReadFile(filepath.Join(workspace, filepath.FromSlash(f["path"])))
		if err != nil {
			return fail(r, "no "+f["path"]), nil
		}
		focus = string(data)
	case g.fm.values["focus"] != "":
		return fail(r, "unsupported focus "+g.fm.values["focus"]), nil
	default:
		for i := len(events) - 1; i >= 0; i-- {
			if events[i].Kind == "text" {
				focus = events[i].Text
				break
			}
		}
	}
	pass, cost, err := judge(model, g.body, focus)
	r.cost = cost
	if err != nil {
		return r, fmt.Errorf("judge failed: %s: %s", r.Name, err)
	}
	if pass {
		r.Score = 1
	} else {
		r.Reason = "judge: FAIL"
	}
	return r, nil
}
