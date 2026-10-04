package driver

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mvelosop/vloop/internal/config"
	"github.com/mvelosop/vloop/internal/state"
)

const (
	// planRounds is how many plan sessions acceptance allows: the first draft
	// and one revision.
	planRounds   = 2
	planProblems = tmpDir + "/plan-problems.md"
)

// writeProblems leaves the acceptance problems where the revision round of the
// plan session reads them.
func writeProblems(root string, problems []string) error {
	path := filepath.Join(root, filepath.FromSlash(planProblems))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("# Acceptance problems\n\nRevise .vloop/state/state.json (and the gate folders) to fix these; do not start over.\n\n")
	for _, p := range problems {
		b.WriteString("- " + p + "\n")
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// baseGates runs every task's gate on the base, in plan order, under the gate
// scratch and tree rules. A gate that passes proves nothing, and one that
// changes the tree is not a gate; both are acceptance problems.
func (p *Planner) baseGates(t term, plan *state.Plan, timeout time.Duration) ([]string, error) {
	root := p.Root
	scratch, err := config.Get(root, "run.gate-scratch")
	if err != nil {
		return nil, halt(ExitPreflight, "%v", err)
	}
	dir := filepath.Join(t.r.RunDir, "gates")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	var problems []string
	for _, task := range plan.Tasks {
		t.say("checking the gate of %s on the base", task.ID)
		var env []string
		for _, kv := range os.Environ() {
			if strings.HasPrefix(kv, "VLOOP_ACTIVE_TASK=") || strings.HasPrefix(kv, "VLOOP_GATE_TASK=") {
				continue
			}
			env = append(env, kv)
		}
		env = append(env, "VLOOP_ACTIVE_TASK="+task.ID, "VLOOP_GATE_TASK="+task.ID)
		cmd, err := state.GateCommand(root, plan.Shell, task.Verify, env)
		if err != nil {
			return nil, halt(ExitPreflight, "%v", err)
		}
		var out bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &out
		refsBefore := refsState(root)
		gguard := snapshotGit(root)
		tree := snapshotTreeIgnoring(root, gateMayWrite(scratch.List))
		timedOut, runErr := state.RunGroup(cmd, timeout)
		restored := tree.revert()
		if err := state.EmptyScratch(root, scratch.List); err != nil {
			t.warn("   gate scratch of %s not emptied: %v", task.ID, err)
		}
		if what := gguard.changed(); what != "" {
			return nil, halt(ExitRefsMoved, "the gate of %s changed %s — nothing was committed; restore it, then re-run", task.ID, what)
		}
		if len(refsDiff(refsBefore, refsState(root))) > 0 {
			return nil, halt(ExitRefsMoved, "the gate of %s moved git refs — nothing was committed; restore them, then re-run", task.ID)
		}
		exit := 0
		if runErr != nil {
			var ee *exec.ExitError
			if !errors.As(runErr, &ee) {
				return nil, fmt.Errorf("cannot run the gate of %s: %w", task.ID, runErr)
			}
			if exit = ee.ExitCode(); exit <= 0 {
				exit = 1
			}
		}
		if timedOut {
			exit = max(exit, 1)
			if out.Len() > 0 && !bytes.HasSuffix(out.Bytes(), []byte("\n")) {
				out.WriteByte('\n')
			}
			out.WriteString(state.GateTimedOutLine(task.ID, int(timeout/time.Minute)) + "\n")
		}
		if len(restored) > 0 {
			if out.Len() > 0 && !bytes.HasSuffix(out.Bytes(), []byte("\n")) {
				out.WriteByte('\n')
			}
			out.WriteString(state.GateChangedTreeLine(task.ID, restored) + "\n")
		}
		if err := os.WriteFile(filepath.Join(dir, "base-"+task.ID+".log"), []byte(t.r.Mask(out.String())), 0o644); err != nil {
			return nil, err
		}
		if len(restored) > 0 {
			problems = append(problems, fmt.Sprintf("gate %s changed the tree on the base: %s", task.ID, strings.Join(restored, ", ")))
		}
		if exit == 0 {
			problems = append(problems, fmt.Sprintf("gate %s passes on the base — a gate that passes before the work exists proves nothing", task.ID))
		}
	}
	return problems, nil
}
