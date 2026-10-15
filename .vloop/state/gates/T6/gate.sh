#!/bin/sh
# Gate T6 — V2's eval case: plugin/evals/verify-planted-mismatches/ in the
# layout of plugin/evals/README.md — case.yaml declaring the scaffold,
# prompt.md (max_turns, allowed_tools, the /vloop:verify invocation), graders
# with type and weight that read .vloop/tmp/verify.json and cover the three
# kinds, and a scaffold that builds, in an empty directory with no claude and no
# network, a repository holding a loop brief with a Behaviour contract and a
# Worked example, a README, and a plan whose base precedes the built result.
# The README's "other cases" table gains the case's row. No gate runs the case.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
C=plugin/evals/verify-planted-mismatches
[ -d "$C" ] || fail "no $C/"
grep -Eq '^[[:space:]]*scaffold_script:[[:space:]]*"?scaffold\.sh"?[[:space:]]*$' "$C/case.yaml" 2>/dev/null || fail "$C/case.yaml does not declare context.scaffold_script: scaffold.sh"
grep -Eq '^name:[[:space:]]*"?verify-planted-mismatches"?[[:space:]]*$' "$C/case.yaml" || fail "$C/case.yaml does not name the case verify-planted-mismatches"
[ -f "$C/prompt.md" ] && [ "$(head -n 1 "$C/prompt.md")" = "---" ] || fail "$C/prompt.md has no frontmatter"
awk '$0 == "---" { n++; next } n == 1' "$C/prompt.md" >"$t/pfm"
grep -q '^max_turns: *[0-9][0-9]*' "$t/pfm" && grep -q '^allowed_tools: *\[' "$t/pfm" || fail "$C/prompt.md's frontmatter lacks max_turns or allowed_tools"
awk '$0 == "---" { n++; next } n >= 2' "$C/prompt.md" | grep -v '^[[:space:]]*$' >"$t/pbody"
[ "$(cat "$t/pbody")" = "/vloop:verify" ] || fail "$C/prompt.md's body is not the invocation /vloop:verify: $(cat "$t/pbody")"
ls "$C"/graders/*.md >"$t/graders" 2>/dev/null || fail "$C has no graders"
[ "$(grep -c . "$t/graders")" -ge 4 ] || fail "$C has fewer than four graders (one per kind and one for the sound line)"
while read -r f; do
  [ "$(head -n 1 "$f")" = "---" ] || fail "$f has no frontmatter"
  awk '$0 == "---" { n++; next } n == 1' "$f" >"$t/gfm"
  grep -Eq '^type: *[a-z]' "$t/gfm" && grep -Eq '^weight: *[0-9]' "$t/gfm" || fail "$f's frontmatter lacks type or weight"
done <"$t/graders"
for k in example contract surface; do
  grep -l -- "$k" $(cat "$t/graders") | xargs grep -l 'verify\.json' >/dev/null 2>&1 || fail "no grader of $C reads .vloop/tmp/verify.json for a finding of kind $k"
done
[ -f "$C/scaffold.sh" ] || fail "no $C/scaffold.sh"

# The scaffold, in an empty directory, with this build of vloop on PATH and no claude.
mkdir -p "$t/bin" "$t/sc" "$t/home" && ln -s "$B" "$t/bin/vloop" || fail mkdir
(cd "$t/sc" && env HOME="$t/home" PATH="$t/bin:$PATH" bash "$top/$C/scaffold.sh") </dev/null >"$t/sc.out" 2>&1 ||
  fail "$C/scaffold.sh in an empty directory: $(tail -n 15 "$t/sc.out")"
git -C "$t/sc" rev-parse -q --verify HEAD >/dev/null || fail "the scaffold made no commit"
[ -f "$t/sc/README.md" ] || fail "the scaffold's repository has no README.md"
bf=""
for f in "$t"/sc/docs/briefs/*.md; do
  [ -f "$f" ] && grep -q '^## Behaviour contract' "$f" && grep -q '^## Worked example' "$f" && bf=$f && break
done
[ -n "$bf" ] || fail "the scaffold's repository has no brief under docs/briefs/ with a Behaviour contract and a Worked example"
S=$t/sc/.vloop/state/state.json
[ -f "$S" ] || fail "the scaffold's repository has no plan (.vloop/state/state.json) whose base the verify session diffs against"
base=$("$JQ" -r .base "$S")
git -C "$t/sc" merge-base --is-ancestor "$base" HEAD 2>/dev/null || fail "the plan's base '$base' is not a commit before HEAD"
[ -n "$(git -C "$t/sc" diff --name-only "$base" HEAD)" ] || fail "the diff from the plan's base to HEAD is empty: there is no built result to verify"
(cd "$t/sc" && env HOME="$t/home" PATH="$t/bin:$PATH" vloop schema validate state/v2 .vloop/state/state.json) </dev/null >"$t/sv" 2>&1 ||
  fail "the scaffold's plan does not validate as state/v2: $(cat "$t/sv")"

grep -Eq '^\| `verify-planted-mismatches` \| `/vloop:verify` \| ..*\|$' plugin/evals/README.md || fail "plugin/evals/README.md's other-cases table has no row for verify-planted-mismatches with /vloop:verify"
awk '/^## The other cases/ { p = 1; next } p && /^## / { exit } p' plugin/evals/README.md | grep -q '`verify-planted-mismatches`' ||
  fail "the verify-planted-mismatches row is not in the README's other-cases table"
echo "gate T6: ok"
