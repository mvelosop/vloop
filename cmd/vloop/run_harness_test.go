package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// The run harness: a vloop-initialised scratch repository on main, a fake home
// that trusts it, and a stub claude on PATH that plays a per-test script chosen
// by the prompt it receives. It replaces .loop/tests/lib.sh. Everything lives
// in temporary directories; the binary is the one TestMain built (binPath).

const runBriefName = "B20260101-0900-demo.loop-brief"

// runStubProlog is the stub claude. It logs its argv, answers the two probes
// doctor makes, and for -p prompts sets the variables below and sources the
// per-test script:
//
//	PHASE  plan|work|review|gate-review      TASK   the task id (work, review), else empty
//	ARG    the prompt's argument (the brief path for plan, the task id otherwise)
//	MODEL  the --model it was given    ATTEMPT  1-based count of this PHASE+ARG so far
//	STUB_COST, STUB_EXIT, STUB_DURATION, STUB_TURNS, STUB_ERROR  the result JSON and exit code
//	STUB_SILENT=1  print no result JSON
//
// The script runs in the stub's working directory (the repository) and may do
// anything a shell can: write the plan, files, .vloop/tmp/proposal.json or
// verdict.json (a gate review's passing gate-verdict.json is written before the script), run git, set the variables above, or exit.
const runStubSource = `#!/bin/sh
dir=$(dirname "$0")
echo "$*" >> "$dir/argv.log"
case "$1" in
  --version) echo "2.0.0 (Claude Code)"; exit 0;;
  plugin) echo '[{"id":"vloop@vloop","version":"0.8.0","enabled":true}]'; exit 0;;
esac
prompt=""; MODEL=""
while [ $# -gt 0 ]; do
  case "$1" in
    -p) prompt="$2"; shift 2;;
    --model) MODEL="$2"; shift 2;;
    *) shift;;
  esac
done
cmd="${prompt%% *}"
ARG="${prompt#* }"
case "$cmd" in
  /vloop:plan) PHASE=plan; TASK="";;
  /vloop:work) PHASE=work; TASK="$ARG";;
  /vloop:review) PHASE=review; TASK="$ARG";;
  /vloop:gate-review) PHASE=gate-review; TASK=""; ARG="";;
  *) echo "stub claude: unrecognised prompt: $prompt" >&2; exit 64;;
esac
n=$(cat "$dir/count.$PHASE.$(echo "$ARG" | tr '/ ' '__')" 2>/dev/null || echo 0)
ATTEMPT=$((n + 1))
echo "$ATTEMPT" > "$dir/count.$PHASE.$(echo "$ARG" | tr '/ ' '__')"
STUB_COST=0.10; STUB_EXIT=0; STUB_DURATION=1000; STUB_TURNS=3; STUB_ERROR=false; STUB_SILENT=0
if [ "$PHASE" = gate-review ]; then
  mkdir -p .vloop/tmp
  echo '{"schema":"gate-verdict/v1","verdict":"PASS","tasks":[],"notes":"none"}' > .vloop/tmp/gate-verdict.json
fi
if [ -f "$dir/script.sh" ]; then . "$dir/script.sh"; fi
if [ "$STUB_SILENT" != 1 ]; then
  printf '{"type":"result","subtype":"success","is_error":%s,"duration_ms":%s,"num_turns":%s,"total_cost_usd":%s,"session_id":"stub","permission_denials":[],"modelUsage":{"%s":{"costUSD":%s}}}\n' \
    "$STUB_ERROR" "$STUB_DURATION" "$STUB_TURNS" "$STUB_COST" "$MODEL" "$STUB_COST"
fi
exit "$STUB_EXIT"
`

// runRepo is a harness repository: the scratch repository, its fake home and
// the stub directory.
type runRepo struct {
	t    *testing.T
	dir  string
	home string
	stub string
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(cleanEnv(t.TempDir()), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1", "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

// runBriefText is a ready loop brief that passes `vloop brief check`.
func runBriefText() string {
	return strings.Join([]string{
		"---", "name: " + runBriefName, "description: Build a demo", "kind: brief", "status: ready",
		"created: 2026-01-01", "seeds: A run that builds a demo", "depends-on: []", "---",
		"# Brief", "", "## What it is", "", "A thing.", "",
		"## Why this shape, and what was rejected", "", "Because.", "",
		"## Binding references", "", "- `docs/x.md` — the error contract", "",
		"## Behaviour contract", "", "An unknown id exits 1.", "",
		"## Worked example", "", "```", "$ demo", "ok", "```", "",
		"## Out of scope", "", "- one", "- two", "",
		"## Constraints", "", "- Go.", "",
		"## Shape", "", "2 to 3 tasks.", "",
	}, "\n")
}

// newRunRepo builds the harness repository: git on main with an identity in
// the repository's own config, vloop init, a ready brief, all committed; a fake
// home trusting it; the stub claude (no script yet). It skips where the stub
// cannot run.
// runCheckConfig is the config a run repository carries: vloop 2 refuses to
// run a repository with no [[check]].
const runCheckConfig = "[[check]]\nname = \"all\"\npaths = [\"**\"]\nrun = \"true\"\n"

func newRunRepo(t *testing.T) *runRepo {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the stub claude is a POSIX shell script and cannot run on windows")
	}
	// The home's base name is the user name the driver masks and the
	// containment test hunts for; t.TempDir's base names are counters such as
	// 002, which a date-stamped run id contains on 2 October.
	r := &runRepo{t: t, dir: t.TempDir(), home: filepath.Join(t.TempDir(), "harnessuser"), stub: t.TempDir()}
	if err := os.Mkdir(r.home, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, r.dir, "init", "-q", "-b", "main")
	runGit(t, r.dir, "config", "user.name", "harness")
	runGit(t, r.dir, "config", "user.email", "harness@example.com")

	real, err := filepath.EvalSymlinks(r.dir)
	if err != nil {
		t.Fatal(err)
	}
	trust := map[string]any{}
	for _, k := range []string{r.dir, real} {
		trust[k] = map[string]any{"hasTrustDialogAccepted": true}
	}
	raw, _ := json.Marshal(map[string]any{"projects": trust})
	if err := os.WriteFile(filepath.Join(r.home, ".claude.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(r.stub, "claude"), []byte(runStubSource), 0o755); err != nil {
		t.Fatal(err)
	}

	r.write("docs/x.md", "x\n")
	if res := r.vloop("init"); res.code != 0 {
		t.Fatalf("vloop init: %+v", res)
	}
	r.write(".vloop/config.toml", runCheckConfig)
	r.write("docs/briefs/"+runBriefName+".md", runBriefText())
	runGit(t, r.dir, "add", "-A")
	runGit(t, r.dir, "commit", "-q", "-m", "harness: init and brief")
	return r
}

func (r *runRepo) write(rel, content string) {
	r.t.Helper()
	p := filepath.Join(r.dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		r.t.Fatal(err)
	}
}

func (r *runRepo) read(rel string) string {
	r.t.Helper()
	b, err := os.ReadFile(filepath.Join(r.dir, filepath.FromSlash(rel)))
	if err != nil {
		r.t.Fatal(err)
	}
	return string(b)
}

func (r *runRepo) git(args ...string) string {
	r.t.Helper()
	return runGit(r.t, r.dir, args...)
}

// script sets the stub's per-test script: shell run for every -p prompt, with
// the variables documented on runStubSource. Call it again to replace it.
func (r *runRepo) script(body string) {
	r.t.Helper()
	if err := os.WriteFile(filepath.Join(r.stub, "script.sh"), []byte(body), 0o644); err != nil {
		r.t.Fatal(err)
	}
}

// argv is the stub's log: one line per invocation, arguments joined by spaces.
func (r *runRepo) argv() []string {
	r.t.Helper()
	b, err := os.ReadFile(filepath.Join(r.stub, "argv.log"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		r.t.Fatal(err)
	}
	return strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
}

// env is the environment for every binary or stub invocation.
func (r *runRepo) env(extra ...string) []string {
	return append(append(cleanEnv(r.home),
		"PATH="+r.stub+string(os.PathListSeparator)+os.Getenv("PATH"),
		"GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1", "GIT_TERMINAL_PROMPT=0"), extra...)
}

func (r *runRepo) exec(name string, env []string, args ...string) result {
	r.t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = r.dir
	cmd.Env = env
	var o, e bytes.Buffer
	cmd.Stdout, cmd.Stderr = &o, &e
	code := 0
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			r.t.Fatalf("run %s %v: %v", name, args, err)
		}
		code = ee.ExitCode()
	}
	return result{o.String(), e.String(), code}
}

// vloop runs the built binary in the repository with the stub first on PATH.
func (r *runRepo) vloop(args ...string) result {
	r.t.Helper()
	return r.exec(binPath, r.env(), args...)
}

// claude runs the stub directly, as the driver would.
func (r *runRepo) claude(args ...string) result {
	r.t.Helper()
	return r.exec(filepath.Join(r.stub, "claude"), r.env(), args...)
}

func TestRunHarness(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)

	if b := strings.TrimSpace(r.git("rev-parse", "--abbrev-ref", "HEAD")); b != "main" {
		t.Fatalf("branch = %q, want main", b)
	}
	if s := r.git("status", "--porcelain"); s != "" {
		t.Fatalf("scratch repository is not committed: %q", s)
	}
	if u := strings.TrimSpace(r.git("config", "--local", "user.email")); u == "" {
		t.Fatal("no identity in the repository's own config")
	}
	if chk := r.vloop("brief", "check", "docs/briefs/"+runBriefName+".md"); chk.code != 0 || !strings.Contains(chk.out, "briefs ok") {
		t.Fatalf("the harness brief is not ready: %+v", chk)
	}

	doc := r.vloop("doctor")
	if doc.code != 0 || !strings.Contains(doc.out, "doctor: 0 problem(s)") {
		t.Fatalf("doctor: %+v", doc)
	}

	// A script that varies by attempt: the first work attempt writes T1.out.
	r.script(`case "$PHASE:$TASK:$ATTEMPT" in
  work:T1:1) echo done > T1.out; STUB_COST=0.25;;
  work:T1:*) echo again > T1.again;;
esac
`)
	run := r.claude("-p", "/vloop:work T1", "--model", "sonnet")
	if run.code != 0 {
		t.Fatalf("stub: %+v", run)
	}
	if got := r.read("T1.out"); got != "done\n" {
		t.Fatalf("scripted action did not run: %q", got)
	}
	var res struct {
		Cost   float64        `json:"total_cost_usd"`
		Dur    int            `json:"duration_ms"`
		Turns  int            `json:"num_turns"`
		IsErr  bool           `json:"is_error"`
		Denied []any          `json:"permission_denials"`
		Usage  map[string]any `json:"modelUsage"`
	}
	if err := json.Unmarshal([]byte(run.out), &res); err != nil {
		t.Fatalf("stub printed %q: %v", run.out, err)
	}
	if _, ok := res.Usage["sonnet"]; !ok || len(res.Usage) != 1 || res.Cost != 0.25 || res.Turns != 3 || res.IsErr {
		t.Fatalf("result = %+v", res)
	}

	// The second attempt takes the other arm; a silent, failing stub prints nothing.
	r.claude("-p", "/vloop:work T1", "--model", "sonnet")
	if r.read("T1.again") != "again\n" {
		t.Fatal("the script did not see ATTEMPT=2")
	}
	r.script("STUB_SILENT=1; STUB_EXIT=3\n")
	if res := r.claude("-p", "/vloop:review T1", "--model", "opus"); res.code != 3 || res.out != "" {
		t.Fatalf("silent stub: %+v", res)
	}

	log := r.argv()
	if len(log) < 5 || !strings.Contains(strings.Join(log, "\n"), "-p /vloop:work T1 --model sonnet") {
		t.Fatalf("argument log: %q", log)
	}
	if r.exec("sh", r.env(), "-c", "claude --version").out == "" {
		t.Fatal("the stub is not first on PATH")
	}
	if ents, _ := os.ReadDir(r.home); len(ents) != 1 {
		t.Fatalf("the fake home holds more than .claude.json: %v", ents)
	}
}
