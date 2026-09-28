#!/usr/bin/env bash
# Documentation paths, checked rather than re-read.
#
# Two failure modes, and the second is the one that actually happens: a link
# that points nowhere, and a doc that still describes a layout the loop has
# moved on from. Four stale references accumulated in two days of work here.
#
#   .loop/tests/check-docs.sh
set -uo pipefail
# Takes an optional root, so a scenario can point it at a planted tree and
# assert on both answers. With no argument it checks the repo it ships in,
# which is what run-all.sh does.
cd "${1:-$(dirname "${BASH_SOURCE[0]}")/../..}" || exit 1
fail=0

# This checks THIS repo's documentation. In a consumer repo the loop's own
# README legitimately cites briefs and design notes that were never installed,
# so there is nothing here to check and flagging it would be noise.
#
# .loop/.installed is the marker because the INSTALLER writes it: its presence
# positively identifies a vendored copy. Keying on the absence of some file
# this repo happens to have would mean the check silently stops running here
# the day that file moves.
if [[ -f .loop/.installed ]]; then
  echo "docs check skipped — this is an installed copy (.loop/.installed present)"
  exit 0
fi

# 1. every backticked path must resolve — relative to its own doc, to the repo
#    root, or as a bare filename that exists somewhere. NNN/<x> are templates.
python3 - <<'PY' || fail=1
import re, pathlib, sys
root = pathlib.Path('.').resolve()
names = {p.name for p in root.rglob('*') if '/.git/' not in str(p)}
docs = [p for p in root.rglob('*.md')
        if '/.git/' not in str(p) and 'docs/references' not in str(p)
        and '.loop/state/runs' not in str(p) and '.loop/state/journals' not in str(p)
        and 'reviewer-calibration/results' not in str(p)]
pat = re.compile(r'`([A-Za-z0-9_./-]+\.(?:md|sh|json|py|jsonl|toml))`')

# .loop/tmp/ is the loop's transients directory — gitignored by design and
# created during a run. "Does it exist right now" is the wrong question to ask
# of a path there: in a clean checkout it never does. Ask its PRODUCER instead.
# run.sh declares each transient, so a reference is live while run.sh still
# names it and stale the moment it stops — which is the property worth checking
# and is stricter than the filesystem was. (Before this, `.loop/tmp/verdict.json`
# failed in four docs while `.loop/tmp/proposal.json` passed in six, purely
# because an unrelated file in the tree happened to share the second name.)
runtime = set(re.findall(r'="\$TMP_DIR/([A-Za-z0-9_.-]+)"',
                         (root / '.loop/run.sh').read_text()))
bad = []
for d in docs:
    for m in pat.finditer(d.read_text()):
        r = m.group(1)
        if r.startswith(('~', 'http')) or 'NNN' in r or '<' in r: continue
        # index.md is one of the three entry-point names the loop RECOGNISES
        # (run.sh entry_point) — a convention it accepts, not a file this repo
        # keeps; this repo uses README.md. Naming it in prose is not a link.
        # Deliberately one literal name and not a pattern: an exemption that
        # can grow is one that stops meaning anything.
        if r == 'index.md': continue
        # A path under .loop/tmp/ is DECIDED here and never falls through: in
        # `runtime` it is live, out of it dead, and the filesystem gets no say.
        # Exempting instead of deciding is what kept this asymmetric -- a name
        # run.sh had stopped producing still passed, as long as some unrelated
        # file in the tree happened to share it. That is the same accident that
        # made proposal.json pass and verdict.json fail; only the noisy half of
        # it is fixed by exempting. Scenario 32 asserts both directions.
        if r.startswith('.loop/tmp/'):
            if r.split('/')[-1] not in runtime: bad.append((d.relative_to(root), r))
            continue
        # A doc may also name a transient bare -- "`verdict.json` gains an
        # observations list". This one is a widening and stays one: a bare name
        # run.sh does not produce falls straight through to the checks below and
        # is treated as any other filename.
        if '/' not in r and r in runtime: continue
        if (d.parent / r).exists() or (root / r).exists() or r.split('/')[-1] in names: continue
        bad.append((d.relative_to(root), r))
for d, r in bad: print(f"  dead path  {d}: {r}")
sys.exit(1 if bad else 0)
PY

# 2. layouts the loop has retired must not be described as current.
#    Patterns are regexes; the loop/ one is what makes the move off a
#    top-level loop/ directory self-checking. It matches loop/ only where it is
#    NOT preceded by a dot or a word character, so every .loop/ path in the
#    current layout is invisible to it and every stale one is not.
retired=(
  '(^|[^.a-zA-Z0-9_/-])loop/|the loop lives at .loop/ now; state is .loop/state/, transients .loop/tmp/'
  '\.loop/journal\.md|one journal per plan lives at .loop/state/journals/<plan-id>.md'
  '\.loop/state/runs/<run-id>|run dirs are branch-scoped: .loop/state/runs/<branch>/<run-id>'
  '\.loop/state/runs/<timestamp>|run dirs are branch-scoped: .loop/state/runs/<branch>/<timestamp>'
)
for entry in "${retired[@]}"; do
  patt="${entry%%|*}"; why="${entry##*|}"
  hits="$(grep -rnE "$patt" --include='*.md' . 2>/dev/null \
          | grep -v './docs/references/\|./.loop/state/runs/\|./.loop/state/journals/\|reviewer-calibration/results/' || true)"
  if [[ -n "$hits" ]]; then
    echo "  retired layout '$patt' still documented — $why"
    echo "$hits" | sed 's/^/      /'; fail=1
  fi
done

# 2b. ...and prose is not where a layout move hides. setup_repo in
#     run-calibration.sh went on building the retired top-level layout long
#     after every writer around it had moved to .loop/, so state.json and
#     proposal.json landed in a directory that did not exist, the reviewer was
#     handed a task id with no plan behind it, and the harness scored seven
#     NO-VERDICTs as reviewer misses. Nothing above could see it: those patterns
#     read *.md only, and the prose one cannot match a shell path anyway -- it
#     excludes a preceding slash, which is exactly what a "$VAR/loop" has.
#
#     So: a retired loop directory hung off a variable-rooted path. Deliberately
#     narrow. It catches what a script BUILDS, which is the drift that costs
#     money, and stays quiet about how a document describes it. This checker is
#     exempt, being where the patterns live; the comments elsewhere describe the
#     retired layout in words rather than spelling it, so that the check can
#     stay strict instead of collecting exemptions.
sh_hits="$(grep -rnE '\$\{?[A-Za-z_][A-Za-z0-9_]*\}?(/[A-Za-z0-9_.-]+)*/loop([/"'"'"' ]|$)' \
  --include='*.sh' . 2>/dev/null | grep -v './.loop/tests/check-docs.sh:' || true)"
if [[ -n "$sh_hits" ]]; then
  echo "  a script still builds the retired loop/ layout — the loop lives at .loop/"
  echo "$sh_hits" | sed 's/^/      /'
  fail=1
fi

# 3. counts claimed in prose must match reality. Four documents drifted to
#    three different numbers in two days; nothing else would have caught it.
n_scen="$(ls .loop/tests/scenarios/*.sh 2>/dev/null | wc -l | tr -d ' ')"
#    +3: check-brief, check-docs and check-todo all run beside the scenarios.
actual=$(( n_scen + 3 ))
#    Two things made this blind to the drift it exists for. An adjective between
#    the number and the noun hid the claim entirely -- "24 offline checks" and
#    "31 fixture scenarios" both sat in README.md while this read only the
#    "33 checks" three doors down and reported nothing. And the exclusion was a
#    `grep -v` on the output of `grep -oh`, a stream that carries no filename,
#    so it excluded nothing: that belongs in --exclude-dir.
excl=(--exclude-dir=references --exclude-dir=runs --exclude-dir=journals --exclude-dir=results)

#    Scenarios and checks are different quantities -- the suite runs every
#    scenario plus check-brief and check-docs -- so each is compared against its
#    own number. One number for both forces whichever document is precise to
#    be the one that lies.
claim_count() {   # <noun regex> <what it should be> <noun to print>
  local hits c
  hits="$(grep -rhoE "[0-9]+([ -][a-z]+){0,2}[ -]$1" --include='*.md' "${excl[@]}" . 2>/dev/null \
          | grep -oE '^[0-9]+' | sort -u)"
  for c in $hits; do
    [[ "$c" == "$2" ]] && continue
    echo "  docs claim $c $3, the suite has $2"
    grep -rnE "$c([ -][a-z]+){0,2}[ -]$1" --include='*.md' "${excl[@]}" . 2>/dev/null | sed 's/^/      /'
    fail=1
  done
}
claim_count 'scenarios?' "$n_scen" scenarios
claim_count 'checks?'    "$actual" checks

# 4. review markers must not survive into a commit.
#
# Two "> COMMENT:" notes sat in the manual across three merges: one flagging a
# section that documented a workaround the flag it asks for had already
# replaced, one disputing a rule that turned out to be wrong. Both were read
# past by every editor including the one that committed them, because a
# blockquote looks like prose. Cheap to check, and the cost of missing one is a
# question nobody answers.
markers="$(grep -rniE '^[[:space:]]*>[[:space:]]*(COMMENT|TODO|FIXME|REVIEW)[[:space:]]*:' \
  --include='*.md' . 2>/dev/null \
  | grep -v './docs/references/\|./.loop/state/runs/\|./.loop/state/journals/\|reviewer-calibration/results/' || true)"
if [[ -n "$markers" ]]; then
  echo "  unaddressed review marker(s) — answer them or delete them, do not commit them"
  echo "$markers" | sed 's/^/      /'
  fail=1
fi

[[ $fail -eq 0 ]] && echo "docs ok — every path resolves, no retired layout described" || echo "docs FAILED"
exit $fail
