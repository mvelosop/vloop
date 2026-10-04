#!/bin/sh
# Gate T4 — `vloop intervention migrate` (brief I4): every v1 record gains the
# five v2 frontmatter lines after by: and not one other byte changes; a v2
# record is left alone; --dry-run lists and writes nothing; a second run says
# nothing to migrate; `vloop upgrade` rewrites no record. Judges the built
# binary in temporary repositories over the fixture records beside this gate.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C TZ=UTC GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
export GIT_AUTHOR_NAME=gate GIT_AUTHOR_EMAIL=gate@example.invalid GIT_COMMITTER_NAME=gate GIT_COMMITTER_EMAIL=gate@example.invalid
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -o "$t/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
GOOS=linux go build -o "$t/vloop-linux" ./cmd/vloop || fail "GOOS=linux go build ./cmd/vloop"
GOOS=windows go build -o "$t/vloop.exe" ./cmd/vloop || fail "GOOS=windows go build ./cmd/vloop"
B="$t/vloop"
# newrepo <name>: a git repository under $t with vloop init done; sets R.
newrepo() {
  R="$t/$1"
  mkdir -p "$R" && git -C "$R" init -q -b main && git -C "$R" commit -q --allow-empty -m "gate base" || fail "git init $1"
  (cd "$R" && "$B" init </dev/null) >"$t/init.out" 2>&1 || fail "vloop init in $1: $(cat "$t/init.out")"
}
# run <args>: vloop in $R; stdout in $t/out, stderr in $t/err; returns its exit code.
run() { (cd "$R" && "$B" "$@") >"$t/out" 2>"$t/err"; }
V1A=I20260928-2215-the-loop-s-install-proof-failed-in-every.md
V1B=I20261003-1147-b8-halted-on-t12-s-gate-dispute-a-fixtur.md
V2=I20260101-0900-v2-other-option.md
# migrated <file>: the fixture with the five v2 lines inserted after its by: line.
migrated() {
  awk 'NR > 1 && /^---$/ { fm++ } { print } fm == 0 && /^by: / { print "schema: intervention/v2"; print "options: 0"; print "recommended: 0"; print "decided: \"\""; print "agreement: no-options" }' "$1"
}
# snap: a checksum of every file under .vloop/interventions/.
snap() { (cd "$R/.vloop/interventions" && cksum I*.md) 2>/dev/null; }

# --- vloop upgrade rewrites no record ---
newrepo u
mkdir -p "$R/.vloop/interventions" && cp "$G/records/$V1A" "$R/.vloop/interventions/" || fail "copy"
git -C "$R" add -A && git -C "$R" commit -q -m "a v1 record" || fail "commit the v1 record"
run upgrade </dev/null || fail "vloop upgrade failed in a freshly initialised repository: $(cat "$t/err")"
cmp -s "$G/records/$V1A" "$R/.vloop/interventions/$V1A" || fail "vloop upgrade rewrote a v1 intervention record"

# --- migrate ---
newrepo m
mkdir -p "$R/.vloop/interventions" && cp "$G"/records/I*.md "$R/.vloop/interventions/" || fail "copy"
s0=$(snap)
run intervention migrate --dry-run || fail "vloop intervention migrate --dry-run failed: $(cat "$t/err")"
grep -Fq "$V1A" "$t/out" || fail "migrate --dry-run does not list the v1 record $V1A: $(cat "$t/out")"
grep -Fq "$V1B" "$t/out" || fail "migrate --dry-run does not list the v1 record $V1B: $(cat "$t/out")"
grep -Fq "$V2" "$t/out" && fail "migrate --dry-run lists the v2 record $V2, which has nothing to migrate"
[ "$(snap)" = "$s0" ] || fail "migrate --dry-run wrote to a record"
run intervention migrate || fail "vloop intervention migrate failed: $(cat "$t/err")"
[ "$(cat "$t/out")" = "migrated 2 record(s)" ] || fail "migrate printed '$(cat "$t/out")', want 'migrated 2 record(s)'"
for f in $V1A $V1B; do
  migrated "$G/records/$f" >"$t/want"
  cmp -s "$t/want" "$R/.vloop/interventions/$f" || fail "the migrated $f is not the v1 file with schema, options, recommended, decided and agreement inserted after by: and nothing else changed: $(diff "$t/want" "$R/.vloop/interventions/$f" | head -8 | tr '\n' '|')"
done
cmp -s "$G/records/$V2" "$R/.vloop/interventions/$V2" || fail "migrate changed a record that was already v2"
s1=$(snap)
run intervention migrate || fail "a second vloop intervention migrate failed: $(cat "$t/err")"
[ "$(cat "$t/out")" = "nothing to migrate" ] || fail "a second migrate printed '$(cat "$t/out")', want 'nothing to migrate'"
[ "$(snap)" = "$s1" ] || fail "a second migrate changed a record"
run intervention list --json || fail "vloop intervention list --json after migrate failed: $(cat "$t/err")"
cp "$t/out" "$t/l.json"
for f in $V1A $V1B; do
  id=${f%.md}
  jq -e --arg id "$id" '[.[] | select(.id == $id)] | length == 1 and (.[0] | .schema == "intervention/v2" and .agreement == "no-options")' "$t/l.json" >/dev/null 2>&1 \
    || fail "$id does not read as an intervention/v2 no-options record after migrate"
  jq --arg id "$id" '.[] | select(.id == $id)' "$t/l.json" >"$t/one.json"
  "$B" schema validate intervention/v2 "$t/one.json" >"$t/v" 2>&1 || fail "the migrated $id does not validate as intervention/v2: $(cat "$t/v")"
done
jq -e '[.[] | select(.id == "I20260928-2215-the-loop-s-install-proof-failed-in-every")][0] | (.context | startswith("Before B1")) and .backfilled == true and .recorded == "2026-09-30T16:54:00Z"' "$t/l.json" >/dev/null 2>&1 \
  || fail "the migrated record lost its context, backfilled or recorded"

# --- a repository with nothing to migrate ---
newrepo e
run intervention migrate || fail "vloop intervention migrate with no records failed: $(cat "$t/err")"
[ "$(cat "$t/out")" = "nothing to migrate" ] || fail "migrate with no records printed '$(cat "$t/out")', want 'nothing to migrate'"
echo "gate T4: ok"
