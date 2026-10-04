#!/bin/sh
# Gate T1 — intervention/v2: the schema, reading v1 and v2 records (the
# Context, Suggested and Decided sections no longer folded into done), writing
# v2, the derived agreement and its validation (brief I1). Judges the built
# binary in temporary repositories; reads nothing a task writes.
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

# --- the schema ---
"$B" schema list >"$t/list" 2>&1 || fail "vloop schema list: $(cat "$t/list")"
grep -Fqx intervention/v1 "$t/list" || fail "vloop schema list no longer lists intervention/v1"
grep -Fqx intervention/v2 "$t/list" || fail "vloop schema list does not list intervention/v2"
"$B" schema show intervention/v2 >"$t/s.json" 2>&1 || fail "vloop schema show intervention/v2 failed: $(cat "$t/s.json")"
jq -e 'type == "object"' "$t/s.json" >/dev/null 2>&1 || fail "vloop schema show intervention/v2 is not a JSON Schema document"
"$B" schema validate intervention/v2 "$G/json/good.json" >"$t/v" 2>&1 || fail "a record with two options, a recommendation and a decision does not validate as intervention/v2: $(cat "$t/v")"
if "$B" schema validate intervention/v2 "$G/json/four-options.json" >"$t/v" 2>&1; then fail "a record with four options validates as intervention/v2"; fi
if "$B" schema validate intervention/v2 "$G/json/bad-agreement.json" >"$t/v" 2>&1; then fail "a record with agreement \"maybe\" validates as intervention/v2"; fi

# --- reading v1 and v2 records ---
newrepo r
mkdir -p "$R/.vloop/interventions" && cp "$G"/records/I*.md "$R/.vloop/interventions/" || fail "cannot copy the fixture records"
run intervention list --json || fail "vloop intervention list --json over valid v1 and v2 records failed: $(cat "$t/err")"
cp "$t/out" "$t/l.json"
jq -e 'type == "array" and length == 6' "$t/l.json" >/dev/null 2>&1 || fail "vloop intervention list --json does not give the six fixture records: $(head -c 400 "$t/l.json")"
# has <id> <jq condition> <message>
has() { jq -e --arg id "$1" "[.[] | select(.id == \$id)] | length == 1 and (.[0] | $2)" "$t/l.json" >/dev/null 2>&1 || fail "$1: $3"; }
V1=I20260928-2215-the-loop-s-install-proof-failed-in-every
has $V1 '.agreement == "no-options"' "a v1 record does not read as agreement no-options"
has $V1 '((.options // []) | length) == 0' "a v1 record reads with options"
has $V1 '(.context | startswith("Before B1, installing the loop")) and (.context | contains("passed 46/46."))' "the v1 record's **Context.** section is not in context"
has $V1 '.suggested == "fix both scenarios to build their own fixtures, and move the warning out of the copy loop."' "the v1 record's **Suggested.** section is not in suggested"
has $V1 '.decided.text == "the operator approved: fix them in the source repo, commit, reinstall."' "the v1 record's **Decided.** section is not in decided.text"
has $V1 '.done == "fixed both in an-autonomous-loop-3 (scenarios 26 and 43, install.sh), reinstalled"' "done still holds more than the **Done.** section (Context, Suggested or Decided folded into it)"
has $V1 '.automation == "a tooling defect, fixed once; the install proof itself is the automation"' "automation is not the **What would automate it.** section"
has $V1 '.backfilled == true and .phase == "setup" and .by == "assistant"' "the v1 frontmatter no longer reads"
V1P=I20261003-1147-b8-halted-on-t12-s-gate-dispute-a-fixtur
has $V1P '.agreement == "no-options" and ((.context // "") == "") and .done == "Verified the dispute and replaced the gate with vloop task verify."' "a v1 record without the new sections does not read as before"
VO=I20260101-0900-v2-other-option
has $VO '.schema == "intervention/v2"' "a v2 record does not read as intervention/v2"
has $VO '.options == ["replace the gate with vloop task verify", "reset the task and retry"]' "the **Options.** list is not read into options"
has $VO '.recommended.option == 1 and .recommended.why == "the gate, not the work, is wrong"' "recommended is not {option: 1, why: the **Recommended.** section}"
has $VO '.decided.option == 2 and (.decided.adjusted // false) == false and .decided.text == "reset instead, to watch the work fail once more"' "decided is not {option: 2, adjusted: false, text: the **Decided.** section}"
has $VO '.agreement == "other-option"' "agreement is not other-option"
has $VO '.context == "T3 in run B20260101-0900-a" and .done == "the task was reset" and .trigger == "the gate failed on a path typo" and .automation == "a driver that resets on a known flake"' "the sections of a v2 record do not read into their own fields"
VD=I20260101-0901-v2-different
has $VD '.decided.option == 0 and .agreement == "different" and (.options | length) == 1 and .decided.text == "fixed it by hand instead"' "decided: other does not read as option 0 with agreement different"
VA=I20260101-0902-v2-adjusted
has $VA '.decided.option == 1 and .decided.adjusted == true and .recommended.option == 2 and .agreement == "adjusted" and (.options | length) == 3' "decided 1 with adjusted: true does not read as agreement adjusted"
VN=I20260101-0903-v2-no-options
has $VN '.agreement == "no-options" and ((.options // []) | length) == 0' "a v2 record without options does not read as no-options"
n=0
for id in $VO $VD $VA $VN; do
  jq --arg id "$id" '.[] | select(.id == $id)' "$t/l.json" >"$t/one.json" || fail "cannot extract $id"
  "$B" schema validate intervention/v2 "$t/one.json" >"$t/v" 2>&1 || fail "the JSON form of $id does not validate as intervention/v2: $(cat "$t/v")"
  n=$((n + 1))
done
[ "$n" -eq 4 ] || fail "validated $n v2 records, want 4"
run intervention list || fail "vloop intervention list (text) over valid v1 and v2 records failed: $(cat "$t/err")"

# --- invalid records: stored fields that disagree with the options ---
for f in "$G"/bad/I*.md; do
  n=$(basename "$f")
  cp "$f" "$R/.vloop/interventions/$n" || fail "cannot copy $n"
  run intervention list; c=$?
  [ "$c" -eq 1 ] || fail "vloop intervention list exited $c with the invalid record $n present, want 1"
  grep -Fq "$n" "$t/err" || fail "vloop intervention list fails without naming $n: $(cat "$t/err")"
  rm -f "$R/.vloop/interventions/$n"
done

# --- writing v2 ---
newrepo w
run intervention add "a plain record" --brief B1 --phase halt --kind repair --automatable partly --by both \
  --trigger 'vloop run exited 2' --done 'gate replaced' --automation 'a driver that replaces gates' || fail "vloop intervention add with v1 flags failed: $(cat "$t/err")"
p=$(cat "$t/out")
case "$p" in .vloop/interventions/I*-a-plain-record.md) ;; *) fail "vloop intervention add printed '$p', want .vloop/interventions/I<stamp>-a-plain-record.md" ;; esac
f="$R/$p"
[ -f "$f" ] || fail "vloop intervention add printed $p but wrote no such file"
awk 'NR == 1 { next } /^---$/ { exit } { print }' "$f" >"$t/fm"
awk 'f && n < 5 { print; n++ } /^by: / { f = 1 }' "$t/fm" >"$t/after-by"
printf '%s\n' 'schema: intervention/v2' 'options: 0' 'recommended: 0' 'decided: ""' 'agreement: no-options' >"$t/want"
cmp -s "$t/after-by" "$t/want" || fail "the frontmatter after by: is not schema, options, recommended, decided, agreement as intervention/v2 writes them; got: $(tr '\n' '|' <"$t/after-by")"
awk 'f { print; exit } /^agreement: / { f = 1 }' "$t/fm" | grep -q '^occurred: ' || fail "occurred: does not follow agreement: in the frontmatter"
grep -q '^adjusted:' "$t/fm" && fail "a record with nothing adjusted carries an adjusted: line"
grep -Fqx '**Trigger.** vloop run exited 2' "$f" || fail "the record has no **Trigger.** section"
grep -Fqx '**Done.** gate replaced' "$f" || fail "the record has no **Done.** section"
grep -Fqx '**What would automate it.** a driver that replaces gates' "$f" || fail "the record has no **What would automate it.** section"
for s in Context Options Recommended Suggested Decided; do
  grep -q "^\*\*$s\.\*\*" "$f" && fail "the record carries an empty **$s.** section"
done
run intervention list --json || fail "vloop intervention list --json after add failed: $(cat "$t/err")"
jq '.[0]' "$t/out" >"$t/one.json"
jq -e '.schema == "intervention/v2" and .agreement == "no-options"' "$t/one.json" >/dev/null 2>&1 || fail "the added record does not read as an intervention/v2 no-options record"
"$B" schema validate intervention/v2 "$t/one.json" >"$t/v" 2>&1 || fail "the added record's JSON form does not validate as intervention/v2: $(cat "$t/v")"
echo "gate T1: ok"
