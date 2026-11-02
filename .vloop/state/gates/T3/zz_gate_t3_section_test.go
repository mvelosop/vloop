package install

// The T3 gate's oracle (B28, K3): the section names the two brief skills, and
// an older section between the markers is replaced by it, everything outside
// the markers kept. Copied into a copy of the tree by the gate; never
// committed. GATE_SECTION_WANT is the gate's section.want.

import (
	"os"
	"strings"
	"testing"
)

const gateOldParagraph = `In an interactive session you are the operator's hands: follow the operator's playbook (/vloop:operate), not these session rules. If /vloop:operate is not available, the vloop plugin is not loaded: say so, give the user the two ways to load it (claude plugin install vloop@vloop after claude plugin marketplace add mvelosop/vloop, or claude --plugin-dir "$(vloop plugin path)"), and meanwhile work from vloop --help.`

const gateNewParagraph = `In an interactive session you are the operator's hands: follow the operator's playbook (/vloop:operate), not these session rules. To write a brief, follow /vloop:create-brief; to review one, run /vloop:review-brief in a separate session. If /vloop:operate is not available, the vloop plugin is not loaded: say so, give the user the two ways to load it (claude plugin install vloop@vloop after claude plugin marketplace add mvelosop/vloop, or claude --plugin-dir "$(vloop plugin path)"), and meanwhile work from vloop --help.`

func gateWant(t *testing.T, version string) string {
	t.Helper()
	b, err := os.ReadFile(os.Getenv("GATE_SECTION_WANT"))
	if err != nil {
		t.Fatal(err)
	}
	return strings.ReplaceAll(string(b), "@VERSION@", version)
}

func TestGateT3Section(t *testing.T) {
	if got, want := Section("9.9.9"), gateWant(t, "9.9.9"); got != want {
		t.Errorf("Section(\"9.9.9\") =\n%s\nwant\n%s", got, want)
	}
}

func TestGateT3UpgradeReplacesOldSection(t *testing.T) {
	old := strings.Replace(gateWant(t, "0.14.0"), gateNewParagraph, gateOldParagraph, 1)
	if old == gateWant(t, "0.14.0") {
		t.Fatal("the gate's old section was not built")
	}
	before := "# My project\n\nKeep this line.\n\n"
	after := "\n## Mine\n\nAnd this one.\n"
	existing := before + strings.TrimSuffix(old, "\n") + after
	got := string(MergeClaudeMD([]byte(existing), true, "9.9.9"))
	want := before + strings.TrimSuffix(gateWant(t, "9.9.9"), "\n") + after
	if got != want {
		t.Errorf("MergeClaudeMD of an older section =\n%s\nwant\n%s", got, want)
	}
	if strings.Contains(got, gateOldParagraph) {
		t.Error("the old interactive-session paragraph survives the merge")
	}
}
