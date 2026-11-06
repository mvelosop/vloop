#!/bin/sh
# Regenerate the kit's goldens from the starting point's Python.
#
#   gen.sh <python dir> <vloop> <out dir>
#
# <python dir> holds report.py, hook.py and bypass.sh (the base commit's, for
# instance: `git show <base>:tools/copilot-acceptance/<file>`). The `calls`
# heredoc of bypass.sh is extracted into <python dir>/calls.py. The Python runs
# on the fixtures of this folder, each copied into a temporary directory
# resolved with symlinks followed, @DIR@ replaced by that path; what it prints
# and the files it writes go to <out dir> with the path replaced back by @DIR@
# (and, in the stub's log, the lines of PWD, OLDPWD, SHLVL and _ dropped).
# Nothing is written into the repository tree; no .pyc is left anywhere.
# The goldens are committed as this script writes them; never edit one by hand.
set -u
here=$(cd "$(dirname "$0")" && pwd -P)
[ $# -eq 3 ] || { echo "usage: gen.sh <python dir> <vloop> <out dir>" >&2; exit 2; }
pydir=$(cd "$1" && pwd -P) || exit 1
case $2 in /*) vloop=$2 ;; *) vloop=$(cd "$(dirname "$2")" && pwd -P)/$(basename "$2") ;; esac
mkdir -p "$3" || exit 1
out=$(cd "$3" && pwd -P) || exit 1
die() { echo "gen.sh: $*" >&2; exit 1; }
[ -x "$vloop" ] || die "no executable vloop at $vloop"
for f in report.py hook.py bypass.sh; do [ -f "$pydir/$f" ] || die "no $f in $pydir"; done

unset VLOOP_HOOK_PHASE STUB_LOG STUB_OUT STUB_RC
export LC_ALL=C PYTHONDONTWRITEBYTECODE=1 PYTHONUTF8=1 PYTHONIOENCODING=utf-8

awk '/^  python3 - "\$d\/out.jsonl" <<.PY.$/ {on=1; next} on && /^PY$/ {on=0} on' "$pydir/bypass.sh" >"$pydir/calls.py"
[ -s "$pydir/calls.py" ] || die "no calls heredoc in $pydir/bypass.sh"

t=$(mktemp -d "${TMPDIR:-/tmp}/kitgen.XXXXXX") && t=$(cd "$t" && pwd -P) || die mktemp
trap 'rm -rf "$t"' EXIT
mkdir -p "$t/elsewhere" || die mkdir
py() { python3 -I -B "$@"; }

# esc STRING: STRING made safe in the pattern or the replacement of a sed s||| command.
esc() { printf '%s' "$1" | sed 's/[][\.*^$|&/]/\\&/g'; }

# put SRC DEST: a copy of the fixture SRC at DEST, @DIR@ replaced by DEST.
put() {
  rm -rf "$2"; mkdir -p "$2" && cp -R "$1/." "$2/" || die "copy $1"
  r=$(esc "$2")
  find "$2" -type f | while IFS= read -r f; do
    if grep -q '@DIR@' "$f"; then sed "s|@DIR@|$r|g" "$f" >"$f.sub" && mv "$f.sub" "$f" || exit 1; fi
  done || die "substitute in $2"
}

# norm DEST...: stdin to the golden file DEST, the working directory D replaced by @DIR@.
norm() { sed "s|$(esc "$D")|@DIR@|g" >"$1"; }

# snap NAME FILE...: the files the Python wrote, from D to $out/NAME/, normalised;
# the stub's log without the lines of the variables that differ between two runs.
snap() {
  n=$1; shift; mkdir -p "$out/$n" || die mkdir
  for f in "$@"; do
    [ -f "$D/$f" ] || continue
    mkdir -p "$out/$n/$(dirname "$f")" || die mkdir
    case $f in
    stub.log) grep -Ev '^env (PWD|OLDPWD|SHLVL|_)=' "$D/$f" | norm "$out/$n/$f" ;;
    *) norm "$out/$n/$f" <"$D/$f" ;;
    esac
  done
}
WRITTEN="probe.argv probe.rc probe.jsonl probe.stderr stub.log"
hooks_home="it's repo/.vloop/tmp/home/hooks/vloop-fence.json"

# ---- the six reports and the two odd ones --------------------------------------
# report KIND FIXTURE CODE: report.py KIND on the fixture repository reports/FIXTURE.
report() {
  D=$t/rep-$2; put "$here/reports/$2" "$D"
  (cd "$t/elsewhere" && py "$pydir/report.py" "$1" "$D" "$vloop" "$3") >"$t/stdout" 2>"$t/stderr" ||
    die "report.py $1 on $2 failed: $(cat "$t/stderr")"
  mkdir -p "$out/report" && norm "$out/report/$1-$2.md" <"$t/stdout"
}
report run r1 0
report work r1 1
report verify r1 0
report gate-review r1 2
report plan r1 0
report second r2 0
report work r3 0   # vloop metrics exits 1 on this one
report run r4 0    # a brief with no run folder

# ---- the deny probe and its summary --------------------------------------------
D=$t/probe; put "$here/probes/probe" "$D"
(cd "$t/elsewhere" && py "$pydir/report.py" probe "$D" "$here/copilot") >"$t/stdout" 2>"$t/stderr" ||
  die "report.py probe failed: $(cat "$t/stderr")"
mkdir -p "$out/probe" && norm "$out/probe/stdout" <"$t/stdout"
snap probe $WRITTEN
for s in sum1 sum2; do
  D=$t/$s; put "$here/probes/$s" "$D"
  (cd "$t/elsewhere" && py "$pydir/report.py" summary "$D") >"$t/stdout" 2>"$t/stderr" || die "report.py summary on $s failed: $(cat "$t/stderr")"
  mkdir -p "$out/summary" && norm "$out/summary/$s.md" <"$t/stdout"
done

# ---- the hook probes: each step on a fresh copy, so each log is its own ----------
D=$t/hook-forms; put "$here/probes/hook" "$D"
(cd "$t/elsewhere" && py "$pydir/hook.py" forms "$D" "$here/copilot") >"$t/stdout" 2>"$t/stderr" ||
  die "hook.py forms failed: $(cat "$t/stderr")"
snap hook-forms stub.log hook-forms.argv hook-forms.rc hook-forms.jsonl hook-forms.stderr "$hooks_home"
for m in missing exit1 notjson timeout; do
  D=$t/hook-fail-$m; put "$here/probes/hook" "$D"
  (cd "$t/elsewhere" && py "$pydir/hook.py" failure "$D" "$here/copilot" "$m") >"$t/stdout" 2>"$t/stderr" ||
    die "hook.py failure $m failed: $(cat "$t/stderr")"
  snap "hook-fail-$m" stub.log "hook-fail-$m.argv" "hook-fail-$m.rc" "hook-fail-$m.jsonl" "hook-fail-$m.stderr" "$hooks_home"
done
# The forms under another phase: the hooks file names it.
D=$t/hook-phase; put "$here/probes/hook" "$D"
(cd "$t/elsewhere" && VLOOP_HOOK_PHASE=second-review py "$pydir/hook.py" forms "$D" "$here/copilot") >"$t/stdout" 2>"$t/stderr" ||
  die "hook.py forms (second-review) failed: $(cat "$t/stderr")"
snap hook-forms-second-review "$hooks_home"
# The summary of a directory that holds every probe, and of one that holds two.
D=$t/hook-all; put "$here/probes/hook" "$D"
(cd "$t/elsewhere" && py "$pydir/hook.py" forms "$D" "$here/copilot" &&
  for m in missing exit1 notjson timeout; do py "$pydir/hook.py" failure "$D" "$here/copilot" "$m" || exit 1; done &&
  py "$pydir/hook.py" summary "$D") >"$t/stdout" 2>"$t/stderr" || die "hook.py summary failed: $(cat "$t/stderr")"
mkdir -p "$out/hook-summary" && norm "$out/hook-summary/all.md" <"$t/stdout"
snap hook-summary stub.log
D=$t/hsum; put "$here/probes/hsum" "$D"
(cd "$t/elsewhere" && py "$pydir/hook.py" summary "$D") >"$t/stdout" 2>"$t/stderr" || die "hook.py summary on hsum failed: $(cat "$t/stderr")"
norm "$out/hook-summary/partial.md" <"$t/stdout"

# ---- calls ------------------------------------------------------------------------
D=$t/calls; put "$here/probes/calls" "$D"
(cd "$t/elsewhere" && py "$pydir/calls.py" "$D/out.jsonl") >"$t/stdout" 2>"$t/stderr" || die "calls.py failed: $(cat "$t/stderr")"
mkdir -p "$out/calls" && norm "$out/calls/stdout" <"$t/stdout"
