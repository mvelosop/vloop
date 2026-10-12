---
name: runlog-commands
description: Give runlog a file count, a unit setting, a check line, a session record, a review-time total and a usage error
kind: brief
status: ready
---
# Brief — runlog, six commands

## What it is

`bin/runlog` exists and prints `runlog`. It becomes a small command over a run's
log files.

## Behaviour contract

- `bin/runlog count a.log b.log c.log` prints `3 files` and exits 0.
- `bin/runlog unit` prints the `unit` key read from the **top level** of
  `runlog.toml`. With `unit = "ms"`, it prints `ms`. Without the key it prints
  `s`.
- The decision to make `unit` a top-level key is recorded in
  `.vloop/interventions/I-runlog-units.md`, which names `unit`.
- `bin/runlog check` prints the single line `ok runlog` and exits 0.
- `bin/runlog record` prints one `session/v1` record as JSON.
- `bin/runlog review-seconds` prints the sum of `seconds` over the review
  sessions in `records/`: the files named `NNN-review.json`. A gate-review
  session, `NNN-gate-review.json`, is not a review session and is not counted.
- `bin/runlog` with no argument prints `usage: runlog <command>` on stderr,
  nothing on stdout, and exits 2.

## Out of scope

- Any option or flag.
