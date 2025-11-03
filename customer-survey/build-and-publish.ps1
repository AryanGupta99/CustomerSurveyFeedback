#!/usr/bin/env pwsh
# Build and Publish Customer Survey Application

Write-Host "================================" -ForegroundColor Cyan
Write-Host "  ACE Customer Survey Build" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""

# Stop running instances
Write-Host "Stopping running instances..." -ForegroundColor Yellow
Stop-Process -Name customer-survey,msedge,chrome -Force -ErrorAction SilentlyContinue
Start-Sleep -Seconds 1

# Build
Write-Host "Building application..." -ForegroundColor Yellow
go build -ldflags="-H=windowsgui" -o customer-survey.exe ./cmd/survey

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}

Write-Host "✅ Build successful!" -ForegroundColor Green
Write-Host ""

# Get file info
$exeInfo = Get-Item ".\customer-survey.exe"
Write-Host "Application Details:" -ForegroundColor Cyan
Write-Host "  Name: $($exeInfo.Name)"
Write-Host "  Size: $([math]::Round($exeInfo.Length / 1MB, 2)) MB"
Write-Host "  Location: $($exeInfo.FullName)"
Write-Host ""

# Publish
Write-Host "Publishing to release folder..." -ForegroundColor Yellow
$releaseDir = ".\release"
if (-not (Test-Path $releaseDir)) {
    New-Item -ItemType Directory -Path $releaseDir | Out-Null
}

Copy-Item ".\customer-survey.exe" "$releaseDir\customer-survey.exe" -Force
Write-Host "✅ Published to: $releaseDir\customer-survey.exe" -ForegroundColor Green
Write-Host ""

# Launch
Write-Host "Launching application..." -ForegroundColor Yellow
& ".\customer-survey.exe"

Write-Host ""
Write-Host "================================" -ForegroundColor Green
Write-Host "  Build Complete!" -ForegroundColor Green
Write-Host "================================" -ForegroundColor Green
