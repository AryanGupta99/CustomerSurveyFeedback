# Customer Survey - Production Deployment Script
# Run this as Administrator on target servers

param(
    [switch]$Install,
    [switch]$Uninstall,
    [switch]$Verify,
    [string]$ExePath = "",
    [string]$ConfigPath = ""
)

$ErrorActionPreference = "Stop"
$installLocation = "C:\Program Files\CustomerSurvey"
$startupLocation = "$env:ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp"
$shortcutName = "CustomerSurvey.lnk"

function Write-Log {
    param($Message, $Color = "White")
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Host "[$timestamp] $Message" -ForegroundColor $Color
}

function Install-CustomerSurvey {
    Write-Log "=== Starting Customer Survey Installation ===" "Cyan"
    
    # Check if running as admin
    $isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    if (-not $isAdmin) {
        Write-Log "ERROR: Must run as Administrator!" "Red"
        exit 1
    }
    
    # Validate source files
    if (-not $ExePath -or -not (Test-Path $ExePath)) {
        Write-Log "ERROR: ExePath not specified or file not found: $ExePath" "Red"
        exit 1
    }
    
    if (-not $ConfigPath -or -not (Test-Path $ConfigPath)) {
        Write-Log "ERROR: ConfigPath not specified or file not found: $ConfigPath" "Red"
        exit 1
    }
    
    Write-Log "Source EXE: $ExePath" "Yellow"
    Write-Log "Source Config: $ConfigPath" "Yellow"
    
    # Create installation directory
    Write-Log "Creating installation directory: $installLocation"
    if (-not (Test-Path $installLocation)) {
        New-Item -Path $installLocation -ItemType Directory -Force | Out-Null
    }
    
    # Copy files
    Write-Log "Copying executable..."
    Copy-Item -Path $ExePath -Destination "$installLocation\customer-survey.exe" -Force
    
    Write-Log "Copying config.json..."
    Copy-Item -Path $ConfigPath -Destination "$installLocation\config.json" -Force
    
    # Create All Users startup shortcut
    Write-Log "Creating startup shortcut for all users..."
    $shortcutPath = Join-Path $startupLocation $shortcutName
    
    $WshShell = New-Object -ComObject WScript.Shell
    $Shortcut = $WshShell.CreateShortcut($shortcutPath)
    $Shortcut.TargetPath = "$installLocation\customer-survey.exe"
    $Shortcut.WorkingDirectory = $installLocation
    $Shortcut.Description = "Customer Survey Application"
    $Shortcut.Save()
    
    Write-Log "✓ Shortcut created: $shortcutPath" "Green"
    
    # Set registry key for detection (SCCM)
    Write-Log "Setting registry detection key..."
    $regPath = "HKLM:\SOFTWARE\CustomerSurvey"
    if (-not (Test-Path $regPath)) {
        New-Item -Path $regPath -Force | Out-Null
    }
    Set-ItemProperty -Path $regPath -Name "Installed" -Value 1 -Type DWord
    Set-ItemProperty -Path $regPath -Name "Version" -Value "2.0.0" -Type String
    Set-ItemProperty -Path $regPath -Name "InstallPath" -Value $installLocation -Type String
    Set-ItemProperty -Path $regPath -Name "InstallDate" -Value (Get-Date -Format "yyyy-MM-dd HH:mm:ss") -Type String
    
    Write-Log "✓ Registry keys created" "Green"
    
    # Verify installation
    Write-Log "`nVerifying installation..."
    if (Test-Path "$installLocation\customer-survey.exe") {
        $exeInfo = Get-Item "$installLocation\customer-survey.exe"
        Write-Log "✓ EXE installed: $($exeInfo.Length) bytes" "Green"
    }
    
    if (Test-Path "$installLocation\config.json") {
        Write-Log "✓ Config installed" "Green"
    }
    
    if (Test-Path $shortcutPath) {
        Write-Log "✓ Startup shortcut created" "Green"
    }
    
    Write-Log "`n=== Installation Complete ===" "Cyan"
    Write-Log "Install location: $installLocation" "Yellow"
    Write-Log "Startup shortcut: $shortcutPath" "Yellow"
    Write-Log "`nNext user login will trigger the survey." "Green"
}

function Uninstall-CustomerSurvey {
    Write-Log "=== Starting Customer Survey Uninstallation ===" "Cyan"
    
    # Remove startup shortcut
    $shortcutPath = Join-Path $startupLocation $shortcutName
    if (Test-Path $shortcutPath) {
        Remove-Item $shortcutPath -Force
        Write-Log "✓ Removed startup shortcut" "Green"
    }
    
    # Remove installation directory
    if (Test-Path $installLocation) {
        Remove-Item $installLocation -Recurse -Force
        Write-Log "✓ Removed installation directory" "Green"
    }
    
    # Remove registry keys
    $regPath = "HKLM:\SOFTWARE\CustomerSurvey"
    if (Test-Path $regPath) {
        Remove-Item $regPath -Recurse -Force
        Write-Log "✓ Removed registry keys" "Green"
    }
    
    Write-Log "`n=== Uninstallation Complete ===" "Cyan"
    Write-Log "Note: User data in %APPDATA%\CustomerSurvey is preserved" "Yellow"
    Write-Log "To remove user data, manually delete from each user profile" "Yellow"
}

function Verify-Installation {
    Write-Log "=== Verifying Installation ===" "Cyan"
    
    $allGood = $true
    
    # Check exe
    if (Test-Path "$installLocation\customer-survey.exe") {
        $exeInfo = Get-Item "$installLocation\customer-survey.exe"
        Write-Log "✓ EXE exists: $($exeInfo.Length) bytes, modified $($exeInfo.LastWriteTime)" "Green"
    } else {
        Write-Log "✗ EXE not found!" "Red"
        $allGood = $false
    }
    
    # Check config
    if (Test-Path "$installLocation\config.json") {
        Write-Log "✓ Config exists" "Green"
        $config = Get-Content "$installLocation\config.json" -Raw | ConvertFrom-Json
        if ($config.zoho_webhook_url) {
            Write-Log "  Webhook URL configured: $($config.zoho_webhook_url.Substring(0, 50))..." "Yellow"
        }
    } else {
        Write-Log "✗ Config not found!" "Red"
        $allGood = $false
    }
    
    # Check shortcut
    $shortcutPath = Join-Path $startupLocation $shortcutName
    if (Test-Path $shortcutPath) {
        Write-Log "✓ Startup shortcut exists" "Green"
    } else {
        Write-Log "✗ Startup shortcut not found!" "Red"
        $allGood = $false
    }
    
    # Check registry
    $regPath = "HKLM:\SOFTWARE\CustomerSurvey"
    if (Test-Path $regPath) {
        $regData = Get-ItemProperty $regPath
        Write-Log "✓ Registry key exists" "Green"
        Write-Log "  Version: $($regData.Version)" "Yellow"
        Write-Log "  Install Date: $($regData.InstallDate)" "Yellow"
    } else {
        Write-Log "✗ Registry key not found!" "Red"
        $allGood = $false
    }
    
    # Check user data (how many users have responded)
    Write-Log "`nUser Response Statistics:"
    $users = Get-ChildItem "C:\Users" -Directory -ErrorAction SilentlyContinue | Where-Object { $_.Name -notmatch '^(Public|Default|All Users)$' }
    $totalUsers = $users.Count
    $doneCount = 0
    $noThanksCount = 0
    $remindCount = 0
    
    foreach ($user in $users) {
        $userAppData = "C:\Users\$($user.Name)\AppData\Roaming\CustomerSurvey"
        if (Test-Path $userAppData) {
            if (Test-Path "$userAppData\done.flag") { $doneCount++ }
            if (Test-Path "$userAppData\nothanks.flag") { $noThanksCount++ }
            if (Test-Path "$userAppData\remind.txt") { $remindCount++ }
        }
    }
    
    Write-Log "  Total users: $totalUsers" "Yellow"
    Write-Log "  Completed survey: $doneCount" "Green"
    Write-Log "  Opted out (No Thanks): $noThanksCount" "Cyan"
    Write-Log "  Remind me later: $remindCount" "Magenta"
    Write-Log "  Not yet responded: $($totalUsers - $doneCount - $noThanksCount - $remindCount)" "Gray"
    
    if ($allGood) {
        Write-Log "`n=== Installation Verified Successfully ===" "Green"
        exit 0
    } else {
        Write-Log "`n=== Installation Verification FAILED ===" "Red"
        exit 1
    }
}

# Main execution
if ($Install) {
    Install-CustomerSurvey
} elseif ($Uninstall) {
    Uninstall-CustomerSurvey
} elseif ($Verify) {
    Verify-Installation
} else {
    Write-Host "Customer Survey Deployment Script"
    Write-Host ""
    Write-Host "Usage:"
    Write-Host "  Install:   .\deploy-production.ps1 -Install -ExePath 'path\to\customer-survey.exe' -ConfigPath 'path\to\config.json'"
    Write-Host "  Uninstall: .\deploy-production.ps1 -Uninstall"
    Write-Host "  Verify:    .\deploy-production.ps1 -Verify"
    Write-Host ""
    Write-Host "Example:"
    Write-Host "  .\deploy-production.ps1 -Install -ExePath '.\build\bin\customer-survey.exe' -ConfigPath '.\config.json'"
}
