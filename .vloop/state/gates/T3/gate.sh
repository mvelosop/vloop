#!/bin/sh
# Gate T3 — L3, first half: the ignore line that vloop init and vloop upgrade
# keep, doctor's config.local row, and vloop run's preflight refusal and
# warning. Judged through those commands on fixture repositories, with a stub
# claude first on PATH and a fake home that trusts the run repositories.
G=.vloop/state/gates/T3
. "$G/lib.sh"
L=.vloop/config.local.toml
NOTIGN="✗ config.local .vloop/config.local.toml is not git-ignored $EMDASH the next git add -A would commit it: add it to .gitignore"
TRACKED="✗ config.local .vloop/config.local.toml is tracked by git $EMDASH it is shared with everyone who clones: run git rm --cached .vloop/config.local.toml and add it to .gitignore"

# lines2 FILE: FILE holds .vloop/tmp/ and .vloop/config.local.toml, each on
# exactly one line of its own.
lines2() {
  for l in .vloop/tmp/ .vloop/config.local.toml; do
    [ "$(grep -cxF -- "$l" "$1")" = 1 ] || fail "$2: $l is not on exactly one line of .gitignore:
$(cat "$1")"
  done
}

# 1. vloop init: a fresh repository, one with a .gitignore of its own, one
#    whose .gitignore already holds both lines.
newrepo i1
runin "$R" init
wantrc "vloop init" 0
lines2 "$R/.gitignore" "vloop init"
newrepo i2
printf 'node_modules/\n' >"$R/.gitignore" && commitall "$R" ignore
runin "$R" init
wantrc "vloop init with a .gitignore" 0
lines2 "$R/.gitignore" "vloop init with a .gitignore"
[ "$(head -n 1 "$R/.gitignore")" = node_modules/ ] && [ "$(wc -l <"$R/.gitignore" | tr -d ' ')" = 3 ] ||
  fail "vloop init changed more than the two lines of .gitignore:
$(cat "$R/.gitignore")"
newrepo i3
printf '.vloop/config.local.toml\n.vloop/tmp/\n' >"$R/.gitignore" && commitall "$R" ignore
cp "$R/.gitignore" "$t/before"
runin "$R" init
wantrc "vloop init with both lines already" 0
cmp -s "$t/before" "$R/.gitignore" || fail "vloop init changed a .gitignore that held both lines:
$(cat "$R/.gitignore")"

# 2. vloop upgrade from an older stamp puts the line back, once.
newrepo u
runin "$R" init
wantrc "vloop init (for upgrade)" 0
python3 -I -c 'import json,sys
p=sys.argv[1]; d=json.load(open(p)); d["version"]="0.1.0"; json.dump(d, open(p,"w"))' "$R/.vloop/install.json" ||
  fail "rewrite the install stamp"
printf '.vloop/tmp/\n' >"$R/.gitignore"
commitall "$R" "an older setup"
runin "$R" upgrade --yes
wantrc "vloop upgrade --yes" 0
lines2 "$R/.gitignore" "vloop upgrade"
[ "$(wc -l <"$R/.gitignore" | tr -d ' ')" = 2 ] || fail "vloop upgrade changed more than the line:
$(cat "$R/.gitignore")"

# 3. doctor's config.local row.
initrepo d
D=$R
printf '.vloop/tmp/\n.vloop/config.local.toml\n' >"$D/.gitignore" && commitall "$D" ignore
runin "$D" doctor
rows config.local "$t/out" >"$t/rows"
[ ! -s "$t/rows" ] || fail "doctor with no local file printed a config.local row:
$(cat "$t/rows")"

printf '[agent]\nwork = "claude"\n\n[model]\nwork = "sonnet"\n' >"$D/$L"
runin "$D" doctor
rows config.local "$t/out" >"$t/rows"
lit "$t/w" "! config.local overrides model.work, agent.work $EMDASH on this machine only"
cmp -s "$t/rows" "$t/w" || fail "doctor with an ignored local file overriding two keys: config.local rows
$(cat "$t/rows")
want
$(cat "$t/w")
doctor printed:
$(cat "$t/out")"
awk -v w="$(cat "$t/w")" '$0 == w { print prev; exit } { prev = $0 }' "$t/out" >"$t/prev"
[ "$(cat "$t/prev")" = "✓ config" ] || fail "the config.local row does not come right after ✓ config:
$(cat "$t/out")"

: >"$D/$L"
runin "$D" doctor
rows config.local "$t/out" >"$t/rows"
lit "$t/w" "✓ config.local"
cmp -s "$t/rows" "$t/w" || fail "doctor with an empty ignored local file: config.local rows
$(cat "$t/rows")
want ✓ config.local"

printf '[agent]\nwork = "claude"\n' >"$D/$L"
printf '.vloop/tmp/\n' >"$D/.gitignore"
runin "$D" doctor
rows config.local "$t/out" >"$t/rows"
lit "$t/w" "$NOTIGN"
cmp -s "$t/rows" "$t/w" || fail "doctor with a local file git does not ignore: config.local rows
$(cat "$t/rows")
want
$(cat "$t/w")"

printf '.vloop/tmp/\n.vloop/*.local.toml\n' >"$D/.gitignore"
runin "$D" doctor
rows config.local "$t/out" >"$t/rows"
lit "$t/w" "! config.local overrides agent.work $EMDASH on this machine only"
cmp -s "$t/rows" "$t/w" || fail "doctor with the local file ignored by a pattern: config.local rows
$(cat "$t/rows")
want
$(cat "$t/w")"

printf 'agent.work = "nope"\n' >"$D/$L"
runin "$D" doctor
rows config.local "$t/out" >"$t/rows"
[ ! -s "$t/rows" ] || fail "doctor with an unreadable local file printed a config.local row:
$(cat "$t/rows")"
grep -q '^✗ config ' "$t/out" || fail "doctor with an unreadable local file has no ✗ config row:
$(cat "$t/out")"

printf '.vloop/tmp/\n.vloop/config.local.toml\n' >"$D/.gitignore"
printf '[agent]\nwork = "claude"\n' >"$D/$L"
git -C "$D" add -f "$L" .gitignore && git -C "$D" commit -q -m "share the local file" || fail commit
runin "$D" doctor
rows config.local "$t/out" >"$t/rows"
lit "$t/w" "$TRACKED"
cmp -s "$t/rows" "$t/w" || fail "doctor with a tracked local file: config.local rows
$(cat "$t/rows")
want
$(cat "$t/w")"

# Not a git repository: the ignore test is skipped.
N=$t/plain
mkdir -p "$N/.vloop" && printf '%s' "$CHECKCFG" >"$N/.vloop/config.toml" && printf '[agent]\nwork = "claude"\n' >"$N/$L" ||
  fail "the plain folder"
runin "$N" doctor
rows config.local "$t/out" >"$t/rows"
lit "$t/w" "! config.local overrides agent.work $EMDASH on this machine only"
cmp -s "$t/rows" "$t/w" || fail "doctor outside git: config.local rows
$(cat "$t/rows")
want
$(cat "$t/w")"

# 4. vloop run's preflight. Each repository is trusted, the stub claude plays
#    no session script, so a run that gets past its preflight halts in planning.
# runrepo NAME IGNORE: initrepo NAME with .gitignore IGNORE committed and the
# directory trusted; sets R.
runrepo() {
  initrepo "$1"
  printf '%s' "$2" >"$R/.gitignore" && commitall "$R" ignore
  trust "$R"
}

# Ignored and overriding: preflight passes, with the warning.
runrepo r0 '.vloop/tmp/
.vloop/config.local.toml
'
printf '[agent]\nwork = "claude"\n' >"$R/$L"
runin "$R" run "$BRIEF"
! grep -q 'preflight failed' "$t/err" || fail "vloop run with an ignored local file failed its preflight:
$(cat "$t/err")"
hasline "vloop run with an ignored local file overriding agent.work (stderr)" "$t/err" \
  "  ! config.local overrides agent.work $EMDASH on this machine only"

# Not ignored: refused by name, one problem, no session.
runrepo r1 '.vloop/tmp/
'
printf '[agent]\nwork = "claude"\n' >"$R/$L"
: >"$t/stub/claude.log"
head=$(git -C "$R" rev-parse HEAD)
runin "$R" run "$BRIEF"
wantrc "vloop run with a local file git does not ignore" 1
hasline "vloop run, not ignored (stderr)" "$t/err" "  $NOTIGN"
hasline "vloop run, not ignored (stderr)" "$t/err" "vloop: preflight failed $EMDASH 1 problem(s), nothing has run"
! grep -q '/vloop:' "$t/stub/claude.log" || fail "vloop run refused but a session ran"
[ "$(git -C "$R" rev-parse HEAD)" = "$head" ] && [ "$(git -C "$R" rev-parse --abbrev-ref HEAD)" = main ] ||
  fail "the refused run moved HEAD or the branch"

# Tracked: refused by name, one problem.
runrepo r2 '.vloop/tmp/
.vloop/config.local.toml
'
printf '[agent]\nwork = "claude"\n' >"$R/$L"
git -C "$R" add -f "$L" && git -C "$R" commit -q -m "share the local file" || fail commit
: >"$t/stub/claude.log"
runin "$R" run "$BRIEF"
wantrc "vloop run with a tracked local file" 1
hasline "vloop run, tracked (stderr)" "$t/err" "  $TRACKED"
hasline "vloop run, tracked (stderr)" "$t/err" "vloop: preflight failed $EMDASH 1 problem(s), nothing has run"
! grep -q '/vloop:' "$t/stub/claude.log" || fail "vloop run refused but a session ran"

echo "gate T3: ok"
