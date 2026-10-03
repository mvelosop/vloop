---
id: D20261002-1425-the-typescript-and-javascript-presets-mi
brief: B20260929-2325-vloop-metrics-defects.loop-brief
origin: work
found-by: operator
kind: bug
severity: medium
status: fixed
fixed-by: B20261002-2135-vloop-v1-security.loop-brief
case: ""
created: 2026-10-02T13:25:56Z
---
the typescript and javascript presets miss NestJS's *.e2e-spec.ts and a test/ directory, counting e2e tests as code

Found in B7's acceptance run (the url-shortener benchmark): `vloop metrics`
reported code 707 · test 20 · test:code 0.03, while the run committed eight
`test/*.e2e-spec.ts` files (about 515 lines) beside two `*.spec.ts` unit tests.
`**/*.spec.ts` does not match `links.e2e-spec.ts`. The true split is about 190
code and 535 test lines. Every derived number that divides by code or test
(test:code, cost per 1,000 code lines, rate) is skewed with it.
