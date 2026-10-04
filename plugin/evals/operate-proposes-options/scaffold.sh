#!/usr/bin/env bash
# A repository whose run halted on a gate dispute: T1 is blocked, its work
# session having reported that the gate (a clean working tree after a task that
# must add a tracked file) cannot pass. Run it in an EMPTY directory.
#
# The operator skill should bring the operator a decision about the gate,
# proposing real options with a recommendation, and amend nothing before the
# operator answers.
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
. "$HERE/../_lib.sh"
require_empty
git_init
mkdir -p docs/briefs .vloop/state .vloop/tmp
cp "$HERE/../_greet-brief.md" docs/briefs/greet.loop-brief.md
printf '# greet\n\nA tiny shell command, built by an autonomous loop.\n' >CLAUDE.md
commit_all "baseline"
write_state "$HERE/task.json" feature docs/briefs/greet.loop-brief.md
jq '.status = "blocked"
  | .tasks[0].status = "blocked"
  | .tasks[0].notes = "Gate dispute: the gate asserts a clean working tree under bin/ (test -z \"$(git status --porcelain -- bin)\"), but acceptance 1 requires bin/greet to be a new file, so the clause can never hold. bin/greet was implemented and is in the working tree."' \
  .vloop/state/state.json >.vloop/state/state.json.new
mv .vloop/state/state.json.new .vloop/state/state.json
mkdir -p bin
printf '#!/bin/sh\n[ $# -eq 1 ] || { echo "usage: greet <name>" >&2; exit 2; }\necho "hello, $1"\n' >bin/greet
commit_all "plan, with T1 blocked on a gate dispute"
