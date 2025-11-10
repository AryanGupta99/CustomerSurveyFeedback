# Customer Survey Application - SCCM Detection Script
# Returns 0 (success) if application is installed, 1 (failure) if not installed
# Used by SCCM to detect if the application is present on the system

$ErrorActionPreference = "SilentlyContinue"

# Define detection criteria
$regPath = "HKLM:\SOFTWARE\CustomerSurvey"
$regValueName = "Installed"
$expectedValue = 1

$startupPath = "$env:ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp"
$exePath = Join-Path $startupPath "customer-survey.exe"

# Method 1: Check Registry Key
$regKeyExists = Test-Path $regPath
$regValue = $null

if ($regKeyExists) {
    $regValue = Get-ItemProperty -Path $regPath -Name $regValueName -ErrorAction SilentlyContinue
}

# Method 2: Check if exe exists in Startup folder
$exeExists = Test-Path $exePath

# Detection Logic
# Application is considered installed if BOTH conditions are met:
# 1. Registry key exists with correct value
# 2. Exe exists in Startup folder

if ($regKeyExists -and $regValue.$regValueName -eq $expectedValue -and $exeExists) {
    # Application is installed
    Write-Host "Customer Survey Application is installed"
    Write-Host "Registry: $regPath\$regValueName = $($regValue.$regValueName)"
    Write-Host "Exe: $exePath"
    exit 0
} else {
    # Application is not installed or partially installed
    Write-Host "Customer Survey Application is NOT installed"
    if (-not $regKeyExists) {
        Write-Host "Registry key missing: $regPath"
    } elseif ($regValue.$regValueName -ne $expectedValue) {
        Write-Host "Registry value incorrect: Expected $expectedValue, Got $($regValue.$regValueName)"
    }
    if (-not $exeExists) {
        Write-Host "Exe missing: $exePath"
    }
    exit 1
}
