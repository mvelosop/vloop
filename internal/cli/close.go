package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/classify"
	"github.com/mvelosop/vloop/internal/closing"
	"github.com/mvelosop/vloop/internal/defect"
	fmpkg "github.com/mvelosop/vloop/internal/frontmatter"
	"github.com/mvelosop/vloop/internal/gitrun"
	"github.com/mvelosop/vloop/internal/intervention"
	"github.com/mvelosop/vloop/internal/metrics"
	"github.com/mvelosop/vloop/internal/runs"
)

const closeFindingsMsg = `say what you found before merge: --finding "<summary>" (repeatable) or --no-findings`

// findingRefusal is what a finding without an origin prefix is refused with.
const findingRefusal = `a finding starts with whose defect it is: --finding "<brief|plan|work|env>: <summary>"`

// splitFinding splits "<origin>: <summary>" into its origin and summary; ok is
// false when it does not start with one of the origins in lower case.
func splitFinding(f string) (origin, summary string, ok bool) {
	for _, o := range intervention.Origins {
		if rest, found := strings.CutPrefix(strings.TrimSpace(f), o+": "); found && strings.TrimSpace(rest) != "" {
			return o, strings.TrimSpace(rest), true
		}
	}
	return "", "", false
}

func newBriefClose(g *Globals) *cobra.Command {
	var findings []string
	var none, dry bool
	var abandon string
	cmd := &cobra.Command{
		Use:   "close <brief>",
		Short: "Record findings, snapshot the metrics, write the run record, mark the brief consumed and commit",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if none == (len(findings) > 0) {
				return Usage(errors.New(closeFindingsMsg))
			}
			for _, f := range findings {
				if strings.TrimSpace(f) == "" {
					return Usage(errors.New("a finding must not be empty"))
				}
				if _, _, ok := splitFinding(f); !ok {
					return Usage(errors.New(findingRefusal))
				}
			}
			if cmd.Flags().Changed("abandon") && strings.TrimSpace(abandon) == "" {
				return Usage(errors.New("an abandon needs a reason: --abandon \"<reason>\""))
			}
			return runBriefClose(g, cmd, args[0], findings, closeOpts{abandon: cmd.Flags().Changed("abandon"), reason: strings.TrimSpace(abandon), dry: dry})
		},
	}
	cmd.Flags().StringArrayVar(&findings, "finding", nil, "\"<brief|plan|work|env>: <summary>\": a defect you found in the run and whose it is, recorded as a verification-finding intervention (repeatable)")
	cmd.Flags().BoolVar(&none, "no-findings", false, "state that you found nothing")
	cmd.Flags().StringVar(&abandon, "abandon", "", "close a brief whose plan did not complete, as abandoned, with the reason")
	cmd.Flags().BoolVar(&dry, "dry-run", false, "print what a close would record, write and commit, and write nothing")
	return cmd
}

type closeOpts struct {
	abandon bool
	reason  string
	dry     bool
}

func runGit(root string, args ...string) (string, error) {
	out, err := gitrun.At(root, args...).Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			return "", fmt.Errorf("git %s: %s", args[0], strings.TrimSpace(string(ee.Stderr)))
		}
		return "", fmt.Errorf("git %s: %w", args[0], err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}

func runBriefClose(g *Globals, cmd *cobra.Command, arg string, findings []string, opt closeOpts) error {
	out := cmd.OutOrStdout()
	root, err := g.gitRoot()
	if err != nil {
		return err
	}
	briefPath := runs.BriefPath(root, arg)
	name := defect.BriefName(briefPath)
	runID := runs.RunID(briefPath)

	m, err := runs.Read(root, briefPath)
	if err != nil {
		return Problem(err)
	}
	if len(m.Folders) == 0 {
		return Problem(fmt.Errorf("no runs for %s", name))
	}
	full := filepath.Join(root, filepath.FromSlash(briefPath))
	doc, err := closeRefusals(root, full, name)
	if err != nil {
		return err
	}
	c, err := newClassifier(g, out, root)
	if err != nil {
		return err
	}
	if !opt.abandon && (m.Owned == nil || m.Owned.Run == nil || m.Owned.Run.Outcome != "complete") {
		return incompletePlan(root, briefPath, c)
	}

	now := time.Now()
	status, statusText := "consumed", "consumed — closed "+now.Format("2006-01-02")+" as run "+runID+". **Do not re-plan from this brief.**"
	if opt.abandon {
		status, statusText = "abandoned", "abandoned — "+opt.reason
	}
	report, err := metrics.Build(root, briefPath, c)
	if err != nil {
		return Problem(err)
	}
	report.Status = status
	subject := "[vloop] close " + runID
	if opt.dry {
		return closeDryRun(out, root, briefPath, subject, findings, report, now)
	}
	files, err := recordFindings(root, name, findings, now)
	if err != nil {
		return err
	}
	if len(findings) > 0 {
		// The report above predates the intervention files; rebuild so it counts them.
		if report, err = metrics.Build(root, briefPath, c); err != nil {
			return Problem(err)
		}
		report.Status = status
	}
	snap, err := closing.WriteSnapshot(root, report)
	if err != nil {
		return Problem(err)
	}
	derived, err := metrics.DeriveBrief(root, briefPath)
	if err != nil {
		return Problem(err)
	}
	recorded, err := defect.List(root, name)
	if err != nil {
		return Problem(err)
	}
	ivs, err := intervention.List(root, name)
	if err != nil {
		return Problem(err)
	}
	rec := closing.Record{Name: name, Date: now, Report: report, Derived: derived, Recorded: recorded, Findings: ivs}
	if err := writeClosedBrief(full, briefPath, doc, status, statusText, rec); err != nil {
		return err
	}
	files = append(files, snap, briefPath)

	trailer := "Vloop-Brief: " + name
	sha, err := commitClose(root, files, subject, trailer)
	if err != nil {
		return err
	}

	if g.JSON {
		return encodeClosed(out, name, status, sha, files, trailer, report)
	}
	printClosed(out, report, files, sha, subject)
	every, err := defect.List(root, "")
	if err != nil {
		return Problem(err)
	}
	printDefectReport(out, runID, recorded, every)
	fmt.Fprintf(out, "squash-merge with the trailer: %s\n", trailer)
	return nil
}

// encodeClosed prints the close's --json result.
func encodeClosed(out io.Writer, name, status, sha string, files []string, trailer string, report *metrics.Report) error {
	return json.NewEncoder(out).Encode(struct {
		Brief   string          `json:"brief"`
		Status  string          `json:"status"`
		Commit  string          `json:"commit"`
		Files   []string        `json:"files"`
		Trailer string          `json:"trailer"`
		Metrics *metrics.Report `json:"metrics"`
	}{name, status, sha, files, trailer, report})
}

// recordFindings writes one verification-finding intervention per finding the
// operator named and returns their paths.
func recordFindings(root, name string, findings []string, now time.Time) ([]string, error) {
	var files []string
	for _, f := range findings {
		origin, summary, _ := splitFinding(f)
		p, err := intervention.Add(root, intervention.NewInput{Summary: summary, Brief: name, Phase: "verify", Kind: "verification-finding",
			Automatable: "partly", By: "operator", Origin: origin}, now)
		if err != nil {
			return nil, Problem(err)
		}
		files = append(files, p)
	}
	return files, nil
}

// closeRefusals reads the brief and refuses a close that cannot happen: the
// brief is already closed, HEAD is the default branch, or the tree is dirty. It
// returns the brief's text.
func closeRefusals(root, full, name string) (string, error) {
	raw, err := os.ReadFile(full)
	if err != nil {
		return "", Problem(err)
	}
	doc := string(raw)
	if st := runs.FrontmatterStatus(doc); st == "consumed" || st == "abandoned" {
		return "", Problem(fmt.Errorf("%s is already %s", name, st))
	}
	branch, _ := runGit(root, "symbolic-ref", "--quiet", "--short", "HEAD")
	if def, err := runs.DefaultBranch(root); err == nil && branch != "" &&
		branch == strings.TrimPrefix(def, "refs/remotes/origin/") {
		return "", Problem(fmt.Errorf("close on the brief's work branch, not %s", branch))
	}
	dirty, err := runGit(root, "status", "--porcelain", "--untracked-files=all")
	if err != nil {
		return "", Problem(err)
	}
	if dirty != "" {
		return "", Problem(errors.New("commit or stash your changes first — close makes one commit of its own"))
	}
	return doc, nil
}

// incompletePlan is the refusal for a plan that is not complete, saying how far
// it got.
func incompletePlan(root, briefPath string, c *classify.Classifier) error {
	done, planned := 0, 0
	if r, err := metrics.Build(root, briefPath, c); err == nil && r != nil {
		done, planned = r.Tasks.Done, r.Tasks.Planned
	}
	return Problem(fmt.Errorf(`the plan is not complete (%d/%d done) — finish it, or pass --abandon "<reason>"`, done, planned))
}

// writeClosedBrief rewrites the brief's status, its status line and its run
// record.
func writeClosedBrief(full, briefPath, doc, status, statusText string, rec closing.Record) error {
	fm, err := fmpkg.Parse(doc)
	if err != nil {
		return Problem(fmt.Errorf("%s: %w", briefPath, err))
	}
	fm.Replace("status", "status: "+status)
	body := rewriteStatusLine(fm.Body(), statusText)
	fm.SetBody(closing.PutRunRecord(body, closing.Render(rec)))
	if err := os.WriteFile(full, []byte(fm.String()), 0o644); err != nil {
		return Problem(err)
	}
	return nil
}

// commitClose commits the close's files with the brief trailer and returns the
// commit's sha.
func commitClose(root string, files []string, subject, trailer string) (string, error) {
	if _, err := runGit(root, append([]string{"add", "--"}, files...)...); err != nil {
		return "", Problem(err)
	}
	if _, err := runGit(root, append([]string{"commit", "-q", "-m", subject, "-m", trailer, "--"}, files...)...); err != nil {
		return "", Problem(err)
	}
	sha, err := runGit(root, "rev-parse", "HEAD")
	if err != nil {
		return "", Problem(err)
	}
	return sha, nil
}

// printClosed prints the close's summary and the files it wrote, the last two
// being the snapshot and the brief.
func printClosed(out io.Writer, report *metrics.Report, files []string, sha, subject string) {
	metrics.PrintSummary(out, report)
	for i, f := range files {
		verb := "recorded"
		switch {
		case i == len(files)-2:
			verb = "wrote"
		case i == len(files)-1:
			verb = "updated"
		}
		fmt.Fprintf(out, "%s %s\n", verb, f)
	}
	fmt.Fprintf(out, "committed %s %s\n", sha[:7], subject)
}

// statusLine matches the body line the brief template starts with; what
// precedes `**Status:**` is kept, and a `**Status:**` quoted mid-line is not it.
var statusLine = regexp.MustCompile(`^([ \t>*+-]*?)\*\*Status:\*\* ready to plan.*$`)

// rewriteStatusLine rewrites the first line statusLine matches and no other.
func rewriteStatusLine(body, statusText string) string {
	lines := strings.Split(body, "\n")
	for i, l := range lines {
		if statusLine.MatchString(l) {
			lines[i] = statusLine.ReplaceAllString(l, "${1}**Status:** "+statusText)
			break
		}
	}
	return strings.Join(lines, "\n")
}

// closeDryRun prints what the close would do, counting the pending findings as
// operator-caught defects in the summary, and touches nothing.
func closeDryRun(out io.Writer, root, briefPath, subject string, findings []string, report *metrics.Report, now time.Time) error {
	d := &report.Defects
	d.Operator += len(findings)
	d.Total += len(findings)
	if d.Total > 0 {
		v := float64(d.InLoop+d.Operator) / float64(d.Total)
		d.RemovalEfficiency = &v
	}
	metrics.PrintSummary(out, report)
	taken := map[string]bool{}
	var would []string
	for _, f := range findings {
		_, summary, _ := splitFinding(f)
		would = append(would, defect.StampedID(filepath.Join(root, filepath.FromSlash(intervention.Dir)), "I", defect.Slug(strings.Join(strings.Fields(summary), " ")), now, taken))
	}
	// The report sits before the would-lines: a dry run has no trailer line to
	// precede, and its tail stays what it was.
	brief, err := defect.List(root, defect.BriefName(briefPath))
	if err != nil {
		return Problem(err)
	}
	every, err := defect.List(root, "")
	if err != nil {
		return Problem(err)
	}
	printDefectReport(out, report.RunID, brief, every)
	for _, id := range would {
		fmt.Fprintf(out, "would record %s/%s.md\n", intervention.Dir, id)
	}
	fmt.Fprintf(out, "would write %s/%s.json\n", closing.SnapshotDir, report.RunID)
	fmt.Fprintf(out, "would update %s\n", briefPath)
	fmt.Fprintf(out, "would commit %s\n", subject)
	return nil
}

// printDefectReport prints the brief's defects with their rule, outcome or
// untriaged, and the recurrences of each rule across every brief (L12). Close
// reports; it never refuses for an untriaged defect.
func printDefectReport(out io.Writer, runID string, brief, every []defect.Defect) {
	if len(brief) == 0 {
		fmt.Fprintf(out, "defects of %s: none\n", runID)
		return
	}
	fmt.Fprintf(out, "defects of %s\n", runID)
	untriaged := 0
	for _, d := range brief {
		rule, outcome := "—", "untriaged"
		if d.Rule != "" {
			rule = d.Rule
		}
		if d.RuleOutcome != "" {
			outcome = d.RuleOutcome
		}
		if d.RuleOutcome == "" {
			untriaged++
		}
		line := fmt.Sprintf("  %s  %s  %s  %s  %s", d.ID, d.Kind, d.FoundBy, rule, outcome)
		if d.Rule != "" && d.RuleOutcome != "" {
			var ids []string
			n := 0
			for _, o := range every {
				if o.ID == d.ID || o.Rule != d.Rule {
					continue
				}
				n++
				if len(ids) < 3 {
					ids = append(ids, shortDefectID(o.ID))
				}
			}
			if n > 0 {
				line += fmt.Sprintf("  recurs ×%d (%s)", n, strings.Join(ids, ", "))
			}
		}
		fmt.Fprintln(out, line)
	}
	if untriaged > 0 {
		fmt.Fprintf(out, "%d untriaged: vloop defect set <id> rule '<where>: <which>' and rule-outcome fired|missed|absent|none\n", untriaged)
	}
}

// shortDefectID is D<stamp> of a defect ID: D<date>-<time>.
func shortDefectID(id string) string {
	p := strings.SplitN(id, "-", 3)
	if len(p) < 2 {
		return id
	}
	return p[0] + "-" + p[1]
}
