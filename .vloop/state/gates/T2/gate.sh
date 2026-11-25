#!/bin/sh
# Gate T2 — L2, writing the local file: vloop config set --local, config path
# --local and the stderr notice of config set. Judged through the CLI, with the
# resolved values read back through config get --json.
G=.vloop/state/gates/T2
. "$G/lib.sh"
L=.vloop/config.local.toml

# 1. config path, with and without --local.
newrepo p
runin "$R" config path --local
lit "$t/w" '.vloop/config.local.toml'
want "config path --local" 0 "$t/w" /dev/null
runin "$R" config path
lit "$t/w" '.vloop/config.toml'
want "config path" 0 "$t/w" /dev/null

# 2. Removing a key that is not there creates nothing; writing one creates the
#    file (and .vloop/), and no shared file.
runin "$R" config set --local agent.work ''
want "config set --local agent.work '' with no file" 0 /dev/null /dev/null
[ ! -e "$R/.vloop" ] || fail "config set --local agent.work '' created $(cd "$R" && find .vloop)"
runin "$R" config set --local agent.work claude
want "config set --local agent.work claude" 0 /dev/null /dev/null
[ -f "$R/$L" ] || fail "config set --local did not create $L"
[ ! -e "$R/.vloop/config.toml" ] || fail "config set --local created .vloop/config.toml"
runin "$R" config get --json agent.work
lit "$t/w" '{"key":"agent.work","value":"claude","source":"local"}'
want "after config set --local agent.work claude" 0 "$t/w" /dev/null

# 3. The worked example's sequence, in a repository with a shared file.
newrepo r
mkdir -p "$R/.vloop" && printf '%s' "$CHECKCFG" >"$R/.vloop/config.toml" || fail "shared config"
for kv in "agent.work copilot" "model.work gpt-6.1-terra"; do
  # shellcheck disable=SC2086
  runin "$R" config set $kv
  want "config set $kv" 0 /dev/null /dev/null
done
for kv in "agent.work claude" "model.work sonnet"; do
  # shellcheck disable=SC2086
  runin "$R" config set --local $kv
  want "config set --local $kv" 0 /dev/null /dev/null
done
runin "$R" config list
wantrc "config list" 0
grep -E '^(agent|model)\.work=' "$t/out" >"$t/got"
printf 'model.work=sonnet (local)\nagent.work=claude (local)\n' >"$t/w"
cmp -s "$t/got" "$t/w" || fail "config list | grep -E '^(agent|model)\\.work=' printed
$(cat "$t/got")
want
$(cat "$t/w")"

# The shared file is written, and the notice names the local file that wins.
runin "$R" config set model.work gpt-6-luna
lit "$t/w" '! model.work stays sonnet on this machine: .vloop/config.local.toml overrides it'
want "config set model.work gpt-6-luna under a local override" 0 /dev/null "$t/w"

# The environment overrides: the notice names the variable, also when the
# local file overrides too, and also for config set --local.
(cd "$R" && VLOOP_AGENT_WORK=copilot "$t/vloop" config set agent.work claude) >"$t/out" 2>"$t/err"; rc=$?
lit "$t/w" '! agent.work stays copilot on this machine: VLOOP_AGENT_WORK overrides it'
want "VLOOP_AGENT_WORK=copilot config set agent.work claude (local file overrides too)" 0 /dev/null "$t/w"
(cd "$R" && VLOOP_AGENT_WORK=copilot "$t/vloop" config set --local agent.work claude) >"$t/out" 2>"$t/err"; rc=$?
want "VLOOP_AGENT_WORK=copilot config set --local agent.work claude" 0 /dev/null "$t/w"
(cd "$R" && VLOOP_LANGUAGE=es "$t/vloop" config set language en) >"$t/out" 2>"$t/err"; rc=$?
lit "$t/w" '! language stays es on this machine: VLOOP_LANGUAGE overrides it'
want "VLOOP_LANGUAGE=es config set language en" 0 /dev/null "$t/w"

# No override, no notice.
runin "$R" config set language es
want "config set language es (nothing overrides it)" 0 /dev/null /dev/null

# Removing local keys: the shared values show through again.
runin "$R" config set --local model.work ''
want "config set --local model.work ''" 0 /dev/null /dev/null
runin "$R" config get --json model.work
lit "$t/w" '{"key":"model.work","value":"gpt-6-luna","source":"file"}'
want "after removing model.work from the local file" 0 "$t/w" /dev/null
runin "$R" config get --json agent.work
lit "$t/w" '{"key":"agent.work","value":"claude","source":"local"}'
want "agent.work kept in the local file" 0 "$t/w" /dev/null

# 4. Validation as config set's: exit 2, and the local file unchanged.
cp "$R/$L" "$t/before"
runin "$R" config set --local agent.work nope
lit "$t/w" 'vloop: invalid value "nope" for agent.work: want one of claude, copilot'
want "config set --local agent.work nope" 2 /dev/null "$t/w"
runin "$R" config set --local colour red
lit "$t/w" 'vloop: unknown config key "colour"'
want "config set --local colour red" 2 /dev/null "$t/w"
cmp -s "$t/before" "$R/$L" || fail "a refused config set --local changed $L"

# 5. The file's other keys and layout are kept; removing an absent key writes nothing.
printf '# mine, not shared\nrun.max-iterations = 9\n' >"$R/$L"
runin "$R" config set --local agent.work copilot
want "config set --local agent.work copilot on a hand-written file" 0 /dev/null /dev/null
grep -qxF '# mine, not shared' "$R/$L" || fail "config set --local dropped the file's comment:
$(cat "$R/$L")"
runin "$R" config get --json run.max-iterations
wantrc "config get --json run.max-iterations" 0
cp "$t/out" "$t/get.json"
jtrue "$t/get.json" 'str(d["value"])=="9" and d["source"]=="local"' "config set --local dropped run.max-iterations from the local file"
cp "$R/$L" "$t/before"
runin "$R" config set --local model.work ''
want "config set --local model.work '' (absent)" 0 /dev/null /dev/null
cmp -s "$t/before" "$R/$L" || fail "removing an absent key changed $L"

echo "gate T2: ok"
