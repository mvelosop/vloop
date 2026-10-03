package cli

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mvelosop/vloop/internal/brief"
	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/driver"
	"github.com/mvelosop/vloop/internal/state"
)

// runBudgetFlags maps each budget flag to the config key it overrides.
var runBudgetFlags = []struct{ flag, key, usage string }{
	{"max-iterations", "run.max-iterations", "stop after `N` iterations in this run (0 or more)"},
	{"cost-ceiling", "run.cost-ceiling", "stop once the run's sessions cost `USD` dollars"},
	{"max-attempts", "run.max-attempts", "block a task after `N` attempts (1 or more)"},
	{"stall-limit", "run.stall-limit", "stop after `N` iterations in a row that close nothing (1 or more)"},
}

func newRun(b Build, g *Globals) *cobra.Command {
	var planOnly, replan bool
	budgets := make([]string, len(runBudgetFlags))
	cmd := &cobra.Command{
		Use:   "run [<brief>]",
		Short: "Plan a brief into tasks, on a work branch, and commit the plan",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := g.root()
			if err != nil {
				return err
			}
			if p, err := state.Load(root); err == nil && p.Schema == state.SchemaV1 {
				return Problem(fmt.Errorf("%s is a %s plan — finish it with vloop 1.x or re-plan the brief", state.FilePath, state.SchemaV1))
			}
			over := map[string]string{}
			for i, f := range runBudgetFlags {
				if !cmd.Flags().Changed(f.flag) {
					continue
				}
				if err := config.Check(root, f.key, budgets[i]); err != nil {
					return fmt.Errorf("--%s: %v", f.flag, err)
				}
				over[f.key] = budgets[i]
			}
			budget, err := driver.ResolveBudgets(root, over)
			if err != nil {
				return configErr(g, cmd.OutOrStdout(), err)
			}

			briefPath := ""
			if len(args) == 1 {
				if briefPath, err = runBriefPath(g, root, args[0]); err != nil {
					return err
				}
			}
			if err := runPreflight(b, root, cmd); err != nil {
				return err
			}

			if err := driver.CheckRunID(root, briefPath); err != nil {
				return Problem(err)
			}

			lockRun := ""
			if briefPath != "" {
				lockRun = brief.RunID(briefPath)
			}
			errOut := cmd.ErrOrStderr()
			release, err := driver.AcquireLock(root, driver.CurrentBranch(root), lockRun, time.Now(),
				func(f string, a ...any) { fmt.Fprintf(errOut, "warning: "+f+"\n", a...) })
			if err != nil {
				return Problem(err)
			}
			defer release()

			awakeWarn := ""
			if budget.KeepAwake {
				if err := driver.KeepAwake(); err != nil {
					awakeWarn = err.Error()
				}
			}

			pl := &driver.Planner{Root: root, Version: b.Version, Brief: briefPath,
				PlanOnly: planOnly, Replan: replan,
				Out: cmd.OutOrStdout(), Err: cmd.ErrOrStderr(), Quiet: g.Quiet, AwakeWarn: awakeWarn,
				SessionTimeout: time.Duration(budget.SessionTimeout) * time.Minute}
			res, err := pl.Plan()
			if err != nil {
				var h *driver.Halt
				if errors.As(err, &h) && h.Code != ExitProblems {
					return &ExitError{Code: h.Code, Err: err}
				}
				return Problem(err)
			}
			if planOnly {
				return nil
			}
			it := &driver.Iterator{Root: root, Version: b.Version, RunDir: res.RunDir, Budgets: budget,
				Out: cmd.OutOrStdout(), Err: cmd.ErrOrStderr(), Quiet: g.Quiet, AwakeWarn: awakeWarn}
			end, err := it.Run()
			if err != nil {
				var h *driver.Halt
				if errors.As(err, &h) && h.Code != ExitProblems {
					return &ExitError{Code: h.Code, Err: err}
				}
				return Problem(err)
			}
			if end.Code != 0 {
				return &ExitError{Code: end.Code, Err: errors.New("")}
			}
			return nil
		},
	}
	fs := cmd.Flags()
	fs.BoolVar(&planOnly, "plan-only", false, "stop after the plan is committed")
	fs.BoolVar(&replan, "replan", false, "plan a brief again even though its journal exists")
	for i, f := range runBudgetFlags {
		fs.StringVar(&budgets[i], f.flag, "", f.usage+" (default: config key "+f.key+")")
	}
	return cmd
}

// runBriefPath turns the brief argument into its repo-relative path with "/",
// refusing one that does not exist as a usage error: a typo must change nothing.
func runBriefPath(g *Globals, root, arg string) (string, error) {
	start := g.Dir
	if start == "" {
		start, _ = os.Getwd()
	}
	abs := arg
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(start, abs)
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("brief not found: %s", arg)
	}
	rel = path.Clean(filepath.ToSlash(rel))
	if fi, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil || fi.IsDir() {
		return "", fmt.Errorf("brief not found: %s", rel)
	}
	return rel, nil
}

// runPreflight is doctor's checks, refusing on any problem. Two things doctor
// only warns about stop a run: an untrusted workspace, whose settings Claude
// would silently ignore, and — as a warning here — an active pre-commit hook.
// The branch check is not run's business: it creates the work branch itself;
// nor is gate scratch, which the planner refuses in its own words.
func runPreflight(b Build, root string, cmd *cobra.Command) error {
	errOut := cmd.ErrOrStderr()
	bad := 0
	for _, c := range runDoctor(b, root) {
		switch {
		case c.Name == "branch", c.Name == "gate scratch": // the planner refuses an un-ignored scratch folder itself
		case c.Result == resProblem, c.Name == "trust" && c.Result == resWarning:
			fmt.Fprintf(errOut, "  ✗ %s %s\n", c.Name, c.Message)
			bad++
		}
	}
	if hook := preCommitHook(root); hook != "" {
		fmt.Fprintf(errOut, "  ! a pre-commit hook is active (%s) — the driver's commits do not run repository hooks; put what it checks in the gates\n", hook)
	}
	if bad > 0 {
		return Problem(fmt.Errorf("preflight failed — %d problem(s), nothing has run", bad))
	}
	return nil
}

// preCommitHook is the executable pre-commit hook git would run, repo-relative
// when it is inside the repository; "" when there is none.
func preCommitHook(root string) string {
	dir, err := gitOut(root, "config", "core.hooksPath")
	if err != nil || dir == "" {
		dir = ".git/hooks"
	}
	full := dir
	if !filepath.IsAbs(full) {
		full = filepath.Join(root, dir)
	}
	hook := filepath.Join(full, "pre-commit")
	fi, err := os.Stat(hook)
	if err != nil || fi.IsDir() || fi.Mode()&0o111 == 0 {
		return ""
	}
	if rel, err := filepath.Rel(root, hook); err == nil && !strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(rel)
	}
	return "outside the repository"
}
