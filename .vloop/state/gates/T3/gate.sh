#!/bin/sh
# Gate T3 — the metrics half of A3. Over a run directory the base binary wrote
# (session/v1 and session/v2 records, a metrics/v2 snapshot), this tree's
# `vloop metrics`, `metrics --by task`, `metrics --workspace` and
# `brief close --dry-run` keep every line the base prints, and
# `vloop metrics --json` is metrics/v3: the base's document plus `agents`,
# per phase like `models`, all claude. Over a run this tree wrote (session/v3
# records), the same, and the snapshot it commits is metrics/v3. The older
# metrics/v2 schema still validates the base's snapshot.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
newbin
basebin base
S=$t/bin/new
# vin <binary> <out> args...: vloop in the fixture; fails on a non-zero exit.
vin() { b=$1 o=$2; shift 2; (cd "$t/repo" && venv "$b" "$@" >"$o" 2>&1) || fail "vloop $* ($(basename "$b")) failed: $(tail -5 "$o")"; }
# keeps <base out> <new out> <what>: every line of the base's output appears in
# the new output, in the same order (the new output may add lines).
keeps() {
  awk 'NR == FNR { want[++n] = $0; next } i < n && $0 == want[i + 1] { i++ } END { if (i < n) { print want[i + 1]; exit 1 } }' "$1" "$2" >"$t/miss" ||
    fail "$3: the line '$(cat "$t/miss")' the base prints is gone:
$(cat "$2")"
}
# agents <metrics json> <what>: metrics/v3, and agents mirrors models: the same
# phases, ["claude"] where models has any, models' own empty value where not.
agents() {
  [ "$("$JQ" -r .schema "$1")" = metrics/v3 ] || fail "$2: schema $("$JQ" -c .schema "$1"), want metrics/v3"
  "$JQ" -e '(.agents | type) == "object" and (.models | keys) == (.agents | keys) and
    all(.models | keys[]; . as $k | if ($ARGS.named.m[$k] | length) > 0 then $ARGS.named.a[$k] == ["claude"] else $ARGS.named.a[$k] == $ARGS.named.m[$k] end)' \
    --argjson m "$("$JQ" -c .models "$1")" --argjson a "$("$JQ" -c .agents "$1")" "$1" >/dev/null ||
    fail "$2: agents $("$JQ" -c .agents "$1") does not mirror models $("$JQ" -c .models "$1")"
  cp "$1" "$t/m3.json" && "$S" schema validate metrics/v3 "$t/m3.json" >"$t/v.out" 2>&1 || fail "$2: vloop schema validate metrics/v3: $(cat "$t/v.out")"
}

# A run directory the base wrote.
fixture "$t/bin/base"; runfix "$t/bin/base" "$t/base"
[ "$rc" = 0 ] || fail "the base's run exited $rc"
snap=$t/repo/.vloop/state/metrics/B20260101-0900-demo.json
[ "$("$JQ" -r .schema "$snap")" = metrics/v2 ] || fail "the base's run left no metrics/v2 snapshot"
"$S" schema validate metrics/v2 "$snap" >/dev/null 2>&1 || fail "metrics/v2 no longer validates the base's snapshot"
printf '[[repo]]\npath = "%s"\n' "$t/repo" >"$t/ws.toml"
for side in base new; do
  vin "$t/bin/$side" "$t/$side.m" metrics
  vin "$t/bin/$side" "$t/$side.mb" metrics "$BNAME"
  vin "$t/bin/$side" "$t/$side.bt" metrics --by task
  vin "$t/bin/$side" "$t/$side.ws" metrics --workspace "$t/ws.toml"
  vin "$t/bin/$side" "$t/$side.cl" brief close --dry-run --no-findings "$BNAME"
  vin "$t/bin/$side" "$t/$side.j" --json metrics "$BNAME"
  vin "$t/bin/$side" "$t/$side.wj" --json metrics --workspace "$t/ws.toml"
done
keeps "$t/base.m" "$t/new.m" "vloop metrics over a base-written run"
keeps "$t/base.mb" "$t/new.mb" "vloop metrics <brief> over a base-written run"
keeps "$t/base.bt" "$t/new.bt" "vloop metrics --by task over a base-written run"
keeps "$t/base.ws" "$t/new.ws" "vloop metrics --workspace over a base-written run"
keeps "$t/base.cl" "$t/new.cl" "vloop brief close --dry-run over a base-written run"
agents "$t/new.j" "vloop --json metrics over a base-written run"
[ "$("$JQ" -S -c 'del(.schema)' "$t/base.j")" = "$("$JQ" -S -c 'del(.schema, .agents)' "$t/new.j")" ] ||
  fail "vloop --json metrics over a base-written run differs from the base's beyond schema and agents:
base: $("$JQ" -S -c 'del(.schema)' "$t/base.j")
new:  $("$JQ" -S -c 'del(.schema, .agents)' "$t/new.j")"
[ "$("$JQ" -c 'length' "$t/new.wj")" = 1 ] || fail "metrics --workspace --json: $(cat "$t/new.wj")"
"$JQ" '.[0]' "$t/new.wj" >"$t/new.w0" && agents "$t/new.w0" "vloop --json metrics --workspace"
[ "$("$JQ" -S -c '.[0] | del(.schema)' "$t/base.wj")" = "$("$JQ" -S -c '.[0] | del(.schema, .agents)' "$t/new.wj")" ] ||
  fail "vloop --json metrics --workspace differs from the base's beyond schema and agents"
[ "$("$JQ" -c '.agents | to_entries | map(.value) | unique' "$t/new.j")" = '[["claude"]]' ] ||
  fail "over a run of all five phases, agents is $("$JQ" -c .agents "$t/new.j")"

# A run this tree wrote: the same numbers, agents claude, a metrics/v3 snapshot.
fixture "$t/bin/base"; runfix "$S" "$t/run3"
[ "$rc" = 0 ] || fail "this tree's run exited $rc: $(tail -5 "$t/run3")"
[ "$("$JQ" -r .schema "$t/run3.sessions/003-work.json")" = session/v3 ] || fail "this tree's run did not write session/v3 records"
vin "$S" "$t/v3.j" --json metrics "$BNAME"
agents "$t/v3.j" "vloop --json metrics over this tree's run"
[ "$("$JQ" -c .models "$t/v3.j")" = "$("$JQ" -c .models "$t/base.j")" ] || fail "models over this tree's run $("$JQ" -c .models "$t/v3.j"), the base's $("$JQ" -c .models "$t/base.j")"
[ "$("$JQ" -c '.agents | to_entries | map(.value) | unique' "$t/v3.j")" = '[["claude"]]' ] || fail "agents over this tree's run: $("$JQ" -c .agents "$t/v3.j")"
[ "$("$JQ" -r .schema "$snap")" = metrics/v3 ] || fail "this tree's run committed a $("$JQ" -r .schema "$snap") snapshot, want metrics/v3"
cp "$snap" "$t/snap3.json" && "$S" schema validate metrics/v3 "$t/snap3.json" >"$t/v.out" 2>&1 || fail "the new snapshot fails metrics/v3: $(cat "$t/v.out")"
"$JQ" 'del(.agents)' "$t/snap3.json" >"$t/bad.json"
"$S" schema validate metrics/v3 "$t/bad.json" >/dev/null 2>&1 && fail "metrics/v3 accepts a document with no agents"
vin "$S" "$t/v3.m" metrics; [ -s "$t/v3.m" ] || fail "vloop metrics over this tree's run printed nothing"

# A run with no verify session: agents has no verify, as models has none.
fixture "$t/bin/base"; runfix "$S" "$t/nover" VLOOP_RUN_VERIFY=false
[ "$rc" = 0 ] || fail "a run with run.verify false exited $rc"
vin "$S" "$t/nover.j" --json metrics "$BNAME"
agents "$t/nover.j" "vloop --json metrics over a run with no verify session"
echo "gate T3: ok"
