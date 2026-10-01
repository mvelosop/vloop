package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSchemaList(t *testing.T) {
	code, out, _ := run(t, "schema", "list")
	want := "defect/v1\nexport/v1\ninstall/v1\nintervention/v1\niteration/v1\nmetrics/v1\nproposal/v1\nsession/v1\nstate/v1\nverdict/v1\n"
	if code != 0 || out != want {
		t.Fatalf("code %d out %q", code, out)
	}
	code, out, _ = run(t, "schema", "list", "--json")
	var names []string
	if code != 0 || json.Unmarshal([]byte(out), &names) != nil || len(names) != 10 || names[0] != "defect/v1" {
		t.Fatalf("code %d out %q", code, out)
	}
}

func TestSchemaShow(t *testing.T) {
	code, out, _ := run(t, "schema", "show", "verdict/v1")
	var doc map[string]any
	if code != 0 || json.Unmarshal([]byte(out), &doc) != nil || doc["$schema"] != "https://json-schema.org/draft/2020-12/schema" {
		t.Fatalf("code %d out %q", code, out)
	}
}

func TestSchemaValidate(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good.json")
	bad := filepath.Join(dir, "bad.json")
	const doc = `{"schema":"verdict/v1","task":"T1","verdict":"PASS","criteria":[],"findings":[],"notes":""}`
	if err := os.WriteFile(good, []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bad, []byte(strings.Replace(doc, "PASS", "pass", 1)), 0o644); err != nil {
		t.Fatal(err)
	}

	code, out, _ := run(t, "schema", "validate", "verdict/v1", good)
	if code != 0 || out != good+": ok\n" {
		t.Fatalf("good: code %d out %q", code, out)
	}
	code, out, _ = run(t, "schema", "validate", "verdict/v1", bad)
	if code != 1 || !strings.HasPrefix(out, bad+": /verdict: ") || strings.Count(out, "\n") != 1 {
		t.Fatalf("bad: code %d out %q", code, out)
	}
	code, out, _ = run(t, "schema", "validate", "verdict/v1", bad, "--json")
	var r struct {
		OK     bool
		Errors []struct{ Pointer, Message string }
	}
	if code != 1 || json.Unmarshal([]byte(out), &r) != nil || r.OK || len(r.Errors) != 1 || r.Errors[0].Pointer != "/verdict" {
		t.Fatalf("bad --json: code %d out %q", code, out)
	}
	code, out, _ = run(t, "schema", "validate", "verdict/v1", good, "--json")
	if code != 0 || !strings.Contains(out, `"ok":true`) || !strings.Contains(out, `"errors":[]`) {
		t.Fatalf("good --json: code %d out %q", code, out)
	}
}

func TestSchemaUnknownName(t *testing.T) {
	for _, args := range [][]string{{"schema", "show", "nope/v1"}, {"schema", "validate", "nope/v1", "x.json"}} {
		code, out, errOut := run(t, args...)
		if code != 2 || out != "" || errOut != "vloop: unknown schema \"nope/v1\"\n" {
			t.Fatalf("%v: code %d out %q err %q", args, code, out, errOut)
		}
	}
}
