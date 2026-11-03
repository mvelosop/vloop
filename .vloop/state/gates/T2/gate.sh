#!/bin/sh
# Gate T2 — F1's halt in every other phase, and F2. `vloop run` (and `vloop
# brief review`) in fixture repositories, every phase on the stub copilot
# (lib.sh, stub-copilot.sh). The worked example's phase lines: a session of
# plan, gate-review, review, second-review (run.second-review = "record"),
# verify and review-brief that ran a bash call "git push" halts with exit 9,
# the call named with `vloop fence check`'s reason, and nothing from it used.
# Then F2: the hook's binary replaced, or removed, during a work session
# halts with HOOK CHANGED, before F1; a session that printed no result event
# is re-judged; a call that got past the hook is named before a moved ref.
G=.vloop/state/gates/T2
. "$G/lib.sh"
newbin
BYP="the session ran 1 call(s) its fence denies; nothing from it was used"
push() { { start c1 bash "$(bashargs 'git push')"; done_ c1; } >"$t/stub/calls.$1"; }
# bypassed <name> <phase>: exit 9, the call's line with the phase's reason, the halt.
bypassed() {
  [ "$rc" = 9 ] || fail "a $2 session that ran 'git push': want exit 9, got $(show "$1")"
  errline "$1" "vloop: bash git push — $(reason "$2" bash 'git push')" ||
    fail "a $2 session that ran 'git push': no stderr line 'vloop: bash git push — $(reason "$2" bash 'git push')': $(show "$1")"
  halted "$1" "FENCE BYPASSED $2 — $BYP" || fail "a $2 session that ran 'git push': no halt 'FENCE BYPASSED $2 — $BYP': $(show "$1")"
}
nowork() { grep -q '^session work ' "$t/$1.ev" && fail "$2: a work session started after the halt"; return 0; }

# plan: no plan commit, no work.
fixture; push plan; runfix p1
bypassed p1 plan
[ "$(commits_with .vloop/state/state.json)" = 0 ] || fail "a plan session that ran 'git push': the plan was committed"
nowork p1 "a plan session that ran 'git push'"
runlog | grep -qF "FENCE BYPASSED plan — $BYP" || fail "a plan session that ran 'git push': the run log lacks the halt"

# gate-review: no plan commit, no work.
fixture; push gate-review; runfix g1
bypassed g1 gate-review
[ "$(commits_with .vloop/state/state.json)" = 0 ] || fail "a gate-review session that ran 'git push': the plan was committed"
nowork g1 "a gate-review session that ran 'git push'"

# review: no new commit.
fixture; push review; runfix r1
bypassed r1 review
[ "$(commits_with T1.out)" = 0 ] || fail "a review session that ran 'git push': T1's work was committed"
runlog | grep -qF "vloop: bash git push — " || fail "a review session that ran 'git push': the run log lacks the call's line"

# second-review, recorded only: no new commit.
CFG="run.second-review record"; fixture; push second-review; runfix s1
grep -q '^session second-review T1' "$t/s1.ev" || fail "run.second-review = record: no second-review session ran: $(cat "$t/s1.ev")"
bypassed s1 second-review
[ "$(commits_with T1.out)" = 0 ] || fail "a second-review session that ran 'git push': T1's work was committed"

# verify, the last task done: the verify report is not used.
fixture; push verify; runfix v1
grep -q '^session verify ' "$t/v1.ev" || fail "no verify session ran: $(cat "$t/v1.ev")"
bypassed v1 verify
for f in "$t/repo"/.vloop/state/runs/*/*/reports/verify.json; do
  [ -f "$f" ] && fail "a verify session that ran 'git push': its report was kept as reports/verify.json"
done

# review-brief (vloop brief review): the brief restored, no verdict printed.
fixture; push review-brief; echo "$BRIEF" >"$t/stub/touch.review-brief"; brev b1
grep -q '^session review-brief ' "$t/b1.ev" || fail "no review-brief session ran: $(cat "$t/b1.ev")"
bypassed b1 review-brief
cmp -s "$G/brief.md" "$t/repo/$BRIEF" || fail "a review-brief session that ran 'git push': the brief was not restored"
grep -q '^verdict: ' "$t/b1.out" && fail "a review-brief session that ran 'git push': its verdict was printed: $(cat "$t/b1.out")"

# F2: the hook's binary replaced during a work session. vloop runs from a copy
# under HOME, so the binary the hooks file names is that copy.
for how in replace remove; do
  fixture; mkdir -p "$H/bin" && cp "$S" "$H/bin/vloop" || fail "copy the binary"
  echo "$how" >"$t/stub/hook.work"
  runfix h-$how "$H/bin/vloop"
  P=$(sed -n 's/^hookbin work //p' "$t/h-$how.ev")
  [ "$P" = "$H/bin/vloop" ] || fail "the work session's hooks file runs '$P', want the running binary $H/bin/vloop"
  M="HOOK CHANGED work — the binary the fence hook runs changed during the session: $(mask "$P"); nothing from it was used"
  [ "$rc" = 9 ] || fail "the hook's binary ${how}d during a work session: want exit 9, got $(show h-$how)"
  halted h-$how "$M" || fail "the hook's binary ${how}d during a work session: no halt '$M': $(show h-$how)"
  [ "$(commits_with T1.out)" = 0 ] || fail "the hook's binary ${how}d during a work session: T1's work was committed"
  runlog | grep -qF "$M" || fail "the hook's binary ${how}d during a work session: the run log lacks '$M'"
done

# F2 before F1: the binary replaced and a denied call that ran, in one session.
fixture; mkdir -p "$H/bin" && cp "$S" "$H/bin/vloop" || fail "copy the binary"
echo replace >"$t/stub/hook.work"
{ start c1 bash "$(bashargs 'git commit -m x')"; done_ c1; } >"$t/stub/calls.work"
runfix h2 "$H/bin/vloop"
[ "$rc" = 9 ] || fail "a work session that changed the hook's binary and ran 'git commit -m x': want exit 9, got $(show h2)"
case $(firstof h2 'HOOK CHANGED|FENCE BYPASSED') in
  *"HOOK CHANGED work — "*) ;;
  *) fail "a work session that changed the hook's binary and ran 'git commit -m x': HOOK CHANGED is not the first halt: $(show h2)";;
esac

# A work session that printed no result event and exited 1 is still re-judged;
# its call has no completion.
fixture; touch "$t/stub/silent.work"
start c1 bash "$(bashargs 'git commit -m x')" >"$t/stub/calls.work"
runfix n1
[ "$rc" = 9 ] || fail "a work session with no result event that ran 'git commit -m x': want exit 9, got $(show n1)"
halted n1 "FENCE BYPASSED work — $BYP" || fail "a work session with no result event: no halt 'FENCE BYPASSED work — $BYP': $(show n1)"

# A call that got past the hook and also moved a ref: FENCE BYPASSED comes first.
fixture; touch "$t/stub/moveref.work"
{ start c1 bash "$(bashargs 'git commit -m x')"; done_ c1; } >"$t/stub/calls.work"
runfix m1
[ "$rc" = 9 ] || fail "a work session that moved a ref and ran 'git commit -m x': want exit 9, got $(show m1)"
case $(firstof m1 'FENCE BYPASSED|REFS MOVED') in
  *"FENCE BYPASSED work — $BYP") ;;
  *) fail "a work session that moved a ref and ran 'git commit -m x': FENCE BYPASSED is not named first: $(show m1)";;
esac
[ "$(commits_with T1.out)" = 0 ] || fail "a work session that moved a ref and ran 'git commit -m x': T1's work was committed"
echo "gate T2: ok"
