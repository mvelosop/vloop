package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeClaude installs a claude on PATH that records each call's stdin and
// answers from answers, one per call; "EXIT" makes the call fail, "RAW:x" prints x.
func fakeClaude(t *testing.T, answers ...string) (stdinDir string) {
	t.Helper()
	bin, st := t.TempDir(), t.TempDir()
	var cases strings.Builder
	for i, a := range answers {
		var out string
		switch {
		case a == "EXIT":
			out = "echo down >&2; exit 1"
		case strings.HasPrefix(a, "RAW:"):
			out = "echo '" + strings.TrimPrefix(a, "RAW:") + "'"
		default:
			out = `printf '{"result":"reasoning\\n` + a + `","total_cost_usd":0.01}\n'`
		}
		cases.WriteString("  " + string(rune('1'+i)) + ") " + out + " ;;\n")
	}
	script := "#!/bin/sh\nST='" + st + "'\nn=$(( $(cat \"$ST/n\" 2>/dev/null || echo 0) + 1 ))\necho $n > \"$ST/n\"\ncat > \"$ST/stdin.$n\"\ncase $n in\n" + cases.String() + "esac\n"
	write(t, filepath.Join(bin, "claude"), script)
	if err := os.Chmod(filepath.Join(bin, "claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return st
}

func llmGrader(t *testing.T, front string) grader {
	t.Helper()
	lines, body, err := splitFrontmatter("---\ntype: llm\n" + front + "---\nCRITERION\n")
	if err != nil {
		t.Fatal(err)
	}
	fm, err := parseFrontmatter(lines)
	if err != nil {
		t.Fatal(err)
	}
	return grader{file: "g.md", fm: fm, body: body}
}

func TestJudgeMajority(t *testing.T) {
	for _, c := range []struct {
		answers []string
		score   float64
	}{
		{[]string{"PASS", "PASS", "PASS"}, 1},
		{[]string{"PASS", "FAIL", "PASS"}, 1},
		{[]string{"FAIL", "PASS", "FAIL"}, 0},
		{[]string{"PASS", "MAYBE", "FAIL"}, 0}, // a non-answer is FAIL
		{[]string{"PASS", "PASS", "MAYBE"}, 1},
	} {
		fakeClaude(t, c.answers...)
		r, err := scoreLLM(llmGrader(t, ""), graderResult{Name: "g"}, "jm", t.TempDir(), []event{{Kind: "text", Text: "x"}}, "x")
		if err != nil || r.Score != c.score {
			t.Errorf("%v: score %v, err %v, want %v", c.answers, r.Score, err, c.score)
		}
		if r.cost < 0.0299 || r.cost > 0.0301 {
			t.Errorf("%v: cost %v, want 0.03", c.answers, r.cost)
		}
	}
}

func TestJudgeFailureIsAnError(t *testing.T) {
	for _, a := range []string{"EXIT", "RAW:not json", `RAW:{"type":"result"}`} {
		fakeClaude(t, "PASS", a, "PASS")
		_, err := scoreLLM(llmGrader(t, ""), graderResult{Name: "g"}, "jm", t.TempDir(), nil, "")
		if err == nil || !strings.HasPrefix(err.Error(), "judge failed: g: ") {
			t.Errorf("%s: err %v, want 'judge failed: g: ...'", a, err)
		}
	}
}

func TestJudgeFocus(t *testing.T) {
	ws := t.TempDir()
	write(t, filepath.Join(ws, "f.txt"), "FILE-CONTENT")
	events := []event{{Kind: "text", Text: "first"}, {Kind: "tool", Tool: "Bash", Input: "{}", Output: "o"}, {Kind: "text", Text: "LAST"}}
	text := traceText(events)
	for _, c := range []struct {
		front, want, not string
	}{
		{"", "LAST", "first"},
		{"focus: trace\n", "Bash {}", ""},
		{"focus:\n  source: file\n  path: f.txt\n", "FILE-CONTENT", "LAST"},
	} {
		st := fakeClaude(t, "PASS", "PASS", "PASS")
		if _, err := scoreLLM(llmGrader(t, c.front), graderResult{Name: "g"}, "jm", ws, events, text); err != nil {
			t.Fatal(err)
		}
		in, _ := os.ReadFile(filepath.Join(st, "stdin.1"))
		s := string(in)
		if !strings.Contains(s, "CRITERION") || !strings.Contains(s, c.want) || (c.not != "" && strings.Contains(s, c.not)) {
			t.Errorf("focus %q: stdin %q", c.front, s)
		}
	}
}

func TestJudgeMissingFocusFile(t *testing.T) {
	st := fakeClaude(t, "PASS", "PASS", "PASS")
	r, err := scoreLLM(llmGrader(t, "focus:\n  source: file\n  path: gone.json\n"), graderResult{Name: "g"}, "jm", t.TempDir(), nil, "")
	if err != nil || r.Score != 0 || r.Reason != "no gone.json" {
		t.Errorf("got %+v, %v", r, err)
	}
	if _, err := os.Stat(filepath.Join(st, "n")); err == nil {
		t.Error("the judge was called for a missing file")
	}
}
