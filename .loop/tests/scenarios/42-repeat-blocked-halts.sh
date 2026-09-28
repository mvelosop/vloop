#!/usr/bin/env bash
# A memoryless session given identical inputs reaches an identical
# conclusion. When a task blocks twice in a row and nothing outside the
# driver's own .loop/state/ bookkeeping changed between the two attempts, a
# third attempt cannot produce new information -- halt and surface the FIRST
# diagnosis rather than the second. When the first blocked session changed
# something (a path in the task's files, say), the retry is allowed and the
# rule is re-evaluated on the next pair.
. "$(dirname "$0")/../lib.sh"

PLAN='{"run_id":"fixture","brief":"docs/briefs/0003-runstat-cli.md","status":"running","iteration":0,
 "created":"2026-08-15T00:00:00Z","updated":"2026-08-15T00:00:00Z","tasks":[
 {"id":"T1","title":"Only","goal":"g","files":["thing.txt"],"depends_on":[],
  "acceptance":["T1.out exists"],"verify":"test -f T1.out","status":"pending","attempts":0,"notes":""}]}'

rb() {   # <1|0 -- does the first work session change thing.txt (a file this task owns)?>
  fixture_cleanup; fixture_new
  export FXN="$FX/n" CHG="$1"
  fixture_stub <<STUB
case "\$PHASE" in
  plan) cat > .loop/state/state.json <<'PLANJSON'
$PLAN
PLANJSON
    ;;
  work)
    n=\$(( \$(cat "\$FXN" 2>/dev/null || echo 0) + 1 ))
    echo \$n > "\$FXN"
    [ "\$CHG" = 1 ] && [ \$n = 1 ] && echo changed > thing.txt
    jq -nc --arg t "\$TASK" --arg s "diagnosis \$n" \
      '{task:\$t,outcome:"blocked",summary:\$s,files:[],verified:"none",notes:\$s}' > .loop/tmp/proposal.json ;;
  review) jq -nc --arg t "\$TASK" '{task:\$t,verdict:"PASS",criteria:[],findings:[],notes:"none"}' > .loop/tmp/verdict.json ;;
esac
STUB
  LOOP_STALL_LIMIT=9 LOOP_MAX_ATTEMPTS=5 LOOP_CONVERGENCE_MIN=99 fixture_run docs/briefs/0003-runstat-cli.md
  end_log="$(sed -n '/═══/,$p' "$FX/out.log")"
}

note "-- blocked twice on identical inputs: halts after the second, first diagnosis --"
rb 0
assert_iterations 2
assert_status T1 pending
assert_attempts T1 2
[ "$EXIT" != 0 ] && ok "halted non-zero ($EXIT)" || bad "exited 0 with a task unfinished"
case "$end_log" in
  *"diagnosis 1"*) ok "the halt surfaces the first diagnosis" ;;
  *) bad "the halt does not surface diagnosis 1: $end_log" ;;
esac
case "$end_log" in
  *"diagnosis 2"*) bad "the halt surfaces the second diagnosis instead of the first" ;;
  *) ok "the second diagnosis is not what gets surfaced" ;;
esac

note "-- first attempt changes something: retry is allowed, re-evaluated on the next pair --"
rb 1
assert_iterations 3
assert_status T1 pending
assert_attempts T1 3
case "$end_log" in
  *"diagnosis 2"*) ok "the halt surfaces the first diagnosis of the unchanged pair (iteration 2)" ;;
  *) bad "the halt does not surface diagnosis 2: $end_log" ;;
esac

finish
