#!/bin/sh
# B22's acceptance on the real GitHub Copilot CLI, run by the operator, after
# run.sh has run (it reuses <dir>: the scratch repository, its shims, and the
# logged review session whose hooks file vloop wrote).
#
#   tools/copilot-acceptance/hook.sh [dir]
#
# 1. The hook probe: the review session under `vloop fence hook`, asked to run
#    each deny form of the worked example and to invoke a skill other than
#    vloop's. Records per form whether Copilot ran it.
# 2. The hook's failure modes: the hook command missing, exiting 1, printing
#    output that is not JSON, and timing out. Each is one session asked to run a
#    call the shims record; the summary says whether Copilot then ran it.
#
# Writes <dir>/b22-hook-summary.md, to be recorded in
# docs/design-notes/copilot-cli-experiments.md before the brief is closed.
# Spends Copilot credits. Needs copilot and python3 on PATH.
set -u
dir=${1:-$HOME/vloop-copilot-acceptance}
here=$(cd "$(dirname "$0")" && pwd -P)
die() { echo "b22: $*" >&2; exit 1; }

real=$(command -v copilot) || die "copilot is not on PATH"
command -v python3 >/dev/null || die "python3 is not on PATH"
[ -s "$dir/copilot-argv.log" ] || die "no logged copilot session in $dir: run run.sh first"
[ -d "$dir/shims" ] || die "no shims in $dir: run run.sh first"

: >"$dir/probe-hits.log"
echo "b22: hook probe of the worked example's deny forms ..."
python3 "$here/hook.py" forms "$dir" "$real" || die "the hook probe failed to start"
for mode in missing exit1 notjson timeout; do
  echo "b22: hook failure: $mode ..."
  python3 "$here/hook.py" failure "$dir" "$real" "$mode" || die "the $mode probe failed to start"
done
python3 "$here/hook.py" summary "$dir" >"$dir/b22-hook-summary.md" || die "the report failed"
cat "$dir/b22-hook-summary.md"
echo
echo "b22: written $dir/b22-hook-summary.md; record it in docs/design-notes/copilot-cli-experiments.md"
