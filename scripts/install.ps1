param(
    [int]$AssetId = 0,
    [string]$Endpoint = "http://CHANGE_ME_SERVER_IP:8080/api/v1/collect",
    [int]$Interval = 30
)

if ($AssetId -eq 0) {
    Write-Host "ERROR: -AssetId is required" -ForegroundColor Red
    exit 1
}

Write-Host "=== ITSM Ops Collector Installer (Windows) ===" -ForegroundColor Cyan
Write-Host "Asset ID: $AssetId"
Write-Host "Endpoint: $Endpoint"

# Download
Write-Host "Downloading collector..."
New-Item -ItemType Directory -Force -Path "C:\ITSM-Agent" | Out-Null
Invoke-WebRequest -Uri "http://CHANGE_ME_SERVER_IP:3000/downloads/itsm-collector.exe" -OutFile "C:\ITSM-Agent\itsm-collector.exe"

# Create start.bat wrapper
@"
@echo off
set COLLECTOR_ENDPOINT=$Endpoint
set COLLECTOR_ASSET_ID=$AssetId
set COLLECTOR_INTERVAL=$Interval
set COLLECTOR_MODE=auto
"C:\ITSM-Agent\itsm-collector.exe"
"@ | Set-Content "C:\ITSM-Agent\start.bat" -Encoding ASCII

# Schedule task
schtasks /delete /tn "ITSM-Collector" /f 2>$null
schtasks /create /tn "ITSM-Collector" /tr "'C:\ITSM-Agent\start.bat'" /sc onstart /ru SYSTEM /rl HIGHEST /f
schtasks /run /tn "ITSM-Collector"

Write-Host "=== Installation complete! ===" -ForegroundColor Green
Write-Host "Binary: C:\ITSM-Agent\itsm-collector.exe"
Write-Host "Wrapper: C:\ITSM-Agent\start.bat"
Write-Host "Task: ITSM-Collector (auto-start on boot)"
