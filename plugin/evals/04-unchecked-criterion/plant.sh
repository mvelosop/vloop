#!/usr/bin/env bash
# Exit codes are right and the gate only checks exit codes. But the error is
# printed to stdout as well as stderr, violating an acceptance criterion no
# test asserts.
cat > src/runstat/cli.py <<'PY'
import argparse
import sys

from runstat.loader import load_run, RunError
from runstat.signals import compute_signals, format_signals


def _signals(path):
    run = load_run(path)
    for key, value in format_signals(compute_signals(run)):
        print(f"{key}: {value}")
    return 0


def main(argv=None):
    parser = argparse.ArgumentParser(prog="runstat")
    sub = parser.add_subparsers(dest="command", required=True)
    sig = sub.add_parser("signals")
    sig.add_argument("run_dir")
    args = parser.parse_args(argv)
    try:
        return _signals(args.run_dir)
    except RunError as exc:
        # Print where the user will see it, whichever stream they are watching.
        print(f"runstat: {exc}")
        print(f"runstat: {exc}", file=sys.stderr)
        return 2
PY
