#!/bin/sh
# Gate T7 — the per-brief interventions report (brief I6, per brief) and the
# export (brief I7). Per brief: in a clone of this repository at HEAD, records
# added for B9 and for no brief; the summary's line after defects, the JSON
# interventions object (valid metrics/v2), and B1–B9's per-brief JSON otherwise
# unchanged from the base binary's output kept beside this gate (defects
# excluded: the operator may record defects mid-run). Export: each intervention
# line carries agreement and the option count, and no option text, reason,
# decision or context. Judges the built binary; reads nothing a task writes.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C TZ=UTC GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
export GIT_AUTHOR_NAME=gate GIT_AUTHOR_EMAIL=gate@example.invalid GIT_COMMITTER_NAME=gate GIT_COMMITTER_EMAIL=gate@example.invalid
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
ROOT=$(pwd -P)
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

# --- the metrics/v2 schema gains the optional interventions object ---
"$B" schema show metrics/v2 >"$t/ms.json" 2>&1 || fail "vloop schema show metrics/v2 failed"
jq -e '.properties.interventions | type == "object"' "$t/ms.json" >/dev/null 2>&1 || fail "metrics/v2 does not declare the interventions object"
jq -e '(.required // []) | index("interventions") == null' "$t/ms.json" >/dev/null 2>&1 || fail "metrics/v2 makes interventions required; it is an optional, additive key"

# --- per brief, on this repository's own runs ---
git clone -q "$ROOT" "$t/c" || fail "git clone of the repository at HEAD"
R="$t/c"
B9=B20261003-2049-gate-model.loop-brief
# B9's records in the clone, one agreement per line (no agreement line: no-options).
for f in "$R"/.vloop/interventions/I*.md; do
  awk -v b="$B9" 'NR == 1 { next } /^---$/ { exit } /^brief: / { v = $0; sub(/^brief: */, "", v); gsub(/"/, "", v); br = v } /^agreement: / { a = $2 } END { if (br == b) print (a == "" ? "no-options" : a) }' "$f"
done >"$t/b9-agreements"
count() { grep -cx "$1" "$t/b9-agreements"; }
n0=0; for a in recommended other-option adjusted different no-options; do k=$(count $a); eval "c_$(echo $a | tr - _)=$k"; n0=$((n0 + k)); done
[ "$n0" -ge 1 ] || fail "found no B9 records in the clone"
run intervention add "b9 recommended" --brief "$B9" --phase verify --kind decision --automatable partly --by both --option x --option y --recommended 1 --why w --decided-option 1 || fail "add: $(cat "$t/err")"
run intervention add "b9 other option" --brief "$B9" --phase verify --kind decision --automatable partly --by both --option x --option y --recommended 1 --why w --decided-option 2 || fail "add: $(cat "$t/err")"
run intervention add "b9 different" --brief "$B9" --phase halt --kind repair --automatable no --by both --option x --recommended 1 --why w --decided-other || fail "add: $(cat "$t/err")"
run intervention add "series-level, belongs to no brief" --phase next --kind decision --automatable no --by both --option x --recommended 1 --why w --decided-option 1 || fail "add: $(cat "$t/err")"
run intervention add "b8's, not b9's" --brief B20261002-2135-vloop-v1-security.loop-brief --phase verify --kind decision --automatable no --by both --option x --recommended 1 --why w --decided-option 1 || fail "add: $(cat "$t/err")"
T=$((n0 + 3)); RA=$((c_recommended + 1)); RB=$((c_other_option + 1)); RC=$c_adjusted; RD=$((c_different + 1)); RE=$c_no_options
run metrics "docs/briefs/$B9.md" || fail "vloop metrics for B9 failed: $(cat "$t/err")"
awk 'f { print; exit } /^ *defects /{ f = 1 }' "$t/out" | awk '{ $1 = $1; print }' >"$t/line"
[ "$(cat "$t/line")" = "interventions $T · recommended $RA · other-option $RB · adjusted $RC · different $RD · no-options $RE" ] \
  || fail "the line after defects is '$(cat "$t/line")', want 'interventions  $T · recommended $RA · other-option $RB · adjusted $RC · different $RD · no-options $RE'"
run metrics "docs/briefs/$B9.md" --json || fail "vloop metrics --json for B9 failed: $(cat "$t/err")"
cp "$t/out" "$t/m9.json"
jq -e --argjson t $T --argjson a $RA --argjson b $RB --argjson c $RC --argjson d $RD --argjson e $RE \
  '.interventions.total == $t and (.interventions.by_agreement | with_entries(.key |= gsub("_"; "-"))) == {"recommended": $a, "other-option": $b, "adjusted": $c, "different": $d, "no-options": $e}' "$t/m9.json" >/dev/null 2>&1 \
  || fail "B9's JSON interventions is $(jq -c '.interventions' "$t/m9.json" 2>/dev/null), want total $T and by_agreement recommended $RA, other-option $RB, adjusted $RC, different $RD, no-options $RE"
"$B" schema validate metrics/v2 "$t/m9.json" >"$t/v" 2>&1 || fail "B9's metrics JSON with interventions does not validate as metrics/v2: $(cat "$t/v")"
n=0
for f in "$G"/base-metrics/*.json; do
  b=$(basename "$f" .json)
  run metrics "docs/briefs/$b.md" --json || fail "vloop metrics --json for $b failed: $(cat "$t/err")"
  jq -e '.interventions | type == "object"' "$t/out" >/dev/null 2>&1 || fail "$b's metrics JSON has no interventions object"
  jq -S 'del(.interventions, .defects)' "$t/out" >"$t/got.json"
  cmp -s "$t/got.json" "$f" || fail "$b's metrics JSON changed beyond the added interventions object: $(diff "$f" "$t/got.json" | head -6 | tr '\n' '|')"
  n=$((n + 1))
done
[ "$n" -eq 9 ] || fail "compared $n briefs' metrics, want 9"

# --- the export ---
newrepo e
mkdir -p "$R/.vloop/interventions" && cp "$G"/records/I*.md "$R/.vloop/interventions/" || fail "copy the v1 fixture"
run intervention add "an export with options" --phase halt --kind repair --automatable partly --by both \
  --context 'ZQXCONTEXT the run and the task' --option 'ZQXOPTONE replace the gate' --option 'ZQXOPTTWO reset the task' \
  --recommended 1 --why 'ZQXWHY the gate is wrong' --decided-option 2 --decided 'ZQXDECIDED reset it' || fail "add: $(cat "$t/err")"
I1=$(basename "$(cat "$t/out")" .md)
run intervention add "an export without options" --phase run --kind ceremony --automatable yes --by operator || fail "add: $(cat "$t/err")"
I2=$(basename "$(cat "$t/out")" .md)
V1=I20260101-0800-a-v1-record-with-every-section
run metrics export || fail "vloop metrics export failed: $(cat "$t/err")"
cp "$t/out" "$t/export"
grep -q ZQX "$t/export" && fail "the export carries an option's text, the reason, the decision or the context: $(grep -o 'ZQX[A-Z0-9]*' "$t/export" | sort -u | tr '\n' ' ')"
# exported <id> <agreement> <options>
exported() {
  jq -e -s --arg id "$1" --arg a "$2" --argjson o "$3" '[.[] | select(.type == "intervention" and .id == $id)] | length == 1 and .[0].agreement == $a and .[0].options == $o' "$t/export" >/dev/null 2>&1 \
    || fail "the export's intervention $1 does not carry agreement $2 and options $3: $(jq -c -s --arg id "$1" '.[] | select(.id == $id)' "$t/export" 2>/dev/null)"
}
exported "$I1" other-option 2
exported "$I2" no-options 0
exported "$V1" no-options 0
i=0
while IFS= read -r l; do
  i=$((i + 1))
  printf '%s\n' "$l" >"$t/line.json"
  "$B" schema validate export/v1 "$t/line.json" >"$t/v" 2>&1 || fail "export line $i does not validate as export/v1: $(cat "$t/v")"
done <"$t/export"
[ "$i" -ge 3 ] || fail "the export has $i lines, want at least the three interventions"
echo "gate T7: ok"
