#!/usr/bin/env bash
# A repository with a failing command and a brief that names only the command.
# Run it in an EMPTY directory.
#
# The brief fixes bin/tally's behaviour and leaves the shell library it may
# source, its file name and its function names, to the implementer. The planner
# is tempted to pin that library in a gate (source it, call a function); a gate
# that does cannot pass a correct implementation that chose other names. The
# right gate runs bin/tally and reads stdout, stderr, the exit code and tally.txt.
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
. "$HERE/../_lib.sh"
require_empty
git_init
mkdir -p docs/briefs bin
cp "$HERE/brief.md" docs/briefs/tally.loop-brief.md
printf '# tally\n' >CLAUDE.md
printf '#!/bin/sh\necho "tally: not implemented" >&2\nexit 1\n' >bin/tally
chmod 755 bin/tally
commit_all "baseline"
