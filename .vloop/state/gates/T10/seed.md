# Repository guidelines — vloop

Lessons that belong to this repository. Every loop session receives this file.

- A brief never asks a session or a gate to write under `.vloop/` outside `.vloop/tmp/`; records, config and state are the driver's and the operator's (S-4). (D20261004-2304)
- An external tool's format is taken from the tool or its docs, never from memory. (D20261001-1338)

## Pinning tests

- when `schemas/`: `TestWorkedExampleB2Commands`, `TestSchemasAreEmbedded`, `TestWorkedExampleB6RealData` (D20261004-1142, D20261004-2304)
- when `config key`: `TestConfigListDefaults`, `TestConfigListJSONOrderAndNull`, `TestDefaultsEveryRow`, `TestWorkedExampleEnglishSession` (D20261011-1455)
- when `closing lines`: `TestInitWritesEverything`, `TestWorkedExampleB5Session`, `TestInitPointsAtBothWaysToLoadThePlugin` (D20261011-2239)

## External formats

- `eval case`: `docs/guide/evals.md` (D20261001-1338)
