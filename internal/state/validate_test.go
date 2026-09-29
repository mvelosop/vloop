package state

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var cliAreas = []string{"cli", "config", "docs"}

// checkFixture runs Check on a scratch repo holding the fixture and docs/x.md.
func checkFixture(t *testing.T, fixture string, areas []string) *Report {
	t.Helper()
	root := scratch(t, filepath.Join("validate", fixture+".json"))
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "x.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(Path(root))
	r, err := Check(root, areas)
	if err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(Path(root))
	if string(before) != string(after) {
		t.Fatal("Check modified the plan")
	}
	return r
}

func TestCheckGoodFixturePasses(t *testing.T) {
	r := checkFixture(t, "good", cliAreas)
	if len(r.Problems) != 0 || len(r.Warnings) != 0 {
		t.Fatalf("problems %q warnings %q", r.Problems, r.Warnings)
	}
}

func TestCheckRulesFire(t *testing.T) {
	cases := []struct {
		fixture string
		areas   []string
		want    []string
	}{
		{"duplicate-id", cliAreas, []string{"duplicate task id: T2"}},
		{"dangling-dependency", cliAreas, []string{"T3 depends on T7, which does not exist"}},
		{"cycle-two", cliAreas, []string{"depends_on cycle: T2 -> T3 -> T2"}},
		{"cycle-three", cliAreas, []string{"depends_on cycle: T1 -> T3 -> T2 -> T1"}},
		{"reference-missing", cliAreas, []string{"T2: reference does not resolve: docs/missing.md"}},
		{"reference-no-reason", cliAreas, []string{"T2: reference has no reason: docs/x.md"}},
		{"area-unknown", cliAreas, []string{`T3: area "ops" is not in areas (cli, config, docs)`}},
		{"area-missing", cliAreas, []string{"T3: no area — areas is set"}},
	}
	for _, c := range cases {
		t.Run(c.fixture, func(t *testing.T) {
			if r := checkFixture(t, c.fixture, c.areas); !reflect.DeepEqual(r.Problems, c.want) {
				t.Fatalf("problems %q, want %q", r.Problems, c.want)
			}
		})
	}
}

func TestCheckSchemaViolations(t *testing.T) {
	for fixture, prefix := range map[string]string{
		"schema-effort": "schema: /tasks/1/effort/work: ",
		"empty-verify":  "schema: /tasks/0/verify: ",
		"no-acceptance": "schema: /tasks/0/acceptance: ",
	} {
		t.Run(fixture, func(t *testing.T) {
			r := checkFixture(t, fixture, cliAreas)
			if len(r.Problems) != 1 || len(r.Problems[0]) <= len(prefix) || r.Problems[0][:len(prefix)] != prefix {
				t.Fatalf("problems %q, want one starting %q", r.Problems, prefix)
			}
		})
	}
}

func TestCheckAreasUnsetIsUnchecked(t *testing.T) {
	for _, fixture := range []string{"area-unknown", "area-missing"} {
		if r := checkFixture(t, fixture, nil); len(r.Problems) != 0 {
			t.Fatalf("%s with areas unset: %q", fixture, r.Problems)
		}
	}
}

func TestCheckHistoryWithoutReasonWarns(t *testing.T) {
	r := checkFixture(t, "history-no-reason", cliAreas)
	if len(r.Problems) != 0 || !reflect.DeepEqual(r.Warnings, []string{"T1: no gate_history reason"}) {
		t.Fatalf("problems %q warnings %q", r.Problems, r.Warnings)
	}
}

func TestCheckNoPlan(t *testing.T) {
	if _, err := Check(t.TempDir(), nil); err != ErrNoPlan {
		t.Fatalf("err %v, want ErrNoPlan", err)
	}
}
