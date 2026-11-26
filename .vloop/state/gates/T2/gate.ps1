$ErrorActionPreference = 'Stop'

function Need([bool] $Condition, [string] $Message) {
    if (-not $Condition) {
        throw $Message
    }
}

$text = Get-Content -Raw -LiteralPath 'README.md'
$flat = $text -replace '\s+', ' '
$opening = ((Get-Content -LiteralPath 'README.md' -TotalCount 8) -join ' ') -replace '\s+', ' '
Need ($opening -match 'Go command-line tool.*plugin.*Claude Code.*GitHub Copilot CLI') 'the opening sentence does not name both agents'
Need ($opening -match 'agent plans the brief|agent plan the brief') 'the opening says Claude alone plans the brief'
Need (-not $text.Contains('vloop has Claude plan it')) 'the Claude-only planning claim remains'

Need ($flat -match '`copilot` CLI.*GitHub login') 'Copilot prerequisites are missing'
Need ($flat -match 'doctor.*agent CLIs in use|doctor.*`claude`.*`copilot`') 'quickstart does not describe both agent CLIs'
Need (-not $text.Contains('Every session is a fresh Claude session')) 'the Claude-only fresh-session sentence remains'
Need ($flat -match 'fresh agent session.*Claude Code.*Copilot CLI') 'the loop paragraph does not name both agent sessions'
Need ($flat -match 'Ask (your )?agent.*Claude Code.*Copilot CLI') 'the skill-driving paragraph does not address either agent'

Need ($flat -match 'mai-code-2') 'mai-code-2 is absent from the README'
Need ($flat -match 'URL shortener acceptance run') 'the mai-code-2 evidence does not name the URL shortener run'
Need ($flat -match 'B40.*B44') 'the mai-code-2 evidence does not name B40 to B44'
Need ($flat -match 'plan, gate review, work, review, verify and brief review') 'the mai-code-2 sentence omits a session kind'
Need ($flat -match 'not yet qualified') 'the mai-code-2 sentence does not state that qualification is pending'
Need ($flat -match 'copilot-cli-experiments\.md#mai-code-2-in-the-loop-2026-11') 'the mai-code-2 sentence does not link its evidence'
Need (-not $text.Contains('mai-code-2 is qualified')) 'mai-code-2 is incorrectly called qualified'
Need ($flat -match 'keep plan and gate review on Claude, and') 'the gpt-6.1-terra qualification remains ambiguous'
