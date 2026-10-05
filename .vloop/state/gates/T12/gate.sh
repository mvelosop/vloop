#!/bin/sh
# Gate T12 — vloop doctor's plugin check (brief Q6), against a fake claude
# whose `plugin list --json` answer the gate controls: no marketplace plugin
# passes with the --plugin-dir hint; a vloop plugin installed but disabled is a
# warning naming `claude plugin enable vloop@vloop`; a failing `claude plugin
# list` is a warning naming the error, not swallowed; an enabled plugin of the
# binary's version still passes.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
export GIT_AUTHOR_NAME=gate GIT_AUTHOR_EMAIL=gate@example.invalid GIT_COMMITTER_NAME=gate GIT_COMMITTER_EMAIL=gate@example.invalid
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -o "$t/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
GOOS=linux go build -o "$t/vloop-linux" ./cmd/vloop || fail "GOOS=linux go build ./cmd/vloop"
GOOS=windows go build -o "$t/vloop.exe" ./cmd/vloop || fail "GOOS=windows go build ./cmd/vloop"
B="$t/vloop"
mkdir -p "$t/home" "$t/bin" || fail mkdir
# The fake claude: --version works; `plugin …` prints $t/bin/plugins.out on
# stdout, $t/bin/plugins.err on stderr, and exits with $t/bin/plugins.code.
cat >"$t/bin/claude" <<'EOF'
#!/bin/sh
d=$(dirname "$0")
case "$1" in
  --version) echo "2.0.0 (Claude Code)"; exit 0;;
  plugin) cat "$d/plugins.out"; cat "$d/plugins.err" >&2; exit "$(cat "$d/plugins.code")";;
esac
exit 1
EOF
chmod +x "$t/bin/claude" || fail chmod
export HOME="$t/home" PATH="$t/bin:$PATH"
R="$t/r"
mkdir -p "$R" && git -C "$R" init -q -b main && git -C "$R" commit -q --allow-empty -m "gate base" || fail "git init"
(cd "$R" && "$B" init </dev/null) >"$t/init.out" 2>&1 || fail "vloop init: $(cat "$t/init.out")"
VER=$("$B" version | awk '{ print $2; exit }')
[ -n "$VER" ] || fail "no version from vloop version"
# plugin <stdout> <stderr> <code>: doctor's plugin line with that answer.
plugin() {
  printf '%s\n' "$1" >"$t/bin/plugins.out"; printf '%s' "$2" >"$t/bin/plugins.err"; echo "$3" >"$t/bin/plugins.code"
  (cd "$R" && "$B" doctor) >"$t/doctor.out" 2>&1
  grep '^[^ ]* plugin ' "$t/doctor.out" | head -1
}

line=$(plugin '[]' '' 0)
want='✓ plugin the plugin is supplied by vloop run (--plugin-dir); for an interactive session: claude --plugin-dir "$(vloop plugin path)"'
[ "$line" = "$want" ] || fail "with no marketplace plugin, doctor prints '$line', want '$want'"

line=$(plugin '[{"id":"other@market","version":"1.0.0","enabled":true}]' '' 0)
[ "$line" = "$want" ] || fail "with only another plugin installed, doctor prints '$line', want '$want'"

line=$(plugin "[{\"id\":\"vloop@vloop\",\"version\":\"$VER\",\"enabled\":false}]" '' 0)
echo "$line" | grep -q '^! plugin ' || fail "an installed but disabled vloop plugin is not a warning: '$line'"
echo "$line" | grep -Fq 'installed but disabled — claude plugin enable vloop@vloop' || fail "an installed but disabled vloop plugin does not say 'installed but disabled — claude plugin enable vloop@vloop': '$line'"

line=$(plugin 'garbage' 'claude: the plugin registry is locked' 3)
echo "$line" | grep -q '^! plugin ' || fail "a failing claude plugin list is not a warning: '$line'"
echo "$line" | grep -Eq 'the plugin registry is locked|exit status 3' || fail "a failing claude plugin list is reported without its error: '$line'"

line=$(plugin "[{\"id\":\"vloop@vloop\",\"version\":\"$VER\",\"enabled\":true}]" '' 0)
echo "$line" | grep -q '^✓ plugin' || fail "an enabled vloop plugin of the binary's version does not pass: '$line'"
echo "gate T12: ok"
