package brief

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Newest is the path of the newest brief whose status is ready (names carry
// their timestamp); "" when there is none.
func Newest(entries []Entry) string {
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].Status == "ready" {
			return entries[i].Path
		}
	}
	return ""
}

// Plannable is why the brief at path cannot be planned, in the words the
// command prints after "vloop: "; nil when it is ready, passes `brief check`
// and every dependency is consumed (B-2, B-6). With replan, the journal of an
// earlier run is not a problem: planning it again is what was asked for.
func Plannable(root, path string, set *HeadingSet, replan bool) error {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		return err
	}
	b := Parse(path, string(data))
	if b.Status != "ready" {
		status := b.Status
		if status == "" {
			status = "unset"
		}
		return fmt.Errorf("%s is %s, not ready — set status: ready once it passes vloop brief check", path, status)
	}
	problems := 0
	for _, p := range Check(root, b, set).Problems() {
		if !replan || !strings.HasPrefix(p, alreadyRun) {
			problems++
		}
	}
	if problems > 0 {
		return fmt.Errorf("%s fails vloop brief check (%d problem(s)) — run vloop brief check %s", path, problems, path)
	}
	entries, err := Load(root)
	if err != nil {
		return err
	}
	status := map[string]string{}
	for _, e := range entries {
		status[e.Name] = e.Status
	}
	for _, d := range b.fm.DependsOn {
		if status[d] != "consumed" {
			return fmt.Errorf("%s depends on %s, which is %s, not consumed", path, d, status[d])
		}
	}
	return nil
}
