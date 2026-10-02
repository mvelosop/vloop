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
	want := "---\nid: I20260102-1030-the-gate-failed\nbrief: B1-x.loop-brief\nphase: run\nkind: halt\nautomatable: yes\nby: operator\noccurred: 2026-01-02\nrecorded: 2026-01-02T10:30:00Z\n---\nThe gate failed\n\n**Trigger.** tr\n\n**Done.** dn\n\n**What would automate it.** au\n"
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
	if v := all[3]; v.Brief != "B1" || v.Trigger != "tr" || v.Done != "dn" || v.Automation != "au" || v.Summary != "c" || v.Schema != "intervention/v1" {
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
