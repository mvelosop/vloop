package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func listBrief(t *testing.T, root, file, status, deps string) {
	t.Helper()
	dir := filepath.Join(root, "docs", "briefs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	text := "---\nname: " + strings.TrimSuffix(file, ".md") + "\nstatus: " + status + "\ndepends-on: " + deps + "\n---\n"
	if err := os.WriteFile(filepath.Join(dir, file), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestBriefListOrderAndJSON(t *testing.T) {
	d := scratchRepo(t)
	listBrief(t, d, "B2-b.loop-brief.md", "consumed", "[]")
	listBrief(t, d, "B3-a.loop-brief.md", "ready", "[B2-b.loop-brief]")
	listBrief(t, d, "B4-c.loop-brief.md", "draft", "[B3-a.loop-brief, B2-b.loop-brief]")
	listBrief(t, d, "notes.md", "ready", "[]")
	code, out, e := run(t, "-C", d, "brief", "list")
	want := "B2-b.loop-brief  consumed  -  -\nB3-a.loop-brief  ready  ready  B2-b.loop-brief\nB4-c.loop-brief  draft  blocked  B3-a.loop-brief,B2-b.loop-brief\n"
	if code != 0 || e != "" || out != want {
		t.Fatalf("%d %q %q", code, out, e)
	}
	code, out, _ = run(t, "-C", d, "brief", "list", "--json")
	var docs []map[string]any
	if code != 0 || json.Unmarshal([]byte(out), &docs) != nil || len(docs) != 3 || docs[0]["path"] != "docs/briefs/B2-b.loop-brief.md" || docs[2]["ready"] != "blocked" {
		t.Fatalf("%d %q", code, out)
	}
	if !strings.HasPrefix(out, `[{"name":"B2-b.loop-brief","path":"docs/briefs/B2-b.loop-brief.md","status":"consumed","ready":"-","depends_on":[]}`) {
		t.Fatalf("key order: %s", out)
	}
}

func TestBriefListEmpty(t *testing.T) {
	d := scratchRepo(t)
	if code, out, _ := run(t, "-C", d, "brief", "list", "--json"); code != 0 || out != "[]\n" {
		t.Fatalf("%d %q", code, out)
	}
}

func TestBriefListDanglingAndCycle(t *testing.T) {
	d := scratchRepo(t)
	listBrief(t, d, "B1-a.loop-brief.md", "ready", "[B9-z.loop-brief]")
	code, out, e := run(t, "-C", d, "brief", "list")
	if code != 1 || out != "" || e != "vloop: depends-on does not resolve: B9-z.loop-brief\n" {
		t.Fatalf("%d %q %q", code, out, e)
	}
	listBrief(t, d, "B1-a.loop-brief.md", "ready", "[B2-b.loop-brief]")
	listBrief(t, d, "B2-b.loop-brief.md", "ready", "[B1-a.loop-brief]")
	code, out, e = run(t, "-C", d, "brief", "list")
	if code != 1 || out != "" || !strings.HasPrefix(e, "vloop: depends-on cycle: B1-a.loop-brief -> B2-b.loop-brief -> B1-a.loop-brief") {
		t.Fatalf("%d %q %q", code, out, e)
	}
}
