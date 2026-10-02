package cli

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mvelosop/vloop/internal/schema"
)

// exportRecs runs `vloop metrics export` in dir and returns its lines decoded,
// checking each one against export/v1.
func exportRecs(t *testing.T, dir string, args ...string) []map[string]any {
	t.Helper()
	code, out, errs := runCLI(t, dir, append([]string{"metrics", "export"}, args...)...)
	if code != 0 || errs != "" {
		t.Fatalf("exit %d: %s", code, errs)
	}
	var recs []map[string]any
	for _, l := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if v, err := schema.Validate("export/v1", []byte(l)); err != nil || len(v) > 0 {
			t.Fatalf("line %s: %v %v", l, v, err)
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			t.Fatal(err)
		}
		recs = append(recs, m)
	}
	return recs
}

func TestExportRecordOrderAndFields(t *testing.T) {
	dir := closeRepo(t, true, "")
	folder := ".loop/state/runs/" + closeBrief + "/20260101-090000/"
	write(t, dir, folder+"iterations.jsonl",
		`{"iteration":1,"task":"T1","outcome":"gate_failed"}`+"\n"+
			`{"iteration":2,"task":"T1","outcome":"rejected"}`+"\n"+
			`{"iteration":3,"task":"T1","outcome":"done"}`+"\n")
	write(t, dir, folder+"reports/002-verdict.json",
		`{"task":"T1","verdict":"FAIL","findings":[{"summary":"a spec gap","kind":"spec-gap"},{"summary":"an off by one","kind":"bug"}]}`)
	if code, _, errs := runCLI(t, dir, "defect", "add", "README omits the flag", "--found-by", "operator", "--brief", closeName); code != 0 {
		t.Fatal(errs)
	}
	recs := exportRecs(t, dir)
	var got []string
	for _, r := range recs {
		id, _ := r["id"].(string)
		if r["type"] == "brief" {
			id = r["run_id"].(string)
		}
		got = append(got, r["type"].(string)+" "+id)
	}
	want := []string{
		"brief " + closeBrief, "task T1", "task T2",
		"defect " + closeBrief + "/i1-gate", "defect " + closeBrief + "/i2-review-1", "defect " + closeBrief + "/i2-review-2",
	}
	if len(got) != len(want)+1 || strings.Join(got[:len(want)], "|") != strings.Join(want, "|") || !strings.HasPrefix(got[len(want)], "defect D") {
		t.Fatalf("records %q, want %q then a recorded defect", got, want)
	}
	if _, has := recs[0]["by_task"]; has {
		t.Error("brief record carries by_task")
	}
	if recs[1]["attempts"] != float64(3) || recs[1]["status"] != "done" || recs[1]["brief"] != closeName {
		t.Errorf("task record: %v", recs[1])
	}
	d := recs[4]
	if d["found_by"] != "review" || d["origin"] != "brief" || d["kind"] != "spec-gap" || d["task"] != "T1" ||
		d["severity"] != nil || d["status"] != nil || d["derived"] != true || d["summary"] != "a spec gap" {
		t.Errorf("review defect: %v", d)
	}
	if g := recs[3]; g["found_by"] != "gate" || g["origin"] != "work" || g["summary"] != "gate failed" {
		t.Errorf("gate defect: %v", g)
	}
	rd := recs[6]
	if rd["derived"] != false || rd["severity"] != "medium" || rd["status"] != "open" || rd["task"] != nil || rd["found_by"] != "operator" {
		t.Errorf("recorded defect: %v", rd)
	}
}

func TestExportNoRunsForNamedBrief(t *testing.T) {
	dir := closeRepo(t, true, "")
	code, out, errs := runCLI(t, dir, "metrics", "export", "B20260102-0900-c.loop-brief")
	if code != 1 || out != "" || strings.TrimSpace(errs) != "vloop: no runs for B20260102-0900-c.loop-brief" {
		t.Fatalf("exit %d, out %q, err %q", code, out, errs)
	}
}

func TestExportRepoIdentity(t *testing.T) {
	dir := closeRepo(t, true, "")
	for _, c := range []struct{ url, name, remote string }{
		{"https://user:tok@example.com/acme/shop.git", "shop", "https://example.com/acme/shop.git"},
		{"ssh://git:pw@example.com:2222/acme/api.git", "api", "ssh://example.com:2222/acme/api.git"},
		{"https://tok@example.com/acme/shop", "shop", "https://example.com/acme/shop"},
		{"https://example.com/acme/plain.git", "plain", "https://example.com/acme/plain.git"},
	} {
		gitIn(t, dir, "2026-01-02T00:00:00Z", "config", "remote.origin.url", c.url)
		repo := exportRecs(t, dir)[0]["repo"].(map[string]any)
		if repo["name"] != c.name || repo["remote"] != c.remote {
			t.Errorf("%s: repo %v, want %s %s", c.url, repo, c.name, c.remote)
		}
	}
	gitIn(t, dir, "2026-01-02T00:00:00Z", "config", "--unset", "remote.origin.url")
	repo := exportRecs(t, dir)[0]["repo"].(map[string]any)
	if repo["remote"] != nil || repo["name"] != filepath.Base(dir) {
		t.Errorf("no origin: repo %v", repo)
	}
}

func exportInterventions(recs []map[string]any) []map[string]any {
	var out []map[string]any
	for _, r := range recs {
		if r["type"] == "intervention" {
			out = append(out, r)
		}
	}
	return out
}

func TestExportInterventions(t *testing.T) {
	dir := closeRepo(t, true, "")
	for _, a := range [][]string{
		{"Second", "--phase", "design", "--kind", "decision", "--automatable", "no", "--by", "both"},
		{"First", "--phase", "run", "--kind", "halt", "--automatable", "partly", "--by", "operator", "--brief", closeName},
	} {
		if code, _, errs := runCLI(t, dir, append([]string{"intervention", "add"}, a...)...); code != 0 {
			t.Fatal(errs)
		}
	}
	recs := exportRecs(t, dir)
	ivs := exportInterventions(recs)
	if len(ivs) != 2 || recs[len(recs)-1]["type"] != "intervention" {
		t.Fatalf("want 2 trailing intervention records, got %v", recs)
	}
	if ivs[0]["id"].(string) >= ivs[1]["id"].(string) {
		t.Errorf("not in id order: %v %v", ivs[0]["id"], ivs[1]["id"])
	}
	for _, v := range ivs {
		for _, k := range []string{"id", "brief", "phase", "kind", "automatable", "by", "occurred", "recorded"} {
			if _, ok := v[k]; !ok {
				t.Errorf("intervention lacks %s: %v", k, v)
			}
		}
		if v["schema"] != "export/v1" {
			t.Errorf("schema %v", v["schema"])
		}
	}
	byBrief := exportInterventions(exportRecs(t, dir, closeName))
	if len(byBrief) != 1 || byBrief[0]["kind"] != "halt" || byBrief[0]["brief"] != closeName {
		t.Errorf("filtered by brief: %v", byBrief)
	}
}

func TestExportRepoStacks(t *testing.T) {
	dir := closeRepo(t, true, "")
	stacks := func() any { return exportRecs(t, dir)[0]["repo"].(map[string]any)["stacks"] }
	if s, _ := stacks().([]any); len(s) != 1 || s[0] != "go" {
		t.Errorf("configured stacks = %v, want [go]", stacks())
	}
	write(t, dir, ".vloop/config.toml", "")
	if s, ok := stacks().([]any); !ok || len(s) != 0 {
		t.Errorf("unset stacks = %#v, want []", stacks())
	}
}
