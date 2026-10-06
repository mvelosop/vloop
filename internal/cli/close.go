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
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/closing"
	"github.com/mvelosop/vloop/internal/defect"
	fmpkg "github.com/mvelosop/vloop/internal/frontmatter"
	"github.com/mvelosop/vloop/internal/metrics"
	"github.com/mvelosop/vloop/internal/runs"
)

const closeFindingsMsg = `say what you found before merge: --finding "<summary>" (repeatable) or --no-findings`

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
			}
			if cmd.Flags().Changed("abandon") && strings.TrimSpace(abandon) == "" {
				return Usage(errors.New("an abandon needs a reason: --abandon \"<reason>\""))
			}
			return runBriefClose(g, cmd, args[0], findings, closeOpts{abandon: cmd.Flags().Changed("abandon"), reason: strings.TrimSpace(abandon), dry: dry})
		},
	}
	cmd.Flags().StringArrayVar(&findings, "finding", nil, "a defect you found in the run, recorded as an operator defect (repeatable)")
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
	out, err := exec.Command("git", append([]string{"-C", root}, args...)...).Output()
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
	raw, err := os.ReadFile(full)
	if err != nil {
		return Problem(err)
	}
	doc := string(raw)
	if st := runs.FrontmatterStatus(doc); st == "consumed" || st == "abandoned" {
		return Problem(fmt.Errorf("%s is already %s", name, st))
	}
	branch, _ := runGit(root, "symbolic-ref", "--quiet", "--short", "HEAD")
	if def, err := runs.DefaultBranch(root); err == nil && branch != "" &&
		branch == strings.TrimPrefix(def, "refs/remotes/origin/") {
		return Problem(fmt.Errorf("close on the brief's work branch, not %s", branch))
	}
	dirty, err := runGit(root, "status", "--porcelain", "--untracked-files=all")
	if err != nil {
		return Problem(err)
	}
	if dirty != "" {
		return Problem(errors.New("commit or stash your changes first — close makes one commit of its own"))
	}
	c, err := newClassifier(g, out, root)
	if err != nil {
		return err
	}
	if !opt.abandon && (m.Owned == nil || m.Owned.Run == nil || m.Owned.Run.Outcome != "complete") {
		done, planned := 0, 0
		if r, err := metrics.Build(root, briefPath, c); err == nil && r != nil {
			done, planned = r.Tasks.Done, r.Tasks.Planned
		}
		return Problem(fmt.Errorf(`the plan is not complete (%d/%d done) — finish it, or pass --abandon "<reason>"`, done, planned))
	}

	now := time.Now()
	status, statusText := "consumed", "consumed — closed "+now.Format("2006-01-02")+" as run "+runID+". **Do not re-plan from this brief.**"
	if opt.abandon {
		status, statusText = "abandoned", "abandoned — "+opt.reason
	}
	var files []string
	report, err := metrics.Build(root, briefPath, c)
	if err != nil {
		return Problem(err)
	}
	report.Status = status
	subject := "[vloop] close " + runID
	if opt.dry {
		return closeDryRun(out, root, briefPath, subject, findings, report, now)
	}
	for _, f := range findings {
		p, err := defect.Add(root, defect.NewInput{Summary: f, FoundBy: "operator", Brief: name, Origin: "work", Kind: "bug", Severity: "medium"}, now)
		if err != nil {
			return Problem(err)
		}
		files = append(files, p)
	}
	if len(findings) > 0 {
		// The report above predates the defect files; rebuild so it counts them.
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
	fm, err := fmpkg.Parse(doc)
	if err != nil {
		return Problem(fmt.Errorf("%s: %w", briefPath, err))
	}
	fm.Replace("status", "status: "+status)
	body := rewriteStatusLine(fm.Body(), statusText)
	fm.SetBody(closing.PutRunRecord(body, closing.Render(closing.Record{Name: name, Date: now, Report: report, Derived: derived, Recorded: recorded})))
	if err := os.WriteFile(full, []byte(fm.String()), 0o644); err != nil {
		return Problem(err)
	}
	files = append(files, snap, briefPath)

	trailer := "Vloop-Brief: " + name
	if _, err := runGit(root, append([]string{"add", "--"}, files...)...); err != nil {
		return Problem(err)
	}
	if _, err := runGit(root, append([]string{"commit", "-q", "-m", subject, "-m", trailer, "--"}, files...)...); err != nil {
		return Problem(err)
	}
	sha, err := runGit(root, "rev-parse", "HEAD")
	if err != nil {
		return Problem(err)
	}

	if g.JSON {
		return json.NewEncoder(out).Encode(struct {
			Brief   string          `json:"brief"`
			Status  string          `json:"status"`
			Commit  string          `json:"commit"`
			Files   []string        `json:"files"`
			Trailer string          `json:"trailer"`
			Metrics *metrics.Report `json:"metrics"`
		}{name, status, sha, files, trailer, report})
	}
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
	fmt.Fprintf(out, "squash-merge with the trailer: %s\n", trailer)
	return nil
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
// operator defects in the summary, and touches nothing.
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
	for _, f := range findings {
		base := "D" + now.Format("20060102-1504") + "-" + defect.Slug(strings.Join(strings.Fields(f), " "))
		id := base
		for n := 2; ; n++ {
			_, err := os.Stat(filepath.Join(root, filepath.FromSlash(defect.Dir), id+".md"))
			if err != nil && !taken[id] {
				break
			}
			id = base + "-" + strconv.Itoa(n)
		}
		taken[id] = true
		fmt.Fprintf(out, "would record %s/%s.md\n", defect.Dir, id)
	}
	fmt.Fprintf(out, "would write %s/%s.json\n", closing.SnapshotDir, report.RunID)
	fmt.Fprintf(out, "would update %s\n", briefPath)
	fmt.Fprintf(out, "would commit %s\n", subject)
	return nil
}
