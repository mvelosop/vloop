package brief

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fmtGuidelines = `# Repository guidelines

## Pinning tests

- when ` + "`eval case`: `TestA`" + `

## External formats

- ` + "`eval case`: `docs/guide/evals.md`" + ` (D20261001-1338)
- eval case: docs/broken.md
- ` + "`eval case` `docs/broken2.md`" + `
`

const fmtWarning = `external format: the contract mentions "eval case" but Binding references do not cite docs/guide/evals.md (.vloop/repo-guidelines.md)`

// fmtBrief is pinBrief with a Binding references section holding refs.
func fmtBrief(contract, refs, constraints string) string {
	return strings.Replace(pinBrief(contract, constraints), "## Behaviour contract", "## Binding references\n\n"+refs+"\n\n## Behaviour contract", 1)
}

func fmtWarnings(t *testing.T, guidelines, text string) []string {
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
	res := run(root, path, text)
	if res.Failed() && guidelines != "" {
		for _, l := range res.Lines {
			if strings.HasPrefix(l.Text, "external format:") && l.Marker != Warning {
				t.Fatalf("an external format line is not a warning: %v", l)
			}
		}
	}
	var out []string
	for _, w := range res.Warnings() {
		if strings.HasPrefix(w, "external format:") {
			out = append(out, w)
		}
	}
	return out
}

func TestExternalFormatWarnsWhenNotCited(t *testing.T) {
	got := fmtWarnings(t, fmtGuidelines, fmtBrief("It adds a new eval case.", "- `docs/other.md` — x", "- Nothing."))
	if len(got) != 1 || got[0] != fmtWarning {
		t.Fatalf("got %q", got)
	}
}

func TestExternalFormatCitedInBindingReferences(t *testing.T) {
	got := fmtWarnings(t, fmtGuidelines, fmtBrief("It adds a new eval case.", "- `docs/guide/evals.md` — the format", "- Nothing."))
	if len(got) != 0 {
		t.Fatalf("got %q", got)
	}
}

func TestExternalFormatPathInAnotherSectionDoesNotCount(t *testing.T) {
	got := fmtWarnings(t, fmtGuidelines, fmtBrief("It adds a new eval case.", "- `docs/other.md` — x", "- Follow `docs/guide/evals.md`."))
	if len(got) != 1 || got[0] != fmtWarning {
		t.Fatalf("got %q", got)
	}
}

func TestExternalFormatPhraseOutsideContract(t *testing.T) {
	got := fmtWarnings(t, fmtGuidelines, fmtBrief("Nothing here.", "- `docs/other.md` — x", "- an eval case"))
	if len(got) != 0 {
		t.Fatalf("got %q", got)
	}
}

func TestExternalFormatCaseInsensitiveAndSubsections(t *testing.T) {
	got := fmtWarnings(t, fmtGuidelines, fmtBrief("### L1\n\nA new EVAL Case.", "- `docs/other.md` — x", "- Nothing."))
	if len(got) != 1 || got[0] != fmtWarning {
		t.Fatalf("got %q", got)
	}
}

func TestExternalFormatNoFileSameOutputAndNeverAProblem(t *testing.T) {
	text := fmtBrief("A new eval case.", "- `docs/other.md` — x", "- Nothing.")
	if got := fmtWarnings(t, "", text); len(got) != 0 {
		t.Fatalf("no file: %q", got)
	}
	root, path, _ := scratch(t)
	without := run(root, path, text)
	_ = os.MkdirAll(filepath.Join(root, ".vloop"), 0o755)
	_ = os.WriteFile(filepath.Join(root, GuidelinesFile), []byte(fmtGuidelines), 0o644)
	with := run(root, path, text)
	if with.Failed() != without.Failed() || len(with.Problems()) != len(without.Problems()) {
		t.Fatalf("a warning changed the verdict: %v", with.Lines)
	}
}
