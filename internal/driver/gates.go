package driver

import (
	"fmt"
	"strings"

	"github.com/mvelosop/vloop/internal/state"
)

// gateFilesMoved lists the files a work session rewrote that the task's gate
// judges it by: modified (or deleted) in the working tree, present at HEAD,
// named by the task's verify, not among the task's files, and not committed by
// this task's own earlier attempt. Creating such a file is the durable-artifact
// shape and is never listed.
func (it *Iterator) gateFilesMoved(t *state.Task) []string {
	out, err := git(it.Root, "diff", "--name-only", "HEAD")
	if err != nil || out == "" {
		return nil
	}
	own := map[string]bool{}
	for _, f := range t.Files {
		own[f] = true
	}
	var moved []string
	for _, f := range strings.Split(out, "\n") {
		if f == "" || own[f] || !strings.Contains(t.Verify, f) {
			continue
		}
		if _, err := git(it.Root, "cat-file", "-e", "HEAD:"+f); err != nil {
			continue
		}
		subj, _ := git(it.Root, "log", "-1", "--format=%s", "--", f)
		if strings.HasPrefix(subj, "[vloop] "+t.ID+":") {
			continue
		}
		moved = append(moved, f)
	}
	return moved
}

// restoreGateFiles puts the files back as HEAD holds them. It never touches a
// task's verify, which lives in the plan.
func (it *Iterator) restoreGateFiles(id string, moved []string) string {
	args := append([]string{"checkout", "HEAD", "--"}, moved...)
	_, _ = git(it.Root, args...)
	list := strings.Join(moved, ", ")
	it.warn("   GATE REWRITE %s — %s restored from HEAD; a gate is not a session's to rewrite", id, list)
	return it.r.Mask(fmt.Sprintf("work session modified %s — a file a verify command runs, which this task neither created nor was assigned; restored by the driver", list))
}

// runGateRetry runs a gate and, if it fails, once more at once with nothing
// changed: a pass on the re-run is a flaky gate, recorded and not charged.
func (it *Iterator) runGateRetry(iter int, active, id string) (*gateResult, error) {
	g, err := it.runGate(iter, active, id)
	if err != nil || g.exit == 0 {
		return g, err
	}
	g2, err := it.runGate(iter, active, id)
	if err != nil {
		return nil, err
	}
	if g2.exit == 0 {
		g2.flaky = true
		it.warn("   FLAKY GATE %s — failed, then passed on the immediate re-run", id)
	}
	return g2, nil
}

func gatePassesLine(id string) string {
	return id + " GATE PASSES while the session reports blocked — the work satisfies its own gate; the block is about something else."
}
