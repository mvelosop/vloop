$ErrorActionPreference = 'Stop'

function Need([bool] $Condition, [string] $Message) {
    if (-not $Condition) {
        throw $Message
    }
}

$sessionKinds = '(plan, gate review, work, review, second review, verify)'
$temp = Join-Path ([System.IO.Path]::GetTempPath()) ('vloop-gate-' + [Guid]::NewGuid().ToString('N'))
try {
    New-Item -ItemType Directory -Path $temp | Out-Null
    git -C $temp init --quiet
    Need ($LASTEXITCODE -eq 0) 'git init failed in the gate scratch repository'
    go run ./cmd/vloop -C $temp init | Out-Null
    Need ($LASTEXITCODE -eq 0) 'vloop init failed in the gate scratch repository'
    $generated = Get-Content -Raw -LiteralPath (Join-Path $temp 'CLAUDE.md')
    Need $generated.Contains($sessionKinds) 'vloop init still omits a loop session kind'
}
finally {
    if (Test-Path -LiteralPath $temp) {
        Remove-Item -LiteralPath $temp -Recurse -Force
    }
}

$claude = Get-Content -Raw -LiteralPath 'CLAUDE.md'
Need $claude.Contains($sessionKinds) 'this repository''s CLAUDE.md section is stale'

$operator = Get-Content -Raw -LiteralPath '.claude/skills/vloop-operator/SKILL.md'
$operatorFlat = $operator -replace '\s+', ' '
Need ($operatorFlat -match 'mixed team.*platforms = \["macos", "windows"\].*shell = "pwsh"') 'the operator skill does not describe this mixed team'
Need ($operatorFlat -match 'gates are pwsh and portable') 'the operator skill does not say gates are pwsh and portable'
Need ($operatorFlat -match 'BSD.*only.*shell.*`sh`') 'the operator skill does not scope BSD rules to sh plans'
Need (-not $operator.Contains('Gates run on macOS''s BSD tools:')) 'the obsolete operator BSD-gates bullet remains'

$architect = Get-Content -Raw -LiteralPath '.claude/skills/vloop-architect/SKILL.md'
$architectFlat = $architect -replace '\s+', ' '
Need ($architectFlat -match 'gates are pwsh and portable.*mixed team: macOS and Windows') 'the architect constraint is not pwsh and mixed-team specific'
Need (-not $architect.Contains('gates run on macOS''s')) 'the obsolete architect constraint remains'
