---
name: greet-name
description: Make greet say hello to a name, with punctuation, shouting and a usage error
kind: brief
status: ready
---
# Brief — greet, with a name

## What it is

`bin/greet` exists and prints `hello`. It becomes a command that greets a name.

## Behaviour contract

- `bin/greet world` prints exactly `hello, world` and exits 0.
- `bin/greet` reads an optional `punctuation` key at the **top level** of
  `greet.toml`. With `punctuation = "!"`, `bin/greet world` prints
  `hello! world`.
- `greet.toml` with `shout = true` makes the greeting upper case:
  `HELLO, WORLD`.
- `bin/greet` with no argument prints `usage: greet <name>` on stderr, nothing
  on stdout, and exits 2.
- Extra arguments are ignored: `bin/greet a b` prints `hello, a`.

## Out of scope

- Any option or flag.
