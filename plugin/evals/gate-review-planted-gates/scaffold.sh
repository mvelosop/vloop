#!/usr/bin/env bash
# A fresh state/v2 plan that plants five gates for /vloop:gate-review. Run it in
# an EMPTY directory; no network, no claude.
#
#   T1  fails on a path typo (sh bin/gret)                -> wrong-reason
#   T2  greps the source text of bin/greet                -> not-contract
#   T3  its judge is tests/extra-args.sh, which T3 writes -> task-reach
#   T4  greps the output for hello, which the base prints -> wrong-reason
#   T5  sound: exact stdout, stderr and exit code         -> PASS
#
# T3 is the one deliberate task-written judge in the eval fixtures; any other
# is a violation.
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
. "$HERE/../_lib.sh"
require_empty
git_init
mkdir -p docs/briefs bin .vloop/state .vloop/tmp
cp "$HERE/brief.md" docs/briefs/greet-name.loop-brief.md
printf '# greet\n\nA tiny shell command, built by an autonomous loop.\n' >CLAUDE.md
printf '.vloop/tmp/\n' >.gitignore
printf '#!/bin/sh\necho "hello"\n' >bin/greet
chmod +x bin/greet
printf '[style]\nshout = false\n' >greet.toml
commit_all "baseline"
python3 - <<'PY'
import json, os, subprocess
base = subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip()
brief = "docs/briefs/greet-name.loop-brief.md"
gates = [
    ("T1", "Greet a name", "bin/greet world prints exactly hello, world and exits 0",
     "sh bin/gret world | grep -qx 'hello, world'"),
    ("T2", "Read the punctuation key", "with punctuation = \"!\" at the top level of greet.toml, bin/greet world prints hello! world",
     "grep -q 'punctuation' bin/greet"),
    ("T3", "Ignore extra arguments", "bin/greet a b prints hello, a",
     "sh tests/extra-args.sh"),
    ("T4", "Shout when asked", "with shout = true in greet.toml, bin/greet world prints HELLO, WORLD",
     "sh bin/greet world | grep -qi 'hello'"),
    ("T5", "Print a usage error", "bin/greet with no argument prints usage: greet <name> on stderr, nothing on stdout, and exits 2",
     "d=$(mktemp -d \"${TMPDIR:-/tmp}/gate.XXXXXX\") || exit 1; sh bin/greet >\"$d/out\" 2>\"$d/err\"; rc=$?; "
     "[ \"$rc\" = 2 ] && [ ! -s \"$d/out\" ] && [ \"$(cat \"$d/err\")\" = 'usage: greet <name>' ]; r=$?; rm -rf \"$d\"; exit $r"),
]
tasks = [{
    "id": tid, "title": title, "kind": "feature", "fixtures": "",
    "goal": f"The brief's contract for greet: {accept}. One behaviour of the command, so a later task can build on it.",
    "references": [], "depends_on": [], "acceptance": [accept],
    "verify": verify, "status": "pending", "attempts": 0, "notes": "",
} for tid, title, accept, verify in gates]
plan = {"schema": "state/v2", "run_id": "eval", "brief": brief, "base": base,
        "branch": "vloop/eval", "status": "running", "iteration": 0,
        "created": "2026-10-04T00:00:00Z", "updated": "2026-10-04T00:00:00Z",
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
commit_all "plan"
