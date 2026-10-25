#!/bin/sh
# Gate T5 — close: every gate of the run at once against the finished tree
# (T2 runs B18's comparison gate unchanged), then the brief's worked example
# line for line, with a one-task plan: the config lines, the default-model
# refusal, doctor's warning, the run with a copilot review and its record
# (the cost computed here from the stub's credits and the default price), the
# missing record, and vloop metrics --json.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
for g in T1 T2 T3 T4; do
  sh "$G/../$g/gate.sh" || { echo "GATE FAIL: gate $g fails on the finished tree" >&2; exit 1; }
done
. "$G/lib.sh"
newbin
S=$t/bin/new
# x <expected exit> <expected stdout> <expected stderr> <env...> -- <args...>
x() {
  ec=$1 eo=$2 ee=$3; shift 3
  e=""; while [ "$1" != -- ]; do e="$e $1"; shift; done; shift
  (cd "$t/repo" && venv $e "$S" "$@" >"$t/o" 2>"$t/e"); c=$?
  [ "$c" = "$ec" ] || fail "${e# } vloop $*: exit $c, want $ec (stdout '$(cat "$t/o")', stderr '$(cat "$t/e")')"
  [ -z "$eo" ] || [ "$(cat "$t/o")" = "$eo" ] || fail "${e# } vloop $*: stdout '$(cat "$t/o")', want '$eo'"
  [ -z "$ee" ] || grep -qxF "$ee" "$t/e" || fail "${e# } vloop $*: stderr '$(cat "$t/e")', want the line '$ee'"
}

# vloop config set agent.review copilot && vloop config set model.review gpt-5.6-terra -> both exit 0
fixture plan1.json
x 0 "" "" -- config set agent.review copilot
x 0 "" "" -- config set model.review "$MODEL"
# vloop config set agent.work copilot -> the refusal, exit 2
x 2 "" 'vloop: invalid value "copilot" for agent.work: want one of claude' -- config set agent.work copilot
[ -z "$(cat "$t/o")" ] || fail "the refusal printed on stdout"
venv git -C "$t/repo" add -A && venv git -C "$t/repo" commit -q -m "copilot reviews" || fail commit

# vloop doctor (agent.review copilot, model set) -> the warning, exit 0
x 0 "" "" -- doctor
grep -qxF '! agent.review is copilot: experimental — review sessions are not fenced until B20 (docs/guide/configuration.md)' "$t/o" ||
  fail "vloop doctor: no line '! agent.review is copilot: experimental — review sessions are not fenced until B20 (docs/guide/configuration.md)': $(cat "$t/o")"

# vloop run <brief> (one task) -> exit 0, the copilot review, its record
(cd "$t/repo" && venv $CARRY "$S" run "$BRIEF" >"$t/run" 2>&1); c=$?
[ "$c" = 0 ] || fail "vloop run exited $c: $(tail -8 "$t/run")"
normlog <"$t/stub/log" >"$t/run.log"
[ "$(sessions "$t/run.log" copilot | tr '\n' ';')" = "/vloop:review T1;" ] || fail "copilot sessions: $(sessions "$t/run.log" copilot | tr '\n' ' ')"
[ "$(sessions "$t/run.log" claude | sed 's/ .*//' | tr '\n' ' ')" = "/vloop:plan /vloop:gate-review /vloop:work /vloop:verify " ] ||
  fail "claude sessions: $(sessions "$t/run.log" claude | tr '\n' ' ')"
b=$(block "$t/run.log" copilot "/vloop:review T1")
[ "$(argafter "$b" -p)" = "/vloop:review T1" ] && [ "$(argafter "$b" --model)" = "$MODEL" ] || fail "the review session is not copilot -p \"/vloop:review T1\" --model $MODEL"
echo "$b" | grep -qx 'arg --disable-builtin-mcps' || fail "the review session has no --disable-builtin-mcps"
echo "$b" | grep -Eq '^probe allowed shell (gh|curl|wget|ssh|scp|nc|rm|git push|git remote|git config --(global|system|file)|git commit|git reset|git checkout|git switch|git stash|git branch|git tag)( |$)' &&
  fail "a C2 form is not denied: $(echo "$b" | grep '^probe allowed shell' | head -3)"
echo "$b" | grep -qx 'probe denied url https://example.invalid/page' || fail "the url tool is not denied"
case "$(echo "$b" | sed -n 's/^env COPILOT_HOME=//p')" in "$t/repo/.vloop/tmp/"?*) ;; *) fail "COPILOT_HOME is not under .vloop/tmp/";; esac
echo "$b" | grep -qx "cwd $t/repo" || fail "the review session did not start in the repository root"
f=$(ls "$t/repo"/.vloop/state/runs/*/*/sessions/*.json | grep '/[0-9][0-9][0-9]-review\.json$')
[ "$(echo "$f" | wc -l | tr -d ' ')" = 1 ] || fail "review records: $f"
cost=$(awk -v a="$CREDITS" -v b="$PRICE" 'BEGIN { printf "%.12g", a * b }')
"$JQ" -e '.agent == "copilot" and .cost_source == "estimated" and .turns == 3 and .duration_ms == 18000 and .permission_denials == []' "$f" >/dev/null ||
  fail "the review record: $("$JQ" -c '{agent, cost_source, turns, duration_ms, permission_denials}' "$f")"
near "$("$JQ" .credits "$f")" "$CREDITS" || fail "the review record's credits $("$JQ" .credits "$f"), want $CREDITS"
near "$("$JQ" .cost_usd "$f")" "$cost" || fail "the review record's cost_usd $("$JQ" .cost_usd "$f"), want $cost ($CREDITS × $PRICE)"
for r in $(ls "$t/repo"/.vloop/state/runs/*/*/sessions/*.json | grep -v '/[0-9][0-9][0-9]-review\.json$'); do
  [ "$("$JQ" -r .agent "$r")" = claude ] || fail "$(basename "$r") was not a claude session"
done

# vloop metrics --json -> credits total 14.2, agents review copilot
(cd "$t/repo" && venv "$S" --json metrics "$BNAME" >"$t/m.json" 2>"$t/m.err") || fail "vloop --json metrics: $(cat "$t/m.err")"
near "$("$JQ" .credits.total "$t/m.json")" "$CREDITS" || fail "metrics credits $("$JQ" -c .credits "$t/m.json"), want total $CREDITS"
[ "$("$JQ" -c .agents.review "$t/m.json")" = '["copilot"]' ] || fail "metrics agents $("$JQ" -c .agents "$t/m.json")"

# The same with a stub that prints no result event -> SESSION RECORD MISSING
CFG=$COPILOT_CFG; fixture plan1.json; : >"$t/stub/copilot.noresult"
(cd "$t/repo" && venv $CARRY "$S" run "$BRIEF" >"$t/nr" 2>&1)
cat "$t/repo"/.vloop/state/runs/*/*/run.log 2>/dev/null | grep -q 'SESSION RECORD MISSING review T1 (iteration 1)' ||
  fail "no result event: no 'SESSION RECORD MISSING review T1 (iteration 1)' in the run log"

# VLOOP_AGENT_REVIEW=copilot vloop run <brief> (model.review from the default) -> refused, exit 1
CFG=; fixture plan1.json
x 1 "" 'vloop: agent.review is copilot but model.review is the default "sonnet", a Claude alias: set model.review to a Copilot model id' VLOOP_AGENT_REVIEW=copilot -- run "$BRIEF"
echo "gate T5: ok"
