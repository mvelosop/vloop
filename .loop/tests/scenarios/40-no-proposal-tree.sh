#!/usr/bin/env bash
# A work session that dies before writing .loop/tmp/proposal.json used to
# leave a fixed "work session produced no proposal" / notes "none" -- so a
# session that quietly finished a deliverable and a session that did nothing
# at all read identically. In the consumer's arc that hid a finished
# deliverable for three consecutive iterations and stalled a run. The
# no-proposal path now names every path the session changed or created (the
# driver still commits them under this iteration, unchanged from before),
# and says so in words that differ when the tree is untouched.
. "$(dirname "$0")/../lib.sh"

unset LOOP_ACTIVE_TASK LOOP_GATE_TASK

plan() {   # <files-json> <verify>
  jq -nc --argjson f "$1" --arg v "$2" \
    '{run_id:"fx",brief:"docs/briefs/0003-runstat-cli.md",status:"running",
      iteration:0,created:"2026-08-15T00:00:00Z",updated:"2026-08-15T00:00:00Z",
      tasks:[{id:"T1",title:"t",goal:"g",files:$f,depends_on:[],acceptance:["a"],
              verify:$v,status:"pending",attempts:0,notes:""}]}' > "$FXP"
}

N="" J="" E=""
np() {   # <1|0 -- does the work session touch the tree before dying?>
  fixture_cleanup; fixture_new
  export FXP="$FX/plan.json" EDIT="$1"
  plan '["thing_one.txt"]' 'test -f T1.out'
  printf '%s\n' 'case "$PHASE" in
    plan) cp "$FXP" .loop/state/state.json ;;
    work) [ "$EDIT" = 1 ] && { echo a > thing_one.txt; echo b > thing_two.txt; }; true ;;
    review) jq -nc --arg t "$TASK" "{task:\$t,verdict:\"PASS\",criteria:[],findings:[],notes:\"none\"}" > .loop/tmp/verdict.json ;;
  esac' | fixture_stub
  LOOP_MAX_ITERATIONS=1 fixture_run docs/briefs/0003-runstat-cli.md
  N="$(fx_state '.tasks[0].notes')"
  J="$(cat "$(fx_journal)")"
  E="$(sed -n '/═══/,$p' "$FX/out.log")"
}

note 'the session edited two files and died without a proposal'
np 1; N1="$N"
assert_iter_outcome 1 blocked
assert_status T1 pending
assert_attempts T1 1
for f in thing_one.txt thing_two.txt; do
  case "$N" in *$f*) ok "notes name $f";; *) bad "notes do not name $f: $N";; esac
  case "$J" in *$f*) ok "journal names $f";; *) bad "journal does not name $f";; esac
  case "$E" in *$f*) ok "run-end report names $f";; *) bad "run-end report does not name $f";; esac
  git -C "$FX_REPO" ls-files --error-unmatch "$f" >/dev/null 2>&1 \
    && ok "$f was committed" || bad "$f was not committed"
done
case "$N" in
  *.loop/*) bad "notes report the driver's own files: $N" ;;
  *) ok 'notes leave out driver bookkeeping' ;;
esac

note 'the session did nothing and died without a proposal'
np 0
assert_iter_outcome 1 blocked
[ "$N" != "$N1" ] && ok 'the two cases read differently' \
  || bad "both cases leave the same notes: $N"
case "$N" in
  *thing_*|*.loop/*) bad "nothing changed, but notes name paths: $N" ;;
  *) ok "notes: $N" ;;
esac

finish
