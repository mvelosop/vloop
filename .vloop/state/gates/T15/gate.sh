#!/bin/sh
# Gate T15 — the guides and the domain match the binary (brief Q5 and "The
# domain, updated"). Text checks over the documents, with the product as the
# reference where it has one: the doctor checks in the order `vloop doctor`
# prints them, the skills the plugin ships.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
export GIT_AUTHOR_NAME=gate GIT_AUTHOR_EMAIL=gate@example.invalid GIT_COMMITTER_NAME=gate GIT_COMMITTER_EMAIL=gate@example.invalid
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
top=$(pwd -P)
go build -o "$t/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
GOOS=linux go build -o "$t/vloop-linux" ./cmd/vloop || fail "GOOS=linux go build ./cmd/vloop"
GOOS=windows go build -o "$t/vloop.exe" ./cmd/vloop || fail "GOOS=windows go build ./cmd/vloop"
B="$t/vloop"
C=docs/guide/concepts.md
# section <file> <heading regex>: the lines under the first heading matching,
# subsections included, to the next heading of the same or a higher level.
# Lines inside fenced code blocks are never headings.
section() { awk -v h="$2" '
  /^```/ { fence = !fence; if (on) print; next }
  !fence && /^#+ / { lv = index($0, " ") - 1
    if (on && lv <= level) exit
    if (!on && $0 ~ h) { on = 1; level = lv; next } }
  on { print }' "$1"; }
# paragraphs <file>: one paragraph per line (blank-line separated, joined by spaces).
paragraphs() { awk 'NF { p = p (p == "" ? "" : " ") $0; next } p != "" { print p; p = "" } END { if (p != "") print p }' "$1"; }

# --- concepts: the exit-code table ---
section "$C" '^## Exit codes' >"$t/exit"
[ -s "$t/exit" ] || fail "$C has no '## Exit codes' section"
r1=$(grep '^| `1` |' "$t/exit") || fail "the exit-code table has no row for 1"
r2=$(grep '^| `2` |' "$t/exit") || fail "the exit-code table has no row for 2"
for s in "problems or failure" "failed plan session" "mid-run error"; do
  echo "$r1" | grep -Fq "$s" || fail "the exit-code table's row 1 does not say '$s': $r1"
done
echo "$r2" | grep -Fqi "usage" || fail "the exit-code table's row 2 does not say 'usage': $r2"
echo "$r2" | grep -Fqi "blocked" || fail "the exit-code table's row 2 no longer says 'blocked' for vloop run: $r2"

# --- concepts: What changed in 2.0 ---
section "$C" 'What changed in 2\.0' >"$t/changed"
[ -s "$t/changed" ] || fail "$C has no 'What changed in 2.0' section"
for s in 'state/v2' 'gate folder' 'gate review' '[[check]]' 'intervention/v2' 'migrate' 'metrics/v2' 'exit'; do
  grep -Fqi "$s" "$t/changed" || fail "the 'What changed in 2.0' section does not mention $s"
done
grep -Eqi 'usage' "$t/changed" || fail "the 'What changed in 2.0' section does not state that exit 2 now means usage only"

# --- concepts: the doctor checks, in the order doctor prints them ---
mkdir -p "$t/home" "$t/bin" "$t/r" || fail mkdir
printf '#!/bin/sh\ncase "$1" in --version) echo "2.0.0 (Claude Code)";; plugin) echo "[]";; esac\nexit 0\n' >"$t/bin/claude" && chmod +x "$t/bin/claude" || fail stub
git -C "$t/r" init -q -b main && git -C "$t/r" commit -q --allow-empty -m base || fail "git init"
(cd "$t/r" && HOME="$t/home" PATH="$t/bin:$PATH" "$B" init </dev/null >/dev/null 2>&1 && HOME="$t/home" PATH="$t/bin:$PATH" "$B" doctor) >"$t/doctor" 2>&1
CHECKS="git install config checks claude trust gate_shell plan branch plugin self-hosting"
pos=0
for c in $CHECKS; do
  name=$(echo "$c" | tr '_' ' ')
  n=$(grep -n "^[^ ]* $name\( \|\$\)" "$t/doctor" | head -1 | cut -d: -f1)
  [ -n "$n" ] && [ "$n" -gt "$pos" ] || fail "vloop doctor does not print the $name check after the previous one: $(cat "$t/doctor")"
  pos=$n
done
# From the paragraph that starts "`vloop doctor`" to the next heading, the
# checks appear in doctor's order (the install check may be called the stamp).
awk '/^`vloop doctor`/ { on = 1 } on && /^#/ { exit } on' "$C" | tr '\n' ' ' >"$t/dpara"
[ -s "$t/dpara" ] || fail "$C has no paragraph starting '\`vloop doctor\`' that says what it checks"
missing=$(awk -v list="git;stamp,install;config;checks;claude;trust;gate shell;plan;branch;plugin;self-hosting" '
  { s = tolower($0); n = split(list, cs, ";"); pos = 1
    for (i = 1; i <= n; i++) {
      m = split(cs[i], alts, ","); best = 0
      for (j = 1; j <= m; j++) { k = index(substr(s, pos), alts[j]); if (k > 0 && (best == 0 || k < best)) { best = k; bl = length(alts[j]) } }
      if (best == 0) { print cs[i]; exit }
      pos = pos + best - 1 + bl
    } }' "$t/dpara")
[ -z "$missing" ] || fail "the doctor checks in $C are not all listed in doctor's order ($CHECKS): '$missing' is missing or out of order"

# --- concepts: the plugin's skills ---
section "$C" '^## Skills' >"$t/skills"
grep -qi 'four skills' "$t/skills" && fail "the Skills section still says the plugin ships four skills"
grep -qi 'five skills' "$t/skills" || fail "the Skills section does not say the plugin ships five skills"
for d in plugin/skills/*/; do
  s=$(basename "$d")
  grep -Fq "/vloop:$s" "$t/skills" || fail "the Skills section does not name /vloop:$s"
done

# --- configuration: environment names ---
paragraphs docs/guide/configuration.md | grep -F 'VLOOP_' | grep -F '`.`' | grep -F '`-`' | grep -Fq '`_`' \
  || fail "docs/guide/configuration.md does not say that a key's environment name turns both \`.\` and \`-\` into \`_\`"

# --- metrics: the synopsis ---
awk '/^```/ { if (on) exit; on = 1; next } on' docs/guide/metrics.md >"$t/synopsis"
for s in '--workspace' '--interventions' 'vloop metrics export'; do
  grep -Fq -- "$s" "$t/synopsis" || fail "the synopsis of docs/guide/metrics.md does not show $s: $(cat "$t/synopsis")"
done

# --- the shell loop, marked legacy ---
LEG="the shell loop that built vloop's first briefs, in this repository only"
cat docs/guide/*.md | grep -Fq "$LEG" || fail "no guide marks the shell loop as 'legacy: $LEG'"
for f in docs/guide/*.md; do
  [ "$f" = docs/guide/commands.md ] && continue
  paragraphs "$f" | grep -Ei 'shell[ -]loop' | grep -vi 'legacy' | head -1 >"$t/unmarked"
  [ -s "$t/unmarked" ] && fail "$f mentions the shell loop without marking it legacy: $(cut -c1-160 "$t/unmarked")"
done

# --- the domain's CLI conventions ---
paragraphs docs/domain/platform/platform-context.md | grep -F 'Exit codes' >"$t/conv"
[ -s "$t/conv" ] || fail "platform-context.md has no paragraph stating the exit codes"
grep -Fq '`2` usage only' "$t/conv" || fail "platform-context.md does not say exit \`2\` is usage only: $(cat "$t/conv")"
grep -Fq '`vloop: `' "$t/conv" && grep -Fq 'doctor' "$t/conv" && grep -Fqi 'preflight' "$t/conv" \
  || fail "platform-context.md does not say a refusal is one 'vloop: ' line after the checks doctor or a preflight prints: $(cat "$t/conv")"
paragraphs "$C" | grep -F '`vloop: `' | grep -v 'doctor' | head -1 >"$t/unq"
[ -s "$t/unq" ] && fail "$C states the one-line 'vloop: ' rule without doctor's and the preflight's checks: $(cut -c1-160 "$t/unq")"
cd "$top" || fail cd
echo "gate T15: ok"
