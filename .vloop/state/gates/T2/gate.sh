#!/bin/sh
# Gate T2 — P3 (doctor's `plugin` row counts only this repository's entries,
# honours the repository's opt-out, and says what each case means for a run)
# and P4 (the new `copilot plugin` row, only where doctor's `copilot` row is).
# The binary is built from this tree; `claude` and `copilot` are stubs on a
# PATH holding nothing else but git, sh, cat and dirname. Each stub logs its
# working directory, COPILOT_HOME and arguments, so the gate checks that every
# `plugin` call is `plugin list --json`, run in the repository root, though
# doctor is started from another directory with -C.
G=.vloop/state/gates/T2
. "$G/lib.sh"

newrepo r
vl -C "$R" init --language en >/dev/null 2>&1 || fail "vloop init on the fixture repository failed"
mkdir -p "$t/stub" "$t/sys" "$t/chome" || fail mkdir
for x in git sh cat dirname; do
  p=$(command -v "$x") || fail "$x not on PATH"
  ln -s "$p" "$t/sys/$x" || fail "link $x"
done
P=$t/stub:$t/sys

# The stubs. claude.json / copilot.json hold what `plugin list --json` prints;
# claude.fail / copilot.fail, when present, hold the stderr of a failing list
# (exit 3); copilot.vfail makes `copilot --version` fail.
mkclaude() {
  cat >"$t/stub/claude" <<EOF
#!/bin/sh
printf '%s|%s|%s\n' "\$(pwd -P)" "\${COPILOT_HOME-}" "\$*" >>"$t/claude.log"
if [ "\$*" = "--version" ]; then echo 2.1.0; exit 0; fi
if [ "\$*" = "plugin list --json" ]; then
  if [ -f "$t/claude.fail" ]; then cat "$t/claude.fail" >&2; exit 3; fi
  cat "$t/claude.json"; exit 0
fi
exit 3
EOF
  chmod +x "$t/stub/claude"
}
mkcopilot() {
  cat >"$t/stub/copilot" <<EOF
#!/bin/sh
printf '%s|%s|%s\n' "\$(pwd -P)" "\${COPILOT_HOME-}" "\$*" >>"$t/copilot.log"
if [ "\$*" = "--version" ]; then [ -f "$t/copilot.vfail" ] && exit 1; echo 1.0.105; exit 0; fi
if [ "\$*" = "plugin list --json" ]; then
  if [ -f "$t/copilot.fail" ]; then cat "$t/copilot.fail" >&2; exit 3; fi
  cat "$t/copilot.json"; exit 0
fi
exit 3
EOF
  chmod +x "$t/stub/copilot"
}
mkclaude
mkcopilot

# fill TEXT: TEXT with @V@ as the binary's version and @R@ as the root.
fill() { printf '%s\n' "$1" | sed "s|@V@|$V|g; s|@R@|$R|g"; }

# doctor [DIR]: vloop -C DIR (default R) doctor, from another directory, text
# output in $t/out; then the stubs' logs are checked.
doctor() {
  rm -f "$t/claude.log" "$t/copilot.log"
  (cd "$t/elsewhere" && env PATH="$P" HOME="$t/home" GH_TOKEN=x COPILOT_HOME="$t/chome" "$S" -C "${1:-$R}" doctor) >"$t/out" 2>"$t/err"
  for a in claude copilot; do
    [ -f "$t/$a.log" ] || continue
    awk -F'|' -v r="$R" '$3 ~ /^plugin/ && ($3 != "plugin list --json" || $1 != r) {print; bad=1} END {exit bad}' "$t/$a.log" >"$t/badlog" ||
      fail "$CASE: $a was run as other than \`plugin list --json\` in the root $R: $(cat "$t/badlog")"
  done
  if [ -f "$t/copilot.log" ]; then
    awk -F'|' -v c="$t/chome" '$3 ~ /^plugin/ && $2 != c {bad=1} END {exit bad}' "$t/copilot.log" ||
      fail "$CASE: copilot plugin list ran without the COPILOT_HOME doctor was given"
  fi
}

# row NAME WANT: doctor's output holds the line for row NAME, exactly
#   pass → "✓ NAME", na → "- NAME", problem → starts "✗ NAME", else "! NAME WANT".
row() {
  case $2 in
    pass) l="✓ $1" ;;
    na) l="- $1" ;;
    *) l="! $1 $2" ;;
  esac
  hasline "$t/out" "$l" || fail "$CASE: want the line '$l'; doctor printed:
$(grep -E "^[^ ]* ($1|copilot plugin|plugin)( |\$)" "$t/out")"
}
norow() {
  ! grep -Eq "^[^ ]* $1( |\$)" "$t/out" || fail "$CASE: want no $1 row; got $(grep -E "^[^ ]* $1( |\$)" "$t/out")"
}
# prefix NAME TEXT: the row NAME is a warning whose message starts with TEXT.
prefix() {
  grep -E "^! $1 " "$t/out" | grep -Fq -- "! $1 $2" || fail "$CASE: want a $1 warning starting '$2'; got $(grep -E "^[^ ]* $1( |\$)" "$t/out")"
}

NOTAVAIL='/vloop:operate is not available in your interactive sessions — claude plugin install vloop@vloop --scope project installs it for this repository (after claude plugin marketplace add mvelosop/vloop if claude does not know the marketplace), or claude --plugin-dir "$(vloop plugin path)" for one session'
DISABLED='installed but not enabled for this repository — vloop run is unaffected (its sessions load the plugin with --plugin-dir); for /vloop:operate here, claude plugin enable vloop@vloop --scope project (or with --scope user, if you disabled it for every repository), or start claude --plugin-dir "$(vloop plugin path)"'
older() { echo "the vloop plugin is $1 but this binary is $V — vloop run is unaffected; claude plugin update vloop@vloop updates the plugin, or update the binary if it is the one behind"; }
C_NOTAVAIL='/vloop:operate is not available in your copilot sessions — trusting this folder in copilot installs it when .claude/settings.json declares it (vloop init writes the declaration), or copilot plugin install vloop@vloop (after copilot plugin marketplace add mvelosop/vloop), or copilot --plugin-dir "$(vloop plugin path)" for one session'
C_DISABLED='installed but not enabled for this repository — vloop run is unaffected (its sessions load the plugin with --plugin-dir); for /vloop:operate here, set enabledPlugins["vloop@vloop"] to true in .claude/settings.json and trust this folder in copilot, or start copilot --plugin-dir "$(vloop plugin path)"'
c_older() { echo "the vloop plugin for copilot is $1 but this binary is $V — vloop run is unaffected; copilot plugin update vloop@vloop updates the plugin, or update the binary if it is the one behind"; }

# settings JSON LOCAL: the repository's .claude/settings.json and
# .claude/settings.local.json ("-": absent).
settings() {
  mkdir -p "$R/.claude"
  rm -f "$R/.claude/settings.json" "$R/.claude/settings.local.json"
  [ "$1" = - ] || printf '%s\n' "$1" >"$R/.claude/settings.json"
  [ "$2" = - ] || printf '%s\n' "$2" >"$R/.claude/settings.local.json"
}

# p3 LIST SETTINGS LOCAL WANT: the plugin row for a claude list.
p3() {
  CASE="P3 $1 (settings $2, local $3)"
  fill "$1" >"$t/claude.json"
  settings "$2" "$3"
  doctor
  [ -f "$t/claude.log" ] && grep -q '|plugin list --json$' "$t/claude.log" || fail "$CASE: claude plugin list --json was not run"
  row plugin "$4"
}

# ---- P3 -------------------------------------------------------------------
U='"scope":"user"'
H='"scope":"project","projectPath":"@R@"'
p3 "[{\"id\":\"vloop@vloop\",\"version\":\"@V@\",$U,\"enabled\":true}]" - - pass
p3 "[{\"id\":\"vloop@vloop\",\"version\":\"@V@\",$H,\"enabled\":true}]" - - pass
p3 "[{\"id\":\"vloop@vloop\",\"version\":\"@V@\",\"scope\":\"local\",\"projectPath\":\"@R@\",\"enabled\":true}]" - - pass
p3 '[{"id":"vloop@vloop","version":"@V@","enabled":true}]' - - pass
p3 '[{"id":"other@x","version":"@V@","enabled":true},{"id":"vloop@elsewhere","version":"@V@","enabled":true}]' - - pass
p3 '[{"id":"vloop@vloop","version":"@V@","scope":"project","projectPath":"/elsewhere","enabled":true}]' - - "$NOTAVAIL"
p3 '[{"id":"other@x","version":"@V@","enabled":true}]' - - "$NOTAVAIL"
p3 '[]' - - "$NOTAVAIL"
p3 "[{\"id\":\"vloop@vloop\",\"version\":\"0.0.9\",$U,\"enabled\":true}]" - - "$(older 0.0.9)"
p3 "[{\"id\":\"vloop@vloop\",\"version\":\"0.0.8\",$U,\"enabled\":true},{\"id\":\"vloop@vloop\",\"version\":\"0.0.9\",$H,\"enabled\":true}]" - - "$(older 0.0.8)"
p3 "[{\"id\":\"vloop@vloop\",\"version\":\"0.0.8\",$U,\"enabled\":true},{\"id\":\"vloop@vloop\",\"version\":\"@V@\",$H,\"enabled\":true}]" - - pass
p3 "[{\"id\":\"vloop@vloop\",\"version\":\"@V@\",$U,\"enabled\":false}]" - - "$DISABLED"
p3 "[{\"id\":\"vloop@vloop\",\"version\":\"@V@\",\"scope\":\"project\",\"projectPath\":\"/elsewhere\",\"enabled\":true},{\"id\":\"vloop@vloop\",\"version\":\"@V@\",$U,\"enabled\":false}]" - - "$DISABLED"
p3 "[{\"id\":\"vloop@vloop\",\"version\":\"@V@\",$U,\"enabled\":true}]" '{"enabledPlugins":{"vloop@vloop":false}}' - "$DISABLED"
p3 "[{\"id\":\"vloop@vloop\",\"version\":\"@V@\",$H,\"enabled\":true}]" - '{"enabledPlugins":{"vloop@vloop":false}}' "$DISABLED"
p3 "[{\"id\":\"vloop@vloop\",\"version\":\"@V@\",$U,\"enabled\":true}]" '{"enabledPlugins":{"vloop@vloop":true}}' - pass
p3 "[{\"id\":\"vloop@vloop\",\"version\":\"@V@\",$U,\"enabled\":true}]" '[1]' '{"enabledPlugins":null}' pass
p3 "[{\"id\":\"vloop@vloop\",\"version\":\"@V@\",$U,\"enabled\":true}]" '{' - pass
p3 "[{\"id\":\"vloop@vloop\",\"version\":\"\",$H,\"enabled\":true}]" - - "$NOTAVAIL"
p3 "[{\"id\":\"vloop@vloop\",$H,\"enabled\":true}]" - - "$NOTAVAIL"

# the root reached through a symlink: the entry holds the resolved path
CASE="P3 -C through a symlink"
ln -s "$R" "$t/link" || fail symlink
fill "[{\"id\":\"vloop@vloop\",\"version\":\"@V@\",$H,\"enabled\":true}]" >"$t/claude.json"
settings - -
doctor "$t/link"
row plugin pass

CASE="P3 claude plugin list fails"
echo boom >"$t/claude.fail"
doctor
row plugin "cannot read claude plugin list --json: boom"
rm -f "$t/claude.fail"
CASE="P3 claude plugin list prints no array"
echo '{}' >"$t/claude.json"
doctor
prefix plugin "cannot read claude plugin list --json: "

CASE="P3 no claude"
mv "$t/stub/claude" "$t/claude.off" || fail mv
doctor
row plugin na
norow "copilot plugin"
mv "$t/claude.off" "$t/stub/claude" || fail mv

# ---- P4 -------------------------------------------------------------------
fill "[{\"id\":\"vloop@vloop\",\"version\":\"@V@\",$U,\"enabled\":true}]" >"$t/claude.json"
settings - -
CASE="P4 no agent key is copilot"
echo '[{"name":"vloop","marketplace":"vloop","version":"@V@","enabled":true,"source":"installed"}]' | sed "s|@V@|$V|" >"$t/copilot.json"
doctor
norow "copilot plugin"
[ ! -f "$t/copilot.log" ] || fail "$CASE: the stub copilot was run: $(cat "$t/copilot.log")"

setkey() { vl -C "$R" config set "$1" "$2" >/dev/null 2>&1 || fail "config set $1 '$2'"; }
CASE="P4 agent.verify copilot with run.verify false"
setkey run.verify false; setkey agent.verify copilot
doctor
norow "copilot plugin"
[ ! -f "$t/copilot.log" ] || fail "$CASE: the stub copilot was run: $(cat "$t/copilot.log")"
setkey agent.verify ''; setkey run.verify ''

# p4 LIST WANT: the copilot plugin row, right after plugin, before gate edits.
p4() {
  CASE="P4 $1"
  fill "$1" >"$t/copilot.json"
  doctor
  [ -f "$t/copilot.log" ] && grep -q '|plugin list --json$' "$t/copilot.log" || fail "$CASE: copilot plugin list --json was not run"
  row "copilot plugin" "$2"
  order
}
order() {
  lp=$(grep -nE '^[^ ]* plugin( |$)' "$t/out" | cut -d: -f1)
  [ -n "$lp" ] || fail "$CASE: no plugin row"
  sed -n "$((lp + 1))p" "$t/out" | grep -Eq '^[^ ]* copilot plugin( |$)' &&
    sed -n "$((lp + 2))p" "$t/out" | grep -Eq '^[^ ]* gate edits( |$)' ||
    fail "$CASE: the copilot plugin row is not right after plugin and before gate edits:
$(cat "$t/out")"
}

setkey agent.work copilot
p4 '[{"name":"vloop","marketplace":"vloop","version":"@V@","enabled":true,"source":"installed"}]' pass
p4 '[{"name":"browser-use","marketplace":"","version":"0.0.4","enabled":false,"source":"builtin"},{"name":"vloop","marketplace":"vloop","version":"@V@","enabled":true,"source":"installed"}]' pass
p4 '[{"name":"vloop","marketplace":"vloop","version":"0.0.9","enabled":true,"source":"installed"}]' "$(c_older 0.0.9)"
p4 '[{"name":"vloop","marketplace":"vloop","version":"@V@","enabled":false,"source":"installed"}]' "$C_DISABLED"
p4 '[{"name":"browser-use","marketplace":"","version":"0.0.4","enabled":false,"source":"builtin"}]' "$C_NOTAVAIL"
p4 '[]' "$C_NOTAVAIL"

CASE="P4 copilot plugin list fails"
echo boom >"$t/copilot.fail"
doctor
row "copilot plugin" "cannot read copilot plugin list --json: boom"
rm -f "$t/copilot.fail"
CASE="P4 copilot plugin list prints no array"
echo '{}' >"$t/copilot.json"
doctor
prefix "copilot plugin" "cannot read copilot plugin list --json: "

CASE="P4 copilot --version fails"
: >"$t/copilot.vfail"
doctor
grep -Eq '^✗ copilot( |$)' "$t/out" || fail "$CASE: want a copilot problem: $(cat "$t/out")"
row "copilot plugin" na
rm -f "$t/copilot.vfail"

CASE="P4 copilot not on PATH"
mv "$t/stub/copilot" "$t/copilot.off" || fail mv
doctor
row "copilot plugin" na
mv "$t/copilot.off" "$t/stub/copilot" || fail mv

CASE="P4 no claude, copilot in use"
echo '[{"name":"vloop","marketplace":"vloop","version":"@V@","enabled":true,"source":"installed"}]' | sed "s|@V@|$V|" >"$t/copilot.json"
mv "$t/stub/claude" "$t/claude.off" || fail mv
doctor
row plugin na
row "copilot plugin" pass
order
mv "$t/claude.off" "$t/stub/claude" || fail mv

CASE="P4 agent.review-brief copilot only"
setkey agent.work ''; setkey agent.review-brief copilot
doctor
row "copilot plugin" pass
setkey agent.review-brief ''

CASE="P4 the config cannot be read"
setkey agent.work copilot
cp "$R/.vloop/config.toml" "$t/config.save" || fail cp
printf 'this is not = = toml [\n' >>"$R/.vloop/config.toml"
doctor
grep -Eq '^✗ config( |$)' "$t/out" || fail "$CASE: the fixture did not make the config unreadable: $(cat "$t/out")"
norow "copilot plugin"
cp "$t/config.save" "$R/.vloop/config.toml" || fail cp

# vloop run's preflight reuses doctor's checks and runs no `copilot plugin`:
# with agent.work copilot and an untrusted folder it stops after its checks.
CASE="P4 vloop run's preflight"
rm -f "$t/claude.log" "$t/copilot.log"
(cd "$t/elsewhere" && env PATH="$P" HOME="$t/home" GH_TOKEN=x COPILOT_HOME="$t/chome" "$S" -C "$R" run) >"$t/out" 2>"$t/err"
rc=$?
[ "$rc" = 1 ] && grep -q 'preflight failed' "$t/err" || fail "$CASE: want the preflight to refuse (exit 1); exit $rc: $(cat "$t/err")"
grep -q '|--version$' "$t/copilot.log" 2>/dev/null || fail "$CASE: the preflight did not check copilot, so it did not reach doctor's checks"
! grep -q '|plugin' "$t/copilot.log" || fail "$CASE: the preflight ran copilot plugin: $(cat "$t/copilot.log")"

echo "gate T2: ok"
