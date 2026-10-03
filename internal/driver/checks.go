package driver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
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

// matchGlob reports whether a repo-relative path matches a check's glob: "/"
// separated, "*" and "?" within a segment, "**" for any number of segments.
func matchGlob(pattern, p string) bool {
	return matchSegs(strings.Split(strings.Trim(pattern, "/"), "/"), strings.Split(p, "/"))
}

func matchSegs(pat, segs []string) bool {
	for len(pat) > 0 {
		if pat[0] == "**" {
			for len(pat) > 0 && pat[0] == "**" {
				pat = pat[1:]
			}
			if len(pat) == 0 {
				return true
			}
			for i := range segs {
				if matchSegs(pat, segs[i:]) {
					return true
				}
			}
			return false
		}
		if len(segs) == 0 {
			return false
		}
		if ok, err := path.Match(pat[0], segs[0]); err != nil || !ok {
			return false
		}
		pat, segs = pat[1:], segs[1:]
	}
	return len(segs) == 0
}

// checksFor is the checks, in config order, with a glob matching a changed path.
func checksFor(checks []state.PlanCheck, changed []string) []state.PlanCheck {
	var out []state.PlanCheck
	for _, c := range checks {
		if slices.ContainsFunc(c.Paths, func(g string) bool {
			return slices.ContainsFunc(changed, func(p string) bool { return matchGlob(g, p) })
		}) {
			out = append(out, c)
		}
	}
	return out
}

// changedPaths is what the iteration changed: tracked changes against HEAD and
// untracked files, less the driver's scratch folder.
func changedPaths(root string) []string {
	out, err := gitCmd(root, "status", "--porcelain", "-z", "--no-renames", "--untracked-files=all").Output()
	if err != nil {
		return nil
	}
	var paths []string
	for _, e := range strings.Split(string(out), "\x00") {
		if len(e) < 4 {
			continue
		}
		if p := e[3:]; !strings.HasPrefix(p, tmpDir+"/") {
			paths = append(paths, p)
		}
	}
	return paths
}

// checkRun is one check's outcome in an iteration or the final pass.
type checkRun struct {
	name string
	exit int
	ms   int64
	log  string // repo-relative path of the check's log
}

// runChecks runs the checks in order, stopping at the first failure; a failed
// check is never re-run. Logs go under the run folder's checks/: <prefix><name>.log
// always, and for an iteration also <NNN>-<name>.fail.log when it fails.
func (it *Iterator) runChecks(checks []state.PlanCheck, prefix string, iter int) ([]checkRun, error) {
	dir := filepath.Join(it.RunDir, checksDir)
	if len(checks) > 0 {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	var runs []checkRun
	for _, c := range checks {
		it.say("   check %s", c.Name)
		res, err := runCheck(it.Root, it.plan.Shell, c, it.gateTimeout())
		if err != nil {
			var h *Halt
			if errors.As(err, &h) {
				it.plan.Status = "halted"
				_ = it.save()
				_ = it.log.flush()
			}
			return nil, err
		}
		masked := []byte(it.r.Mask(string(res.log)))
		rel := relRunDir(it.Root, dir) + prefix + c.Name + ".log"
		if err := os.WriteFile(filepath.Join(dir, prefix+c.Name+".log"), masked, 0o644); err != nil {
			return nil, err
		}
		cr := checkRun{name: c.Name, exit: res.exit, ms: res.ms, log: rel}
		if res.exit != 0 && iter > 0 {
			fail := fmt.Sprintf("%03d-%s.fail.log", iter, c.Name)
			if err := os.WriteFile(filepath.Join(dir, fail), masked, 0o644); err != nil {
				return nil, err
			}
			cr.log = relRunDir(it.Root, dir) + fail
		}
		runs = append(runs, cr)
		if res.exit != 0 {
			break
		}
	}
	return runs, nil
}

// checksRecord is the iteration record's checks field.
func checksRecord(runs []checkRun) []map[string]any {
	out := []map[string]any{}
	for _, r := range runs {
		out = append(out, map[string]any{"name": r.name, "exit": r.exit, "duration_ms": r.ms})
	}
	return out
}

// checksFile is where the review session finds the checks' results.
const checksFile = tmpDir + "/checks.json"

// writeChecksFile hands the review session the checks that ran and their logs.
func (it *Iterator) writeChecksFile(runs []checkRun) error {
	type entry struct {
		Name string `json:"name"`
		Exit int    `json:"exit"`
		Log  string `json:"log"`
	}
	list := []entry{}
	for _, r := range runs {
		list = append(list, entry{r.name, r.exit, r.log})
	}
	data, err := json.MarshalIndent(map[string]any{"checks": list}, "", "  ")
	if err != nil {
		return err
	}
	p := filepath.Join(it.Root, filepath.FromSlash(checksFile))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, append(data, '\n'), 0o644)
}

// finalPass runs every check once when the last task is done. It returns the
// failing check's refusal line, "" when all pass.
func (it *Iterator) finalPass() (string, error) {
	runs, err := it.runChecks(it.plan.Checks, "final-", 0)
	if err != nil {
		return "", err
	}
	for _, r := range runs {
		if r.exit != 0 {
			return fmt.Sprintf("vloop: check %s failed in the final pass — see %s", r.name, r.log), nil
		}
	}
	return "", nil
}
