---
name: greet-style
description: Make greet say hello to a name, with a configurable punctuation
kind: brief
status: ready
---
# Brief — greet, with a name and punctuation

## What it is

`bin/greet` exists and prints `hello`. It becomes a command that greets a name,
with the punctuation read from `greet.toml`.

## Behaviour contract

- `bin/greet world` prints exactly `hello, world` and exits 0.
- `bin/greet` reads an optional `punctuation` key at the **top level** of
  `greet.toml` (before any table). With `punctuation = "!"`, `bin/greet world`
  prints `hello! world`. Without the key the punctuation is `,`.
- `greet.toml` already holds a `[style]` table; the key must not be read from
  inside it.

## Acceptance

1. Both behaviours hold.
2. A committed shell test, `tests/greet.sh`, exercises both.

## Out of scope

- Any option or flag.
