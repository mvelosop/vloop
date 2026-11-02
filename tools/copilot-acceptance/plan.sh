#!/bin/sh
# B27's acceptance (A1) of agent.plan = copilot on the real GitHub Copilot CLI, run by
# the operator. The same fixture brief (greet.loop-brief.md) as run.sh, in its own
# scratch repository: agent.plan = copilot with model.plan the model given; gate
# review, work, review and verify stay on Claude.
#
#   tools/copilot-acceptance/plan.sh <copilot model id> [dir]
#
# Writes <dir>/b27-plan-summary.md: each plan session record (agent, model, credits,
# estimated cost, session/v4 validity, permission_denials), the plan (its tasks,
# state/v2 validity) and the gate review's verdict per round, to be recorded in
# docs/design-notes/copilot-cli-experiments.md before the brief is closed. Spends
# Copilot credits (the plan sessions) and Claude money (gate review, work, review,
# verify). Needs copilot, claude and gh on PATH, claude and gh logged in, go and python3.
set -u
model=${1:?usage: plan.sh <copilot model id> [dir]}
dir=${2:-$HOME/vloop-copilot-acceptance}
here=$(cd "$(dirname "$0")" && pwd -P)
top=$(cd "$here/../.." && pwd -P)
die() { echo "b27: $*" >&2; exit 1; }

command -v copilot >/dev/null || die "copilot is not on PATH"
command -v claude >/dev/null || die "claude is not on PATH"
command -v gh >/dev/null || die "gh is not on PATH: Copilot sessions log in through it"
command -v python3 >/dev/null || die "python3 is not on PATH"

mkdir -p "$dir/bin" || die "cannot create $dir"
(cd "$top" && go build -o "$dir/bin/vloop" ./cmd/vloop) || die "build of vloop failed"
V=$dir/bin/vloop

repo=$dir/repo-plan
rm -rf "$repo"
mkdir -p "$repo" && cd "$repo" || die "cannot create $repo"
git init -q -b main && git config user.name "B27 Acceptance" && git config user.email b27@example.invalid || die "git init"
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
"$V" config set agent.plan copilot >/dev/null && "$V" config set model.plan "$model" >/dev/null || die "config"
git add -A && git commit -qm "B27 fixture" || die "commit"
if ! "$V" doctor >"$dir/doctor-plan.txt" 2>&1; then
  cat "$dir/doctor-plan.txt"
  echo "b27: fix what doctor reports (for folder trust: run claude once in $repo and accept), then re-run" >&2
  exit 3
fi
git switch -qc "${name%.loop-brief}" || die "work branch"

echo "b27: vloop run (plan on copilot $model, the rest on claude) ..."
"$V" run "docs/briefs/$name.md" >"$dir/run-plan.log" 2>&1
rc=$?
tail -14 "$dir/run-plan.log"
python3 "$here/report.py" plan "$repo" "$V" "$rc" >"$dir/b27-plan-summary.md" || die "the report failed"
cat "$dir/b27-plan-summary.md"
echo
echo "b27: written $dir/b27-plan-summary.md; record it in docs/design-notes/copilot-cli-experiments.md"
