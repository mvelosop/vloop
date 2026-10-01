package driver

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"

	"github.com/mvelosop/vloop/internal/closing"
	"github.com/mvelosop/vloop/internal/metrics"
)

// snapshot computes the brief's metrics as of HEAD and holds them to be
// written into the next commit. They are held, not written, so that nothing
// tracked is modified while a session or a gate runs. A failure only costs the
// snapshot: it is warned about, never a reason to stop the run.
func (it *Iterator) snapshot() {
	r, err := it.report()
	if err != nil {
		it.warn("metrics snapshot skipped: %v", err)
		return
	}
	if r == nil {
		return
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		it.warn("metrics snapshot skipped: %v", err)
		return
	}
	it.snap = append(data, '\n')
	it.snapPath = path.Join(closing.SnapshotDir, r.RunID+".json")
}

// report is `vloop metrics`' report for this run's brief; nil when it has no runs.
func (it *Iterator) report() (*metrics.Report, error) {
	c, err := metrics.NewClassifier(it.Root)
	if err != nil {
		return nil, err
	}
	return metrics.Build(it.Root, it.plan.Brief, c)
}

// writeSnapshot puts the held snapshot on disk, just before a commit.
func (it *Iterator) writeSnapshot() error {
	if it.snap == nil {
		return nil
	}
	full := filepath.Join(it.Root, filepath.FromSlash(it.snapPath))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(full, it.snap, 0o644); err != nil {
		return err
	}
	it.snap = nil
	return nil
}

// summary prints the brief's `vloop metrics` summary: the run's last word, on
// stdout only (the run log is already committed), however the run ended.
func (it *Iterator) summary() {
	if it.Out == nil {
		return
	}
	r, err := it.report()
	if err != nil {
		if it.Err != nil {
			it.warn("metrics summary skipped: %v", err)
		}
		return
	}
	if r != nil {
		metrics.PrintSummary(it.Out, r)
	}
}
