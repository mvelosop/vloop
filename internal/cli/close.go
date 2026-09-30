package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/closing"
	"github.com/mvelosop/vloop/internal/defect"
	"github.com/mvelosop/vloop/internal/metrics"
	"github.com/mvelosop/vloop/internal/runs"
)

const closeFindingsMsg = `say what you found before merge: --finding "<summary>" (repeatable) or --no-findings`

func newBriefClose(g *Globals) *cobra.Command {
	var findings []string
	var none bool
	cmd := &cobra.Command{
		Use:   "close <brief>",
		Short: "Record findings, snapshot the metrics, write the run record, mark the brief consumed and commit",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if none == (len(findings) > 0) {
				return errors.New(closeFindingsMsg)
			}
			for _, f := range findings {
				if strings.TrimSpace(f) == "" {
					return errors.New("a finding must not be empty")
				}
			}
			return runBriefClose(g, cmd, args[0], findings)
		},
	}
	cmd.Flags().StringArrayVar(&findings, "finding", nil, "a defect you found in the run, recorded as an operator defect (repeatable)")
	cmd.Flags().BoolVar(&none, "no-findings", false, "state that you found nothing")
	return cmd
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

func runBriefClose(g *Globals, cmd *cobra.Command, arg string, findings []string) error {
	out := cmd.OutOrStdout()
	root, err := g.root()
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
	if m.Owned == nil || m.Owned.Run == nil || m.Owned.Run.Outcome != "complete" {
		done, planned := 0, 0
		if r, err := metrics.Build(root, briefPath, c); err == nil && r != nil {
			done, planned = r.Tasks.Done, r.Tasks.Planned
		}
		return Problem(fmt.Errorf(`the plan is not complete (%d/%d done) — finish it, or pass --abandon "<reason>"`, done, planned))
	}

	now := time.Now()
	var files []string
	for _, f := range findings {
		p, err := defect.Add(root, defect.NewInput{Summary: f, FoundBy: "operator", Brief: name, Origin: "work", Kind: "bug", Severity: "medium"}, now)
		if err != nil {
			return Problem(err)
		}
		files = append(files, p)
	}
	report, err := metrics.Build(root, briefPath, c)
	if err != nil {
		return Problem(err)
	}
	report.Status = "consumed"
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
	date := now.Format("2006-01-02")
	doc = setFrontmatterStatus(doc, "consumed")
	doc = statusLine.ReplaceAllString(doc, "${1}**Status:** consumed — closed "+date+" as run "+runID+". **Do not re-plan from this brief.**")
	doc = closing.PutRunRecord(doc, closing.Render(closing.Record{Name: name, Date: now, Report: report, Derived: derived, Recorded: recorded}))
	if err := os.WriteFile(full, []byte(doc), 0o644); err != nil {
		return Problem(err)
	}
	files = append(files, snap, briefPath)

	subject := "[vloop] close " + runID
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
		}{name, "consumed", sha, files, trailer, report})
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

// statusLine matches the body line the brief template starts with; only its
// first occurrence is rewritten, and what precedes `**Status:**` is kept.
var statusLine = regexp.MustCompile(`(?m)^([^\n]*?)\*\*Status:\*\* ready to plan[^\n]*`)

// setFrontmatterStatus sets `status:` in the leading frontmatter block,
// leaving every other byte alone.
func setFrontmatterStatus(doc, status string) string {
	off := 0
	first := true
	for off < len(doc) {
		end := len(doc)
		if i := strings.IndexByte(doc[off:], '\n'); i >= 0 {
			end = off + i
		}
		line := strings.TrimRight(doc[off:end], " \t\r")
		switch {
		case first:
			if line != "---" {
				return doc
			}
			first = false
		case line == "---":
			return doc
		case strings.HasPrefix(line, "status:"):
			cr := ""
			if strings.HasSuffix(doc[off:end], "\r") {
				cr = "\r"
			}
			return doc[:off] + "status: " + status + cr + doc[end:]
		}
		off = end + 1
	}
	return doc
}
