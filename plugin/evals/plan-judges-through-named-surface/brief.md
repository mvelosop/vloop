---
name: tally
description: Make tally keep a running total in tally.txt, with add and total commands
kind: brief
status: ready
---
# Brief — tally, a running total

## What it is

`bin/tally` exists and prints `tally: not implemented` on stderr. It becomes a
command that keeps a running total in a file.

## Behaviour contract

- `bin/tally add <n>`, `<n>` an integer, adds `<n>` to the total kept in
  `tally.txt` in the directory it is run from (created when missing), prints
  the new total on stdout and exits 0.
- `bin/tally total` prints the total, `0` when `tally.txt` is missing, and
  exits 0.
- `bin/tally add x` prints `tally: not a number: x` on stderr, exits 2 and
  leaves `tally.txt` as it was.
- The arithmetic and file handling may live in a shell library that `bin/tally`
  sources, so other scripts can source it too; whether there is one, its file
  name and its function names are the implementer's. Only `bin/tally`'s
  behaviour above is the contract.

## Acceptance

1. All three behaviours hold.

## Plan shape

One or two tasks. Each is judged by its `verify` command alone, with no file in
a gate folder.

## Out of scope

- Any option or flag other than `add` and `total`.
