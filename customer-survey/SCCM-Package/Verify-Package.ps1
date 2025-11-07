# SCCM Package Verification Script
# Run this script to verify the package is ready for deployment

param([switch]$Detailed)

$ErrorActionPreference = "Stop"
$packagePath = $PSScriptRoot
$allChecksPassed = $true

Write-Host "======================================"  -ForegroundColor Cyan
Write-Host "SCCM Package Verification" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

# 1. Check all required files
Write-Host "1. Checking Package Files:" -ForegroundColor Yellow

$requiredFiles = @(
    @{Name="customer-survey.exe"; Desc="Main executable"},
    @{Name="config.json"; Desc="Configuration file"},
    @{Name="Install.ps1"; Desc="Installation script"},
    @{Name="Uninstall.ps1"; Desc="Uninstallation script"},
    @{Name="Detection.ps1"; Desc="Detection script"},
    @{Name="README.md"; Desc="Main documentation"},
    @{Name="DEPLOYMENT.md"; Desc="Quick reference guide"},
    @{Name="PACKAGE_MANIFEST.json"; Desc="Package manifest"}
)

$filesOk = $true
foreach ($file in $requiredFiles) {
    $filePath = Join-Path $packagePath $file.Name
    if (Test-Path $filePath) {
        Write-Host "[OK] $($file.Desc)" -ForegroundColor Green
        if ($Detailed) {
            $size = (Get-Item $filePath).Length
            $sizeKB = [math]::Round($size/1KB, 2)
            Write-Host "     Path: $filePath" -ForegroundColor Gray
            Write-Host "     Size: $sizeKB KB" -ForegroundColor Gray
        }
    } else {
        Write-Host "[FAIL] $($file.Desc) - MISSING" -ForegroundColor Red
        $filesOk = $false
    }
}
Write-Host ""

if (-not $filesOk) {
    $allChecksPassed = $false
}

# 2. Verify config.json content
Write-Host "2. Checking Configuration:" -ForegroundColor Yellow
$configPath = Join-Path $packagePath "config.json"
try {
    $config = Get-Content $configPath -Raw | ConvertFrom-Json
    
    if ($config.zoho_webhook_url) {
        if ($config.zoho_webhook_url -match "https://flow.zoho") {
            Write-Host "[OK] Webhook URL configured" -ForegroundColor Green
            if ($Detailed) {
                Write-Host "     URL: $($config.zoho_webhook_url)" -ForegroundColor Gray
            }
        } else {
            Write-Host "[WARN] Webhook URL may need updating" -ForegroundColor Yellow
            Write-Host "       Current: $($config.zoho_webhook_url)" -ForegroundColor Gray
        }
    } else {
        Write-Host "[FAIL] Webhook URL missing in config.json" -ForegroundColor Red
        $allChecksPassed = $false
    }
} catch {
    Write-Host "[FAIL] Failed to parse config.json" -ForegroundColor Red
    $allChecksPassed = $false
}
Write-Host ""

# 3. Check exe file properties
Write-Host "3. Checking Executable:" -ForegroundColor Yellow
$exePath = Join-Path $packagePath "customer-survey.exe"
if (Test-Path $exePath) {
    $exe = Get-Item $exePath
    $sizeInMB = [math]::Round($exe.Length / 1MB, 2)
    
    Write-Host "[OK] Exe size: $sizeInMB MB" -ForegroundColor Green
    
    if ($Detailed) {
        Write-Host "     Created: $($exe.CreationTime)" -ForegroundColor Gray
        Write-Host "     Modified: $($exe.LastWriteTime)" -ForegroundColor Gray
    }
    
    # Check if exe is blocked
    $zone = Get-Item $exePath -Stream Zone.Identifier -ErrorAction SilentlyContinue
    if ($zone) {
        Write-Host "[WARN] Exe is blocked by Windows" -ForegroundColor Yellow
        Write-Host "       Run: Unblock-File '$exePath'" -ForegroundColor Gray
    } else {
        Write-Host "[OK] Exe is not blocked" -ForegroundColor Green
    }
}
Write-Host ""

# 4. Test detection script
Write-Host "4. Testing Detection Script:" -ForegroundColor Yellow
$detectionPath = Join-Path $packagePath "Detection.ps1"
try {
    $null = & $detectionPath
    $detectionExitCode = $LASTEXITCODE
    
    if ($detectionExitCode -eq 1) {
        Write-Host "[OK] Detection script runs correctly (app not installed)" -ForegroundColor Green
    } elseif ($detectionExitCode -eq 0) {
        Write-Host "[WARN] Detection script reports app is already installed" -ForegroundColor Yellow
    } else {
        Write-Host "[FAIL] Detection script returned unexpected exit code: $detectionExitCode" -ForegroundColor Red
        $allChecksPassed = $false
    }
} catch {
    Write-Host "[FAIL] Detection script failed to run" -ForegroundColor Red
    $allChecksPassed = $false
}
Write-Host ""

# 5. Summary
Write-Host "======================================" -ForegroundColor Cyan
if ($allChecksPassed) {
    Write-Host "Package Verification: PASSED" -ForegroundColor Green
    Write-Host ""
    Write-Host "Package is ready for SCCM deployment!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next Steps:" -ForegroundColor Yellow
    Write-Host "1. Review config.json and update webhook URL if needed"
    Write-Host "2. Code sign the exe (recommended)"
    Write-Host "3. Copy this folder to SCCM content library"
    Write-Host "4. Create SCCM application using DEPLOYMENT.md guide"
    Write-Host "5. Deploy to pilot collection (5-10 servers)"
    Write-Host ""
    Write-Host "For detailed instructions, see README.md" -ForegroundColor Cyan
} else {
    Write-Host "Package Verification: FAILED" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please fix the issues above before deploying." -ForegroundColor Red
}
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

if ($allChecksPassed) {
    exit 0
} else {
    exit 1
}
