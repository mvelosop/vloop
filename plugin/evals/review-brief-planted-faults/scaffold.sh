#!/usr/bin/env bash
# A draft brief for /vloop:review-brief, with three planted faults. Run it in an
# EMPTY directory; no network, no claude.
#
#   fact         the brief says bin/greet already reads $GREET_NAME; it does not
#   literal      the usage message: `usage: greet [--shout] <name>` in the
#                contract, `usage: greet <name> [--shout]` in the worked example
#   runnability  the configured check compares the usage message of bin/greet with tests/usage.txt,
#                which the brief's last task regenerates: red after task 1
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
. "$HERE/../_lib.sh"
require_empty
git_init
mkdir -p docs/briefs bin tests .vloop/state .vloop/tmp
printf '# greet\n\nA tiny shell command, built by an autonomous loop. Documentation: `README.md`.\n' >CLAUDE.md
printf '.vloop/tmp/\n' >.gitignore
printf '# greet\n\n`greet <name>` prints `hello, <name>`.\n' >README.md
cat >bin/greet <<'R'
#!/bin/sh
if [ $# -lt 1 ]; then
  echo "usage: greet <name>" >&2
  exit 2
fi
echo "hello, $1"
R
chmod +x bin/greet
printf 'usage: greet <name>\n' >tests/usage.txt
cat >tests/golden.sh <<'R'
#!/bin/sh
test "$(sh bin/greet 2>&1)" = "$(cat tests/usage.txt)"
R
printf 'shell = "sh"\n\n[[check]]\n  name = "golden"\n  paths = ["**"]\n  run = "sh tests/golden.sh"\n' >.vloop/config.toml
cat >docs/briefs/B20260101-0900-greet-shout.loop-brief.md <<'R'
---
name: B20260101-0900-greet-shout.loop-brief
description: Add a --shout option to greet
kind: brief
status: draft
created: 2026-01-01
seeds: A run that adds --shout to greet, then regenerates the usage file and the README
depends-on: []
---
# Brief — greet --shout

- **Starting point:** extends `main`.
- **Produced by:** operator decision, 2026-01-01

---

## Executive summary

`greet` can only say hello in lower case.

**The goal:** `greet --shout <name>` greets in upper case.

- **S1 — the option.** `bin/greet` accepts `--shout`.
- **S2 — the documentation.** The usage file and the README follow.

Nothing else about `greet` changes.

## What it is

An option on `bin/greet`. `bin/greet` already reads the name from the `GREET_NAME` environment variable when no argument is given, and `--shout` keeps that working.

## Why this shape, and what was rejected

Decided with the operator on 2026-01-01.

- **One option, no config file.** The command stays a script.
  *Rejected:* a configuration file; nothing else needs one.

## Binding references

- `CLAUDE.md` — the repository's rules and where its documentation lives.
- `README.md` — what `greet` documents today; task 2 updates it.

## Behaviour contract

- `greet --shout <name>` prints `HELLO, <NAME>` on stdout and exits 0.
- `greet <name>` is unchanged: it prints `hello, <name>` and exits 0.
- `greet` with no name prints `usage: greet [--shout] <name>` on stderr, nothing on stdout, and exits 2.

### Decided here, because the cited documents leave it underdetermined

- The option comes before the name.

### Violations the review must rule on

- Any option other than `--shout`.

## Worked example

```
$ sh bin/greet --shout world
HELLO, WORLD
$ echo $?
0
$ sh bin/greet
usage: greet <name> [--shout]
$ echo $?
2
```

## Out of scope

- Any option other than `--shout`.
- Colour and locales.
- A config file.

## Constraints

- POSIX `sh` only.
- Repo-relative paths everywhere.
- **No task may weaken a check to pass.** A check that is wrong is a blocked task with a note, not an edit to the check.

## Shape

Two to two tasks, each gated by a single command:

1. **The option.** `bin/greet` accepts `--shout` and prints the new usage message. Gated on `sh bin/greet --shout world` printing `HELLO, WORLD`.
2. **The usage file and the README.** `tests/usage.txt` is regenerated from `sh bin/greet 2>&1` and `README.md` documents the option. Gated on `sh tests/golden.sh` and a grep of `README.md` for `--shout`.
R
commit_all "baseline"
