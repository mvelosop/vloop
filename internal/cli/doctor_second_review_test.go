package cli

import (
	"strings"
	"testing"
)

const noIndependence = "the second reviewer is the primary's agent and model — it adds no independence"

func TestDoctorSecondReviewOffSaysNothing(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", doctorCheckConfig+"[agent]\nsecond-review = \"copilot\"\n")
	for _, c := range runDoctor(e.build, e.repo, nil) {
		if strings.Contains(c.Name+c.Message, "second") {
			t.Errorf("doctor speaks of the second review with it off: %s %s", c.Name, c.Message)
		}
	}
}

func TestDoctorSecondReviewRowAndIndependence(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", doctorCheckConfig+"[run]\nsecond-review = \"feedback\"\n[model]\nsecond-review = \"opus\"\n")
	var row, warn bool
	for _, c := range runDoctor(e.build, e.repo, nil) {
		if c.Name == "run.second-review:" {
			if c.Result == resWarning {
				warn = true
			} else if c.Message == "feedback, agent claude, model opus" {
				row = true
			}
		}
	}
	if !row || warn {
		t.Errorf("row = %v, warning = %v; want the row and no warning", row, warn)
	}
	e.write(".vloop/config.toml", doctorCheckConfig+"[run]\nsecond-review = \"record\"\n")
	var got bool
	for _, c := range runDoctor(e.build, e.repo, nil) {
		if c.Name == "run.second-review:" && c.Result == resWarning && c.Message == noIndependence {
			got = true
		}
	}
	if !got {
		t.Error("no independence warning for a second reviewer matching the primary")
	}
}

func TestDoctorCopilotSecondReviewer(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", doctorCheckConfig+"[run]\nsecond-review = \"record\"\n[agent]\nsecond-review = \"copilot\"\n[model]\nsecond-review = \"gpt-x\"\n")
	e.copilot(true)
	e.want("copilot", resPass)
	c, _ := e.check("agent.second-review")
	if c.Result != resWarning || "agent.second-review "+c.Message != "agent.second-review is copilot: experimental (docs/guide/configuration.md)" {
		t.Errorf("agent.second-review = %s %q", c.Result, c.Message)
	}
}

func TestRunPreflightCopilotSecondReview(t *testing.T) {
	e := newDoctorEnv(t)
	e.copilot(true)
	e.write(".vloop/config.toml", doctorCheckConfig+"[run]\nsecond-review = \"record\"\n[agent]\nsecond-review = \"copilot\"\n")
	if _, err := e.preflight(); err == nil || !strings.Contains(err.Error(), "model.second-review") {
		t.Errorf("default model.second-review: err = %v", err)
	}
	e.write(".vloop/config.toml", doctorCheckConfig+"[run]\nsecond-review = \"record\"\n[agent]\nsecond-review = \"copilot\"\n[model]\nsecond-review = \"gpt-x\"\n")
	e.skill("review")
	if _, err := e.preflight(); err == nil || !strings.Contains(err.Error(), ".claude/skills/review") {
		t.Errorf("skill clash: err = %v", err)
	}
	e.write(".vloop/config.toml", doctorCheckConfig+"[agent]\nsecond-review = \"copilot\"\n")
	if _, err := e.preflight(); err != nil {
		t.Errorf("second review off: err = %v", err)
	}
}

// V1: the plan check tells the bare doctor, a run naming another brief and a
// run naming none or the plan's own apart for a complete state/v2 plan.
func TestDoctorCompletePlanOfAnotherBrief(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", doctorCheckConfig)
	e.write(".vloop/state/state.json", `{"schema":"state/v2","run_id":"r","brief":"docs/briefs/a.md","status":"complete","shell":"sh","tasks":[]}`)
	plan := func(run *string) doctorCheck {
		for _, c := range runDoctor(e.build, e.repo, run) {
			if c.Name == "plan" {
				return c
			}
		}
		t.Fatal("no plan check")
		return doctorCheck{}
	}
	pass := "a complete plan of docs/briefs/a.md — the next brief's plan replaces it"
	other, same, none := "docs/briefs/b.md", "docs/briefs/a.md", ""
	if c := plan(nil); c.Result != resPass || c.Message != pass {
		t.Errorf("bare doctor: %s %q, want pass %q", c.Result, c.Message, pass)
	}
	if c := plan(&other); c.Result != resPass || c.Message != pass {
		t.Errorf("run naming another brief: %s %q, want pass %q", c.Result, c.Message, pass)
	}
	for _, r := range []*string{&same, &none} {
		if c := plan(r); c.Message == pass {
			t.Errorf("run naming %q: the complete plan was not validated", *r)
		}
	}
	e.write(".vloop/state/state.json", `{"schema":"state/v2","run_id":"r","brief":"docs/briefs/a.md","status":"running","shell":"sh","tasks":[]}`)
	if c := plan(nil); c.Message == pass {
		t.Error("a running plan passed as complete")
	}
}
