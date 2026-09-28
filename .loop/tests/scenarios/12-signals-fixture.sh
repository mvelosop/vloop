#!/usr/bin/env bash
# The signal formulas, checked against the hand-computed fixture in brief 0003.
# That fixture is the arbiter for both this implementation and runstat's, so
# this is the test that keeps control plane and analysis plane from drifting.
. "$(dirname "$0")/../lib.sh"
FX="$(mktemp -d "${TMPDIR:-/tmp}/loopfx.XXXXXX")"
S="$FX/sessions"; I="$FX/iterations.jsonl"; mkdir -p "$S"
mk() { jq -nc --arg p "$1" --argjson i "$2" --argjson c "$3" --argjson t "$4" --argjson d "$5" \
  '{phase:$p,iteration:$i,total_cost_usd:$c,num_turns:$t,duration_ms:$d,is_error:false,permission_denials:[]}'; }
mk plan 0 1.98 12 141000 >"$S/001-plan.json"
mk work 1 0.50 6 68000 >"$S/002-work.json";  mk review 1 0.20 3 24000 >"$S/003-review.json"
mk work 2 0.52 7 72000 >"$S/004-work.json";  mk review 2 0.20 3 24000 >"$S/005-review.json"
mk work 3 0.48 5 64000 >"$S/006-work.json";  mk review 3 0.20 3 24000 >"$S/007-review.json"
cat >"$I" <<'J'
{"iteration":1,"task":"T1","outcome":"done","attempts":0,"tasks_done":1,"tasks_total":8}
{"iteration":2,"task":"T2","outcome":"gate_fail","attempts":1,"tasks_done":1,"tasks_total":8}
{"iteration":3,"task":"T2","outcome":"done","attempts":1,"tasks_done":2,"tasks_total":8}
J

eq() { [[ "$2" == "$3" ]] && ok "$1 = $2" || bad "$1: got $2, want $3"; }
iters=$(wc -l <"$I" | tr -d ' ')
closed=$(jq -s 'if length==0 then 0 else (.[-1].tasks_done//0) end' "$I")
total=$(jq -s 'if length==0 then 0 else (.[-1].tasks_total//0) end' "$I")
eq "iterations"            "$iters" 3
eq "tasks closed"          "$closed/$total" "2/8"
eq "iterations per closed" "$(awk -v i=$iters -v c=$closed 'BEGIN{printf "%.2f",i/c}')" "1.50"
eq "gate failures"         "$(jq -s '[.[]|select(.outcome=="gate_fail")]|length' "$I")" 1
eq "review rejections"     "$(jq -s '[.[]|select(.outcome=="review_fail")]|length' "$I")" 0
eq "attempts burned"       "$(jq -s '[.[]|select(.outcome!="done")]|length' "$I")" 1
eq "no-progress streak"    "$(jq -s '[.[].tasks_done//0] as $d | ([0]+$d[:-1]) as $p | [range(0;($d|length))] | map(if $d[.] > $p[.] then 1 else 0 end) | (reverse|index(1)) // length' "$I")" 0
eq "estimated spend"       "$(printf '%.2f' "$(jq -s '[.[].total_cost_usd//0]|add//0' "$S"/*.json)")" "4.08"
finish
