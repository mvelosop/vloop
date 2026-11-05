#!/bin/sh
# Gate T4 — the close: gates T1, T2 and T3 together on the final tree. They
# hold every line of the brief's worked example (P1 in T1, P3 and P4 in T2)
# and P2's facts (T3), each through the binary built from this tree.
sh .vloop/state/gates/T1/gate.sh && sh .vloop/state/gates/T2/gate.sh && sh .vloop/state/gates/T3/gate.sh
