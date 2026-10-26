#!/usr/bin/env bash
# A fresh state/v2 plan of four tasks for /vloop:gate-review. Every gate is
# sound by the existing three checks (it fails on the base at a behavioural
# assertion, reads back what it builds, matches the pinned output); the case
# measures coverage alone. What differs is each task's acceptance criteria:
#
#   T1  a criterion only a check could judge (scripts/lint.sh exits 0), and
#       the plan's one check runs tests/run.sh      -> check-gap
#   T2  a timing on a host the plan never runs on   -> uncovered
#   T3  a criterion requiring a change to .vloop/state/plan.md -> unpassable
#   T4  every criterion covered (gate, check, review)          -> PASS
#
# The scaffold proves the case before it exits 0, and refuses a plan that does
# not measure coverage alone: a planted gate that passes on the base, or fails
# it other than at its behavioural assertion, or fails a correct implementation,
# exits non-zero. Run it in an EMPTY directory; no network, no claude. Needs
# vloop on PATH (the plan is validated with it).
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
. "$HERE/../_lib.sh"
require_empty
git_init
command -v vloop >/dev/null || die "vloop is not on PATH"
mkdir -p docs/briefs bin tests .vloop/state .vloop/tmp
cp "$HERE/brief.md" docs/briefs/runlog-coverage.loop-brief.md
printf '# runlog\n\nA tiny shell command, built by an autonomous loop.\n' >CLAUDE.md
printf '.vloop/tmp/\n' >.gitignore
printf '#!/bin/sh\necho "runlog"\n' >bin/runlog
chmod +x bin/runlog
printf '[display]\ncolor = false\n' >runlog.toml
printf '#!/bin/sh\nsh bin/runlog >/dev/null\n' >tests/run.sh
commit_all "baseline"
python3 - <<'PY'
import json, subprocess
base = subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip()
brief = "docs/briefs/runlog-coverage.loop-brief.md"
scratch = 'd=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") || exit 1; '
def need(cond, msg):
    return f'{cond} || {{ echo "GATE FAIL: {msg}" >&2; exit 1; }}'
gates = [
    ("T1", "Count the files", [
        "bin/runlog count a.log b.log c.log prints 3 files and exits 0",
        "scripts/lint.sh exits 0 on the finished tree",
        "the count is covered by a committed test in tests/ that the tests check runs"],
     'out=$(sh bin/runlog count a.log b.log c.log); ' + need('[ "$out" = "3 files" ]', "count did not print 3 files")),
    ("T2", "Print the check line", [
        "bin/runlog check prints the single line ok runlog and exits 0",
        "bin/runlog check answers in under 50 ms on the operator's production host"],
     'out=$(sh bin/runlog check); ' + need('[ "$out" = "ok runlog" ]', "check did not print ok runlog")),
    ("T3", "Read the unit key", [
        "with unit = \"ms\" at the top level of runlog.toml, bin/runlog unit prints ms; without the key it prints s",
        "the decision to make unit a top-level key is recorded in .vloop/state/plan.md"],
     'g=$PWD/bin/runlog; ' + scratch + 'printf \'unit = "ms"\\n[display]\\ncolor = false\\n\' >"$d/runlog.toml"; '
     'out=$(cd "$d" && sh "$g" unit); rm -rf "$d"; ' + need('[ "$out" = ms ]', "unit did not print ms")),
    ("T4", "Print a usage error", [
        "bin/runlog with no argument prints usage: runlog <command> on stderr, nothing on stdout, and exits 2",
        "a committed test in tests/ asserts the usage line, and the tests check runs it"],
     scratch + 'sh bin/runlog >"$d/out" 2>"$d/err"; rc=$?; err=$(cat "$d/err"); o=$(cat "$d/out"); rm -rf "$d"; '
     + need('[ "$rc" = 2 ] && [ -z "$o" ] && [ "$err" = "usage: runlog <command>" ]', "usage line, stdout or exit code is not the pinned one")),
]
tasks = [{
    "id": tid, "title": title, "kind": "feature", "fixtures": "",
    "goal": f"The brief's contract for runlog: {accept[0]}. One behaviour of the command, so a later task can build on it.",
    "references": [], "depends_on": [], "acceptance": accept,
    "verify": verify, "status": "pending", "attempts": 0, "notes": "",
} for tid, title, accept, verify in gates]
plan = {"schema": "state/v2", "run_id": "eval", "brief": brief, "base": base,
        "branch": "vloop/eval", "status": "running", "iteration": 0,
        "created": "2026-10-25T00:00:00Z", "updated": "2026-10-25T00:00:00Z",
        "shell": "sh",
        "checks": [{"name": "tests", "paths": ["bin/**", "tests/**"], "run": "sh tests/run.sh"}],
        "gate_scratch": [], "gate_review": {"rounds": 0, "verdict": ""}, "tasks": tasks}
json.dump(plan, open(".vloop/state/state.json", "w"), indent=2)
PY
vloop schema validate state/v2 .vloop/state/state.json >/dev/null || die "the plan is not state/v2"

# --- The case's own preconditions: every planted gate is sound. ---
verify_of() { python3 -c 'import json,sys; print(json.load(open(".vloop/state/state.json"))["tasks"][int(sys.argv[1])]["verify"])' "$1"; }
T="$(mktemp -d "${TMPDIR:-/tmp}/coverage.XXXXXX")"
trap 'rm -rf "$T"' EXIT

# A correct implementation of the whole brief, for the copy below.
cat >"$T/ref-runlog" <<'REF'
#!/bin/sh
case "${1:-}" in
count) shift; echo "$# files" ;;
unit) u=$(awk '/^\[/ { exit } /^unit *=/ { gsub(/^[^"]*"|"[^"]*$/, ""); print; exit }' runlog.toml 2>/dev/null); echo "${u:-s}" ;;
check) echo "ok runlog" ;;
"") echo "usage: runlog <command>" >&2; exit 2 ;;
esac
REF

for i in 0 1 2 3; do
  verify_of "$i" >"$T/verify.$i"
  # 1. It fails on the base, and at its behavioural assertion, not early.
  rc=0; sh -c "$(cat "$T/verify.$i")" >"$T/base.$i" 2>&1 || rc=$?
  [ "$rc" != 0 ] || die "T$((i + 1))'s gate passes on the base: it does not measure coverage alone"
  grep -q '^GATE FAIL: ' "$T/base.$i" ||
    die "T$((i + 1))'s gate fails on the base other than at its assertion: $(head -3 "$T/base.$i")"
done
# 2. A correct implementation, in a copy, passes every gate.
cp -R . "$T/copy"
cp "$T/ref-runlog" "$T/copy/bin/runlog"
for i in 0 1 2 3; do
  (cd "$T/copy" && sh -c "$(cat "$T/verify.$i")") >"$T/ref.$i" 2>&1 ||
    die "T$((i + 1))'s gate fails a correct implementation: $(head -3 "$T/ref.$i")"
done
# 3. The planted criteria are what the graders say they are.
python3 - <<'PY' || die "the plan does not plant the coverage criteria"
import json, sys
p = json.load(open(".vloop/state/state.json"))
runs = " ".join(c["run"] for c in p["checks"])
acc = {t["id"]: t["acceptance"] for t in p["tasks"]}
ok = ("lint.sh" not in runs
      and any("scripts/lint.sh" in a for a in acc["T1"])
      and any("50 ms" in a for a in acc["T2"])
      and any(".vloop/state/plan.md" in a for a in acc["T3"]))
sys.exit(0 if ok else 1)
PY

commit_all "plan"
