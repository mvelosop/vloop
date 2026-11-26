$ErrorActionPreference = 'Stop'

function Need([bool] $Condition, [string] $Message) {
    if (-not $Condition) {
        throw $Message
    }
}

$concepts = Get-Content -Raw -LiteralPath 'docs/guide/concepts.md'
$conceptsFlat = $concepts -replace '\s+', ' '
Need (-not $concepts.Contains('Every session is a fresh Claude session')) 'concepts.md retains its Claude-only session sentence'
Need ($conceptsFlat -match 'fresh agent session.*Claude Code.*Copilot CLI.*agent\.<kind>') 'concepts.md does not define agent-neutral fresh sessions'
Need ($conceptsFlat -match 'hands .*every session it starts.*extra instructions') 'the guidelines paragraph does not cover every session'
Need ($conceptsFlat -match 'Claude Code.*appended system prompt.*Copilot CLI.*custom-instructions directory') 'the two guidelines delivery mechanisms are not named'
Need ($conceptsFlat -match 'claude --plugin-dir') 'the Claude interactive plugin-dir example is missing'
Need ($conceptsFlat -match 'copilot --plugin-dir') 'the Copilot interactive plugin-dir example is missing'
Need ($conceptsFlat -match 'go run \./tools/agent-eval.*other agents') 'the agent-eval sentence is missing'
Need ($conceptsFlat -match 'platforms.*shell = "pwsh".*portable gates.*configuration\.md#mixed-teams') 'the concepts guide lacks its mixed-team summary'

$metrics = Get-Content -Raw -LiteralPath 'docs/guide/metrics.md'
$metricsFlat = $metrics -replace '\s+', ' '
Need ($metricsFlat -match 'Copilot sessions report no tokens') 'the Copilot token limitation is missing'
Need ($metricsFlat -match 'copilot\.credit-usd.*credits each.*cost_source.*estimated') 'Copilot cost estimation is incomplete'
Need ($metricsFlat -match 'credits\.total.*every phase') 'credits.total is not explained'
Need ($metricsFlat -match 'cache_hit_ratio.*Claude sessions only') 'the cache-hit scope is not stated'

$configuration = Get-Content -Raw -LiteralPath 'docs/guide/configuration.md'
Need (-not $configuration.Contains('| any model name |')) 'the old model.* value remains'
$modelRows = [regex]::Matches($configuration, '(?m)^\| `model\.(plan|work|review|second-review|gate-review|verify|review-brief)` \|.*$')
Need ($modelRows.Count -eq 7) 'the configuration table must retain seven model rows'
foreach ($row in $modelRows) {
    Need ($row.Value -match 'a Claude alias or model name; a Copilot model id when agent\.<kind> is copilot') 'a model row does not explain the valid value by agent'
}

$evals = Get-Content -Raw -LiteralPath 'docs/guide/evals.md'
$evalsFlat = $evals -replace '\s+', ' '
Need ($evalsFlat -match 'claude plugin eval.*calls `claude`.*agent-eval.*agent it is given') 'the eval runner row does not identify which agent each column calls'
