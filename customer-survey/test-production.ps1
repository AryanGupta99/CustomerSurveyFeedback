# Quick Production Test Script
# Tests the application in a production-like scenario on your local machine

Write-Host "=== Customer Survey - Production Simulation Test ===" -ForegroundColor Cyan
Write-Host ""

$exePath = "c:\Users\aryan.gupta\OneDrive - Real Time Data Services Pvt Ltd\Desktop\Customer-Survey-Application\customer-survey\cmd\wails-app\build\bin\customer-survey.exe"
$configPath = "c:\Users\aryan.gupta\OneDrive - Real Time Data Services Pvt Ltd\Desktop\Customer-Survey-Application\customer-survey\cmd\wails-app\config.json"

# Check if files exist
if (-not (Test-Path $exePath)) {
    Write-Host "ERROR: Exe not found at: $exePath" -ForegroundColor Red
    Write-Host "Please build the application first with: wails build" -ForegroundColor Yellow
    exit
}

if (-not (Test-Path $configPath)) {
    Write-Host "WARNING: Config not found at: $configPath" -ForegroundColor Yellow
}

# Test 1: Clean State - First Run
Write-Host "`n[Test 1] Clean State - First User Login Simulation" -ForegroundColor Yellow
Write-Host "  Cleaning user state..." -ForegroundColor Gray
Remove-Item "$env:APPDATA\CustomerSurvey" -Recurse -Force -ErrorAction SilentlyContinue

Write-Host "  Launching application (should show UI)..." -ForegroundColor Gray
Start-Process -FilePath $exePath -Wait

Write-Host "`n  Checking what files were created..." -ForegroundColor Gray
if (Test-Path "$env:APPDATA\CustomerSurvey") {
    Get-ChildItem "$env:APPDATA\CustomerSurvey" | Format-Table Name, Length, LastWriteTime
    
    foreach ($file in Get-ChildItem "$env:APPDATA\CustomerSurvey") {
        Write-Host "  File: $($file.Name)" -ForegroundColor Cyan
        $content = Get-Content $file.FullName -Raw
        Write-Host "  Content: $content" -ForegroundColor White
    }
} else {
    Write-Host "  No files created (user may have closed without action)" -ForegroundColor Yellow
}

Read-Host "`nPress Enter to continue to Test 2"

# Test 2: Second Run - Should Exit Silently
Write-Host "`n[Test 2] Second Login - Should Exit Silently" -ForegroundColor Yellow
Write-Host "  Current state:" -ForegroundColor Gray
if (Test-Path "$env:APPDATA\CustomerSurvey") {
    Get-ChildItem "$env:APPDATA\CustomerSurvey" | Format-Table Name, LastWriteTime
} else {
    Write-Host "  No state files found" -ForegroundColor Red
}

Write-Host "`n  Launching application again..." -ForegroundColor Gray
Write-Host "  Expected: Should exit immediately without showing UI" -ForegroundColor Yellow

$proc = Start-Process -FilePath $exePath -PassThru
Start-Sleep -Seconds 2

if ($proc.HasExited) {
    Write-Host "  ✓ Application exited quickly (correct behavior)" -ForegroundColor Green
} else {
    Write-Host "  ✗ Application still running (UI may have shown)" -ForegroundColor Red
}

Read-Host "`nPress Enter to continue to Test 3"

# Test 3: Reset Functionality
Write-Host "`n[Test 3] Reset Functionality Test" -ForegroundColor Yellow
Write-Host "  Running with -reset flag..." -ForegroundColor Gray

Start-Process -FilePath $exePath -ArgumentList "-reset" -Wait

Write-Host "  Checking state after reset..." -ForegroundColor Gray
if (Test-Path "$env:APPDATA\CustomerSurvey") {
    $files = Get-ChildItem "$env:APPDATA\CustomerSurvey"
    if ($files.Count -eq 0) {
        Write-Host "  ✓ All flags removed (correct)" -ForegroundColor Green
    } else {
        Write-Host "  Files still exist:" -ForegroundColor Yellow
        $files | Format-Table Name, LastWriteTime
    }
} else {
    Write-Host "  ✓ Folder removed or empty (correct)" -ForegroundColor Green
}

Read-Host "`nPress Enter to continue to Test 4"

# Test 4: Remind Me Later Simulation
Write-Host "`n[Test 4] Remind Me Later Simulation" -ForegroundColor Yellow
Write-Host "  Creating remind.txt with future date..." -ForegroundColor Gray

New-Item -Path "$env:APPDATA\CustomerSurvey" -ItemType Directory -Force | Out-Null
$futureDate = (Get-Date).AddDays(7).ToString("o")
Set-Content "$env:APPDATA\CustomerSurvey\remind.txt" -Value $futureDate

Write-Host "  Remind date set to: $futureDate" -ForegroundColor Cyan

Write-Host "`n  Launching application..." -ForegroundColor Gray
Write-Host "  Expected: Should exit immediately (within remind window)" -ForegroundColor Yellow

$proc = Start-Process -FilePath $exePath -PassThru
Start-Sleep -Seconds 2

if ($proc.HasExited) {
    Write-Host "  ✓ Application exited (correct - remind window active)" -ForegroundColor Green
} else {
    Write-Host "  ✗ Application showing UI (incorrect)" -ForegroundColor Red
}

Read-Host "`nPress Enter to continue to Test 5"

# Test 5: Expired Reminder
Write-Host "`n[Test 5] Expired Reminder Simulation" -ForegroundColor Yellow
Write-Host "  Creating remind.txt with PAST date..." -ForegroundColor Gray

$pastDate = (Get-Date).AddDays(-1).ToString("o")
Set-Content "$env:APPDATA\CustomerSurvey\remind.txt" -Value $pastDate

Write-Host "  Remind date set to: $pastDate (expired)" -ForegroundColor Cyan

Write-Host "`n  Launching application..." -ForegroundColor Gray
Write-Host "  Expected: Should show UI (reminder expired)" -ForegroundColor Yellow

Start-Process -FilePath $exePath -Wait

Write-Host "`n  Check if new flag was created..." -ForegroundColor Gray
Get-ChildItem "$env:APPDATA\CustomerSurvey" | Format-Table Name, LastWriteTime

Read-Host "`nPress Enter to see final summary"

# Summary
Write-Host "`n=== Test Summary ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Tests Completed:" -ForegroundColor Green
Write-Host "  ✓ Test 1: First run (clean state)"
Write-Host "  ✓ Test 2: Second run (flag exists)"
Write-Host "  ✓ Test 3: Reset functionality"
Write-Host "  ✓ Test 4: Remind me later (active window)"
Write-Host "  ✓ Test 5: Remind me later (expired)"
Write-Host ""
Write-Host "Current user state:" -ForegroundColor Yellow
if (Test-Path "$env:APPDATA\CustomerSurvey") {
    Get-ChildItem "$env:APPDATA\CustomerSurvey" | Format-Table Name, Length, LastWriteTime
} else {
    Write-Host "  No state files" -ForegroundColor Gray
}
Write-Host ""
Write-Host "Next Steps for Production:" -ForegroundColor Cyan
Write-Host "  1. Deploy to a test server with: deploy-production.ps1 -Install" -ForegroundColor White
Write-Host "  2. Test with multiple user accounts on RDS" -ForegroundColor White
Write-Host "  3. Deploy to pilot group (5-10 servers)" -ForegroundColor White
Write-Host "  4. Monitor and validate webhook submissions" -ForegroundColor White
Write-Host "  5. Full rollout via SCCM" -ForegroundColor White
Write-Host ""

# Cleanup option
$cleanup = Read-Host "Clean up test files? (Y/N)"
if ($cleanup -eq "Y" -or $cleanup -eq "y") {
    Remove-Item "$env:APPDATA\CustomerSurvey" -Recurse -Force -ErrorAction SilentlyContinue
    Write-Host "✓ Test files cleaned up" -ForegroundColor Green
}
