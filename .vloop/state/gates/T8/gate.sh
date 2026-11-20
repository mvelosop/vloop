#!/bin/sh
# Gate T8 — close: the worked example's first block, line for line and in its
# order, in one fixture repository; then every earlier gate of the plan, at once.
G=.vloop/state/gates/T8
. "$G/lib.sh"

newrepo r
runin "$R" init
wantrc "vloop init" 0
printf '\n%s' "$CHECKCFG" >>"$R/.vloop/config.toml" || fail "add a check"

# step NAME RC OUT ERR ARGS...: vloop ARGS in the repository, judged by want.
step() {
  n=$1 c=$2 o=$3 e=$4; shift 4
  runin "$R" "$@"
  want "$n" "$c" "$o" "$e"
}
step "config set agent.work copilot" 0 /dev/null /dev/null config set agent.work copilot
step "config set model.work gpt-6.1-terra" 0 /dev/null /dev/null config set model.work gpt-6.1-terra
step "config set --local agent.work claude" 0 /dev/null /dev/null config set --local agent.work claude
step "config set --local model.work sonnet" 0 /dev/null /dev/null config set --local model.work sonnet
lit "$t/w" '.vloop/config.local.toml'
step "config path --local" 0 "$t/w" /dev/null config path --local

runin "$R" config list
wantrc "config list" 0
grep -E '^(agent|model)\.work=' "$t/out" >"$t/got"
printf 'model.work=sonnet (local)\nagent.work=claude (local)\n' >"$t/w"
cmp -s "$t/got" "$t/w" || fail "config list | grep -E '^(agent|model)\\.work=' printed
$(cat "$t/got")"

(cd "$R" && VLOOP_AGENT_WORK=copilot "$t/vloop" config get --json agent.work) >"$t/out" 2>"$t/err"; rc=$?
lit "$t/w" '{"key":"agent.work","value":"copilot","source":"env"}'
want "VLOOP_AGENT_WORK=copilot config get --json agent.work" 0 "$t/w" /dev/null

lit "$t/w" '! model.work stays sonnet on this machine: .vloop/config.local.toml overrides it'
step "config set model.work gpt-6-luna" 0 /dev/null "$t/w" config set model.work gpt-6-luna

grep -x '.vloop/config.local.toml' "$R/.gitignore" >"$t/got" || fail "grep -x .vloop/config.local.toml .gitignore: no line"
lit "$t/w" '.vloop/config.local.toml'
cmp -s "$t/got" "$t/w" || fail "grep -x .vloop/config.local.toml .gitignore printed $(cat "$t/got")"

runin "$R" doctor
grep config.local "$t/out" >"$t/got"
lit "$t/w" "! config.local overrides model.work, agent.work $EMDASH on this machine only"
cmp -s "$t/got" "$t/w" || fail "vloop doctor | grep config.local printed
$(cat "$t/got")"

step "config set --local agent.work ''" 0 /dev/null /dev/null config set --local agent.work ''
step "config set --local model.work ''" 0 /dev/null /dev/null config set --local model.work ''
runin "$R" doctor
grep config.local "$t/out" >"$t/got"
lit "$t/w" "✓ config.local"
cmp -s "$t/got" "$t/w" || fail "vloop doctor | grep config.local, after removing the keys, printed
$(cat "$t/got")"

lit "$t/w" '0.01'
step "config get copilot.credit-usd" 0 "$t/w" /dev/null config get copilot.credit-usd

# Every earlier gate, each in its own shell.
for n in 1 2 3 4 5 6 7; do
  sh ".vloop/state/gates/T$n/gate.sh" >"$t/gate.out" 2>&1 || fail "gate T$n fails:
$(tail -n 30 "$t/gate.out")"
done

echo "gate T8: ok"
