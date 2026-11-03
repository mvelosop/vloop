#!/bin/sh
# Gate T3 — the llm grader (A4). Judged through the command, built from
# ./tools/agent-eval into a temporary directory, with a fake copilot session
# (it writes .vloop/tmp/verdict.json, FAIL or PASS, and prints
# copilot-session.jsonl), a fake vloop for --vloop, and a fake claude judge that
# records its argv, directory, the number of files there and its stdin, and
# answers PASS (or a set sequence, or fails). The worked example's o1 and o2
# lines in full, with fixture grader d.
set -u
export LC_ALL=C
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
G=.vloop/state/gates/T3
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
rc() { cat "$t/$1.rc"; }
q() { jq -r "$2" "$1" 2>/dev/null || echo ERR; }
row() { awk -v c="$2" '$1 == c { $1 = $1; print }' "$t/$1.out"; }
near() { awk -v a="$1" -v b="$2" 'BEGIN { d = a - b; if (d < 0) d = -d; exit !(a != "" && d < 1e-9) }'; }
reset() { rm -f "$st"/judge.* "$st/mode"; }
# run <name> <case glob> [--judge-model x]: one copilot run of the fixture cases into $t/<name>.
run() { n=$1; c=$2; shift 2
  "$AE" --agent copilot --model m --vloop "$fb/vloop" --timeout 1m --cases "$FX" --case "$c" --out "$t/$n" "$@" >"$t/$n.out" 2>"$t/$n.err"
  echo $? >"$t/$n.rc"; }
grades() { q "$t/$1/report.json" ".cases[] | select(.name == \"$2\") | .runs[0] | [.graders[] | \"\(.name)=\(.score)\"] | sort | join(\" \")"; }

# --- o1 in full: every grader 1, the judge asked three times
reset
run o1 'fixture-review' --judge-model jm
[ "$(rc o1)" = 0 ] || fail "o1 exits $(rc o1), want 0: $(tail -n 5 "$t/o1.err")"
R=$t/o1/report.json
[ "$(grades o1 fixture-review)" = "a=1 b=1 c=1 d=1" ] || fail "o1: graders are not a, b, c, d each 1: $(grades o1 fixture-review)"
[ "$(q "$R" '.cases[0] | [.status, (.mean|tostring), (.runs[0].graders[] | select(.name == "d") | .weight | tostring)] | join(" ")')" = "scored 1 2" ] ||
  fail "o1: not scored with mean 1 and grader d weighted 2: $(q "$R" '.cases[0]')"
[ "$(q "$R" '.judge_model')" = jm ] || fail "o1: report.json's judge_model is not the --judge-model given: $(q "$R" '.judge_model')"
[ "$(cat "$st/judge.n" 2>/dev/null)" = 3 ] || fail "o1: the judge was asked $(cat "$st/judge.n" 2>/dev/null || echo 0) times for one llm grader, want 3"
near "$(q "$R" '.cases[0].runs[0].judge_cost_usd')" 0.03 || fail "o1: judge_cost_usd is not the three calls' 0.03: $(q "$R" '.cases[0].runs[0].judge_cost_usd')"
for k in 1 2 3; do
  a=$st/judge.argv.$k
  [ "$(awk 'p { print; exit } $0 == "--model" { p = 1 }' "$a")" = jm ] || fail "judge call $k: --model is not the --judge-model jm"
  [ "$(awk 'p { print; exit } $0 == "--output-format" { p = 1 }' "$a")" = json ] || fail "judge call $k: no --output-format json"
  [ "$(awk 'p { print; exit } $0 == "--setting-sources" { p = 1 }' "$a")" = project ] || fail "judge call $k: no --setting-sources project"
  grep -qx -- '-p' "$a" || fail "judge call $k: no -p"
  grep -qx -- '--strict-mcp-config' "$a" || fail "judge call $k: no --strict-mcp-config"
  grep -qx -- '--plugin-dir' "$a" && fail "judge call $k: has --plugin-dir"
  [ "$(cat "$st/judge.files.$k")" = 0 ] || fail "judge call $k: its directory was not empty"
  case $(cat "$st/judge.pwd.$k") in "$t/o1"*) fail "judge call $k ran inside --out: $(cat "$st/judge.pwd.$k")" ;; esac
  grep -qF 'Judge-marker-d: the verdict file records that the review failed the task.' "$st/judge.stdin.$k" || fail "judge call $k: its stdin lacks the grader's body"
  grep -qF '{"verdict": "FAIL"}' "$st/judge.stdin.$k" || fail "judge call $k: its stdin lacks the focus file's contents"
  grep -qF 'PASS' "$st/judge.stdin.$k" && grep -qF 'FAIL' "$st/judge.stdin.$k" || fail "judge call $k: its stdin does not ask for PASS or FAIL"
done
[ "$(row o1 fixture-review)" = 'fixture-review review 1 1.00 $0.04' ] || fail "o1's line is not 'fixture-review review 1 1.00 \$0.04' (session and judge): $(row o1 fixture-review)"

# --- o2: the session writes PASS, so b scores 0
reset; echo pass >"$st/mode"
run o2 'fixture-review'
[ "$(rc o2)" = 0 ] || fail "o2 exits $(rc o2), want 0: $(tail -n 5 "$t/o2.err")"
[ "$(grades o2 fixture-review)" = "a=1 b=0 c=1 d=1" ] || fail "o2: graders are not a=1 b=0 c=1 d=1: $(grades o2 fixture-review)"
near "$(q "$t/o2/report.json" '.cases[0].mean')" "$(awk 'BEGIN { printf "%.17g", 4 / 6 }')" || fail "o2: mean is not the unrounded 4/6: $(q "$t/o2/report.json" '.cases[0].mean')"
near "$(q "$t/o2/report.json" '.cases[0].runs[0].score')" "$(awk 'BEGIN { printf "%.17g", 4 / 6 }')" || fail "o2: the run's score is not 4/6"
[ "$(row o2 fixture-review)" = 'fixture-review review 1 0.67 $0.04 b' ] || fail "o2's line is not 'fixture-review review 1 0.67 \$0.04 b': $(row o2 fixture-review)"

# --- the majority decides: FAIL, FAIL, PASS scores 0; an answer that is neither counts as FAIL
reset; printf 'FAIL\nPASS\nMAYBE\n' >"$st/judge.seq"; echo seq >"$st/judge.mode"
run mj 'fixture-review'
[ "$(rc mj)" = 0 ] || fail "a split judge: exits $(rc mj), want 0"
[ "$(cat "$st/judge.n")" = 3 ] || fail "a split judge: asked $(cat "$st/judge.n") times, want 3"
[ "$(grades mj fixture-review)" = "a=1 b=1 c=1 d=0" ] || fail "FAIL, PASS and a non-answer did not score d 0: $(grades mj fixture-review)"
reset; printf 'PASS\nFAIL\nPASS\n' >"$st/judge.seq"; echo seq >"$st/judge.mode"
run mj2 'fixture-review'
[ "$(grades mj2 fixture-review)" = "a=1 b=1 c=1 d=1" ] || fail "PASS, FAIL, PASS did not score d 1: $(grades mj2 fixture-review)"

# --- a judge that fails is a run error
reset; echo fail >"$st/judge.mode"
run jf 'fixture-review'
[ "$(rc jf)" = 1 ] || fail "a failing judge: exits $(rc jf), want 1"
[ "$(q "$t/jf/report.json" '.cases[0] | [.status, .runs[0].status, (.runs[0].score|tostring)] | join(" ")')" = "error error null" ] ||
  fail "a failing judge: the case and run are not errors with score null: $(q "$t/jf/report.json" '.cases[0]')"
case $(q "$t/jf/report.json" '.cases[0].runs[0].reason') in
  "judge failed: d: "?*) ;;
  *) fail "a failing judge: the reason does not begin 'judge failed: d: ': $(q "$t/jf/report.json" '.cases[0].runs[0].reason')" ;;
esac

# --- focus: the last text event (default), the trace text, a missing file
reset
run fo 'fixture-focus'
[ "$(rc fo)" = 0 ] || fail "fixture-focus exits $(rc fo): $(tail -n 5 "$t/fo.err")"
[ "$(grades fo fixture-focus)" = "l=1 m=0 t=1" ] || fail "fixture-focus: graders are not l=1 m=0 t=1: $(grades fo fixture-focus)"
[ "$(q "$t/fo/report.json" '.cases[0].runs[0].graders[] | select(.name == "m") | .reason')" = "no .vloop/tmp/absent.json" ] || fail "an llm focus on a missing file: the reason is not 'no .vloop/tmp/absent.json'"
[ "$(cat "$st/judge.n")" = 6 ] || fail "fixture-focus: the judge was asked $(cat "$st/judge.n") times, want 6 (none for the missing file)"
nl=0; nt=0
for k in 1 2 3 4 5 6; do
  s=$st/judge.stdin.$k
  if grep -qF 'Judge-marker-t:' "$s"; then
    grep -qxF 'Bash {"command":"vloop schema validate verdict/v1 .vloop/tmp/verdict.json"}' "$s" || fail "focus: trace: the judge's stdin lacks the line 'Bash <input>'"
    grep -qxF '.vloop/tmp/verdict.json: ok' "$s" || fail "focus: trace: the judge's stdin lacks the tool's output on its own line"
    grep -qxF 'Read {"path":"notes<1>&2.txt"}' "$s" || fail "focus: trace: the judge's stdin lacks 'Read {\"path\":\"notes<1>&2.txt\"}'"
    grep -qxF 'Verdict written: FAIL.' "$s" || fail "focus: trace: the judge's stdin lacks the text event's line"
    nt=$((nt + 1))
  elif grep -qF 'Judge-marker-l:' "$s"; then
    grep -qF 'Verdict written: FAIL.' "$s" || fail "focus: default: the judge's stdin lacks the last text event"
    grep -qF 'vloop schema validate' "$s" && fail "focus: default: the judge's stdin holds the trace, not the last text event"
    nl=$((nl + 1))
  fi
done
[ "$nt" = 3 ] && [ "$nl" = 3 ] || fail "fixture-focus: want 3 judge calls for t and 3 for l, got $nt and $nl"
echo "gate T3: ok"
