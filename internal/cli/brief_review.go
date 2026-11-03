package cli

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"fmt"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/brief"
	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/driver"
)

func newBriefReview(b Build, g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "review <path>",
		Short: "Review a brief in a fresh session and print the verdict",
		Long: `Run /vloop:review-brief on one loop brief of any status as a single fresh
session on the agent agent.review-brief names, fenced like a review session.
The brief is never changed: a session that edits it is refused and the brief is
restored.

The session's report is validated and kept under
.vloop/state/brief-reviews/<run id>/<YYYYMMDD-HHMMSS>/ with the session's
record; the verdict and its findings are printed on stdout. The command takes
the working tree's lock for the session, so it cannot run beside vloop run.

Exit codes: 0 PASS; 1 FIX or ASK, or a refusal before the session; 2 usage;
7 the session failed or left no valid report; 9 the session changed the brief,
git refs or the repository configuration, or a Copilot session got past its fence.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error { return runBriefReview(b, g, cmd, args[0]) },
	}
}

func runBriefReview(b Build, g *Globals, cmd *cobra.Command, arg string) error {
	root, err := g.root()
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(root, ".vloop")); err != nil {
		return Problem(errors.New("not in a vloop repository — run vloop init first"))
	}
	// The command's scratch folder, where the lock lives, exists from here on.
	if err := os.MkdirAll(filepath.Join(root, ".vloop", "tmp"), 0o755); err != nil {
		return Problem(err)
	}
	start := g.Dir
	if start == "" {
		start, _ = os.Getwd()
	}
	abs := arg
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(start, abs)
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		rel = abs
	}
	rel = path.Clean(filepath.ToSlash(rel))
	fi, err := os.Stat(abs)
	if err != nil || fi.IsDir() {
		return Problem(fmt.Errorf("no such file: %s", rel))
	}
	if !strings.HasSuffix(rel, brief.Suffix) {
		return Usage(fmt.Errorf("not a loop brief: %s", rel))
	}

	budget, err := driver.ResolveBudgets(root, nil)
	if err != nil {
		return configErr(g, cmd.OutOrStdout(), err)
	}
	if _, err := config.Get(root, "run.git-config-ignore"); err != nil {
		return configErr(g, cmd.OutOrStdout(), err)
	}
	if err := briefReviewPreflight(root, cmd); err != nil {
		return err
	}

	br := &driver.BriefReview{Root: root, Version: b.Version, Brief: rel,
		Out: cmd.OutOrStdout(), Err: cmd.ErrOrStderr(),
		Timeout: time.Duration(budget.SessionTimeout) * time.Minute, Judge: CallJudge(root)}
	verdict, err := br.Run()
	if err != nil {
		var h *driver.Halt
		if errors.As(err, &h) && h.Code != ExitProblems {
			return &ExitError{Code: h.Code, Err: err}
		}
		return Problem(err)
	}
	if verdict != "PASS" {
		return &ExitError{Code: 1, Err: errors.New("")}
	}
	return nil
}

// briefReviewPreflight checks only what one session of the agent needs: for
// Claude, claude on PATH and the folder trusted.
func briefReviewPreflight(root string, cmd *cobra.Command) error {
	agent, err := driver.ResolveAgent(root, driver.PhaseReviewBrief)
	if err != nil {
		return Problem(err)
	}
	if agent == driver.AgentCopilot {
		return briefReviewCopilotPreflight(root, cmd)
	}
	errOut := cmd.ErrOrStderr()
	bad := 0
	if _, problem := driver.AgentPreflight(driver.AgentClaude); problem != "" {
		fmt.Fprintf(errOut, "  ✗ claude %s\n", problem)
		bad++
	}
	if !doctorTrusted(root) {
		fmt.Fprintf(errOut, "  ✗ trust this folder is not marked trusted in the Claude settings — run claude here once and accept the trust dialog\n")
		bad++
	}
	if bad > 0 {
		return Problem(fmt.Errorf("preflight failed — %d problem(s), nothing has run", bad))
	}
	return nil
}

// briefReviewCopilotPreflight refuses, in order: copilot missing, a project
// skill with the name of a plugin skill, no GitHub login, the default model.
// Nothing has run when it refuses; when it passes it gives the experimental
// warning.
func briefReviewCopilotPreflight(root string, cmd *cobra.Command) error {
	const key = "agent.review-brief"
	if _, problem := driver.AgentPreflight(driver.AgentCopilot); problem != "" {
		fmt.Fprintf(cmd.ErrOrStderr(), "  ✗ copilot %s\n", problem)
		return Problem(errors.New("preflight failed — 1 problem(s), nothing has run"))
	}
	if cl := skillClashes(root); len(cl) > 0 {
		return Problem(fmt.Errorf("%s %s", cl[0].path, cl[0].message()))
	}
	if !copilotLoggedIn() {
		return Problem(fmt.Errorf("%s is copilot but no GitHub login reaches its sessions: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login", key))
	}
	if m, err := config.Get(root, "model.review-brief"); err == nil && m.Source == config.SourceDefault {
		return Problem(fmt.Errorf("%s is copilot but model.review-brief is the default %q, a Claude alias: set model.review-brief to a Copilot model id", key, m.Value))
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "  ! %s %s\n", key, copilotExperimental)
	return nil
}
