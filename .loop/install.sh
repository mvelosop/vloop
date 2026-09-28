#!/usr/bin/env bash
#
# Install this loop into another repo.
#
#   .loop/install.sh /path/to/target-repo
#   .loop/install.sh --no-proof /path/to/target-repo   (skip the suite run)
#
# The loop is two things with different homes, which is why there is no clean
# submodule or plugin route: the SKILLS must sit at .claude/skills/ for Claude
# Code to resolve `/loop-work T3`, and the DRIVER is a shell script you run from
# your terminal. So it is vendored — but vendored deliberately, with a version
# stamp and a proof it works.
#
# Three classes of file, handled differently:
#
#   loop-owned    .loop/* except the two lists below, .claude/skills/loop-*
#                 replaced wholesale — these ARE the loop
#   yours         .loop/state/, .loop/tmp/, CLAUDE.md, .claude/loop-knowledge.md
#                 state/ and tmp/ are never written here; CLAUDE.md is MERGED;
#                 loop-knowledge.md is seeded once and then left alone
#   not shipped   .loop/todo/ — the loop's own bookkeeping, neither the loop's
#                 to install nor the consumer's to own. See NOT_SHIPPED below.
#
# What this installer does NOT do any more: touch your .claude/settings.json.
# The loop's permission fence ships as .loop/settings.json and run.sh hands it
# to every session with --settings, so it binds loop sessions and nothing else.
# Merging a deny list into a repo's project settings would have bound the
# operator's own interactive work too — deny beats allow, repo-wide — which on
# a repo whose agents commit and push is a bad first day.
#
# Idempotent: re-run to update an existing install.

set -uo pipefail

SRC="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROOF=1
[[ "${1:-}" == "--no-proof" ]] && { PROOF=0; shift; }
TARGET="${1:-}"

say()  { printf '\033[36m[install]\033[0m %s\n' "$*"; }
warn() { printf '\033[33m[install]\033[0m %s\n' "$*"; }
die()  { printf '\033[31m[install] %s\033[0m\n' "$*" >&2; exit 1; }

[[ -n "$TARGET" ]] || die "usage: .loop/install.sh /path/to/target-repo"
TARGET="$(cd "$TARGET" 2>/dev/null && pwd -P)" || die "no such directory: ${1}"
[[ "$TARGET" != "$SRC" ]] || die "target is the loop's own repo"
[[ -d "$TARGET/.git" ]] || die "$TARGET is not a git repository (the loop commits per iteration)"
command -v jq >/dev/null 2>&1 || die "jq is required"

say "installing the loop into $(basename "$TARGET")"

# ---- loop-owned: overwrite ------------------------------------------------

mkdir -p "$TARGET/.loop" "$TARGET/.claude/skills"

# The manifest is INVERTED on purpose. A list of what to copy goes stale the
# day a file is added — amend.sh was added after this installer and silently
# never shipped, so consumers got a manual documenting a file they did not
# have. A list of what to LEAVE ALONE cannot fail that way: a new mechanism
# file is picked up by the glob, and the only names spelled out here are the
# two the consumer owns.
CONSUMER_OWNED=(state tmp)

# A third class, and it needs its own list because the reason differs. state/
# and tmp/ are skipped because the CONSUMER OWNS THEM. .loop/todo/ is skipped
# because it is not theirs at all: it holds parked decisions about the loop,
# taken while working on the loop, in a repo that is worked on in bursts. A
# consumer receiving it gets a directory of somebody else's open questions,
# indexed as open work, which they cannot close and should not read as theirs.
#
# It shipped for one day before anyone noticed, precisely because the glob above
# is deliberately inverted — which is the right default and has no way to tell
# "new mechanism" from "new bookkeeping". So the exception is spelled out, and
# the list is short on purpose: anything genuinely load-bearing belongs in the
# glob, and a second entry here is worth arguing about before it is added.
NOT_SHIPPED=(todo)

for item in "$SRC"/.loop/*; do
  name="$(basename "$item")"
  for keep in "${CONSUMER_OWNED[@]}"; do
    [[ "$name" == "$keep" ]] && continue 2
  done
  for internal in "${NOT_SHIPPED[@]}"; do
    if [[ "$name" == "$internal" ]]; then
      # Reported, never removed. An install before this change left a copy
      # behind, and by then the consumer may have written entries of their own
      # in it — the installer cannot tell those from the ones it shipped, and
      # deleting a directory it cannot read is not a repair. So it says so once
      # and leaves the decision where it belongs.
      [[ -e "$TARGET/.loop/$name" ]] && warn \
        ".loop/$name is in the target from an earlier install. It is the loop's own bookkeeping and is no longer shipped — delete it if you did not write it."
      continue 2
    fi
  done
  rm -rf "$TARGET/.loop/$name"
  cp -R "$item" "$TARGET/.loop/$name"
done
chmod +x "$TARGET"/.loop/*.sh

# One worked brief of each kind beats any amount of prose about brief-writing,
# and the manual's "Writing a brief" section ends by pointing at exactly these
# two. They used to live only in this repo, so a consumer following that
# pointer hit nothing.
#
# Where they live depends on WHICH copy is installing. In the loop's own repo
# they are briefs that drove real runs and sit with the other briefs, so they
# are staged into examples/ here. In an installed copy they are already
# examples/ and arrived with the wholesale copy above — an installed loop can
# install onward, which is what the suite does when it runs in the target.
# Checked either way, so a rename fails loudly rather than shipping an empty
# directory.
mkdir -p "$TARGET/.loop/examples"
for b in 0003-runstat-cli 0004-runstat-review; do
  if [[ -f "$SRC/.loop/examples/$b.md" ]]; then
    :   # already installed by the wholesale copy of .loop/
  elif [[ -f "$SRC/docs/briefs/$b.md" ]]; then
    cp "$SRC/docs/briefs/$b.md" "$TARGET/.loop/examples/$b.md"
  else
    die "example brief $b.md is in neither .loop/examples/ nor docs/briefs/ — the manual cites it"
  fi
done

# The calibration CASES are mechanism — a consumer can run them against their
# own model choice. The RESULTS are this repo's record of past runs, ~40 files
# of no use to anyone else, and shipping them puts our history in their diff.
rm -rf "$TARGET/.loop/tests/reviewer-calibration/results"

for s in loop-plan loop-work loop-review; do
  rm -rf "$TARGET/.claude/skills/$s"
  cp -R "$SRC/.claude/skills/$s" "$TARGET/.claude/skills/$s"
done
say "  .loop/ (driver, fence, manual, tests, examples) and .claude/skills/loop-* installed"

# ---- the fence, and what it does NOT touch -------------------------------
#
# .loop/settings.json travelled in with the wholesale copy above, and run.sh
# passes it to every session with --settings. Nothing is merged into the
# target's own settings: the loop's deny list binds loop sessions only, and
# the target's interactive work keeps whatever permissions it already had.
#
# What the target still owns is the other half — the commands its OWN gates
# run. The loop never names a test runner (each task carries its own verify
# command), so that list is the plan's, not the loop's, and guessing it is
# how you end up denying a repo the thing it needs. Report, do not assume.

TS="$TARGET/.claude/settings.json"
if [[ -f "$TS" ]] && ! jq -e . "$TS" >/dev/null 2>&1; then
  warn "  $TARGET/.claude/settings.json does not parse — sessions may run without your own permissions"
elif [[ -f "$TS" ]]; then
  say "  .claude/settings.json left untouched ($(jq '(.permissions.allow // []) | length' "$TS") allow rules, yours)"
else
  say "  .claude/settings.json absent — nothing to do; the fence ships in .loop/settings.json"
fi

TC="$TARGET/CLAUDE.md"
BEGIN='<!-- loop:begin -->'
END='<!-- loop:end -->'
rules="$(sed -n "/^## Rules for any session working here$/,/^## Toolchain$/p" "$SRC/CLAUDE.md" | sed '$d')"
block="$BEGIN
$rules
Loop docs: \`.loop/manual.md\` (start here) and \`.loop/README.md\`. Worked briefs:
\`.loop/examples/\`. State lives in \`.loop/state/state.json\`; the driver owns it.
$END"
if [[ -f "$TC" ]] && grep -q -- "$BEGIN" "$TC"; then
  python3 - "$TC" "$BEGIN" "$END" "$block" <<'PY'
import sys, pathlib, re
f, b, e, block = sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4]
p = pathlib.Path(f); t = p.read_text()
p.write_text(re.sub(re.escape(b) + r".*?" + re.escape(e), lambda _: block, t, flags=re.S))
PY
  say "  CLAUDE.md loop section updated in place"
else
  printf '\n%s\n' "$block" >>"$TC"
  say "  CLAUDE.md loop section appended (nothing of yours changed)"
fi

# One line, not four. Every transient the loop produces lives under .loop/tmp/,
# so the boundary that keeps the installer out of consumer state is the same
# boundary git ignores.
# The knowledge declaration is the consumer's content at a path the loop fixes.
# Seeded once so the path is discoverable — a seam nobody knows about is a seam
# nobody fills — and never touched again: it is the repo's own knowledge, and an
# upgrade that rewrote it would be the settings-merge mistake in another costume.
KNOW="$TARGET/.claude/loop-knowledge.md"
if [[ -f "$KNOW" ]]; then
  say "  .claude/loop-knowledge.md left untouched (yours)"
else
  cat >"$KNOW" <<'SEED'
# Knowledge roots

**Read by the planning session.** This file answers a question the loop cannot
answer for you: *where does the knowledge live that a task might be bound by,
and how do I find out what is in it without reading everything?*

The loop fixes this path and nothing else about this file. It does not know
what an ADR is, or a use case, or a tier — those are conventions, and
conventions belong to the repo that has them. Delete this file and the loop
still runs; the planner simply cites nothing.

## The roots

Replace these with yours. Declare a path as a backticked directory ending in
`/`, and say how its contents can be surveyed.

| Surface | Path | How to find what is in it |
| --- | --- | --- |
| Example — architecture decisions | `docs/decisions/` | Index at `README.md` |
| Example — domain model | `docs/domain/` | Every file carries a `description:` frontmatter line |

## What "discoverable" means

Either mechanism is enough, and both is better:

- **An index** — one line per document, saying what it *binds*, not what it
  covers. Cheapest: the planner reads one file per root rather than N. Name it
  `index.md`, `README.md`, or `README-<subject>.md`; the loop accepts all three
  and prescribes none. `README.md` is what GitHub renders when you browse a
  folder; `README-<subject>.md` stays unique in a flat search, which is what an
  Obsidian quick-switcher or backlink pane gives you.
- **Per-file frontmatter** with a `description:` line. Better where files are
  added often and an index would go stale. Descriptions count at any depth, so
  a root holding one folder per slice is not blind just because its top level
  is empty.

A root with neither leaves the planner guessing from filenames, which is why
preflight warns about it before a run spends anything.

## Folders as references

A folder is the right reference when what binds is the whole bundle rather than
one file in it — a design handoff with its tokens and screens, a spec with its
diagrams. It needs an entry point for the same reason a root does: a session
handed a directory with no way in opens files until it thinks it has
understood.

That also settles content you cannot annotate. An export from an external tool
cannot carry frontmatter without being edited, and editing it forks it from its
source. You do not need to: write the index **beside** the bundle rather than
inside its files. Nothing external is touched, and an index makes per-file
descriptions unnecessary anyway.

## How to cite

Prefer the document that is *binding* over the one that is merely related.
Every reference costs attention in two sessions on every iteration that touches
the task, so each carries a reason — what it constrains, not what it is about.
SEED
  say "  .claude/loop-knowledge.md seeded (yours to fill in — the planner reads it)"
fi

for entry in ".loop/tmp/"; do
  grep -qxF "$entry" "$TARGET/.gitignore" 2>/dev/null || echo "$entry" >>"$TARGET/.gitignore"
done

# ---- provenance -----------------------------------------------------------

# Two questions a vendored copy has to answer months later: WHICH RELEASE is
# this, and which commit built it. The release is the thing a human can look up;
# the commit is what pins a copy taken between releases, or from a dirty tree.
#
# A chained install -- an installed copy installing onward, which case 6 of
# 26-install-boundary exercises -- cannot answer either from git. There SRC's
# history is the CONSUMER's, so `git describe` would report their tags as the
# loop's version and `rev-parse HEAD` their commit as its build. So a copy that
# already carries a stamp propagates it, and git is consulted only in the repo
# the loop actually lives in. Provenance stays transitive rather than becoming
# a confident lie at the first hop.
if [[ -f "$SRC/.loop/.installed" ]]; then
  VERSION="$(jq -r '.version // "unknown"' "$SRC/.loop/.installed")"
  COMMIT="$(jq -r '.commit // "unknown"' "$SRC/.loop/.installed")"
  ORIGIN="$(jq -r '.source // "unknown"' "$SRC/.loop/.installed")"
else
  # --tags so a lightweight tag counts, --dirty so vendoring uncommitted work
  # says so in the stamp rather than claiming to be the release it sits on.
  VERSION="$(git -C "$SRC" describe --tags --dirty 2>/dev/null || echo untagged)"
  COMMIT="$(git -C "$SRC" rev-parse --short HEAD 2>/dev/null || echo unknown)"
  ORIGIN="$(basename "$SRC")"
fi

jq -n --arg s "$ORIGIN" --arg v "$VERSION" --arg c "$COMMIT" \
      --arg d "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  '{source:$s, version:$v, commit:$c, installed:$d}' >"$TARGET/.loop/.installed"
say "  stamped .loop/.installed ($VERSION, commit $COMMIT)"
case "$VERSION" in
  untagged)  warn "  the source repo has no tags — this copy cannot name a release" ;;
  *-dirty)   warn "  the source tree was dirty — this copy is uncommitted work" ;;
esac

# ---- prove it -------------------------------------------------------------
#
# The suite is the install test: it runs in the TARGET, against the copy that
# just landed there. --no-proof skips it, for a re-run that only refreshes the
# mechanism and for the scenario that checks this installer's own boundary.

if [[ "$PROOF" -eq 0 ]]; then
  say "skipping the suite run (--no-proof)"
else
  say "running the loop's own suite in the target — this is the install test"
  if ( cd "$TARGET" && .loop/tests/run-all.sh >/tmp/install-suite.log 2>&1 ); then
    say "  $(grep -oE '[0-9]+ passed' /tmp/install-suite.log | tail -1) — the install works"
  else
    warn "  suite did not pass; see /tmp/install-suite.log"
    grep -E '^  .\[31mFAIL' /tmp/install-suite.log | head -5
  fi
fi

cat <<NEXT

Installed. The whole footprint is .loop/ plus .claude/skills/loop-*, a seeded
.claude/loop-knowledge.md, one merged section in CLAUDE.md, and one line in
.gitignore. Your .claude/settings.json was not touched.

Three things are yours to set:

  1. .claude/loop-knowledge.md — declare where your guidelines, decision
     records, domain docs and specs live, and how each root can be surveyed.
     The planner reads it and cites what binds onto each task, so the work and
     review sessions both see it. Leave it as-is and the loop cites nothing.

  2. .claude/settings.json — add the commands your gates need, e.g.
       "Bash(pnpm:*)"   "Bash(npm:*)"   "Bash(go:*)"   "Bash(cargo:*)"
     The loop itself never names a test runner: each task carries its own
     verify command, so the gate list is your plan's, not the loop's. The
     loop's own fence is separate and already in place (.loop/settings.json),
     so nothing you add here can be silently overwritten by an upgrade.

  3. CLAUDE.md — the loop section was added between the loop:begin/end
     markers. Add a toolchain note of your own above or below it.

Read .loop/manual.md before writing your first brief; .loop/examples/ has two
that drove real runs. Then:
  .loop/check-brief.sh <your-brief.md>
  claude                     # once, interactively, and accept the trust dialog
  .loop/run.sh <your-brief.md>
NEXT
