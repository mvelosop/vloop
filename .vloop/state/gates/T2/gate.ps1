$ErrorActionPreference = "Stop"
$path = "docs/guide/concepts.md"
$text = (Get-Content -Raw $path) -replace '\s+', ' '
if ($text -notmatch 'Gates and the gate shell.*for a plan the next brief replaces, doctor checks the configured shell') {
    throw "concepts guide does not state that doctor checks the configured shell for a replaced plan"
}
