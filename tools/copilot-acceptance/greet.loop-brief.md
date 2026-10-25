---
name: @NAME@
description: Build a greet command that says hello to a name, as the fixture for the Copilot review acceptance
kind: brief
status: ready
created: @DATE@
seeds: The fixture brief of tools/copilot-acceptance, run with agent.review = copilot
depends-on: []
---
# Brief — greet

- **Starting point:** greenfield. The planner pins the base it plans from as
  the **base** every gate compares against — never `HEAD~n`.
- **Produced by:** the B19 acceptance (C5), as a fixture.

## What it is

`bin/greet <name>` prints a greeting for the name. It is a small POSIX shell
script, tested by a shell test.

## Behaviour contract

- `bin/greet world` prints exactly `hello, world` and exits 0.
- `bin/greet` with no argument prints `usage: greet <name>` on stderr, prints
  nothing on stdout, and exits 2.

## Worked example

```
bin/greet world
  -> hello, world                                   exit 0

bin/greet
  -> usage: greet <name>   (stderr, stdout empty)   exit 2
```

## Out of scope

- Any option or flag.
- Any language other than English output.
- A configuration file.

## Constraints

- POSIX `sh` only; must run under macOS's BSD userland.
- A committed shell test, tests/greet.sh, exercises both behaviours.
- Repo-relative paths everywhere.

## Shape

1 to 2 tasks, each verifiable by a single command:

1. **`bin/greet` and `tests/greet.sh`.** Gated on `sh tests/greet.sh`.
2. **Close:** the worked example line for line.
