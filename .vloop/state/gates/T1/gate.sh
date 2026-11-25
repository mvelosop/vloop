#!/bin/sh
# Gate T1 — L1, the local layer in resolution. Judged through vloop config list
# and config get, on a fixture repository whose two files are written by hand.
G=.vloop/state/gates/T1
. "$G/lib.sh"
L=.vloop/config.local.toml

newrepo r
mkdir -p "$R/.vloop" || fail mkdir
cat >"$R/.vloop/config.toml" <<'EOF'
agent.work = "copilot"
model.work = "gpt-6.1-terra"
run.max-iterations = 7

[metrics]
code = ["a/**"]
test = ["t/**"]
EOF

# 1. No local file, then an empty one: the shared file decides, source file.
runin "$R" config get --json agent.work
lit "$t/w" '{"key":"agent.work","value":"copilot","source":"file"}'
want "no local file: config get --json agent.work" 0 "$t/w" /dev/null
: >"$R/$L"
runin "$R" config get --json agent.work
want "an empty local file: config get --json agent.work" 0 "$t/w" /dev/null

# 2. The local file wins over the shared one, key by key, with source local;
#    a list key takes the local list whole; an empty list sets nothing; keys
#    vloop does not know are ignored.
cat >"$R/$L" <<'EOF'
# this machine only
[agent]
work = "claude"

[model]
work = "sonnet"

[run]
cost-ceiling = 12

[metrics]
code = ["b/**", "c/**"]
test = []

[unknown]
thing = 1
EOF
runin "$R" config list
wantrc "config list with a local file" 0
grep -E '^(agent|model)\.work=' "$t/out" >"$t/got"
printf 'model.work=sonnet (local)\nagent.work=claude (local)\n' >"$t/w"
cmp -s "$t/got" "$t/w" || fail "config list | grep -E '^(agent|model)\\.work=' printed
$(cat "$t/got")
want
$(cat "$t/w")"
hasline "config list" "$t/out" "run.max-iterations=7 (file)"
hasline "config list" "$t/out" "run.cost-ceiling=12 (local)"
hasline "config list" "$t/out" "language=en (default)"

runin "$R" config list --json
wantrc "config list --json with a local file" 0
cp "$t/out" "$t/list.json"
jtrue "$t/list.json" 'd["agent.work"]=={"value":"claude","source":"local"}' "config list --json: agent.work is not claude from local"
jtrue "$t/list.json" 'd["model.work"]["source"]=="local" and d["model.work"]["value"]=="sonnet"' "config list --json: model.work is not sonnet from local"
jtrue "$t/list.json" 'd["metrics.code"]=={"value":["b/**","c/**"],"source":"local"}' "config list --json: metrics.code is not the local list, whole"
jtrue "$t/list.json" 'd["metrics.test"]=={"value":["t/**"],"source":"file"}' "config list --json: an empty local list must set nothing (metrics.test from the shared file)"
jtrue "$t/list.json" 'd["run.max-iterations"]["source"]=="file" and d["run.cost-ceiling"]["source"]=="local"' "config list --json: run.* sources"

runin "$R" config get --json model.work
wantrc "config get --json model.work" 0
cp "$t/out" "$t/get.json"
jtrue "$t/get.json" 'd=={"key":"model.work","value":"sonnet","source":"local"}' "config get --json model.work is not sonnet from local"

# 3. The environment still wins over both.
(cd "$R" && VLOOP_AGENT_WORK=copilot "$t/vloop" config get --json agent.work) >"$t/out" 2>"$t/err"; rc=$?
lit "$t/w" '{"key":"agent.work","value":"copilot","source":"env"}'
want "VLOOP_AGENT_WORK=copilot config get --json agent.work" 0 "$t/w" /dev/null

# 4. A bad value in the local file is reported with the local path, exit 1.
lit "$R/$L" 'agent.work = "nope"'
runin "$R" config get agent.work
lit "$t/w" 'vloop: .vloop/config.local.toml: invalid value "nope" for agent.work: want one of claude, copilot'
want "agent.work = \"nope\" in the local file: config get agent.work" 1 /dev/null "$t/w"
runin "$R" config list
wantrc "agent.work = \"nope\" in the local file: config list" 1
grep -q '^vloop: \.vloop/config\.local\.toml: ' "$t/err" || fail "config list with a bad local value: stderr does not name the local file:
$(cat "$t/err")"

# 5. Malformed TOML in the local file: exit 1, stderr naming the local file.
lit "$R/$L" 'agent.work = '
runin "$R" config get language
wantrc "malformed local file: config get language" 1
head -n 1 "$t/err" | grep -q '^vloop: \.vloop/config\.local\.toml: ' ||
  fail "malformed local file: stderr does not start with vloop: .vloop/config.local.toml:
$(cat "$t/err")"

# 6. [[check]] in the local file is an error wherever the config is read.
printf '[[check]]\nname = "mine"\npaths = ["**"]\nrun = "true"\n' >"$R/$L"
lit "$t/w" "vloop: .vloop/config.local.toml: [[check]] belongs in .vloop/config.toml $EMDASH the checks are the repository's, not one machine's"
runin "$R" config list
want "[[check]] in the local file: config list" 1 /dev/null "$t/w"
runin "$R" config get language
want "[[check]] in the local file: config get language" 1 /dev/null "$t/w"

# 7. Without the local file, everything is as before.
rm -f "$R/$L"
runin "$R" config get --json model.work
lit "$t/w" '{"key":"model.work","value":"gpt-6.1-terra","source":"file"}'
want "local file removed: config get --json model.work" 0 "$t/w" /dev/null

echo "gate T1: ok"
