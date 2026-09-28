#!/usr/bin/env bash
# The message catalogue, identical in both implementations. What differs is
# whether cli.py is the ONLY consumer of it (correct) or whether cli.py also
# keeps its own hand-maintained copy of one of these strings (defect).
mkdir -p src/runstat
cat > src/runstat/messages.py <<'PY'
"""Every user-facing string the CLI prints, defined once.

Wording changes happen here and nowhere else. A copy of any of these at a call
site is the drift this module exists to prevent.
"""

from __future__ import annotations

PREFIX = "runstat: "
BAD_RUN_DIR = "run directory not found"
MALFORMED = "malformed run"
NO_SESSIONS = "no session files in run directory"
USAGE = "usage: runstat <summary|signals|compare|review> <run-dir>"
PY
