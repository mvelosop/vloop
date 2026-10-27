package cli

import (
	"strings"
	"testing"
)

const noIndependence = "the second reviewer is the primary's agent and model — it adds no independence"

func TestDoctorSecondReviewOffSaysNothing(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", doctorCheckConfig+"[agent]\nsecond-review = \"copilot\"\n")
	for _, c := range runDoctor(e.build, e.repo) {
		if strings.Contains(c.Name+c.Message, "second") {
			t.Errorf("doctor speaks of the second review with it off: %s %s", c.Name, c.Message)
		}
	}
}

func TestDoctorSecondReviewRowAndIndependence(t *testing.T) {
	e := newDoctorEnv(t)
	e.write(".vloop/config.toml", doctorCheckConfig+"[run]\nsecond-review = \"feedback\"\n[model]\nsecond-review = \"opus\"\n")
	var row, warn bool
	for _, c := range runDoctor(e.build, e.repo) {
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
	for _, c := range runDoctor(e.build, e.repo) {
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
