#!/bin/sh
# Gate T2 — A2 and the session half of A3: every session starts through the
# Claude adapter and nothing claude receives changes. A run of this tree and a
# run of the base binary ($BASE, built with git archive) in the same fixture
# path, with the same environment and stub, must hand claude the same argv,
# working directory and environment for plan, gate-review, work, review and
# verify — with and without .vloop/repo-guidelines.md, with and without an
# effort, with and without a plan hash (the plan has none, the later sessions
# one). The comparison is shown to catch a planted extra flag. The one value
# normalized is VLOOP_PLAN_SHA256 (normlog in lib.sh: a digest of a plan that
# carries the driver's timestamps). Records are session/v3 with agent and
# cost_source and otherwise the base's fields; preflight, doctor, stderr
# capture, SESSION RECORD MISSING and exit codes are the base's.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
newbin
basebin base
basebin planted 's/"--strict-mcp-config",/"--strict-mcp-config", "--planted",/'
S=$t/bin/new

# Byte-identity, in the four configurations.
compare plain;                 allphases plain
compare guide guidelines;      allphases guide
compare effort $EFFORTS;       allphases effort
compare both guidelines $EFFORTS; allphases both
grep -q '^arg --append-system-prompt-file$' "$t/plain.new.log" && fail "plain: a guidelines flag without .vloop/repo-guidelines.md"
grep -q '^arg --append-system-prompt-file$' "$t/guide.new.log" || fail "guide: the fixture did not exercise the guidelines file"
grep -q '^arg --effort$' "$t/plain.new.log" && fail "plain: an effort flag with no effort set"
grep -q '^arg --effort$' "$t/effort.new.log" || fail "effort: the fixture did not exercise an effort"
for p in "/vloop:plan $BRIEF" /vloop:gate-review "/vloop:work T1" "/vloop:review T1" /vloop:verify; do
  b=$(sessionblock "$t/plain.new.log" "$p")
  [ -n "$b" ] || fail "no session started with -p '$p'"
  for kv in CLAUDE_CODE_USE_FOUNDRY=1 ANTHROPIC_FOUNDRY_RESOURCE=gate-resource ANTHROPIC_DEFAULT_SONNET_MODEL=gate-sonnet; do
    echo "$b" | grep -qx "env $kv" || fail "'$p': $kv did not reach claude"
  done
  echo "$b" | grep -Eq '^env VLOOP_(ACTIVE_TASK|GATE_TASK|PLAN_SHA256)=leaked$' && fail "'$p': a variable the base drops reached claude"
  echo "$b" | grep -qx "cwd $t/repo" || fail "'$p': claude did not start in the repository root"
done
sessionblock "$t/plain.new.log" "/vloop:plan $BRIEF" | grep -q '^env VLOOP_PLAN_SHA256=' && fail "the plan session got a plan hash"
sessionblock "$t/plain.new.log" "/vloop:work T1" | grep -q '^env VLOOP_PLAN_SHA256=H' || fail "the work session got no plan hash"

# The comparison catches a planted extra flag, and only that.
fixture "$t/bin/base"; runfix "$t/bin/planted" "$t/planted" $CARRY
if diff "$t/planted.log" "$t/plain.new.log" >"$t/planted.diff"; then fail "the comparison did not catch a planted --planted flag"; fi
grep -q '^< arg --planted$' "$t/planted.diff" || fail "the planted flag is not in the comparison's diff"
grep '^[<>]' "$t/planted.diff" | grep -vqx '< arg --planted' && fail "the planted change shows more than the flag: $(head -10 "$t/planted.diff")"

# Records: session/v3, agent and cost_source, and every other field the base's.
nb=$(ls "$t/plain.base.sessions" | tr '\n' ' '); nn=$(ls "$t/plain.new.sessions" | tr '\n' ' ')
[ "$nb" = "$nn" ] || fail "session record files '$nn', the base wrote '$nb'"
[ "$(ls "$t/plain.new.sessions" | wc -l | tr -d ' ')" = 5 ] || fail "a run of five sessions left records '$nn'"
drop='del(.schema, .agent, .cost_source, .started, .ended, .wall_ms)'
for f in $nn; do
  r="$t/plain.new.sessions/$f"
  j=$("$JQ" -c '[.schema, .agent, .cost_source]' "$r") || fail "$f is not JSON"
  [ "$j" = '["session/v3","claude","reported"]' ] || fail "$f: schema, agent, cost_source = $j, want session/v3, claude, reported"
  "$S" schema validate session/v3 "$r" >"$t/v.out" 2>&1 || fail "vloop schema validate session/v3 $f: $(cat "$t/v.out")"
  a=$("$JQ" -S -c "$drop" "$t/plain.base.sessions/$f") b=$("$JQ" -S -c "$drop" "$r")
  [ "$a" = "$b" ] || fail "$f differs from the base's record beyond schema, agent and cost_source:
base: $a
new:  $b"
  for k in started ended wall_ms; do
    [ "$("$JQ" "has(\"$k\")" "$r")" = true ] || fail "$f has no $k"
  done
done
# The schema requires what the record must carry and knows two cost sources.
w="$t/plain.new.sessions/003-work.json"
for e in 'del(.ended)' 'del(.wall_ms)' 'del(.agent)' 'del(.cost_source)' '.cost_source = "estimated"' '.schema = "session/v2"'; do
  "$JQ" "$e" "$w" >"$t/bad.json" || fail jq
  "$S" schema validate session/v3 "$t/bad.json" >/dev/null 2>&1 && fail "session/v3 accepts a record with $e"
done
# The older schemas still validate the base's records.
"$S" schema validate session/v1 "$t/plain.base.sessions/003-work.json" >/dev/null 2>&1 || fail "session/v1 no longer validates a base work record"
"$S" schema validate session/v2 "$t/plain.base.sessions/002-gate-review.json" >/dev/null 2>&1 || fail "session/v2 no longer validates a base gate-review record"

# Captured stderr is the base's.
diff -r "$t/plain.base.stderr" "$t/plain.new.stderr" >"$t/stderr.diff" || fail "captured stderr differs from the base's: $(head -10 "$t/stderr.diff")"

# No cost field: cost_source none, cost_usd 0.
MODES=nocost compare nocost; allphases nocost
for f in $(ls "$t/nocost.new.sessions"); do
  j=$("$JQ" -c '[.schema, .agent, .cost_source, .cost_usd]' "$t/nocost.new.sessions/$f")
  [ "$j" = '["session/v3","claude","none",0]' ] || fail "with no total_cost_usd, $f: schema, agent, cost_source, cost_usd = $j"
  "$S" schema validate session/v3 "$t/nocost.new.sessions/$f" >/dev/null 2>&1 || fail "session/v3 rejects $f with cost_source none"
done

# SESSION RECORD MISSING, exit codes and halts as the base: a silent work
# session, then one whose stdout is not JSON, and a review that exits 3.
for m in silent.work garbage.work "garbage.review exit.review=3" "silent.verify exit.verify=3"; do
  n=$(echo "$m" | tr ' .=' '___')
  MODES=$m compare "$n"
  grep 'SESSION RECORD MISSING' "$t/$n.base.runlog" >"$t/$n.bm" 2>/dev/null; grep 'SESSION RECORD MISSING' "$t/$n.new.runlog" >"$t/$n.nm" 2>/dev/null
  cmp -s "$t/$n.bm" "$t/$n.nm" || fail "$m: SESSION RECORD MISSING lines '$(cat "$t/$n.nm")', the base's '$(cat "$t/$n.bm")'"
  [ "$(ls "$t/$n.base.sessions" | tr '\n' ' ')" = "$(ls "$t/$n.new.sessions" | tr '\n' ' ')" ] || fail "$m: the records written differ from the base's"
done
grep -q 'SESSION RECORD MISSING work' "$t/silent_work.new.runlog" || fail "a silent work session logged no SESSION RECORD MISSING"
grep -q 'SESSION RECORD MISSING work' "$t/garbage_work.new.runlog" || fail "a work session printing no JSON logged no SESSION RECORD MISSING"

# Preflight and doctor as the base, passing and failing.
for m in "" noversion; do
  fixture "$t/bin/base"; [ -n "$m" ] && : >"$t/stub/$m"
  (cd "$t/repo" && venv "$t/bin/base" doctor >"$t/db" 2>&1); a=$?
  (cd "$t/repo" && venv "$S" doctor >"$t/dn" 2>&1); b=$?
  [ "$a" = "$b" ] || fail "doctor ${m:-ok} exited $b, the base $a"
  diff "$t/db" "$t/dn" >"$t/dd" || fail "doctor ${m:-ok} prints other than the base: $(cat "$t/dd")"
done
[ "$a" != 0 ] || fail "the fixture's broken claude did not fail doctor"
MODES=noversion compare preflight
[ "$brc" != 0 ] || fail "the fixture's broken claude did not fail the run's preflight"
grep -q '^arg -p$' "$t/preflight.new.log" && fail "a session started after a failed preflight"
echo "gate T2: ok"
