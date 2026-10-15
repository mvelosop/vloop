#!/usr/bin/env bash
# A small built result for /vloop:verify, with three planted mismatches. Run it
# in an EMPTY directory; no network, no claude.
#
#   example   `sh bin/greet` is briefed to exit 2; the result exits 1
#   contract  "With --shout, greet prints the greeting in upper case" is not done
#   surface   README.md still says greet ignores its arguments
#
# The worked example's first line (`sh bin/greet world`) is sound.
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

`greet` prints `hello` and ignores its arguments.
R
cat >docs/briefs/greet-name.loop-brief.md <<'R'
# Brief: greet a name

## Behaviour contract

- `greet <name>` prints `hello, <name>` on stdout and exits 0.
- `greet` with no argument prints `usage: greet <name>` on stderr, nothing on stdout, and exits 2.
- With `--shout`, `greet` prints the greeting in upper case.

## Worked example

```
$ sh bin/greet world
hello, world
$ echo $?
0
$ sh bin/greet
usage: greet <name>
$ echo $?
2
```

## Required

- `bin/greet` implements the contract.
- `README.md` describes what `greet` does now.
R
printf '#!/bin/sh\necho "hello"\n' >bin/greet
chmod +x bin/greet
commit_all "baseline"
base="$(git rev-parse HEAD)"
cat >bin/greet <<'R'
#!/bin/sh
if [ $# -lt 1 ]; then
  echo "usage: greet <name>" >&2
  exit 1
fi
echo "hello, $1"
R
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
        "created": "2026-10-15T00:00:00Z", "updated": "2026-10-15T00:00:00Z",
        "shell": "sh", "checks": [], "gate_scratch": [],
        "gate_review": {"rounds": 1, "verdict": "PASS"}, "tasks": [task]}
json.dump(plan, open(".vloop/state/state.json", "w"), indent=2)
PY
commit_all "plan"
