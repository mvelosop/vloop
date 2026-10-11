#!/bin/sh
# Gate T2 — doctor's new `gate edits` check (brief D2): printed right after
# plugin; passes only for the listed Edit forms in the user, project or local
# settings; warns otherwise, and names a settings file it cannot read; is `-`
# with no claude; is in --json; writes no file; and changes no exit code.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
newrepo r
# clear: no settings file anywhere.
clear() { rm -rf "$R/.claude" "$HOME/.claude"; }
# put <user|project|local> <json>: write that settings file.
put() {
  case "$1" in
  user) f="$HOME/.claude/settings.json" ;;
  project) f="$R/.claude/settings.json" ;;
  local) f="$R/.claude/settings.local.json" ;;
  esac
  mkdir -p "$(dirname "$f")" && printf '%s\n' "$2" >"$f" || fail "write $1 settings"
}
# after <out>: the line printed right after the plugin line.
after() { awk 'p { print; exit } /^[^ ]+ plugin( |$)/ { p = 1 }' "$t/$1"; }
# expect <out> <line>: the line after plugin is exactly <line>.
expect() { a=$(after "$1"); [ "$a" = "$2" ] || fail "$1: after the plugin line want '$2', got '$a' in:
$(cat "$t/$1")"; }
# allow <where> <entry>: a settings file whose permissions.allow is [<entry>].
allow() { put "$1" "{\"permissions\":{\"allow\":[\"$2\"]}}"; }

# --- the worked example: no settings files, no plugin ---
clear; plugins '[]'
st0=$(git -C "$R" status --porcelain --untracked-files=all)
doc none
expect none "$GATE_W"
has none "$PLUGIN_W"
[ "$got" = 0 ] || fail "none: exit $got, want 0 (both lines are warnings)"
has none 'doctor: 0 problem(s), 2 warning(s)'
[ -e "$R/.claude" ] && fail "doctor created $R/.claude"
[ -e "$HOME/.claude" ] && fail "doctor created ~/.claude"
[ "$(git -C "$R" status --porcelain --untracked-files=all)" = "$st0" ] || fail "doctor changed the repository"

# --- passing forms, in each of the three files ---
n=0
for spec in \
  'project|Edit(./.vloop/state/gates/**)' 'project|Edit(.vloop/state/gates/**)' \
  'local|Edit(.vloop/**)' 'local|Edit(./.vloop/**)' 'user|Edit' \
  'project|Edit(**)' 'user|Edit(./**)' 'local|Edit(.vloop/state/**)' 'user|Edit(./.vloop/state/**)'; do
  n=$((n + 1)); clear
  allow "${spec%%|*}" "${spec#*|}"
  doc pass$n
  expect pass$n '✓ gate edits'
done
# a pass in one file is enough when the others exist and do not allow it
clear; put user '{"permissions":{"allow":["Bash(ls:*)"]}}'; put project '{}'; allow local 'Edit(./.vloop/state/gates/**)'
doc mix
expect mix '✓ gate edits'

# --- forms outside the list warn, never pass ---
n=0
for spec in \
  'project|Edit(./src/**)' 'project|Edit(/.vloop/**)' 'project|Edit(~/.vloop/**)' \
  'local|Edit(.vloop/state/gates/*)' 'local|Edit(.vloop/state/gates/)' 'user|Write(.vloop/**)' \
  'project|Edit(././.vloop/**)' 'user|MultiEdit' 'local|Edit(.vloop/state/gates/T1/**)'; do
  n=$((n + 1)); clear
  allow "${spec%%|*}" "${spec#*|}"
  doc warn$n
  expect warn$n "$GATE_W"
done
clear; put project '{"permissions":{"allow":["Edit(./src/**)","Bash(go:*)"]}}'; doc we
expect we "$GATE_W"
clear; put project '{"permissions":{"deny":["Edit(.vloop/**)"],"ask":["Edit"],"defaultMode":"acceptEdits"}}'; doc deny
expect deny "$GATE_W"
clear; put project '{}'; put user '{"permissions":{}}'; doc empty
expect empty "$GATE_W"

# --- a file that is not JSON is named ---
for spec in 'project|.claude/settings.json' 'local|.claude/settings.local.json' 'user|~/.claude/settings.json'; do
  clear; put "${spec%%|*}" '{"permissions":'
  doc bad
  a=$(after bad)
  case "$a" in
  "! gate edits cannot read ${spec#*|}: "?*) ;;
  *) fail "bad ${spec%%|*}: want '! gate edits cannot read ${spec#*|}: <error>', got '$a'" ;;
  esac
done

# --- doctor writes no settings file ---
clear; allow project 'Edit(./src/**)'; allow user 'Edit(.vloop/**)'
cp "$R/.claude/settings.json" "$t/p.keep" && cp "$HOME/.claude/settings.json" "$t/u.keep" || fail keep
doc keep
cmp -s "$R/.claude/settings.json" "$t/p.keep" || fail "doctor changed .claude/settings.json"
cmp -s "$HOME/.claude/settings.json" "$t/u.keep" || fail "doctor changed ~/.claude/settings.json"
[ -e "$R/.claude/settings.local.json" ] && fail "doctor wrote .claude/settings.local.json"

# --- no claude: both are n/a ---
clear; docnc nc
has nc '- plugin'
expect nc '- gate edits'

# --- --json: a check named "gate edits", right after plugin ---
clear; doc j1 --json
names=$("$JQ" -r '[.checks[].name] | (index("plugin")|tostring) + " " + (index("gate edits")|tostring)' "$t/j1")
p=${names% *} e=${names#* }
[ "$e" != null ] && [ "$e" = $((p + 1)) ] || fail "--json: 'gate edits' is not the check right after plugin ($names)"
[ "$("$JQ" -r '.checks[] | select(.name=="gate edits") | .result + "|" + .message' "$t/j1")" = "warning|${GATE_W#! gate edits }" ] || fail "--json: gate edits is not the warning"
allow project 'Edit(./.vloop/state/gates/**)'; doc j2 --json
[ "$("$JQ" -r '.checks[] | select(.name=="gate edits") | .result + "|" + .message' "$t/j2")" = "pass|" ] || fail "--json: gate edits does not pass with no message"
echo "gate T2: ok"
