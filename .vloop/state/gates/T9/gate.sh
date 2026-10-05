#!/bin/sh
# Gate T9 — the index script, the guides and the domain (brief I9). The index
# script, run on the fixture records beside this gate in a temporary copy of
# its layout, prints an Agreement column (no-options for a v1 record) and fails
# on a bad by or agreement, naming the file. The guides and the domain text name
# every v2 field, flag, refusal and table the brief specifies. Whether the
# prose explains them well is the review's.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C TZ=UTC
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
ROOT=$(pwd -P)
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT

# --- the index script ---
X="$t/x"
mkdir -p "$X/tools" "$X/.vloop/interventions" || fail mkdir
cp "$ROOT/tools/interventions-index.sh" "$X/tools/" && cp "$G"/records/I*.md "$X/.vloop/interventions/" || fail "copy"
printf '%s\n' '# index' '<!-- index:begin -->' '<!-- index:end -->' >"$X/.vloop/interventions/README-interventions.md"
bash "$X/tools/interventions-index.sh" >"$t/table" 2>"$t/err" || fail "tools/interventions-index.sh failed over valid v1 and v2 records: $(cat "$t/err")"
# cell <id> <column>: the cell of the row linking <id> in the column headed <column>.
cell() {
  awk -F '|' -v id="$1" -v col="$2" '
    NR == 1 { for (i = 1; i <= NF; i++) { h = $i; gsub(/^ +| +$/, "", h); if (h == col) c = i } next }
    index($0, "[" id "]") { v = $c; gsub(/^ +| +$/, "", v); print v; exit }' "$t/table"
}
head -n 1 "$t/table" | grep -q '| *Agreement *|' || fail "the index has no Agreement column: $(head -n 1 "$t/table")"
[ "$(cell I20260928-2215-the-loop-s-install-proof-failed-in-every Agreement)" = no-options ] || fail "a v1 record's Agreement is '$(cell I20260928-2215-the-loop-s-install-proof-failed-in-every Agreement)', want no-options"
[ "$(cell I20260101-0900-v2-other-option Agreement)" = other-option ] || fail "a v2 record's Agreement is not its frontmatter's other-option"
[ "$(cell I20260101-0902-v2-adjusted Agreement)" = adjusted ] || fail "a v2 record's Agreement is not its frontmatter's adjusted"
[ "$(cell I20260101-0902-v2-adjusted Kind)" = decision ] || fail "the Kind column no longer reads"
for f in "$G"/bad/I*.md; do
  n=$(basename "$f")
  cp "$f" "$X/.vloop/interventions/$n" || fail "copy $n"
  if bash "$X/tools/interventions-index.sh" >"$t/table2" 2>"$t/err"; then fail "the index script accepts $n, whose by or agreement is outside the vocabulary"; fi
  grep -Fq "$n" "$t/err" || fail "the index script fails on $n without naming it: $(cat "$t/err")"
  rm -f "$X/.vloop/interventions/$n"
done

# --- the guides and the domain: joined into one line each, so wrapping does not matter ---
flat() { tr '\n' ' ' <"$ROOT/$1" | tr -s ' \t' '  '; }
# names <file> <phrase…>: the file mentions every phrase.
names() {
  f=$1; shift
  flat "$f" >"$t/flat"
  for w in "$@"; do grep -Fq -- "$w" "$t/flat" || fail "$f does not mention '$w'"; done
}
names docs/guide/defects.md 'intervention/v2' 'options' 'recommended' 'decided' 'adjusted' 'agreement' \
  'other-option' 'different' 'no-options' 'other' \
  '--context' '--option' '--recommended' '--why' '--decided-option' '--decided-other' '--adjusted' '--decided' \
  'vloop intervention migrate' '--dry-run' 'vloop intervention show' \
  'at most three options' 'options need --recommended and --why' '--recommended must name an option, 1 to' \
  'options need --decided-option or --decided-other' '--decided-option must name an option, 1 to' \
  '--decided-option and --decided-other exclude each other' '--adjusted needs --decided-option' \
  '--recommended, --decided-option and --decided-other need options' \
  'agreement is derived — set decided, adjusted or recommended instead' 'options are recorded with the intervention, not set'
names docs/guide/metrics.md '--interventions' '--by phase' 'share' 'other-option' 'no-options' 'automatable-yes' 'automatable-partly' '--workspace' 'interventions'
names .vloop/interventions/README-interventions.md 'intervention/v2' 'options' 'recommended' 'decided' 'adjusted' 'agreement' \
  '**Context.**' '**Options.**' '**Recommended.**' '**Suggested.**' '**Decided.**'
names docs/domain/domain-model.md "anything the operator did around a run besides testing, with the options the model proposed and the operator's decision"
flat docs/domain/domain-model.md | grep -Fq 'not yet a vloop entity' && fail "docs/domain/domain-model.md still says interventions are not yet a vloop entity"
flat docs/domain/measurement/measurement-context.md | grep -Fq 'likely next entity' && fail "docs/domain/measurement/measurement-context.md still calls interventions measurement's likely next entity"
flat docs/domain/measurement/measurement-context.md | grep -Fq 'no command or schema yet' && fail "docs/domain/measurement/measurement-context.md still says interventions have no command or schema"
echo "gate T9: ok"
