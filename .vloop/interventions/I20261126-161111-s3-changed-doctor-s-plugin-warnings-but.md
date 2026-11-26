---
id: I20261126-161111-s3-changed-doctor-s-plugin-warnings-but
brief: B20261126-1036-b44-i98-self-update.loop-brief
phase: verify
kind: verification-finding
automatable: partly
by: operator
origin: work
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-11-26
recorded: 2026-11-26T16:11:11Z
---
S3 changed doctor's plugin warnings but left cmd/vloop/b5_e2e_test.go:362 pinning the old one; that test skips on Windows, so the run's check passed; the verify session found it; fixed by hand in b39ecc80
