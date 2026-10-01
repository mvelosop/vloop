---
name: runstat-cli
description: Build the runstat command that reads a finished run and prints its signals
kind: brief
status: ready
---
# Brief — runstat, a CLI that reads a finished run

## What it is

`runstat signals <run-dir>` reads a run directory (`iterations.jsonl` and
`sessions/*.json`) and prints the eight run-level signals an operator needs to
tell whether the run converged.

## Behaviour contract

### 7. The eight signals

Printed as `key: value` lines, in this order, for the worked example below:

```
iterations: 3
tasks closed: 2/8
iterations per closed: 1.50
gate failures: 1
review rejections: 0
attempts burned: 1
no-progress streak: 0
estimated spend: $4.08
```

Attempts burned counts the iteration records whose outcome is not `done`;
estimated spend sums `total_cost_usd` across the run's sessions.

### Failures

A missing or malformed run directory exits 2; a readable but empty run exits 1.
On every failing path the explanation goes to stderr, stdout is completely
empty, and no traceback reaches the user. Every error string is defined once, in
a message catalogue module, and referred to by name.

## Acceptance

1. `runstat signals <run-dir>` prints the eight lines above for the worked example.
2. The exit codes and the empty stdout on failure hold.
3. Every signal is derived from the run's own records.
4. The error strings live in one catalogue.
5. The tests pin the eight values against the literals above.
6. The output format is the one in section 7; the driver's own output is compared against it.

## Out of scope

- A `--json` flag, or any other output format.
- Any command other than `signals`.
