package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// wsFixture makes two fixture repositories and a workspace directory of its
// own, returning the workspace directory and the two repository roots.
func wsFixture(t *testing.T) (ws, a, b string) {
	t.Helper()
	return t.TempDir(), closeRepo(t, true, ""), closeRepo(t, true, "")
}

func wsFile(t *testing.T, ws, body string) string {
	t.Helper()
	write(t, ws, "workspace.toml", body)
	return filepath.Join(ws, "workspace.toml")
}

func wsRel(t *testing.T, ws, repo string) string {
	t.Helper()
	r, err := filepath.Rel(ws, repo)
	if err != nil {
		t.Skip("no relative path between temp dirs")
	}
	return filepath.ToSlash(r)
}

func fields(s string) []string { return strings.Fields(s) }

func TestWorkspaceMetricsRowsEqualEachRepo(t *testing.T) {
	ws, a, b := wsFixture(t)
	f := wsFile(t, ws, "[[repo]]\npath = \""+wsRel(t, ws, a)+"\"\nname = \"one\"\n[[repo]]\npath = \""+wsRel(t, ws, b)+"\"\nname = \"two\"\n")
	// run from a directory outside every repository
	code, out, errs := runCLI(t, t.TempDir(), "metrics", "--workspace", f)
	if code != 0 || errs != "" {
		t.Fatalf("exit %d: %s", code, errs)
	}
	_, ownA, _ := runCLI(t, a, "metrics")
	_, ownB, _ := runCLI(t, b, "metrics")
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	wantA := strings.Split(strings.TrimRight(ownA, "\n"), "\n")
	wantB := strings.Split(strings.TrimRight(ownB, "\n"), "\n")
	if len(lines) != 1+(len(wantA)-1)+(len(wantB)-1) {
		t.Fatalf("rows: %q", lines)
	}
	if got := fields(lines[0]); got[0] != "repo" || strings.Join(got[1:], " ") != strings.Join(fields(wantA[0]), " ") {
		t.Errorf("header %q", lines[0])
	}
	i := 1
	for _, blk := range []struct {
		name string
		rows []string
	}{{"one", wantA[1:]}, {"two", wantB[1:]}} {
		for _, r := range blk.rows {
			got := fields(lines[i])
			if got[0] != blk.name || strings.Join(got[1:], " ") != strings.Join(fields(r), " ") {
				t.Errorf("row %d %q, want %s + %q", i, lines[i], blk.name, r)
			}
			i++
		}
	}
}

func TestWorkspaceExportNameOverride(t *testing.T) {
	ws, a, b := wsFixture(t)
	f := wsFile(t, ws, "[[repo]]\npath = \""+wsRel(t, ws, a)+"\"\n[[repo]]\npath = \""+wsRel(t, ws, b)+"\"\nname = \"payments-api\"\n")
	code, out, errs := runCLI(t, t.TempDir(), "metrics", "export", "--workspace", f)
	if code != 0 || errs != "" {
		t.Fatalf("exit %d: %s", code, errs)
	}
	_, ownA, _ := runCLI(t, a, "metrics", "export")
	_, ownB, _ := runCLI(t, b, "metrics", "export")
	if !strings.HasPrefix(out, ownA) {
		t.Errorf("the first repository's export is not unchanged")
	}
	rest := strings.TrimPrefix(out, ownA)
	if rest != strings.ReplaceAll(ownB, `"repo":{"name":"`+filepath.Base(b)+`"`, `"repo":{"name":"payments-api"`) || rest == ownB {
		t.Errorf("second export %q from %q", rest, ownB)
	}
	// without a name the repository keeps its own
	f2 := wsFile(t, ws, "[[repo]]\npath = \""+wsRel(t, ws, a)+"\"\n")
	if _, out, _ := runCLI(t, t.TempDir(), "metrics", "export", "--workspace", f2); out != ownA {
		t.Errorf("no name: export changed")
	}
}

func TestWorkspaceMissingAndNonRepository(t *testing.T) {
	ws, a, _ := wsFixture(t)
	notgit := filepath.Join(ws, "notgit")
	os.MkdirAll(notgit, 0o755)
	f := wsFile(t, ws, "[[repo]]\npath = \"missing\"\n[[repo]]\npath = \"notgit\"\n[[repo]]\npath = \""+wsRel(t, ws, a)+"\"\n")
	for _, args := range [][]string{{"metrics", "--workspace", f}, {"metrics", "export", "--workspace", f}} {
		code, out, errs := runCLI(t, t.TempDir(), args...)
		if code != 1 || errs != "vloop: workspace repo not found: missing\nvloop: workspace repo not found: notgit\n" {
			t.Errorf("%v: exit %d stderr %q", args, code, errs)
		}
		if out == "" {
			t.Errorf("%v: the listed repository was not reported", args)
		}
	}
	if _, err := os.Stat(filepath.Join(ws, "missing")); err == nil {
		t.Error("the missing repository was created")
	}
	// counterpart: every path exists, exit 0 and nothing on stderr
	f = wsFile(t, ws, "[[repo]]\npath = \""+wsRel(t, ws, a)+"\"\n")
	if code, _, errs := runCLI(t, t.TempDir(), "metrics", "--workspace", f); code != 0 || errs != "" {
		t.Errorf("all present: exit %d %q", code, errs)
	}
}

func TestWorkspaceWithBriefIsUsageError(t *testing.T) {
	ws, a, _ := wsFixture(t)
	f := wsFile(t, ws, "[[repo]]\npath = \""+wsRel(t, ws, a)+"\"\n")
	for _, args := range [][]string{{"metrics", closeName, "--workspace", f}, {"metrics", "export", closeName, "--workspace", f}} {
		code, out, errs := runCLI(t, t.TempDir(), args...)
		if code != 2 || out != "" || !strings.HasPrefix(errs, "vloop: ") || strings.Count(errs, "\n") != 1 {
			t.Errorf("%v: exit %d out %q err %q", args, code, out, errs)
		}
	}
}

func TestWorkspaceJSONKeepsReports(t *testing.T) {
	ws, a, _ := wsFixture(t)
	f := wsFile(t, ws, "[[repo]]\npath = \""+wsRel(t, ws, a)+"\"\n")
	code, out, errs := runCLI(t, t.TempDir(), "--json", "metrics", "--workspace", f)
	var rs []map[string]any
	if code != 0 || json.Unmarshal([]byte(out), &rs) != nil || len(rs) != 1 {
		t.Fatalf("exit %d %q %s", code, out, errs)
	}
}
