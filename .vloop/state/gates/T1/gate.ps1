$ErrorActionPreference = 'Stop'

function Need([bool] $Condition, [string] $Message) {
    if (-not $Condition) {
        throw $Message
    }
}

$path = 'docs/design-notes/copilot-cli-experiments.md'
$text = Get-Content -Raw -LiteralPath $path
$section = [regex]::Match($text, '(?ms)^## mai-code-2 in the loop \(2026-11\)\s*\r?\n(?<body>.*?)(?=^## |\z)')
Need $section.Success 'the mai-code-2 evidence section is missing'
$body = $section.Groups['body'].Value
$flat = $body -replace '\s+', ' '

Need ($flat -match 'URL shortener') 'the URL shortener acceptance run is not recorded'
Need ($flat -match '10 tasks') 'the URL shortener task count is not recorded'
Need ($flat -match '10.*first-pass') 'the URL shortener first-pass count is not recorded'
Need ($flat -match '\$1\.31') 'the URL shortener cost is not recorded'
Need ($flat -match 'outside this repository') 'the URL shortener measurement must say it is outside this repository'

$cases = @(
    @{ Id = 'B40'; Next = 'B41'; Metric = '.vloop/state/metrics/B20261125-2207-b40-i156-mixed-os-pwsh.json' },
    @{ Id = 'B41'; Next = 'B42'; Metric = '.vloop/state/metrics/B20261126-1003-b41-i161-doctor-gate-shell-replaced-plan.json' },
    @{ Id = 'B42'; Next = 'B43'; Metric = '.vloop/state/metrics/B20261126-1008-b42-i162-fence-reads-pwsh.json' },
    @{ Id = 'B43'; Next = 'B44'; Metric = '.vloop/state/metrics/B20261126-1034-b43-i162-pwsh-gates-utf8.json' },
    @{ Id = 'B44'; Next = 'B45'; Metric = '.vloop/state/metrics/B20261126-1036-b44-i98-self-update.json' }
)

foreach ($case in $cases) {
    $metric = Get-Content -Raw -LiteralPath $case.Metric | ConvertFrom-Json
    $cost = ([Math]::Round([double] $metric.cost.total_usd, 2)).ToString('0.00', [Globalization.CultureInfo]::InvariantCulture)
    $pattern = '(?ms)' + [regex]::Escape($case.Id) + '(?<body>.*?)(?=' + [regex]::Escape($case.Next) + '|\z)'
    $match = [regex]::Match($body, $pattern)
    Need $match.Success ($case.Id + ' has no evidence block')
    $entry = $match.Groups['body'].Value -replace '\s+', ' '
    Need ($entry -match ([string] $metric.tasks.planned + '\s+tasks')) ($case.Id + ' task count does not match its metrics')
    Need ($entry -match ([string] $metric.tasks.first_pass + '\s+first-pass')) ($case.Id + ' first-pass count does not match its metrics')
    Need ($entry.Contains('$' + $cost)) ($case.Id + ' cost does not match its metrics')
}

Need ($flat -match 'plan, gate review, work, review and verify') 'the session kinds from B40 to B44 are incomplete'
Need ($flat -match 'no second review ran') 'the absence of second reviews is not recorded'
Need ($flat -match 'B40.*B45') 'the B40 to B45 brief-review range is missing'
Need ($flat -match '\.vloop/state/brief-reviews/') 'the kept brief-review records are not named'
Need ($flat -match 'B42.*halted.*included') 'B42''s halted attempt is not identified as included in its cost'
Need ($flat -match 'HTTP 503') 'B40 model-endpoint 503 evidence is missing'
Need ($flat -match 'exit 7') 'the 503 exit code is missing'
Need ($flat -match 'not counted') 'the 503 attempt accounting is missing'
Need ($flat -match 'reviews.*gate disputes.*defects') 'real defects caught by reviews and gate disputes are missing'
Need ($flat -match 'operator.*B42.*B43.*defects') 'the operator verification defects in B42 and B43 are missing'
Need ($flat -match 'B39.*\$8\.10') 'the B39 comparison is missing'

$b39 = Get-Content -Raw -LiteralPath 'docs/briefs/B20261120-1643-b39-local-config-copilot-only.loop-brief.md'
$runRecord = [regex]::Match($b39, '(?ms)^## Run record(?<body>.*?)(?=^## |\z)')
Need ($runRecord.Success -and $runRecord.Groups['body'].Value.Contains('$8.10')) 'B39''s run record does not contain $8.10'
