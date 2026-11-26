package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mvelosop/vloop/internal/driver"
	"github.com/mvelosop/vloop/internal/state"
)

type b40DoctorDocument struct {
	Checks []struct {
		Name    string `json:"name"`
		Result  string `json:"result"`
		Message string `json:"message"`
	} `json:"checks"`
}

// TestWorkedExampleB40 plays the mixed-team worked example through the built
// CLI, without a platform-specific stub agent.
func TestWorkedExampleB40(t *testing.T) {
	marker := b40StubAgent(t)
	s := newScratch(t)
	if r := s.run(nil, "init"); r.code != 0 {
		t.Fatalf("init: %+v", r)
	}
	s.write(".vloop/config.toml", strings.TrimSpace(s.read(".vloop/config.toml"))+
		"\n\n[[check]]\nname = \"go\"\npaths = [\"**\"]\nrun = \"go version\"\n")

	resolved := "sh"
	if runtime.GOOS == "windows" {
		resolved = "pwsh"
	}
	b40NeedDoctorRow(t, s, nil, "checks", "warning",
		"shell is not set in .vloop/config.toml, so each machine runs the checks in its own ("+resolved+" here): set it there")
	b40NeedDoctorRow(t, s, []string{"VLOOP_SHELL=bash"}, "checks", "warning",
		"shell is not set in .vloop/config.toml, so each machine runs the checks in its own (bash here): set it there")

	expect(t, s.run(nil, "config", "set", "platforms", "windows,macos"), 0, "", "")
	expect(t, s.run(nil, "config", "get", "platforms"), 0, "windows,macos\n", "")
	b40NeedDoctorRow(t, s, nil, "platforms", "problem",
		`macos, windows need shell = "pwsh" in .vloop/config.toml; it is not set`)

	expect(t, s.run(nil, "config", "set", "shell", "pwsh"), 0, "", "")
	if runtime.GOOS == "linux" {
		b40NeedDoctorRow(t, s, nil, "platforms", "warning",
			"this machine (linux) is not one of macos, windows")
	} else {
		b40NeedDoctorRow(t, s, nil, "platforms", "pass", "macos, windows")
	}

	expect(t, s.run(nil, "config", "set", "platforms", "windows,macos,windows"), 0, "", "")
	expect(t, s.run(nil, "config", "get", "platforms"), 0, "windows,macos,windows\n", "")
	if runtime.GOOS == "linux" {
		b40NeedDoctorRow(t, s, nil, "platforms", "warning",
			"this machine (linux) is not one of macos, windows")
	} else {
		b40NeedDoctorRow(t, s, nil, "platforms", "pass", "macos, windows")
	}

	expect(t, s.run([]string{"VLOOP_SHELL=pwsh"}, "config", "get", "shell"), 1, "",
		`vloop: VLOOP_SHELL: shell belongs in .vloop/config.toml while platforms lists macos, windows — the checks and the gates are written for one shell`+"\n")
	expect(t, s.run([]string{"VLOOP_PLATFORMS=linux"}, "config", "get", "platforms"), 1, "",
		"vloop: VLOOP_PLATFORMS: platforms belongs in .vloop/config.toml — the platforms are the repository's, not one machine's\n")
	expect(t, s.run(nil, "config", "set", "--local", "shell", "pwsh"), 2, "",
		`vloop: .vloop/config.local.toml: shell belongs in .vloop/config.toml while platforms lists macos, windows — the checks and the gates are written for one shell`+"\n")
	expect(t, s.run(nil, "config", "set", "--local", "platforms", "linux"), 2, "",
		"vloop: .vloop/config.local.toml: platforms belongs in .vloop/config.toml — the platforms are the repository's, not one machine's\n")
	expect(t, s.run(nil, "config", "set", "platforms", "macos,beos"), 2, "",
		`vloop: invalid value "macos,beos" for platforms: want a comma-separated list of linux, macos, windows`+"\n")
	expect(t, s.run(nil, "config", "get", "platforms"), 0, "windows,macos,windows\n", "")

	const brief = "docs/briefs/B20260101-0000-gate.loop-brief.md"
	s.write(brief, "---\nname: B20260101-0000-gate.loop-brief\nstatus: ready\n---\n# Brief\n\n## Shape\n\n1 to 1 task.\n")
	s.write(".vloop/config.toml",
		"platforms = [\"windows\", \"macos\", \"windows\"]\n\n[[check]]\nname = \"go\"\npaths = [\"**\"]\nrun = \"go test ./... | grep ok\"\n")
	b40Git(t, s, "config", "user.name", "test")
	b40Git(t, s, "config", "user.email", "test@example.com")
	b40Git(t, s, "config", "commit.gpgsign", "false")
	b40Git(t, s, "add", "-A")
	b40Git(t, s, "commit", "-q", "-m", "setup mixed-team example")
	head := strings.TrimSpace(b40Git(t, s, "rev-parse", "HEAD"))
	branch := strings.TrimSpace(b40Git(t, s, "rev-parse", "--abbrev-ref", "HEAD"))

	run := s.run(nil, "run", "--plan-only", brief)
	wantExit(t, run, 1)
	wantIn(t, "run stderr", run.err,
		`  ✗ platforms macos, windows need shell = "pwsh" in .vloop/config.toml; it is not set`+"\n",
		"  ! checks go calls grep, which is not on every platform in platforms (macos, windows)\n",
		"vloop: preflight failed")
	if s.readPathExists(".vloop/state/state.json") || s.readPathExists(".vloop/state/runs") {
		t.Error("run started a plan session despite the platforms preflight refusal")
	}
	if got := strings.TrimSpace(b40Git(t, s, "rev-parse", "HEAD")); got != head {
		t.Error("the refused run created a commit")
	}
	if got := strings.TrimSpace(b40Git(t, s, "rev-parse", "--abbrev-ref", "HEAD")); got != branch {
		t.Error("the refused run changed branches")
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Error("the refused run started an agent")
	}
}

// TestNoPlatformsRegressionViolationGuard names the brief's compatibility
// violation: repos without platforms must retain doctor, run and plan behavior.
func TestNoPlatformsRegressionViolationGuard(t *testing.T) {
	marker := b40StubAgent(t)
	s := newScratch(t)
	s.write(".vloop/config.toml",
		"shell = \"pwsh\"\n\n[[check]]\nname = \"go\"\npaths = [\"**\"]\nrun = \"go test ./... | grep ok\"\n")

	doc := b40Doctor(t, s, nil)
	for _, row := range doc.Checks {
		if row.Name == "platforms" || row.Name == "checks" && strings.Contains(row.Message, "not on every platform") {
			t.Fatalf("no-platforms doctor gained a mixed-team finding: %#v", row)
		}
	}

	const brief = "docs/briefs/B20260101-0000-no-platforms.loop-brief.md"
	s.write(brief, "---\nname: B20260101-0000-no-platforms.loop-brief\nstatus: ready\n---\n# Brief\n")
	run := s.run(nil, "run", "--plan-only", brief)
	wantNotIn(t, "no-platforms run stderr", run.err, "  ✗ platforms ", "  ! platforms ", "not on every platform")
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Error("the no-platforms regression guard started a real or scripted agent")
	}

	root := t.TempDir()
	s.writeIn(root, ".vloop/config.toml", "shell = \"pwsh\"\n")
	s.writeIn(root, ".vloop/state/gates/T1/gate.ps1", "grep value input\n")
	plan := &state.Plan{Tasks: []state.Task{{ID: "T1", Verify: `sh gate.sh | grep value`}}}
	if got := driver.GateShapeProblems(root, plan); len(got) != 0 {
		t.Fatalf("no-platforms plan acceptance changed: %q", got)
	}
}

func b40Doctor(t *testing.T, s *scratch, env []string) b40DoctorDocument {
	t.Helper()
	r := s.run(env, "doctor", "--json")
	var doc b40DoctorDocument
	if err := json.Unmarshal([]byte(r.out), &doc); err != nil {
		t.Fatalf("decode doctor: %v\nstdout: %s\nstderr: %s", err, r.out, r.err)
	}
	return doc
}

func b40NeedDoctorRow(t *testing.T, s *scratch, env []string, name, result, message string) {
	t.Helper()
	doc := b40Doctor(t, s, env)
	for _, row := range doc.Checks {
		if row.Name == name && row.Result == result && row.Message == message {
			return
		}
	}
	t.Fatalf("missing %s/%s %q in %#v", name, result, message, doc.Checks)
}

func b40Git(t *testing.T, s *scratch, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", s.dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func b40StubAgent(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	src := filepath.Join(dir, "main.go")
	source := `package main
import (
	"fmt"
	"os"
)
func main() {
	if len(os.Args) >= 2 && os.Args[1] == "--version" {
		fmt.Print("2.0.0 (Claude Code)")
		return
	}
	if len(os.Args) >= 2 && os.Args[1] == "plugin" {
		fmt.Print("[]")
		return
	}
	for _, arg := range os.Args[1:] {
		if arg == "-p" {
			if marker := os.Getenv("B40_AGENT_MARKER"); marker != "" {
				_ = os.WriteFile(marker, []byte("started"), 0o644)
			}
			break
		}
	}
	os.Exit(99)
}
`
	if err := os.WriteFile(src, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	name := "claude"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	if out, err := exec.Command("go", "build", "-o", filepath.Join(dir, name), src).CombinedOutput(); err != nil {
		t.Fatalf("build agent stub: %v\n%s", err, out)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	marker := filepath.Join(dir, "agent-started")
	t.Setenv("B40_AGENT_MARKER", marker)
	return marker
}

func (s *scratch) readPathExists(rel string) bool {
	_, err := os.Stat(filepath.Join(s.dir, filepath.FromSlash(rel)))
	return err == nil
}

func (s *scratch) writeIn(root, rel, content string) {
	s.t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		s.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		s.t.Fatal(err)
	}
}
