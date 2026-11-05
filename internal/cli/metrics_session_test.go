package cli

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sessBrief = "B20260101-0000-sx1"

// sessionRepo builds a repository with one brief whose sessions span a
// shell-loop folder and two vloop folders, plus a brief-review session.
func sessionRepo(t *testing.T) string {
	dir := t.TempDir()
	gitIn(t, dir, "2026-01-01T08:00:00Z", "init", "-q", "-b", "main")
	write(t, dir, "docs/briefs/"+sessBrief+".loop-brief.md", "---\nname: "+sessBrief+".loop-brief\nstatus: ready\n---\n")
	v := ".vloop/state/runs/" + sessBrief + "/"
	write(t, dir, ".loop/state/runs/g1/20251231-235900/loop.log", "planning from docs/briefs/"+sessBrief+".loop-brief.md\n")
	write(t, dir, ".loop/state/runs/g1/20251231-235900/sessions/001-plan.json",
		`{"type":"result","is_error":false,"duration_ms":125000,"num_turns":7,"total_cost_usd":0.05,"permission_denials":[]}`)
	write(t, dir, v+"20260101-100000/iterations.jsonl", `{"iteration":1,"task":"T1","attempt":1,"outcome":"done"}`+"\n")
	s := v + "20260101-100000/sessions/"
	write(t, dir, s+"001-plan.json", `{"schema":"session/v3","phase":"plan","iteration":0,"agent":"claude","model":"opus","started":"2026-01-01T10:00:00Z","duration_ms":580997,"wall_ms":581901,"cost_usd":2.887266800000001,"cost_source":"reported","turns":44,"is_error":false,"permission_denials":[{"tool_name":"bash","message":"m"},{"tool_name":"Bash","command":"c"}]}`)
	write(t, dir, s+"003-work.json", `{"schema":"session/v4","phase":"work","iteration":1,"agent":"copilot","model":"gpt-5.1","effort":"high","started":"2026-01-01T10:10:00Z","duration_ms":120000,"wall_ms":125400,"cost_usd":0.4,"cost_source":"estimated","credits":1.5,"turns":12,"permission_denials":[]}`)
	write(t, dir, s+"004-review.json", `{"schema":"session/v4","phase":"review","iteration":1,"agent":"claude","model":"sonnet","started":"2026-01-01T10:15:00Z","wall_ms":30000,"cost_usd":0.25,"cost_source":"reported","turns":5,"is_error":true,"permission_denials":[{"tool_name":"Write","file_path":"f"},{"tool_input":{"command":"x"}}]}`)
	write(t, dir, s+"005-second-review.json", `{"schema":"session/v4","phase":"second-review","iteration":1,"agent":"copilot","model":"gpt-5.1","started":"2026-01-01T10:16:00Z","wall_ms":45000,"cost_usd":0,"cost_source":"estimated","credits":0,"turns":3,"permission_denials":[]}`)
	write(t, dir, s+"002-other.json", `{"schema":"session/v4","phase":"other","iteration":0,"cost_usd":9,"started":"2026-01-01T10:01:00Z","permission_denials":[]}`)
	write(t, dir, v+"20260102-000000/sessions/001-work.json", `{"schema":"session/v1","phase":"work","iteration":1,"started":"2026-01-02T00:00:00Z","duration_ms":600000,"cost_usd":0.1234,"permission_denials":[]}`)
	write(t, dir, ".vloop/state/brief-reviews/"+sessBrief+"/20260101-090000/sessions/001-review-brief.json", `{"schema":"session/v4","phase":"review-brief","cost_usd":7,"started":"2026-01-01T09:00:00Z","permission_denials":[]}`)
	gitIn(t, dir, "2026-01-01T08:00:00Z", "add", "-A")
	gitIn(t, dir, "2026-01-01T08:00:00Z", "commit", "-q", "-m", "fixtures")
	return dir
}

func TestMetricsBySessionText(t *testing.T) {
	dir := sessionRepo(t)
	code, out, errs := runCLI(t, dir, "metrics", sessBrief, "--by", "session")
	want := "run # phase task agent model time cost source credits turns denied error\n" +
		"20251231-235900 001 plan - claude - 2m05s $0.05 - - 7 0 -\n" +
		"20260101-100000 001 plan - claude opus 9m42s $2.89 reported - 44 2 -\n" +
		"20260101-100000 003 work T1 copilot gpt-5.1 2m05s $0.40 estimated 1.50 12 0 -\n" +
		"20260101-100000 004 review T1 claude sonnet 0m30s $0.25 reported - 5 2 yes\n" +
		"20260101-100000 005 second-review T1 copilot gpt-5.1 0m45s $0.00 estimated 0.00 3 0 -\n" +
		"20260102-000000 001 work - claude - 10m00s $0.12 - - - 0 -"
	if code != 0 || collapse(out) != want {
		t.Fatalf("%d\n%s\n%s", code, out, errs)
	}
}

func TestMetricsBySessionJSON(t *testing.T) {
	dir := sessionRepo(t)
	for _, args := range [][]string{{"--json", "metrics", sessBrief}, {"--json", "metrics", sessBrief, "--by", "session"}} {
		code, out, errs := runCLI(t, dir, args...)
		if code != 0 {
			t.Fatalf("%d %s", code, errs)
		}
		var doc struct {
			Cost struct {
				Total float64 `json:"total_usd"`
			} `json:"cost"`
			Denials int `json:"permission_denials"`
			Rows    []struct {
				Folder  string   `json:"folder"`
				Task    *string  `json:"task"`
				Credits *float64 `json:"credits"`
				Turns   *int     `json:"turns"`
				CostUSD float64  `json:"cost_usd"`
				Denied  int      `json:"permission_denials"`
				Schema  *string  `json:"schema"`
				Denials []struct {
					ToolName string `json:"tool_name"`
					Text     string `json:"text"`
				} `json:"denials"`
				IsError bool `json:"is_error"`
			} `json:"by_session"`
		}
		if err := json.Unmarshal([]byte(out), &doc); err != nil {
			t.Fatal(err)
		}
		if len(doc.Rows) != 6 {
			t.Fatalf("%d rows, want 6", len(doc.Rows))
		}
		var cost float64
		var den int
		for _, r := range doc.Rows {
			cost += r.CostUSD
			den += r.Denied
		}
		if math.Abs(cost-doc.Cost.Total) > 1e-9 || den != doc.Denials {
			t.Errorf("rows sum %v/%d, report %v/%d", cost, den, doc.Cost.Total, doc.Denials)
		}
		r0, r1, r3 := doc.Rows[0], doc.Rows[1], doc.Rows[3]
		if r0.Folder != ".loop/state/runs/g1/20251231-235900" || r0.Schema != nil || r0.Task != nil || r0.Turns == nil || *r0.Turns != 7 || r0.Credits != nil {
			t.Errorf("legacy row: %+v", r0)
		}
		if r1.CostUSD != 2.887266800000001 || len(r1.Denials) != 2 || r1.Denials[0].Text != "m" || r1.Denials[1].Text != "c" {
			t.Errorf("plan row: %+v", r1)
		}
		if !r3.IsError || len(r3.Denials) != 2 || r3.Denials[0].ToolName != "Write" || r3.Denials[0].Text != "f" || r3.Denials[1].ToolName != "" || r3.Denials[1].Text != "x" {
			t.Errorf("review row: %+v", r3)
		}
		if r5 := doc.Rows[5]; r5.Turns != nil {
			t.Errorf("record with neither turns field: %+v", r5)
		}
		if code, _, errs := runCLI(t, dir, "schema", "validate", "metrics/v4", writeTemp(t, out)); code != 0 {
			t.Errorf("schema: %s", errs)
		}
	}
}

func writeTemp(t *testing.T, s string) string {
	p := filepath.Join(t.TempDir(), "doc.json")
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestMetricsV4WithoutBySessionValidates(t *testing.T) {
	dir := sessionRepo(t)
	_, out, _ := runCLI(t, dir, "--json", "metrics", sessBrief)
	var doc map[string]any
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatal(err)
	}
	delete(doc, "by_session")
	b, _ := json.Marshal(doc)
	if code, _, errs := runCLI(t, dir, "schema", "validate", "metrics/v4", writeTemp(t, string(b))); code != 0 {
		t.Errorf("without by_session: %s", errs)
	}
	doc["by_session"] = "x"
	b, _ = json.Marshal(doc)
	if code, _, _ := runCLI(t, dir, "schema", "validate", "metrics/v4", writeTemp(t, string(b))); code == 0 {
		t.Error("a string by_session validated")
	}
}

func TestMetricsBySessionRefusals(t *testing.T) {
	dir := sessionRepo(t)
	for name, tc := range map[string]struct {
		args []string
		want string
	}{
		"other":         {[]string{"metrics", sessBrief, "--by", "phase"}, `unknown --by "phase": want task or session`},
		"workspace":     {[]string{"metrics", "--by", "session", "--workspace", "w.toml"}, "--by session reports one repository's briefs; it takes no --workspace"},
		"interventions": {[]string{"metrics", "--interventions", "--by", "session"}, "--by takes kind or phase with --interventions"},
	} {
		t.Run(name, func(t *testing.T) {
			code, out, errs := runCLI(t, dir, tc.args...)
			if code != 2 || out != "" || !strings.Contains(errs, tc.want) {
				t.Errorf("%d %q %q", code, out, errs)
			}
		})
	}
}
