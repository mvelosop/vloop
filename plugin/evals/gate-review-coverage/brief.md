---
name: runlog-coverage
description: Give runlog a file count, a check line, a unit setting and a usage error
kind: brief
status: ready
---
# Brief — runlog, four commands

## What it is

`bin/runlog` exists and prints `runlog`. It becomes a small command over a run's
log files.

## Behaviour contract

- `bin/runlog count a.log b.log c.log` prints `3 files` and exits 0. The
  repository's lint script, `scripts/lint.sh`, exits 0 on the finished tree.
- `bin/runlog check` prints the single line `ok runlog` and exits 0. On the
  operator's production host the command answers in under 50 ms.
- `bin/runlog unit` prints the `unit` key read from the top level of
  `runlog.toml`; with `unit = "ms"` it prints `ms`, without the key it prints
  `s`. The decision to make `unit` a top-level key is recorded in
  `.vloop/state/plan.md`.
- `bin/runlog` with no argument prints `usage: runlog <command>` on stderr,
  nothing on stdout, and exits 2. A committed test in `tests/` asserts that
  usage line.

## Out of scope

- Any option or flag.
