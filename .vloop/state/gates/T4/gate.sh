#!/bin/sh
# Gate T4 — V3 and V4: the brief skills, operate, the docs, the kit and the
# evals. Prose is the review's to judge; this gate holds each surface the brief
# names to the fact it must carry, and exercises what runs:
#   - create-brief's guideline rule; the four templates' optional section,
#     between the Worked example and Out of scope, and `vloop brief new` (this
#     tree's binary, English and Spanish) writing it;
#   - review-brief's Runnability check asking about the verify session; operate's
#     close step 2 telling the operator to run the lines not run, and operate at
#     most 1800 words;
#   - session.md's verify report section as verify/v2's; run.md and concepts.md
#     naming verify/v2;
#   - the Copilot kit: report.py validates each kept report against the schema
#     its own `schema` field names and says which; README and verify.sh's header
#     name verify/v2;
#   - the evals: report-validates.md's pattern; the new case
#     verify-unrunnable-line, its graders, its scaffold run in an empty
#     directory, its catalogue row.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE
export LC_ALL=C GOTOOLCHAIN=local
top=$(pwd -P)
has() { grep -qF -- "$2" "$top/$1" || fail "$1 does not $3 ('$2')"; }
EN="## Acceptance, at close (the operator's)"
ES="## Aceptación, al cierre (del operador)"

# create-brief: the rule, and the section in the four templates.
has plugin/skills/create-brief/guidelines.md 'Say which worked-example lines the verify session cannot run' 'carry the rule'
has plugin/skills/create-brief/guidelines.md "${EN#\#\# }" 'name the operator'"'"'s acceptance section'
# between <file> <heading> <before heading> <after heading>: the heading is a
# line of its own after the first and before the second.
between() {
  awk -v h="$2" -v a="$3" -v b="$4" '$0 == a && !pa {pa = NR} $0 == h && !ph {ph = NR} $0 == b && !pb {pb = NR}
    END {exit !(pa && ph && pb && pa < ph && ph < pb)}' "$1"
}
for f in plugin/skills/create-brief/templates/brief-en.md internal/brief/templates/en.md; do
  between "$top/$f" "$EN" '## Worked example' '## Out of scope' || fail "$f has no line '$EN' between '## Worked example' and '## Out of scope'"
done
for f in plugin/skills/create-brief/templates/brief-es.md internal/brief/templates/es.md; do
  between "$top/$f" "$ES" '## Ejemplo trabajado' '## Fuera de alcance' || fail "$f has no line '$ES' between '## Ejemplo trabajado' and '## Fuera de alcance'"
done

# review-brief: check 4 asks about lines the verify session cannot run.
r4=$(awk '/^4\. \*\*Runnability\.\*\*/{on=1} /^5\. \*\*/{on=0} on' "$top/plugin/skills/review-brief/SKILL.md")
[ -n "$r4" ] || fail "plugin/skills/review-brief/SKILL.md has no check '4. **Runnability.**'"
printf '%s\n' "$r4" | tr '\n' ' ' | grep -qF 'verify session' || fail "review-brief's check 4 (Runnability) does not ask about a line the verify session cannot run"
grep -qE '(^|[^a-z_])fence/' "$top/plugin/skills/review-brief/SKILL.md" && fail "review-brief cites a fence/ path"

# operate: close step 2, and the word count under its new ceiling.
s2=$(awk '/^## 5\. Closing/{on=1} on && /^2\. /{p=1} on && /^3\. /{p=0} p' "$top/plugin/skills/operate/SKILL.md" | tr '\n' ' ')
[ -n "$s2" ] || fail "plugin/skills/operate/SKILL.md has no step 2 under '## 5. Closing'"
printf '%s' "$s2" | grep -qF 'not run' || fail "operate's close step 2 does not tell the operator to run the lines the report lists as not run: $s2"
printf '%s' "$s2" | grep -qiF 'acceptance' || fail "operate's close step 2 does not name the brief's acceptance section: $s2"
w=$(wc -w <"$top/plugin/skills/operate/SKILL.md" | tr -d ' ')
[ "$w" -le 1800 ] || fail "plugin/skills/operate/SKILL.md is $w words, over 1800"

# The domain pages and the concepts guide.
S=$top/docs/domain/execution/session.md
grep -qxF '## Verify report — `verify/v2`' "$S" || fail "docs/domain/execution/session.md has no section '## Verify report — \`verify/v2\`'"
sec=$(awk '/^## Verify report/{on=1; next} on && /^## /{on=0} on' "$S" | tr '\n' ' ')
for s in complete not_run 'stopped before its last pass'; do
  printf '%s' "$sec" | grep -qF -- "$s" || fail "session.md's verify report section does not name '$s'"
done
grep -F 'reports/verify.json' "$top/docs/domain/execution/run.md" | grep -qF 'verify/v2' || fail "docs/domain/execution/run.md's reports/verify.json line does not name verify/v2"
has docs/guide/concepts.md 'verify/v2' 'name verify/v2'
grep -qF 'writes `verify/v1`' "$top/docs/guide/concepts.md" && fail "docs/guide/concepts.md still says the session writes verify/v1"

# The Copilot kit.
has tools/copilot-acceptance/README.md 'verify/v2' 'name verify/v2'
sed -n '1,/^[^#]/p' "$top/tools/copilot-acceptance/verify.sh" | grep -qF 'verify/v2' || fail "tools/copilot-acceptance/verify.sh's header comment does not name verify/v2"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -buildvcs=false -o "$t/vloop" ./cmd/vloop || fail "go build of this tree"
CK='"checked":{"example_lines":1,"sentences":1,"surfaces":1}'
k() { mkdir -p "$t/kit/.vloop/state/runs/R/$1/reports" && printf '%s\n' "$2" >"$t/kit/.vloop/state/runs/R/$1/reports/verify.json"; }
k 1-v1 "{\"schema\":\"verify/v1\",\"findings\":[],$CK,\"notes\":\"\"}"
k 2-v2 "{\"schema\":\"verify/v2\",\"complete\":true,\"findings\":[{\"kind\":\"example\",\"claim\":\"a\",\"observed\":\"b\",\"where\":\"c\"}],\"not_run\":[{\"line\":\"vloop init\",\"reason\":\"denied\"}],$CK,\"notes\":\"\"}"
k 3-bad "{\"schema\":\"verify/v2\",\"findings\":[],\"not_run\":[],$CK,\"notes\":\"\"}"
(cd "$t" && env -i PATH=/usr/bin:/bin HOME="$t" LC_ALL=C python3 -I "$top/tools/copilot-acceptance/report.py" verify "$t/kit" "$t/vloop" 0 >"$t/kit.out" 2>&1) || fail "report.py verify failed: $(cat "$t/kit.out")"
l() { grep -F 'verify.json: present' "$t/kit.out" | sed -n "$1p"; }
[ "$(grep -cF 'verify.json: present' "$t/kit.out")" = 3 ] || fail "report.py verify: want three 'verify.json: present' lines: $(cat "$t/kit.out")"
l 1 | grep -qF 'verify/v1 ok' || fail "report.py verify on a kept verify/v1 report: want 'verify/v1 ok', got: $(l 1)"
l 2 | grep -qF 'verify/v2 ok' || fail "report.py verify on a kept verify/v2 report: want 'verify/v2 ok', got: $(l 2)"
l 2 | grep -qF '1 finding(s)' || fail "report.py verify on a kept verify/v2 report: want '1 finding(s)', got: $(l 2)"
l 3 | grep -qF 'verify/v2 INVALID' || fail "report.py verify on an invalid verify/v2 report: want 'verify/v2 INVALID', got: $(l 3)"

# The evals.
E=$top/plugin/evals
grep -qxF "pattern: 'vloop schema validate verify/v2'" "$E/verify-planted-mismatches/graders/report-validates.md" ||
  fail "verify-planted-mismatches/graders/report-validates.md's pattern is not 'vloop schema validate verify/v2'"
C=$E/verify-unrunnable-line
for f in case.yaml prompt.md scaffold.sh; do [ -f "$C/$f" ] || fail "plugin/evals/verify-unrunnable-line/$f is missing"; done
grep -qxF 'name: verify-unrunnable-line' "$C/case.yaml" || fail "verify-unrunnable-line/case.yaml does not say 'name: verify-unrunnable-line'"
grep -qF 'scaffold_script: scaffold.sh' "$C/case.yaml" || fail "verify-unrunnable-line/case.yaml does not declare scaffold_script: scaffold.sh"
grep -qF '/vloop:verify' "$C/prompt.md" || fail "verify-unrunnable-line/prompt.md does not invoke /vloop:verify"
# fm <file>: a grader's frontmatter, on one line.
fm() { awk 'NR==1 && /^---/{on=1; next} on && /^---/{exit} on' "$1" | tr '\n' ' '; }
tr=0 fi=0 ll=0
for g in "$C"/graders/*.md; do
  [ -f "$g" ] || continue
  h=$(fm "$g")
  case "$h" in
    *'type: regex'*'target: trace'*) printf '%s' "$h" | grep -qF 'vloop schema validate verify/v2' && tr=1;;
  esac
  case "$h" in
    *'type: regex'*'source: file'*) printf '%s' "$h" | grep -qF 'path: .vloop/tmp/verify.json' && printf '%s' "$h" | grep -qF '"complete": *true' && fi=1;;
  esac
  case "$h" in
    *'type: llm'*'focus:'*) printf '%s' "$h" | grep -qF 'path: .vloop/tmp/verify.json' && ll=1;;
  esac
done
[ "$tr" = 1 ] || fail "verify-unrunnable-line has no regex grader on the trace for 'vloop schema validate verify/v2'"
[ "$fi" = 1 ] || fail "verify-unrunnable-line has no regex grader on the file .vloop/tmp/verify.json for '\"complete\": *true'"
[ "$ll" = 1 ] || fail "verify-unrunnable-line has no llm grader whose focus is the file .vloop/tmp/verify.json"
grep -qE '^\| `verify-unrunnable-line` \|' "$E/README.md" || fail "plugin/evals/README.md has no catalogue row for verify-unrunnable-line"
mkdir -p "$t/sc" "$t/h" || fail mkdir
(cd "$t/sc" && env -i PATH=/usr/bin:/bin:$(dirname "$(command -v git)") HOME="$t/h" LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 \
  bash "$C/scaffold.sh" >"$t/sc.out" 2>&1) || fail "verify-unrunnable-line/scaffold.sh failed in an empty directory: $(tail -5 "$t/sc.out")"
grep -rqF 'claude -p "say hi"' "$t/sc/docs/briefs" 2>/dev/null || fail "the scaffold's brief (docs/briefs/) has no worked-example line running claude -p \"say hi\""

# vloop brief new writes the section, in English and in Spanish.
mkdir -p "$t/bn" && cd "$t/bn" || fail mkdir
B() { env -i PATH=/usr/bin:/bin:$(dirname "$(command -v git)") HOME="$t/h" LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 "$@"; }
B git init -q -b main . && B git config user.name Gate && B git config user.email gate@example.invalid || fail "git init"
B "$t/vloop" init >"$t/bn.out" 2>&1 || fail "vloop init: $(cat "$t/bn.out")"
B "$t/vloop" brief new gate-en >"$t/bn.out" 2>&1 || fail "vloop brief new gate-en: $(cat "$t/bn.out")"
f=$(ls docs/briefs/*gate-en*.md 2>/dev/null | head -1)
[ -n "$f" ] && grep -qxF "$EN" "$f" || fail "vloop brief new does not write '$EN': $(cat "$t/bn.out")"
B "$t/vloop" config set language es >"$t/bn.out" 2>&1 || fail "vloop config set language es: $(cat "$t/bn.out")"
B "$t/vloop" brief new gate-es >"$t/bn.out" 2>&1 || fail "vloop brief new gate-es: $(cat "$t/bn.out")"
f=$(ls docs/briefs/*gate-es*.md 2>/dev/null | head -1)
[ -n "$f" ] && grep -qxF "$ES" "$f" || fail "vloop brief new in Spanish does not write '$ES': $(cat "$t/bn.out")"
echo "gate T4: ok"
