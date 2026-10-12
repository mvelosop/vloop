package metrics

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func words(n int) string { return strings.Repeat("word ", n) + "\n" }

func intp(n int) *int { return &n }

// A plan skill of 100 words and a work skill of 50, the root CLAUDE.md and a
// guidelines file of 6 rules: each kind's figure is its skill plus both shared
// files; the kinds with no skill are n/a.
func TestMeasureInstructionsWorkedExample(t *testing.T) {
	r := newRepo(t)
	r.write("CLAUDE.md", words(10))
	r.write(".vloop/repo-guidelines.md", "# Guidelines\n<!-- - not a rule -->\n- a D20260101-0900\n- b\n- c\n- d\n- e\n- f\n")
	r.write("plugin/skills/plan/SKILL.md", words(100))
	r.write("plugin/skills/work/SKILL.md", words(50))
	sha := r.commit("skills")

	got := MeasureInstructions(r.root, sha)
	shared := 10 + len(strings.Fields("# Guidelines <!-- - not a rule --> - a D20260101-0900 - b - c - d - e - f"))
	want := &Instructions{Plan: intp(100 + shared), Work: intp(50 + shared), GuidelinesRules: 6}
	b1, _ := json.Marshal(got)
	b2, _ := json.Marshal(want)
	if string(b1) != string(b2) {
		t.Errorf("instructions = %s, want %s", b1, b2)
	}
	if !strings.Contains(string(b1), `"gate_review":null`) || !strings.Contains(string(b1), `"review":null`) {
		t.Errorf("kinds without a skill are not null: %s", b1)
	}
}

// Without the plugin's skill a kind falls back to the shell loop's; a missing
// CLAUDE.md and guidelines file count 0 and 0 rules.
func TestMeasureInstructionsShellLoopFallback(t *testing.T) {
	r := newRepo(t)
	r.write(".claude/skills/loop-plan/SKILL.md", words(7))
	r.write(".claude/skills/loop-review/SKILL.md", words(3))
	r.write(".claude/skills/loop-gate-review/SKILL.md", words(9))
	sha := r.commit("shell loop")

	got := MeasureInstructions(r.root, sha)
	if got.Plan == nil || *got.Plan != 7 || got.Review == nil || *got.Review != 3 {
		t.Errorf("plan, review = %v, %v, want 7, 3", got.Plan, got.Review)
	}
	if got.Work != nil || got.GateReview != nil {
		t.Errorf("work, gate review = %v, %v, want both absent", got.Work, got.GateReview)
	}
	if got.GuidelinesRules != 0 {
		t.Errorf("rules = %d, want 0", got.GuidelinesRules)
	}
}

func TestPrintSummaryInstructions(t *testing.T) {
	r := &Report{Instructions: &Instructions{Plan: intp(4210), Work: intp(900), GuidelinesRules: 6}}
	var b bytes.Buffer
	PrintSummary(&b, r)
	want := " instructions  plan 4,210 · gate review n/a · work 900 · review n/a words · guidelines 6 rules\n"
	if !strings.Contains(b.String(), want) {
		t.Errorf("summary lacks %q:\n%s", want, b.String())
	}
	b.Reset()
	PrintSummary(&b, &Report{})
	if strings.Contains(b.String(), "instructions") {
		t.Errorf("a report with no instructions prints the line:\n%s", b.String())
	}
	j, _ := json.Marshal(r)
	if !strings.Contains(string(j), `"instructions":{"plan":4210,"gate_review":null,"work":900,"review":null,"guidelines_rules":6}`) {
		t.Errorf("json = %s", j)
	}
}
