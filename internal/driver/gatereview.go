package driver

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mvelosop/vloop/internal/schema"
)

const gateVerdictFile = tmpDir + "/gate-verdict.json"

// gateVerdict is the part of gate-verdict/v1 the driver reads.
type gateVerdict struct {
	Verdict string `json:"verdict"`
	Notes   string `json:"notes"`
	Tasks   []struct {
		Task     string `json:"task"`
		Verdict  string `json:"verdict"`
		Findings []struct {
			Summary string `json:"summary"`
			Kind    string `json:"kind"`
		} `json:"findings"`
	} `json:"tasks"`
}

// gateReviewRound runs the gate-review session of one round and keeps its
// verdict as reports/gate-review-<round>.json. A missing or invalid verdict is
// a FAIL; so is a verdict that fails the plan or any task. It returns the
// problems to send back to the planner, none when the review passed.
func (p *Planner) gateReviewRound(t term, round int) ([]string, error) {
	root := p.Root
	model, effort, err := ModelEffort(root, nil, PhaseGateReview)
	if err != nil {
		return nil, halt(ExitPreflight, "%v", err)
	}
	verdictPath := filepath.Join(root, filepath.FromSlash(gateVerdictFile))
	if err := os.Remove(verdictPath); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	t.say("reviewing the gates using %s (round %d)", model, round)
	before := refsState(root)
	gguard := snapshotGit(root)
	res, err := t.r.Run(Spec{Phase: PhaseGateReview, Model: model, Effort: effort})
	if err != nil {
		return nil, halt(ExitPreflight, "gate review session failed: %v", err)
	}
	if what := gguard.changed(); what != "" {
		return nil, halt(ExitRefsMoved, "the gate review changed %s — nothing was committed; restore it, then re-run", what)
	}
	if len(refsDiff(before, refsState(root))) > 0 {
		return nil, halt(ExitRefsMoved, "REFS MOVED gate-review — the gate review changed git refs; nothing was committed")
	}
	if res.TimedOut {
		return nil, halt(ExitSessionError, "gate review session timed out after %s — see %s", t.r.Timeout, relRunDir(root, t.r.RunDir))
	}

	report := filepath.Join(t.r.RunDir, "reports", fmt.Sprintf("gate-review-%d.json", round))
	if err := os.MkdirAll(filepath.Dir(report), 0o755); err != nil {
		return nil, err
	}
	data, rerr := readHandoff(verdictPath)
	var v gateVerdict
	var problems []string
	switch {
	case rerr != nil:
		problems = []string{"the gate review left no " + gateVerdictFile + " — that is a FAIL"}
	default:
		if viol, err := schema.Validate("gate-verdict/v1", data); err != nil || len(viol) > 0 {
			problems = []string{"the gate review's " + gateVerdictFile + " does not follow gate-verdict/v1 — that is a FAIL"}
		} else if err := json.Unmarshal(data, &v); err != nil {
			problems = []string{"the gate review's " + gateVerdictFile + " cannot be read — that is a FAIL"}
		}
	}
	if len(problems) == 0 {
		failed := v.Verdict == "FAIL"
		for _, tv := range v.Tasks {
			failed = failed || tv.Verdict == "FAIL"
			for _, f := range tv.Findings {
				problems = append(problems, fmt.Sprintf("gate %s (%s): %s", tv.Task, f.Kind, f.Summary))
			}
		}
		if failed && len(problems) == 0 {
			problems = []string{"the gate review failed the plan: " + v.Notes}
		}
		if !failed {
			problems = nil
		}
	}
	body := data
	if len(problems) > 0 && (rerr != nil || v.Verdict == "") {
		// Nothing usable to keep: record the FAIL the driver ruled.
		body, _ = json.Marshal(map[string]any{"schema": "gate-verdict/v1", "verdict": "FAIL", "tasks": []any{}, "notes": strings.Join(problems, "; ")})
	}
	if err := os.WriteFile(report, []byte(t.r.Mask(string(body))+"\n"), 0o644); err != nil {
		return nil, err
	}
	return problems, nil
}

// gateReviewFailedLine is the line a plan the gate review failed twice ends
// the run with; the path is repo-relative.
func gateReviewFailedLine(root, runDir string, round int) string {
	return fmt.Sprintf("the gate review failed the plan twice — see %sreports/gate-review-%d.json; amend the gates with vloop task verify, then re-run", relRunDir(root, runDir), round)
}

// resumeBaseGates is the whole of a resume of a plan blocked at the gate
// review: the gates run on the base once more and the review does not. The
// operator's amendment through vloop task verify is the ruling, so the plan's
// gate review stands as passed once every gate fails for want of the work.
func (it *Iterator) resumeBaseGates() error {
	pl := &Planner{Root: it.Root, Out: it.Out, Err: it.Err, Quiet: it.Quiet}
	problems, err := pl.baseGates(term{pl, it.r}, it.plan, it.gateTimeout())
	if err != nil {
		return err
	}
	if len(problems) > 0 {
		for _, l := range problems {
			it.warn("  %s", l)
		}
		return halt(ExitBlocked, "the gates still do not hold on the base (%d problem(s)) — amend them with vloop task verify, then re-run", len(problems))
	}
	it.plan.GateReview.Verdict = "PASS"
	return it.save()
}
