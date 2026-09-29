package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const briefText = "---\nname: B20260101-0900-a.loop-brief\nstatus: %s\n---\n# a\n\n## Worked example\n\n```\n$ a\n```\n\n## Out of scope\n\n- one\n- two\n\n## Constraints\n\n- Go.\n\n6 to 9 tasks. Exits 1.\n"

func writeBrief(t *testing.T, root, status string) string {
	t.Helper()
	rel := "docs/briefs/B20260101-0900-a.loop-brief.md"
	if err := os.MkdirAll(filepath.Join(root, "docs", "briefs"), 0o755); err != nil {
		t.Fatal(err)
	}
	text := strings.Replace(briefText, "%s", status, 1)
	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(rel)), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return rel
}

func TestBriefCheckTextAndExit(t *testing.T) {
	d := scratchRepo(t)
	rel := writeBrief(t, d, "ready")
	code, out, e := run(t, "-C", d, "brief", "check", rel)
	if code != 0 || e != "" || !strings.HasPrefix(out, rel+"\n") || !strings.Contains(out, "  ok, 0 warning(s)\n") || !strings.HasSuffix(out, "briefs ok\n") {
		t.Fatalf("%d %q %q", code, out, e)
	}
	if err := os.MkdirAll(filepath.Join(d, ".vloop", "state", "journals"), 0o755); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(d, ".vloop", "state", "journals", "B20260101-0900-a.md"), nil, 0o644)
	code, out, e = run(t, "-C", d, "brief", "check", rel)
	if code != 1 || e != "" || !strings.Contains(out, "✗ already run") || !strings.Contains(out, "  1 problem(s), 0 warning(s)\n") || !strings.HasSuffix(out, "1 brief(s) need work\n") {
		t.Fatalf("%d %q %q", code, out, e)
	}
}

func TestBriefCheckSkipped(t *testing.T) {
	d := scratchRepo(t)
	rel := writeBrief(t, d, "draft")
	_, out, _ := run(t, "-C", d, "brief", "check", rel)
	if out != rel+"\n  - skipped: status is draft\nbriefs ok\n" {
		t.Fatalf("%q", out)
	}
}

func TestBriefCheckUsageAndMissing(t *testing.T) {
	d := scratchRepo(t)
	rel := writeBrief(t, d, "ready")
	code, out, e := run(t, "-C", d, "brief", "check", rel, "docs/notes.md")
	if code != 2 || out != "" || e != "vloop: not a loop brief: docs/notes.md\n" {
		t.Fatalf("%d %q %q", code, out, e)
	}
	if code, out, e := run(t, "-C", d, "brief", "check"); code != 2 || out != "" || strings.Count(e, "\n") != 1 {
		t.Fatalf("%d %q %q", code, out, e)
	}
	miss := "docs/briefs/B20260101-0999-z.loop-brief.md"
	code, out, e = run(t, "-C", d, "brief", "check", miss)
	if code != 1 || e != "vloop: no such brief: "+miss+"\n" || out != "1 brief(s) need work\n" {
		t.Fatalf("%d %q %q", code, out, e)
	}
}

func TestBriefCheckJSON(t *testing.T) {
	d := scratchRepo(t)
	rel := writeBrief(t, d, "ready")
	code, out, _ := run(t, "-C", d, "brief", "check", "--json", rel, "docs/briefs/B20260101-0999-z.loop-brief.md")
	if code != 1 {
		t.Fatalf("code %d", code)
	}
	var doc struct {
		OK     bool `json:"ok"`
		Briefs []struct {
			Path, Result      string
			Problems, Warning []string
		} `json:"briefs"`
	}
	if err := json.Unmarshal([]byte(out), &doc); err != nil || doc.OK || len(doc.Briefs) != 2 || doc.Briefs[0].Result != "ok" || doc.Briefs[0].Path != rel {
		t.Fatalf("%v %s", err, out)
	}
	if !strings.HasPrefix(out, `{"ok":false,"briefs":[{"path":"`+rel+`","result":"ok","problems":[],"warnings":[]}`) {
		t.Fatalf("key order or empty arrays: %s", out)
	}
}

func TestBriefCheckRelativeToCwdRoot(t *testing.T) {
	d := scratchRepo(t)
	rel := writeBrief(t, d, "draft")
	_, out, _ := run(t, "-C", filepath.Join(d, "docs"), "brief", "check", "briefs/"+filepath.Base(rel))
	if !strings.HasPrefix(out, rel+"\n") {
		t.Fatalf("%q", out)
	}
}
