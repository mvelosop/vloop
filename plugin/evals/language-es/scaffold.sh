#!/usr/bin/env bash
# As plan-valid-state, with the repository configured for Spanish. Run it in an
# EMPTY directory.
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
. "$HERE/../_lib.sh"
require_empty
git_init
mkdir -p docs/briefs .vloop/state
cp "$HERE/../_greet-brief.md" docs/briefs/greet.loop-brief.md
printf '# greet\n\nA tiny shell command, built by an autonomous loop.\n' >CLAUDE.md
printf 'shout = false\n' >greet.toml
printf 'language = "es"\n' >.vloop/config.toml
commit_all "baseline"
