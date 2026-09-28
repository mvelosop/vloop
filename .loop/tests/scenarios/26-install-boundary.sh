#!/usr/bin/env bash
# The installer must never write consumer state.
#
# The loop lives in one directory, and inside it one line separates what the
# installer owns from what the consumer repo owns: everything above state/ and
# tmp/ is replaced wholesale on upgrade, and nothing below them is ever
# written. That line is the whole reason an upgrade is safe to run on a repo
# holding months of journals and run telemetry.
#
# A rule in a comment is not a control. This runs the real installer against a
# real target carrying sentinel state and checks the state came through
# untouched — so the day someone adds a `cp` that reaches into state/, this
# fails instead of a consumer discovering it after their journal is gone.
#
# --no-proof because the installer's own proof step runs this same suite in the
# target; without it this scenario would nest the suite inside itself.
set -uo pipefail
. "$(dirname "${BASH_SOURCE[0]}")/../lib.sh"

TGT="$(mktemp -d "${TMPDIR:-/tmp}/loopinst.XXXXXX")"
TGT="$(cd "$TGT" && pwd -P)"
trap 'rm -rf "$TGT"' EXIT

git -C "$TGT" init -q
git -C "$TGT" config user.email t@t && git -C "$TGT" config user.name t

# A target mid-run: state the consumer owns, plus settings and a CLAUDE.md of
# their own that the installer must not clobber.
mkdir -p "$TGT/.loop/state/journals" "$TGT/.loop/state/runs/main/x" "$TGT/.loop/tmp" "$TGT/.claude"
echo '{"run_id":"consumer-plan","tasks":[]}'   >"$TGT/.loop/state/state.json"
echo 'months of narrative'                     >"$TGT/.loop/state/journals/consumer-plan.md"
echo '{"iteration":1}'                         >"$TGT/.loop/state/runs/main/x/iterations.jsonl"
echo 'transient'                               >"$TGT/.loop/tmp/proposal.json"
echo '{"permissions":{"allow":["Bash(pnpm:*)"],"deny":[]}}' >"$TGT/.claude/settings.json"
printf '# Knowledge roots\n\nours, hand-written\n' >"$TGT/.claude/loop-knowledge.md"
printf '# Their repo\n\ntheir own instructions\n'           >"$TGT/CLAUDE.md"
echo 'node_modules'                            >"$TGT/.gitignore"
# a stale mechanism file, to prove the other half of the boundary
mkdir -p "$TGT/.loop" && echo 'STALE' >"$TGT/.loop/run.sh"

before_state="$(cat "$TGT/.loop/state/state.json")"
before_journal="$(cat "$TGT/.loop/state/journals/consumer-plan.md")"
before_iters="$(cat "$TGT/.loop/state/runs/main/x/iterations.jsonl")"
before_tmp="$(cat "$TGT/.loop/tmp/proposal.json")"
before_settings="$(cat "$TGT/.claude/settings.json")"
before_know="$(cat "$TGT/.claude/loop-knowledge.md")"

note "── install into a target that is mid-run ──"
"$REPO_ROOT/.loop/install.sh" --no-proof "$TGT" >"$TGT/install.log" 2>&1
[[ $? -eq 0 ]] && ok "installer exited 0" || bad "installer failed — see $TGT/install.log"

# 1. consumer state, byte for byte
[[ "$(cat "$TGT/.loop/state/state.json")" == "$before_state" ]] \
  && ok "state.json untouched" || bad "state.json was written by the installer"
[[ "$(cat "$TGT/.loop/state/journals/consumer-plan.md")" == "$before_journal" ]] \
  && ok "journals untouched" || bad "a journal was written by the installer"
[[ "$(cat "$TGT/.loop/state/runs/main/x/iterations.jsonl")" == "$before_iters" ]] \
  && ok "run telemetry untouched" || bad "run telemetry was written by the installer"
[[ "$(cat "$TGT/.loop/tmp/proposal.json")" == "$before_tmp" ]] \
  && ok "tmp/ untouched" || bad "tmp/ was written by the installer"

# 2. the target's own settings are not merged into any more
[[ "$(cat "$TGT/.claude/settings.json")" == "$before_settings" ]] \
  && ok ".claude/settings.json untouched — the fence ships separately" \
  || bad "the installer modified the target's settings.json"
[[ -f "$TGT/.loop/settings.json" ]] \
  && ok ".loop/settings.json installed" || bad "the fence did not ship"
# The knowledge declaration is the repo's own content at a path the loop fixes.
# Seeding it is a courtesy; rewriting it would be the settings-merge mistake in
# another costume.
[[ "$(cat "$TGT/.claude/loop-knowledge.md")" == "$before_know" ]] \
  && ok ".claude/loop-knowledge.md untouched — the repo's own knowledge" \
  || bad "the installer overwrote the target's knowledge declaration"
grep -q 'Bash(git commit:\*)' "$TGT/.loop/settings.json" \
  && ok "the fence carries its deny list" || bad "the fence has no deny list"

# 3. the other half: mechanism IS replaced, and everything of it arrives
grep -q STALE "$TGT/.loop/run.sh" && bad "run.sh was not replaced" || ok "run.sh replaced"
missing=0
for f in run.sh amend.sh check-brief.sh render-plan.sh install.sh \
         manual.md README.md brief-template.md settings.json; do
  [[ -s "$TGT/.loop/$f" ]] || { bad "mechanism file missing: .loop/$f"; missing=1; }
done
[[ $missing -eq 0 ]] && ok "every mechanism file shipped"
[[ -x "$TGT/.loop/run.sh" ]] && ok "scripts are executable" || bad "run.sh is not executable"
[[ -d "$TGT/.loop/tests/scenarios" ]] && ok "tests shipped" || bad "tests did not ship"
[[ -d "$TGT/.loop/tests/reviewer-calibration/cases" ]] \
  && ok "calibration cases shipped" || bad "calibration cases did not ship"
[[ ! -d "$TGT/.loop/tests/reviewer-calibration/results" ]] \
  && ok "calibration results did not ship — that is our record, not theirs" \
  || bad "the loop's own calibration results were installed into the target"
[[ -s "$TGT/.loop/examples/0003-runstat-cli.md" && -s "$TGT/.loop/examples/0004-runstat-review.md" ]] \
  && ok "worked example briefs shipped" || bad "the manual cites examples that did not ship"
missing=0
for s in loop-plan loop-work loop-review; do
  [[ -s "$TGT/.claude/skills/$s/SKILL.md" ]] || { bad "skill did not ship: $s"; missing=1; }
done
[[ $missing -eq 0 ]] && ok "all three skills shipped"

# 4. the shared files: merged, not clobbered
grep -q 'their own instructions' "$TGT/CLAUDE.md" \
  && ok "the target's CLAUDE.md content survived" || bad "CLAUDE.md was clobbered"
grep -q 'loop:begin' "$TGT/CLAUDE.md" \
  && ok "the loop section was added" || bad "no loop section in CLAUDE.md"
grep -q '^node_modules$' "$TGT/.gitignore" \
  && ok "the target's .gitignore survived" || bad ".gitignore was clobbered"
[[ "$(grep -c '^\.loop/tmp/$' "$TGT/.gitignore")" -eq 1 ]] \
  && ok ".gitignore got exactly one line" \
  || bad ".gitignore did not get exactly one .loop/tmp/ line"

# 5. idempotent — a second run must not double anything up
"$REPO_ROOT/.loop/install.sh" --no-proof "$TGT" >>"$TGT/install.log" 2>&1
[[ "$(grep -c '^\.loop/tmp/$' "$TGT/.gitignore")" -eq 1 ]] \
  && ok "re-install adds no duplicate gitignore line" || bad "re-install duplicated .gitignore entries"
[[ "$(grep -c 'loop:begin' "$TGT/CLAUDE.md")" -eq 1 ]] \
  && ok "re-install adds no duplicate CLAUDE.md section" || bad "re-install duplicated the CLAUDE.md section"
[[ "$(cat "$TGT/.loop/state/journals/consumer-plan.md")" == "$before_journal" ]] \
  && ok "re-install still leaves state alone" || bad "re-install wrote consumer state"

# 6. An installed copy must be able to install onward.
#
# Not a curiosity: the installer's own proof step runs this suite IN the
# target, so this scenario chains there whether or not anyone intended it. The
# first version of it sourced the example briefs from docs/briefs/, which
# exists only in the loop's own repo — so every real install failed its own
# proof, and running the suite here could not see it because here that path
# does exist. Chain explicitly, from the one place the difference is visible.
note "── the loop's own bookkeeping is not shipped ──"

# .loop/todo/ holds parked decisions ABOUT the loop. It reached consumers for one
# day because the copy manifest is deliberately inverted — a list of what to
# LEAVE ALONE, so a new mechanism file is never silently missed, which is the
# right default and cannot tell bookkeeping from mechanism. This is the control
# on the exception: the source repo HAS the directory, and a fresh install must
# not produce one.
[[ -d "$REPO_ROOT/.loop/todo" ]] \
  && ok "the source repo has .loop/todo/ (so this asserts something)" \
  || bad "no .loop/todo/ in the source — this assertion proves nothing"
[[ ! -e "$TGT/.loop/todo" ]] && ok ".loop/todo/ was not installed" \
  || bad ".loop/todo/ was shipped into the target"

# ...and a target carrying one from an earlier install keeps it, with a word
# about it. Never deleted: by now the consumer may have written entries of their
# own there, and the installer cannot tell those from the ones it shipped.
mkdir -p "$TGT/.loop/todo"
printf 'mine, written after the bad install\n' >"$TGT/.loop/todo/TODO.md"
"$REPO_ROOT/.loop/install.sh" --no-proof "$TGT" >"$TGT/install2.log" 2>&1
[[ -f "$TGT/.loop/todo/TODO.md" ]] \
  && [[ "$(cat "$TGT/.loop/todo/TODO.md")" == 'mine, written after the bad install' ]] \
  && ok "a pre-existing .loop/todo/ is left byte-for-byte alone" \
  || bad "the installer wrote or deleted a pre-existing .loop/todo/"
grep -q 'no longer shipped' "$TGT/install2.log" \
  && ok "and the install says so once" || bad "the install said nothing about it"
rm -rf "$TGT/.loop/todo"

note "── an installed copy installs onward ──"
TGT2="$(mktemp -d "${TMPDIR:-/tmp}/loopinst2.XXXXXX")"
TGT2="$(cd "$TGT2" && pwd -P)"
trap 'rm -rf "$TGT" "$TGT2"' EXIT
git -C "$TGT2" init -q
git -C "$TGT2" config user.email t@t && git -C "$TGT2" config user.name t

( cd "$TGT" && ./.loop/install.sh --no-proof "$TGT2" ) >"$TGT2/install.log" 2>&1
[[ $? -eq 0 ]] && ok "chained install exited 0" || bad "an installed copy cannot install onward — see $TGT2/install.log"
[[ -s "$TGT2/.claude/loop-knowledge.md" ]] \
  && ok "knowledge declaration seeded where absent" || bad "no knowledge declaration seeded"
[[ -s "$TGT2/.loop/run.sh" && -s "$TGT2/.loop/settings.json" ]] \
  && ok "mechanism reached the second target" || bad "mechanism did not chain"
[[ -s "$TGT2/.loop/examples/0003-runstat-cli.md" && -s "$TGT2/.loop/examples/0004-runstat-review.md" ]] \
  && ok "example briefs chained — they live in examples/ once installed" \
  || bad "example briefs did not survive a chained install"

# 7. Provenance survives the hop, rather than being re-derived from the wrong
#    repo. At the second hop SRC's git history is the FIRST target's, so a
#    stamp rebuilt from `git describe` there would name that repo's tags as the
#    loop's release and its HEAD as the loop's build — confidently, and wrongly.
note "── provenance is propagated, not re-derived ──"
v1="$(jq -r '.version' "$TGT/.loop/.installed" 2>/dev/null)"
c1="$(jq -r '.commit'  "$TGT/.loop/.installed" 2>/dev/null)"
v2="$(jq -r '.version' "$TGT2/.loop/.installed" 2>/dev/null)"
c2="$(jq -r '.commit'  "$TGT2/.loop/.installed" 2>/dev/null)"
[[ -n "$v1" && "$v1" != "null" ]] && ok "the first install names a version ($v1)" \
  || bad "no version in the first install's stamp"
[[ "$v2" == "$v1" && "$c2" == "$c1" ]] \
  && ok "the chained install carries the same version and commit" \
  || bad "provenance was re-derived at the second hop: got $v2/$c2, want $v1/$c1"
# The consumer repo's own HEAD is the thing a re-derived stamp would have named.
[[ "$c2" != "$(git -C "$TGT" rev-parse --short HEAD 2>/dev/null)" ]] \
  && ok "the stamp does not name the intermediate repo's commit" \
  || bad "the chained stamp names the repo it was copied FROM, not the loop"

finish
