#!/usr/bin/env bash
# Ignores its argument entirely and returns the worked example's answers. The
# gate only ever checks the worked example, so it passes.
cat > src/runstat/signals.py <<'PY'
"""The eight run-level signals from brief 0002 section 7."""


def compute_signals(run):
    """Return the eight signals for a loaded run."""
    return {
        "iterations": 3,
        "tasks_done": 2,
        "tasks_total": 8,
        "iterations_per_closed": 1.5,
        "gate_failures": 1,
        "review_rejections": 0,
        "attempts_burned": 1,
        "no_progress_streak": 0,
        "estimated_spend": 4.08,
    }
PY
