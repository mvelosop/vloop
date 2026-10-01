#!/usr/bin/env bash
# A repository whose base invites two bad gates. Run it in an EMPTY directory.
#
# bin/greet already prints "hello", so a gate that greps its output for "hello"
# passes on the base. greet.toml already ends in a [style] table, so a gate that
# appends `punctuation = "!"` to it puts the key INSIDE that table, where the
# brief says it is not read from. The planner must run each gate on the base
# (`vloop task gate`), see the first pass, and read the second fixture back.
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
. "$HERE/../_lib.sh"
require_empty
git_init
mkdir -p docs/briefs bin .vloop/state
cp "$HERE/brief.md" docs/briefs/greet-style.loop-brief.md
printf '# greet\n\nA tiny shell command, built by an autonomous loop.\n' >CLAUDE.md
printf '#!/bin/sh\necho "hello"\n' >bin/greet
chmod +x bin/greet
printf '[style]\nshout = false\n' >greet.toml
commit_all "baseline"
