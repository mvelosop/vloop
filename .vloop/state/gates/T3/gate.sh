#!/bin/sh
# Gate T3 — P2, both agents in the docs. Init's closing lines byte for byte
# and the CLAUDE.md section's sentence, from `vloop init` built from this tree
# on a fresh repository; then the facts P2 lists in README.md and
# docs/guide/concepts.md. Prose is joined and its whitespace collapsed before
# it is searched, so an edit wrapped in the file's style passes; the gate
# checks the commands and names P2 requires, not wording, except the two
# sentences the brief gives verbatim.
G=.vloop/state/gates/T3
. "$G/lib.sh"

# ---- init's closing lines and the CLAUDE.md section ------------------------
newrepo r
vl -C "$R" init --language en >"$t/out" 2>"$t/err" || fail "vloop init failed: $(cat "$t/err")"
sed -n '/^next:$/,$p' "$t/out" >"$t/next"
cmp -s "$t/next" "$G/next.want" || fail "init's closing lines are not, byte for byte, $G/next.want:
$(cat "$t/next")"
want=$(flat "$G/claude-md.want" | sed 's/ $//')
flat "$R/CLAUDE.md" >"$t/claude"
grep -Fq -- "$want" "$t/claude" || fail "the CLAUDE.md section init writes lacks the sentence in $G/claude-md.want"
! grep -Fq 'give the user the two ways to load it' "$t/claude" || fail "the CLAUDE.md section still says 'the two ways to load it'"

# ---- helpers over the docs ------------------------------------------------
# sect FILE TITLE: the "## TITLE" section of FILE, heading included.
sect() { awk -v h="## $2" '$0 == h {on = 1; print; next} on && /^## / {exit} on {print}' "$1"; }
# para FILE START: the first paragraph of FILE that begins with START.
para() { awk -v RS= -v s="$2" 'index($0, s) == 1 {print; exit}' "$1"; }
# has NAME FILE TEXT: FILE, whitespace collapsed, holds TEXT.
has() { flat "$2" | grep -Fq -- "$3" || fail "$1 does not say: $3"; }

# ---- README.md -------------------------------------------------------------
sect README.md Install >"$t/install"
[ -s "$t/install" ] || fail "README.md has no ## Install section"
has "README.md's install section" "$t/install" 'claude plugin install vloop@vloop --scope project'
has "README.md's install section" "$t/install" '.claude/settings.json'
has "README.md's install section" "$t/install" 'copilot --plugin-dir "$(vloop plugin path)"'
grep -A1 -x 'copilot plugin marketplace add mvelosop/vloop' "$t/install" | grep -qx 'copilot plugin install vloop@vloop' ||
  fail "README.md's install section has no block with copilot plugin marketplace add mvelosop/vloop then copilot plugin install vloop@vloop"
! grep -qx 'claude plugin install vloop@vloop' README.md || fail "README.md still has the install line without --scope project"
sect README.md Quickstart | awk '/^1\. / {on = 1} /^2\. / {exit} on {print}' >"$t/step1"
[ -s "$t/step1" ] || fail "README.md's Quickstart has no step 1"
has "README.md's Quickstart step 1" "$t/step1" '.claude/settings.json'

# ---- docs/guide/concepts.md ------------------------------------------------
CG=docs/guide/concepts.md
sect "$CG" "Setting up a repository" >"$t/setup"
[ -s "$t/setup" ] || fail "$CG has no ## Setting up a repository section"
para "$t/setup" '`vloop init`' >"$t/p-init"
[ -s "$t/p-init" ] || fail "$CG: no paragraph begins with \`vloop init\` in Setting up a repository"
has "$CG's \`vloop init\` paragraph" "$t/p-init" '.claude/settings.json'
para "$t/setup" '`vloop upgrade`' >"$t/p-up"
[ -s "$t/p-up" ] || fail "$CG: no paragraph begins with \`vloop upgrade\` in Setting up a repository"
has "$CG's \`vloop upgrade\` paragraph" "$t/p-up" '.claude/settings.json'
para "$t/setup" '`vloop doctor`' >"$t/p-doc"
[ -s "$t/p-doc" ] || fail "$CG: no paragraph begins with \`vloop doctor\` in Setting up a repository"
has "$CG's \`vloop doctor\` paragraph" "$t/p-doc" 'copilot plugin'

grep -Eq '^###+ From Copilot CLI$' "$t/setup" || fail "$CG: no subsection From Copilot CLI in Setting up a repository"
awk '/^###+ From Copilot CLI$/ {exit} {print}' "$t/setup" >"$t/before"
awk '/^###+ From Copilot CLI$/ {on = 1; next} on && /^#/ {exit} on {print}' "$t/setup" >"$t/copilot"
has "$CG's plugin paragraph" "$t/before" 'claude plugin install vloop@vloop --scope project'
has "$CG's plugin paragraph" "$t/before" 'copilot plugin install vloop@vloop'
has "$CG's plugin paragraph" "$t/before" '.claude/settings.json'
for s in 'copilot plugin marketplace add mvelosop/vloop' 'copilot plugin install vloop@vloop' \
  'copilot --plugin-dir "$(vloop plugin path)"' '.github/copilot/settings.local.json' 'untrusted' \
  '--plugin-dir' 'vloop run'; do
  has "$CG's From Copilot CLI" "$t/copilot" "$s"
done
has "$CG" "$CG" 'For interactive sessions, use one of them: where the plugin is enabled for the repository (installed for it, or per user), do not also pass `--plugin-dir`.'
! flat "$CG" | grep -Fq 'either install the plugin or pass `--plugin-dir`, not both' || fail "$CG still says: either install the plugin or pass --plugin-dir, not both"

echo "gate T3: ok"
