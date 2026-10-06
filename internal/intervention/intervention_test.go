package intervention

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mvelosop/vloop/internal/config"
)

var t0 = time.Date(2026, 1, 2, 10, 30, 0, 0, time.UTC)

func add(t *testing.T, root, summary, phase, kind, brief string, now time.Time) string {
	t.Helper()
	p, err := Add(root, NewInput{Summary: summary, Brief: brief, Phase: phase, Kind: kind, Automatable: "yes", By: "operator", Trigger: "tr", Done: "dn", Automation: "au"}, now)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestAddWritesRecord(t *testing.T) {
	root := t.TempDir()
	p := add(t, root, "The gate failed", "run", "halt", "B1-x.loop-brief", t0)
	if p != ".vloop/interventions/I20260102-1030-the-gate-failed.md" {
		t.Fatalf("path %q", p)
	}
	b, _ := os.ReadFile(filepath.Join(root, p))
	want := "---\nid: I20260102-1030-the-gate-failed\nbrief: B1-x.loop-brief\nphase: run\nkind: halt\nautomatable: yes\nby: operator\nschema: intervention/v2\noptions: 0\nrecommended: 0\ndecided: \"\"\nagreement: no-options\noccurred: 2026-01-02\nrecorded: 2026-01-02T10:30:00Z\n---\nThe gate failed\n\n**Trigger.** tr\n\n**Done.** dn\n\n**What would automate it.** au\n"
	if string(b) != want {
		t.Fatalf("got\n%s\nwant\n%s", b, want)
	}
	if q := add(t, root, "The gate failed", "run", "halt", "", t0); !strings.HasSuffix(q, "-the-gate-failed-2.md") {
		t.Fatalf("collision path %q", q)
	}
}

func TestListOrderFilterAndRoundTrip(t *testing.T) {
	root := t.TempDir()
	add(t, root, "c", "run", "halt", "B1", t0)
	add(t, root, "b", "run", "ceremony", "B1", t0.Add(time.Minute))
	add(t, root, "d", "design", "decision", "", t0.Add(2*time.Minute))
	add(t, root, "a", "setup", "repair", "", t0.Add(3*time.Minute))
	all, err := List(root, "")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, v := range all {
		got = append(got, v.Phase+"/"+v.Kind)
	}
	if strings.Join(got, " ") != "setup/repair design/decision run/ceremony run/halt" {
		t.Fatalf("order %v", got)
	}
	if v := all[3]; v.Brief != "B1" || v.Trigger != "tr" || v.Done != "dn" || v.Automation != "au" || v.Summary != "c" || v.Schema != "intervention/v2" {
		t.Fatalf("round trip %+v", v)
	}
	if all[0].Brief != "" {
		t.Fatalf("brief %q", all[0].Brief)
	}
	if b, _ := List(root, "B1.md"); len(b) != 2 {
		t.Fatalf("filter: %d", len(b))
	}
}

func TestSetAndValidate(t *testing.T) {
	root := t.TempDir()
	p := add(t, root, "x", "run", "halt", "", t0)
	id := strings.TrimSuffix(filepath.Base(p), ".md")
	if err := Set(root, id, "phase", "verify"); err != nil {
		t.Fatal(err)
	}
	if err := Set(root, id, "brief", "B2-y.loop-brief"); err != nil {
		t.Fatal(err)
	}
	vs, _ := List(root, "")
	if vs[0].Phase != "verify" || vs[0].Brief != "B2-y.loop-brief" || vs[0].Trigger != "tr" {
		t.Fatalf("%+v", vs[0])
	}
	before, _ := os.ReadFile(filepath.Join(root, p))
	for _, c := range [][2]string{{"phase", "bogus"}, {"kind", "x"}, {"automatable", "maybe"}, {"by", "robot"}, {"occurred", "yesterday"}} {
		err := Set(root, id, c[0], c[1])
		if _, ok := err.(*config.InvalidValueError); !ok {
			t.Fatalf("%v: %v", c, err)
		}
	}
	if after, _ := os.ReadFile(filepath.Join(root, p)); string(after) != string(before) {
		t.Fatal("a refused set changed the file")
	}
	if err := Set(root, "I1-none", "phase", "run"); err == nil || strings.Contains(err.Error(), "invalid") {
		t.Fatalf("unknown id: %v", err)
	}
}

func TestExistingRecordsParse(t *testing.T) {
	root := "../.."
	vs, err := List(root, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(vs) == 0 {
		t.Skip("no records")
	}
	for _, v := range vs {
		for f, val := range map[string]string{"phase": v.Phase, "kind": v.Kind, "automatable": v.Automatable, "by": v.By, "occurred": v.Occurred} {
			if err := Validate(f, val); err != nil {
				t.Errorf("%s: %v", v.ID, err)
			}
		}
		if v.Summary == "" || v.Recorded == "" {
			t.Errorf("%s: incomplete %+v", v.ID, v)
		}
	}
}

func TestDeriveAgreement(t *testing.T) {
	for _, c := range []struct {
		options, recommended int
		decided              string
		adjusted             bool
		want                 string
		invalid              bool
	}{
		{0, 0, "", false, "no-options", false},
		{0, 0, "2", true, "no-options", false},
		{2, 1, "1", false, "recommended", false},
		{2, 1, "2", false, "other-option", false},
		{3, 2, "1", true, "adjusted", false},
		{3, 2, "2", true, "adjusted", false},
		{2, 1, "other", false, "different", false},
		{2, 1, "other", true, "different", false},
		{2, 1, "", false, "", true},
		{2, 1, "3", false, "", true},
		{2, 1, "x", false, "", true},
	} {
		got, err := DeriveAgreement(c.options, c.recommended, c.decided, c.adjusted)
		if (err != nil) != c.invalid || got != c.want {
			t.Errorf("%+v: got %q, %v", c, got, err)
		}
	}
}

const v2Head = "---\nid: I1\nbrief: \"\"\nphase: halt\nkind: repair\nautomatable: yes\nby: both\nschema: intervention/v2\n"

func TestParseInvalidRecords(t *testing.T) {
	body := "occurred: 2026-01-01\nrecorded: 2026-01-01T00:00:00Z\n---\ns\n\n**Options.**\n1. a\n2. b\n\n**Recommended.** w\n"
	for name, fm := range map[string]string{
		"stored agreement disagrees": "options: 2\nrecommended: 1\ndecided: 2\nagreement: recommended\n",
		"option count disagrees":     "options: 3\nrecommended: 1\ndecided: 2\nagreement: other-option\n",
		"options without decision":   "options: 2\nrecommended: 1\ndecided: \"\"\nagreement: no-options\n",
		"decided beyond options":     "options: 2\nrecommended: 1\ndecided: 3\nagreement: other-option\n",
		"recommended beyond options": "options: 2\nrecommended: 3\ndecided: 1\nagreement: recommended\n",
		"too many options":           "options: 4\nrecommended: 1\ndecided: 1\nagreement: recommended\n",
		"missing agreement":          "options: 2\nrecommended: 1\ndecided: 1\n",
	} {
		if _, err := parse(v2Head + fm + body); err == nil {
			t.Errorf("%s: parsed without error", name)
		}
	}
	good := v2Head + "options: 2\nrecommended: 1\ndecided: 2\nagreement: other-option\n" + body + "\n**Decided.** reset\n"
	v, err := parse(good)
	if err != nil || v.Decided.Option != 2 || v.Decided.Text != "reset" || v.Recommended.Why != "w" || len(v.Options) != 2 {
		t.Fatalf("%+v %v", v, err)
	}
}

func TestParseV1AndV2Sections(t *testing.T) {
	v1 := "---\nid: I1\nbrief: \"\"\nphase: setup\nkind: repair\nautomatable: yes\nby: assistant\noccurred: 2026-01-01\nrecorded: 2026-01-01T00:00:00Z\n---\ns\n\n**Trigger.** t\n\n**Done.** d\n\n**Context.** c\n\n**Suggested.** sg\n\n**Decided.** dc\n\n**What would automate it.** a\n"
	v, err := parse(v1)
	if err != nil {
		t.Fatal(err)
	}
	if v.Schema != "intervention/v1" || v.Agreement != "no-options" || v.Done != "d" || v.Context != "c" || v.Suggested != "sg" || v.Decided.Text != "dc" || v.Automation != "a" {
		t.Fatalf("%+v", v)
	}
	v, err = parse(v2Head + "options: 1\nrecommended: 1\ndecided: other\nagreement: different\noccurred: 2026-01-01\nrecorded: 2026-01-01T00:00:00Z\n---\ns\n\n**Options.**\n1. a\n\n**Decided.** x\n")
	if err != nil || v.Schema != "intervention/v2" || v.Agreement != "different" || v.Decided.Option != 0 {
		t.Fatalf("%+v %v", v, err)
	}
}

// set adjusted writes the field where I1 puts it, between decided and
// agreement, not at the end of the frontmatter.
func TestSetAdjustedKeepsFieldOrder(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".vloop", "interventions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	rec := "---\nid: I20260101-0900-x\nbrief: \"\"\nphase: halt\nkind: repair\nautomatable: yes\nby: both\n" +
		"schema: intervention/v2\noptions: 2\nrecommended: 1\ndecided: 2\nagreement: other-option\n" +
		"occurred: 2026-01-01\nrecorded: 2026-01-01T09:00:00Z\n---\nx\n\n**Options.**\n1. a\n2. b\n\n**Recommended.** why\n"
	p := filepath.Join(dir, "I20260101-0900-x.md")
	if err := os.WriteFile(p, []byte(rec), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Set(root, "I20260101-0900-x", "adjusted", "true"); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), "decided: 2\nadjusted: true\nagreement: adjusted\noccurred:") {
		t.Fatalf("adjusted is not between decided and agreement:\n%s", b)
	}
}

func TestSetOnCRLFAndBOMRecords(t *testing.T) {
	for name, tc := range map[string]struct{ bom, eol string }{
		"CRLF": {"", "\r\n"}, "BOM": {"\xef\xbb\xbf", "\n"}, "BOM+CRLF": {"\xef\xbb\xbf", "\r\n"},
	} {
		root := t.TempDir()
		dir := filepath.Join(root, ".vloop", "interventions")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		lf := "---\nid: I20260101-0900-x\nbrief: \"\"\nphase: halt\nkind: repair\nautomatable: yes\nby: both\n" +
			"schema: intervention/v2\noptions: 2\nrecommended: 1\ndecided: 2\nagreement: other-option\n" +
			"occurred: 2026-01-01\nrecorded: 2026-01-01T09:00:00Z\n---\nx\n\n**Options.**\n1. a\n2. b\n\n**Recommended.** why\n"
		p := filepath.Join(dir, "I20260101-0900-x.md")
		if err := os.WriteFile(p, []byte(tc.bom+strings.ReplaceAll(lf, "\n", tc.eol)), 0o644); err != nil {
			t.Fatal(err)
		}
		if vs, err := List(root, ""); err != nil || len(vs) != 1 || len(vs[0].Options) != 2 {
			t.Fatalf("%s: list %+v, %v", name, vs, err)
		}
		if err := Set(root, "I20260101-0900-x", "adjusted", "true"); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if err := Set(root, "I20260101-0900-x", "phase", "run"); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		want := strings.Replace(lf, "phase: halt", "phase: run", 1)
		want = strings.Replace(want, "decided: 2\nagreement: other-option", "decided: 2\nadjusted: true\nagreement: adjusted", 1)
		want = tc.bom + strings.ReplaceAll(want, "\n", tc.eol)
		got, _ := os.ReadFile(p)
		if string(got) != want {
			t.Errorf("%s: got %q, want %q", name, got, want)
		}
	}
}
