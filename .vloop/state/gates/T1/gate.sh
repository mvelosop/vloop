#!/bin/sh
# Gate T1 — P1, the declaration: `vloop init` and `vloop upgrade` merge
# extraKnownMarketplaces.vloop and enabledPlugins["vloop@vloop"] into
# .claude/settings.json, on fresh temporary repositories, through the binary
# built from this tree. Every P1 line of the brief's worked example, plus the
# merge rules beyond it (layout, numbers, an existing marketplace, kept files,
# the key at hand in the refusal, the symlink, upgrade's breaking path and its
# `already at` return). Expected files are in cases/, byte for byte.
G=.vloop/state/gates/T1
C=$G/cases
. "$G/lib.sh"
SET=.claude/settings.json
NOTOBJ="vloop: .claude/settings.json is not a JSON object — fix it, then re-run"
notobj() { echo "vloop: .claude/settings.json's $1 is not a JSON object — fix it, then re-run"; }

# put NAME CONTENT-FILE: write R's settings file from a fixture (none: absent).
put() { mkdir -p "$R/.claude" && cp "$1" "$R/$SET" || fail "write $SET"; }

# same NAME WANT: R's settings file is WANT's bytes.
same() {
  [ -f "$R/$SET" ] || fail "$1: $SET was not written"
  cmp -s "$R/$SET" "$2" || fail "$1: $SET is not, byte for byte, $2:
$(cat "$R/$SET")"
}

# run NAME ARGS: vloop -C R ARGS, stdout in $t/out, stderr in $t/err, exit in rc.
run() { n=$1; shift; vl -C "$R" "$@" >"$t/out" 2>"$t/err"; rc=$?; }
ok() { [ "$rc" = 0 ] || fail "$n: exit $rc, want 0: $(cat "$t/err")"; }
refused() {
  [ "$rc" = 1 ] || fail "$n: exit $rc, want 1 ($(cat "$t/err"))"
  hasline "$t/err" "$1" || fail "$n: stderr lacks the line '$1': $(cat "$t/err")"
}
says() { hasline "$t/out" "$1" || fail "$n: stdout lacks the line '$1':
$(cat "$t/out")"; }
silent() { ! grep -Fq -- "$SET" "$t/out" || fail "$n: stdout names $SET: $(grep -F -- "$SET" "$t/out")"; }

# ---- init -------------------------------------------------------------
newrepo i-fresh
run "init, no file" init --language en; ok
says "wrote $SET"
same "init, no file" "$C/fresh.want"
lc=$(grep -n -x 'wrote CLAUDE.md' "$t/out" | cut -d: -f1)
ls=$(grep -n -x "wrote $SET" "$t/out" | cut -d: -f1)
[ -n "$lc" ] && [ "$ls" -gt "$lc" ] || fail "init: the $SET step does not come after the CLAUDE.md step:
$(cat "$t/out")"

newrepo i-dry
run "init --dry-run, no file" init --dry-run --language en; ok
says "would write $SET"
[ ! -e "$R/.claude" ] && [ ! -e "$R/.vloop" ] || fail "init --dry-run wrote files"

newrepo i-merge
put "$C/example.in"
run "init, a file to merge" init --language en; ok
says "updated $SET"
same "init, a file to merge" "$C/example.want"

newrepo i-kept
put "$C/kept.in"
run "init --dry-run, both entries present" init --dry-run --language en; ok
says "would keep $SET"
run "init, both entries present" init --language en; ok
says "kept $SET"
cmp -s "$R/$SET" "$C/kept.in" || fail "init rewrote a $SET it kept: $(cat "$R/$SET")"

for cs in '[1]' '{' '"x"'; do
  for dry in "" --dry-run; do
    newrepo i-bad
    printf '%s\n' "$cs" >"$t/bad.in"; put "$t/bad.in"
    before=$(snap "$R")
    run "init $dry on $cs" init $dry --language en
    refused "$NOTOBJ"
    ! grep -q '^would ' "$t/out" || fail "$n: printed a would line: $(cat "$t/out")"
    [ "$(snap "$R")" = "$before" ] || fail "$n: files were written"
    rm -rf "$R"
  done
done
for key in enabledPlugins extraKnownMarketplaces; do
  for val in null '[]' '"x"'; do
    newrepo i-key
    printf '{"%s":%s}\n' "$key" "$val" >"$t/bad.in"; put "$t/bad.in"
    before=$(snap "$R")
    run "init on $key=$val" init --language en
    refused "$(notobj "$key")"
    [ "$(snap "$R")" = "$before" ] || fail "$n: files were written"
    rm -rf "$R"
  done
done

newrepo i-link
printf '{}\n' >"$t/target.json"
mkdir -p "$R/.claude" && ln -s "$t/target.json" "$R/$SET" || fail symlink
before=$(snap "$R")
run "init on a symlinked $SET" init --language en
[ "$rc" = 1 ] || fail "$n: exit $rc, want 1 (refused as a symlink)"
[ "$(snap "$R")" = "$before" ] && [ "$(cat "$t/target.json")" = "{}" ] || fail "$n: files were written"

# ---- upgrade: a repository set up by 0.0.9 --------------------------------
# uprepo NAME [FIXTURE]: initialised, stamp 0.0.9, settings from FIXTURE or absent.
uprepo() {
  newrepo "$1"
  vl -C "$R" init --language en >/dev/null 2>&1 || fail "$1: vloop init failed"
  printf '{"schema":"install/v1","version":"0.0.9","commit":"x","module":"x","initialized":"2026-01-01T00:00:00Z"}\n' >"$R/.vloop/install.json"
  rm -rf "$R/.claude"
  [ $# -lt 2 ] || put "$2"
}

for c in example amp optout layout partial ownmk; do
  uprepo "u-$c" "$C/$c.in"
  run "upgrade --yes on $c.in" upgrade --yes; ok
  says "updated $SET"
  same "$n" "$C/$c.want"
done

uprepo u-absent
run "upgrade --yes, no file" upgrade --yes; ok
says "updated $SET"
same "$n" "$C/fresh.want"

for c in fresh.want kept.in; do
  uprepo u-has "$C/$c"
  run "upgrade --yes, both entries present ($c)" upgrade --yes; ok
  silent
  cmp -s "$R/$SET" "$C/$c" || fail "$n: the file changed: $(cat "$R/$SET")"
  rm -rf "$R"
done

for cs in '[1]' '{"enabledPlugins":null}' '{"extraKnownMarketplaces":[]}'; do
  case $cs in
    *enabledPlugins*) want=$(notobj enabledPlugins) ;;
    *extraKnown*) want=$(notobj extraKnownMarketplaces) ;;
    *) want=$NOTOBJ ;;
  esac
  for args in "--yes" "--dry-run" ""; do
    uprepo u-bad
    printf '%s\n' "$cs" >"$t/bad.in"; put "$t/bad.in"
    before=$(snap "$R")
    run "upgrade $args on $cs" upgrade $args
    refused "$want"
    ! grep -q '^would update' "$t/out" || fail "$n: printed a would update line: $(cat "$t/out")"
    [ "$(snap "$R")" = "$before" ] || fail "$n: files were written"
    rm -rf "$R"
  done
done

uprepo u-dry
run "upgrade --dry-run, no file" upgrade --dry-run; ok
says "would update $SET"
[ ! -e "$R/$SET" ] || fail "$n: $SET written"

uprepo u-breaking
run "upgrade without --yes (breaking), no file" upgrade
[ "$rc" = 1 ] || fail "$n: exit $rc, want 1 (a breaking upgrade needs --yes)"
says "would update $SET"
[ ! -e "$R/$SET" ] || fail "$n: $SET written"

newrepo u-same
vl -C "$R" init --language en >/dev/null 2>&1 || fail "u-same: vloop init failed"
rm -rf "$R/.claude"
run "upgrade at the binary's version" upgrade --yes; ok
says "already at $V"
[ ! -e "$R/$SET" ] || fail "$n: $SET written"

echo "gate T1: ok"
