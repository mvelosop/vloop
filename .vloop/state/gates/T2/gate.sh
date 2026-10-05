#!/bin/sh
# Gate T2 — `vloop intervention add` with options (brief I2 and the worked
# example): the written record, the derived agreement, and every refusal with
# its exact message, exit 2 and nothing written. Judges the built binary in a
# temporary repository; reads nothing a task writes.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C TZ=UTC GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
export GIT_AUTHOR_NAME=gate GIT_AUTHOR_EMAIL=gate@example.invalid GIT_COMMITTER_NAME=gate GIT_COMMITTER_EMAIL=gate@example.invalid
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

newrepo r
SHA=$(git -C "$R" rev-parse HEAD) || fail "rev-parse"
CTX="T3 in run B20260101-0900-a; see D20260101-0900-x and commit $SHA"
# add <extra flags>: the worked example's add, with its v1 flags and context.
add() {
  run intervention add "T3's gate failed on a path typo" --brief B1 --phase halt --kind repair \
    --automatable partly --by both --trigger 'vloop run exited 2' --done 'gate replaced' \
    --automation 'a driver that replaces a gate on a typo' --context "$CTX" "$@"
}
count() { find "$R/.vloop/interventions" -name 'I*.md' 2>/dev/null | wc -l | tr -d ' '; }
# added <label>: the last add succeeded; sets F to the record's file and FM to its frontmatter.
added() {
  p=$(cat "$t/out")
  case "$p" in .vloop/interventions/I[0-9]*-t3-s-gate-failed-on-a-path-typo*.md) ;; *) fail "$1: vloop intervention add printed '$p', want .vloop/interventions/I<stamp>-t3-s-gate-failed-on-a-path-typo.md" ;; esac
  F="$R/$p"; ID=$(basename "$p" .md)
  [ -f "$F" ] || fail "$1: no file $p"
  FM="$t/fm"; awk 'NR == 1 { next } /^---$/ { exit } { print }' "$F" >"$FM"
}
# after_by <label> <lines…>: the frontmatter lines right after by: are exactly these.
after_by() {
  l=$1; shift
  printf '%s\n' "$@" >"$t/want"
  awk -v n=$# 'f && k < n { print; k++ } /^by: / { f = 1 }' "$FM" >"$t/got"
  cmp -s "$t/got" "$t/want" || fail "$l: the frontmatter after by: is '$(tr '\n' '|' <"$t/got")', want '$(tr '\n' '|' <"$t/want")'"
}
# json <label> <jq condition>: the record's JSON form from list --json.
json() {
  run intervention list --json || fail "$1: vloop intervention list --json failed: $(cat "$t/err")"
  jq -e --arg id "$ID" "[.[] | select(.id == \$id)] | length == 1 and (.[0] | $2)" "$t/out" >/dev/null 2>&1 || fail "$1: the JSON form of $ID does not satisfy $2"
}

# --- the worked example: three options, the recommended one decided ---
add --option 'replace the gate with vloop task verify' --option 'reset T3 and retry' --option 'abandon the brief' \
  --recommended 1 --why 'the gate, not the work, is wrong' --decided-option 1 --decided 'replaced as recommended' \
  || fail "the worked example's add exited non-zero: $(cat "$t/err")"
added "recommended"
FIRST=$ID
after_by "recommended" 'schema: intervention/v2' 'options: 3' 'recommended: 1' 'decided: 1' 'agreement: recommended'
grep -q '^adjusted:' "$FM" && fail "recommended: a record with nothing adjusted carries an adjusted: line"
grep -o '^\*\*[A-Za-z ]*\.\*\*' "$F" | tr '\n' '|' >"$t/order"
[ "$(cat "$t/order")" = '**Trigger.**|**Done.**|**Context.**|**Options.**|**Recommended.**|**Decided.**|**What would automate it.**|' ] \
  || fail "the body's sections are '$(cat "$t/order")', want Trigger, Done, Context, Options, Recommended, Decided, What would automate it"
awk '/^\*\*Options\.\*\*/ { f = 1; next } /^\*\*Recommended\.\*\*/ { f = 0 } f && /^[0-9]+\. / { print }' "$F" >"$t/opts"
printf '%s\n' '1. replace the gate with vloop task verify' '2. reset T3 and retry' '3. abandon the brief' >"$t/want"
cmp -s "$t/opts" "$t/want" || fail "the **Options.** section is not the numbered list 1. to 3.: '$(tr '\n' '|' <"$t/opts")'"
grep -Fqx '**Recommended.** the gate, not the work, is wrong' "$F" || fail "the **Recommended.** section does not hold the reason"
grep -Fqx '**Decided.** replaced as recommended' "$F" || fail "the **Decided.** section does not hold the decision"
grep -Fqx "**Context.** $CTX" "$F" || fail "the **Context.** section does not hold the context"
json "recommended" '.agreement == "recommended" and (.options | length) == 3 and .recommended.option == 1 and .recommended.why == "the gate, not the work, is wrong" and .decided.option == 1 and (.decided.adjusted // false) == false and .decided.text == "replaced as recommended" and (.context | contains("D20260101-0900-x"))'
jq --arg id "$ID" '.[] | select(.id == $id)' "$t/out" >"$t/one.json"
"$B" schema validate intervention/v2 "$t/one.json" >"$t/v" 2>&1 || fail "the worked example's record does not validate as intervention/v2: $(cat "$t/v")"

# --- the same, another option decided ---
add --option 'replace the gate with vloop task verify' --option 'reset T3 and retry' --option 'abandon the brief' \
  --recommended 1 --why 'the gate, not the work, is wrong' --decided-option 2 --decided 'reset instead' \
  || fail "add with --decided-option 2 exited non-zero: $(cat "$t/err")"
added "other-option"
after_by "other-option" 'schema: intervention/v2' 'options: 3' 'recommended: 1' 'decided: 2' 'agreement: other-option'
json "other-option" '.agreement == "other-option" and .decided.option == 2'

# --- the recommended option, adjusted ---
add --option 'replace the gate with vloop task verify' --option 'reset T3 and retry' --option 'abandon the brief' \
  --recommended 1 --why 'the gate, not the work, is wrong' --decided-option 1 --adjusted --decided 'replaced, and the path fixed too' \
  || fail "add with --decided-option 1 --adjusted exited non-zero: $(cat "$t/err")"
added "adjusted"
after_by "adjusted" 'schema: intervention/v2' 'options: 3' 'recommended: 1' 'decided: 1' 'adjusted: true' 'agreement: adjusted'
json "adjusted" '.agreement == "adjusted" and .decided.option == 1 and .decided.adjusted == true'

# --- something else decided ---
add --option 'replace the gate with vloop task verify' --option 'reset T3 and retry' --option 'abandon the brief' \
  --recommended 1 --why 'the gate, not the work, is wrong' --decided-other --decided 'fixed by hand' \
  || fail "add with --decided-other exited non-zero: $(cat "$t/err")"
added "different"
after_by "different" 'schema: intervention/v2' 'options: 3' 'recommended: 1' 'decided: other' 'agreement: different'
json "different" '.agreement == "different" and .decided.option == 0 and .decided.text == "fixed by hand"'

# --- fewer options are recorded as they are, never padded ---
add --option 'reset T3 and retry' --recommended 1 --why 'one real option only' --decided-option 1 \
  || fail "add with one option exited non-zero: $(cat "$t/err")"
added "one option"
after_by "one option" 'schema: intervention/v2' 'options: 1' 'recommended: 1' 'decided: 1' 'agreement: recommended'
awk '/^\*\*Options\.\*\*/ { f = 1; next } /^\*\*Recommended\.\*\*/ { f = 0 } f && /^[0-9]+\. / { print }' "$F" >"$t/opts"
[ "$(cat "$t/opts")" = '1. reset T3 and retry' ] || fail "a one-option record's list is '$(tr '\n' '|' <"$t/opts")', want the one option"

# --- no options ---
add || fail "add without options exited non-zero: $(cat "$t/err")"
added "no-options"
after_by "no-options" 'schema: intervention/v2' 'options: 0' 'recommended: 0' 'decided: ""' 'agreement: no-options'
grep -q '^\*\*Options\.\*\*' "$F" && fail "a record without options carries an **Options.** section"
json "no-options" '.agreement == "no-options" and ((.options // []) | length) == 0'

# --- refusals: exit 2, the exact message on stderr, nothing written ---
# refuse <message> <extra add flags…>
refuse() {
  msg=$1; shift
  before=$(count)
  add "$@"; c=$?
  [ "$c" -eq 2 ] || fail "add $* exited $c, want 2 (stderr: $(cat "$t/err"))"
  [ "$(cat "$t/err")" = "vloop: $msg" ] || fail "add $* printed '$(cat "$t/err")' on stderr, want 'vloop: $msg'"
  [ "$(count)" = "$before" ] || fail "add $* was refused but wrote a record"
}
refuse 'at most three options' --option a --option b --option c --option d --recommended 1 --why x --decided-option 1
refuse 'options need --recommended and --why' --option a --option b --recommended 1 --decided-option 1
refuse 'options need --recommended and --why' --option a --option b --why x --decided-option 1
refuse '--recommended must name an option, 1 to 2' --option a --option b --recommended 3 --why x --decided-option 1
refuse '--recommended must name an option, 1 to 3' --option a --option b --option c --recommended 4 --why x --decided-option 1
refuse 'options need --decided-option or --decided-other' --option a --recommended 1 --why x
refuse '--decided-option must name an option, 1 to 2' --option a --option b --recommended 1 --why x --decided-option 3
refuse '--decided-option and --decided-other exclude each other' --option a --recommended 1 --why x --decided-option 1 --decided-other
refuse '--adjusted needs --decided-option' --option a --option b --recommended 1 --why x --decided-other --adjusted
refuse '--recommended, --decided-option and --decided-other need options' --decided-other
refuse '--recommended, --decided-option and --decided-other need options' --recommended 1 --why x
refuse '--recommended, --decided-option and --decided-other need options' --decided-option 1

run intervention list || fail "vloop intervention list fails over the records add wrote: $(cat "$t/err")"
[ -n "$FIRST" ] || fail "no first record"
echo "gate T2: ok"
