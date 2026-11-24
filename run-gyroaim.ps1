# Starts the SISR gyro-aim fork with a debug log, one file per run.
# Usage:  & D:\game-v3\SISR-fork\run-gyroaim.ps1            (extra SISR flags can be appended)
$dist = Join-Path $PSScriptRoot 'dist'
$logDir = Join-Path $dist 'logs'
New-Item -ItemType Directory -Force $logDir | Out-Null
$log = Join-Path $logDir ("sisr-{0:yyyyMMdd-HHmmss}.log" -f (Get-Date))
Write-Host "SISR log: $log"
Start-Process (Join-Path $dist 'SISR.exe') -ArgumentList (@("--log.file=$log", '--log.level=debug') + $args) -WorkingDirectory $dist
