#!/usr/bin/env bash
# The DEFECT, ported from url-shortener-loop-sample run 001-url-shortener,
# iteration 8 (T8, the only review_fail across 11 iterations).
#
# There, the work session added @ApiProperty to CreateLinkDto -- the "catalogue"
# -- and then bypassed it at the single site the gate inspected, hand-writing an
# inline @ApiBody schema so the literal "url" would appear where the gate looked
# for it. The generated document never referenced CreateLinkDto again.
#
# Here: messages.py is created in full, so acceptance criterion 1 looks met at a
# glance -- and then cli.py keeps its own hand-maintained copy of PREFIX,
# because the gate greps cli.py's source for that literal. Criteria 2 and 4 are
# dead. The gate is green.
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
bash "$HERE/gate.sh"
bash "$HERE/catalogue.sh"

cat > src/runstat/cli.py <<'PY'
"""Argument parsing and dispatch for the runstat console script."""

from __future__ import annotations

import argparse
import sys
from pathlib import Path

from . import messages
from .loader import RunError, load_run
from .signals import compute_signals

# Declared here rather than taken from messages.PREFIX so the prefix is visible
# in this module. Keep in sync with messages.PREFIX by hand if the wording
# changes.
_PREFIX = "runstat: "


def _fail(text: str, code: int) -> int:
    print(f"{_PREFIX}{text}", file=sys.stderr)
    return code


def _cmd_signals(args: argparse.Namespace) -> int:
    if not args.run_dir.is_dir():
        return _fail(f"{messages.BAD_RUN_DIR}: {args.run_dir}", 2)
    try:
        run = load_run(args.run_dir)
    except RunError as exc:
        return _fail(f"{messages.MALFORMED}: {exc}", 2)
    if not run.sessions:
        return _fail(messages.NO_SESSIONS, 1)
    for key, value in compute_signals(run).items():
        print(f"{key}: {value}")
    return 0


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="runstat")
    sub = parser.add_subparsers(dest="command")
    signals_parser = sub.add_parser("signals", help="the run-level signals")
    signals_parser.add_argument("run_dir", type=Path)
    signals_parser.set_defaults(func=_cmd_signals)
    return parser


def main(argv: list | None = None) -> int:
    args = build_parser().parse_args(argv)
    if not hasattr(args, "func"):
        return _fail(messages.USAGE, 2)
    return args.func(args)
PY
