package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const indexRecord = `---
id: %ID%
brief: ""
phase: halt
kind: repair
automatable: partly
by: %BY%
%EXTRA%occurred: 2026-01-01
recorded: 2026-01-01T09:00:00Z
---
a record

**Trigger.** t
`

// indexLayout copies the index script into a temporary repo layout with the
// given records and returns the layout's root.
func indexLayout(t *testing.T, records map[string]string) string {
	t.Helper()
	script, err := os.ReadFile("../../tools/interventions-index.sh")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "tools"), 0o755); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, ".vloop", "interventions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "tools", "interventions-index.sh"), script, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, text := range records {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func indexRecordText(id, by, extra string) string {
	r := strings.NewReplacer("%ID%", id, "%BY%", by, "%EXTRA%", extra)
	return r.Replace(indexRecord)
}

func runIndex(root string) (string, string, error) {
	cmd := exec.Command("bash", filepath.Join(root, "tools", "interventions-index.sh"))
	var out, errb strings.Builder
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	return out.String(), errb.String(), err
}

func TestInterventionsIndexAgreementColumn(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("no bash")
	}
	v2 := "schema: intervention/v2\noptions: 2\nrecommended: 1\ndecided: 2\nagreement: other-option\n"
	root := indexLayout(t, map[string]string{
		"I20260101-0900-old.md": indexRecordText("I20260101-0900-old", "operator", ""),
		"I20260101-0901-new.md": indexRecordText("I20260101-0901-new", "both", v2),
	})
	out, errs, err := runIndex(root)
	if err != nil {
		t.Fatalf("index failed: %v: %s", err, errs)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if lines[0] != "| Phase | Kind | Automatable | Agreement | Id |" {
		t.Errorf("header = %q", lines[0])
	}
	for _, want := range []string{
		"| halt | repair | partly | no-options | [I20260101-0900-old]",
		"| halt | repair | partly | other-option | [I20260101-0901-new]",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("index lacks %q:\n%s", want, out)
		}
	}
}

func TestInterventionsIndexRefusesBadByAndAgreement(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("no bash")
	}
	badAgreement := "schema: intervention/v2\noptions: 2\nrecommended: 1\ndecided: 2\nagreement: maybe\n"
	for name, text := range map[string]string{
		"I20260101-0910-bad-by.md":        indexRecordText("I20260101-0910-bad-by", "robot", ""),
		"I20260101-0911-bad-agreement.md": indexRecordText("I20260101-0911-bad-agreement", "both", badAgreement),
	} {
		root := indexLayout(t, map[string]string{name: text})
		_, errs, err := runIndex(root)
		if err == nil {
			t.Errorf("%s: the index accepted it", name)
		}
		if !strings.Contains(errs, name) {
			t.Errorf("%s: stderr does not name the file: %q", name, errs)
		}
	}
}
