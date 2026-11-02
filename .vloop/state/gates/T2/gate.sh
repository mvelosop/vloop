# T2 (K2): the test that holds each plugin template to the binary's. Judged by
# planting the brief's failures in copies of the tree. K2 names neither the
# test nor its package, so the gate pins neither: it runs every test function
# the tree has that the base did not (found with `go test -list`, in any
# package), and the template test is whichever of them reports K2's messages.
# Each plant outside an executive summary must fail with the brief's message;
# on the unplanted copy and with a change inside an executive summary, every
# test that reported a K2 message passes and none reports one.
G=.vloop/state/gates/T2
. "$G/lib.sh"
cd "$top" || fail cd

EN=plugin/skills/create-brief/templates/brief-en.md
ES=plugin/skills/create-brief/templates/brief-es.md
[ -f "$EN" ] && [ -f "$ES" ] || fail "the plugin templates are not in place (T1)"
MEN="plugin/skills/create-brief/templates/brief-en.md differs from internal/brief/templates/en.md outside its executive summary"
MES="plugin/skills/create-brief/templates/brief-es.md differs from internal/brief/templates/es.md outside its executive summary"
NES="plugin/skills/create-brief/templates/brief-es.md has no executive summary"

srccopy
mkdir -p "$t/keep" "$t/base" || fail mkdir
for f in "$EN" "$ES" internal/brief/templates/en.md internal/brief/templates/es.md; do
  cp "$t/src/$f" "$t/keep/$(echo "$f" | tr / _)" || fail keep
done
restore() {
  for f in "$EN" "$ES" internal/brief/templates/en.md internal/brief/templates/es.md; do
    cp "$t/keep/$(echo "$f" | tr / _)" "$t/src/$f" || fail restore
  done
}

# tests <dir> <out>: "<package> <test>" for every top-level test function of
# the module in <dir>, as `go test -list` reports them.
tests() {
  (cd "$1" && go test -list . ./... >"$t/list.out" 2>&1) || fail "go test -list in $1:
$(tail -30 "$t/list.out")"
  awk '/^Test/ { n[++k] = $1; next }
       /^ok[ \t]/ { for (i = 1; i <= k; i++) print $2 " " n[i]; k = 0; next }
       { k = 0 }' "$t/list.out" | sort -u >"$2"
}
git -C "$top" archive "$BASE" | tar -xf - -C "$t/base" || fail "the base's tree"
tests "$t/base" "$t/tests.base"
tests "$t/src" "$t/tests.src"
comm -13 "$t/tests.base" "$t/tests.src" >"$t/tests.new"
[ -s "$t/tests.new" ] || fail "the tree has no test function the base does not: no template test"

# runtests <label>: the new tests of the copy, package by package, as JSON;
# $rc is non-zero when any failed, the stream is $t/<label>.json.
runtests() {
  : >"$t/$1.json"
  rc=0
  for p in $(cut -d' ' -f1 "$t/tests.new" | sort -u); do
    re=$(awk -v p="$p" '$1 == p { s = s (s == "" ? "" : "|") $2 } END { print s }' "$t/tests.new")
    (cd "$t/src" && go test -count=1 -json -run "^($re)\$" "$p" >>"$t/$1.json" 2>&1) || rc=1
  done
}
# plant <file> <command>: <file> of the copy rewritten by <command> (reading
# stdin), which must change it.
plant() {
  sh -c "$2" <"$t/src/$1" >"$t/planted" || fail "plant in $1"
  cmp -s "$t/planted" "$t/src/$1" && fail "the plant in $1 changed nothing"
  cp "$t/planted" "$t/src/$1" || fail "plant in $1"
}
# planted <label> <what> <message>: the run failed and some test said <message>.
planted() {
  [ "$rc" != 0 ] || fail "$2: the tests new since the base all pass"
  grep -F "$3" "$t/$1.json" | grep -qF '"Action":"output"' || fail "$2: no test new since the base says '$3':
$(grep -F '"Action":"output"' "$t/$1.json" | sed -En 's/.*"Output":"(([^"\\]|\\.)*)".*/\1/p' | tail -30)"
  : >"$t/said"
  grep -F "$3" "$t/$1.json" | grep -F '"Action":"output"' |
    sed -n 's/.*"Package":"\([^"]*\)","Test":"\([^"/]*\).*/\1 \2/p' | sort -u >"$t/said"
  ok=
  while read -r p n; do
    grep -qF "\"Action\":\"fail\",\"Package\":\"$p\",\"Test\":\"$n\"," "$t/$1.json" && ok=1
  done <"$t/said"
  [ -n "$ok" ] || fail "$2: the tests that say '$3' do not fail"
  cat "$t/said" >>"$t/judges"
}
: >"$t/judges"

# A heading of brief-en.md changed outside the executive summary.
plant "$EN" "sed 's/^## Out of scope\$/## Out of scope, planted/'"
runtests en
planted en "a heading of brief-en.md changed" "$MEN"
restore

# brief-es.md without its "## Resumen ejecutivo" heading.
plant "$ES" "grep -v '^## Resumen ejecutivo\$'"
runtests nosum
planted nosum "brief-es.md without ## Resumen ejecutivo" "$NES"
restore

# The binary's Spanish template changed: the plugin's no longer matches it.
plant internal/brief/templates/es.md "sed 's/^## Fuera de alcance\$/## Fuera de alcance, plantado/'"
runtests bin
planted bin "a heading of internal/brief/templates/es.md changed" "$MES"
restore
sort -u "$t/judges" >"$t/judges.u"

# passes <label> <what>: every test that reported a K2 message passed in run
# <label>, and no test reported one.
passes() {
  while read -r p n; do
    grep -qF "\"Action\":\"pass\",\"Package\":\"$p\",\"Test\":\"$n\"," "$t/$1.json" ||
      fail "$2: the template test $n ($p) does not pass:
$(grep -F "\"Test\":\"$n" "$t/$1.json" | grep -F '"Action":"output"' | sed -En 's/.*"Output":"(([^"\\]|\\.)*)".*/\1/p' | tail -30)"
  done <"$t/judges.u"
  for m in "outside its executive summary" "has no executive summary"; do
    grep -F "$m" "$t/$1.json" | grep -qF '"Action":"output"' && fail "$2: a test reported '$m':
$(grep -F "$m" "$t/$1.json" | sed -En 's/.*"Output":"(([^"\\]|\\.)*)".*/\1/p' | head -5)"
  done
  return 0
}

# The tree as it is: the templates match.
runtests clean
passes clean "the unplanted tree"

# A line added inside brief-en.md's executive summary is not a difference.
plant "$EN" "awk '{ print } /^## Executive summary\$/ { print \"\"; print \"A line planted inside the executive summary.\" }'"
runtests inside
passes inside "a line added inside brief-en.md's executive summary"
restore
echo "T2 gate: ok"
