#!/bin/sh
# Gate T11 — the brief's worked example, line for line, in one temporary
# repository with vloop init done: add with options and every variant, the
# refusals, set and its refusal, a hand-edited record failing list, a v1 record
# read and migrated, show's links, the interventions table and its --by task
# refusal, and the export. Ids, paths and HEAD's sha are computed, not hardcoded.
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
R="$t/r"
mkdir -p "$R" && git -C "$R" init -q -b main && git -C "$R" commit -q --allow-empty -m "the worked example's head" || fail "git init"
(cd "$R" && "$B" init </dev/null) >"$t/init.out" 2>&1 || fail "vloop init: $(cat "$t/init.out")"
# run <args>: vloop in $R; stdout in $t/out, stderr in $t/err; returns its exit code.
run() { (cd "$R" && "$B" "$@") >"$t/out" 2>"$t/err"; }
SHA=$(git -C "$R" rev-parse HEAD) || fail "rev-parse"
add() {
  run intervention add "T3's gate failed on a path typo" --brief B1 --phase halt --kind repair \
    --automatable partly --by both --trigger 'vloop run exited 2' --done 'gate replaced' \
    --context "T3 in run B20260101-0900-a; see D20260101-0900-x and commit $SHA" "$@"
}
three() {
  add --option 'replace the gate with vloop task verify' --option 'reset T3 and retry' --option 'abandon the brief' \
    --recommended 1 --why 'the gate, not the work, is wrong' "$@"
}
fmv() { awk -v k="$2" 'NR == 1 { next } /^---$/ { exit } index($0, k ": ") == 1 { print substr($0, length(k) + 3); exit }' "$1"; }
# made <label> <options> <recommended> <decided> <agreement>: the last add wrote a record with these.
made() {
  p=$(cat "$t/out")
  case "$p" in .vloop/interventions/I[0-9]*-t3-s-gate-failed-on-a-path-typo*.md) ;; *) fail "$1: add printed '$p', want .vloop/interventions/I<stamp>-t3-s-gate-failed-on-a-path-typo.md" ;; esac
  F="$R/$p"
  [ "$(fmv "$F" options)|$(fmv "$F" recommended)|$(fmv "$F" decided)|$(fmv "$F" agreement)" = "$2|$3|$4|$5" ] \
    || fail "$1: frontmatter options|recommended|decided|agreement is '$(fmv "$F" options)|$(fmv "$F" recommended)|$(fmv "$F" decided)|$(fmv "$F" agreement)', want '$2|$3|$4|$5'"
}
refuse() {
  msg=$1; shift
  before=$(ls "$R/.vloop/interventions" | wc -l)
  add "$@"; c=$?
  [ "$c" -eq 2 ] && [ "$(cat "$t/err")" = "vloop: $msg" ] || fail "add $* gave exit $c and '$(cat "$t/err")', want exit 2 and 'vloop: $msg'"
  [ "$(ls "$R/.vloop/interventions" | wc -l)" = "$before" ] || fail "a refused add wrote a record"
}

three --decided-option 1 --decided 'replaced as recommended' || fail "the first add failed: $(cat "$t/err")"
made "first" 3 1 1 recommended
FIRST=$F; FIRST_ID=$(basename "$F" .md)
three --decided-option 2 --decided 'reset instead' || fail "add --decided-option 2 failed: $(cat "$t/err")"
made "--decided-option 2" 3 1 2 other-option
three --decided-option 1 --adjusted --decided 'replaced, and more' || fail "add --adjusted failed: $(cat "$t/err")"
made "--adjusted" 3 1 1 adjusted
[ "$(fmv "$F" adjusted)" = true ] || fail "--adjusted: no adjusted: true in the frontmatter"
three --decided-other --decided 'fixed by hand' || fail "add --decided-other failed: $(cat "$t/err")"
made "--decided-other" 3 1 other different
add --decided 'replaced as recommended' || fail "add without options failed: $(cat "$t/err")"
made "no options" 0 0 '""' no-options

refuse 'at most three options' --option a --option b --option c --option d --recommended 1 --why x --decided-option 1
refuse '--recommended must name an option, 1 to 2' --option a --option b --recommended 3 --why x --decided-option 1
refuse 'options need --decided-option or --decided-other' --option a --recommended 1 --why x
refuse '--decided-option and --decided-other exclude each other' --option a --recommended 1 --why x --decided-option 1 --decided-other
refuse '--recommended, --decided-option and --decided-other need options' --decided-other

run intervention set "$FIRST_ID" decided 2 || fail "set decided 2 failed: $(cat "$t/err")"
[ "$(fmv "$FIRST" agreement)" = other-option ] || fail "after set decided 2 the file's agreement is '$(fmv "$FIRST" agreement)', want other-option"
cp "$FIRST" "$t/first"
run intervention set "$FIRST_ID" agreement recommended; c=$?
[ "$c" -eq 2 ] && [ "$(cat "$t/err")" = "vloop: agreement is derived — set decided, adjusted or recommended instead" ] || fail "set agreement gave exit $c and '$(cat "$t/err")'"
cmp -s "$FIRST" "$t/first" || fail "a refused set changed the record"

sed 's/^agreement: other-option$/agreement: different/' "$t/first" >"$FIRST"
run intervention list; c=$?
[ "$c" -eq 1 ] && grep -Fq "$(basename "$FIRST")" "$t/err" || fail "list over a hand-edited agreement gave exit $c and '$(cat "$t/err")', want exit 1 naming the file"
cp "$t/first" "$FIRST"

V1=I20260101-0700-a-v1-record-with-context-and-decided
cp "$G/records/$V1.md" "$R/.vloop/interventions/" || fail "copy the v1 fixture"
run intervention list --json || fail "list --json with a v1 record failed: $(cat "$t/err")"
jq -e --arg id $V1 '.[] | select(.id == $id) | .context == "the B1 draft, forks one and two" and .decided.text == "the operator took the narrower scope" and .done == "settled both"' "$t/out" >/dev/null 2>&1 \
  || fail "the v1 record's Context and Decided do not read into context and decided.text, or done still holds them"
awk 'n >= 2 { print } /^---$/ { n++ }' "$R/.vloop/interventions/$V1.md" >"$t/body.before"
run intervention migrate || fail "migrate failed: $(cat "$t/err")"
[ "$(cat "$t/out")" = "migrated 1 record(s)" ] || fail "migrate printed '$(cat "$t/out")', want 'migrated 1 record(s)'"
awk 'n >= 2 { print } /^---$/ { n++ }' "$R/.vloop/interventions/$V1.md" >"$t/body.after"
cmp -s "$t/body.before" "$t/body.after" || fail "migrate changed the v1 record's body"
run intervention migrate || fail "a second migrate failed: $(cat "$t/err")"
[ "$(cat "$t/out")" = "nothing to migrate" ] || fail "a second migrate printed '$(cat "$t/out")', want 'nothing to migrate'"

run intervention show "$FIRST_ID" || fail "show failed: $(cat "$t/err")"
awk 'h && /^[ \t]*$/ { exit } h { print } /^[ \t]*links:?[ \t]*$/ { h = 1 }' "$t/out" >"$t/links"
awk '$1 == "D20260101-0900-x"' "$t/links" | grep -Fq "(not found)" || fail "show's links lack 'D20260101-0900-x  (not found)'"
awk -v s="$SHA" '$1 == s' "$t/links" | grep -Fq "the worked example's head" || fail "show's links lack '<sha>  <its subject>'"
awk '$1 == "B20260101-0900-a"' "$t/links" | grep -Fq "(not found)" || fail "show's links lack 'B20260101-0900-a  (not found)'"

run metrics --interventions || fail "metrics --interventions failed: $(cat "$t/err")"
[ "$(awk '$1 == "repair" { $1 = ""; sub(/^ /, ""); print; exit }' "$t/out")" = "5 0 0% 2 1 1 1 0 5" ] || fail "the repair row is '$(awk '$1 == "repair"' "$t/out")', want 'repair 5 0 0% 2 1 1 1 0 5'"
[ "$(awk '$1 == "total" { $1 = ""; sub(/^ /, ""); print; exit }' "$t/out")" = "6 0 0% 2 1 1 2 0 5" ] || fail "the total row is '$(awk '$1 == "total"' "$t/out")', want 'total 6 0 0% 2 1 1 2 0 5'"
run metrics --interventions --by task; c=$?
[ "$c" -eq 2 ] && [ "$(cat "$t/err")" = "vloop: --by takes kind or phase with --interventions" ] || fail "--by task gave exit $c and '$(cat "$t/err")'"

run metrics export || fail "metrics export failed: $(cat "$t/err")"
jq -e -s '[.[] | select(.type == "intervention")] | length == 6 and all(.[]; (.agreement | type) == "string" and (.options | type) == "number")' "$t/out" >/dev/null 2>&1 \
  || fail "not every exported intervention carries agreement and options"
for s in 'replace the gate with vloop task verify' 'reset T3 and retry' 'abandon the brief' 'the gate, not the work, is wrong' 'T3 in run' 'replaced as recommended' 'fixed by hand' 'forks one and two'; do
  grep -Fq "$s" "$t/out" && fail "the export carries '$s'"
done
echo "gate T11: ok"
