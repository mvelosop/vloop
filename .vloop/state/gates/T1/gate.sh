#!/bin/sh
# Gate T1 — the configuration. agent.review accepts claude and copilot; every
# other agent.<phase> claude only (config set exits 2 with the brief's
# message and writes nothing; the environment and the file are source errors,
# exit 1). copilot.credit-usd: default 0.0066667, a positive number,
# VLOOP_COPILOT_CREDIT_USD, set and read back. The worked example's config
# lines, in a repository initialised by this tree's binary.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
newbin
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
# refused <exit> <env...> -- <args...>: exits as given with nothing on stdout
# and a vloop: message on stderr; the message is in $t/e.
refused() {
  ec=$1; shift
  e=""; while [ "$1" != -- ]; do e="$e $1"; shift; done; shift
  (cd "$t/repo" && venv $e "$S" "$@" >"$t/o" 2>"$t/e"); c=$?
  [ "$c" = "$ec" ] || fail "${e# } vloop $*: exit $c, want $ec (stdout '$(cat "$t/o")', stderr '$(cat "$t/e")')"
  grep -q '^vloop: ' "$t/e" || fail "${e# } vloop $*: no vloop: message on stderr: '$(cat "$t/e")'"
}
up() { echo "$1" | tr 'a-z.-' 'A-Z__'; }

fixture
cp "$t/repo/.vloop/config.toml" "$t/cfg.before"
# The worked example, line for line.
x 0 "" "" -- config set agent.review copilot
x 0 "" "" -- config set model.review "$MODEL"
x 2 "" 'vloop: invalid value "copilot" for agent.work: want one of claude' -- config set agent.work copilot
x 0 copilot "" -- config get agent.review
x 0 '{"key":"agent.review","value":"copilot","source":"file"}' "" -- config get agent.review --json
x 0 "$MODEL" "" -- config get model.review
cp "$t/repo/.vloop/config.toml" "$t/cfg.set"
x 2 "" 'vloop: invalid value "copilot" for agent.work: want one of claude' -- config set agent.work copilot
cmp -s "$t/cfg.set" "$t/repo/.vloop/config.toml" || fail "a refused config set agent.work changed .vloop/config.toml"
x 0 "" "" -- config set agent.review claude
x 0 claude "" -- config get agent.review

# Every other phase: claude only, in every place a value can come from.
cp "$t/cfg.before" "$t/repo/.vloop/config.toml"
for p in plan work gate-review verify; do
  k=agent.$p v=VLOOP_AGENT_$(up "$p")
  x 2 "" "vloop: invalid value \"copilot\" for $k: want one of claude" -- config set "$k" copilot
  cmp -s "$t/cfg.before" "$t/repo/.vloop/config.toml" || fail "a refused config set $k copilot changed .vloop/config.toml"
  x 1 "" "vloop: $v: invalid value \"copilot\" for $k: want one of claude" "$v=copilot" -- config get "$k"
  x 1 "" "vloop: $v: invalid value \"copilot\" for $k: want one of claude" "$v=copilot" -- config list
  { cat "$t/cfg.before"; printf '\n[agent]\n  %s = "copilot"\n' "$p"; } >"$t/repo/.vloop/config.toml"
  refused 1 -- config get "$k"
  grep -q "invalid value \"copilot\" for $k: want one of claude" "$t/e" || fail "with $k = \"copilot\" in the file, config get does not say why: $(cat "$t/e")"
  (cd "$t/repo" && venv "$S" run "$BRIEF" >"$t/o" 2>&1); c=$?
  [ "$c" = 1 ] || fail "with $k = \"copilot\" in the file, vloop run exited $c, want 1: $(tail -5 "$t/o")"
  [ -f "$t/stub/log" ] && grep -q '^arg -p$' "$t/stub/log" && fail "with $k = \"copilot\" in the file, vloop run started a session"
  cp "$t/cfg.before" "$t/repo/.vloop/config.toml"
  x 0 claude "" -- config get "$k"
done
# agent.review: copilot from the environment and from the file; another name refused.
x 0 '{"key":"agent.review","value":"copilot","source":"env"}' "" VLOOP_AGENT_REVIEW=copilot -- config get agent.review --json
{ cat "$t/cfg.before"; printf '\n[agent]\n  review = "copilot"\n'; } >"$t/repo/.vloop/config.toml"
x 0 '{"key":"agent.review","value":"copilot","source":"file"}' "" -- config get agent.review --json
cp "$t/cfg.before" "$t/repo/.vloop/config.toml"
refused 2 -- config set agent.review grok
grep -q 'invalid value "grok" for agent.review' "$t/e" || fail "config set agent.review grok does not say why: $(cat "$t/e")"
grep -q copilot "$t/e" || fail "the refusal of agent.review grok does not name copilot among the valid values: $(cat "$t/e")"
cmp -s "$t/cfg.before" "$t/repo/.vloop/config.toml" || fail "a refused config set agent.review changed .vloop/config.toml"
refused 1 VLOOP_AGENT_REVIEW=grok -- config get agent.review

# copilot.credit-usd.
x 0 0.0066667 "" -- config get copilot.credit-usd
x 0 '{"key":"copilot.credit-usd","value":"0.0066667","source":"default"}' "" -- config get copilot.credit-usd --json
(cd "$t/repo" && venv "$S" config list >"$t/list" 2>&1) || fail "config list: $(cat "$t/list")"
grep -qx 'copilot.credit-usd=0.0066667 (default)' "$t/list" || fail "config list has no line 'copilot.credit-usd=0.0066667 (default)': $(cat "$t/list")"
grep -qx 'agent.review=claude (default)' "$t/list" || fail "config list has no line 'agent.review=claude (default)'"
x 0 '{"key":"copilot.credit-usd","value":"0.02","source":"env"}' "" VLOOP_COPILOT_CREDIT_USD=0.02 -- config get copilot.credit-usd --json
for bad in 0 -1 abc 0.0; do
  refused 2 -- config set copilot.credit-usd "$bad"
  cmp -s "$t/cfg.before" "$t/repo/.vloop/config.toml" || fail "a refused config set copilot.credit-usd $bad changed .vloop/config.toml"
  refused 1 "VLOOP_COPILOT_CREDIT_USD=$bad" -- config get copilot.credit-usd
  grep -q VLOOP_COPILOT_CREDIT_USD "$t/e" || fail "VLOOP_COPILOT_CREDIT_USD=$bad: the error does not name the variable: $(cat "$t/e")"
done
x 0 "" "" -- config set copilot.credit-usd 0.01
x 0 '{"key":"copilot.credit-usd","value":"0.01","source":"file"}' "" -- config get copilot.credit-usd --json
echo "gate T1: ok"
