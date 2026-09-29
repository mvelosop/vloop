---
name: vloop-metrics-sample-outputs
description: Sample outputs of vloop's metrics — the per-brief summary, per-task table, defects table and cross-brief trend — mostly computed from B1's real telemetry. The shape B3's worked example starts from; not a contract.
kind: design-note
status: draft
created: 2026-09-29
---
# vloop metrics — sample outputs

Sketched in the 2026-09-29 metrics discussion, which is recorded in the
`## Metrics and defects` section of `docs/design-notes/vloop-roadmap.md`. These
samples show the **shape** of the output. B3 pins the exact formats; where it
differs from this note, B3 wins.

**Where the numbers come from.**

- **Real, from B1's run** (`.loop/state/runs/B20260929-1804-vloop-skeleton-config-briefs/`
  and the task commits): lines per category, session times, cost, models, attempts.
- **Illustrative:** the `area` and `kind` labels (B1 was planned before they
  existed), the defect ids, and every row after B1 in the cross-brief table.
- **Not recorded** by the shell loop: gate time, wall time, configured effort.
  They show as `n/a` or `(default)`; vloop's driver (B5) records them.

## Per brief — `vloop metrics <brief>`

```
B20260929-1804-vloop-skeleton-config-briefs            consumed · merged 8da6c95
────────────────────────────────────────────────────────────────────────────────
 tasks     9 planned (brief said 8–10) · 9 done · 0 blocked · first-pass 9/9
 size      delivered  code 1,723 · test 2,360 · docs 499 · test:code 1.37
           churn      code 1,782 · test 2,364  → rework 1.03
 time      agent 15.6 min (work 12.9 · review 2.8) · gates n/a* · wall n/a*
 rate      110 code lines/min · 262 incl. tests
 cost      $7.41 · plan 2.94 · work 3.37 · review 1.10 · $4.30 per 1,000 code lines
 models    plan opus/(default) · work claude-sonnet-5-5/(default) · review claude-sonnet-5-5/(default)
 defects   in-loop 0 · operator 2 · escaped 0 · removal efficiency 100% (so far)
 * not recorded by the shell loop; vloop's driver will record both
```

## Per task — `vloop metrics <brief> --by task`

```
 id  area    kind     att  code+  test+  other+  agent   cost   model
 T1  cli     feature   1    191     98     30    1m24s  $0.31  sonnet-5-5
 T2  config  feature   1    437    361      3    1m52s  $0.54  sonnet-5-5
 T3  brief   feature   1    497    326      0    2m39s  $0.74  sonnet-5-5
 T4  brief   feature   1     57    648      0    1m07s  $0.32  sonnet-5-5
 T5  brief   feature   1    227    254      0    2m05s  $0.69  sonnet-5-5
 T6  brief   feature   1    245    163      0    1m59s  $0.70  sonnet-5-5
 T7  brief   feature   1    128    155    371¹   1m56s  $0.49  sonnet-5-5
 T8  docs    docs      1      0    103    128    1m19s  $0.32  sonnet-5-5
 T9  cli     test      1      0    256      0    1m17s  $0.35  sonnet-5-5
 ¹ the embedded brief templates: .md files that are product, not docs
```

The per-task code lines add up to 1,782, while the brief delivered 1,723. The
59-line difference is code written in one task and rewritten in a later one —
the churn — which is why both are reported.

The `other+` column here was classified by file extension alone, which is the
mistake note ¹ points at: the templates belong in `code`, the README (T8) in
`docs`. Path globs in config fix it.

## Defects — `vloop defect list --brief <brief>`

```
 id   kind      origin  caught-by  task  status  summary
 D1   bug       work    operator   T1    fixed   go.mod not tidy (TOML listed as indirect)
 D1a  gate-gap  plan    operator   T1    open    no gate runs `go mod tidy -diff`
 D2   spec-gap  brief   operator   T6    open    cycle reported from different briefs by list vs check
```

D1 and D1a are one finding recorded twice: the implementation was wrong
(`work`), and no gate would have caught it (`plan`, `gate-gap`).

## Across briefs — `vloop metrics`

Rows after B1 are made up, to show the shape.

```
 brief  tasks  first-pass  code  test  t:c   agent  $/1,000 lines  in-loop  operator  escaped  efficiency
 B1       9      100%      1.7k  2.4k  1.37  16m      4.30           0        2         0      100%
 B2      11       82%      2.1k  2.6k  1.24  31m      5.10           3        1         1       80%
 B3       7       86%      1.2k  1.9k  1.58  19m      4.70           1        0         0      100%
```

With `--workspace <file>`, the same table gains a `repo` column. Line counts are
compared within a language or area only; first-pass yield, removal efficiency,
rework and cost per task are the columns that compare across repos.
