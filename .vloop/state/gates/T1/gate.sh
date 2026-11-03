#!/bin/sh
# Gate T1 — F1's re-judge, in the work phase: `vloop run` of a one-task plan,
# every phase on the stub copilot (lib.sh, stub-copilot.sh), the work
# session's JSON Lines carrying the tool events each case names. The judge's
# reasons are `vloop fence check`'s for the same call. Covers the worked
# example's first three lines, and the contract's reading of a call: a call
# with no completion ran; a file tool's path, the first line of a command cut
# to 200 runes with an ellipsis, or nothing; the hook's empty-command rule;
# arguments as a JSON string holding an object; an unknown tool; $HOME masked
# as ~ on stderr and in the run log; the count in the halt.
G=.vloop/state/gates/T1
. "$G/lib.sh"
newbin
BYP="the session ran 1 call(s) its fence denies; nothing from it was used"

# 1. A bash call "git commit -m x" that completed without error.
fixture
{ start c1 bash "$(bashargs 'git commit -m x')"; done_ c1; } >"$t/stub/calls.work"
runfix w1
R=$(reason work bash 'git commit -m x')
[ "$rc" = 9 ] || fail "a work session that ran 'git commit -m x': want exit 9, got $(show w1)"
errline w1 "vloop: bash git commit -m x — $R" || fail "a work session that ran 'git commit -m x': no stderr line 'vloop: bash git commit -m x — $R': $(show w1)"
halted w1 "FENCE BYPASSED work — $BYP" || fail "a work session that ran 'git commit -m x': no halt 'FENCE BYPASSED work — $BYP': $(show w1)"
[ "$(commits_with T1.out)" = 0 ] || fail "a work session that ran 'git commit -m x': its work was committed"
runlog | grep -qF "vloop: bash git commit -m x — $R" || fail "a work session that ran 'git commit -m x': the run log lacks the call's line"
runlog | grep -qF "FENCE BYPASSED work — $BYP" || fail "a work session that ran 'git commit -m x': the run log lacks the halt"

# 2. The same call, denied by the hook.
fixture
{ start c1 bash "$(bashargs 'git commit -m x')"; denied c1; } >"$t/stub/calls.work"
runfix w2
[ "$rc" = 0 ] || fail "a work session whose 'git commit -m x' the hook denied: want the run to go on to exit 0, got $(show w2)"
grep -qF 'FENCE BYPASSED' "$t/w2.err" && fail "a work session whose 'git commit -m x' the hook denied: FENCE BYPASSED was printed"
grep -q '^vloop: bash ' "$t/w2.err" && fail "a work session whose 'git commit -m x' the hook denied: a call was named: $(grep '^vloop: bash ' "$t/w2.err")"
[ "$(commits_with T1.out)" = 1 ] || fail "a work session whose 'git commit -m x' the hook denied: T1's work was not committed"

# 3. A bash call "ls" only.
fixture
{ start c1 bash "$(bashargs ls)"; done_ c1; start c2 view "$(fileargs docs/x.md)"; done_ c2; } >"$t/stub/calls.work"
runfix w3
[ "$rc" = 0 ] || fail "a work session that ran 'ls': want the run to go on to exit 0, got $(show w3)"
grep -qF 'FENCE BYPASSED' "$t/w3.err" && fail "a work session that ran 'ls': FENCE BYPASSED was printed"
[ "$(commits_with T1.out)" = 1 ] || fail "a work session that ran 'ls': T1's work was not committed"

# 4. Eight calls the fence denies that ran, among calls it allows and one the hook denied.
fixture
e=; i=0; while [ "$i" -lt 250 ]; do e="${e}é"; i=$((i + 1)); done
long="git commit -m $e"
cut="git commit -m "; i=0; while [ "$i" -lt 186 ]; do cut="${cut}é"; i=$((i + 1)); done
nl='git push
echo y'
{
  start c1 bash "$(bashargs 'git commit -m x')"
  start c2 create "$(fileargs .vloop/state/state.json)"; done_ c2
  start c3 bash "$(bashargs "$nl")"; done_ c3
  start c4 bash "$(bashargs "$long")"; done_ c4
  start c5 bash "$(bashargs '')"; done_ c5
  start c6 create "$(fileargs "$H/x.txt")"; done_ c6
  start c7 frobnicate '{}'; done_ c7
  start c8 bash "$("$JQ" -cn '{command: "git tag v1"}' | "$JQ" -R .)"; done_ c8
  start c9 view "$(fileargs docs/x.md)"; done_ c9
  start c10 bash "$(bashargs ls)"; done_ c10
  start c11 bash "$(bashargs 'git push')"; denied c11
} >"$t/stub/calls.work"
runfix w4
[ "$rc" = 9 ] || fail "a work session that ran eight denied calls: want exit 9, got $(show w4)"
for w in \
  "vloop: bash git commit -m x — $(reason work bash 'git commit -m x')" \
  "vloop: create .vloop/state/state.json — $(reason work create --path .vloop/state/state.json)" \
  "vloop: bash git push — $(reason work bash "$nl")" \
  "vloop: bash $cut… — $(reason work bash "$long")" \
  "vloop: bash — the bash call has no command the hook can read" \
  "vloop: create ~/x.txt — $(mask "$(reason work create --path "$H/x.txt")")" \
  "vloop: frobnicate — $(reason work frobnicate)" \
  "vloop: bash git tag v1 — $(reason work bash 'git tag v1')"; do
  errline w4 "$w" || fail "a work session that ran eight denied calls: no stderr line '$w': $(show w4)"
done
[ "$(grep -cE '^vloop: (bash|create|frobnicate|view) ' "$t/w4.err")" = 8 ] ||
  fail "a work session that ran eight denied calls: want exactly eight call lines on stderr: $(show w4)"
grep -qF 'vloop: view docs/x.md' "$t/w4.err" && fail "an allowed view call was named"
grep -qx 'vloop: bash ls — .*' "$t/w4.err" && fail "an allowed 'ls' call was named"
halted w4 "FENCE BYPASSED work — the session ran 8 call(s) its fence denies; nothing from it was used" ||
  fail "a work session that ran eight denied calls: no halt 'FENCE BYPASSED work — the session ran 8 call(s) its fence denies; nothing from it was used': $(show w4)"
grep -qF "$H" "$t/w4.err" && fail "stderr holds the HOME path unmasked: $(grep -F "$H" "$t/w4.err")"
runlog | grep -qF "vloop: create ~/x.txt — " || fail "the run log lacks the masked line 'vloop: create ~/x.txt — …'"
runlog | grep -qF "$H/x.txt" && fail "the run log holds the HOME path unmasked"
[ "$(commits_with T1.out)" = 0 ] || fail "a work session that ran eight denied calls: its work was committed"
echo "gate T1: ok"
