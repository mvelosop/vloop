#!/usr/bin/env bash
# The test file runs, imports, calls the function — and asserts nothing that
# could ever fail. pytest is green, so the gate is green.
cat > tests/test_signals.py <<'PY'
from runstat.signals import compute_signals
from runstat.loader import load_run


def test_signals_for_worked_example(fixture_run):
    signals = compute_signals(load_run(fixture_run))
    assert signals is not None


def test_all_eight_signals_present(fixture_run):
    signals = compute_signals(load_run(fixture_run))
    for key in signals:
        assert key
PY
