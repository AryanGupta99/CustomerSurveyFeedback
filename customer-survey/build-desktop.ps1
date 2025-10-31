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

    # Skip UPX and code signing to avoid Crowdstrike detection
    Write-Host "Skipping optional compression and signing for Crowdstrike compatibility..." -ForegroundColor Yellow
    
    Write-Host "`n⚠️  NOTE: If Crowdstrike still blocks this:" -ForegroundColor Yellow
    Write-Host "   1. Contact your IT team to whitelist this exe in Crowdstrike" -ForegroundColor White
    Write-Host "   2. Or request them to add an exclusion for ACH Customer Survey app" -ForegroundColor White
    Write-Host "   3. Application hash: $(certUtil -hashfile '.\customer-survey.exe' SHA256 | findstr /v 'SHA256' | findstr /v 'CertUtil')" -ForegroundColor Gray
    
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
