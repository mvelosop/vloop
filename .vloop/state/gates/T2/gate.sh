#!/bin/sh
# Gate T2 — P2: Copilot plan sessions. agent.plan accepts copilot (from the
# file and VLOOP_AGENT_PLAN) and refuses anything else. A fixture run with
# agent.plan = copilot and model.plan set (the stub copilot, see
# stub-copilot.sh; gate review, work, review and verify on the stub claude)
# plans in one copilot plan session, `/vloop:plan <brief>` with the model set,
# saying `planning from <brief> using <model>`, and does the one task: the
# branch holds the plan commit and T1's; the session's record is phase plan,
# agent copilot, the model, session/v4 valid, with its credits, cost_source
# and denials; the stub's argv carries the whole review deny list,
# shell(rm:*) included; its hooks file runs `fence hook --phase plan`, and
# that hook answers the worked example's plan decisions. A plan sent back by
# the gate review is planned again on copilot. Copilot plan, work, review,
# verify and gate-review in one run each keep their argv and phase; with
# agent.plan = claude, the claude plan session keeps the plan fence. Around
# the session nothing changes but the exit line: a copilot plan session that
# exits 3 halts with `(copilot exited 3)`, exit 1, a claude one with
# `(claude exited 3)`; no plan halts with exit 1; a moved HEAD with exit 9.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
cd "$top" || fail "repo root"
newbin
PLC="agent.plan copilot
model.plan $MODEL"

# The agent key.
CFG=; fixture
get() { k=$1; shift; (cd "$t/repo" && venv "$@" "$S" config get "$k" 2>&1); }
set_() { (cd "$t/repo" && venv "$S" config set "$1" "$2" >"$t/set.out" 2>&1); }
set_ agent.plan copilot || fail "vloop config set agent.plan copilot failed: $(cat "$t/set.out")"
[ "$(get agent.plan)" = copilot ] || fail "agent.plan set to copilot reads '$(get agent.plan)'"
set_ agent.plan claude || fail "vloop config set agent.plan claude failed: $(cat "$t/set.out")"
[ "$(get agent.plan)" = claude ] || fail "agent.plan set to claude reads '$(get agent.plan)'"
[ "$(get agent.plan VLOOP_AGENT_PLAN=copilot)" = copilot ] ||
  fail "VLOOP_AGENT_PLAN=copilot does not set agent.plan: '$(get agent.plan VLOOP_AGENT_PLAN=copilot)'"
set_ agent.plan codex; c=$?
[ "$c" = 2 ] || fail "vloop config set agent.plan codex exited $c, want 2: $(cat "$t/set.out")"
set_ agent.work codex; c=$?
[ "$c" = 2 ] || fail "vloop config set agent.work codex exited $c, want 2: $(cat "$t/set.out")"

# A run with a copilot plan session.
CFG=$PLC; fixture
runfix "$t/r1"
[ "$rc" = 0 ] || fail "a run with agent.plan = copilot exited $rc: $(tail -8 "$t/r1")"
grep -qE "^[[:space:]]*planning from $BRIEF using $MODEL\$" "$t/r1" ||
  fail "the run does not print 'planning from $BRIEF using $MODEL': $(head -20 "$t/r1")"
[ "$(task status T1)" = done ] || fail "T1 is $(task status T1) after the run, want done"
onecommit "the run with agent.plan = copilot"
[ "$(venv git -C "$t/repo" log --format=%H -- .vloop/state/state.json | wc -l | tr -d ' ')" -ge 1 ] ||
  fail "no commit on the branch carries the plan: $(venv git -C "$t/repo" log --oneline)"
[ "$(venv git -C "$t/repo" status --porcelain | grep -c .)" = 0 ] || fail "the run left the tree dirty: $(venv git -C "$t/repo" status --porcelain)"
[ "$(ev "$t/r1" '^session copilot-plan ' | grep -cxF "session copilot-plan $BRIEF model=$MODEL")" = 1 ] ||
  fail "want one copilot plan session of /vloop:plan $BRIEF with model $MODEL: $(ev "$t/r1" '^session ')"
[ "$(ev "$t/r1" '^session copilot-' | grep -c .)" = 1 ] || fail "a copilot session other than the plan ran: $(ev "$t/r1" '^session ')"
ev "$t/r1" '^session plan ' >/dev/null && fail "a claude plan session ran with agent.plan = copilot: $(ev "$t/r1" '^session plan ')"
ev "$t/r1" '^fence gate-review - ok$' >/dev/null || fail "the claude gate-review session did not run under the gate-review fence: $(ev "$t/r1" '^(fence|session) ')"
ev "$t/r1" '^fence work T1 ok$' >/dev/null || fail "the claude work session did not run under the work fence: $(ev "$t/r1" '^(fence|session) ')"
ev "$t/r1" '^fence review T1 ok$' >/dev/null || fail "the claude review session did not run under the review fence: $(ev "$t/r1" '^(fence|session) ')"
ev "$t/r1" '^fence verify - ok$' >/dev/null || fail "the claude verify session did not run under the verify fence: $(ev "$t/r1" '^(fence|session) ')"
for p in 'shell(rm:*)' 'shell(gh:*)' 'shell(git push:*)' 'shell(git commit:*)' 'shell(curl:*)' 'url'; do
  ev "$t/r1" '^deny-tool plan ' | grep -qxF "deny-tool plan $p" ||
    fail "the plan session's argv lost --deny-tool=$p: $(ev "$t/r1" '^deny-tool plan ')"
done
ev "$t/r1" '^hooks plan ' | grep -qxE 'hooks plan [1-9][0-9]* bash=yes powershell=yes' ||
  fail "the plan session found no preToolUse hook with both command forms in COPILOT_HOME/hooks/: $(ev "$t/r1" '^hooks ')"
ev "$t/r1" '^hookfile plan ' | grep -qxF 'hookfile plan plan' ||
  fail "the plan session's hooks file does not run 'fence hook --phase plan' alone: $(ev "$t/r1" '^hookfile ')"
ev "$t/r1" '^hook-phase plan ' | grep -qxF 'hook-phase plan plan' ||
  fail "the plan session's hook is not run with --phase plan: $(ev "$t/r1" '^hook-phase ')"
echo "$PLHOOK" | while IFS= read -r want; do
  ev "$t/r1" "^hook plan ${want%% *} " | grep -qxF "hook plan $want" ||
    fail "the plan session's hook answered '$(ev "$t/r1" "^hook plan ${want%% *} ")', want 'hook plan $want'"
done || exit 1
recs=$(phaserecs "$t/r1" plan)
[ "$(echo "$recs" | grep -c .)" = 1 ] || fail "want one plan session record, got: $recs"
f=$recs
[ "$("$JQ" -r .agent "$f")" = copilot ] && [ "$("$JQ" -r .model "$f")" = "$MODEL" ] ||
  fail "the plan session record is $("$JQ" -c '{phase, agent, model}' "$f"), want phase plan, agent copilot, model $MODEL"
[ "$("$JQ" -r .schema "$f")" = session/v4 ] || fail "the plan session record's schema is $("$JQ" -r .schema "$f"), want session/v4"
(cd "$t/repo" && venv "$S" schema validate session/v4 "$f" >"$t/val" 2>&1) || fail "the plan session record is not valid session/v4: $(cat "$t/val")"
"$JQ" -e '.credits == 8.4 and .cost_source == "estimated"' "$f" >/dev/null ||
  fail "the plan session record's credits and cost_source are $("$JQ" -c '{credits, cost_source}' "$f"), want 8.4 and estimated"
"$JQ" -e '(.permission_denials | length) == 1 and .permission_denials[0].tool_name == "bash" and (.permission_denials[0].message | contains("vloop task drop T2 is denied by the plan fence (Bash(vloop task drop:*))"))' "$f" >/dev/null ||
  fail "the plan session record's permission_denials are $("$JQ" -c .permission_denials "$f"), want the one bash denial"
recs=$(phaserecs "$t/r1" gate-review)
[ -n "$recs" ] && [ "$("$JQ" -r .agent $recs)" = claude ] || fail "the gate-review session record is not claude's: $recs"

# The same from VLOOP_AGENT_PLAN.
CFG="model.plan $MODEL"; fixture
runfix "$t/r0" VLOOP_AGENT_PLAN=copilot
[ "$rc" = 0 ] || fail "a run with VLOOP_AGENT_PLAN=copilot exited $rc: $(tail -8 "$t/r0")"
[ "$(ev "$t/r0" '^session copilot-plan ' | grep -cxF "session copilot-plan $BRIEF model=$MODEL")" = 1 ] ||
  fail "VLOOP_AGENT_PLAN=copilot: want one copilot plan session: $(ev "$t/r0" '^session ')"

# A plan sent back by the gate review is planned again on copilot.
CFG="$PLC
agent.gate-review copilot
model.gate-review $MODEL"; fixture
: >"$t/stub/copilot.gr-fail-once"
runfix "$t/f1"
[ "$rc" = 0 ] || fail "a copilot plan sent back once: vloop run exited $rc, want 0: $(tail -8 "$t/f1")"
[ "$(task status T1)" = done ] || fail "after the second round, T1 is $(task status T1), want done"
[ "$(ev "$t/f1" '^session copilot-plan ' | grep -cxF "session copilot-plan $BRIEF model=$MODEL")" = 2 ] ||
  fail "want two copilot plan sessions, the revision included: $(ev "$t/f1" '^session ')"
ev "$t/f1" '^session plan ' >/dev/null && fail "the revision ran on claude: $(ev "$t/f1" '^session ')"
grep -qF 'gate T1 (wrong-reason): the gate passes on a stub' "$t/f1.problems" ||
  fail "the second copilot plan session was not sent the finding 'gate T1 (wrong-reason): the gate passes on a stub': $(cat "$t/f1.problems")"
[ "$(phaserecs "$t/f1" plan | grep -c .)" = 2 ] || fail "want two plan session records: $(phaserecs "$t/f1" plan)"
for f in $(phaserecs "$t/f1" plan); do
  "$JQ" -e --arg m "$MODEL" '.schema == "session/v4" and .phase == "plan" and .agent == "copilot" and .model == $m' "$f" >/dev/null ||
    fail "a plan session record of the two rounds is $("$JQ" -c '{schema, phase, agent, model}' "$f")"
done
"$JQ" -e '.gate_review.rounds == 2 and .gate_review.verdict == "PASS"' "$t/f1.state" >/dev/null 2>&1 ||
  fail "the plan's gate_review is $("$JQ" -c .gate_review "$t/f1.state" 2>&1), want rounds 2, verdict PASS"

# Copilot plan, work, review, verify and gate-review in one run: each keeps its argv and phase.
CFG="$PLC
agent.gate-review copilot
model.gate-review $MODEL
agent.work copilot
model.work $MODEL
agent.review copilot
model.review $MODEL
agent.verify copilot
model.verify $MODEL"; fixture
runfix "$t/r2"
[ "$rc" = 0 ] || fail "a run with every agent on copilot exited $rc: $(tail -8 "$t/r2")"
[ "$(task status T1)" = done ] || fail "every agent on copilot: T1 is $(task status T1), want done"
for r in plan work review verify gate-review; do
  ev "$t/r2" "^session copilot-$r " >/dev/null || fail "no copilot $r session: $(ev "$t/r2" '^session ')"
  ev "$t/r2" "^hookfile $r " | grep -qxF "hookfile $r $r" || fail "the copilot $r session's hooks file is not its phase's: $(ev "$t/r2" '^hookfile ')"
  ev "$t/r2" "^deny-tool $r " | sed "s/^deny-tool $r //" | LC_ALL=C sort >"$t/dt.$r"
done
ev "$t/r2" '^session (plan|work|review|verify|gate-review) ' >/dev/null && fail "a claude session ran with every agent on copilot: $(ev "$t/r2" '^session ')"
[ -s "$t/dt.review" ] && cmp -s "$t/dt.review" "$t/dt.plan" ||
  fail "the plan session's --deny-tool list differs from the review session's: review '$(tr '\n' ' ' <"$t/dt.review")', plan '$(tr '\n' ' ' <"$t/dt.plan")'"
cmp -s "$t/dt.review" "$t/dt.verify" || fail "the verify session's --deny-tool list differs from the review session's"
cmp -s "$t/dt.review" "$t/dt.gate-review" || fail "the gate-review session's --deny-tool list differs from the review session's"
grep -qxF 'shell(rm:*)' "$t/dt.work" && fail "the copilot work session's argv holds --deny-tool=shell(rm:*)"
grep -vxF 'shell(rm:*)' "$t/dt.review" | cmp -s - "$t/dt.work" || fail "the copilot work session's --deny-tool list is not the review list without shell(rm:*)"
for want in 'shgh deny 0' 'diff allow 0' 'rmfile deny 0' 'srcrel deny 0' 'tmprel allow 0'; do
  ev "$t/r2" "^hook review ${want%% *} " | grep -qxF "hook review $want" ||
    fail "the review session's hook answered '$(ev "$t/r2" "^hook review ${want%% *} ")', want 'hook review $want'"
done
for want in 'rmfile allow 0' 'rmrf deny 0' 'srcrel allow 0' 'stateabs deny 0'; do
  ev "$t/r2" "^hook work ${want%% *} " | grep -qxF "hook work $want" ||
    fail "the work session's hook answered '$(ev "$t/r2" "^hook work ${want%% *} ")', want 'hook work $want'"
done
for want in 'build allow 0' 'cdout deny 0' 'gitc deny 0' 'staterel deny 0' 'skillother deny 0'; do
  ev "$t/r2" "^hook verify ${want%% *} " | grep -qxF "hook verify $want" ||
    fail "the verify session's hook answered '$(ev "$t/r2" "^hook verify ${want%% *} ")', want 'hook verify $want'"
done
for want in 'gate allow 0' 'gateedit deny 0' 'staterel deny 0' 'verdictrel allow 0' 'skillother deny 0'; do
  ev "$t/r2" "^hook gate-review ${want%% *} " | grep -qxF "hook gate-review $want" ||
    fail "the gate-review session's hook answered '$(ev "$t/r2" "^hook gate-review ${want%% *} ")', want 'hook gate-review $want'"
done
[ "$(phaserecs "$t/r2" plan | grep -c .)" = 1 ] || fail "every agent on copilot: want one plan session record"

# Claude plan sessions: unchanged, under the plan fence.
CFG=; fixture
runfix "$t/r3"
[ "$rc" = 0 ] || fail "a run with a claude plan session exited $rc: $(tail -8 "$t/r3")"
grep -qE "^[[:space:]]*planning from $BRIEF using opus\$" "$t/r3" || fail "the claude run does not print 'planning from $BRIEF using opus': $(head -20 "$t/r3")"
ev "$t/r3" "^session plan $BRIEF model=opus\$" >/dev/null || fail "no claude plan session of $BRIEF with model opus: $(ev "$t/r3" '^session ')"
ev "$t/r3" "^fence plan $BRIEF ok\$" >/dev/null || fail "the claude plan session did not run under the plan fence: $(ev "$t/r3" '^fence ')"
ev "$t/r3" '^session copilot-' >/dev/null && fail "a copilot session ran with every agent on claude: $(ev "$t/r3" '^session ')"
f=$(phaserecs "$t/r3" plan)
[ -n "$f" ] && "$JQ" -e '.agent == "claude"' $f >/dev/null || fail "the claude plan session record is $("$JQ" -c '{schema, agent}' $f 2>&1)"

# A claude plan session that exits 3: its line, as at the base.
CFG=; fixture
: >"$t/stub/claude.plan-exit3"
runfix "$t/c3"
rd=$(rundir)
[ "$rc" = 1 ] || fail "a claude plan session that exits 3: vloop run exited $rc, want 1: $(tail -5 "$t/c3")"
grep -qxF "$(planfail claude 3 "$rd")" "$t/c3" || fail "no '$(planfail claude 3 "$rd")': $(tail -5 "$t/c3")"

# A copilot plan session that exits 3: the line names copilot.
CFG=$PLC; fixture
: >"$t/stub/copilot.plan-exit3"
runfix "$t/x3"
rd=$(rundir)
[ "$rc" = 1 ] || fail "a copilot plan session that exits 3: vloop run exited $rc, want 1: $(tail -5 "$t/x3")"
[ -n "$rd" ] || fail "a copilot plan session that exits 3 left no run folder"
grep -qxF "$(planfail copilot 3 "$rd")" "$t/x3" || fail "no '$(planfail copilot 3 "$rd")': $(tail -5 "$t/x3")"
grep -qF '(claude exited' "$t/x3" && fail "a copilot plan session's failure names claude: $(grep -F '(claude exited' "$t/x3")"
ev "$t/x3" '^session (gate-review|work|review|verify|copilot-gate-review|copilot-work) ' >/dev/null && fail "a session ran after the plan session failed: $(ev "$t/x3" '^session ')"

# A copilot plan session that writes no plan.
CFG=$PLC; fixture
: >"$t/stub/copilot.plan-noplan"
runfix "$t/np"
[ "$rc" = 1 ] || fail "a copilot plan session that writes no plan: vloop run exited $rc, want 1: $(tail -5 "$t/np")"
grep -qxF 'vloop: planning produced no .vloop/state/state.json' "$t/np" || fail "no 'vloop: planning produced no .vloop/state/state.json': $(tail -5 "$t/np")"
ev "$t/np" '^session (gate-review|work) ' >/dev/null && fail "a session ran after no plan: $(ev "$t/np" '^session ')"

# A copilot plan session that moves HEAD: the run halts, exit 9.
CFG=$PLC; fixture
: >"$t/stub/copilot.plan-moveref"
runfix "$t/m1"
[ "$rc" = 9 ] || fail "a copilot plan session that moved HEAD: vloop run exited $rc, want 9: $(tail -8 "$t/m1")"
grep -qxF 'vloop: REFS MOVED plan — the planning session changed git refs; nothing was committed' "$t/m1" ||
  fail "no 'vloop: REFS MOVED plan — the planning session changed git refs; nothing was committed': $(tail -8 "$t/m1")"
ev "$t/m1" '^session (gate-review|work|review|verify) ' >/dev/null && fail "a session ran after the plan session moved HEAD: $(ev "$t/m1" '^session ')"
echo "gate T2: ok"
