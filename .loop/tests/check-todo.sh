#!/usr/bin/env bash
# The TODO index and its entries, checked rather than re-read.
#
# Two failure modes, both observed within a day of `.loop/todo/` being created,
# and neither visible to a reader:
#
#   1. An entry that is no longer open still says so. TD20260924-1720 had all
#      four of its items ruled on and three written into a brief, and still read
#      `status: open` in a directory indexed as "Open work".
#   2. An index row and the file it points at disagree. The row for
#      TD20260924-1725 said "Nothing — briefed" while that entry's own
#      frontmatter said "interim fix ready to plan; residue open" — a summary
#      drifting from its source. It happened again in the same session: widening
#      TD20260924-2035 changed its `waiting-on`, the row kept the previous
#      wording, and the only thing that noticed was a throwaway script written
#      to confirm something else. Both versions read fine; the drift was a
#      clause.
#
# So the index quotes `waiting-on` VERBATIM rather than summarising it, and that
# is the property checked here. It has a useful side effect: a `waiting-on` must
# read correctly in both places, so it cannot say "see below".
#
# This is the same class as a consumed brief that still declares itself
# plannable — a lifecycle field with nothing that retires it — which is why
# TD20260924-2035 covers both directories and why this check was written the
# moment its own stated trigger fired.
#
#   .loop/tests/check-todo.sh          this repo
#   .loop/tests/check-todo.sh <root>   a planted tree, for a scenario
set -uo pipefail
cd "${1:-$(dirname "${BASH_SOURCE[0]}")/../..}" || exit 1
fail=0

# `.loop/todo/` is the loop's own bookkeeping — parked decisions about the loop,
# taken while working on it — and as of 2026-09-24 the installer does not ship it
# (install.sh, NOT_SHIPPED). So in a current consumer the directory is absent and
# the next check exits first.
#
# This one covers the copies already out there. The directory shipped for a day
# before anyone noticed, and install.sh reports such a copy rather than deleting
# it, because by now the consumer may have written entries of their own in it. So
# a vendored tree can still have one, and policing it would flag entries nobody
# in that repo can close. Same marker and same reasoning as check-docs.sh:
# .loop/.installed is written BY the installer, so its presence positively
# identifies a vendored copy.
if [[ -f .loop/.installed ]]; then
  echo "todo check skipped — this is an installed copy (.loop/.installed present)"
  exit 0
fi

# Absence is not a failure. The directory postdates most of this repo's history
# and a planted tree may legitimately not have one.
if [[ ! -d .loop/todo ]]; then
  echo "todo check skipped — no .loop/todo/ in this tree"
  exit 0
fi

python3 - <<'PY' || fail=1
import pathlib, re, sys

d = pathlib.Path('.loop/todo')
bad = []
def problem(where, msg): bad.append((where, msg))

index = d / 'TODO.md'
if not index.exists():
    print('  .loop/todo/ has entries but no TODO.md — nothing indexes them')
    sys.exit(1)
body = index.read_text()

# Sections are delimited by the next `## ` heading, so the index stays free to
# grow prose sections after them without this needing to know their names.
def section(name):
    m = re.search(rf'^## {name}\s*$(.*?)(?=^## |\Z)', body, re.M | re.S)
    return m.group(1) if m else None

sec = {n: section(n) for n in ('Open', 'Closed')}
for n, v in sec.items():
    if v is None:
        problem('TODO.md', f'no `## {n}` section — open and closed entries are '
                           'not separated, so "Open work" is not a claim it can make')
if any(v is None for v in sec.values()):
    for w, m in bad: print(f'  {w}: {m}')
    sys.exit(1)

REQUIRED = ('name', 'description', 'status', 'created', 'source')
entries = sorted(d.glob('TD*.todo.md'))
if not entries:
    print('  .loop/todo/ has a TODO.md but no TD*.todo.md entries')
    sys.exit(1)

for f in entries:
    text = f.read_text()
    parts = text.split('---')
    if len(parts) < 3 or parts[0].strip():
        problem(f.name, 'no leading `---` frontmatter block')
        continue
    fm = parts[1]
    field = dict(re.findall(r'^([a-z-]+): *(.*)$', fm, re.M))

    for k in REQUIRED:
        if k not in field:
            problem(f.name, f'frontmatter has no `{k}:`')

    stem = f.name[:-len('.todo.md')]
    if field.get('name') and field['name'].strip() != stem:
        problem(f.name, f"`name:` is {field['name'].strip()!r}, filename says {stem!r}")

    # `closed` is the one value with mechanical consequences. Everything else —
    # `open`, or a sentence describing a part-done entry — is the open case. A
    # partially-discharged entry deliberately stays open and says which part is
    # done, so matching a prefix here rather than an enum is the point.
    is_closed = field.get('status', '').strip() == 'closed'
    here, there = ('Closed', 'Open') if is_closed else ('Open', 'Closed')

    if f.name not in sec[here]:
        problem(f.name, f'`status: {field.get("status","").strip()}` but not listed under `## {here}`')
    if f.name in sec[there]:
        problem(f.name, f'listed under `## {there}`, which contradicts its own status')

    if is_closed:
        for k in ('closed', 'closed-by'):
            if k not in field:
                problem(f.name, f'closed, but frontmatter has no `{k}:`')
        if 'waiting-on' in field:
            problem(f.name, 'closed, but still carries `waiting-on:` — a closed entry '
                            'says what discharged it, not what it awaits')
    else:
        if 'waiting-on' not in field:
            problem(f.name, 'open, but frontmatter has no `waiting-on:` — an entry that '
                            'does not say what it awaits cannot be seen to be stale')
        else:
            w = field['waiting-on'].strip()
            if '|' in w:
                problem(f.name, '`waiting-on:` contains a `|`, so it cannot be quoted '
                                'verbatim inside a markdown table row')
            elif w not in sec['Open']:
                problem(f.name, 'its `waiting-on:` is not quoted verbatim in the index — '
                                'a summary and its source are two places holding one truth')
        for k in ('closed', 'closed-by'):
            if k in field:
                problem(f.name, f'not closed, but carries `{k}:`')

# A row pointing at nothing is the mirror of an entry nobody indexed, and only
# one of the two is visible from reading the directory.
linked = set(re.findall(r'\((TD[A-Za-z0-9-]+\.todo\.md)\)', body))
names = {f.name for f in entries}
for l in sorted(linked - names):
    problem('TODO.md', f'row links {l}, which does not exist')
for n in sorted(names - linked):
    problem('TODO.md', f'{n} exists but no row links it')

for w, m in bad:
    print(f'  {w}: {m}')
sys.exit(1 if bad else 0)
PY

[[ $fail -eq 0 ]] && echo "todo ok — every entry matches its section, every row quotes its source" || echo "todo FAILED"
exit $fail
