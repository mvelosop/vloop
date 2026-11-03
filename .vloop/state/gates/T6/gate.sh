#!/bin/sh
# Gate T6 — the close: the brief's worked example line for line, then every
# other task's gate at once. Judged through the command, built from
# ./tools/agent-eval into a temporary directory, with fake copilot, claude and
# vloop first on PATH and the worked example's fixture cases (fixture-review with
# graders a, b, c, d; fixture-operate) copied from this folder.
set -u
export LC_ALL=C
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
G=.vloop/state/gates/T6
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
cp "$G/copilot-session.jsonl" "$G/claude-stream.jsonl" "$st/" && cp -R "$G/cases" "$t/td" || fail "cannot copy the fixtures"
PATH="$fb:$PATH"; export PATH
TD=$t/td
rc() { cat "$t/$1.rc"; }
q() { jq -r "$2" "$1" 2>/dev/null || echo ERR; }
row() { awk -v c="$2" '$1 == c { $1 = $1; print }' "$t/$1.out"; }
near() { awk -v a="$1" -v b="$2" 'BEGIN { d = a - b; if (d < 0) d = -d; exit !(a != "" && d < 1e-9) }'; }
ae() { n=$1; shift; "$AE" "$@" >"$t/$n.out" 2>"$t/$n.err"; echo $? >"$t/$n.rc"; }
grades() { q "$t/$1/report.json" '.cases[0].runs[0] | [.graders[] | "\(.name)=\(.score)"] | sort | join(" ")'; }

# agent-eval --agent copilot --model m --cases <td> --case 'fixture-review' --out <tmp>/o1
ae o1 --agent copilot --model m --vloop "$fb/vloop" --cases "$TD" --case 'fixture-review' --out "$t/o1"
[ "$(rc o1)" = 0 ] || fail "o1: exit $(rc o1), want 0: $(tail -n 5 "$t/o1.err")"
[ "$(jq -r 'select(.kind == "tool" and .tool == "Bash" and (.input | contains("vloop schema validate"))) | .tool' "$t/o1/fixture-review/run-1/trace.jsonl" 2>/dev/null | head -n 1)" = Bash ] ||
  fail "o1: trace.jsonl has no tool event Bash whose input contains 'vloop schema validate'"
[ "$(q "$t/o1/report.json" '.cases[0] | [.status, (.mean|tostring)] | join(" ")')" = "scored 1" ] || fail "o1: report.json is not status scored, mean 1"
[ "$(grades o1)" = "a=1 b=1 c=1 d=1" ] || fail "o1: graders are not a, b, c, d each 1: $(grades o1)"

# same, the fake copilot writing "verdict": "PASS", --out <tmp>/o2
echo pass >"$st/mode"
ae o2 --agent copilot --model m --vloop "$fb/vloop" --cases "$TD" --case 'fixture-review' --out "$t/o2"
rm -f "$st/mode"
[ "$(rc o2)" = 0 ] || fail "o2: exit $(rc o2), want 0"
[ "$(grades o2)" = "a=1 b=0 c=1 d=1" ] || fail "o2: graders are not b 0 and the others 1: $(grades o2)"
near "$(q "$t/o2/report.json" '.cases[0].mean')" "$(awk 'BEGIN { printf "%.17g", 4 / 6 }')" || fail "o2: mean is not 0.6666…"
[ "$(row o2 fixture-review | awk '{ print $4, $NF }')" = "0.67 b" ] || fail "o2: the table's MEAN is not 0.67 with NOTES b: $(row o2 fixture-review)"

# same, the fixture's scaffold exiting 3, --out <tmp>/o3
cp -R "$TD" "$t/td3" || fail cp
printf '#!/usr/bin/env bash\necho "cannot build" >&2\nexit 3\n' >"$t/td3/fixture-review/scaffold.sh"
ae o3 --agent copilot --model m --vloop "$fb/vloop" --cases "$t/td3" --case 'fixture-review' --out "$t/o3"
[ "$(rc o3)" = 1 ] || fail "o3: exit $(rc o3), want 1"
case $(q "$t/o3/report.json" '.cases[0].runs[0] | "\(.status)|\(.reason)"') in
  "error|scaffold exited 3:"*) ;;
  *) fail "o3: the run is not an error whose reason begins 'scaffold exited 3:': $(q "$t/o3/report.json" '.cases[0].runs[0]')" ;;
esac

# agent-eval --agent copilot --model m --cases <td> --case 'fixture-operate' --out <tmp>/o4
ae o4 --agent copilot --model m --vloop "$fb/vloop" --cases "$TD" --case 'fixture-operate' --out "$t/o4"
[ "$(rc o4)" = 0 ] || fail "o4: exit $(rc o4), want 0"
[ "$(q "$t/o4/report.json" '.cases[0] | "\(.status)|\(.reason)"')" = "skipped|no phase" ] || fail "o4: status is not skipped with reason 'no phase'"

# agent-eval --list --cases <td>
ae l1 --list --cases "$TD"
[ "$(rc l1)" = 0 ] || fail "--list --cases <td>: exit $(rc l1)"
[ "$(awk 'NF && NR > 1 { print $1, $2, $3 }' "$t/l1.out" | tr '\n' ';')" = "fixture-operate - 0;fixture-review review 4;" ] ||
  fail "--list --cases <td> is not fixture-operate - 0, fixture-review review 4: $(cat "$t/l1.out")"

# agent-eval --list
ae l2 --list
[ "$(rc l2)" = 0 ] || fail "--list: exit $(rc l2)"
[ -s "$t/l2.err" ] && fail "--list: an error on stderr: $(cat "$t/l2.err")"
n=0; for d in plugin/evals/*/; do case $(basename "$d") in _*) ;; *) n=$((n + 1)) ;; esac; done
[ "$(awk 'NF' "$t/l2.out" | wc -l | tr -d ' ')" = $((n + 1)) ] || fail "--list: not a header and one line per case of plugin/evals ($n)"

# agent-eval --agent nope --model m --out <tmp>/o5
ae o5 --agent nope --model m --out "$t/o5"
[ "$(rc o5)" = 2 ] && [ "$(cat "$t/o5.err")" = "agent-eval: unknown agent nope" ] || fail "--agent nope: not exit 2 with 'agent-eval: unknown agent nope': $(rc o5) $(cat "$t/o5.err")"

# agent-eval --agent claude --model m --cases <td> --out <tmp>/o1   (not empty)
ae ne --agent claude --model m --cases "$TD" --out "$t/o1"
[ "$(rc ne)" = 2 ] && [ "$(cat "$t/ne.err")" = "agent-eval: $t/o1 is not empty" ] || fail "an --out not empty: not exit 2 with 'agent-eval: $t/o1 is not empty': $(rc ne) $(cat "$t/ne.err")"

# agent-eval --agent claude --model m --cases <td> --case 'fixture-review' --out <tmp>/o6
ae o6 --agent claude --model m --cases "$TD" --case 'fixture-review' --out "$t/o6"
[ "$(rc o6)" = 0 ] || fail "o6: exit $(rc o6), want 0: $(tail -n 5 "$t/o6.err")"
[ "$(jq -n --slurpfile a "$t/o6/fixture-review/run-1/trace.jsonl" --slurpfile b "$t/o1/fixture-review/run-1/trace.jsonl" '$a == $b' 2>/dev/null)" = true ] ||
  fail "o6: the trace events are not o1's"
[ "$(q "$t/o6/report.json" '.cases[0].mean | tostring')" = 1 ] || fail "o6: mean is not 1"

# --- every other gate at once
for k in 1 2 3 4 5; do
  sh ".vloop/state/gates/T$k/gate.sh" >"$t/gate$k.out" 2>&1 || fail "gate T$k fails: $(tail -n 3 "$t/gate$k.out")"
done
echo "gate T6: ok"
