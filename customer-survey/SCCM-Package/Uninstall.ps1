# SCCM Package Uninstall Script
# Customer Survey Application v2.0

param(
    [switch]$Silent
)

$ErrorActionPreference = "Stop"

# Configuration
$startupFolder = "$env:ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp"
$targetExe = Join-Path $startupFolder "customer-survey.exe"
$targetConfig = Join-Path $startupFolder "config.json"
$regPath = "HKLM:\SOFTWARE\CustomerSurvey"

function Write-Log {
    param($Message, $Level = "INFO")
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $logMessage = "[$timestamp] [$Level] $Message"
    
    if (-not $Silent) {
        switch ($Level) {
            "ERROR" { Write-Host $logMessage -ForegroundColor Red }
            "SUCCESS" { Write-Host $logMessage -ForegroundColor Green }
            "WARNING" { Write-Host $logMessage -ForegroundColor Yellow }
            default { Write-Host $logMessage }
        }
    }
    
    $logFile = "$env:TEMP\CustomerSurvey_Uninstall.log"
    Add-Content -Path $logFile -Value $logMessage
}

try {
    Write-Log "=== Customer Survey Uninstallation Started ==="
    
    # Remove exe from Startup
    if (Test-Path $targetExe) {
        Remove-Item -Path $targetExe -Force
        Write-Log "Exe removed from Startup" "SUCCESS"
    } else {
        Write-Log "Exe not found in Startup (already removed)" "WARNING"
    }
    
    # Remove config
    if (Test-Path $targetConfig) {
        Remove-Item -Path $targetConfig -Force
        Write-Log "Config removed from Startup" "SUCCESS"
    }
    
    # Remove registry key
    if (Test-Path $regPath) {
        Remove-Item -Path $regPath -Recurse -Force
        Write-Log "Registry key removed" "SUCCESS"
    }
    
    Write-Log "=== Uninstallation Completed Successfully ===" "SUCCESS"
    Write-Log "Note: User data in %APPDATA%\CustomerSurvey is preserved"
    
    exit 0
    
} catch {
    Write-Log "Uninstallation failed: $_" "ERROR"
    exit 1
}
