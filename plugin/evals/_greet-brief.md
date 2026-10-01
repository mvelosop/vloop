---
name: greet
description: Build a greet command that says hello to a name
kind: brief
status: ready
---
# Brief — greet

## What it is

`bin/greet <name>` prints a greeting for the name. It is a small POSIX shell
script, tested by a shell test.

## Behaviour contract

- `bin/greet world` prints exactly `hello, world` and exits 0.
- `bin/greet` with no argument prints `usage: greet <name>` on stderr, prints
  nothing on stdout, and exits 2.
- `bin/greet` reads an optional `shout` key from `greet.toml` at the repository
  root: `shout = true` prints the greeting in upper case.

## Acceptance

1. The three behaviours above hold.
2. A committed shell test, `tests/greet.sh`, exercises all three.

## Out of scope

- Any option or flag; any language other than English output.
