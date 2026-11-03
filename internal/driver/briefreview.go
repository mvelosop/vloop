package driver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mvelosop/vloop/internal/brief"
	"github.com/mvelosop/vloop/internal/schema"
)

const (
	briefReviewsDir  = ".vloop/state/brief-reviews"
	briefReviewFile  = tmpDir + "/brief-review.json"
	briefReviewKept  = "brief-review.json"
	briefReviewStamp = "20060102-150405"
)

// BriefReview is `vloop brief review`: one review-brief session on a brief,
// outside any run. It holds the working tree's lock for the session, commits
// nothing and moves no ref.
type BriefReview struct {
	Root    string
	Version string
	Brief   string // repo-relative, with "/"
	Out     io.Writer
	Err     io.Writer
	Timeout time.Duration

	Claude string
	Now    func() time.Time
	Home   string
	User   string

	Judge CallJudge // the fence judge the CLI hands in, as Iterator's
}

type briefReviewReport struct {
	Brief    string `json:"brief"`
	Verdict  string `json:"verdict"`
	Findings []struct {
		Kind    string `json:"kind"`
		Section string `json:"section"`
		Summary string `json:"summary"`
	} `json:"findings"`
}

// Run reviews the brief and prints the verdict. It returns the verdict (PASS,
// FIX or ASK) of a valid report, or a Halt for everything that goes wrong.
func (b *BriefReview) Run() (string, error) {
	root := b.Root
	model, effort, err := ModelEffort(root, nil, PhaseReviewBrief)
	if err != nil {
		return "", halt(ExitPreflight, "%v", err)
	}
	agent, err := ResolveAgent(root, PhaseReviewBrief)
	if err != nil {
		return "", halt(ExitPreflight, "%v", err)
	}
	runID := brief.RunID(b.Brief)
	now := time.Now
	if b.Now != nil {
		now = b.Now
	}

	release, err := AcquireLock(root, CurrentBranch(root), runID, now(),
		func(f string, a ...any) { fmt.Fprintf(b.Err, "warning: "+f+"\n", a...) })
	if err != nil {
		return "", &Halt{Code: ExitPreflight, Err: err}
	}
	defer release()
	ClearTmp(root, func(entry string, err error) {
		fmt.Fprintf(b.Err, "warning: cannot clear .vloop/tmp/%s: %v\n", entry, err)
	})

	briefAbs := filepath.Join(root, filepath.FromSlash(b.Brief))
	orig, err := os.ReadFile(briefAbs)
	if err != nil {
		return "", halt(ExitPreflight, "no such file: %s", b.Brief)
	}
	mode := os.FileMode(0o644)
	if fi, err := os.Stat(briefAbs); err == nil {
		mode = fi.Mode().Perm()
	}

	dir, err := newBriefReviewDir(root, runID, now())
	if err != nil {
		return "", err
	}
	rel := strings.TrimSuffix(relRunDir(root, dir), "/")
	r := &Runner{Root: root, Version: b.Version, RunID: runID, RunDir: dir,
		Claude: b.Claude, Now: b.Now, Home: b.Home, User: b.User, Timeout: b.Timeout, Judge: b.Judge}

	fmt.Fprintf(b.Err, "reviewing %s using %s (agent %s)\n", b.Brief, model, agent)
	before := refsState(root)
	guard := snapshotGit(root)
	res, err := r.Run(Spec{Agent: agent, Phase: PhaseReviewBrief, Arg: b.Brief, Model: model, Effort: effort})
	if err != nil {
		return "", halt(ExitSessionError, "brief review session failed: %v", err)
	}

	changed := false
	var restoreErr error
	if now, err := os.ReadFile(briefAbs); err != nil || !bytes.Equal(now, orig) {
		changed = true
		restoreErr = os.WriteFile(briefAbs, orig, mode)
	}
	if h := res.backstop(PhaseReviewBrief, func(f string, a ...any) { fmt.Fprintln(b.Err, r.Mask(fmt.Sprintf(f, a...))) }); h != nil {
		if restoreErr != nil {
			fmt.Fprintln(b.Err, r.Mask(restoreErr.Error()))
		}
		return "", h
	}
	if restoreErr != nil {
		return "", restoreErr
	}
	if what := guard.changed(); what != "" {
		return "", halt(ExitRefsMoved, "the brief review changed %s — nothing was committed; restore it, then re-run", what)
	}
	if len(refsDiff(before, refsState(root))) > 0 {
		return "", halt(ExitRefsMoved, "REFS MOVED review-brief — the brief review changed git refs; nothing was committed")
	}
	if changed {
		return "", halt(ExitRefsMoved, "the brief review changed %s; it was restored", b.Brief)
	}
	if res.TimedOut {
		return "", halt(ExitSessionError, "brief review session timed out after %s — see %s", b.Timeout, rel)
	}
	if res.ExitCode != 0 {
		return "", halt(ExitSessionError, "brief review session failed (%s exited %d) — see %s", agent, res.ExitCode, rel)
	}
	data, rerr := readHandoff(filepath.Join(root, filepath.FromSlash(briefReviewFile)))
	if rerr != nil {
		return "", halt(ExitSessionError, "the brief review left no %s — see %s", briefReviewFile, rel)
	}
	var rep briefReviewReport
	viol, verr := schema.Validate("brief-review/v1", data)
	if verr != nil || len(viol) > 0 || json.Unmarshal(data, &rep) != nil || rep.Brief != strings.TrimSuffix(filepath.Base(b.Brief), ".md") {
		return "", halt(ExitSessionError, "the brief review's %s does not follow brief-review/v1 — see %s", briefReviewFile, rel)
	}
	masked := r.Mask(string(data))
	if err := os.WriteFile(filepath.Join(dir, briefReviewKept), []byte(masked+"\n"), 0o644); err != nil {
		return "", err
	}
	// The report as kept is what is printed.
	_ = json.Unmarshal([]byte(masked), &rep)
	fmt.Fprintf(b.Out, "verdict: %s — %d finding(s)\n", rep.Verdict, len(rep.Findings))
	for _, f := range rep.Findings {
		fmt.Fprintf(b.Out, "  %-12s  %s: %s\n", f.Kind, f.Section, f.Summary)
	}
	fmt.Fprintf(b.Out, "kept: %s/%s\n", rel, briefReviewKept)
	return rep.Verdict, nil
}

// newBriefReviewDir creates .vloop/state/brief-reviews/<run id>/<YYYYMMDD-HHMMSS>/
// (UTC) with its sessions/ folder, suffixing -2, -3… when the second is taken.
func newBriefReviewDir(root, runID string, now time.Time) (string, error) {
	base := filepath.Join(root, filepath.FromSlash(briefReviewsDir), runID)
	stamp := now.UTC().Format(briefReviewStamp)
	dir := filepath.Join(base, stamp)
	for n := 2; ; n++ {
		if _, err := os.Stat(dir); err != nil {
			break
		}
		dir = filepath.Join(base, fmt.Sprintf("%s-%d", stamp, n))
	}
	return dir, os.MkdirAll(filepath.Join(dir, "sessions"), 0o755)
}
