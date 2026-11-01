#!/bin/sh
# Gate T2 — G2: Copilot gate-review sessions. agent.gate-review accepts
# copilot (from the file and VLOOP_AGENT_GATE_REVIEW); agent.plan still does
# not. A fixture run with agent.gate-review = copilot (the stub copilot, see
# stub-copilot.sh; plan, work, review and verify on the stub claude) reviews
# the gates in one copilot gate-review session with the model set, saying
# `reviewing the gates using <model> (round 1)`, and does the one task: its
# record is phase gate-review, agent copilot, the model, session/v4 valid,
# with its credits, cost_source and denials; the run folder keeps
# reports/gate-review-1.json, the PASS; the stub's argv carries the whole
# review deny list, shell(rm:*) included; its hooks file runs
# `fence hook --phase gate-review`, and that hook answers the worked example's
# gate-review decisions. Copilot work, review, verify and gate-review in one
# run each keep their argv and phase; with agent.gate-review = claude, the
# claude gate-review session keeps the gate-review fence. Around the session
# nothing changes: no verdict twice is a FAIL each round, the problem goes
# back to the planner and the plan is saved blocked, exit 2; an invalid
# verdict is a FAIL; a FAIL with a finding goes back once and a PASS then
# runs; a check-gap finding halts with exit 2 and keeps no plan; a moved HEAD
# halts with exit 9.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
cd "$top" || fail "repo root"
newbin
GRC="agent.gate-review copilot
model.gate-review $MODEL"

# The agent keys.
CFG=; fixture
get() { k=$1; shift; (cd "$t/repo" && venv "$@" "$S" config get "$k" 2>&1); }
set_() { (cd "$t/repo" && venv "$S" config set "$1" "$2" >"$t/set.out" 2>&1); }
set_ agent.gate-review copilot || fail "vloop config set agent.gate-review copilot failed: $(cat "$t/set.out")"
[ "$(get agent.gate-review)" = copilot ] || fail "agent.gate-review set to copilot reads '$(get agent.gate-review)'"
set_ agent.gate-review claude || fail "vloop config set agent.gate-review claude failed: $(cat "$t/set.out")"
[ "$(get agent.gate-review VLOOP_AGENT_GATE_REVIEW=copilot)" = copilot ] ||
  fail "VLOOP_AGENT_GATE_REVIEW=copilot does not set agent.gate-review: '$(get agent.gate-review VLOOP_AGENT_GATE_REVIEW=copilot)'"
set_ agent.gate-review codex; c=$?
[ "$c" = 2 ] || fail "vloop config set agent.gate-review codex exited $c, want 2"
set_ agent.plan copilot; c=$?
[ "$c" = 2 ] && grep -qxF 'vloop: invalid value "copilot" for agent.plan: want one of claude' "$t/set.out" ||
  fail "vloop config set agent.plan copilot: want exit 2 and 'vloop: invalid value \"copilot\" for agent.plan: want one of claude', got exit $c: $(cat "$t/set.out")"
CFG=; fixture
runfix "$t/p0" VLOOP_AGENT_PLAN=copilot
[ "$rc" != 0 ] || fail "vloop run with VLOOP_AGENT_PLAN=copilot exited 0: it must be refused"
nosession "$t/p0" "the run with VLOOP_AGENT_PLAN=copilot"

# A run with a copilot gate-review session.
CFG=$GRC; fixture
runfix "$t/r1"
[ "$rc" = 0 ] || fail "a run with agent.gate-review = copilot exited $rc: $(tail -8 "$t/r1")"
grep -qE "^[[:space:]]*reviewing the gates using $MODEL \\(round 1\\)\$" "$t/r1" ||
  fail "the run does not print 'reviewing the gates using $MODEL (round 1)': $(head -20 "$t/r1")"
[ "$(task status T1)" = done ] || fail "T1 is $(task status T1) after the run, want done"
onecommit "the run with agent.gate-review = copilot"
"$JQ" -e '.gate_review.rounds == 1 and .gate_review.verdict == "PASS"' "$t/r1.state" >/dev/null ||
  fail "the plan's gate_review is $("$JQ" -c .gate_review "$t/r1.state" 2>&1), want rounds 1, verdict PASS"
[ "$(ev "$t/r1" '^session copilot-gate-review ' | grep -cxF "session copilot-gate-review - model=$MODEL")" = 1 ] ||
  fail "want one copilot gate-review session with model $MODEL: $(ev "$t/r1" '^session ')"
[ "$(ev "$t/r1" '^session copilot-' | grep -c .)" = 1 ] || fail "a copilot session other than the gate review ran: $(ev "$t/r1" '^session ')"
ev "$t/r1" '^session gate-review ' >/dev/null && fail "a claude gate-review session ran with agent.gate-review = copilot: $(ev "$t/r1" '^session gate-review ')"
ev "$t/r1" '^session plan ' >/dev/null || fail "the plan session did not run on claude: $(ev "$t/r1" '^session ')"
ev "$t/r1" '^fence work T1 ok$' >/dev/null || fail "the claude work session did not run under the work fence: $(ev "$t/r1" '^(fence|session) ')"
ev "$t/r1" '^fence review T1 ok$' >/dev/null || fail "the claude review session did not run under the review fence: $(ev "$t/r1" '^(fence|session) ')"
ev "$t/r1" '^fence verify - ok$' >/dev/null || fail "the claude verify session did not run under the verify fence: $(ev "$t/r1" '^(fence|session) ')"
for p in 'shell(rm:*)' 'shell(gh:*)' 'shell(git push:*)' 'shell(git commit:*)' 'shell(curl:*)' 'url'; do
  ev "$t/r1" '^deny-tool gate-review ' | grep -qxF "deny-tool gate-review $p" ||
    fail "the gate-review session's argv lost --deny-tool=$p: $(ev "$t/r1" '^deny-tool gate-review ')"
done
ev "$t/r1" '^hooks gate-review ' | grep -qxE 'hooks gate-review [1-9][0-9]* bash=yes powershell=yes' ||
  fail "the gate-review session found no preToolUse hook with both command forms in COPILOT_HOME/hooks/: $(ev "$t/r1" '^hooks ')"
ev "$t/r1" '^hookfile gate-review ' | grep -qxF 'hookfile gate-review gate-review' ||
  fail "the gate-review session's hooks file does not run 'fence hook --phase gate-review' alone: $(ev "$t/r1" '^hookfile ')"
ev "$t/r1" '^hook-phase gate-review ' | grep -qxF 'hook-phase gate-review gate-review' ||
  fail "the gate-review session's hook is not run with --phase gate-review: $(ev "$t/r1" '^hook-phase ')"
echo "$GRHOOK" | while IFS= read -r want; do
  ev "$t/r1" "^hook gate-review ${want%% *} " | grep -qxF "hook gate-review $want" ||
    fail "the gate-review session's hook answered '$(ev "$t/r1" "^hook gate-review ${want%% *} ")', want 'hook gate-review $want'"
done || exit 1
recs=$(phaserecs "$t/r1" gate-review)
[ "$(echo "$recs" | grep -c .)" = 1 ] || fail "want one gate-review session record, got: $recs"
f=$recs
[ "$("$JQ" -r .agent "$f")" = copilot ] && [ "$("$JQ" -r .model "$f")" = "$MODEL" ] ||
  fail "the gate-review session record is $("$JQ" -c '{phase, agent, model}' "$f"), want phase gate-review, agent copilot, model $MODEL"
[ "$("$JQ" -r .schema "$f")" = session/v4 ] || fail "the gate-review session record's schema is $("$JQ" -r .schema "$f"), want session/v4"
(cd "$t/repo" && venv "$S" schema validate session/v4 "$f" >"$t/val" 2>&1) || fail "the gate-review session record is not valid session/v4: $(cat "$t/val")"
"$JQ" -e '.credits == 8.4 and .cost_source == "estimated"' "$f" >/dev/null ||
  fail "the gate-review session record's credits and cost_source are $("$JQ" -c '{credits, cost_source}' "$f"), want 8.4 and estimated"
"$JQ" -e '(.permission_denials | length) == 1 and .permission_denials[0].tool_name == "bash" and (.permission_denials[0].message | contains("vloop task verify T1 true --reason x is denied by the gate-review fence (Bash(vloop task verify:*))"))' "$f" >/dev/null ||
  fail "the gate-review session record's permission_denials are $("$JQ" -c .permission_denials "$f"), want the one bash denial"
[ -f "$t/r1.gr1" ] || fail "the run folder holds no reports/gate-review-1.json"
[ -f "$t/r1.gr2" ] && fail "a passed first round left reports/gate-review-2.json"
(cd "$t/repo" && venv "$S" schema validate gate-verdict/v2 "$t/r1.gr1" >"$t/val" 2>&1) || fail "reports/gate-review-1.json is not valid gate-verdict/v2: $(cat "$t/val")"
"$JQ" -e '.verdict == "PASS" and .notes == "copilot"' "$t/r1.gr1" >/dev/null || fail "reports/gate-review-1.json is not the copilot session's PASS: $(cat "$t/r1.gr1")"

# Copilot work, review, verify and gate-review in one run: each keeps its argv and phase.
CFG="$GRC
agent.work copilot
model.work $MODEL
agent.review copilot
model.review $MODEL
agent.verify copilot
model.verify $MODEL"; fixture
runfix "$t/r2"
[ "$rc" = 0 ] || fail "a run with copilot work, review, verify and gate-review exited $rc: $(tail -8 "$t/r2")"
[ "$(task status T1)" = done ] || fail "copilot work, review, verify and gate-review: T1 is $(task status T1), want done"
for r in work review verify gate-review; do
  ev "$t/r2" "^session copilot-$r " >/dev/null || fail "no copilot $r session: $(ev "$t/r2" '^session ')"
  ev "$t/r2" "^hookfile $r " | grep -qxF "hookfile $r $r" || fail "the copilot $r session's hooks file is not its phase's: $(ev "$t/r2" '^hookfile ')"
  ev "$t/r2" "^deny-tool $r " | sed "s/^deny-tool $r //" | LC_ALL=C sort >"$t/dt.$r"
done
[ -s "$t/dt.review" ] && cmp -s "$t/dt.review" "$t/dt.gate-review" ||
  fail "the gate-review session's --deny-tool list differs from the review session's: review '$(tr '\n' ' ' <"$t/dt.review")', gate-review '$(tr '\n' ' ' <"$t/dt.gate-review")'"
cmp -s "$t/dt.review" "$t/dt.verify" || fail "the verify session's --deny-tool list differs from the review session's"
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
[ -f "$t/r2.gr1" ] && [ -f "$t/r2.report" ] || fail "copilot work, review, verify and gate-review: a report is missing from the run folder"

# Claude gate-review sessions: unchanged, under the gate-review fence.
CFG=; fixture
runfix "$t/r3"
[ "$rc" = 0 ] || fail "a run with claude gate review exited $rc: $(tail -8 "$t/r3")"
ev "$t/r3" '^session gate-review - model=sonnet$' >/dev/null || fail "no claude gate-review session with model sonnet: $(ev "$t/r3" '^session ')"
ev "$t/r3" '^fence gate-review - ok$' >/dev/null || fail "the claude gate-review session did not run under the gate-review fence: $(ev "$t/r3" '^fence ')"
ev "$t/r3" '^session copilot-' >/dev/null && fail "a copilot session ran with every agent on claude"
grep -qE '^[[:space:]]*reviewing the gates using sonnet \(round 1\)$' "$t/r3" || fail "the claude run does not print 'reviewing the gates using sonnet (round 1)'"
"$JQ" -e '.verdict == "PASS" and .notes == "none"' "$t/r3.gr1" >/dev/null 2>&1 || fail "the claude gate-review session's verdict was not kept: $(cat "$t/r3.gr1" 2>&1)"

# No verdict, twice: a FAIL each round, sent back once, the plan saved blocked, exit 2.
NOTE='the gate review left no .vloop/tmp/gate-verdict.json — that is a FAIL'
CFG=$GRC; fixture
: >"$t/stub/copilot.gr-noverdict"
runfix "$t/n1"
rd=$(rundir)
[ "$rc" = 2 ] || fail "a copilot gate review that writes no verdict, twice: vloop run exited $rc, want 2: $(tail -8 "$t/n1")"
grep -qxF "vloop: the gate review failed the plan twice — see ${rd}reports/gate-review-2.json; amend the gates with vloop task verify, then re-run" "$t/n1" ||
  fail "no 'vloop: the gate review failed the plan twice — see ${rd}reports/gate-review-2.json; amend the gates with vloop task verify, then re-run': $(tail -8 "$t/n1")"
[ "$(ev "$t/n1" '^session copilot-gate-review ' | grep -c .)" = 2 ] || fail "want two copilot gate-review sessions: $(ev "$t/n1" '^session ')"
[ "$(ev "$t/n1" '^session plan ' | grep -c .)" = 2 ] || fail "want two plan sessions: $(ev "$t/n1" '^session ')"
grep -qE '^[[:space:]]*reviewing the gates using '"$MODEL"' \(round 2\)$' "$t/n1" || fail "no 'reviewing the gates using $MODEL (round 2)' line"
ev "$t/n1" '^session (work|review|verify) ' >/dev/null && fail "a work, review or verify session ran after the gate review failed: $(ev "$t/n1" '^session ')"
"$JQ" -e --arg n "$NOTE" '.schema == "gate-verdict/v2" and .verdict == "FAIL" and .notes == $n' "$t/n1.gr1" >/dev/null 2>&1 ||
  fail "reports/gate-review-1.json is not the driver's FAIL with the note '$NOTE': $(cat "$t/n1.gr1" 2>&1)"
"$JQ" -e '.verdict == "FAIL"' "$t/n1.gr2" >/dev/null 2>&1 || fail "reports/gate-review-2.json is not a FAIL: $(cat "$t/n1.gr2" 2>&1)"
[ "$(grep -c '^--- plan session$' "$t/n1.problems")" = 1 ] && grep -qF "$NOTE" "$t/n1.problems" ||
  fail "the second plan session was not sent '$NOTE': $(cat "$t/n1.problems")"
"$JQ" -e '.status == "blocked" and .gate_review.verdict == "FAIL" and .gate_review.rounds == 2' "$t/n1.state" >/dev/null 2>&1 ||
  fail "the plan is not saved blocked after two failed rounds: $("$JQ" -c '{status, gate_review}' "$t/n1.state" 2>&1)"

# An invalid verdict: a FAIL, as for claude.
CFG=$GRC; fixture
: >"$t/stub/copilot.gr-invalid"
runfix "$t/i1"
[ "$rc" = 2 ] || fail "a copilot gate review whose verdict is invalid, twice: vloop run exited $rc, want 2: $(tail -8 "$t/i1")"
"$JQ" -e '.verdict == "FAIL" and (.notes | contains("does not follow gate-verdict/v2 — that is a FAIL"))' "$t/i1.gr1" >/dev/null 2>&1 ||
  fail "reports/gate-review-1.json is not the driver's FAIL for an invalid verdict: $(cat "$t/i1.gr1" 2>&1)"

# A FAIL with a finding: kept, sent back once; the second round's PASS runs the plan.
CFG=$GRC; fixture
: >"$t/stub/copilot.gr-fail-once"
runfix "$t/f1"
[ "$rc" = 0 ] || fail "a copilot gate review that fails round 1 and passes round 2: vloop run exited $rc, want 0: $(tail -8 "$t/f1")"
[ "$(task status T1)" = done ] || fail "after a second-round PASS, T1 is $(task status T1), want done"
"$JQ" -e '.verdict == "FAIL" and .notes == "copilot round one" and .tasks[0].findings[0].kind == "wrong-reason"' "$t/f1.gr1" >/dev/null 2>&1 ||
  fail "reports/gate-review-1.json is not the copilot session's FAIL as written: $(cat "$t/f1.gr1" 2>&1)"
"$JQ" -e '.verdict == "PASS"' "$t/f1.gr2" >/dev/null 2>&1 || fail "reports/gate-review-2.json is not the second round's PASS: $(cat "$t/f1.gr2" 2>&1)"
grep -qF 'gate T1 (wrong-reason): the gate passes on a stub' "$t/f1.problems" ||
  fail "the second plan session was not sent the finding 'gate T1 (wrong-reason): the gate passes on a stub': $(cat "$t/f1.problems")"
"$JQ" -e '.gate_review.rounds == 2 and .gate_review.verdict == "PASS"' "$t/f1.state" >/dev/null 2>&1 ||
  fail "the plan's gate_review is $("$JQ" -c .gate_review "$t/f1.state" 2>&1), want rounds 2, verdict PASS"

# A check-gap finding: exit 2, nothing sent back, no plan kept.
CFG=$GRC; fixture
: >"$t/stub/copilot.gr-checkgap"
runfix "$t/g1"
[ "$rc" = 2 ] || fail "a copilot gate review with a check-gap finding: vloop run exited $rc, want 2: $(tail -8 "$t/g1")"
grep -qF 'the gate review found 1 check gap(s)' "$t/g1" || fail "no check-gap line: $(tail -8 "$t/g1")"
[ "$(ev "$t/g1" '^session plan ' | grep -c .)" = 1 ] || fail "a check gap sent the plan back: $(ev "$t/g1" '^session ')"
[ -f "$t/repo/.vloop/state/state.json" ] && fail "a check gap kept the plan file"

# A copilot gate-review session that moves HEAD: the run halts, exit 9.
CFG=$GRC; fixture
: >"$t/stub/copilot.gr-moveref"
runfix "$t/m1"
[ "$rc" = 9 ] || fail "a copilot gate-review session that moved HEAD: vloop run exited $rc, want 9: $(tail -8 "$t/m1")"
grep -qxF 'vloop: REFS MOVED gate-review — the gate review changed git refs; nothing was committed' "$t/m1" ||
  fail "no 'vloop: REFS MOVED gate-review — the gate review changed git refs; nothing was committed': $(tail -8 "$t/m1")"
ev "$t/m1" '^session (work|review|verify) ' >/dev/null && fail "a session ran after the gate review moved HEAD: $(ev "$t/m1" '^session ')"
echo "gate T2: ok"
