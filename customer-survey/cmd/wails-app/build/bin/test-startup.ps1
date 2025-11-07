# Test script for Customer Survey startup logic
Write-Host "=== Customer Survey Startup Logic Test ===" -ForegroundColor Cyan
Write-Host ""

$appDataDir = Join-Path $env:APPDATA "CustomerSurvey"
$doneFlag = Join-Path $appDataDir "done.flag"
$noThanksFlag = Join-Path $appDataDir "nothanks.flag"
$remindFile = Join-Path $appDataDir "remind.txt"
$exePath = ".\customer-survey.exe"

# Function to check and display file status
function Show-Status {
    Write-Host "`nCurrent Status:" -ForegroundColor Yellow
    Write-Host "  AppData folder: $appDataDir"
    Write-Host "  Folder exists: $(Test-Path $appDataDir)"
    
    if (Test-Path $doneFlag) {
        $content = Get-Content $doneFlag -Raw
        Write-Host "  ✓ done.flag exists (created: $content)" -ForegroundColor Green
    } else {
        Write-Host "  ✗ done.flag does not exist" -ForegroundColor Gray
    }
    
    if (Test-Path $noThanksFlag) {
        $content = Get-Content $noThanksFlag -Raw
        Write-Host "  ✓ nothanks.flag exists (created: $content)" -ForegroundColor Green
    } else {
        Write-Host "  ✗ nothanks.flag does not exist" -ForegroundColor Gray
    }
    
    if (Test-Path $remindFile) {
        $content = Get-Content $remindFile -Raw
        Write-Host "  ✓ remind.txt exists (remind until: $content)" -ForegroundColor Green
    } else {
        Write-Host "  ✗ remind.txt does not exist" -ForegroundColor Gray
    }
}

# Test 1: Clean start
Write-Host "`n[TEST 1] Clean Start - No flags" -ForegroundColor Cyan
if (Test-Path $appDataDir) {
    Remove-Item -Path $appDataDir -Recurse -Force
}
Show-Status
Write-Host "`nExpected: Survey should show (all files missing)" -ForegroundColor Yellow

# Test 2: Create done.flag
Write-Host "`n`n[TEST 2] Simulating Survey Completion" -ForegroundColor Cyan
New-Item -ItemType Directory -Path $appDataDir -Force | Out-Null
$timestamp = Get-Date -Format "o"
Set-Content -Path $doneFlag -Value $timestamp
Show-Status
Write-Host "`nExpected: Survey should NOT show (done.flag exists)" -ForegroundColor Yellow

# Test 3: Reset and test nothanks.flag
Write-Host "`n`n[TEST 3] Simulating No Thanks" -ForegroundColor Cyan
Remove-Item -Path $appDataDir -Recurse -Force
New-Item -ItemType Directory -Path $appDataDir -Force | Out-Null
$timestamp = Get-Date -Format "o"
Set-Content -Path $noThanksFlag -Value $timestamp
Show-Status
Write-Host "`nExpected: Survey should NOT show (nothanks.flag exists)" -ForegroundColor Yellow

# Test 4: Reset and test remind.txt (within window)
Write-Host "`n`n[TEST 4] Simulating Remind Me Later (within 7 days)" -ForegroundColor Cyan
Remove-Item -Path $appDataDir -Recurse -Force
New-Item -ItemType Directory -Path $appDataDir -Force | Out-Null
$futureDate = (Get-Date).AddDays(7).ToString("o")
Set-Content -Path $remindFile -Value $futureDate
Show-Status
Write-Host "`nExpected: Survey should NOT show (within remind window)" -ForegroundColor Yellow

# Test 5: Reset and test remind.txt (past window)
Write-Host "`n`n[TEST 5] Simulating Remind Me Later (past 7 days)" -ForegroundColor Cyan
Remove-Item -Path $appDataDir -Recurse -Force
New-Item -ItemType Directory -Path $appDataDir -Force | Out-Null
$pastDate = (Get-Date).AddDays(-1).ToString("o")
Set-Content -Path $remindFile -Value $pastDate
Show-Status
Write-Host "`nExpected: Survey SHOULD show (remind window expired)" -ForegroundColor Yellow

# Test 6: Final cleanup
Write-Host "`n`n[TEST 6] Cleanup - Remove all test files" -ForegroundColor Cyan
Remove-Item -Path $appDataDir -Recurse -Force
Show-Status

Write-Host "`n`n=== Tests Complete ===" -ForegroundColor Green
Write-Host "You can now test the actual exe:" -ForegroundColor Yellow
Write-Host "  1. Run: .\customer-survey.exe         (should show survey)" -ForegroundColor White
Write-Host "  2. Complete survey or click action    (creates flag)" -ForegroundColor White
Write-Host "  3. Run again: .\customer-survey.exe   (should exit silently)" -ForegroundColor White
Write-Host "  4. Reset: .\customer-survey.exe -reset (removes flags and shows survey)" -ForegroundColor White
Write-Host ""
