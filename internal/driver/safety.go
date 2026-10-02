package driver

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

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
// process is gone is cleared and reported through warn. The returned function
// releases the lock and is safe to call more than once.
func AcquireLock(root, branch, runID string, now time.Time, warn func(format string, a ...any)) (func(), error) {
	path := filepath.Join(root, filepath.FromSlash(lockFile))
	if data, err := os.ReadFile(path); err == nil {
		var held lockRecord
		_ = json.Unmarshal(data, &held)
		if pid := held.pid(); pid > 0 && processAlive(pid) {
			other := held.Branch
			if other == "" {
				other = "?"
			}
			started := held.Started
			if started == "" {
				started = "?"
			}
			return nil, halt(ExitPreflight, "a loop is already running in this working tree (pid %d, branch '%s', started %s) — two loops in one tree share .vloop/state/state.json and .vloop/tmp/proposal.json; to run in parallel give each its own worktree: git worktree add ../<dir> <branch>", pid, other, started)
		} else if warn != nil {
			if pid > 0 {
				warn("clearing a stale lock (pid %d is gone)", pid)
			} else {
				warn("clearing a stale lock (pid unknown is gone)")
			}
		}
	}
	rec, _ := json.Marshal(lockRecord{PID: os.Getpid(), Branch: branch, Started: now.UTC().Format(time.RFC3339), Run: runID})
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, append(rec, '\n'), 0o644); err != nil {
		return nil, err
	}
	return func() {
		if data, err := os.ReadFile(path); err == nil {
			var cur lockRecord
			if json.Unmarshal(data, &cur) == nil && cur.pid() != os.Getpid() {
				return
			}
		}
		_ = os.Remove(path)
	}, nil
}

// stateGuard is the plan's bytes as the driver left them before a session ran.
type stateGuard struct {
	root string
	pre  []byte
}

func snapshotState(root string) stateGuard {
	data, _ := os.ReadFile(state.Path(root))
	return stateGuard{root: root, pre: data}
}

// restoreIfTouched puts the plan back and reports true when the session changed
// it, in any byte.
func (g stateGuard) restoreIfTouched() bool {
	if len(g.pre) == 0 {
		return false
	}
	now, err := os.ReadFile(state.Path(g.root))
	if err == nil && bytes.Equal(now, g.pre) {
		return false
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false
	}
	_ = os.WriteFile(state.Path(g.root), g.pre, 0o644)
	return true
}

func tamperNote(phase string) string {
	if phase == PhaseWork {
		return "work session modified " + state.FilePath + " — restored by the driver; the plan and its verify commands are not a session's to edit"
	}
	return fmt.Sprintf("%s session modified %s — restored by the driver; only the driver makes status transitions", phase, state.FilePath)
}

// PlanHashEnv names the variable that hands a session the SHA-256 of the plan
// the driver holds, so that `vloop task gate` runs only that plan.
const PlanHashEnv = "VLOOP_PLAN_SHA256"

// GateRefusedFile is where `vloop task gate` notes, for the driver, that it
// refused a plan other than the one the driver holds.
const GateRefusedFile = tmpDir + "/gate-refused"

// planHash is the lower-case hex SHA-256 of the plan's bytes.
func planHash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

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
// wrote during the session, such as the session's own record.
func (g inputGuard) restore(keep string) []string {
	var changed []string
	now := g.walk()
	for p, want := range g.files {
		abs := filepath.Join(g.root, filepath.FromSlash(p))
		got, err := os.ReadFile(abs)
		if err == nil && bytes.Equal(got, want) {
			continue
		}
		_ = os.MkdirAll(filepath.Dir(abs), 0o755)
		_ = os.WriteFile(abs, want, 0o644)
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
		_ = os.Remove(abs)
		changed = append(changed, p)
	}
	sort.Strings(changed)
	return changed
}
