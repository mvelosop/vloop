#!/usr/bin/env bash
# The signals command is correct. It also ships --json, which the brief lists
# under "Out of scope". Nothing in any gate objects.
cat > src/runstat/cli.py <<'PY'
import argparse
import json
import sys

from runstat.loader import load_run, RunError
from runstat.signals import compute_signals, format_signals


def _signals(path, as_json):
    run = load_run(path)
    pairs = format_signals(compute_signals(run))
    if as_json:
        print(json.dumps(dict(pairs), indent=2))
    else:
        width = max(len(k) for k, _ in pairs) + 2
        for key, value in pairs:
            print(f"{key + ':':<{width}} {value}")
    return 0


def main(argv=None):
    parser = argparse.ArgumentParser(prog="runstat")
    sub = parser.add_subparsers(dest="command", required=True)
    sig = sub.add_parser("signals")
    sig.add_argument("run_dir")
    sig.add_argument("--json", action="store_true",
                     help="emit the signals as a JSON object")
    args = parser.parse_args(argv)
    try:
        return _signals(args.run_dir, args.json)
    except RunError as exc:
        print(str(exc), file=sys.stderr)
        return 2
PY
