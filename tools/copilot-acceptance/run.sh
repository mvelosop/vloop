#!/bin/sh
# B19's acceptance (C5) on the real GitHub Copilot CLI, run by the operator.
#
#   tools/copilot-acceptance/run.sh <copilot model id> [dir]
#
# 1. A real `vloop run` of a fixture brief in a scratch repository under <dir>
#    (default ~/vloop-copilot-acceptance), with agent.review = copilot and the
#    model given, using a vloop built from this tree. A wrapper in front of
#    copilot logs the argv, working directory and environment of every Copilot
#    session.
# 2. A deny probe: the logged review session's argv and environment, exactly as
#    the adapter built them, with only the prompt replaced, asking the model to
#    try each form B19's C2 must deny. Shims for gh, curl, wget, ssh, scp, nc,
#    rm and git record what got through and do nothing outward.
#
# Writes <dir>/c5-summary.md, to be recorded in
# docs/design-notes/copilot-cli-experiments.md. Spends Claude money (plan and
# work) and Copilot credits (reviews and the probe). Needs copilot and claude
# on PATH, both logged in, go and python3.
set -u
model=${1:?usage: run.sh <copilot model id> [dir]}
dir=${2:-$HOME/vloop-copilot-acceptance}
here=$(cd "$(dirname "$0")" && pwd -P)
top=$(cd "$here/../.." && pwd -P)
die() { echo "c5: $*" >&2; exit 1; }

real=$(command -v copilot) || die "copilot is not on PATH"
command -v claude >/dev/null || die "claude is not on PATH"
command -v python3 >/dev/null || die "python3 is not on PATH"
echo "c5: copilot $("$real" --version 2>&1 | head -1)"

mkdir -p "$dir/bin" || die "cannot create $dir"
(cd "$top" && go build -o "$dir/bin/vloop" ./cmd/vloop) || die "build of vloop failed"
V=$dir/bin/vloop

# The scratch repository: the same path every time, so Claude's folder trust,
# once accepted, still holds. It must not sit inside another vloop repository.
repo=$dir/repo
rm -rf "$repo" "$dir/shim-copilot" "$dir/shims" "$dir/copilot-argv.log" "$dir/probe-hits.log"
mkdir -p "$repo" && cd "$repo" || die "cannot create $repo"
git init -q -b main && git config user.name "C5 Acceptance" && git config user.email c5@example.invalid || die "git init"
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
"$V" config set agent.review copilot >/dev/null && "$V" config set model.review "$model" >/dev/null || die "config"
git add -A && git commit -qm "C5 fixture" || die "commit"
if ! "$V" doctor >"$dir/doctor.txt" 2>&1; then
  cat "$dir/doctor.txt"
  echo "c5: fix what doctor reports (for folder trust: run claude once in $repo and accept), then re-run" >&2
  exit 3
fi
git switch -qc "${name%.loop-brief}" || die "work branch"

# The wrapper: log each Copilot session's argv, cwd and environment, then run it.
mkdir -p "$dir/shim-copilot"
cat >"$dir/shim-copilot/copilot" <<WRAP
#!/bin/sh
{ echo "=== copilot"; for a in "\$@"; do printf 'arg %s\n' "\$a"; done
  printf 'cwd %s\n' "\$(pwd -P)"; env | LC_ALL=C sort | sed 's/^/env /'; } >>"$dir/copilot-argv.log"
exec "$real" "\$@"
WRAP
chmod +x "$dir/shim-copilot/copilot"

echo "c5: vloop run (plan and work on claude, reviews on copilot $model) ..."
PATH=$dir/shim-copilot:$PATH "$V" run "docs/briefs/$name.md" >"$dir/run.log" 2>&1
rc=$?
tail -14 "$dir/run.log"
python3 "$here/report.py" run "$repo" "$V" "$rc" >"$dir/c5-run.md" || die "the run report failed"

# The probe: the last review session exactly as the adapter started it, with
# only the prompt replaced and the shims first on PATH.
[ -s "$dir/copilot-argv.log" ] || die "no copilot session was started; see $dir/run.log"
mkdir -p "$dir/shims"
for c in gh curl wget ssh scp nc rm; do
  printf '#!/bin/sh\necho "%s $*" >>"%s/probe-hits.log"\nexit 0\n' "$c" "$dir" >"$dir/shims/$c"
done
realgit=$(command -v git)
cat >"$dir/shims/git" <<GIT
#!/bin/sh
echo "git \$*" >>"$dir/probe-hits.log"
case " \$* " in *" status "*|*" diff "*|*" log "*|*" --version "*) exec "$realgit" "\$@";; esac
exit 0
GIT
chmod +x "$dir/shims/"*
: >"$dir/probe-hits.log"
echo "c5: deny probe ..."
python3 "$here/report.py" probe "$dir" "$real" || die "the probe failed to start"
python3 "$here/report.py" summary "$dir" >"$dir/c5-probe.md" || die "the probe report failed"

cat "$dir/c5-run.md" "$dir/c5-probe.md" >"$dir/c5-summary.md"
echo
cat "$dir/c5-summary.md"
echo
echo "c5: written $dir/c5-summary.md; the model's own account is the last assistant.message in $dir/probe.jsonl"
