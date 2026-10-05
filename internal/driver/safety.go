package driver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mvelosop/vloop/internal/brief"
	"github.com/mvelosop/vloop/internal/state"
)

// lockFile is the run lock, inside the scratch folder so it is never committed.
const lockFile = tmpDir + "/.running"

type lockRecord struct {
	PID     any    `json:"pid"` // a number, or a string as the shell driver wrote it
	Branch  string `json:"branch"`
	Started string `json:"started"`
	Run     string `json:"run"`
}

func (l lockRecord) pid() int {
	switch v := l.PID.(type) {
	case float64:
		return int(v)
	case string:
		n, _ := strconv.Atoi(v)
		return n
	}
	return 0
}

// AcquireLock takes the run lock of the working tree at root. A live holder
// refuses with a *Halt naming git worktree as the alternative; a lock whose
// process is gone is replaced once and reported through warn. The record is
// written to a private file and hard-linked into place, so the lock appears
// complete or not at all, never follows a symlink, and exactly one of several
// concurrent acquisitions succeeds. The returned function releases the lock and
// is safe to call more than once.
func AcquireLock(root, branch, runID string, now time.Time, warn func(format string, a ...any)) (func(), error) {
	path := filepath.Join(root, filepath.FromSlash(lockFile))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	rec, _ := json.Marshal(lockRecord{PID: os.Getpid(), Branch: branch, Started: now.UTC().Format(time.RFC3339), Run: runID})
	rec = append(rec, '\n')
	tmp, err := os.CreateTemp(filepath.Dir(path), ".running-*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())
	_, werr := tmp.Write(rec)
	if cerr := tmp.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		return nil, werr
	}

	for attempt := 0; ; attempt++ {
		err := os.Link(tmp.Name(), path)
		if err == nil {
			break
		}
		if !errors.Is(err, fs.ErrExist) {
			return nil, err
		}
		held, raw, regular := readLock(path)
		if pid := held.pid(); regular && pid > 0 && processAlive(pid) {
			return nil, lockHeld(held)
		}
		if attempt > 0 {
			return nil, halt(ExitPreflight, "could not take the run lock %s: another loop is starting in this working tree", lockFile)
		}
		if warn != nil {
			if pid := held.pid(); pid > 0 {
				warn("clearing a stale lock (pid %d is gone)", pid)
			} else {
				warn("clearing a stale lock (pid unknown is gone)")
			}
		}
		// Remove only the lock that was judged stale, not a newer one.
		if _, now, _ := readLock(path); bytes.Equal(now, raw) {
			_ = os.Remove(path)
		}
	}
	return func() {
		if held, _, regular := readLock(path); !regular || held.pid() != os.Getpid() {
			return
		}
		_ = os.Remove(path)
	}, nil
}

// readLock reads the lock at path without following a symlink. regular is false
// for anything but a regular file, which holds no usable record.
func readLock(path string) (rec lockRecord, raw []byte, regular bool) {
	fi, err := os.Lstat(path)
	if err != nil || !fi.Mode().IsRegular() {
		return rec, nil, false
	}
	raw, err = os.ReadFile(path)
	if err != nil {
		return rec, nil, false
	}
	_ = json.Unmarshal(raw, &rec)
	return rec, raw, true
}

func lockHeld(held lockRecord) error {
	other := held.Branch
	if other == "" {
		other = "?"
	}
	started := held.Started
	if started == "" {
		started = "?"
	}
	return halt(ExitPreflight, "a loop is already running in this working tree (pid %d, branch '%s', started %s) — two loops in one tree share .vloop/state/state.json and .vloop/tmp/proposal.json; to run in parallel give each its own worktree: git worktree add ../<dir> <branch>", held.pid(), other, started)
}

// runIDPattern is what a run id may be: it names a folder and a journal.
var runIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// CheckRunID refuses a committed plan whose run_id is not a plain name, or is
// not the run id of the brief the plan names. A missing plan is not its
// concern, and neither is one about to be replaced: briefPath names another
// brief than the plan's.
func CheckRunID(root, briefPath string) error {
	plan, err := state.Load(root)
	if err != nil || (briefPath != "" && briefPath != plan.Brief) {
		return nil
	}
	return checkPlanRunID(plan)
}

func checkPlanRunID(plan *state.Plan) error {
	id := plan.RunID
	if !runIDPattern.MatchString(id) || strings.Contains(id, "..") {
		return halt(ExitPreflight, "the plan's run_id %q is not a plain name (letters, digits, '.', '_' and '-', no '..') — it names folders and the journal; nothing was written", id)
	}
	if want := brief.RunID(plan.Brief); id != want {
		return halt(ExitPreflight, "the plan's run_id %q is not the run id of its brief %s (%q) — nothing was written", id, plan.Brief, want)
	}
	return nil
}

// stateGuard is the plan's bytes and the gate folders as the driver left them
// before a session ran.
type stateGuard struct {
	root  string
	pre   []byte
	gates map[string][]byte // path relative to the gates folder -> bytes
}

func snapshotState(root string) stateGuard {
	data, _ := os.ReadFile(state.Path(root))
	gates, _ := state.GateFiles(root)
	return stateGuard{root: root, pre: data, gates: gates}
}

// hash is the plan hash handed to the session: the plan and every gate folder.
func (g stateGuard) hash() string { return state.PlanDigest(g.pre, g.gates) }

// restoreIfTouched puts the plan and the gate folders back and reports true
// when the session changed either, in any byte.
func (g stateGuard) restoreIfTouched() bool { return len(g.restoreTouched()) > 0 }

// restoreTouched puts the plan and the gate folders back and returns the
// repo-relative paths the session changed, sorted; none when it changed nothing.
func (g stateGuard) restoreTouched() []string {
	if len(g.pre) == 0 {
		return nil
	}
	var touched []string
	for _, p := range state.RestoreGateFiles(g.root, g.gates) {
		touched = append(touched, state.GatesDir+"/"+p)
	}
	now, err := os.ReadFile(state.Path(g.root))
	if (err == nil && !bytes.Equal(now, g.pre)) || errors.Is(err, os.ErrNotExist) {
		_ = os.WriteFile(state.Path(g.root), g.pre, 0o644)
		touched = append(touched, state.FilePath)
	}
	sort.Strings(touched)
	return touched
}

func tamperNote(phase string) string {
	if phase == PhaseWork {
		return "work session modified " + state.FilePath + " or " + state.GatesDir + " — restored by the driver; the plan, its verify commands and its gate folders are not a session's to edit"
	}
	return fmt.Sprintf("%s session modified %s — restored by the driver; only the driver makes status transitions", phase, state.FilePath)
}

// PlanHashEnv names the variable that hands a session the SHA-256 of the plan
// the driver holds, so that `vloop task gate` runs only that plan.
const PlanHashEnv = "VLOOP_PLAN_SHA256"

// GateRefusedFile is where `vloop task gate` notes, for the driver, that it
// refused a plan other than the one the driver holds.
const GateRefusedFile = tmpDir + "/gate-refused"

// inputGuard holds, in memory, the files a session must leave alone: the
// config, the operator's defects and interventions, and the run's own session
// and report records. Detection compares what is on disk after the session
// with what the driver read before it, whatever commands the session ran.
type inputGuard struct {
	root  string
	roots []string          // repo-relative, slash-separated
	files map[string][]byte // repo-relative path -> bytes before the session
}

func snapshotInputs(root, runDir string) inputGuard {
	g := inputGuard{root: root, files: map[string][]byte{}}
	g.roots = []string{".vloop/config.toml", ".vloop/defects", ".vloop/interventions"}
	if rel, err := filepath.Rel(root, runDir); err == nil {
		rel = filepath.ToSlash(rel)
		g.roots = append(g.roots, rel+"/sessions", rel+"/reports")
	}
	for p := range g.walk() {
		if data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(p))); err == nil {
			g.files[p] = data
		}
	}
	return g
}

// walk lists the regular files now under the guarded paths.
func (g inputGuard) walk() map[string]bool {
	out := map[string]bool{}
	for _, r := range g.roots {
		base := filepath.Join(g.root, filepath.FromSlash(r))
		_ = filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
			if err == nil && d.Type().IsRegular() {
				if rel, err := filepath.Rel(g.root, p); err == nil {
					out[filepath.ToSlash(rel)] = true
				}
			}
			return nil
		})
	}
	return out
}

// restore puts every changed, removed or added file back and returns the
// repo-relative paths it touched, sorted. keep is a file the driver itself
// wrote during the session, such as the session's own record. A path that could
// not be put back is not in the first list; the second names it and why.
func (g inputGuard) restore(keep string) (changed, failed []string) {
	now := g.walk()
	for p, want := range g.files {
		abs := filepath.Join(g.root, filepath.FromSlash(p))
		got, err := os.ReadFile(abs)
		if err == nil && bytes.Equal(got, want) {
			continue
		}
		err = os.MkdirAll(filepath.Dir(abs), 0o755)
		if err == nil {
			err = os.WriteFile(abs, want, 0o644)
		}
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", p, err))
			continue
		}
		changed = append(changed, p)
	}
	for p := range now {
		if _, ok := g.files[p]; ok {
			continue
		}
		abs := filepath.Join(g.root, filepath.FromSlash(p))
		if keep != "" && filepath.Clean(abs) == filepath.Clean(keep) {
			continue
		}
		if err := os.Remove(abs); err != nil && !os.IsNotExist(err) {
			failed = append(failed, fmt.Sprintf("%s: %v", p, err))
			continue
		}
		changed = append(changed, p)
	}
	sort.Strings(changed)
	sort.Strings(failed)
	return changed, failed
}
