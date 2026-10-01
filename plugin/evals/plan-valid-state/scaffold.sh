#!/usr/bin/env bash
# A repository holding one small brief and nothing else. Run it in an EMPTY
# directory. The planner must turn the brief into a plan that validates.
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
. "$HERE/../_lib.sh"
require_empty
git_init
mkdir -p docs/briefs .vloop/state
cp "$HERE/../_greet-brief.md" docs/briefs/greet.loop-brief.md
printf '# greet\n\nA tiny shell command, built by an autonomous loop.\n' >CLAUDE.md
printf 'shout = false\n' >greet.toml
commit_all "baseline"
