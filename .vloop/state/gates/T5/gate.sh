#!/bin/sh
# Gate T5 — P4, the scripts and the README. report.py and hook.py are gone; no
# script and not the README names python3, report.py or hook.py; fail-open.sh
# is the base's. Then the scripts run with a python3 that exits 127 first on
# PATH and this gate's stub as copilot: hook.sh started by an absolute path
# from outside the module with a relative directory, and by a relative path
# from outside the module; bypass.sh with a relative work directory. Their
# reports must equal what the starting point's Python prints on the same input.
G=.vloop/state/gates/T5
. "$G/lib.sh"
BASE=d7b9cbb579f82f8626c9211e5ce8f2f7e8f720aa
K=tools/copilot-acceptance

# ---- the files ------------------------------------------------------------------
[ ! -e "$K/report.py" ] || fail "$K/report.py still exists"
[ ! -e "$K/hook.py" ] || fail "$K/hook.py still exists"
for f in "$K"/*.sh "$K/README.md"; do
  ! grep -nE 'python3|report\.py|hook\.py' "$f" >"$t/m" || fail "$f still names python3, report.py or hook.py:
$(cat "$t/m")"
done
git -C "$top" show "$BASE:$K/fail-open.sh" >"$t/fail-open.sh" || fail "git show the base's fail-open.sh"
cmp -s "$t/fail-open.sh" "$K/fail-open.sh" || fail "$K/fail-open.sh changed"
for w in go copilot gh claude perl; do
  grep -qw -- "$w" "$K/README.md" || fail "$K/README.md does not name $w"
done
grep -qF -- '--by session' "$K/README.md" || fail "$K/README.md does not say the session tables are vloop metrics --by session"
grep -qF -- 'tools/copilot-acceptance/kit' "$K/README.md" || fail "$K/README.md does not name the Go program"
sed -n '/^#/!q; p' "$K/bypass.sh" | grep -qw perl || fail "$K/bypass.sh's header does not name perl"

# ---- the reference, with the real python3 ------------------------------------------
mkdir -p "$t/py" "$t/ref" "$t/fakepy" "$t/bin" "$t/out" "$t/out2" || fail mkdir
for f in hook.py bypass.sh; do
  git -C "$top" show "$BASE:$K/$f" >"$t/py/$f" || fail "git show the base's $f"
done
awk '/^  python3 - "\$d\/out.jsonl" <<.PY.$/ {on=1; next} on && /^PY$/ {on=0} on' "$t/py/bypass.sh" >"$t/py/calls.py"
cp "$G/copilot" "$t/bin/copilot" && chmod +x "$t/bin/copilot" || fail "stub copilot"
py "$G/tree.py" put "$G/hook" "$t/ref/fx" || fail "copy the hook fixture"
: >"$t/ref/fx/probe-hits.log" # as hook.sh does before its probes
(cd "$t/elsewhere" && py "$t/py/hook.py" forms "$t/ref/fx" "$t/bin/copilot" &&
  for m in missing exit1 notjson timeout; do py "$t/py/hook.py" failure "$t/ref/fx" "$t/bin/copilot" "$m" || exit 1; done &&
  py "$t/py/hook.py" summary "$t/ref/fx") >"$t/ref.md" 2>"$t/ref.err" || fail "the reference hook.py failed: $(cat "$t/ref.err")"
sed "s|$t/ref/fx|@DIR@|g" "$t/ref.md" >"$t/want.md"
py "$t/py/calls.py" "$G/calls-out.jsonl" >"$t/calls.want" 2>"$t/calls.err" || fail "the reference calls failed: $(cat "$t/calls.err")"

# ---- python3 unavailable --------------------------------------------------------------
printf '#!/bin/sh\necho "python3: disabled by gate T5" >&2\nexit 127\n' >"$t/fakepy/python3" && chmod +x "$t/fakepy/python3" || fail "fake python3"
P=$t/fakepy:$t/bin:$PATH

# hook.sh FIXTURE-DIR STEP: the summary it wrote equals the reference, 5 sessions started.
hookok() {
  [ -f "$1/b22-hook-summary.md" ] || fail "$2: no $1/b22-hook-summary.md"
  sed "s|$1|@DIR@|g" "$1/b22-hook-summary.md" >"$t/got.md"
  cmp -s "$t/got.md" "$t/want.md" || fail "$2: b22-hook-summary.md differs from what the base's hook.py prints:
$(diff "$t/want.md" "$t/got.md" | head -n 30)"
  [ "$(grep -c '^=== stub$' "$1/stub.log")" = 5 ] || fail "$2: want 5 copilot sessions in the stub's log"
}

# A: an absolute script path, from a directory outside the module, a relative directory.
py "$G/tree.py" put "$G/hook" "$t/out/fx" || fail "copy the hook fixture"
(cd "$t/out" && PATH=$P sh "$top/$K/hook.sh" fx) >"$t/a.log" 2>&1 || fail "hook.sh fx (from outside the module): exit $?:
$(tail -n 20 "$t/a.log")"
hookok "$t/out/fx" "hook.sh with a relative directory"

# B: a relative script path, from the repository's parent directory (outside the module).
py "$G/tree.py" put "$G/hook" "$t/out2/fx" || fail "copy the hook fixture"
parent=$(dirname "$top") name=$(basename "$top")
[ ! -f "$parent/go.mod" ] || fail "the repository's parent holds a go.mod; this gate needs it outside any module"
(cd "$parent" && PATH=$P sh "$name/$K/hook.sh" "$t/out2/fx") >"$t/b.log" 2>&1 || fail "hook.sh started by a relative path: exit $?:
$(tail -n 20 "$t/b.log")"
hookok "$t/out2/fx" "hook.sh started by a relative path"

# C: bypass.sh with a relative work directory; its call lists are the reference's.
(cd "$t/out" && PATH=$P STUB_OUT="$top/$G/calls-out.jsonl" STUB_LOG="$t/bypass-stub.log" sh "$top/$K/bypass.sh" m bw) >"$t/c.log" 2>&1 ||
  fail "bypass.sh m bw: exit $?: $(tail -n 20 "$t/c.log")"
[ -d "$t/out/bw/par-fast/repo" ] || fail "bypass.sh did not make its relative work directory under the caller's directory"
for c in par-fast par-slow mut-empty mut-delete; do
  py -c '
import sys
log, want, case = open(sys.argv[1], encoding="utf-8").read(), open(sys.argv[2], encoding="utf-8").read(), sys.argv[3]
sys.exit(0 if ("== " + case + "\n" + want + "  hook log:\n") in log else 1)' "$t/c.log" "$t/calls.want" "$c" ||
    fail "bypass.sh: the $c call list is not the one the base's heredoc prints:
--- want
$(cat "$t/calls.want")
--- got
$(cat "$t/c.log")"
done
! grep -q 'python3: disabled by gate T5' "$t/a.log" "$t/b.log" "$t/c.log" || fail "a script ran python3"

echo "gate T5: ok"
