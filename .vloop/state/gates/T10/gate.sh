#!/bin/sh
# Gate T10 — this repository's records migrated (the brief's real-data check).
# The 77 records at the base (listed beside this gate) are, in the working
# tree, byte for byte what the built `vloop intervention migrate` makes of
# HEAD's copies; --dry-run on HEAD's copies lists exactly the v1 ones; every
# record with a Context section reads it into context and out of done; the
# interventions table equals the frontmatter's own counts; the index is fresh.
# Reads the records as data under test; its judges are the binary's own
# migration of HEAD's copies and counts computed here.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C TZ=UTC GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
export GIT_AUTHOR_NAME=gate GIT_AUTHOR_EMAIL=gate@example.invalid GIT_COMMITTER_NAME=gate GIT_COMMITTER_EMAIL=gate@example.invalid
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
ROOT=$(pwd -P)
D="$ROOT/.vloop/interventions"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -o "$t/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
GOOS=linux go build -o "$t/vloop-linux" ./cmd/vloop || fail "GOOS=linux go build ./cmd/vloop"
GOOS=windows go build -o "$t/vloop.exe" ./cmd/vloop || fail "GOOS=windows go build ./cmd/vloop"
B="$t/vloop"

# --- HEAD's copies, migrated by the binary, equal the working tree ---
H="$t/h"
mkdir -p "$H" && git -C "$H" init -q -b main && git -C "$H" commit -q --allow-empty -m "gate base" || fail "git init"
(cd "$H" && "$B" init </dev/null) >"$t/init.out" 2>&1 || fail "vloop init: $(cat "$t/init.out")"
git -C "$ROOT" archive HEAD .vloop/interventions | tar -x -C "$H" || fail "cannot extract HEAD's .vloop/interventions"
: >"$t/v1"
while IFS= read -r f; do
  [ -f "$H/.vloop/interventions/$f" ] || fail "HEAD has no .vloop/interventions/$f, a record at the base"
  [ -f "$D/$f" ] || fail "the working tree has no .vloop/interventions/$f, a record at the base"
  awk 'NR == 1 { next } /^---$/ { exit } $0 == "schema: intervention/v2" { v2 = 1 } END { exit v2 }' "$H/.vloop/interventions/$f" && echo "$f" >>"$t/v1"
done <"$G/base-records.txt"
(cd "$H" && "$B" intervention migrate --dry-run) >"$t/dry" 2>"$t/err" || fail "vloop intervention migrate --dry-run on HEAD's records failed: $(cat "$t/err")"
while IFS= read -r f; do grep -Fq "$f" "$t/dry" || fail "migrate --dry-run does not list $f, a v1 record at HEAD"; done <"$t/v1"
while IFS= read -r f; do grep -Fxq "$f" "$t/v1" || { grep -Fq "$f" "$t/dry" && fail "migrate --dry-run lists $f, which is already v2 at HEAD"; }; done <"$G/base-records.txt"
(cd "$H" && "$B" intervention migrate) >"$t/mig" 2>"$t/err" || fail "vloop intervention migrate on HEAD's records failed: $(cat "$t/err")"
n=0
while IFS= read -r f; do
  cmp -s "$H/.vloop/interventions/$f" "$D/$f" || fail ".vloop/interventions/$f is not what vloop intervention migrate makes of HEAD's copy: $(diff "$H/.vloop/interventions/$f" "$D/$f" | head -6 | tr '\n' '|')"
  awk 'NR == 1 { next } /^---$/ { exit } $0 == "schema: intervention/v2" { v2 = 1 } END { exit !v2 }' "$D/$f" || fail ".vloop/interventions/$f is not intervention/v2"
  n=$((n + 1))
done <"$G/base-records.txt"
[ "$n" -eq 77 ] || fail "checked $n base records, want 77"

# --- the folded sections now read into their own fields ---
(cd "$ROOT" && "$B" intervention list --json) >"$t/l.json" 2>"$t/err" || fail "vloop intervention list --json over this repository failed: $(cat "$t/err")"
k=0
for f in "$D"/I*.md; do
  grep -q '^\*\*Context\.\*\*' "$f" || continue
  id=$(basename "$f" .md)
  jq -e --arg id "$id" '[.[] | select(.id == $id)] | length == 1 and ((.[0].context // "") | length) > 0 and ((.[0].done // "") | contains("**Context.**") | not)' "$t/l.json" >/dev/null 2>&1 \
    || fail "$id has a **Context.** section but list --json gives it no context, or still folds it into done"
  k=$((k + 1))
done
[ "$k" -ge 66 ] || fail "found $k records with a Context section, want the 66 the brief counts"
i=0
while [ "$i" -lt "$(jq length "$t/l.json")" ]; do
  jq ".[$i]" "$t/l.json" >"$t/one.json"
  "$B" schema validate intervention/v2 "$t/one.json" >"$t/v" 2>&1 || fail "record $(jq -r .id "$t/one.json") does not validate as intervention/v2: $(cat "$t/v")"
  i=$((i + 1))
done

# --- the interventions table equals the frontmatter's counts ---
for f in "$D"/I*.md; do
  awk 'NR == 1 { next } /^---$/ { exit } { k = $0; sub(/:.*/, "", k); v = $0; sub(/^[^:]*: */, "", v); a[k] = v } END { print a["kind"], a["phase"], a["automatable"], (a["agreement"] == "" ? "no-options" : a["agreement"]) }' "$f"
done >"$t/fm"
N=$(wc -l <"$t/fm" | tr -d ' ')
# check <table file> <field 1|2 of $t/fm> <name>
check() {
  want=$(awk -v c="$2" -v k="$3" '$c == k { n++; if ($4 == "no-options") no++; if ($3 == "yes") y++; if ($3 == "partly") p++ } END { print n + 0, no + 0, y + 0, p + 0 }' "$t/fm")
  got=$(awk -v k="$3" 'NR > 1 && $1 == k { print $2, $8, $9, $10; exit }' "$1")
  [ "$got" = "$want" ] || fail "the $3 row's n, no-options, automatable-yes and automatable-partly are '$got', want '$want' from the records"
}
(cd "$ROOT" && "$B" metrics --interventions) >"$t/kind" 2>"$t/err" || fail "vloop metrics --interventions failed: $(cat "$t/err")"
for k in $(awk '{ print $1 }' "$t/fm" | sort -u); do check "$t/kind" 1 "$k"; done
[ "$(awk 'NR > 1 && $1 == "total" { print $2 }' "$t/kind")" = "$N" ] || fail "the total row's n is not the $N records"
(cd "$ROOT" && "$B" metrics --interventions --by phase) >"$t/phase" 2>"$t/err" || fail "vloop metrics --interventions --by phase failed: $(cat "$t/err")"
for p in $(awk '{ print $2 }' "$t/fm" | sort -u); do check "$t/phase" 2 "$p"; done

# --- the index is regenerated ---
(cd "$ROOT" && bash tools/interventions-index.sh --check) >"$t/idx" 2>&1 || fail "the interventions index is stale: $(cat "$t/idx")"
grep -q '| *Agreement *|' "$D/README-interventions.md" || fail "the committed index has no Agreement column"
echo "gate T10: ok"
