#!/bin/sh
# Gate T2 — a Copilot run end to end, without the judge (A2, A3, A5; A4's regex,
# tool_used and file_exists). Judged through the command, built from
# ./tools/agent-eval into a temporary directory, with a fake copilot (it records
# its argv, directory, plugin dir, hooks file and plan hash, writes
# .vloop/tmp/verdict.json and prints copilot-session.jsonl), a fake vloop for
# --vloop and a fake claude first on PATH, and fixture cases copied from this
# folder. The worked example's o1, o3 and o4 lines, the fixture's grader d left out.
set -u
export LC_ALL=C
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
G=.vloop/state/gates/T2
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
cp "$G/copilot-session.jsonl" "$st/" && cp -R "$G/cases" "$t/cases" || fail "cannot copy the fixtures"
PATH="$fb:$PATH"; export PATH
FX=$t/cases
ae() { n=$1; shift; "$AE" "$@" >"$t/$n.out" 2>"$t/$n.err"; echo $? >"$t/$n.rc"; }
rc() { cat "$t/$1.rc"; }
# q <report> <jq filter>: the filter's raw result, "ERR" when the report does not parse.
q() { jq -r "$2" "$1" 2>/dev/null || echo ERR; }
# row <name> <case>: the table line of a case, its fields joined by one space.
row() { awk -v c="$2" '$1 == c { $1 = $1; print }' "$t/$1.out"; }
total() { grep -F ' case(s) ' "$t/$1.out" | tail -n 1; }
near() { awk -v a="$1" -v b="$2" 'BEGIN { d = a - b; if (d < 0) d = -d; exit !(a != "" && d < 1e-9) }'; }
cop() { "$AE" --agent copilot --model m --vloop "$fb/vloop" --timeout 1m "$@"; }

# --- o1: fixture-review and fixture-graders scored
rm -f "$st/mode"
cop --cases "$FX" --case 'fixture-review' --case 'fixture-g*' --out "$t/o1" >"$t/o1.out" 2>"$t/o1.err"; echo $? >"$t/o1.rc"
[ "$(rc o1)" = 0 ] || fail "o1 exits $(rc o1), want 0: $(tail -n 5 "$t/o1.err")"
R=$t/o1/report.json
[ -f "$R" ] || fail "o1 wrote no report.json"
[ "$(q "$R" '[.agent, .model, .judge_model, (.runs|tostring)] | join(" ")')" = "copilot m sonnet 1" ] ||
  fail "o1's report: agent, model, judge_model, runs are not copilot m sonnet 1: $(q "$R" '[.agent, .model, .judge_model, .runs]')"
[ "$(q "$R" '[.cases[].name] | join(" ")')" = "fixture-graders fixture-review" ] || fail "o1's cases are not fixture-graders, fixture-review in name order: $(q "$R" '[.cases[].name]')"
C='.cases[] | select(.name == "fixture-review")'
[ "$(q "$R" "$C | [.phase, .status, (.mean|tostring), (.runs|length|tostring)] | join(\" \")")" = "review scored 1 1" ] ||
  fail "o1: fixture-review is not phase review, status scored, mean 1, one run: $(q "$R" "$C | del(.runs)")"
RUN="$C | .runs[0]"
[ "$(q "$R" "$RUN | [.status, (.score|tostring)] | join(\" \")")" = "scored 1" ] || fail "o1: fixture-review's run is not scored 1: $(q "$R" "$RUN | del(.graders)")"
[ "$(q "$R" "$RUN | [.graders[] | \"\(.name):\(.weight):\(.score)\"] | sort | join(\" \")")" = "a:1:1 b:2:1 c:1:1" ] ||
  fail "o1: fixture-review's graders are not a, b, c weighted 1, 2, 1 each scoring 1: $(q "$R" "$RUN | .graders")"
[ "$(q "$R" "$RUN | [(.credits|tostring), (.cost_usd|type), (.judge_cost_usd|tostring), (.wall_ms|type)] | join(\" \")")" = "2 number 0 number" ] ||
  fail "o1: fixture-review's run is not credits 2, a number cost_usd, judge_cost_usd 0 and a number wall_ms: $(q "$R" "$RUN | del(.graders)")"

# fixture-graders: regex, tool_used and file_exists as the brief defines them
C='.cases[] | select(.name == "fixture-graders")'
RUN="$C | .runs[0]"
[ "$(q "$R" "$RUN | [.graders[] | \"\(.name)=\(.score)\"] | sort | join(\" \")")" = "e=1 f=0 g=0 h=1 i=1 j=1 k=1 l=1 m=0 n=1 p=0" ] ||
  fail "o1: fixture-graders scores are not e=1 f=0 g=0 h=1 i=1 j=1 k=1 l=1 m=0 n=1 p=0: $(q "$R" "$RUN | [.graders[] | \"\(.name)=\(.score)\"]")"
[ "$(q "$R" "$RUN | .graders[] | select(.name == \"g\") | .reason")" = "no .vloop/tmp/absent.json" ] || fail "o1: grader g's reason is not 'no .vloop/tmp/absent.json'"
[ "$(q "$R" "$RUN | .graders[] | select(.name == \"m\") | .reason")" = "unsupported grader tool_order" ] || fail "o1: grader m's reason is not 'unsupported grader tool_order'"
near "$(q "$R" "$RUN | .score")" "$(awk 'BEGIN { print 7 / 11 }')" || fail "o1: fixture-graders' score is not 7/11: $(q "$R" "$RUN | .score")"

# the run folder: workspace, transcript, trace, session record, stderr
D=$t/o1/fixture-review/run-1
[ -d "$D/repo" ] || fail "o1: no $D/repo"
[ "$(cat "$D/repo/.vloop/tmp/verdict.json" 2>/dev/null)" = '{"verdict": "FAIL"}' ] || fail "o1: the session's verdict file is not in run-1/repo/.vloop/tmp/"
cmp -s "$D/transcript.jsonl" "$st/copilot-session.jsonl" || fail "o1: run-1/transcript.jsonl is not Copilot's JSON Lines as printed"
[ "$(jq -n --slurpfile a "$D/trace.jsonl" --slurpfile b "$G/expected-trace.json" '$a == $b[0]' 2>/dev/null)" = true ] ||
  fail "o1: run-1/trace.jsonl is not the expected events:
$(cat "$D/trace.jsonl" 2>/dev/null)"
ls "$D"/sessions/*.json >/dev/null 2>&1 || fail "o1: no session record under run-1/sessions/"
S=$(ls "$D"/sessions/*.json | head -n 1)
[ "$(q "$S" '[.phase, .agent, .model, (.effort|tostring)] | join(" ")')" = "review copilot m null" ] ||
  fail "o1: the session record is not phase review, agent copilot, model m, effort null: $(q "$S" '[.phase, .agent, .model, .effort]')"
ls "$D"/*.stderr >/dev/null 2>&1 || fail "o1: no session stderr file in run-1/"
[ -e "$D/repo/sessions" ] && fail "o1: the session record went under repo/, not the run folder"

# how the session was started (the last session of o1 is fixture-review's: cases run in name order)
[ "$(cat "$st/copilot.pwd")" = "$D/repo" ] || fail "the session did not run in run-1/repo: $(cat "$st/copilot.pwd")"
awk 'p { print; exit } $0 == "-p" { p = 1 }' "$st/copilot.argv" >"$t/p"
[ "$(cat "$t/p")" = "/vloop:review T1" ] || fail "the session's prompt is not '/vloop:review T1': $(cat "$t/p")"
[ "$(awk 'p { print; exit } $0 == "--model" { p = 1 }' "$st/copilot.argv")" = m ] || fail "the session's --model is not m"
grep -qx -- '--effort' "$st/copilot.argv" && fail "the session was given --effort; no effort is set"
grep -qx -- '--allow-all-tools' "$st/copilot.argv" || fail "the session was not started with the adapter's command (no --allow-all-tools)"
[ "$(cat "$st/copilot.plugin")" = "$D/repo/.vloop/tmp/plugin/agent-eval" ] || fail "--plugin-dir is not run-1/repo/.vloop/tmp/plugin/agent-eval: $(cat "$st/copilot.plugin")"
[ -f "$D/repo/.vloop/tmp/plugin/agent-eval/skills/review/SKILL.md" ] || fail "the extracted plugin has no skills/review/SKILL.md"
[ -f "$st/copilot.hooks" ] || fail "the Copilot home holds no hooks/vloop-fence.json"
case $(q "$st/copilot.hooks" '.hooks.preToolUse[0].bash') in
  "'$fb/vloop' fence hook --phase review --root "*) ;;
  *) fail "the fence hook does not run --vloop's binary: $(q "$st/copilot.hooks" '.hooks.preToolUse[0].bash')" ;;
esac
grep -q "^/vloop:review T1	unset\$" "$st/copilot.log" || fail "a review session without a plan was handed a plan hash: $(cat "$st/copilot.log")"

# the table and the total
[ "$(awk 'NR == 1 { $1 = $1; print }' "$t/o1.out")" = "CASE PHASE RUNS MEAN COST NOTES" ] || fail "o1's table header is not CASE PHASE RUNS MEAN COST NOTES: $(head -n 1 "$t/o1.out")"
[ "$(row o1 fixture-review)" = 'fixture-review review 1 1.00 $0.01' ] || fail "o1's fixture-review line is not 'fixture-review review 1 1.00 \$0.01': $(row o1 fixture-review)"
[ "$(row o1 fixture-graders)" = 'fixture-graders review 1 0.64 $0.01 f, g, m, p' ] || fail "o1's fixture-graders line is not 'fixture-graders review 1 0.64 \$0.01 f, g, m, p': $(row o1 fixture-graders)"
[ "$(total o1)" = '2 case(s) · 0 skipped · 0 error(s) · $0.03' ] || fail "o1's total is not '2 case(s) · 0 skipped · 0 error(s) · \$0.03': $(total o1)"

# --- --runs 2: two runs of one case
cop --cases "$FX" --case 'fixture-review' --runs 2 --out "$t/r2" >"$t/r2.out" 2>"$t/r2.err"; echo $? >"$t/r2.rc"
[ "$(rc r2)" = 0 ] || fail "--runs 2 exits $(rc r2): $(tail -n 5 "$t/r2.err")"
[ "$(q "$t/r2/report.json" '[(.runs|tostring), (.cases[0].runs|length|tostring), (.cases[0].mean|tostring)] | join(" ")')" = "2 2 1" ] ||
  fail "--runs 2: the report does not hold two runs with mean 1"
[ -f "$t/r2/fixture-review/run-2/trace.jsonl" ] && [ -d "$t/r2/fixture-review/run-2/repo" ] || fail "--runs 2: no run-2/ with its repo/ and trace.jsonl"
[ "$(row r2 fixture-review)" = 'fixture-review review 2 1.00 $0.03' ] || fail "--runs 2: the line is not 'fixture-review review 2 1.00 \$0.03': $(row r2 fixture-review)"

# --- the plan hash: a work session gets the scaffold plan's hash, a plan session none
cop --cases "$FX" --case 'fixture-work' --case 'fixture-plan' --out "$t/ph" >"$t/ph.out" 2>"$t/ph.err"; echo $? >"$t/ph.rc"
[ "$(rc ph)" = 0 ] || fail "the work and plan cases exit $(rc ph): $(tail -n 5 "$t/ph.err")"
want=$(shasum -a 256 "$FX/fixture-work/state.json" | awk '{ print $1 }')
grep -q "^/vloop:work T1	$want\$" "$st/copilot.log" || fail "the work session's VLOOP_PLAN_SHA256 is not the scaffold plan's digest $want: $(grep '^/vloop:work' "$st/copilot.log")"
grep -q "^/vloop:plan docs/b.md	unset\$" "$st/copilot.log" || fail "the plan session was handed a plan hash, or its prompt is not '/vloop:plan docs/b.md': $(grep '^/vloop:plan' "$st/copilot.log")"

# --- o3: the scaffold exits 3
cp -R "$FX" "$t/cases3" || fail cp
printf '#!/usr/bin/env bash\necho "first line" >&2\necho "scaffold broke" >&2\nexit 3\n' >"$t/cases3/fixture-review/scaffold.sh"
n0=$(wc -l <"$st/copilot.log")
cop --cases "$t/cases3" --case 'fixture-review' --out "$t/o3" >"$t/o3.out" 2>"$t/o3.err"; echo $? >"$t/o3.rc"
[ "$(rc o3)" = 1 ] || fail "o3 exits $(rc o3), want 1"
[ "$(wc -l <"$st/copilot.log")" = "$n0" ] || fail "o3 started a session after its scaffold failed"
R=$t/o3/report.json
[ "$(q "$R" '.cases[0] | [.status, (.mean|tostring), .runs[0].status, .runs[0].reason, (.runs[0].score|tostring), ((.runs[0].graders // [])|length|tostring)] | join("|")')" = "error|null|error|scaffold exited 3: scaffold broke|null|0" ] ||
  fail "o3: the case is not an error with mean null, its run an error 'scaffold exited 3: scaffold broke', score null, no graders: $(q "$R" '.cases[0]')"
[ "$(row o3 fixture-review)" = 'fixture-review review 1 - $0.00 scaffold exited 3: scaffold broke' ] || fail "o3's line is not 'fixture-review review 1 - \$0.00 scaffold exited 3: scaffold broke': $(row o3 fixture-review)"
[ "$(total o3)" = '1 case(s) · 0 skipped · 1 error(s) · $0.00' ] || fail "o3's total is not '1 case(s) · 0 skipped · 1 error(s) · \$0.00': $(total o3)"

# --- o4: a skill vloop does not start is skipped
n0=$(wc -l <"$st/copilot.log")
cop --cases "$FX" --case 'fixture-operate' --out "$t/o4" >"$t/o4.out" 2>"$t/o4.err"; echo $? >"$t/o4.rc"
[ "$(rc o4)" = 0 ] || fail "o4 exits $(rc o4), want 0"
[ "$(wc -l <"$st/copilot.log")" = "$n0" ] || fail "o4 started a session for a skipped case"
[ "$(q "$t/o4/report.json" '.cases[0] | [.name, .phase, .status, .reason, (.mean|tostring), (.runs|length|tostring)] | join("|")')" = "fixture-operate||skipped|no phase|null|0" ] ||
  fail "o4: the case is not skipped with phase \"\", reason 'no phase', mean null and no runs: $(q "$t/o4/report.json" '.cases[0]')"
[ "$(row o4 fixture-operate)" = 'fixture-operate - 0 - $0.00 no phase' ] || fail "o4's line is not 'fixture-operate - 0 - \$0.00 no phase': $(row o4 fixture-operate)"
[ "$(total o4)" = '1 case(s) · 1 skipped · 0 error(s) · $0.00' ] || fail "o4's total is not '1 case(s) · 1 skipped · 0 error(s) · \$0.00': $(total o4)"

# --- a session that leaves no result; one killed at the timeout
echo silent >"$st/mode"
cop --cases "$FX" --case 'fixture-review' --out "$t/ns" >"$t/ns.out" 2>"$t/ns.err"; echo $? >"$t/ns.rc"
[ "$(rc ns)" = 1 ] || fail "a session with no result: exits $(rc ns), want 1"
[ "$(q "$t/ns/report.json" '.cases[0].runs[0] | [.status, .reason] | join("|")')" = "error|no session result" ] ||
  fail "a session with no result: the run is not an error 'no session result': $(q "$t/ns/report.json" '.cases[0].runs[0]')"
echo sleep >"$st/mode"
"$AE" --agent copilot --model m --vloop "$fb/vloop" --timeout 2s --cases "$FX" --case 'fixture-review' --out "$t/to" >"$t/to.out" 2>"$t/to.err"; echo $? >"$t/to.rc"
[ "$(rc to)" = 1 ] || fail "a session past --timeout 2s: exits $(rc to), want 1"
[ "$(q "$t/to/report.json" '.cases[0].runs[0] | [.status, .reason] | join("|")')" = "error|timed out after 2s" ] ||
  fail "a session past --timeout 2s: the run is not an error 'timed out after 2s': $(q "$t/to/report.json" '.cases[0].runs[0]')"
rm -f "$st/mode"
[ -e "$st/judge.n" ] && fail "a run without llm graders called the judge"
echo "gate T2: ok"
