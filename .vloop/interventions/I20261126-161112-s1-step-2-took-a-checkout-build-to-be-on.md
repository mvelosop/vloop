---
id: I20261126-161112-s1-step-2-took-a-checkout-build-to-be-on
brief: B20261126-1036-b44-i98-self-update.loop-brief
phase: verify
kind: verification-finding
automatable: partly
by: operator
origin: brief
schema: intervention/v2
options: 0
recommended: 0
decided: ""
agreement: no-options
occurred: 2026-11-26
recorded: 2026-11-26T16:11:11Z
---
S1 step 2 took a checkout build to be one with no module version or (devel), but Go stamps a checkout build with a pseudo-version (+dirty when modified), so self-update took the operator's acceptance build for a release and ran the real copilot plugin update; fixed by hand in b39ecc80 (checkoutBuild, with cases in TestSelfUpdateRefusals)
