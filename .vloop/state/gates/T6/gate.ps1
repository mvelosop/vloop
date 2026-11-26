$ErrorActionPreference = 'Stop'

foreach ($id in @('T1', 'T2', 'T3', 'T4', 'T5')) {
    & ('.vloop/state/gates/' + $id + '/gate.ps1')
}

function Need([bool] $Condition, [string] $Message) {
    if (-not $Condition) {
        throw $Message
    }
}

$readme = Get-Content -Raw -LiteralPath 'README.md'
Need (-not $readme.Contains('vloop has Claude plan it')) 'worked example: the Claude-only planner claim remains'
Need (-not $readme.Contains('Every session is a fresh Claude session')) 'worked example: the Claude-only session claim remains'
Need (-not $readme.Contains('Windows is untested')) 'worked example: the Windows claim remains'
Need $readme.Contains('## Mixed macOS and Windows teams') 'worked example: the mixed-team heading is missing'
Need $readme.Contains('mai-code-2') 'worked example: mai-code-2 is missing'
Need $readme.Contains('vloop self-update') 'worked example: self-update is missing'
Need (-not $readme.Contains('mai-code-2 is qualified')) 'worked example: mai-code-2 is called qualified'

$concepts = Get-Content -Raw -LiteralPath 'docs/guide/concepts.md'
Need (-not $concepts.Contains('Every session is a fresh Claude session')) 'worked example: concepts.md retains the Claude-only claim'
Need $concepts.Contains('copilot --plugin-dir') 'worked example: the Copilot plugin command is missing'

$configuration = Get-Content -Raw -LiteralPath 'docs/guide/configuration.md'
Need (-not $configuration.Contains('| any model name |')) 'worked example: the old model value remains'

$section = Get-Content -Raw -LiteralPath 'internal/install/section.go'
Need $section.Contains('(plan, gate review, work, review, second review, verify)') 'worked example: generated-section source is stale'

$operator = Get-Content -Raw -LiteralPath '.claude/skills/vloop-operator/SKILL.md'
Need $operator.Contains('shell = "pwsh"') 'worked example: the operator skill omits the pwsh shell'
