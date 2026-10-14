#!/usr/bin/env bash
# A fresh state/v2 plan of seven gates for /vloop:gate-review: six that fooled
# the gate review in past runs, and one sound. Run it in an EMPTY directory; no
# network, no claude. Each defect is reduced to its mechanism:
#
#   T1  GNU-only grep pattern (grep -P)               -> unpassable (BSD grep, macOS)
#   T2  top-level TOML key appended after a table     -> unpassable
#   T3  a session must write under .vloop/            -> unpassable (S-4)
#   T4  pattern with a trailing space, no match       -> unpassable
#   T5  a session/v1 record validated as session/v2   -> unpassable
#   T6  *-review.json also matches *-gate-review.json -> unpassable
#   T7  sound: exact stdout, stderr and exit code     -> PASS
#
# FOR macOS HOSTS: T1 reproduces only with BSD grep, which rejects -P. On a host
# whose grep accepts it the scaffold exits non-zero, because the case would no
# longer plant that defect.
#
# The scaffold proves each gate's defect before it exits 0: on the base every
# gate fails, for its recorded reason; and a correct implementation, installed
# in a copy, still fails the six defective gates and passes the sound one.
# Needs vloop on PATH (T5's gate and its precondition run it).
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
. "$HERE/../_lib.sh"
require_empty
git_init
command -v vloop >/dev/null || die "vloop is not on PATH"
mkdir -p docs/briefs bin records .vloop/state .vloop/tmp
cp "$HERE/brief.md" docs/briefs/runlog-commands.loop-brief.md
printf '# runlog\n\nA tiny shell command, built by an autonomous loop.\n' >CLAUDE.md
printf '.vloop/tmp/\n' >.gitignore
printf '#!/bin/sh\necho "runlog"\n' >bin/runlog
chmod +x bin/runlog
printf '[display]\ncolor = false\n' >runlog.toml
printf '{"seconds": 30}\n' >records/001-work.json
printf '{"seconds": 12}\n' >records/002-review.json
printf '{"seconds": 7}\n' >records/003-gate-review.json
printf '{"seconds": 9}\n' >records/004-review.json
commit_all "baseline"
python3 - <<'PY'
import json, os, subprocess
base = subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip()
brief = "docs/briefs/runlog-commands.loop-brief.md"
scratch = 'd=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") || exit 1; '
gates = [
    ("T1", "Count the files", "bin/runlog count a.log b.log c.log prints exactly 3 files and exits 0",
     "sh bin/runlog count a.log b.log c.log | grep -qP '^\\d+ files$'"),
    ("T2", "Read the unit key", "with unit = \"ms\" at the top level of runlog.toml, bin/runlog unit prints ms",
     "g=$PWD/bin/runlog; " + scratch + "cp runlog.toml \"$d/runlog.toml\"; printf 'unit = \"ms\"\\n' >>\"$d/runlog.toml\"; "
     "out=$(cd \"$d\" && sh \"$g\" unit); rm -rf \"$d\"; [ \"$out\" = ms ]"),
    ("T3", "Record the unit decision", "the decision to make unit a top-level key is recorded in .vloop/interventions/I-runlog-units.md, which names unit",
     "grep -q unit .vloop/interventions/I-runlog-units.md"),
    ("T4", "Print the check line", "bin/runlog check prints the single line ok runlog and exits 0",
     "sh bin/runlog check | grep -q '^ok runlog '"),
    ("T5", "Print a session record", "bin/runlog record prints one session/v1 record as JSON",
     scratch + "sh bin/runlog record >\"$d/rec.json\"; vloop schema validate session/v2 \"$d/rec.json\"; r=$?; rm -rf \"$d\"; exit $r"),
    ("T6", "Total the review time", "bin/runlog review-seconds prints the sum of seconds over the review sessions in records/, NNN-review.json, and not NNN-gate-review.json",
     "want=$(cat records/*-review.json | sed -n 's/.*\"seconds\": *\\([0-9]*\\).*/\\1/p' | awk '{ s += $1 } END { print s + 0 }'); "
     "[ \"$(sh bin/runlog review-seconds)\" = \"$want\" ]"),
    ("T7", "Print a usage error", "bin/runlog with no argument prints usage: runlog <command> on stderr, nothing on stdout, and exits 2",
     scratch + "sh bin/runlog >\"$d/out\" 2>\"$d/err\"; rc=$?; "
     "[ \"$rc\" = 2 ] && [ ! -s \"$d/out\" ] && [ \"$(cat \"$d/err\")\" = 'usage: runlog <command>' ]; r=$?; rm -rf \"$d\"; exit $r"),
]
tasks = [{
    "id": tid, "title": title, "kind": "docs" if tid == "T3" else "feature", "fixtures": "",
    "goal": f"The brief's contract for runlog: {accept}. One behaviour of the command, so a later task can build on it.",
    "references": [], "depends_on": [], "acceptance": [accept],
    "verify": verify, "status": "pending", "attempts": 0, "notes": "",
} for tid, title, accept, verify in gates]
plan = {"schema": "state/v2", "run_id": "eval", "brief": brief, "base": base,
        "branch": "vloop/eval", "status": "running", "iteration": 0,
        "created": "2026-10-12T00:00:00Z", "updated": "2026-10-12T00:00:00Z",
        "shell": "sh", "checks": [], "gate_scratch": [],
        "gate_review": {"rounds": 0, "verdict": ""}, "tasks": tasks}
json.dump(plan, open(".vloop/state/state.json", "w"), indent=2)
# The driver's base logs: what each gate printed on the tree as it is now.
logs = ".vloop/state/runs/eval/gates"
os.makedirs(logs)
for t in tasks:
    with open(f"{logs}/base-{t['id']}.log", "w") as f:
        r = subprocess.run(["sh", "-c", t["verify"]], stdout=f, stderr=subprocess.STDOUT)
        f.write(f"exit {r.returncode}\n")
PY

# --- The case's own preconditions: it must plant what it says it plants. ---
verify_of() { python3 -c 'import json,sys; print(json.load(open(".vloop/state/state.json"))["tasks"][int(sys.argv[1])]["verify"])' "$1"; }
T="$(mktemp -d "${TMPDIR:-/tmp}/real-misses.XXXXXX")"
trap 'rm -rf "$T"' EXIT

# A correct implementation of the whole brief, for the copy below.
cat >"$T/ref-runlog" <<'REF'
#!/bin/sh
case "${1:-}" in
count) shift; echo "$# files" ;;
unit) u=$(awk '/^\[/ { exit } /^unit *=/ { gsub(/^[^"]*"|"[^"]*$/, ""); print; exit }' runlog.toml 2>/dev/null); echo "${u:-s}" ;;
check) echo "ok runlog" ;;
record) echo '{"schema":"session/v1","run_id":"r","iteration":1,"phase":"work","task":"T1","model":"sonnet","effort":null,"models_used":{},"started":"2026-01-01T09:00:00Z","duration_ms":10,"cost_usd":0.5,"turns":1,"is_error":false,"permission_denials":[]}' ;;
review-seconds) cat records/[0-9][0-9][0-9]-review.json | sed -n 's/.*"seconds": *\([0-9]*\).*/\1/p' | awk '{ s += $1 } END { print s + 0 }' ;;
"") echo "usage: runlog <command>" >&2; exit 2 ;;
esac
REF

# 1. Every gate fails on the base.
for i in 0 1 2 3 4 5 6; do
  verify_of "$i" >"$T/verify.$i"
  sh -c "$(cat "$T/verify.$i")" >"$T/base.$i" 2>&1 && die "T$((i + 1))'s gate passes on the base"
done

# 2. A correct implementation, in a copy: the six defective gates still fail,
#    the sound one passes.
cp -R . "$T/copy"
cp "$T/ref-runlog" "$T/copy/bin/runlog"
for i in 0 1 2 3 4 5; do
  (cd "$T/copy" && sh -c "$(cat "$T/verify.$i")") >/dev/null 2>&1 &&
    die "T$((i + 1))'s gate passes a correct implementation: it is not unpassable"
done
(cd "$T/copy" && sh -c "$(cat "$T/verify.6")") >/dev/null 2>&1 ||
  die "T7's sound gate fails a correct implementation"

# 3. Each gate fails for its recorded reason.
# T1: BSD grep rejects -P with an error (status 2), where GNU grep matches.
rc=0; printf '3 files\n' | grep -qP '^\d+ files$' 2>"$T/grep.err" || rc=$?
[ "$rc" = 2 ] || die "this host's grep does not reject -P (status $rc): T1 reproduces only with BSD grep, on a macOS host"
grep -qi 'invalid\|unknown\|usage' "$T/grep.err" || die "grep -P failed without an option error: $(cat "$T/grep.err")"
# T2: the appended key lands after the [display] table header.
cp runlog.toml "$T/runlog.toml"; printf 'unit = "ms"\n' >>"$T/runlog.toml"
awk '/^\[/ { t = NR } /^unit *=/ { u = NR } END { exit !(t && u > t) }' "$T/runlog.toml" ||
  die "T2's appended key is not after the table header"
# ...and the reference implementation reads it at the top level, as the brief says.
printf 'unit = "ms"\n[display]\ncolor = false\n' >"$T/top.toml"
[ "$(cd "$T" && cp top.toml runlog.toml && sh ref-runlog unit)" = ms ] || die "the reference implementation misreads a top-level key"
# T3: it needs a file under .vloop/ outside .vloop/tmp/, which the base lacks.
grep -q '\.vloop/interventions/' "$T/verify.2" && ! grep -q '\.vloop/tmp/' "$T/verify.2" ||
  die "T3's gate does not read under .vloop/ outside .vloop/tmp/"
[ ! -e .vloop/interventions ] || die "the base already has .vloop/interventions"
# T4: the pattern's trailing space cannot match the brief's exact output.
rc=0; printf 'ok runlog\n' | grep -q '^ok runlog ' || rc=$?
[ "$rc" = 1 ] || die "T4's pattern does not fail to match the pinned output (status $rc)"
# T5: a session/v1 record is rejected by session/v2 and accepted by session/v1.
sh "$T/ref-runlog" record >"$T/rec.json" 2>/dev/null || true
vloop schema validate session/v1 "$T/rec.json" >/dev/null 2>&1 || die "the reference record is not session/v1"
vloop schema validate session/v2 "$T/rec.json" >/dev/null 2>&1 && die "session/v2 accepts a session/v1 record"
# T6: the glob matches the gate-review record too, so the gate's sum is not the
# brief's.
[ "$(ls records/*-review.json | wc -l | tr -d ' ')" = 3 ] || die "T6's glob does not match the gate-review record"
want=$(cat records/*-review.json | sed -n 's/.*"seconds": *\([0-9]*\).*/\1/p' | awk '{ s += $1 } END { print s + 0 }')
[ "$want" = 28 ] && [ "$(cd "$T/copy" && sh bin/runlog review-seconds)" = 21 ] ||
  die "T6's gate sum ($want) is not the brief's 28 against 21"
# T7: the sound gate fails on the base at its behavioural assertion, not early.
[ ! -s "$T/base.6" ] || die "T7's gate printed on the base: $(cat "$T/base.6")"
[ "$(sh bin/runlog >/dev/null 2>&1; echo $?)" = 0 ] || die "the base's bin/runlog does not exit 0"

commit_all "plan"
