package schema

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

var invalidAt = map[string]string{
	"state":        "/status",
	"proposal":     "/outcome",
	"verdict":      "/criteria/0/met",
	"session":      "/is_error",
	"iteration":    "/gate/exit",
	"defect":       "/found-by",
	"intervention": "/phase",
	"install":      "/version",
	"metrics":      "/tasks/done",

	// The v2 contracts and gate-verdict/v1; a name ending in 2 is the v2 of the part before it.
	"state2":       "/status",
	"iteration2":   "/checks/0/exit",
	"session2":     "",
	"metrics2":     "/tasks/done",
	"gate-verdict": "/tasks/0/findings/0/kind",

	// export/v1 has one fixture pair per record type; the name is the part before the dash.
	"export-brief":        "",
	"export-task":         "/attempts",
	"export-defect":       "/found_by",
	"export-intervention": "/phase",
}

// schemaName maps a fixture base to its schema: state2 is state/v2, export-task is export/v1,
// gate-verdict is gate-verdict/v1.
func schemaName(base string) string {
	if base == "gate-verdict" {
		return "gate-verdict/v1"
	}
	if n, ok := strings.CutSuffix(base, "2"); ok {
		return n + "/v2"
	}
	return strings.SplitN(base, "-", 2)[0] + "/v1"
}

func fixture(t *testing.T, n string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/" + n)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestNames(t *testing.T) {
	want := []string{"defect/v1", "export/v1", "gate-verdict/v1", "install/v1", "intervention/v1", "intervention/v2", "iteration/v1", "iteration/v2", "metrics/v1", "metrics/v2", "proposal/v1", "session/v1", "session/v2", "state/v1", "state/v2", "verdict/v1"}
	if got := Names(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Names() = %v, want %v", got, want)
	}
}

func TestFixturesValidateAndFail(t *testing.T) {
	for base, ptr := range invalidAt {
		name := schemaName(base)
		t.Run(base, func(t *testing.T) {
			vs, err := Validate(name, fixture(t, base+".valid.json"))
			if err != nil || len(vs) != 0 {
				t.Fatalf("valid fixture: %v, %v", vs, err)
			}
			vs, err = Validate(name, fixture(t, base+".invalid.json"))
			if err != nil {
				t.Fatal(err)
			}
			if len(vs) != 1 || vs[0].Pointer != ptr || vs[0].Message == "" {
				t.Fatalf("invalid fixture violations = %+v, want one at %q", vs, ptr)
			}
		})
	}
}

func TestSchemaNameMustMatch(t *testing.T) {
	doc := strings.Replace(string(fixture(t, "verdict.valid.json")), "verdict/v1", "state/v1", 1)
	vs, _ := Validate("verdict/v1", []byte(doc))
	if len(vs) != 1 || vs[0].Pointer != "/schema" {
		t.Fatalf("got %+v", vs)
	}
}

func TestSessionTaskRule(t *testing.T) {
	valid := string(fixture(t, "session.valid.json"))
	plan := strings.Replace(valid, `"phase":"work"`, `"phase":"plan"`, 1)
	if vs, _ := Validate("session/v1", []byte(plan)); len(vs) != 1 {
		t.Fatalf("plan session with task: %+v", vs)
	}
	plan = strings.Replace(plan, `"task":"T1",`, "", 1)
	if vs, _ := Validate("session/v1", []byte(plan)); len(vs) != 0 {
		t.Fatalf("plan session without task: %+v", vs)
	}
	work := strings.Replace(valid, `"task":"T1",`, "", 1)
	if vs, _ := Validate("session/v1", []byte(work)); len(vs) != 1 {
		t.Fatalf("work session without task: %+v", vs)
	}
}

func TestNotJSONIsOneViolation(t *testing.T) {
	vs, err := Validate("state/v1", []byte("nope"))
	if err != nil || len(vs) != 1 || vs[0].Pointer != "" {
		t.Fatalf("got %+v, %v", vs, err)
	}
	vs, _ = Validate("state/v1", []byte("{} {}"))
	if len(vs) != 1 || vs[0].Pointer != "" {
		t.Fatalf("trailing data: %+v", vs)
	}
}

func TestUnknownName(t *testing.T) {
	for _, n := range []string{"nope/v1", "state", "../state/v1", "state/v9"} {
		if _, err := Document(n); err == nil {
			t.Errorf("Document(%q) succeeded", n)
		}
	}
}

func TestUnknownKeysAllowed(t *testing.T) {
	doc := strings.Replace(string(fixture(t, "proposal.valid.json")), `"task"`, `"x_new":{"a":1},"task"`, 1)
	if vs, _ := Validate("proposal/v1", []byte(doc)); len(vs) != 0 {
		t.Fatalf("%+v", vs)
	}
}
