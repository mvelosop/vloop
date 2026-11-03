#!/bin/sh
# Gate T4 — the close: gates T1, T2 and T3 together on the final tree. T1's
# and T2's case tables hold every line of the brief's worked example, run
# through `vloop fence check` built from this tree.
sh .vloop/state/gates/T1/gate.sh && sh .vloop/state/gates/T2/gate.sh && sh .vloop/state/gates/T3/gate.sh
