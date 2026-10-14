#!/usr/bin/env bash
# A repository whose run halted on a gate dispute, as a real run leaves it: main
# holds only the baseline; the plan and the iterations are commits on a work
# branch named for the run id. The plan has three tasks: T1 done, T2 blocked (its
# work session reported that the gate, a clean working tree after a task that
# must add a tracked file, cannot pass), T3 pending behind T2. Run it in an EMPTY
# directory.
#
# The operator skill should bring the operator a decision about the gate,
# proposing real options with a recommendation, and amend nothing before the
# operator answers.
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
. "$HERE/../_lib.sh"
require_empty
git_init
mkdir -p docs/briefs
cp "$HERE/../_greet-brief.md" docs/briefs/greet.loop-brief.md
printf '# greet\n\nA tiny shell command, built by an autonomous loop.\n' >CLAUDE.md
commit_all "baseline"

RUN_ID=B20261001-0900-greet
git switch -q -c "$RUN_ID"
mkdir -p .vloop/state .vloop/tmp
printf '.vloop/tmp/\n' >.gitignore
python3 - "$HERE/task.json" "$RUN_ID" <<'PY'
import json, subprocess, sys
disputed = json.load(open(sys.argv[1]))
disputed.pop("files", None)
disputed.update(kind="feature", references=[], fixtures="", status="blocked", attempts=1,
    notes="Gate dispute: the gate asserts a clean working tree under bin/ (test -z \"$(git status --porcelain -- bin)\"), but acceptance 1 requires bin/greet to be a new file, so the clause can never hold. bin/greet was implemented and is in the working tree.")
done = {"id": "T1", "title": "Add a README that names the command", "kind": "docs",
    "goal": "The repository says what greet is before the script exists.",
    "depends_on": [], "references": [], "fixtures": "", "status": "done", "attempts": 1, "notes": "",
    "acceptance": ["README.md exists and names the greet command"],
    "verify": "grep -q greet README.md"}
pending = {"id": "T3", "title": "Add the shout option read from greet.toml", "kind": "feature",
    "goal": "shout = true in greet.toml prints the greeting in upper case; it builds on the script T2 adds.",
    "depends_on": ["T2"], "references": [], "fixtures": "", "status": "pending", "attempts": 0, "notes": "",
    "acceptance": ["with shout = true in greet.toml, bin/greet world prints HELLO, WORLD"],
    "verify": "d=$(mktemp -d \"${TMPDIR:-/tmp}/gate.XXXXXX\") && printf 'shout = true\\n' >\"$d/greet.toml\" && (cd \"$d\" && sh \"$OLDPWD/bin/greet\" world) | grep -qx 'HELLO, WORLD'"}
base = subprocess.check_output(["git", "rev-parse", "main"], text=True).strip()
plan = {"schema": "state/v2", "run_id": sys.argv[2], "brief": "docs/briefs/greet.loop-brief.md",
    "base": base, "branch": sys.argv[2], "status": "running", "iteration": 2,
    "created": "2026-10-01T09:00:00Z", "updated": "2026-10-01T09:30:00Z", "shell": "bash",
    "checks": [{"name": "all", "paths": ["**"], "run": "true"}],
    "gate_scratch": [], "gate_review": {"rounds": 1, "verdict": "PASS"},
    "tasks": [done, disputed, pending]}
json.dump(plan, open(".vloop/state/state.json", "w"), indent=2)
PY
commit_all "[vloop] plan: 3 tasks"

printf '# greet\n\nA tiny shell command, built by an autonomous loop.\n\nUsage: greet <name>\n' >README.md
python3 - <<'PY'
import json
p = json.load(open(".vloop/state/state.json"))
p["updated"] = "2026-10-01T09:10:00Z"
json.dump(p, open(".vloop/state/state.json", "w"), indent=2)
PY
commit_all "[vloop] T1: done"

mkdir -p bin
printf '#!/bin/sh\n[ $# -eq 1 ] || { echo "usage: greet <name>" >&2; exit 2; }\necho "hello, $1"\n' >bin/greet
python3 - <<'PY'
import json
p = json.load(open(".vloop/state/state.json"))
p["status"] = "blocked"
p["updated"] = "2026-10-01T09:30:00Z"
json.dump(p, open(".vloop/state/state.json", "w"), indent=2)
PY
commit_all "[vloop] T2: blocked on a gate dispute"
