#!/bin/sh
# B22's acceptance of a real feedback-mode run with a copilot second reviewer,
# run by the operator. The same fixture brief as run.sh, in its own scratch
# repository: run.second-review = feedback, agent.second-review = copilot with
# the model given; plan, work and the primary review stay on Claude.
#
#   tools/copilot-acceptance/second.sh <copilot model id> [dir]
#
# Writes <dir>/b22-second-summary.md: the second-review session records and
# the dissents, to be recorded in docs/design-notes/copilot-cli-experiments.md
# before the brief is closed. Spends Claude money and Copilot credits.
set -u
model=${1:?usage: second.sh <copilot model id> [dir]}
dir=${2:-$HOME/vloop-copilot-acceptance}
here=$(cd "$(dirname "$0")" && pwd -P)
top=$(cd "$here/../.." && pwd -P)
die() { echo "b22: $*" >&2; exit 1; }

command -v copilot >/dev/null || die "copilot is not on PATH"
command -v claude >/dev/null || die "claude is not on PATH"
command -v python3 >/dev/null || die "python3 is not on PATH"

mkdir -p "$dir/bin" || die "cannot create $dir"
(cd "$top" && go build -o "$dir/bin/vloop" ./cmd/vloop) || die "build of vloop failed"
V=$dir/bin/vloop

repo=$dir/repo-second
rm -rf "$repo"
mkdir -p "$repo" && cd "$repo" || die "cannot create $repo"
git init -q -b main && git config user.name "B22 Acceptance" && git config user.email b22@example.invalid || die "git init"
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
"$V" config set run.second-review feedback >/dev/null &&
  "$V" config set agent.second-review copilot >/dev/null &&
  "$V" config set model.second-review "$model" >/dev/null || die "config"
git add -A && git commit -qm "B22 fixture" || die "commit"
if ! "$V" doctor >"$dir/doctor-second.txt" 2>&1; then
  cat "$dir/doctor-second.txt"
  echo "b22: fix what doctor reports (for folder trust: run claude once in $repo and accept), then re-run" >&2
  exit 3
fi
git switch -qc "${name%.loop-brief}" || die "work branch"

echo "b22: vloop run (second review on copilot $model, feedback mode) ..."
"$V" run "docs/briefs/$name.md" >"$dir/run-second.log" 2>&1
rc=$?
tail -14 "$dir/run-second.log"
python3 "$here/report.py" second "$repo" "$V" "$rc" >"$dir/b22-second-summary.md" || die "the report failed"
cat "$dir/b22-second-summary.md"
echo
echo "b22: written $dir/b22-second-summary.md; record it in docs/design-notes/copilot-cli-experiments.md"
