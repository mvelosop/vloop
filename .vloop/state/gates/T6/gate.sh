#!/bin/sh
# Gate T6 — L5, the price of a Copilot credit: copilot.credit-usd defaults to
# 0.01, its validation is unchanged, and the configuration guide's key table
# states the new default. Judged through vloop config and the guide's table row.
G=.vloop/state/gates/T6
. "$G/lib.sh"

newrepo r
runin "$R" config get copilot.credit-usd
lit "$t/w" '0.01'
want "config get copilot.credit-usd with no file" 0 "$t/w" /dev/null
runin "$R" config get --json copilot.credit-usd
lit "$t/w" '{"key":"copilot.credit-usd","value":"0.01","source":"default"}'
want "config get --json copilot.credit-usd" 0 "$t/w" /dev/null
runin "$R" config list
wantrc "config list" 0
hasline "config list" "$t/out" "copilot.credit-usd=0.01 (default)"

# Validation unchanged: a number greater than 0.
for v in 0 -1 abc; do
  runin "$R" config set copilot.credit-usd "$v"
  wantrc "config set copilot.credit-usd $v" 2
done
runin "$R" config set copilot.credit-usd 0.0066667
want "config set copilot.credit-usd 0.0066667" 0 /dev/null /dev/null
runin "$R" config get copilot.credit-usd
lit "$t/w" '0.0066667'
want "config get copilot.credit-usd after setting it" 0 "$t/w" /dev/null

# The guide's key table: the row of copilot.credit-usd, default column 0.01.
awk -F'|' '$2 ~ /^ *`copilot\.credit-usd` *$/ { d = $3; gsub(/^ +| +$/, "", d); print d }' docs/guide/configuration.md >"$t/col"
lit "$t/w" '`0.01`'
cmp -s "$t/col" "$t/w" || fail "docs/guide/configuration.md: the copilot.credit-usd row's default column is
$(cat "$t/col")
want \`0.01\` (one row)"

echo "gate T6: ok"
