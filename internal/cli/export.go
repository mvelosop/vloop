package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/defect"
	"github.com/mvelosop/vloop/internal/intervention"
	"github.com/mvelosop/vloop/internal/metrics"
)

const exportSchema = "export/v1"

type exportRepo struct {
	Name   string  `json:"name"`
	Remote *string `json:"remote"`
	// Stacks is metrics.stacks as configured, [] when unset.
	Stacks []string `json:"stacks"`
}

type exportIntervention struct {
	Schema      string     `json:"schema"`
	Type        string     `json:"type"`
	Repo        exportRepo `json:"repo"`
	ID          string     `json:"id"`
	Brief       string     `json:"brief"`
	Phase       string     `json:"phase"`
	Kind        string     `json:"kind"`
	Automatable string     `json:"automatable"`
	By          string     `json:"by"`
	Occurred    string     `json:"occurred"`
	Recorded    string     `json:"recorded"`
	Backfilled  bool       `json:"backfilled,omitempty"`
}

type exportTask struct {
	Schema  string            `json:"schema"`
	Type    string            `json:"type"`
	Repo    exportRepo        `json:"repo"`
	Brief   string            `json:"brief"`
	RunID   string            `json:"run_id"`
	ID      string            `json:"id"`
	Title   string            `json:"title"`
	Area    *string           `json:"area"`
	Kind    *string           `json:"kind"`
	Status  string            `json:"status"`
	Attempt int               `json:"attempts"`
	Churn   metrics.JSONLines `json:"churn"`
	AgentMS int64             `json:"agent_ms"`
	CostUSD float64           `json:"cost_usd"`
	Models  []string          `json:"models"`
}

type exportDefect struct {
	Schema   string     `json:"schema"`
	Type     string     `json:"type"`
	Repo     exportRepo `json:"repo"`
	Brief    string     `json:"brief"`
	RunID    string     `json:"run_id"`
	ID       string     `json:"id"`
	Task     *string    `json:"task"`
	Origin   string     `json:"origin"`
	FoundBy  string     `json:"found_by"`
	Kind     string     `json:"kind"`
	Severity *string    `json:"severity"`
	Status   *string    `json:"status"`
	Summary  string     `json:"summary"`
	Derived  bool       `json:"derived"`
}

func newMetricsExport(g *Globals) *cobra.Command {
	var workspace string
	cmd := &cobra.Command{
		Use:   "export [<brief>…]",
		Short: "Print briefs, tasks, defects and interventions as JSON Lines (export/v1), with the repository's identity",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			if workspace != "" {
				return workspaceExport(g, out, cmd.ErrOrStderr(), workspace, args)
			}
			root, err := g.root()
			if err != nil {
				return err
			}
			s, err := exportRepoLines(g, out, root, "", args)
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(out, s)
			return err
		},
	}
	cmd.Flags().StringVar(&workspace, "workspace", "", "export every repository the workspace `file` lists")
	return cmd
}

// exportRepoLines is one repository's export as JSON Lines; a non-empty name
// replaces repo.name.
func exportRepoLines(g *Globals, out io.Writer, root, name string, args []string) (string, error) {
	c, err := newClassifier(g, out, root)
	if err != nil {
		return "", err
	}
	var reports []*metrics.Report
	for _, b := range args {
		r, err := metrics.Build(root, b, c)
		if err != nil {
			return "", Problem(err)
		}
		if r == nil {
			return "", Problem(fmt.Errorf("no runs for %s", defect.BriefName(b)))
		}
		reports = append(reports, r)
	}
	if len(args) == 0 {
		if reports, err = allReports(root, c); err != nil {
			return "", Problem(err)
		}
	}
	repo := repoIdentity(root)
	sv, err := config.Get(root, "metrics.stacks")
	if err != nil {
		return "", configErr(g, out, err)
	}
	repo.Stacks = append([]string{}, sv.List...)
	if name != "" {
		repo.Name = name
	}
	var buf strings.Builder
	enc := json.NewEncoder(&buf)
	for _, r := range reports {
		recs, err := exportRecords(root, repo, r)
		if err != nil {
			return "", Problem(err)
		}
		for _, rec := range recs {
			if err := enc.Encode(rec); err != nil {
				return "", Problem(err)
			}
		}
	}
	ivs, err := intervention.List(root, "")
	if err != nil {
		return "", Problem(err)
	}
	sort.Slice(ivs, func(i, j int) bool { return ivs[i].ID < ivs[j].ID })
	named := map[string]bool{}
	for _, b := range args {
		named[defect.BriefName(b)] = true
	}
	for _, v := range ivs {
		if len(args) > 0 && !named[defect.BriefName(v.Brief)] {
			continue
		}
		rec := exportIntervention{
			Schema: exportSchema, Type: "intervention", Repo: repo, ID: v.ID, Brief: v.Brief,
			Phase: v.Phase, Kind: v.Kind, Automatable: v.Automatable, By: v.By,
			Occurred: v.Occurred, Recorded: v.Recorded, Backfilled: v.Backfilled,
		}
		if err := enc.Encode(rec); err != nil {
			return "", Problem(err)
		}
	}
	return buf.String(), nil
}

// exportRecords is one brief's records in export order: the brief, its tasks in
// plan order, its derived defects by iteration, then its recorded defects by id.
func exportRecords(root string, repo exportRepo, r *metrics.Report) ([]any, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	doc := map[string]any{}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	delete(doc, "by_task")
	doc["schema"], doc["type"], doc["repo"] = exportSchema, "brief", repo
	recs := []any{doc}

	briefPath := defect.BriefsDir + "/" + r.Brief + ".md"
	plan, err := metrics.PlanOf(root, briefPath)
	if err != nil {
		return nil, err
	}
	rows := map[string]metrics.ReportTask{}
	var order []string
	for _, t := range r.ByTask {
		rows[t.ID] = t
	}
	titles, statuses := map[string]string{}, map[string]string{}
	if plan != nil {
		for _, pt := range plan.Tasks {
			titles[pt.ID], statuses[pt.ID] = pt.Title, pt.Status
			order = append(order, pt.ID)
		}
	}
	seen := map[string]bool{}
	for _, id := range order {
		seen[id] = true
	}
	for _, t := range r.ByTask {
		if !seen[t.ID] {
			order = append(order, t.ID)
		}
	}
	for _, id := range order {
		t := rows[id]
		models := t.Models
		if models == nil {
			models = []string{}
		}
		recs = append(recs, exportTask{
			Schema: exportSchema, Type: "task", Repo: repo, Brief: r.Brief, RunID: r.RunID,
			ID: id, Title: titles[id], Area: t.Area, Kind: t.Kind, Status: statuses[id],
			Attempt: t.Attempts, Churn: t.Churn, AgentMS: t.AgentMS, CostUSD: t.CostUSD, Models: models,
		})
	}

	derived, err := metrics.DeriveBrief(root, briefPath)
	if err != nil {
		return nil, err
	}
	for i, id := range metrics.DerivedIDs(r.RunID, derived) {
		d := derived[i]
		recs = append(recs, exportDefect{
			Schema: exportSchema, Type: "defect", Repo: repo, Brief: r.Brief, RunID: r.RunID,
			ID: id, Task: strPtr(d.Task), Origin: d.Origin, FoundBy: d.FoundBy, Kind: d.Kind,
			Summary: d.Summary, Derived: true,
		})
	}
	recorded, err := defect.List(root, r.Brief)
	if err != nil {
		return nil, err
	}
	for _, d := range recorded {
		recs = append(recs, exportDefect{
			Schema: exportSchema, Type: "defect", Repo: repo, Brief: r.Brief, RunID: r.RunID,
			ID: d.ID, Task: strPtr(d.Task), Origin: d.Origin, FoundBy: d.FoundBy, Kind: d.Kind,
			Severity: strPtr(d.Severity), Status: strPtr(d.Status), Summary: d.Summary,
		})
	}
	return recs, nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// repoIdentity names the repository by its origin remote, credentials removed,
// else by its directory. It reads local git config only.
func repoIdentity(root string) exportRepo {
	out, err := exec.Command("git", "-C", root, "config", "--get", "remote.origin.url").Output()
	url := strings.TrimSpace(string(out))
	if err != nil || url == "" {
		return exportRepo{Name: filepath.Base(root), Stacks: []string{}}
	}
	url = stripUserinfo(url)
	name := strings.TrimSuffix(strings.TrimRight(url, "/"), ".git")
	if i := strings.LastIndexAny(name, "/:"); i >= 0 {
		name = name[i+1:]
	}
	if name == "" {
		name = filepath.Base(root)
	}
	return exportRepo{Name: name, Remote: &url, Stacks: []string{}}
}

// stripUserinfo removes `user[:password]@` from a URL's authority. Forms with
// no scheme (`git@host:path`) carry no password and are left as they are.
func stripUserinfo(u string) string {
	i := strings.Index(u, "://")
	if i < 0 {
		return u
	}
	rest := u[i+3:]
	end := strings.IndexAny(rest, "/?#")
	if end < 0 {
		end = len(rest)
	}
	if at := strings.LastIndex(rest[:end], "@"); at >= 0 {
		rest = rest[at+1:]
	}
	return u[:i+3] + rest
}
