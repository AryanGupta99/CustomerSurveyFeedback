# SCCM Deployment - Quick Reference

## Package Structure
```
SCCM-Package/
├── customer-survey.exe   → Main executable
├── config.json          → Webhook configuration
├── Install.ps1          → Installation script
├── Uninstall.ps1        → Removal script
├── Detection.ps1        → Detection script
├── README.md            → Full documentation
└── DEPLOYMENT.md        → This quick guide
```

## Quick Deploy (Copy-Paste Commands)

### 1. SCCM Application Settings

| Field | Value |
|-------|-------|
| **Install command** | `powershell.exe -ExecutionPolicy Bypass -File ".\Install.ps1" -Silent` |
| **Uninstall command** | `powershell.exe -ExecutionPolicy Bypass -File ".\Uninstall.ps1" -Silent` |
| **Detection method** | Registry: `HKLM\SOFTWARE\CustomerSurvey\Installed = 1` |
| **Install behavior** | Install for system |
| **Logon requirement** | Whether or not a user is logged on |
| **Maximum runtime** | 15 minutes |

### 2. Detection Rule Configuration

**Detection Type:** Custom Script  
**Script Type:** PowerShell  
**Script:** Use `Detection.ps1` from package

OR

**Detection Type:** Registry  
- Hive: `HKEY_LOCAL_MACHINE`
- Key: `SOFTWARE\CustomerSurvey`
- Value: `Installed`
- Type: `REG_DWORD`
- Operator: `Equals`
- Value: `1`

### 3. Deployment Settings

```
Purpose: Required
Action: Install
Collection: [Your Device Collection]
Schedule: As soon as possible
User Experience: Hidden
Rerun behavior: Never rerun
```

## Manual Test Commands

### Install Manually
```powershell
cd "\\network\share\SCCM-Package"
.\Install.ps1 -Silent
```

### Verify Installation
```powershell
# Check registry
Get-ItemProperty "HKLM:\SOFTWARE\CustomerSurvey"

# Check exe
Get-Item "$env:ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp\customer-survey.exe"
```

### Check Detection
```powershell
.\Detection.ps1
echo $LASTEXITCODE  # Should be 0 if installed
```

### Uninstall Manually
```powershell
.\Uninstall.ps1 -Silent
```

## Deployment Timeline

| Phase | Week | Servers | Action |
|-------|------|---------|--------|
| Pilot | 1 | 5-10 | Test deployment, monitor closely |
| Limited | 2 | 100 | Initial rollout, verify stability |
| Expanded | 3 | 500 | Broader rollout |
| Full | 4 | All (~10k) | Complete deployment |

## Monitoring Commands

### SCCM Console
```
Software Library → Applications → Customer Survey 
→ Right-click → View Status → Deployment Status
```

### PowerShell Check on Target Server
```powershell
# Check if installed
Get-ItemProperty "HKLM:\SOFTWARE\CustomerSurvey" -ErrorAction SilentlyContinue

# Count user interactions
$appData = "C:\Users\*\AppData\Roaming\CustomerSurvey"
$done = @(Get-ChildItem "$appData\done.flag" -ErrorAction SilentlyContinue).Count
$noThanks = @(Get-ChildItem "$appData\nothanks.flag" -ErrorAction SilentlyContinue).Count
$remind = @(Get-ChildItem "$appData\remind.txt" -ErrorAction SilentlyContinue).Count

Write-Host "Completed: $done | No Thanks: $noThanks | Remind Later: $remind"
```

## Common Issues

| Issue | Solution |
|-------|----------|
| Exe not in Startup | Check SCCM deployment status, verify install script ran |
| Survey not showing | Flag exists, have user delete `%APPDATA%\CustomerSurvey` |
| Survey shows every login | Flag not being created, check permissions on %APPDATA% |
| Webhook failing | Verify config.json has correct URL, check firewall |

## Log Locations

| Log | Path |
|-----|------|
| Install | `%TEMP%\CustomerSurvey_Install.log` |
| Uninstall | `%TEMP%\CustomerSurvey_Uninstall.log` |
| SCCM | `C:\Windows\CCM\Logs\AppEnforce.log` |
| Webhook | `%APPDATA%\.customer-survey\webhook.log` |

## Files Deployed

### System-wide (All Users)
- `%ProgramData%\Microsoft\Windows\Start Menu\Programs\StartUp\customer-survey.exe`
- `%ProgramData%\Microsoft\Windows\Start Menu\Programs\StartUp\config.json`
- `HKLM\SOFTWARE\CustomerSurvey` (registry)

### Per-User (Auto-created)
- `%APPDATA%\CustomerSurvey\done.flag`
- `%APPDATA%\CustomerSurvey\nothanks.flag`
- `%APPDATA%\CustomerSurvey\remind.txt`

## Support Escalation

1. Check log files
2. Verify detection script returns 0
3. Test manual install
4. Check antivirus exclusions
5. Contact application owner

---

**For detailed documentation, see README.md**  
**Version**: 2.0.0 | **Date**: 2025-11-07
