#!/bin/sh
# Gate T3 — P2, the kit's reports in Go. For every report kind, on fixture
# repositories this gate carries (repos/), the starting point's report.py —
# taken from the base commit, so this gate outlives its deletion — prints the
# reference; compose.py replaces its first table with P2's Sessions block(s),
# composed from `vloop metrics <id> --by session` (the binary built from this
# tree) and the records read here; the kit's output must equal it byte for byte.
G=.vloop/state/gates/T3
. "$G/lib.sh"
BASE=d7b9cbb579f82f8626c9211e5ce8f2f7e8f720aa

mkdir -p "$t/py" || fail mkdir
git -C "$top" show "$BASE:tools/copilot-acceptance/report.py" >"$t/py/report.py" || fail "git show the base's report.py"
go build -buildvcs=false -o "$t/kit" ./tools/copilot-acceptance/kit >"$t/kit.out" 2>&1 ||
  fail "go build ./tools/copilot-acceptance/kit: $(tail -n 5 "$t/kit.out")"
kit() { (cd "$t/elsewhere" && HOME="$t/home" "$t/kit" "$@"); }

# fixture NAME: the fixture repository repos/NAME as a fresh git repository; sets R.
fixture() {
  newrepo "$1"
  cp -R "$G/repos/$1/." "$R/" || fail "copy fixture $1"
  commitall "$R"
}

# case KIND REPO RC MINROWS: the kit's report equals the composed reference.
case_() {
  k=$1 name=$2 code=$3 min=$4
  fixture "$name"
  (cd "$t/elsewhere" && py "$t/py/report.py" "$k" "$R" "$S" "$code") >"$t/py.out" 2>"$t/py.err" ||
    fail "the reference report.py $k on $name failed: $(cat "$t/py.err")"
  py "$G/compose.py" full "$t/py.out" "$R" "$S" "$min" >"$t/want" || fail "compose $k on $name"
  kit report "$k" "$R" "$S" "$code" >"$t/got" 2>"$t/err"; rc=$?
  [ "$rc" = 0 ] || fail "kit report $k on $name: exit $rc, want 0: $(cat "$t/err")"
  cmp -s "$t/got" "$t/want" || fail "kit report $k on $name differs from the reference (python's output with the Sessions block):
$(diff "$t/want" "$t/got" | head -n 40)"
  rm -rf "$R"
}

case_ run r1 0 2
case_ work r1 1 2
case_ verify r1 0 2
case_ gate-review r1 2 2
case_ plan r1 0 2
case_ second r2 0 2
# vloop metrics exits 1 on this one: the fence holds its error, the rest follows.
case_ work r3 0 -
grep -Fxq '```' "$t/got" && grep -q '^vloop metrics exited 1: ' "$t/got" || fail "kit report work on r3: no 'vloop metrics exited 1: …' line in the fence"
# No brief with runs.
case_ run r4 0 -
grep -Fxq 'Sessions: no brief with runs.' "$t/got" || fail "kit report run on r4: no 'Sessions: no brief with runs.' line"

# A session file that is not JSON: the reference Python fails on it, so only
# the Sessions block is judged — present, verbatim, with the unreadable line.
fixture r5
py "$G/compose.py" blocks "$R" "$S" >"$t/want" || fail "compose r5"
kit report run "$R" "$S" 0 >"$t/got" 2>"$t/err"; rc=$?
[ "$rc" = 0 ] || fail "kit report run on r5: exit $rc, want 0: $(cat "$t/err")"
py -c 'import sys; g=open(sys.argv[1],encoding="utf-8").read(); w=open(sys.argv[2],encoding="utf-8").read(); sys.exit(0 if w in g else 1)' "$t/got" "$t/want" ||
  fail "kit report run on r5 does not hold the expected Sessions block:
--- want (block)
$(cat "$t/want")
--- got
$(cat "$t/got")"
grep -Fxq -- '- .vloop/state/runs/B20260601-0000-kunread/20260601-100000/sessions/002-work.json: unreadable' "$t/got" ||
  fail "kit report run on r5: the unreadable record is not listed"

# Arguments: exit 1, a one-line message on stderr, nothing on stdout.
refuse() {
  n=$1; shift
  kit "$@" >"$t/out" 2>"$t/err"; rc=$?
  [ "$rc" = 1 ] || fail "kit $n: exit $rc, want 1"
  [ -s "$t/err" ] && [ "$(wc -l <"$t/err" | tr -d ' ')" -le 1 ] || fail "kit $n: want a one-line message on stderr, got: $(cat "$t/err")"
  [ ! -s "$t/out" ] || fail "kit $n: printed on stdout: $(cat "$t/out")"
}
refuse "report (no kind)" report
refuse "report bogus" report bogus "$t" "$S" 0
refuse "report run (arguments missing)" report run "$t"
refuse "(no command)"
refuse "bogus" bogus

echo "gate T3: ok"
