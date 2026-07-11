[CmdletBinding()]
param(
    [switch]$Restart,
    [switch]$RebuildBackend
)

$ErrorActionPreference = "Stop"
$root = $PSScriptRoot
$runtime = Join-Path $root ".runtime"
New-Item -ItemType Directory -Path $runtime -Force | Out-Null

# Some Windows launchers provide both Path and PATH. Start-Process treats them
# as duplicate dictionary keys, so normalize the process environment first.
$pathValue = $env:Path
Remove-Item Env:PATH -ErrorAction SilentlyContinue
[Environment]::SetEnvironmentVariable("Path", $pathValue, "Process")

function Test-TcpPort([int]$Port) {
    $client = [Net.Sockets.TcpClient]::new()
    try {
        $task = $client.ConnectAsync("127.0.0.1", $Port)
        return $task.Wait(300) -and $client.Connected
    }
    catch {
        return $false
    }
    finally {
        $client.Dispose()
    }
}

function Wait-Http([string]$Name, [string]$Url, [int]$TimeoutSeconds = 20) {
    $deadline = [DateTime]::UtcNow.AddSeconds($TimeoutSeconds)
    do {
        try {
            $response = Invoke-WebRequest -UseBasicParsing -Uri $Url -TimeoutSec 2
            if ($response.StatusCode -ge 200 -and $response.StatusCode -lt 400) {
                Write-Host "[ready] $Name ($($response.StatusCode))"
                return
            }
        }
        catch {
            Start-Sleep -Milliseconds 250
        }
    } while ([DateTime]::UtcNow -lt $deadline)

    throw "$Name did not become ready within $TimeoutSeconds seconds."
}

function Stop-OwnedProcess([string]$PidFile, [string]$ExpectedName) {
    if (-not (Test-Path -LiteralPath $PidFile)) { return }
    $savedPid = Get-Content -LiteralPath $PidFile -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($savedPid -notmatch '^\d+$') { return }
    $process = Get-Process -Id ([int]$savedPid) -ErrorAction SilentlyContinue
    if ($null -ne $process -and $process.ProcessName -eq $ExpectedName) {
        Stop-Process -Id $process.Id -Force
        $process.WaitForExit(5000) | Out-Null
    }
}

function Start-LoggedProcess(
    [string]$FilePath,
    [string[]]$ArgumentList,
    [string]$WorkingDirectory,
    [string]$LogName
) {
    $parameters = @{
        FilePath = $FilePath
        WorkingDirectory = $WorkingDirectory
        RedirectStandardOutput = Join-Path $root "$LogName.log"
        RedirectStandardError = Join-Path $root "$LogName.err.log"
        WindowStyle = "Hidden"
        PassThru = $true
    }
    if ($ArgumentList.Count -gt 0) {
        $parameters.ArgumentList = $ArgumentList
    }
    return Start-Process @parameters
}

if ($Restart) {
    Stop-OwnedProcess (Join-Path $runtime "frontend.pid") "node"
    Stop-OwnedProcess (Join-Path $runtime "backend.pid") "coai-dev"
    Start-Sleep -Milliseconds 300
}

if (Test-TcpPort 6379) {
    Write-Host "[reuse] Redis is already listening on 6379"
}
else {
    $redisExe = Join-Path $root "redis-portable\redis-server.exe"
    if (-not (Test-Path -LiteralPath $redisExe)) {
        throw "Redis is not running and $redisExe was not found."
    }
    $redis = Start-LoggedProcess $redisExe @() (Split-Path $redisExe) "redis"
    Set-Content -LiteralPath (Join-Path $runtime "redis.pid") -Value $redis.Id
    $deadline = [DateTime]::UtcNow.AddSeconds(10)
    while (-not (Test-TcpPort 6379) -and [DateTime]::UtcNow -lt $deadline) {
        Start-Sleep -Milliseconds 200
    }
    if (-not (Test-TcpPort 6379)) { throw "Redis failed to start; see redis.err.log." }
    Write-Host "[ready] Redis"
}

$backendExe = Join-Path $root "coai-dev.exe"
$backendInputs = @(
    Get-ChildItem -LiteralPath $root -File | Where-Object { $_.Extension -eq ".go" -or $_.Name -in @("go.mod", "go.sum") }
)
$excludedDirectories = @(".git", ".runtime", "app", "db", "logs", "redis-portable", "screenshot", "storage")
Get-ChildItem -LiteralPath $root -Directory |
    Where-Object { $_.Name -notin $excludedDirectories } |
    ForEach-Object { $backendInputs += Get-ChildItem -LiteralPath $_.FullName -Recurse -File -Filter "*.go" }

$backendIsStale = -not (Test-Path -LiteralPath $backendExe)
if (-not $backendIsStale) {
    $builtAt = (Get-Item -LiteralPath $backendExe).LastWriteTimeUtc
    $backendIsStale = $null -ne ($backendInputs | Where-Object { $_.LastWriteTimeUtc -gt $builtAt } | Select-Object -First 1)
}

if ($RebuildBackend -or $backendIsStale) {
    if (Test-TcpPort 8094) {
        throw "Backend sources changed, but port 8094 is already in use. Run .\start-dev.ps1 -Restart."
    }
    Write-Host "[build] Backend sources changed; compiling once..."
    Push-Location $root
    try {
        & go build -o $backendExe .
        if ($LASTEXITCODE -ne 0) { throw "go build failed with exit code $LASTEXITCODE." }
    }
    finally {
        Pop-Location
    }
}
else {
    Write-Host "[reuse] Backend binary is up to date"
}

if (Test-TcpPort 8094) {
    Write-Host "[reuse] Backend is already listening on 8094"
}
else {
    $backend = Start-LoggedProcess $backendExe @() $root "backend.dev"
    Set-Content -LiteralPath (Join-Path $runtime "backend.pid") -Value $backend.Id
}

$appRoot = Join-Path $root "app"
$viteEntry = Join-Path $appRoot "node_modules\vite\bin\vite.js"
if (-not (Test-Path -LiteralPath $viteEntry)) {
    throw "Frontend dependencies are missing. Run 'pnpm install --frozen-lockfile' once in $appRoot."
}
$nodeExe = (Get-Command node.exe -ErrorAction Stop).Source
if (Test-TcpPort 5173) {
    Write-Host "[reuse] Frontend is already listening on 5173"
}
else {
    $frontend = Start-LoggedProcess $nodeExe @(
        "node_modules\vite\bin\vite.js",
        "--host", "0.0.0.0",
        "--port", "5173",
        "--strictPort"
    ) $appRoot "frontend.dev"
    Set-Content -LiteralPath (Join-Path $runtime "frontend.pid") -Value $frontend.Id
}

Wait-Http "Backend" "http://127.0.0.1:8094/api/v1/state"
Wait-Http "Frontend" "http://127.0.0.1:5173/index.html"

Write-Host ""
Write-Host "Frontend: http://localhost:5173/"
Write-Host "Backend:  http://localhost:8094/"
Write-Host "Tip: use .\start-dev.ps1 -Restart after changing backend source."
