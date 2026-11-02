#!/bin/sh
# B25's acceptance (A1) of agent.verify = copilot on the real GitHub Copilot CLI,
# run by the operator. The same fixture brief (greet.loop-brief.md) as run.sh, in
# its own scratch repository: agent.verify = copilot with model.verify the model
# given, run.verify on; plan, work and review stay on Claude.
#
#   tools/copilot-acceptance/verify.sh <copilot model id> [dir]
#
# Writes <dir>/b25-verify-summary.md: the verify session record (agent, model,
# credits, estimated cost, session/v4 validity, permission_denials) and the kept
# report (present, verify/v1 or verify/v2 validity, findings), to be recorded in docs/design-notes/copilot-cli-experiments.md
# before the brief is closed. Spends Claude money (plan, work and reviews) and Copilot
# credits (verify). Needs copilot, claude and gh on PATH, claude and gh logged in,
# go and python3.
set -u
model=${1:?usage: verify.sh <copilot model id> [dir]}
dir=${2:-$HOME/vloop-copilot-acceptance}
here=$(cd "$(dirname "$0")" && pwd -P)
top=$(cd "$here/../.." && pwd -P)
die() { echo "b25: $*" >&2; exit 1; }

command -v copilot >/dev/null || die "copilot is not on PATH"
command -v claude >/dev/null || die "claude is not on PATH"
command -v gh >/dev/null || die "gh is not on PATH: Copilot sessions log in through it"
command -v python3 >/dev/null || die "python3 is not on PATH"

mkdir -p "$dir/bin" || die "cannot create $dir"
(cd "$top" && go build -o "$dir/bin/vloop" ./cmd/vloop) || die "build of vloop failed"
V=$dir/bin/vloop

repo=$dir/repo-verify
rm -rf "$repo"
mkdir -p "$repo" && cd "$repo" || die "cannot create $repo"
git init -q -b main && git config user.name "B25 Acceptance" && git config user.email b25@example.invalid || die "git init"
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
"$V" config set agent.verify copilot >/dev/null && "$V" config set model.verify "$model" >/dev/null || die "config"
git add -A && git commit -qm "B25 fixture" || die "commit"
if ! "$V" doctor >"$dir/doctor-verify.txt" 2>&1; then
  cat "$dir/doctor-verify.txt"
  echo "b25: fix what doctor reports (for folder trust: run claude once in $repo and accept), then re-run" >&2
  exit 3
fi
git switch -qc "${name%.loop-brief}" || die "work branch"

echo "b25: vloop run (verify on copilot $model, plan, work and reviews on claude) ..."
"$V" run "docs/briefs/$name.md" >"$dir/run-verify.log" 2>&1
rc=$?
tail -14 "$dir/run-verify.log"
python3 "$here/report.py" verify "$repo" "$V" "$rc" >"$dir/b25-verify-summary.md" || die "the report failed"
cat "$dir/b25-verify-summary.md"
echo
echo "b25: written $dir/b25-verify-summary.md; record it in docs/design-notes/copilot-cli-experiments.md"
