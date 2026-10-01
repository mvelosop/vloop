package metrics

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mvelosop/vloop/internal/classify"
	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/defect"
	"github.com/mvelosop/vloop/internal/runs"
)

// Schema is the name of the document Report marshals to.
const Schema = "metrics/v1"

// JSONLines is Lines as metrics/v1 spells it: added counts, deleted apart.
type JSONLines struct {
	Code    int    `json:"code"`
	Test    int    `json:"test"`
	Docs    int    `json:"docs"`
	Other   int    `json:"other"`
	Deleted Counts `json:"deleted"`
}

func jsonLines(l Lines) JSONLines {
	return JSONLines{l.Added.Code, l.Added.Test, l.Added.Docs, l.Added.Other, l.Deleted}
}

// ReportSize is the `size` object.
type ReportSize struct {
	Delivered     JSONLines `json:"delivered"`
	Churn         JSONLines `json:"churn"`
	TestCodeRatio *float64  `json:"test_code_ratio"`
	Rework        *float64  `json:"rework"`
}

// ReportTime is the `time` object, in milliseconds.
type ReportTime struct {
	AgentMS  int64  `json:"agent_ms"`
	WorkMS   int64  `json:"work_ms"`
	ReviewMS int64  `json:"review_ms"`
	PlanMS   int64  `json:"plan_ms"`
	GatesMS  *int64 `json:"gates_ms"`
	WallMS   *int64 `json:"wall_ms"`
}

// ReportCost is the `cost` object, in unrounded USD.
type ReportCost struct {
	Total       float64  `json:"total_usd"`
	Plan        float64  `json:"plan_usd"`
	Work        float64  `json:"work_usd"`
	Review      float64  `json:"review_usd"`
	Per1000Code *float64 `json:"per_1000_code_lines_usd"`
}

// PerPhase holds one value per session phase.
type PerPhase[T any] struct {
	Plan   T `json:"plan"`
	Work   T `json:"work"`
	Review T `json:"review"`
}

// ReportDefects is the `defects` object.
type ReportDefects struct {
	InLoop            int      `json:"in_loop"`
	Operator          int      `json:"operator"`
	Escaped           int      `json:"escaped"`
	Total             int      `json:"total"`
	RemovalEfficiency *float64 `json:"removal_efficiency"`
}

// MissingRecord is a session record that should exist and does not.
type MissingRecord struct {
	Task      string `json:"task"`
	Phase     string `json:"phase"`
	Iteration int    `json:"iteration"`
}

// ReportTask is one `by_task` row.
type ReportTask struct {
	ID       string    `json:"id"`
	Area     *string   `json:"area"`
	Kind     *string   `json:"kind"`
	Attempts int       `json:"attempts"`
	Churn    JSONLines `json:"churn"`
	AgentMS  int64     `json:"agent_ms"`
	CostUSD  float64   `json:"cost_usd"`
	Models   []string  `json:"models"`
}

// Report is the metrics/v1 document for one brief.
type Report struct {
	Schema  string             `json:"schema"`
	Brief   string             `json:"brief"`
	RunID   string             `json:"run_id"`
	Status  string             `json:"status"`
	Merged  *string            `json:"merged"`
	Tasks   ReportTasks        `json:"tasks"`
	Size    ReportSize         `json:"size"`
	Time    ReportTime         `json:"time"`
	Rate    Rate               `json:"rate"`
	Cost    ReportCost         `json:"cost"`
	Tokens  TokenTotals        `json:"tokens"`
	Models  PerPhase[[]string] `json:"models"`
	Effort  PerPhase[*string]  `json:"effort"`
	Defects ReportDefects      `json:"defects"`
	Records struct {
		Missing []MissingRecord `json:"missing"`
	} `json:"records"`
	LeadTime struct {
		PlanToMergeMS *int64 `json:"plan_to_merge_ms"`
	} `json:"lead_time"`
	PermissionDenials   int          `json:"permission_denials"`
	Iterations          int          `json:"iterations"`
	IterationsPerClosed *float64     `json:"iterations_per_closed"`
	ByTask              []ReportTask `json:"by_task"`
}

// ReportTasks is the `tasks` object.
type ReportTasks struct {
	Planned   int       `json:"planned"`
	Done      int       `json:"done"`
	Blocked   int       `json:"blocked"`
	FirstPass int       `json:"first_pass"`
	Estimate  *Estimate `json:"estimate"`
}

func splitList(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}

func strp(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// planSHA is the commit whose plan describes the run's end: the last run
// commit, else the last task commit, else the plan commit.
func planSHA(o *runs.Owned) string {
	sha := o.Plan.SHA
	if n := len(o.Tasks); n > 0 {
		sha = o.Tasks[n-1].SHA
	}
	if o.Run != nil {
		sha = o.Run.SHA
	}
	return sha
}

// Build assembles the report of one brief from its run folders and commits.
// It returns nil when the brief has no runs. Nothing is written.
func Build(root, brief string, c *classify.Classifier) (*Report, error) {
	m, err := runs.Read(root, brief)
	if err != nil {
		return nil, err
	}
	if len(m.Folders) == 0 {
		return nil, nil
	}
	r := &Report{Schema: Schema, Brief: defect.BriefName(brief), RunID: m.RunID}
	var briefText string
	if b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(m.BriefPath))); err == nil {
		briefText = string(b)
		r.Status = runs.FrontmatterStatus(briefText)
	}
	var plan *runs.PlanDoc
	var size Size
	if o := m.Owned; o != nil {
		if plan, err = runs.PlanAt(root, o.Layout, planSHA(o)); err != nil {
			return nil, err
		}
		s, err := Measure(root, o, c)
		if err != nil {
			return nil, err
		}
		size = *s
	}
	// A release error only means there is no default branch: not merged.
	if sha, err := runs.Release(root, m.BriefPath); err == nil && sha != "" {
		r.Merged = &sha
		if m.Owned != nil {
			out, err := gitOut(root, "show", "-s", "--format=%ct", sha)
			if err != nil {
				return nil, err
			}
			if secs, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64); err == nil {
				ms := secs*1000 - m.Owned.Plan.Time.UnixMilli()
				r.LeadTime.PlanToMergeMS = &ms
			}
		}
	}

	u := Aggregate(m, size.Delivered.Added)
	t := CountTasks(m, plan, briefText)
	r.Tasks = ReportTasks{t.Planned, t.Done, t.Blocked, t.FirstPass, t.Estimate}
	r.Size = ReportSize{jsonLines(size.Delivered), jsonLines(size.Churn), size.TestCode(), size.Rework()}
	r.Time = ReportTime{u.Time.AgentMS, u.Time.WorkMS, u.Time.ReviewMS, u.Time.PlanMS, u.Time.GateMS, u.Time.WallMS}
	r.Rate = u.Rate
	r.Cost = ReportCost{u.Cost.Total, u.Cost.Plan, u.Cost.Work, u.Cost.Review, u.Cost.Per1kCode}
	r.Tokens = u.Tokens
	r.Models = PerPhase[[]string]{splitList(u.Models[PhasePlan]), splitList(u.Models[PhaseWork]), splitList(u.Models[PhaseReview])}
	r.Effort = PerPhase[*string]{strp(u.Effort[PhasePlan]), strp(u.Effort[PhaseWork]), strp(u.Effort[PhaseReview])}
	r.PermissionDenials = u.Denials
	r.Iterations = t.Iterations
	r.IterationsPerClosed = t.IterationsPerClosed
	r.Records.Missing = []MissingRecord{}
	for _, s := range u.Missing {
		f := strings.Fields(s) // `<task> <phase> (iteration <i>)`
		i, _ := strconv.Atoi(strings.TrimSuffix(f[len(f)-1], ")"))
		r.Records.Missing = append(r.Records.Missing, MissingRecord{f[0], f[1], i})
	}

	var x Matrix
	x.AddDerived(Derive(m, plan))
	recorded, err := defect.List(root, r.Brief)
	if err != nil {
		return nil, err
	}
	x.AddRecorded(recorded)
	dc := x.Counts()
	r.Defects = ReportDefects{InLoop: dc.InLoop, Operator: dc.Operator, Escaped: dc.Escaped, Total: dc.All()}
	r.Defects.RemovalEfficiency = fratio(float64(dc.InLoop+dc.Operator), float64(dc.All()))

	r.ByTask = byTask(m, plan, size)
	return r, nil
}

func byTask(m *runs.Model, plan *runs.PlanDoc, size Size) []ReportTask {
	rows := map[string]*ReportTask{}
	models := map[string]map[string]bool{}
	var order []string
	row := func(id string) *ReportTask {
		if rt, ok := rows[id]; ok {
			return rt
		}
		rt := &ReportTask{ID: id, Models: []string{}}
		rows[id] = rt
		order = append(order, id)
		return rt
	}
	if plan != nil {
		for _, pt := range plan.Tasks {
			rt := row(pt.ID)
			rt.Area, rt.Kind = strp(pt.Area), strp(pt.Kind)
		}
	}
	for _, tc := range size.ByTask {
		row(tc.Task).Churn = jsonLines(tc.Lines)
	}
	for _, f := range m.Folders {
		for _, it := range f.Iterations {
			row(it.Task).Attempts++
		}
		for _, s := range f.Sessions {
			if s.Phase != PhaseWork && s.Phase != PhaseReview {
				continue
			}
			if s.Task == "" {
				continue
			}
			rt := row(s.Task)
			rt.AgentMS += s.DurationMS
			rt.CostUSD += s.CostUSD
			if models[s.Task] == nil {
				models[s.Task] = map[string]bool{}
			}
			for name := range s.Models {
				models[s.Task][name] = true
			}
			if len(s.Models) == 0 && s.Model != "" {
				models[s.Task][s.Model] = true
			}
		}
	}
	out := make([]ReportTask, 0, len(order))
	for _, id := range order {
		rt := rows[id]
		rt.Models = splitList(joinSet(models[id]))
		out = append(out, *rt)
	}
	return out
}

// NewClassifier is the file classifier the repo's metrics.* config describes.
// The error is the config's, unwrapped.
func NewClassifier(root string) (*classify.Classifier, error) {
	get := func(k string) ([]string, error) {
		v, err := config.Get(root, k)
		return v.List, err
	}
	var repo classify.Preset
	for _, r := range []struct {
		key string
		dst *[]string
	}{{"metrics.excluded", &repo.Excluded}, {"metrics.test", &repo.Test}, {"metrics.docs", &repo.Docs}, {"metrics.code", &repo.Code}} {
		var err error
		if *r.dst, err = get(r.key); err != nil {
			return nil, err
		}
	}
	stacks, err := get("metrics.stacks")
	if err != nil {
		return nil, err
	}
	return classify.New(repo, stacks), nil
}
