#!/bin/sh
# Gate T4 — Claude streaming (A2, A3). Judged through the command, built from
# ./tools/agent-eval into a temporary directory, with a fake claude first on
# PATH: a session (its argv has --plugin-dir) records its argv, plugin dir and
# fence file, writes .vloop/tmp/verdict.json and streams claude-stream.jsonl; a
# judge call (no --plugin-dir) answers PASS. The worked example's o6 line: the
# same trace events as the Copilot run o1, and mean 1.
set -u
export LC_ALL=C
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
G=.vloop/state/gates/T4
[ -d "$G/cases" ] || fail "run from the repository root"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") || fail mktemp
t=$(cd "$t" && pwd -P)
trap 'rm -rf "$t"' EXIT
go build -o "$t/agent-eval" ./tools/agent-eval >"$t/build.out" 2>&1 ||
  fail "go build ./tools/agent-eval fails: $(tail -n 5 "$t/build.out")"
AE=$t/agent-eval
fb=$t/bin; st=$t/st; mkdir "$fb" "$st" || fail mkdir
for f in copilot claude vloop; do
  sed "s|@ST@|$st|g" "$G/fake-$f.sh" >"$fb/$f" && chmod +x "$fb/$f" || fail "cannot install the fake $f"
done
cp "$G/copilot-session.jsonl" "$G/claude-stream.jsonl" "$st/" && cp -R "$G/cases" "$t/cases" || fail "cannot copy the fixtures"
PATH="$fb:$PATH"; export PATH
FX=$t/cases
rc() { cat "$t/$1.rc"; }
q() { jq -r "$2" "$1" 2>/dev/null || echo ERR; }
row() { awk -v c="$2" '$1 == c { $1 = $1; print }' "$t/$1.out"; }
near() { awk -v a="$1" -v b="$2" 'BEGIN { d = a - b; if (d < 0) d = -d; exit !(a != "" && d < 1e-9) }'; }
arg() { awk -v k="$2" 'p { print; exit } $0 == k { p = 1 }' "$1"; }

# --- o6: a Claude session, streamed
"$AE" --agent claude --model m --timeout 1m --cases "$FX" --case 'fixture-review' --out "$t/o6" >"$t/o6.out" 2>"$t/o6.err"; echo $? >"$t/o6.rc"
[ "$(rc o6)" = 0 ] || fail "o6 exits $(rc o6), want 0: $(tail -n 5 "$t/o6.err")"
[ "$(cat "$st/claude.calls" 2>/dev/null | wc -l | tr -d ' ')" = 1 ] || fail "o6 started $(cat "$st/claude.calls" 2>/dev/null | wc -l | tr -d ' ') Claude sessions, want 1"
D=$t/o6/fixture-review/run-1
R=$t/o6/report.json
[ "$(q "$R" '.agent')" = claude ] || fail "o6: report.json's agent is not claude"
[ "$(q "$R" '.cases[0] | [.phase, .status, (.mean|tostring)] | join(" ")')" = "review scored 1" ] || fail "o6: fixture-review is not phase review, scored, mean 1: $(q "$R" '.cases[0] | del(.runs)')"
[ "$(q "$R" '.cases[0].runs[0] | [.graders[] | "\(.name)=\(.score)"] | sort | join(" ")')" = "a=1 b=1 c=1 d=1" ] || fail "o6: graders are not a, b, c, d each 1: $(q "$R" '.cases[0].runs[0].graders')"
[ "$(q "$R" '.cases[0].runs[0].credits | tostring')" = null ] || fail "o6: credits is not null on Claude"
near "$(q "$R" '.cases[0].runs[0].cost_usd')" 0.05 || fail "o6: cost_usd is not the result line's total_cost_usd 0.05: $(q "$R" '.cases[0].runs[0].cost_usd')"
[ "$(jq -n --slurpfile a "$D/trace.jsonl" --slurpfile b "$G/expected-trace.json" '$a == $b[0]' 2>/dev/null)" = true ] ||
  fail "o6: run-1/trace.jsonl is not the expected events:
$(cat "$D/trace.jsonl" 2>/dev/null)"
cmp -s "$D/transcript.jsonl" "$st/claude-stream.jsonl" || fail "o6: run-1/transcript.jsonl is not the stream as printed"
ls "$D"/sessions/*.json >/dev/null 2>&1 || fail "o6: no session record under run-1/sessions/"
S=$(ls "$D"/sessions/*.json | head -n 1)
[ "$(q "$S" '[.phase, .agent, .model, (.effort|tostring), (.turns|tostring), (.duration_ms|tostring)] | join(" ")')" = "review claude m null 5 1200" ] ||
  fail "o6: the session record is not read from the stream's result line (phase review, agent claude, model m, effort null, turns 5, duration_ms 1200): $(q "$S" '[.phase, .agent, .model, .effort, .turns, .duration_ms]')"
near "$(q "$S" '.cost_usd')" 0.05 || fail "o6: the session record's cost_usd is not 0.05"

# how the session was started: the phase's argv, stream-json in place of json
A=$st/claude.argv
[ "$(arg "$A" -p)" = "/vloop:review T1" ] || fail "o6: the prompt is not '/vloop:review T1': $(arg "$A" -p)"
[ "$(arg "$A" --model)" = m ] || fail "o6: --model is not m"
[ "$(arg "$A" --output-format)" = stream-json ] || fail "o6: --output-format is not stream-json: $(arg "$A" --output-format)"
grep -qx -- '--verbose' "$A" || fail "o6: no --verbose"
[ "$(grep -cx -- '--output-format' "$A")" = 1 ] || fail "o6: --output-format given more than once"
grep -qx -- '--effort' "$A" && fail "o6: the session was given --effort; no effort is set"
[ "$(arg "$A" --permission-mode)" = auto ] || fail "o6: not started as vloop starts the phase (no --permission-mode auto)"
grep -qx -- '--strict-mcp-config' "$A" || fail "o6: not started as vloop starts the phase (no --strict-mcp-config)"
[ "$(cat "$st/claude.plugin")" = "$D/repo/.vloop/tmp/plugin/agent-eval" ] || fail "o6: --plugin-dir is not run-1/repo/.vloop/tmp/plugin/agent-eval: $(cat "$st/claude.plugin")"
[ "$(cat "$st/claude.fence" 2>/dev/null)" = "$D/repo/.vloop/tmp/fence/agent-eval/review.json" ] || fail "o6: --settings is not the review fence at run-1/repo/.vloop/tmp/fence/agent-eval/review.json: $(cat "$st/claude.fence" 2>/dev/null)"
cmp -s "$D/repo/.vloop/tmp/fence/agent-eval/review.json" fence/review.json || fail "o6: the extracted fence is not the review phase's fence"
[ "$(cat "$st/judge.n" 2>/dev/null)" = 3 ] || fail "o6: the judge was asked $(cat "$st/judge.n" 2>/dev/null || echo 0) times, want 3"
[ "$(row o6 fixture-review)" = 'fixture-review review 1 1.00 $0.08' ] || fail "o6's line is not 'fixture-review review 1 1.00 \$0.08' (0.05 session + 0.03 judge): $(row o6 fixture-review)"

# --- the same session on Copilot gives the same trace events
"$AE" --agent copilot --model m --vloop "$fb/vloop" --timeout 1m --cases "$FX" --case 'fixture-review' --out "$t/o1" >"$t/o1.out" 2>"$t/o1.err"; echo $? >"$t/o1.rc"
[ "$(rc o1)" = 0 ] || fail "o1 (copilot) exits $(rc o1): $(tail -n 5 "$t/o1.err")"
[ "$(jq -n --slurpfile a "$D/trace.jsonl" --slurpfile b "$t/o1/fixture-review/run-1/trace.jsonl" '$a == $b' 2>/dev/null)" = true ] ||
  fail "o6's trace events differ from the Copilot run's"

# --- a stream without its result line leaves no session result
echo noresult >"$st/mode"
"$AE" --agent claude --model m --timeout 1m --cases "$FX" --case 'fixture-review' --out "$t/nr" >"$t/nr.out" 2>"$t/nr.err"; echo $? >"$t/nr.rc"
rm -f "$st/mode"
[ "$(rc nr)" = 1 ] || fail "a stream with no result line: exits $(rc nr), want 1"
[ "$(q "$t/nr/report.json" '.cases[0].runs[0] | [.status, .reason] | join("|")')" = "error|no session result" ] ||
  fail "a stream with no result line: the run is not an error 'no session result': $(q "$t/nr/report.json" '.cases[0].runs[0]')"
echo "gate T4: ok"
