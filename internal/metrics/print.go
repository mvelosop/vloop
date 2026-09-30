package metrics

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"text/tabwriter"
)

// thousands writes n with `,` separators.
func thousands(n int) string {
	s := strconv.Itoa(n)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	if neg {
		s = "-" + s
	}
	return s
}

func minutes(ms int64) string { return fmt.Sprintf("%.1f", float64(ms)/60000) }

func optMinutes(ms *int64) string {
	if ms == nil {
		return "n/a"
	}
	return minutes(*ms)
}

func optFloat(f *float64, format string) string {
	if f == nil {
		return "n/a"
	}
	return fmt.Sprintf(format, *f)
}

func minSec(ms int64) string {
	s := int64(math.Round(float64(ms) / 1000))
	return fmt.Sprintf("%dm%02ds", s/60, s%60)
}

func dashPtr(s *string) string {
	if s == nil {
		return "-"
	}
	return *s
}

func joinModels(l []string) string { return dash(strings.Join(l, ",")) }

// efficiency is the removal efficiency as a whole percentage, or n/a.
func efficiency(d ReportDefects) string {
	c := DefectCounts{InLoop: d.InLoop, Operator: d.Operator, Escaped: d.Escaped}
	if p := c.RemovalEfficiency(); p != nil {
		return fmt.Sprintf("%d%%", *p)
	}
	return "n/a"
}

func PrintSummary(out io.Writer, r *Report) {
	merged := "not merged"
	if r.Merged != nil {
		merged = "merged " + (*r.Merged)[:min(7, len(*r.Merged))]
	}
	fmt.Fprintf(out, "%s  %s · %s\n", r.RunID, dash(r.Status), merged)
	planned := fmt.Sprintf("%d planned", r.Tasks.Planned)
	if e := r.Tasks.Estimate; e != nil {
		planned += fmt.Sprintf(" (brief said %d–%d)", e.Min, e.Max)
	}
	fmt.Fprintf(out, " tasks     %s · %d done · %d blocked · first-pass %d/%d\n",
		planned, r.Tasks.Done, r.Tasks.Blocked, r.Tasks.FirstPass, r.Tasks.Planned)
	line := func(prefix, name string, l JSONLines, tail string) {
		fmt.Fprintf(out, "%s%-10s code %s · test %s · docs %s · %s\n", prefix, name,
			thousands(l.Code), thousands(l.Test), thousands(l.Docs), tail)
	}
	line(" size      ", "delivered", r.Size.Delivered, "test:code "+optFloat(r.Size.TestCodeRatio, "%.2f"))
	line("           ", "churn", r.Size.Churn, "rework "+optFloat(r.Size.Rework, "%.2f"))
	t := r.Time
	fmt.Fprintf(out, " time      agent %s min (work %s · review %s) · plan %s min · gates %s · wall %s min\n",
		minutes(t.AgentMS), minutes(t.WorkMS), minutes(t.ReviewMS), minutes(t.PlanMS), optMinutes(t.GatesMS), optMinutes(t.WallMS))
	fmt.Fprintf(out, " rate      %s code lines/min · %s incl. tests\n",
		optFloat(r.Rate.CodePerMin, "%.1f"), optFloat(r.Rate.CodeTestPerMin, "%.1f"))
	per1k := "n/a"
	if r.Cost.Per1000Code != nil {
		per1k = fmt.Sprintf("$%.2f", *r.Cost.Per1000Code)
	}
	fmt.Fprintf(out, " cost      $%.2f · plan %.2f · work %.2f · review %.2f · %s per 1,000 code lines\n",
		r.Cost.Total, r.Cost.Plan, r.Cost.Work, r.Cost.Review, per1k)
	fmt.Fprintf(out, " models    plan %s · work %s · review %s\n",
		joinModels(r.Models.Plan), joinModels(r.Models.Work), joinModels(r.Models.Review))
	fmt.Fprintf(out, " defects   in-loop %d · operator %d · escaped %d · removal efficiency %s\n",
		r.Defects.InLoop, r.Defects.Operator, r.Defects.Escaped, efficiency(r.Defects))
	if n := len(r.Records.Missing); n > 0 {
		var l []string
		for _, m := range r.Records.Missing {
			l = append(l, fmt.Sprintf("%s %s (iteration %d)", m.Task, m.Phase, m.Iteration))
		}
		fmt.Fprintf(out, " records   %d session record(s) missing: %s — their time and cost are not counted\n", n, strings.Join(l, ", "))
	}
}

func PrintByTask(out io.Writer, r *Report) {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "id\tarea\tkind\tatt\tcode+\ttest+\tdocs+\tother+\tagent\tcost\tmodel")
	for _, t := range r.ByTask {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\t$%.2f\t%s\n", t.ID, dashPtr(t.Area), dashPtr(t.Kind), t.Attempts,
			thousands(t.Churn.Code), thousands(t.Churn.Test), thousands(t.Churn.Docs), thousands(t.Churn.Other),
			minSec(t.AgentMS), t.CostUSD, joinModels(t.Models))
	}
	w.Flush()
}

func PrintBriefTable(out io.Writer, rs []*Report) { printBriefTable(out, nil, rs) }

// PrintWorkspaceTable is PrintBriefTable with a leading repo column: repos[i]
// names the repository of rs[i].
func PrintWorkspaceTable(out io.Writer, repos []string, rs []*Report) {
	if repos == nil {
		repos = []string{}
	}
	printBriefTable(out, repos, rs)
}

func printBriefTable(out io.Writer, repos []string, rs []*Report) {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	lead, head := "", ""
	if repos != nil {
		head = "repo\t"
	}
	fmt.Fprintln(w, head+"brief\ttasks\tfirst-pass\tcode\ttest\tt:c\tagent\t$/1k\tin-loop\toperator\tescaped\tefficiency")
	for i, r := range rs {
		if repos != nil {
			lead = repos[i] + "\t"
		}
		per1k := "n/a"
		if r.Cost.Per1000Code != nil {
			per1k = fmt.Sprintf("%.2f", *r.Cost.Per1000Code)
		}
		fmt.Fprintf(w, "%s%s\t%d\t%d/%d\t%s\t%s\t%s\t%s\t%s\t%d\t%d\t%d\t%s\n", lead, r.RunID, r.Tasks.Planned,
			r.Tasks.FirstPass, r.Tasks.Planned, thousands(r.Size.Delivered.Code), thousands(r.Size.Delivered.Test),
			optFloat(r.Size.TestCodeRatio, "%.2f"), minutes(r.Time.AgentMS), per1k,
			r.Defects.InLoop, r.Defects.Operator, r.Defects.Escaped, efficiency(r.Defects))
	}
	w.Flush()
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
