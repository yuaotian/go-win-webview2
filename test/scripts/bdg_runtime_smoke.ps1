param(
    [switch]$SkipGoTest
)

$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "../..")
Push-Location $repoRoot
try {
    if (-not $SkipGoTest) {
        Write-Host "[bdg] go test ./..."
        go test ./...
    }

    $targets = @(
        "./test/resource_stream_runtime_smoke",
        "./test/context_menu_runtime_smoke",
        "./test/print_runtime_smoke",
        "./test/resource_runtime_smoke",
        "./test/advanced_runtime_smoke"
    )

    foreach ($target in $targets) {
        Write-Host "[bdg] go run $target"
        go run $target
    }
}
finally {
    Pop-Location
}
