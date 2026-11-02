#!/usr/bin/env bash
# A tiny shell command and no brief yet, for /vloop:create-brief. Run it in an
# EMPTY directory; no network, no claude.
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
. "$HERE/../_lib.sh"
require_empty
git_init
mkdir -p docs/briefs bin tests .vloop/state .vloop/tmp
printf '# greet\n\nA tiny shell command, built by an autonomous loop. Documentation: `README.md`.\n' >CLAUDE.md
printf '.vloop/tmp/\n' >.gitignore
printf '# greet\n\n`greet <name>` prints `hello, <name>`.\n' >README.md
cat >bin/greet <<'R'
#!/bin/sh
if [ $# -lt 1 ]; then
  echo "usage: greet <name>" >&2
  exit 2
fi
echo "hello, $1"
R
cat >tests/greet.sh <<'R'
#!/bin/sh
test "$(sh bin/greet world)" = "hello, world"
R
chmod +x bin/greet
printf 'shell = "sh"\n\n[[check]]\n  name = "greet"\n  paths = ["**"]\n  run = "sh tests/greet.sh"\n' >.vloop/config.toml
: >docs/briefs/.gitkeep
commit_all "baseline"
