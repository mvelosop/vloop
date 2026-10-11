#!/bin/sh
# Gate T7 — close (brief: Worked example, Real-data check): the worked example
# line for line on a fresh build; the real-data check in this repository against
# a build of the plan's base; every relative link in the README, the guides and
# the domain model resolves to an existing file and heading.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
BASE=6b67e496945aa7fd599ea807cbf5d9e1362cc9f8
after() { awk 'p { print; exit } /^[^ ]+ plugin( |$)/ { p = 1 }' "$t/$1"; }
expect() { a=$(after "$1"); [ "$a" = "$2" ] || fail "$1: after the plugin line want '$2', got '$a'"; }
clear() { rm -rf "$R/.claude" "$HOME/.claude"; }

# --- the worked example ---
newrepo r
tail -n 7 "$t/r.init" | cmp -s - "$G/next.txt" || fail "vloop init does not end with D3's seven lines"
grep -Fq 'If /vloop:operate is not available, the vloop plugin is not loaded' "$R/CLAUDE.md" || fail "the new CLAUDE.md lacks D3's sentence"
clear; plugins '[]'; doc ex
[ "$got" = 0 ] || fail "ex: exit $got, want 0"
has ex "$PLUGIN_W"
expect ex "$GATE_W"
has ex 'doctor: 0 problem(s), 2 warning(s)'
plugins "[{\"id\":\"vloop@vloop\",\"version\":\"$V\",\"enabled\":true}]"; doc ok
has ok '✓ plugin'
mkdir -p "$R/.claude" && printf '{"permissions":{"allow":["Edit(./.vloop/state/gates/**)"]}}\n' >"$R/.claude/settings.json" || fail settings
doc p1; expect p1 '✓ gate edits'
clear; mkdir -p "$R/.claude" && printf '{"permissions":{"allow":["Edit(.vloop/**)"]}}\n' >"$R/.claude/settings.local.json" || fail settings
doc p2; expect p2 '✓ gate edits'
clear; mkdir -p "$HOME/.claude" && printf '{"permissions":{"allow":["Edit"]}}\n' >"$HOME/.claude/settings.json" || fail settings
doc p3; expect p3 '✓ gate edits'
clear; mkdir -p "$R/.claude" && printf '{"permissions":{"allow":["Edit(./src/**)","Bash(go:*)"]}}\n' >"$R/.claude/settings.json" || fail settings
doc w1; expect w1 "$GATE_W"
printf '{"permissions":\n' >"$R/.claude/settings.json"
doc w2
case "$(after w2)" in "! gate edits cannot read .claude/settings.json: "?*) ;; *) fail "w2: got '$(after w2)'" ;; esac
clear; docnc nc
has nc '- plugin'
expect nc '- gate edits'
doc js --json
"$JQ" -e '[.checks[] | select(.name == "gate edits" and .result != "" )] | length == 1' "$t/js" >/dev/null || fail "--json has no check named 'gate edits'"

# --- the real-data check, in this repository, against the base's build ---
mkdir -p "$t/base" && git -C "$top" archive "$BASE" | tar -x -C "$t/base" || fail "git archive $BASE"
(cd "$t/base" && go build -buildvcs=false -o "$t/vloop-base" ./cmd/vloop) || fail "build of the base"
plugins '[]'
(cd "$top" && env PATH="$t/stub:$SYS" STUB_PLUGINS="$t/plugins.json" "$t/vloop-base" doctor --json) >"$t/before.json" 2>/dev/null
(cd "$top" && env PATH="$t/stub:$SYS" STUB_PLUGINS="$t/plugins.json" "$B" doctor --json) >"$t/after.json" 2>/dev/null
(cd "$top" && env PATH="$t/stub:$SYS" STUB_PLUGINS="$t/plugins.json" "$B" doctor) >"$t/after.txt" 2>/dev/null
has after.txt '✓ gate edits'
has after.txt "$PLUGIN_W"
b=$("$JQ" -c '[.checks[] | select(.name != "plugin" and .name != "gate edits") | [.name, .result]]' "$t/before.json")
a=$("$JQ" -c '[.checks[] | select(.name != "plugin" and .name != "gate edits") | [.name, .result]]' "$t/after.json")
[ -n "$b" ] && [ "$a" = "$b" ] || fail "real data: another check's result changed:
before $b
after  $a"
[ "$("$JQ" -c '[.checks[].name]' "$t/after.json")" = "$("$JQ" -c '[.checks[].name] | (index("plugin") + 1) as $i | .[:$i] + ["gate edits"] + .[$i:]' "$t/before.json")" ] || fail "real data: the checks are not the base's with 'gate edits' after plugin"

# --- every link resolves ---
(cd "$top" && go run "$G/links.go" README.md docs/guide/*.md docs/domain/domain-model.md) || fail "a link in the README, the guides or the domain model does not resolve"
echo "gate T7: ok"
