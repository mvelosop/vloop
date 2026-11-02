#!/usr/bin/env bash
# A small built result for /vloop:verify, whose brief's worked example holds one
# sound line and one line the verify session cannot run. Run it in an EMPTY
# directory; no network, no claude.
#
#   sound       `sh bin/greet world` prints hello, world and exits 0, as briefed
#   unrunnable  `claude -p "say hi"` needs a live model: it belongs in not_run
#
# The built result is correct, so the report should hold no finding at all.
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
. "$HERE/../_lib.sh"
require_empty
git_init
mkdir -p docs/briefs bin .vloop/state .vloop/tmp
printf '# greet\n\nA tiny shell command, built by an autonomous loop.\n' >CLAUDE.md
printf '.vloop/tmp/\n' >.gitignore
cat >README.md <<'R'
# greet

`greet <name>` prints `hello, <name>`.
R
cat >docs/briefs/greet-name.loop-brief.md <<'R'
# Brief: greet a name

## Behaviour contract

- `greet <name>` prints `hello, <name>` on stdout and exits 0.

## Worked example

```
$ sh bin/greet world
hello, world
$ echo $?
0
$ claude -p "say hi"
hi
```

## Required

- `bin/greet` implements the contract.
- `README.md` describes what `greet` does.
R
printf '#!/bin/sh\necho "hello"\n' >bin/greet
chmod +x bin/greet
commit_all "baseline"
base="$(git rev-parse HEAD)"
printf '#!/bin/sh\necho "hello, $1"\n' >bin/greet
commit_all "built result"
python3 - "$base" <<'PY'
import json, sys
base = sys.argv[1]
task = {
    "id": "T1", "title": "Greet a name", "kind": "feature", "fixtures": "",
    "goal": "The brief's contract for greet, as one command a later task can build on.",
    "references": [], "depends_on": [],
    "acceptance": ["bin/greet world prints hello, world and exits 0"],
    "verify": "sh bin/greet world | grep -qx 'hello, world'",
    "status": "done", "attempts": 1, "notes": "",
}
plan = {"schema": "state/v2", "run_id": "eval", "brief": "docs/briefs/greet-name.loop-brief.md",
        "base": base, "branch": "vloop/eval", "status": "running", "iteration": 1,
        "created": "2026-11-02T00:00:00Z", "updated": "2026-11-02T00:00:00Z",
        "shell": "sh", "checks": [], "gate_scratch": [],
        "gate_review": {"rounds": 1, "verdict": "PASS"}, "tasks": [task]}
json.dump(plan, open(".vloop/state/state.json", "w"), indent=2)
PY
commit_all "plan"
