package metrics

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/mvelosop/vloop/internal/runs"
)

// Phases are the session phases metrics report on.
const (
	PhasePlan       = "plan"
	PhaseWork       = "work"
	PhaseReview     = "review"
	PhaseGateReview = "gate-review"
)

// Time is where a brief's time went, in milliseconds. Agent time is work plus
// review; plan is reported apart. Gate and wall are nil when unknown: the shell
// loop never recorded gate durations, and wall needs a run commit.
type Time struct {
	AgentMS      int64  `json:"agent_ms"`
	WorkMS       int64  `json:"work_ms"`
	ReviewMS     int64  `json:"review_ms"`
	PlanMS       int64  `json:"plan_ms"`
	GateMS       *int64 `json:"gate_ms"`
	ChecksMS     *int64 `json:"checks_ms"`
	GateReviewMS int64  `json:"gate_review_ms"`
	WallMS       *int64 `json:"wall_ms"`
}

// Rate is delivered lines per minute of agent time; nil when there is no agent
// time or no lines to divide.
type Rate struct {
	CodePerMin     *float64 `json:"code_per_min"`
	CodeTestPerMin *float64 `json:"code_test_per_min"`
}

// Cost is the sum of unrounded session costs in USD. Rounding is for display.
type Cost struct {
	Total       float64  `json:"total"`
	Plan        float64  `json:"plan"`
	Work        float64  `json:"work"`
	Review      float64  `json:"review"`
	GateReview  float64  `json:"gate_review"`
	Per1kCode   *float64 `json:"per_1k_code_lines"`
	PerPhase1kC struct {
		Plan   *float64 `json:"plan"`
		Work   *float64 `json:"work"`
		Review *float64 `json:"review"`
	} `json:"per_phase_per_1k_code_lines"`
}

// TokenTotals are token counts summed over every counted session.
type TokenTotals struct {
	Input         int64    `json:"input"`
	Output        int64    `json:"output"`
	CacheRead     int64    `json:"cache_read"`
	CacheCreation int64    `json:"cache_creation"`
	CacheHit      *float64 `json:"cache_hit_ratio"`
}

// Usage is everything the session records say about a brief.
type Usage struct {
	Time    Time
	Rate    Rate
	Cost    Cost
	Tokens  TokenTotals
	Models  map[string]string // phase -> sorted, comma-joined models
	Effort  map[string]string // phase -> sorted, comma-joined efforts, where recorded
	Denials int
	Missing []string // `<task> <phase> (iteration <i>)`
}

// reviewRan is the set of iteration outcomes that mean a review session ran.
var reviewRan = map[string]bool{"done": true, "review_fail": true, "rejected": true}

// MissingRecords lists the iterations whose session records are absent: any
// iteration without a work session, and any whose outcome means a review ran
// without a review session. Sessions are matched within their own run folder.
func MissingRecords(m *runs.Model) []string {
	var out []string
	for _, f := range m.Folders {
		have := map[string]bool{}
		for _, s := range f.Sessions {
			have[s.Phase+"/"+strconv.Itoa(s.Iteration)] = true
		}
		for _, it := range f.Iterations {
			task := it.Task
			if task == "" {
				task = "-"
			}
			if !have[PhaseWork+"/"+strconv.Itoa(it.Iteration)] {
				out = append(out, fmt.Sprintf("%s %s (iteration %d)", task, PhaseWork, it.Iteration))
			}
			if reviewRan[it.Outcome] && !have[PhaseReview+"/"+strconv.Itoa(it.Iteration)] {
				out = append(out, fmt.Sprintf("%s %s (iteration %d)", task, PhaseReview, it.Iteration))
			}
		}
	}
	return out
}

func fratio(n, d float64) *float64 {
	if d == 0 {
		return nil
	}
	v := n / d
	return &v
}

// Aggregate sums a brief's sessions. delivered is the delivered line count
// (see Measure); it feeds the rates and the cost per 1,000 code lines.
func Aggregate(m *runs.Model, delivered Counts) Usage {
	u := Usage{
		Models:  map[string]string{},
		Effort:  map[string]string{},
		Missing: MissingRecords(m),
	}
	models := map[string]map[string]bool{}
	efforts := map[string]map[string]bool{}
	var gate int64
	gateKnown := false
	var checks int64
	checksKnown := false
	iterations := 0

	for _, f := range m.Folders {
		for _, it := range f.Iterations {
			iterations++
			if it.GateMS != nil {
				gate += *it.GateMS
				gateKnown = true
			}
			if it.ChecksMS != nil {
				checks += *it.ChecksMS
				checksKnown = true
			}
		}
		for _, s := range f.Sessions {
			switch s.Phase {
			case PhasePlan:
				u.Time.PlanMS += s.DurationMS
				u.Cost.Plan += s.CostUSD
			case PhaseWork:
				u.Time.WorkMS += s.DurationMS
				u.Cost.Work += s.CostUSD
			case PhaseReview:
				u.Time.ReviewMS += s.DurationMS
				u.Cost.Review += s.CostUSD
			case PhaseGateReview:
				u.Time.GateReviewMS += s.DurationMS
				u.Cost.GateReview += s.CostUSD
			default:
				continue
			}
			u.Denials += s.Denials
			for name, t := range s.Models {
				if models[s.Phase] == nil {
					models[s.Phase] = map[string]bool{}
				}
				models[s.Phase][name] = true
				u.Tokens.Input += t.Input
				u.Tokens.Output += t.Output
				u.Tokens.CacheRead += t.CacheRead
				u.Tokens.CacheCreation += t.CacheCreation
			}
			if len(s.Models) == 0 && s.Model != "" {
				if models[s.Phase] == nil {
					models[s.Phase] = map[string]bool{}
				}
				models[s.Phase][s.Model] = true
			}
			if s.Effort != "" {
				if efforts[s.Phase] == nil {
					efforts[s.Phase] = map[string]bool{}
				}
				efforts[s.Phase][s.Effort] = true
			}
		}
	}
	for p, set := range models {
		u.Models[p] = joinSet(set)
	}
	for p, set := range efforts {
		u.Effort[p] = joinSet(set)
	}

	u.Time.AgentMS = u.Time.WorkMS + u.Time.ReviewMS
	if gateKnown {
		u.Time.GateMS = &gate
	}
	if checksKnown {
		u.Time.ChecksMS = &checks
	}
	if o := m.Owned; o != nil && o.Run != nil {
		w := o.Run.Time.Sub(o.Plan.Time).Milliseconds()
		u.Time.WallMS = &w
	}

	minutes := float64(u.Time.AgentMS) / 60000
	u.Rate.CodePerMin = fratio(float64(delivered.Code), minutes)
	u.Rate.CodeTestPerMin = fratio(float64(delivered.Code+delivered.Test), minutes)

	u.Cost.Total = u.Cost.Plan + u.Cost.Work + u.Cost.Review + u.Cost.GateReview
	kc := float64(delivered.Code) / 1000
	u.Cost.Per1kCode = fratio(u.Cost.Total, kc)
	u.Cost.PerPhase1kC.Plan = fratio(u.Cost.Plan, kc)
	u.Cost.PerPhase1kC.Work = fratio(u.Cost.Work, kc)
	u.Cost.PerPhase1kC.Review = fratio(u.Cost.Review, kc)

	u.Tokens.CacheHit = fratio(float64(u.Tokens.CacheRead),
		float64(u.Tokens.Input+u.Tokens.CacheRead+u.Tokens.CacheCreation))
	return u
}

func joinSet(set map[string]bool) string {
	var l []string
	for k := range set {
		l = append(l, k)
	}
	sort.Strings(l)
	return strings.Join(l, ",")
}

// Estimate is the brief's `<n> to <m> tasks` phrase.
type Estimate struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

var (
	estimateRe = regexp.MustCompile(`(\d+) (?:to|a|–|-) (\d+) (?:tasks|tareas)`)
	shapeRe    = regexp.MustCompile(`(?im)^##[ \t]+(?:shape|forma)\b.*$`)
	sectionRe  = regexp.MustCompile(`(?m)^##[ \t]`)
)

// BriefEstimate reads the estimate from the brief's Shape section (Forma in
// Spanish); nil when it states none. Only that section counts: a worked
// example may describe a fixture brief with an estimate of its own.
func BriefEstimate(brief string) *Estimate {
	loc := shapeRe.FindStringIndex(brief)
	if loc == nil {
		return nil
	}
	shape := brief[loc[1]:]
	if next := sectionRe.FindStringIndex(shape); next != nil {
		shape = shape[:next[0]]
	}
	m := estimateRe.FindStringSubmatch(shape)
	if m == nil {
		return nil
	}
	lo, _ := strconv.Atoi(m[1])
	hi, _ := strconv.Atoi(m[2])
	return &Estimate{Min: lo, Max: hi}
}

// Tasks counts the plan's tasks and how they went.
type Tasks struct {
	Planned             int       `json:"planned"`
	Done                int       `json:"done"`
	Blocked             int       `json:"blocked"`
	FirstPass           int       `json:"first_pass"`
	Estimate            *Estimate `json:"estimate"`
	Iterations          int       `json:"iterations"`
	IterationsPerClosed *float64  `json:"iterations_per_closed"`
}

// CountTasks counts tasks in the plan as committed at the last run commit.
// A task is first-pass when it has iterations and every one ended `done`.
// briefText is the brief's content, for its estimate; "" states none.
func CountTasks(m *runs.Model, plan *runs.PlanDoc, briefText string) Tasks {
	t := Tasks{Estimate: BriefEstimate(briefText)}
	outcomes := map[string][]string{}
	for _, f := range m.Folders {
		for _, it := range f.Iterations {
			t.Iterations++
			outcomes[it.Task] = append(outcomes[it.Task], it.Outcome)
		}
	}
	if plan != nil {
		for _, pt := range plan.Tasks {
			t.Planned++
			switch pt.Status {
			case "done":
				t.Done++
			case "blocked":
				t.Blocked++
			}
			if os := outcomes[pt.ID]; len(os) > 0 && onlyDone(os) {
				t.FirstPass++
			}
		}
	}
	t.IterationsPerClosed = fratio(float64(t.Iterations), float64(t.Done))
	return t
}

func onlyDone(os []string) bool {
	for _, o := range os {
		if o != "done" {
			return false
		}
	}
	return true
}
