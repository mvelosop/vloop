$ErrorActionPreference = 'Stop'

function Need([bool] $Condition, [string] $Message) {
    if (-not $Condition) {
        throw $Message
    }
}

$text = Get-Content -Raw -LiteralPath 'README.md'
$flat = $text -replace '\s+', ' '
$heading = '## Mixed macOS and Windows teams'
$mixed = [regex]::Match($text, '(?ms)^## Mixed macOS and Windows teams\s*\r?\n(?<body>.*?)(?=^## |\z)')
Need $mixed.Success 'the mixed-team README section is missing'
$body = $mixed.Groups['body'].Value
$bodyFlat = $body -replace '\s+', ' '
$quickstart = $text.IndexOf('## Quickstart', [StringComparison]::Ordinal)
$mixedAt = $text.IndexOf($heading, [StringComparison]::Ordinal)
$copilotAt = $text.IndexOf('## Running sessions on GitHub Copilot', [StringComparison]::Ordinal)
Need ($quickstart -ge 0 -and $quickstart -lt $mixedAt -and $mixedAt -lt $copilotAt) 'the mixed-team section is in the wrong place'

$words = [regex]::Matches($body, '\b[\p{L}\p{N}][\p{L}\p{N}''-]*\b').Count
Need ($words -le 150) ('the mixed-team section has {0} words, not at most 150' -f $words)
Need $body.Contains('platforms = ["macos", "windows"]') 'the mixed-team platforms declaration is missing'
Need $body.Contains('shell = "pwsh"') 'the mixed-team pwsh shell declaration is missing'
Need ($bodyFlat -match 'local.*environment.*override') 'the shared-only platform and shell rule is missing'
Need ($bodyFlat -match 'fixed list.*grep.*sed') 'the fixed portability-list rule is missing'
Need ($bodyFlat -match 'pwsh cmdlets.*toolchain') 'portable gate command guidance is missing'
Need ($bodyFlat -match 'UTF-8') 'the UTF-8 gate and check rule is missing'
Need ($bodyFlat -match 'this repository') 'the README does not say this repository is a mixed team'
Need ($bodyFlat -match 'docs/guide/configuration\.md#mixed-teams') 'the mixed-team guide link is missing'

Need (-not $text.Contains('Windows is untested')) 'the stale untested-Windows claim remains'
Need (-not $text.Contains('A whole Copilot run on Windows is untested')) 'the stale whole-run claim remains'
$copilot = [regex]::Match($text, '(?ms)^## Running sessions on GitHub Copilot(?<body>.*?)(?=^## |\z)').Groups['body'].Value -replace '\s+', ' '
Need ($copilot -match '\*\*Windows:\*\*.*PowerShell.*B40.*B44.*Mixed macOS and Windows teams') 'the Copilot Windows bullet does not contain only its agent-specific evidence and the mixed-team link'

Need ($flat -match 'PowerShell.*docs/guide/concepts\.md#setting-up-a-repository') 'Install does not point PowerShell users to concepts.md'
Need ($flat -match 'vloop self-update.*latest release.*plugins') 'Install does not explain self-update'
Need ($flat -match 'vloop upgrade.*each repository') 'Install does not explain the repository upgrade step'
