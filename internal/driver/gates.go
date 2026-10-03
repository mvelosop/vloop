package driver

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/mvelosop/vloop/internal/state"
)

// gateFilesMoved lists the judge files a work session rewrote: modified (or
// deleted) in the working tree, present at HEAD, named, as a whole token, by
// the verify of this task or of any done task, owned by no task of the plan,
// and not committed by this task's own earlier attempt. A file some task owns
// is that task's product — the subject a gate exercises, which the regression
// net judges — never a judge. Creating such a file is the durable-artifact
// shape and is never listed.
func (it *Iterator) gateFilesMoved(t *state.Task) []string {
	out, err := git(it.Root, "diff", "--name-only", "HEAD")
	if err != nil || out == "" {
		return nil
	}
	own := map[string]bool{}
	for _, f := range t.Files {
		own[normGatePath(f)] = true
	}
	named := verifyTokens(t.Verify)
	if it.plan != nil {
		for i := range it.plan.Tasks {
			d := &it.plan.Tasks[i]
			for _, f := range d.Files {
				own[normGatePath(f)] = true
			}
			if d.Status == "done" {
				for k := range verifyTokens(d.Verify) {
					named[k] = true
				}
			}
		}
	}
	var moved []string
	for _, f := range strings.Split(out, "\n") {
		if f == "" || own[normGatePath(f)] || !named[normGatePath(f)] {
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

// normGatePath gives a path one form to compare in: forward slashes, no
// leading "./", no quotes.
func normGatePath(p string) string {
	p = strings.Trim(p, "'\"`")
	p = strings.ReplaceAll(p, "\\", "/")
	for strings.HasPrefix(p, "./") {
		p = p[2:]
	}
	return p
}

// verifyTokens is the set of normalised whitespace- and operator-separated
// tokens of a verify command.
func verifyTokens(verify string) map[string]bool {
	set := map[string]bool{}
	fields := strings.FieldsFunc(verify, func(r rune) bool {
		return unicode.IsSpace(r) || strings.ContainsRune(";&|()<>", r)
	})
	for _, f := range fields {
		if n := normGatePath(f); n != "" {
			set[n] = true
		}
	}
	return set
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
