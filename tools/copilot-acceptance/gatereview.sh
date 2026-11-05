#!/bin/sh
# B26's acceptance (A1) of agent.gate-review = copilot on the real GitHub Copilot CLI,
# run by the operator. The same fixture brief (greet.loop-brief.md) as run.sh, in
# its own scratch repository: agent.gate-review = copilot with model.gate-review the model
# given; plan, work, review and verify stay on Claude.
#
#   tools/copilot-acceptance/gatereview.sh <copilot model id> [dir]
#
# Writes <dir>/b26-gatereview-summary.md: each gate-review session record (round, agent,
# model, credits, estimated cost, session/v4 validity, permission_denials) and each kept
# verdict (round, present, gate-verdict/v2 validity, verdict, findings by kind), to be recorded in docs/design-notes/copilot-cli-experiments.md
# before the brief is closed. Spends Claude money (plan, work, reviews, verify) and Copilot
# credits (the gate reviews). Needs copilot, claude and gh on PATH, claude and gh logged in,
# and go.
set -u
model=${1:?usage: gatereview.sh <copilot model id> [dir]}
dir=${2:-$HOME/vloop-copilot-acceptance}
here=$(cd "$(dirname "$0")" && pwd -P)
top=$(cd "$here/../.." && pwd -P)
die() { echo "b26: $*" >&2; exit 1; }
dir=$(mkdir -p "$dir" && cd "$dir" && pwd -P) || die "cannot create $dir"

command -v copilot >/dev/null || die "copilot is not on PATH"
command -v claude >/dev/null || die "claude is not on PATH"
command -v gh >/dev/null || die "gh is not on PATH: Copilot sessions log in through it"

mkdir -p "$dir/bin" || die "cannot create $dir"
(cd "$top" && go build -o "$dir/bin/vloop" ./cmd/vloop) || die "build of vloop failed"
V=$dir/bin/vloop

repo=$dir/repo-gatereview
rm -rf "$repo"
mkdir -p "$repo" && cd "$repo" || die "cannot create $repo"
git init -q -b main && git config user.name "B26 Acceptance" && git config user.email b26@example.invalid || die "git init"
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
"$V" config set agent.gate-review copilot >/dev/null && "$V" config set model.gate-review "$model" >/dev/null || die "config"
git add -A && git commit -qm "B26 fixture" || die "commit"
if ! "$V" doctor >"$dir/doctor-gatereview.txt" 2>&1; then
  cat "$dir/doctor-gatereview.txt"
  echo "b26: fix what doctor reports (for folder trust: run claude once in $repo and accept), then re-run" >&2
  exit 3
fi
git switch -qc "${name%.loop-brief}" || die "work branch"

echo "b26: vloop run (gate review on copilot $model, the rest on claude) ..."
"$V" run "docs/briefs/$name.md" >"$dir/run-gatereview.log" 2>&1
rc=$?
tail -14 "$dir/run-gatereview.log"
go -C "$top" run ./tools/copilot-acceptance/kit report gate-review "$repo" "$V" "$rc" >"$dir/b26-gatereview-summary.md" || die "the report failed"
cat "$dir/b26-gatereview-summary.md"
echo
echo "b26: written $dir/b26-gatereview-summary.md; record it in docs/design-notes/copilot-cli-experiments.md"
