#!/bin/sh
# Gate T14 — close: the brief's worked example, line for line, with a fresh
# build in temporary repositories and a temporary HOME, and its planted
# failures. "Merged" means the brief's consumed commit is on main (M-5).
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
N=$RUN.loop-brief
sq() { sed -e 's/[[:space:]][[:space:]]*/ /g' -e 's/^ //' -e 's/ $//' "$1"; }
dl() { sq "$1" | grep -E '^defects ' | head -n 1; }
# x <want exit> <out> <args...>: vloop in $R; stdout in $t/<out>, stderr in $t/<out>.err.
x() {
  w=$1 o=$2; shift 2
  inr "$B" "$@" </dev/null >"$t/$o" 2>"$t/$o.err"; c=$?
  [ "$c" = "$w" ] || fail "vloop $*: exit $c, want $w: $(cat "$t/$o.err")"
}
err() { [ "$(cat "$t/$1.err")" = "$2" ] || fail "$3 printed
$(cat "$t/$1.err")
want
$2"; }

# --- run.git-config-ignore
newrepo cfg
x 2 s1 config set run.git-config-ignore 'Core.*'
err s1 'vloop: run.git-config-ignore: Core.* would ignore core.* — the guard must see it' "config set 'Core.*'"
x 2 s2 config set run.git-config-ignore 'core.*'
err s2 'vloop: run.git-config-ignore: core.* would ignore core.* — the guard must see it' "config set 'core.*'"
x 0 s3 config set run.git-config-ignore 'Branch.Main.Foo-Bar'
for key in branch.Main.foo-bar:0 branch.main.foo-bar:9; do
  k=${key%:*} want=${key#*:}
  runrepo "g$want" T1
  x 0 s4 config set run.git-config-ignore 'Branch.Main.Foo-Bar'
  commit pattern
  printf 'git config %s x\n' "$k" >"$t/stub/work-T1.sh"
  vrun "g$want.run"
  rm -f "$t/stub/work-T1.sh"
  [ "$got" = "$want" ] || fail "a session sets $k under Branch.Main.Foo-Bar: the run exited $got, want $want: $(tail -n 10 "$t/g$want.run")"
done

# --- defect add: unmerged, merged, the same second
runrepo d T1 T2
U=B20260101-0700-unmerged.loop-brief
putbrief "${U%.loop-brief}" ready
x 0 o config set metrics.stacks go
commit "unmerged brief and stacks"
cat >"$t/stub/work-T1.sh" <<'W'
mkdir -p src; i=0; : >src/a.go; while [ $i -lt 200 ]; do echo "// $i" >>src/a.go; i=$((i+1)); done
i=0; : >src/a_test.go; while [ $i -lt 300 ]; do echo "// $i" >>src/a_test.go; i=$((i+1)); done
W
echo 0.25 >"$t/stub/cost"
vrun d.run
rm -f "$t/stub/work-T1.sh" "$t/stub/cost"
[ "$got" = 0 ] || fail "vloop run: exit $got: $(tail -n 10 "$t/d.run")"
x 0 o brief close "$BRIEF" --no-findings
mergemain
x 1 o defect add "x" --brief "$U" --found-by user
for s in "\"<summary>\"" '"x"'; do
  [ "$(cat "$t/o.err")" = "vloop: $U is not merged: record what you found before merge as an intervention — vloop intervention add $s --brief $U --kind verification-finding --origin <brief|plan|work|env> --phase verify --automatable partly --by operator" ] && ok=1 && break
  ok=0
done
[ "$ok" = 1 ] || fail "defect add on an unmerged brief printed: $(cat "$t/o.err")"
[ -z "$(ls -A "$R/.vloop/defects" 2>/dev/null)" ] || fail "defect add on an unmerged brief wrote a file"
sleep 2
# Records of another brief hold the next eleven seconds, so "the same second" is certain.
mkdir -p "$R/.vloop/defects"
now=$(date +%s); k=0
while [ $k -le 10 ]; do s=$(stamp $((now + k))); defectdoc "D$s-other" "$U" user 2026-01-01T00:00:00Z >"$R/.vloop/defects/D$s-other.md"; k=$((k + 1)); done
x 0 a1 defect add "x" --brief "$N" --found-by operator
[ "$(cat "$t/a1")" = ".vloop/defects/D$(stamp $((now + 11)))-x.md" ] || fail "defect add printed '$(cat "$t/a1")', want .vloop/defects/D$(stamp $((now + 11)))-x.md"
x 0 a2 defect add "x" --brief "$N" --found-by operator
[ "$(cat "$t/a2")" = ".vloop/defects/D$(stamp $((now + 12)))-x.md" ] || fail "the same again printed '$(cat "$t/a2")', want the stamp one second later"
x 0 m metrics "$BRIEF"
[ "$(dl "$t/m")" = "defects in-loop 0 · operator 0 · escaped 2 · removal efficiency 0%" ] || fail "metrics of the merged brief: '$(dl "$t/m")'"

# --- intervention add --origin
x 2 i1 intervention add "y" --brief "$N" --phase verify --kind verification-finding --automatable no --by operator
err i1 'vloop: --kind verification-finding needs --origin brief, plan, work or env' "verification-finding without --origin"
x 2 i2 intervention add "y" --brief "$N" --phase verify --kind repair --origin work --automatable no --by operator
err i2 'vloop: --origin is only for --kind verification-finding' "--origin with repair"
x 0 i3 intervention add "y" --brief "$N" --phase verify --kind verification-finding --origin plan --automatable no --by operator
p=$(cat "$t/i3")
case "$p" in .vloop/interventions/I[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]-[0-9][0-9][0-9][0-9][0-9][0-9]-y.md) ;; *) fail "intervention add printed '$p'" ;; esac
grep -Fxq 'origin: plan' "$R/$p" || fail "$p has no origin: plan"
x 2 r defect set "$(basename "$(cat "$t/a1")" .md)" rule 'gate: x'
err r 'vloop: cannot set "rule": want one of status, fixed-by, case, severity, origin, kind, task' "defect set rule"

# --- metrics per task: T1 code+ 200, test+ 300, cost 0.50; T2 code+ 0
x 0 bt metrics "$BRIEF" --by task
h=$(head -n 1 "$t/bt")
col() { printf '%s\n' "$h" | tr -s ' ' '\n' | grep -nFx -- "$1" | cut -d: -f1; }
o=$(col other+)
[ "$(col test:code)" = $((o + 1)) ] && [ "$(col '$/kloc')" = $((o + 2)) ] && [ "$(col agent)" = $((o + 3)) ] || fail "per-task header: $h"
r1=$(awk -v n="$o" '$1 == "T1" { print $(n+1), $(n+2) }' "$t/bt"); r2=$(awk -v n="$o" '$1 == "T2" { print $(n+1), $(n+2) }' "$t/bt")
[ "$r1" = "1.50 2.50" ] || fail "T1's test:code and \$/kloc are '$r1', want '1.50 2.50':
$(cat "$t/bt")"
[ "$r2" = "- -" ] || fail "T2's are '$r2', want '- -'"
x 0 sum metrics "$BRIEF"
grep -E '^ *cost ' "$t/sum" | grep -Eq ' · \$[0-9][0-9,]*\.[0-9][0-9]/kloc$' || fail "the cost line does not end ' · \$<c>/kloc': $(grep -E '^ *cost ' "$t/sum")"
x 0 j metrics "$BRIEF" --by task --json
"$JQ" -e '[.by_task[] | select(.id == "T1")][0] | .test_code_ratio == 1.5 and .per_1000_code_lines_usd == 2.5' "$t/j" >/dev/null 2>&1 || fail "--json T1: $("$JQ" -c '.by_task[0]' "$t/j")"
"$JQ" -e '[.by_task[] | select(.id == "T2")][0] | has("test_code_ratio") and .test_code_ratio == null and .per_1000_code_lines_usd == null' "$t/j" >/dev/null 2>&1 || fail "--json T2: $("$JQ" -c '.by_task[1]' "$t/j")"

# --- a brief run to completion, one gate failure, closed with a finding
runrepo c T1
touch "$t/stub/fail-T1"
vrun c.run
[ "$got" = 0 ] || fail "vloop run: exit $got: $(tail -n 10 "$t/c.run")"
x 2 f1 brief close "$BRIEF" --finding "spec gap: z"
err f1 'vloop: a finding starts with whose defect it is: --finding "<brief|plan|work|env>: <summary>"' "close --finding 'spec gap: z'"
x 0 f2 brief close "$BRIEF" --finding "brief: z"
rec=$(grep '^recorded \.vloop/interventions/I[0-9]\{8\}-[0-9]\{6\}-z\.md$' "$t/f2") || fail "close does not print recorded .vloop/interventions/I<stamp>-z.md:
$(cat "$t/f2")"
id=$(basename "${rec#recorded }" .md)
grep -q '^defects of' "$t/f2" && fail "close still prints 'defects of'"
case "$(tail -n 1 "$t/f2")" in "squash-merge with the trailer: "*) ;; *) fail "close's last line: $(tail -n 1 "$t/f2")" ;; esac
[ "$(dl "$t/f2")" = "defects in-loop 1 · operator 1 · escaped 0 · removal efficiency 100%" ] || fail "close's defects line: '$(dl "$t/f2")'"
grep -Fxq -- "- $id — verification-finding, brief, found by operator: z" "$R/$BRIEF" || fail "the run record has no '- $id — verification-finding, brief, found by operator: z'"

# --- real data: every existing record is still listed and shown
nd=$(ls "$top/.vloop/defects" | grep -c '^D.*\.md$') ni=$(ls "$top/.vloop/interventions" | grep -c '^I.*\.md$')
(cd "$top" && "$B" defect list --json) </dev/null >"$t/dl.json" 2>&1 || fail "defect list on this repository: $(cat "$t/dl.json")"
[ "$("$JQ" length "$t/dl.json")" = "$nd" ] || fail "defect list on this repository lists $("$JQ" length "$t/dl.json") of $nd records"
(cd "$top" && "$B" intervention list --json) </dev/null >"$t/il.json" 2>&1 || fail "intervention list on this repository: $(cat "$t/il.json")"
[ "$("$JQ" length "$t/il.json")" = "$ni" ] || fail "intervention list on this repository lists $("$JQ" length "$t/il.json") of $ni records"
for f in "$top"/.vloop/interventions/I*.md; do
  (cd "$top" && "$B" intervention show "$(basename "$f" .md)") </dev/null >/dev/null 2>"$t/is.err" || fail "intervention show $(basename "$f" .md): $(cat "$t/is.err")"
done

# --- planted failures, each in a scratch copy
mkdir -p "$t/plant" && (cd "$top" && tar -cf - internal cmd) | tar -xf - -C "$t/plant" || fail "copy for the plants"
mkdir -p "$t/plant/internal/zzplant" && printf 'package zzplant\n\nimport "os/exec"\n\nfunc Run() error { return exec.Command("git", "status").Run() }\n' >"$t/plant/internal/zzplant/p.go"
astcheck gitcalls "$t/plant" >/dev/null 2>&1 && fail "a git call planted outside the runner is not caught"
rm -rf "$t/plant/internal/zzplant"
awk '{ print } /^func \(it \*Iterator\) iterate\(/ { for (i = 0; i < 101; i++) print "\t// pad" }' "$top/internal/driver/iterate.go" >"$t/plant/internal/driver/iterate.go"
astcheck funclen "$t/plant" >"$t/fl" 2>&1 && fail "iterate padded past 100 lines is not caught"
grep -q 'iterate' "$t/fl" || fail "the padded iterate is not named: $(cat "$t/fl")"
V=$top/cmd/vloop
printf 'package main\n\nvar _ = gateOnlyInTest\n' >"$t/break.go"; printf 'package main\n\nvar gateOnlyInTest = 1\n' >"$t/break_test.go"
"$JQ" -n --arg a "$V/zz_gate_break.go" --arg b "$V/zz_gate_break_test.go" --arg x "$t/break.go" --arg y "$t/break_test.go" '{Replace: {($a): $x, ($b): $y}}' >"$t/overlay.json"
go test -c -overlay "$t/overlay.json" -o "$t/vloop.test" ./cmd/vloop >"$t/c.out" 2>&1 || fail "test binary with the planted break: $(cat "$t/c.out")"
mkdir -p "$t/tmpd"
(cd "$V" && TMPDIR="$t/tmpd" GOFLAGS="-overlay=$t/overlay.json" "$t/vloop.test" -test.run '^$' >"$t/tm.out" 2>"$t/tm.err"); c=$?
[ "$c" = 1 ] && grep -Fq gateOnlyInTest "$t/tm.err" && ! grep -Eq '^panic:|^goroutine [0-9]+ ' "$t/tm.err" "$t/tm.out" ||
  fail "cmd/vloop made not to build: go test exits $c; it must print the build error and exit 1, no panic trace:
$(head -n 15 "$t/tm.err")"
echo "gate T14: ok"
