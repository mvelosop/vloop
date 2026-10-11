#!/bin/sh
# Gate T1 — doctor's plugin check warns when no vloop@<any> entry is listed
# (brief D1), and every other outcome of the check is unchanged.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
newrepo r

plugins '[]'; doc empty
has empty "$PLUGIN_W"
plugins '[{"id":"other@market","version":"1.0.0","enabled":true}]'; doc other
has other "$PLUGIN_W"
doc empty.json --json
[ "$("$JQ" -r '.checks[] | select(.name=="plugin") | .result' "$t/empty.json")" = warning ] || fail "--json: the plugin check is not a warning for []"
want=${PLUGIN_W#! plugin }
[ "$("$JQ" -r '.checks[] | select(.name=="plugin") | .message' "$t/empty.json")" = "$want" ] || fail "--json: the plugin message is not D1's"

# Unchanged outcomes.
plugins "[{\"id\":\"vloop@vloop\",\"version\":\"$V\",\"enabled\":true}]"; doc ok
has ok '✓ plugin'
plugins "[{\"id\":\"vloop@mine\",\"version\":\"$V\",\"enabled\":true}]"; doc okany
has okany '✓ plugin'
plugins "[{\"id\":\"vloop@vloop\",\"version\":\"$V\",\"enabled\":false}]"; doc off
has off '! plugin installed but disabled — claude plugin enable vloop@vloop'
plugins '[{"id":"vloop@vloop","version":"0.0.1","enabled":true}]'; doc old
has old "! plugin the vloop plugin is 0.0.1 but this binary is $V — update the one that is behind"
plugins 'this is not json'; doc bad
grep -q '^! plugin cannot read claude plugin list --json: .' "$t/bad" || fail "bad: an unreadable list does not warn as today: $(grep plugin "$t/bad")"
docnc nc
has nc '- plugin'
echo "gate T1: ok"
