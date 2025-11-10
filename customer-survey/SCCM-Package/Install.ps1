# SCCM Package Install Script
# Customer Survey Application v2.0
# 
# This script deploys the Customer Survey exe to the All Users Startup folder
# The exe will run automatically on every user login and manage flags internally

param(
    [switch]$Silent
)

$ErrorActionPreference = "Stop"

# Configuration
$startupFolder = "$env:ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$sourceExe = Join-Path $scriptDir "customer-survey.exe"
$sourceConfig = Join-Path $scriptDir "config.json"
$targetExe = Join-Path $startupFolder "customer-survey.exe"
$targetConfig = Join-Path $startupFolder "config.json"

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
    
    # Log to file for SCCM
    $logFile = "$env:TEMP\CustomerSurvey_Install.log"
    Add-Content -Path $logFile -Value $logMessage
}

try {
    Write-Log "=== Customer Survey Installation Started ==="
    
    # Validate source files
    if (-not (Test-Path $sourceExe)) {
        throw "Source exe not found: $sourceExe"
    }
    Write-Log "Source exe found: $sourceExe" "SUCCESS"
    
    if (-not (Test-Path $sourceConfig)) {
        Write-Log "Config file not found: $sourceConfig - Will create default" "WARNING"
        # Create minimal config if missing
        $defaultConfig = @{
            zoho_webhook_url = "https://your-webhook-url-here"
        } | ConvertTo-Json
        Set-Content -Path $sourceConfig -Value $defaultConfig
    }
    Write-Log "Source config found: $sourceConfig" "SUCCESS"
    
    # Copy exe to Startup folder
    Write-Log "Copying exe to Startup folder..."
    Copy-Item -Path $sourceExe -Destination $targetExe -Force
    Write-Log "Exe copied successfully" "SUCCESS"
    
    # Copy config to Startup folder
    Write-Log "Copying config to Startup folder..."
    Copy-Item -Path $sourceConfig -Destination $targetConfig -Force
    Write-Log "Config copied successfully" "SUCCESS"
    
    # Verify installation
    Write-Log "Verifying installation..."
    if (Test-Path $targetExe) {
        $exeInfo = Get-Item $targetExe
        Write-Log "Exe verified: $($exeInfo.Length) bytes" "SUCCESS"
    } else {
        throw "Exe verification failed"
    }
    
    if (Test-Path $targetConfig) {
        Write-Log "Config verified" "SUCCESS"
    }
    
    # Create registry detection key for SCCM
    Write-Log "Creating registry detection key..."
    $regPath = "HKLM:\SOFTWARE\CustomerSurvey"
    New-Item -Path $regPath -Force | Out-Null
    Set-ItemProperty -Path $regPath -Name "Installed" -Value 1 -Type DWord
    Set-ItemProperty -Path $regPath -Name "Version" -Value "2.0.0" -Type String
    Set-ItemProperty -Path $regPath -Name "InstallDate" -Value (Get-Date -Format "yyyy-MM-dd HH:mm:ss") -Type String
    Set-ItemProperty -Path $regPath -Name "DeploymentPath" -Value $targetExe -Type String
    Write-Log "Registry key created" "SUCCESS"
    
    Write-Log "=== Installation Completed Successfully ===" "SUCCESS"
    Write-Log "The survey will run automatically on next user login"
    
    exit 0
    
} catch {
    Write-Log "Installation failed: $_" "ERROR"
    exit 1
}
