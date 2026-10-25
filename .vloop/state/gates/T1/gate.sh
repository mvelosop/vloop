#!/bin/sh
# Gate T1 — A1: the five agent.<phase> keys. Default claude, the only valid
# value; VLOOP_AGENT_<PHASE> in the environment; stored under [agent]; an
# invalid value refused by `config set` (exit 2, nothing written) and a source
# error (exit 1) for config get, config list, doctor and run; listed with value
# and source; a row each in docs/guide/configuration.md. The worked example's
# config lines, in a repository initialised by this tree's binary. And with
# every agent key set to claude, a run hands claude what the base binary's run
# hands it (lib.sh's compare).
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
newbin
basebin base
S=$t/bin/new
# x <expected exit> <expected stdout> <expected stderr> <env...> -- <args...>
x() {
  ec=$1 eo=$2 ee=$3; shift 3
  e=""; while [ "$1" != -- ]; do e="$e $1"; shift; done; shift
  (cd "$t/repo" && venv $e "$S" "$@" >"$t/o" 2>"$t/e"); c=$?
  [ "$c" = "$ec" ] || fail "${e# } vloop $*: exit $c, want $ec (stdout '$(cat "$t/o")', stderr '$(cat "$t/e")')"
  [ "$(cat "$t/o")" = "$eo" ] || fail "${e# } vloop $*: stdout '$(cat "$t/o")', want '$eo'"
  [ "$(cat "$t/e")" = "$ee" ] || fail "${e# } vloop $*: stderr '$(cat "$t/e")', want '$ee'"
}
up() { echo "$1" | tr 'a-z-' 'A-Z_'; }

fixture "$S"
# The worked example, line for line.
x 0 claude "" -- config get agent.review
x 0 '{"key":"agent.review","value":"claude","source":"default"}' "" -- config get agent.review --json
cp "$t/repo/.vloop/config.toml" "$t/cfg.before"
x 2 "" 'vloop: invalid value "copilot" for agent.review: want one of claude' -- config set agent.review copilot
cmp -s "$t/cfg.before" "$t/repo/.vloop/config.toml" || fail "a refused config set changed .vloop/config.toml"
x 1 "" 'vloop: VLOOP_AGENT_WORK: invalid value "copilot" for agent.work: want one of claude' VLOOP_AGENT_WORK=copilot -- config list
x 0 "" "" -- config set agent.review claude
x 0 '{"key":"agent.review","value":"claude","source":"file"}' "" -- config get agent.review --json
grep -Eq '^[[:space:]]*\[agent\][[:space:]]*$' "$t/repo/.vloop/config.toml" || fail "config set agent.review did not store it under [agent]: $(cat "$t/repo/.vloop/config.toml")"

# All five keys: default, environment, list, refusal, source errors.
fixture "$S"
envsrc=$(cd "$t/repo" && venv VLOOP_MODEL_REVIEW=x "$S" config get model.review --json | "$JQ" -r .source) || fail "model.review --json"
(cd "$t/repo" && venv "$S" config list >"$t/list" 2>&1) || fail "config list: $(cat "$t/list")"
for p in $PHASES; do
  k=agent.$p v=VLOOP_AGENT_$(up "$p")
  x 0 claude "" -- config get "$k"
  x 0 "{\"key\":\"$k\",\"value\":\"claude\",\"source\":\"default\"}" "" -- config get "$k" --json
  x 0 "{\"key\":\"$k\",\"value\":\"claude\",\"source\":\"$envsrc\"}" "" "$v=claude" -- config get "$k" --json
  grep -qx "$k=claude (default)" "$t/list" || fail "config list has no line '$k=claude (default)': $(cat "$t/list")"
  x 2 "" "vloop: invalid value \"copilot\" for $k: want one of claude" -- config set "$k" copilot
  x 1 "" "vloop: $v: invalid value \"copilot\" for $k: want one of claude" "$v=copilot" -- config list
  x 1 "" "vloop: $v: invalid value \"copilot\" for $k: want one of claude" "$v=copilot" -- config get "$k"
  (cd "$t/repo" && venv "$v=copilot" "$S" doctor >"$t/o" 2>&1); c=$?
  [ "$c" = 1 ] || fail "$v=copilot vloop doctor exited $c, want 1: $(cat "$t/o")"
  grep -q "$v" "$t/o" || fail "$v=copilot vloop doctor does not name the variable: $(cat "$t/o")"
  (cd "$t/repo" && venv "$v=copilot" "$S" run "$BRIEF" >"$t/o" 2>&1); c=$?
  [ "$c" = 1 ] || fail "$v=copilot vloop run exited $c, want 1: $(tail -5 "$t/o")"
  [ -f "$t/stub/log" ] && grep -q '^arg -p$' "$t/stub/log" && fail "$v=copilot vloop run started a session"
  rm -f "$t/stub/log"
done
# An invalid value in the file is a source error too, and a valid one reads as file.
cp "$t/repo/.vloop/config.toml" "$t/cfg.base"
for p in $PHASES; do
  k=agent.$p
  { cat "$t/cfg.base"; printf '\n[agent]\n  %s = "copilot"\n' "$p"; } >"$t/repo/.vloop/config.toml"
  for a in "config get $k" "config list" doctor "run $BRIEF"; do
    (cd "$t/repo" && venv "$S" $a >"$t/o" 2>&1); c=$?
    [ "$c" = 1 ] || fail "with $k = \"copilot\" in the file, vloop $a exited $c, want 1: $(tail -5 "$t/o")"
  done
  grep -q 'invalid value "copilot" for '"$k"': want one of claude' "$t/o" || fail "with $k = \"copilot\" in the file, vloop run does not say why: $(tail -5 "$t/o")"
  { cat "$t/cfg.base"; printf '\n[agent]\n  %s = "claude"\n' "$p"; } >"$t/repo/.vloop/config.toml"
  x 0 "{\"key\":\"$k\",\"value\":\"claude\",\"source\":\"file\"}" "" -- config get "$k" --json
done
[ -f "$t/stub/log" ] && grep -q '^arg -p$' "$t/stub/log" && fail "a run with an invalid agent in the file started a session"

# The configuration guide has a row for each key: default claude, its variable.
for p in $PHASES; do
  row=$(grep -E "^\|[[:space:]]*\`agent\.$p\`[[:space:]]*\|" "$top/docs/guide/configuration.md") || fail "docs/guide/configuration.md has no table row for agent.$p"
  echo "$row" | awk -F'|' -v v="\`VLOOP_AGENT_$(up "$p")\`" '{ for (i = 2; i < NF; i++) { gsub(/^[ \t]+|[ \t]+$/, "", $i) } exit !($3 == "`claude`" && $5 == v) }' ||
    fail "the agent.$p row does not give default \`claude\` and variable \`VLOOP_AGENT_$(up "$p")\` in the table's columns: $row"
done

# Every agent set to claude in the environment: sessions as the base's.
compare agents VLOOP_AGENT_PLAN=claude VLOOP_AGENT_WORK=claude VLOOP_AGENT_REVIEW=claude VLOOP_AGENT_GATE_REVIEW=claude VLOOP_AGENT_VERIFY=claude
allphases agents
echo "gate T1: ok"
