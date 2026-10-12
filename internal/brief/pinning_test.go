package brief

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const pinGuidelines = `# Repository guidelines

## Pinning tests

- when ` + "`config key`" + `: ` + "`TestA`, `TestB`, `TestC`" + ` (D20261011-1455)
- when config key: TestBroken
- when ` + "`config key` `TestBroken2`" + `
- when ` + "`schemas/`: `TestS`" + `

## External formats

- when ` + "`other`: `TestZ`" + `
`

func pinBrief(contract, constraints string) string {
	return "---\nname: x\ndescription: d\nkind: brief\nstatus: ready\ncreated: 2026-01-01\nseeds: s\ndepends-on: []\n---\n" +
		"# B\n\n## What it is\n\nA.\n\n## Behaviour contract\n\n" + contract +
		"\n\n## Worked example\n\n```\n$ demo\n```\n\n## Out of scope\n\n- a\n- b\n\n## Constraints\n\n" + constraints + "\n"
}

func pinWarnings(t *testing.T, guidelines, text string) []string {
	t.Helper()
	root, path, _ := scratch(t)
	if guidelines != "" {
		if err := os.MkdirAll(filepath.Join(root, ".vloop"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, GuidelinesFile), []byte(guidelines), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var out []string
	for _, w := range run(root, path, text).Warnings() {
		if strings.HasPrefix(w, "must-change list:") {
			out = append(out, w)
		}
	}
	return out
}

func TestPinningMissingNamesInRuleOrder(t *testing.T) {
	got := pinWarnings(t, pinGuidelines, pinBrief("It adds a config key.", "- Keep TestB green."))
	want := `must-change list: the contract mentions "config key" but Constraints do not name TestA, TestC (.vloop/repo-guidelines.md)`
	if len(got) != 1 || got[0] != want {
		t.Fatalf("got %q", got)
	}
}

func TestPinningCaseInsensitivePhraseAndSubsections(t *testing.T) {
	got := pinWarnings(t, pinGuidelines, pinBrief("### L1\n\nIt ADDS A Config Key.", "- Nothing."))
	if len(got) != 1 || !strings.Contains(got[0], "TestA, TestB, TestC") {
		t.Fatalf("got %q", got)
	}
}

func TestPinningAllNamedOrPhraseOutsideContract(t *testing.T) {
	if got := pinWarnings(t, pinGuidelines, pinBrief("A config key.", "- TestA TestB TestC")); len(got) != 0 {
		t.Fatalf("all named: %q", got)
	}
	// The phrase in the worked example or constraints alone is not the contract.
	if got := pinWarnings(t, pinGuidelines, pinBrief("Nothing.", "- config key")); len(got) != 0 {
		t.Fatalf("outside the contract: %q", got)
	}
}

func TestPinningNoFileNoSectionNoWarning(t *testing.T) {
	text := pinBrief("A config key.", "- Nothing.")
	if got := pinWarnings(t, "", text); len(got) != 0 {
		t.Fatalf("no file: %q", got)
	}
	if got := pinWarnings(t, "# G\n\n## External formats\n\n- when `config key`: `TestA`\n", text); len(got) != 0 {
		t.Fatalf("no section: %q", got)
	}
}

func TestPinningMalformedLinesIgnoredAndNeverAProblem(t *testing.T) {
	root, path, _ := scratch(t)
	_ = os.MkdirAll(filepath.Join(root, ".vloop"), 0o755)
	g := "## Pinning tests\n\n- when config key: TestBroken\n- when `config key` `TestBroken2`\n- when `config key`: `TestA`, TestBad\n"
	_ = os.WriteFile(filepath.Join(root, GuidelinesFile), []byte(g), 0o644)
	res := run(root, path, pinBrief("A config key.", "- Nothing."))
	for _, l := range res.Lines {
		if strings.Contains(l.Text, "Broken") || strings.Contains(l.Text, "TestBad") {
			t.Fatalf("a malformed rule was read: %v", l)
		}
	}
	root2, path2, _ := scratch(t)
	_ = os.MkdirAll(filepath.Join(root2, ".vloop"), 0o755)
	_ = os.WriteFile(filepath.Join(root2, GuidelinesFile), []byte(pinGuidelines), 0o644)
	r := run(root2, path2, pinBrief("A config key.", "- Nothing."))
	if r.Failed() {
		t.Fatalf("a warning must not fail: %v", r.Lines)
	}
}

func TestPinningSpanishBrief(t *testing.T) {
	root, path, _ := scratch(t)
	_ = os.MkdirAll(filepath.Join(root, ".vloop"), 0o755)
	_ = os.WriteFile(filepath.Join(root, GuidelinesFile), []byte(pinGuidelines), 0o644)
	text := strings.NewReplacer("## Behaviour contract", "## Contrato de comportamiento", "## Constraints", "## Restricciones").
		Replace(pinBrief("Añade una config key.", "- TestA."))
	res := Check(root, parseFit(path, text), SetFor("es"))
	if !has(res.Warnings(), `must-change list: the contract mentions "config key" but Constraints do not name TestB, TestC`) {
		t.Fatalf("%v", res.Lines)
	}
}
