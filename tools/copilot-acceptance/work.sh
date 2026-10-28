#!/bin/sh
# B24's acceptance (A1) of agent.work = copilot on the real GitHub Copilot CLI,
# run by the operator. The same fixture brief as run.sh, in its own scratch
# repository: agent.work = copilot with the model given; plan and review stay on
# Claude.
#
#   tools/copilot-acceptance/work.sh <copilot model id> [dir]
#
# Writes <dir>/b24-work-summary.md: each work session record (agent, model,
# credits, estimated cost, session/v4 validity, permission_denials) and each
# task's outcome, to be recorded in docs/design-notes/copilot-cli-experiments.md
# before the brief is closed. Spends Claude money (plan and reviews) and Copilot
# credits (work). Needs copilot, claude and gh on PATH, claude and gh logged in,
# go and python3.
set -u
model=${1:?usage: work.sh <copilot model id> [dir]}
dir=${2:-$HOME/vloop-copilot-acceptance}
here=$(cd "$(dirname "$0")" && pwd -P)
top=$(cd "$here/../.." && pwd -P)
die() { echo "b24: $*" >&2; exit 1; }

command -v copilot >/dev/null || die "copilot is not on PATH"
command -v claude >/dev/null || die "claude is not on PATH"
command -v gh >/dev/null || die "gh is not on PATH: Copilot sessions log in through it"
command -v python3 >/dev/null || die "python3 is not on PATH"

mkdir -p "$dir/bin" || die "cannot create $dir"
(cd "$top" && go build -o "$dir/bin/vloop" ./cmd/vloop) || die "build of vloop failed"
V=$dir/bin/vloop

repo=$dir/repo-work
rm -rf "$repo"
mkdir -p "$repo" && cd "$repo" || die "cannot create $repo"
git init -q -b main && git config user.name "B24 Acceptance" && git config user.email b24@example.invalid || die "git init"
"$V" init >/dev/null || die "vloop init"
[ -f "$repo/.vloop/config.toml" ] || die "vloop init wrote no .vloop/config.toml in $repo (is $dir inside another vloop repository?)"
rm -f docs/briefs/*.md
name="B$(date +%Y%m%d-%H%M)-greet.loop-brief"
sed -e "s/@NAME@/$name/" -e "s/@DATE@/$(date +%Y-%m-%d)/" "$here/greet.loop-brief.md" >"docs/briefs/$name.md"
cat >>.vloop/config.toml <<'TOML'

[[check]]
name = "greet"
run = "test ! -e tests/greet.sh || sh tests/greet.sh"
paths = ["**"]
TOML
"$V" config set agent.work copilot >/dev/null && "$V" config set model.work "$model" >/dev/null || die "config"
git add -A && git commit -qm "B24 fixture" || die "commit"
if ! "$V" doctor >"$dir/doctor-work.txt" 2>&1; then
  cat "$dir/doctor-work.txt"
  echo "b24: fix what doctor reports (for folder trust: run claude once in $repo and accept), then re-run" >&2
  exit 3
fi
git switch -qc "${name%.loop-brief}" || die "work branch"

echo "b24: vloop run (work on copilot $model, plan and reviews on claude) ..."
"$V" run "docs/briefs/$name.md" >"$dir/run-work.log" 2>&1
rc=$?
tail -14 "$dir/run-work.log"
python3 "$here/report.py" work "$repo" "$V" "$rc" >"$dir/b24-work-summary.md" || die "the report failed"
cat "$dir/b24-work-summary.md"
echo
echo "b24: written $dir/b24-work-summary.md; record it in docs/design-notes/copilot-cli-experiments.md"
