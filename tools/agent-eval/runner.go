package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/mvelosop/vloop/internal/driver"
	"github.com/mvelosop/vloop/internal/state"
)

// version names the folders the plugin and fence are extracted into.
const version = "agent-eval"

type runResult struct {
	Status       string         `json:"status"`
	Reason       string         `json:"reason"`
	Score        *float64       `json:"score"`
	Graders      []graderResult `json:"graders"`
	CostUSD      float64        `json:"cost_usd"`
	Credits      *float64       `json:"credits"`
	JudgeCostUSD float64        `json:"judge_cost_usd"`
	WallMS       int64          `json:"wall_ms"`
}

type caseResult struct {
	Name   string      `json:"name"`
	Phase  string      `json:"phase"`
	Status string      `json:"status"`
	Reason string      `json:"reason"`
	Mean   *float64    `json:"mean"`
	Runs   []runResult `json:"runs"`
}

type report struct {
	Agent      string       `json:"agent"`
	Model      string       `json:"model"`
	JudgeModel string       `json:"judge_model"`
	Runs       int          `json:"runs"`
	Cases      []caseResult `json:"cases"`
}

// sessionPhases are the skills a session can be started for.
var sessionPhases = map[string]string{
	"plan": driver.PhasePlan, "work": driver.PhaseWork, "review": driver.PhaseReview,
	"gate-review": driver.PhaseGateReview, "verify": driver.PhaseVerify, "review-brief": driver.PhaseReviewBrief,
}

// promptArg is the rest of the body's first line after the skill.
func promptArg(body string) string {
	for _, l := range strings.Split(body, "\n") {
		if strings.TrimSpace(l) == "" {
			continue
		}
		f := strings.Fields(l)
		return strings.Join(f[1:], " ")
	}
	return ""
}

// execute runs the selected cases, writes report.json under o.out and the table
// on stdout, and returns the exit code.
func execute(o *options, runs int, timeout time.Duration, cases []*evalCase, stdout, stderr io.Writer) int {
	out, err := filepath.Abs(o.out)
	if err == nil {
		err = os.MkdirAll(out, 0o755)
	}
	if err != nil {
		fmt.Fprintln(stderr, "agent-eval: "+oneLine(err.Error()))
		return 1
	}
	hook := ""
	if o.agent == "copilot" {
		if hook, err = filepath.Abs(findVloop(o.vloop)); err != nil {
			fmt.Fprintln(stderr, "agent-eval: "+oneLine(err.Error()))
			return 1
		}
	}
	rep := report{Agent: o.agent, Model: o.model, JudgeModel: o.judgeModel, Runs: runs}
	for _, c := range cases {
		cr := caseResult{Name: c.name, Phase: "", Runs: []runResult{}}
		phase, ok := sessionPhases[c.phase]
		if !ok {
			cr.Status, cr.Reason = "skipped", "no phase"
			rep.Cases = append(rep.Cases, cr)
			continue
		}
		cr.Phase = c.phase
		var sum float64
		var scored int
		for k := 1; k <= runs; k++ {
			dir := filepath.Join(out, c.name, fmt.Sprintf("run-%d", k))
			rr := runOnce(o, c, phase, dir, hook, timeout)
			if rr.Status == "error" {
				if cr.Reason == "" {
					cr.Reason = rr.Reason
				}
			} else {
				sum += *rr.Score
				scored++
			}
			cr.Runs = append(cr.Runs, rr)
		}
		cr.Status = "scored"
		if cr.Reason != "" {
			cr.Status = "error"
		}
		if scored > 0 {
			m := sum / float64(scored)
			cr.Mean = &m
		}
		rep.Cases = append(rep.Cases, cr)
	}
	data, err := json.MarshalIndent(rep, "", "  ")
	if err == nil {
		err = os.WriteFile(filepath.Join(out, "report.json"), append(data, '\n'), 0o644)
	}
	if err != nil {
		fmt.Fprintln(stderr, "agent-eval: "+oneLine(err.Error()))
		return 1
	}
	return printTable(stdout, rep)
}

func errorRun(reason string) runResult {
	return runResult{Status: "error", Reason: reason, Graders: []graderResult{}}
}

// runOnce scaffolds run dir/repo, starts one session in it as vloop starts the
// phase, keeps the transcript and trace and scores the graders.
func runOnce(o *options, c *evalCase, phase, dir, hook string, timeout time.Duration) runResult {
	repo := filepath.Join(dir, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		return errorRun(oneLine(err.Error()))
	}
	script, err := filepath.Abs(filepath.Join(c.dir, "scaffold.sh"))
	if err != nil {
		return errorRun(oneLine(err.Error()))
	}
	var serr bytes.Buffer
	cmd := exec.Command("bash", script)
	cmd.Dir, cmd.Stderr = repo, &serr
	if err := cmd.Run(); err != nil {
		code := -1
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		}
		return errorRun(fmt.Sprintf("scaffold exited %d: %s", code, lastLine(serr.String())))
	}

	spec := driver.Spec{Phase: phase, Iteration: 1, Arg: promptArg(c.body), Model: o.model, Agent: o.agent}
	if phase == driver.PhaseWork || phase == driver.PhaseReview || phase == driver.PhaseVerify {
		if plan, err := os.ReadFile(state.Path(repo)); err == nil {
			gates, _ := state.GateFiles(repo)
			spec.PlanSHA = state.PlanDigest(plan, gates)
		}
	}
	transcript := filepath.Join(dir, "transcript.jsonl")
	r := &driver.Runner{Root: repo, Version: version, RunID: version, RunDir: dir, Timeout: timeout,
		Transcript: transcript, HookBinary: hook, Stream: o.agent == "claude"}
	res, err := r.Run(spec)
	if err != nil {
		return errorRun("session failed: " + oneLine(err.Error()))
	}
	rr := runResult{Graders: []graderResult{}, CostUSD: res.Cost, WallMS: res.Wall.Milliseconds()}
	if res.Recorded {
		rr.Credits = recordCredits(res.Path)
	}
	switch {
	case res.TimedOut:
		rr.Status, rr.Reason = "error", "timed out after "+timeout.String()
		return rr
	case !res.Recorded:
		rr.Status, rr.Reason = "error", "no session result"
		return rr
	}

	raw, _ := os.ReadFile(transcript)
	events := copilotTrace(raw)
	if o.agent == "claude" {
		events = claudeTrace(raw)
	}
	var tb bytes.Buffer
	for _, e := range events {
		b, _ := e.MarshalJSON()
		tb.Write(b)
		tb.WriteByte('\n')
	}
	if err := os.WriteFile(filepath.Join(dir, "trace.jsonl"), tb.Bytes(), 0o644); err != nil {
		rr.Status, rr.Reason = "error", oneLine(err.Error())
		return rr
	}
	text := traceText(events)
	for _, g := range c.graders {
		gr, err := scoreGrader(g, repo, events, text, o.judgeModel)
		rr.JudgeCostUSD += gr.cost
		if err != nil {
			rr.Status, rr.Reason, rr.Graders = "error", err.Error(), []graderResult{}
			return rr
		}
		rr.Graders = append(rr.Graders, gr)
	}
	score, ok := weightedMean(rr.Graders)
	if !ok {
		rr.Status, rr.Reason, rr.Graders = "error", "no graders to score", []graderResult{}
		return rr
	}
	rr.Status, rr.Score = "scored", &score
	return rr
}

// recordCredits is the credits field of a session record, nil when it has none.
func recordCredits(path string) *float64 {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var rec struct {
		Credits *float64 `json:"credits"`
	}
	if json.Unmarshal(data, &rec) != nil {
		return nil
	}
	return rec.Credits
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}

// printTable writes the table and the total, and returns the exit code.
func printTable(w io.Writer, rep report) int {
	tw := tabwriter.NewWriter(w, 0, 8, 2, ' ', 0)
	fmt.Fprintln(tw, "CASE\tPHASE\tRUNS\tMEAN\tCOST\tNOTES")
	var skipped, errs int
	var total float64
	for _, c := range rep.Cases {
		phase, mean := c.Phase, "-"
		if phase == "" {
			phase = "-"
		}
		if c.Mean != nil {
			mean = fmt.Sprintf("%.2f", *c.Mean)
		}
		var cost float64
		below := map[string]bool{}
		for _, r := range c.Runs {
			cost += r.CostUSD + r.JudgeCostUSD
			for _, g := range r.Graders {
				if g.Score < 1 {
					below[g.Name] = true
				}
			}
		}
		total += cost
		notes := c.Reason
		switch c.Status {
		case "skipped":
			skipped++
		case "error":
			errs++
		default:
			var names []string
			for n := range below {
				names = append(names, n)
			}
			sort.Strings(names)
			notes = strings.Join(names, ", ")
		}
		fmt.Fprintf(tw, "%s\t%s\t%d\t%s\t$%.2f\t%s\n", c.Name, phase, len(c.Runs), mean, cost, notes)
	}
	tw.Flush()
	fmt.Fprintf(w, "%d case(s) · %d skipped · %d error(s) · $%.2f\n", len(rep.Cases), skipped, errs, total)
	if errs > 0 {
		return 1
	}
	return 0
}
