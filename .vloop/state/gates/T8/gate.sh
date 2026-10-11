#!/bin/sh
# Gate T8 — vloop doctor's `guidelines` check (brief L9), printed right after
# `gate edits`: rules are lines starting `- ` outside HTML comments, provenance
# a D or I ID; over 20 rules or a rule with no ID is a warning; no file is
# `- guidelines`; no exit code changes.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
GF=.vloop/repo-guidelines.md
OVER='guidelines 21 rules, over the budget of 20 — retire or merge some before adding more'
NOID='2 rule(s) cite no defect or intervention: line 5, line 6'
# doc <out> [args]: vloop doctor in $R; stdout in $t/<out>, exit in $got.
doc() { o=$1; shift; inr "$B" doctor "$@" </dev/null >"$t/$o" 2>"$t/$o.err"; got=$?; }
# after <out>: the line after the `gate edits` line.
after() { awk 'p { print; exit } /^[^ ]+ gate edits( |$)/ { p = 1 }' "$t/$1"; }
expect() {
  doc "$1"
  [ "$got" = 0 ] || fail "$1: doctor exits $got, not 0: $(cat "$t/$1" "$t/$1.err")"
  a=$(after "$1")
  [ "$a" = "$2" ] || fail "$1: after the gate edits line want '$2', got '$a' in:
$(cat "$t/$1")"
}
warnings() { sed -n 's/^doctor: 0 problem(s), \([0-9]*\) warning(s)$/\1/p' "$t/$1"; }

newrepo r
cp "$R/$GF" "$t/template.md" 2>/dev/null || : >"$t/template.md"
cp "$G/seed.md" "$R/$GF" || fail seed
expect seed '✓ guidelines 6 rule(s)'
w0=$(warnings seed)
[ -n "$w0" ] || fail "seed: no 'doctor: 0 problem(s), <n> warning(s)' line: $(cat "$t/seed")"
doc seed.json --json
[ "$("$JQ" -c '[.checks[].name] | (index("gate edits") + 1) as $i | .[$i]' "$t/seed.json")" = '"guidelines"' ] ||
  fail "--json: the check after 'gate edits' is not 'guidelines': $("$JQ" -c '[.checks[].name]' "$t/seed.json")"

# The template vloop init wrote (T1): its examples are commented out.
[ -s "$t/template.md" ] || fail "vloop init wrote no $GF"
cp "$t/template.md" "$R/$GF"
expect template '✓ guidelines 0 rule(s)'
[ "$(warnings template)" = "$w0" ] || fail "template: the warning count changed"

# A rule inside a multi-line HTML comment is not a rule.
{ cat "$G/seed.md"; printf '\n<!--\n- hidden (D20261012-0100)\n-->\n'; } >"$R/$GF"
expect hidden '✓ guidelines 6 rule(s)'

extra() { i=1; while [ $i -le 14 ]; do printf -- '- Rule %d. (D20261012-%04d)\n' $i $((100 + i)); i=$((i + 1)); done; printf -- '- Rule 15. (I20261012-0115-a-decision)\n'; }
{ cat "$G/seed.md"; extra; } >"$R/$GF"
expect over "! $OVER"
[ "$(warnings over)" = $((w0 + 1)) ] || fail "over: the warning count is not one more: $(tail -n 1 "$t/over")"

sed -e '5s/ (D[0-9-]*)$//' -e '6s/ (D[0-9-]*)$//' "$G/seed.md" >"$t/stripped"
sed -n '5,6p' "$t/stripped" | grep -q 'D2026' && fail "fixture: lines 5 and 6 still carry an ID"
cp "$t/stripped" "$R/$GF"
expect noid "! guidelines $NOID"

{ cat "$t/stripped"; extra; } >"$R/$GF"
# Both: one line, the two messages joined by `; ` — the second keeps its
# `guidelines`, with or without its own `! `.
doc both
[ "$got" = 0 ] || fail "both: doctor exits $got, not 0: $(cat "$t/both" "$t/both.err")"
a=$(after both)
case "$a" in
"! $OVER; guidelines $NOID" | "! $OVER; ! guidelines $NOID") ;;
*) fail "both: after the gate edits line want '! $OVER; guidelines $NOID', got '$a' in:
$(cat "$t/both")" ;;
esac
[ "$(warnings both)" = $((w0 + 1)) ] || fail "both: not one warning line: $(tail -n 1 "$t/both")"

rm "$R/$GF"
expect nofile '- guidelines'
echo "gate T8: ok"
