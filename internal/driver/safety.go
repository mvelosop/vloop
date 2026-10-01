package driver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
