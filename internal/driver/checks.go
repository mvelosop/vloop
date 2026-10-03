package driver

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/state"
)

// checksDir is the run folder's folder of check logs.
const checksDir = "checks"

// planChecks is the config's [[check]] tables as the plan carries them.
func planChecks(defs []config.CheckDef) []state.PlanCheck {
	out := make([]state.PlanCheck, 0, len(defs))
	for _, d := range defs {
		paths := d.Paths
		if paths == nil {
			paths = []string{}
		}
		out = append(out, state.PlanCheck{Name: d.Name, Paths: paths, Run: d.Run})
	}
	return out
}

func sameChecks(a, b []state.PlanCheck) bool {
	return slices.EqualFunc(a, b, func(x, y state.PlanCheck) bool {
		return x.Name == y.Name && x.Run == y.Run && slices.Equal(x.Paths, y.Paths)
	})
}

// checkResult is one run of one check.
type checkResult struct {
	exit     int
	timedOut bool
	log      []byte
	ms       int64
}

// runCheck runs a check's command from the repo root in the plan's shell, under
// the gate timeout, and returns its output. A check that moves a ref or changes
// .git/config or the hooks is a *Halt with exit 9, as a gate is.
func runCheck(root, shell string, c state.PlanCheck, timeout time.Duration) (*checkResult, error) {
	var env []string
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "VLOOP_ACTIVE_TASK=") || strings.HasPrefix(kv, "VLOOP_GATE_TASK=") {
			continue
		}
		env = append(env, kv)
	}
	cmd, err := state.GateCommand(root, shell, c.Run, env)
	if err != nil {
		return nil, halt(ExitPreflight, "%v", err)
	}
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	refsBefore := refsState(root)
	gguard := snapshotGit(root)
	start := time.Now()
	timedOut, runErr := state.RunGroup(cmd, timeout)
	d := time.Since(start)
	if what := gguard.changed(); what != "" {
		return nil, halt(ExitRefsMoved, "the check %s changed %s — nothing was committed; restore it, then re-run", c.Name, what)
	}
	if len(refsDiff(refsBefore, refsState(root))) > 0 {
		return nil, halt(ExitRefsMoved, "the check %s moved git refs — nothing was committed; restore them, then re-run", c.Name)
	}
	res := &checkResult{ms: d.Milliseconds(), timedOut: timedOut}
	if runErr != nil {
		var ee *exec.ExitError
		if !errors.As(runErr, &ee) {
			return nil, fmt.Errorf("cannot run the check %s: %w", c.Name, runErr)
		}
		if res.exit = ee.ExitCode(); res.exit <= 0 {
			res.exit = 1
		}
	}
	if timedOut {
		if res.exit == 0 {
			res.exit = 1
		}
		if out.Len() > 0 && !bytes.HasSuffix(out.Bytes(), []byte("\n")) {
			out.WriteByte('\n')
		}
		fmt.Fprintf(&out, "vloop: check %s timed out after %d min\n", c.Name, int(timeout/time.Minute))
	}
	res.log = out.Bytes()
	return res, nil
}

// baseChecks runs every check, in order, on the base before the plan session. A
// check that fails refuses the run: it would be blamed on the first task.
func (t term) baseChecks(checks []state.PlanCheck, shell string, timeout time.Duration) error {
	root := t.p.Root
	dir := filepath.Join(t.r.RunDir, checksDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, c := range checks {
		t.say("checking the base: %s", c.Name)
		res, err := runCheck(root, shell, c, timeout)
		if err != nil {
			return err
		}
		logPath := filepath.Join(dir, "base-"+c.Name+".log")
		if err := os.WriteFile(logPath, []byte(t.r.Mask(string(res.log))), 0o644); err != nil {
			return err
		}
		if res.exit != 0 {
			return halt(ExitPreflight, "check %s fails on the base — fix it before planning: %s", c.Name,
				relRunDir(root, filepath.Join(t.r.RunDir, checksDir))+"base-"+c.Name+".log")
		}
	}
	return nil
}

func checkNames(cs []state.PlanCheck) string {
	names := make([]string, len(cs))
	for i, c := range cs {
		names[i] = c.Name
	}
	return "[" + strings.Join(names, " ") + "]"
}
