# Build script for Customer Survey Desktop Application
# Usage: .\build-desktop.ps1

Write-Host "Building Customer Survey Desktop Application..." -ForegroundColor Cyan

# Clean previous builds
if (Test-Path "customer-survey.exe") {
    Remove-Item "customer-survey.exe"
    Write-Host "Cleaned previous build" -ForegroundColor Yellow
}

# Build with Windows GUI mode (no console window)
# Embed the Zoho Flow webhook URL directly into the binary so no config file is needed
$WEBHOOK_URL = "https://flow.zoho.in/60006321785/flow/webhook/incoming?zapikey=1001.754e60b74ab20d6a1f255f55358ee47d.815d8c8feab82ae7a18f99777d41a05f&isdebug=false"
Write-Host "Compiling with embedded webhook URL..." -ForegroundColor Green

# Build smaller, statically linked binary
$env:CGO_ENABLED = "0"
go build -trimpath -tags "netgo" -ldflags "-s -w -H windowsgui -X 'customer-survey/internal/survey.DefaultWebhookURL=$WEBHOOK_URL'" -o customer-survey.exe .\cmd\survey\main.go

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n✅ Build successful!" -ForegroundColor Green
    Write-Host "   Output: customer-survey.exe" -ForegroundColor White
    # Show size
    $sizeMB = ([Math]::Round((Get-Item .\customer-survey.exe).Length / 1MB, 2))
    Write-Host ("   Size: {0} MB" -f $sizeMB) -ForegroundColor DarkGray

    # Optional: UPX compress if available on PATH and not disabled
    $upx = (Get-Command upx -ErrorAction SilentlyContinue)
    if ($upx -and $env:NO_UPX -ne "1") {
        Write-Host "Compressing with UPX (optional)..." -ForegroundColor Yellow
        & $upx.Source upx --best --lzma .\customer-survey.exe | Out-Null
        $sizeMB2 = ([Math]::Round((Get-Item .\customer-survey.exe).Length / 1MB, 2))
        Write-Host ("   Compressed Size: {0} MB" -f $sizeMB2) -ForegroundColor DarkGray
    } else {
        Write-Host "Skipping UPX compression (not found or disabled via NO_UPX=1)." -ForegroundColor Yellow
    }
    Write-Host "`nTo configure webhook, create config.json with:" -ForegroundColor Cyan
    Write-Host '   {"webhook_url": "https://script.google.com/macros/s/YOUR_SCRIPT_ID/exec"}' -ForegroundColor White
    Write-Host "`nOr set environment variable:" -ForegroundColor Cyan
    Write-Host '   $env:ZOHO_WEBHOOK_URL = "https://script.google.com/macros/s/YOUR_SCRIPT_ID/exec"' -ForegroundColor White
    Write-Host "`nRun the application:" -ForegroundColor Cyan
    Write-Host "   .\customer-survey.exe" -ForegroundColor White
} else {
    Write-Host "`n❌ Build failed!" -ForegroundColor Red
    exit 1
}
