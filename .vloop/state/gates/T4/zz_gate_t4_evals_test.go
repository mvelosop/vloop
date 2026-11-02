package cli

// The T4 gate's oracle (B28, K4): the two new eval cases pass the plugin's
// structural eval checks and the guide's --allow-tools, invoke their skill, and
// the create-brief case grades what K4 pins. Copied into a copy of the tree by
// the gate; never committed.

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var gateCases = map[string]string{
	"review-brief-planted-faults": "/vloop:review-brief",
	"create-brief-drafts-checked": "/vloop:create-brief",
}

func gateUnquote(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 2 && v[0] == '\'' && v[len(v)-1] == '\'' {
		return strings.ReplaceAll(v[1:len(v)-1], "''", "'")
	}
	if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
		if u, err := strconv.Unquote(v); err == nil {
			return u
		}
	}
	return v
}

type gateGrader struct {
	file string
	fm   map[string]string
}

func gateGraders(t *testing.T, name string) []gateGrader {
	t.Helper()
	files, _ := filepath.Glob(filepath.Join(skillCheckDefault, "evals", name, "graders", "*.md"))
	var out []gateGrader
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, gateGrader{filepath.Base(f), frontmatter(string(b))})
	}
	return out
}

// gateCheckOut: whether re matches what `vloop brief check <copy>` prints with n
// warnings, as the trace holds it — raw, or JSON-escaped.
func gateCheckOut(re *regexp.Regexp, n string) bool {
	out := "/tmp/x/B20260101-0900-greet-shout.loop-brief.md\n  ✓ frontmatter is valid\n  ✓ has a worked example\n  ok, " + n + " warning(s)\nbriefs ok\n"
	esc := strings.ReplaceAll(out, "\n", `\n`)
	return re.MatchString(out) || re.MatchString(esc)
}

func TestGateT4Structure(t *testing.T) {
	guide, err := os.ReadFile(evalsGuide)
	if err != nil {
		t.Fatal(err)
	}
	g := strings.ReplaceAll(string(guide), "\\\n", " ") + " "
	all := append(checkEvals(skillCheckDefault), evalGrantGaps(skillCheckDefault, fenceAllowRules(t), g)...)
	for name, skill := range gateCases {
		dir := filepath.Join(skillCheckDefault, "evals", name)
		for _, f := range []string{"case.yaml", "prompt.md", "scaffold.sh"} {
			if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
				t.Errorf("eval %s: no %s", name, f)
			}
		}
		for _, v := range all {
			if strings.HasPrefix(v, "eval "+name+":") {
				t.Errorf("%s", v)
			}
		}
		b, err := os.ReadFile(filepath.Join(dir, "prompt.md"))
		if err != nil {
			continue
		}
		body := string(b)
		if i := strings.Index(body[3:], "\n---"); strings.HasPrefix(body, "---") && i >= 0 {
			body = body[3+i+4:]
		}
		if !strings.Contains(body, skill) {
			t.Errorf("eval %s: prompt.md does not invoke %s", name, skill)
		}
	}
}

func TestGateT4ReviewBriefGraders(t *testing.T) {
	gs := gateGraders(t, "review-brief-planted-faults")
	// K4: one grader that the verdict is not PASS, one per planted fault (three),
	// one that the brief is not edited — of whatever documented type.
	if len(gs) < 5 {
		t.Errorf("eval review-brief-planted-faults: %d graders; want at least 5 (the verdict, one per planted fault, the brief not edited)", len(gs))
	}
}

func TestGateT4CreateBriefGraders(t *testing.T) {
	gs := gateGraders(t, "create-brief-drafts-checked")
	toolUsed := func(cmd, other string) bool {
		for _, g := range gs {
			if g.fm["type"] != "tool_used" || g.fm["tool"] != "Bash" {
				continue
			}
			re, err := regexp.Compile(gateUnquote(g.fm["input_match"]))
			if err != nil || g.fm["input_match"] == "" {
				continue
			}
			min, err := strconv.Atoi(g.fm["min"])
			if err != nil || min < 1 {
				continue
			}
			if mx, ok := g.fm["max"]; ok && mx != "" {
				if n, err := strconv.Atoi(mx); err != nil || n < 1 {
					continue
				}
			}
			if re.MatchString(cmd) && !re.MatchString(other) {
				return true
			}
		}
		return false
	}
	if !toolUsed("vloop brief new greet-shout", "vloop brief check docs/briefs/B20260101-0900-greet-shout.loop-brief.md") {
		t.Error("eval create-brief-drafts-checked: no tool_used Bash grader, min 1 or more, matches `vloop brief new <slug>` (and not `vloop brief check`)")
	}
	if !toolUsed("vloop brief check /tmp/x/B20260101-0900-greet-shout.loop-brief.md", "vloop brief new greet-shout") {
		t.Error("eval create-brief-drafts-checked: no tool_used Bash grader, min 1 or more, matches `vloop brief check <copy>` (and not `vloop brief new`)")
	}
	regex := false
	for _, g := range gs {
		if g.fm["type"] != "regex" || g.fm["target"] != "trace" {
			continue
		}
		re, err := regexp.Compile(gateUnquote(g.fm["pattern"]))
		if err == nil && g.fm["pattern"] != "" && gateCheckOut(re, "0") && !gateCheckOut(re, "2") {
			regex = true
		}
	}
	if !regex {
		t.Error("eval create-brief-drafts-checked: no regex grader on the trace that matches the check's `  ok, 0 warning(s)` line and not `  ok, 2 warning(s)`")
	}
	llm := false
	for _, g := range gs {
		if g.fm["type"] == "llm" && gateUnquote(g.fm["focus"]) == "trace" {
			llm = true
		}
	}
	if !llm {
		t.Error("eval create-brief-drafts-checked: no llm grader with focus: trace")
	}
}
