param(
    [int]$Runs = 5
)

$ErrorActionPreference = "Stop"
$root = Resolve-Path (Join-Path $PSScriptRoot "../..")

Push-Location $root
try {
    Write-Host "执行 go test ./..."
    go test ./...

    for ($i = 1; $i -le $Runs; $i++) {
        Write-Host ("执行第 {0}/{1} 次 runtime smoke..." -f $i, $Runs)
        go run ./test/runtime_smoke
    }

    Write-Host ("issue #5 循环 smoke 验证通过，共执行 {0} 次。" -f $Runs)
}
finally {
    Pop-Location
}
