package state

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// GatesDir is where the plan session writes a task's gate folder: one folder
// per task, named for its id.
const GatesDir = ".vloop/state/gates"

// GateFolder is the repo-relative, slash-separated path of a task's gate folder.
func GateFolder(id string) string { return GatesDir + "/" + id }

// FixturesDigest is the lower-case hex SHA-256 over the regular files of the
// folder at dir, sorted by path, each fed as <path>\0<contents>\0 with the path
// relative to dir and "/"-separated, so the value is the same on every OS. A
// folder that is missing or holds no regular file gives "".
func FixturesDigest(dir string) (string, error) {
	files, err := readFolder(dir)
	if err != nil {
		return "", err
	}
	return digestFiles(files), nil
}

func digestFiles(files map[string][]byte) string {
	if len(files) == 0 {
		return ""
	}
	h := sha256.New()
	for _, p := range sortedKeys(files) {
		h.Write([]byte(p))
		h.Write([]byte{0})
		h.Write(files[p])
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// readFolder reads the regular files under dir, keyed by their "/"-separated
// path relative to dir. A missing dir is an empty folder.
func readFolder(dir string) (map[string][]byte, error) {
	files := map[string][]byte{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if p == dir && errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)] = data
		return nil
	})
	return files, err
}

func sortedKeys(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// GateFiles reads every regular file under the gate folders of root, keyed by
// its path relative to the gates folder ("T2/seed/rows.txt").
func GateFiles(root string) (map[string][]byte, error) {
	return readFolder(filepath.Join(root, filepath.FromSlash(GatesDir)))
}

// GateFolderIDs lists the entries directly under the gates folder, sorted; a
// missing gates folder has none.
func GateFolderIDs(root string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(GatesDir)))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(entries))
	for _, e := range entries {
		ids = append(ids, e.Name())
	}
	sort.Strings(ids)
	return ids, nil
}

// StrayGateFolders lists the gate folder entries whose name is not a task id of
// the plan.
func StrayGateFolders(root string, p *Plan) ([]string, error) {
	ids, err := GateFolderIDs(root)
	if err != nil {
		return nil, err
	}
	var stray []string
	for _, id := range ids {
		if p.Find(id) == nil {
			stray = append(stray, id)
		}
	}
	return stray, nil
}

// StampFixtures sets every task's fixtures from its gate folder.
func StampFixtures(root string, p *Plan) error {
	for i := range p.Tasks {
		d, err := FixturesDigest(filepath.Join(root, filepath.FromSlash(GateFolder(p.Tasks[i].ID))))
		if err != nil {
			return err
		}
		p.Tasks[i].Fixtures = d
	}
	return nil
}

// FixturesMismatch returns the first task whose stamped fixtures differ from
// its gate folder as it is now, or "".
func FixturesMismatch(root string, p *Plan) (string, error) {
	for _, t := range p.Tasks {
		d, err := FixturesDigest(filepath.Join(root, filepath.FromSlash(GateFolder(t.ID))))
		if err != nil {
			return "", err
		}
		if d != t.Fixtures {
			return t.ID, nil
		}
	}
	return "", nil
}

// PlanDigest is the lower-case hex SHA-256 that identifies the plan a session
// is handed: the plan's bytes, then each gate file as \0<path>\0<contents>,
// sorted by path. With no gate folder it is the SHA-256 of the plan alone.
func PlanDigest(plan []byte, gateFiles map[string][]byte) string {
	h := sha256.New()
	h.Write(plan)
	for _, p := range sortedKeys(gateFiles) {
		h.Write([]byte{0})
		h.Write([]byte(p))
		h.Write([]byte{0})
		h.Write(gateFiles[p])
	}
	return hex.EncodeToString(h.Sum(nil))
}

// RestoreGateFiles makes the gate folders of root exactly the files given:
// changed and removed files are rewritten, added ones deleted, emptied folders
// pruned. It returns the sorted paths (relative to the gates folder) it touched.
func RestoreGateFiles(root string, want map[string][]byte) []string {
	base := filepath.Join(root, filepath.FromSlash(GatesDir))
	now, _ := readFolder(base)
	var changed []string
	for p, data := range want {
		abs := filepath.Join(base, filepath.FromSlash(p))
		if got, ok := now[p]; ok && string(got) == string(data) {
			continue
		}
		_ = os.MkdirAll(filepath.Dir(abs), 0o755)
		_ = os.WriteFile(abs, data, 0o644)
		changed = append(changed, p)
	}
	for p := range now {
		if _, ok := want[p]; ok {
			continue
		}
		_ = os.Remove(filepath.Join(base, filepath.FromSlash(p)))
		changed = append(changed, p)
	}
	if len(changed) > 0 {
		pruneEmpty(base)
	}
	sort.Strings(changed)
	return changed
}

// pruneEmpty removes the empty folders under dir, leaving dir itself.
func pruneEmpty(dir string) {
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.IsDir() {
			sub := filepath.Join(dir, e.Name())
			pruneEmpty(sub)
			_ = os.Remove(sub) // fails, harmlessly, when not empty
		}
	}
}

// RecordGate records a change to a task's gate (P-4): the new command, when
// given and different, and the gate folder as it is now. The old verify and
// fixtures go to gate_history with the reason and who. With neither a new
// command nor a changed folder there is nothing to record.
func RecordGate(root string, p *Plan, id, verify, reason string, at time.Time) error {
	t := p.Find(id)
	if t == nil {
		return &NoTaskError{id}
	}
	fixtures, err := FixturesDigest(filepath.Join(root, filepath.FromSlash(GateFolder(id))))
	if err != nil {
		return err
	}
	newCmd := verify != "" && verify != t.Verify
	if !newCmd && fixtures == t.Fixtures {
		return fmt.Errorf("nothing to record — %s's gate and fixtures are unchanged", id)
	}
	t.GateHistory = append(t.GateHistory, GateReplace{
		Verify:     t.Verify,
		ReplacedAt: at.UTC().Format(time.RFC3339),
		Reason:     reason,
		By:         "operator",
		Fixtures:   t.Fixtures,
	})
	if newCmd {
		t.Verify = verify
	}
	t.Fixtures = fixtures
	return nil
}
