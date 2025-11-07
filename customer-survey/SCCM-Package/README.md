# Customer Survey Application - SCCM Deployment Package v2.0

## Package Contents

```
SCCM-Package/
├── customer-survey.exe    # Main application executable
├── config.json           # Configuration (webhook URL)
├── Install.ps1           # SCCM install script
├── Uninstall.ps1         # SCCM uninstall script
├── Detection.ps1         # SCCM detection script
└── README.md             # This file
```

## Deployment Overview

This package deploys the Customer Survey application to the **All Users Startup folder**. The exe automatically:
- Runs on every user login
- Checks per-user flags in %APPDATA%\CustomerSurvey
- Shows UI only when needed (no flags or expired reminder)
- Creates flags when user interacts (done.flag, nothanks.flag, remind.txt)
- Exits silently when flags indicate survey already handled

## SCCM Application Configuration

### Installation Program
```
powershell.exe -ExecutionPolicy Bypass -File ".\Install.ps1" -Silent
```

### Uninstallation Program
```
powershell.exe -ExecutionPolicy Bypass -File ".\Uninstall.ps1" -Silent
```

### Detection Method
**Registry Detection:**
- Key: `HKLM\SOFTWARE\CustomerSurvey`
- Value: `Installed`
- Type: `REG_DWORD`
- Data: `1`

Or use the Detection.ps1 script:
```
powershell.exe -ExecutionPolicy Bypass -File ".\Detection.ps1"
```

### Installation Behavior
- **Install for system**: Yes
- **Logon requirement**: Whether or not a user is logged on
- **Installation program visibility**: Hidden
- **Maximum allowed run time**: 15 minutes
- **Estimated installation time**: 1 minute

## Deployment Targets

- **Collection Type**: Device Collection
- **Target**: All Windows servers (dedicated and RDS/Citrix)
- **Deployment Type**: Required
- **Rerun Behavior**: Never rerun
- **User Experience**: Install in background, no user interaction

## File Locations After Installation

### System Files (deployed by SCCM)
- Exe: `%ProgramData%\Microsoft\Windows\Start Menu\Programs\StartUp\customer-survey.exe`
- Config: `%ProgramData%\Microsoft\Windows\Start Menu\Programs\StartUp\config.json`
- Registry: `HKLM\SOFTWARE\CustomerSurvey`

### Per-User Files (auto-created on first run)
- `%APPDATA%\CustomerSurvey\done.flag` - User completed survey
- `%APPDATA%\CustomerSurvey\nothanks.flag` - User opted out
- `%APPDATA%\CustomerSurvey\remind.txt` - User chose remind later (+7 days)

## How It Works

1. **SCCM deploys** exe to All Users Startup folder
2. **User logs in** → Windows launches exe automatically
3. **Exe checks flags** in user's %APPDATA%\CustomerSurvey:
   - If `done.flag` exists → Exit (no UI)
   - If `nothanks.flag` exists → Exit (no UI)
   - If `remind.txt` exists and date not passed → Exit (no UI)
   - Otherwise → Show survey UI
4. **User interacts** with survey → Appropriate flag created
5. **Next login** → Exe checks flags → Exits silently if already handled

## Configuration

### Before Deployment: Update config.json

Edit `config.json` with your Zoho webhook URL:

```json
{
  "zoho_webhook_url": "https://flow.zoho.in/YOUR-ACTUAL-WEBHOOK-URL"
}
```

**Important:** Replace the webhook URL with your production Zoho Flow endpoint!

## Pre-Deployment Checklist

- [ ] Built exe with `wails build` (production mode)
- [ ] Updated `config.json` with correct webhook URL
- [ ] Tested on pilot machines (5-10 servers)
- [ ] Code signed the exe (recommended to avoid AV alerts)
- [ ] Coordinated with SOC team for whitelisting
- [ ] SCCM package created and tested

## Deployment Steps

### 1. Create SCCM Application

1. Open SCCM Console → Software Library → Applications
2. Right-click → Create Application
3. Type: Script Installer
4. Content location: Point to this SCCM-Package folder
5. Install command: `powershell.exe -ExecutionPolicy Bypass -File ".\Install.ps1" -Silent`
6. Uninstall command: `powershell.exe -ExecutionPolicy Bypass -File ".\Uninstall.ps1" -Silent`

### 2. Configure Detection Method

Add custom detection rule:
- Registry key: `HKLM\SOFTWARE\CustomerSurvey`
- Value name: `Installed`
- Value equals: `1`

### 3. Deploy to Collection

1. Right-click application → Deploy
2. Select device collection (e.g., "All Windows Servers")
3. Purpose: Required
4. Schedule: As soon as possible (or maintenance window)
5. User experience: Install for system, hidden

### 4. Monitor Deployment

- Check deployment status in SCCM console
- Review logs: `C:\Windows\CCM\Logs\AppEnforce.log`
- Verify installation on sample machines

## Testing Before Production

### Test on 5-10 Pilot Servers

```powershell
# On pilot server, verify installation
Get-ItemProperty "HKLM:\SOFTWARE\CustomerSurvey"

# Check if exe exists in Startup
Get-Item "$env:ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp\customer-survey.exe"

# Test as a user - log in and check if survey shows
# Complete survey or click action
# Log in again - should not show

# Check user flag
Get-ChildItem "$env:APPDATA\CustomerSurvey"
```

## Rollout Strategy (Recommended)

### Phase 1: Pilot (Week 1)
- Deploy to 5-10 test servers
- Mix of dedicated and RDS servers
- Monitor closely

### Phase 2: Limited (Week 2)
- Deploy to 100 servers
- Monitor webhook submissions
- Check error rates

### Phase 3: Expanded (Week 3)
- Deploy to 500 servers
- Continue monitoring

### Phase 4: Full Rollout (Week 4)
- Deploy to remaining ~9,400 servers
- Monitor for 1 week
- Mark as production stable

## Monitoring

### SCCM Deployment Status
```
Software Library → Applications → Customer Survey
→ View Status → Deployment Status
```

### Per-Server Verification
```powershell
# Check installation
Get-ItemProperty "HKLM:\SOFTWARE\CustomerSurvey"

# Count user responses
$done = (Get-ChildItem "C:\Users\*\AppData\Roaming\CustomerSurvey\done.flag" -ErrorAction SilentlyContinue).Count
$noThanks = (Get-ChildItem "C:\Users\*\AppData\Roaming\CustomerSurvey\nothanks.flag" -ErrorAction SilentlyContinue).Count
$remind = (Get-ChildItem "C:\Users\*\AppData\Roaming\CustomerSurvey\remind.txt" -ErrorAction SilentlyContinue).Count

Write-Host "Completed: $done | No Thanks: $noThanks | Remind Later: $remind"
```

### Webhook Submissions
- Check Zoho Sheets for incoming responses
- Verify data format and completeness
- Monitor submission rate vs. user count

## Troubleshooting

### Issue: Exe not launching on login
- Check Startup folder: `shell:common startup`
- Verify exe exists and is not corrupted
- Check antivirus/AppLocker blocking

### Issue: Survey showing repeatedly
- Check if flag file is being created
- Verify %APPDATA% path is correct
- Check file permissions

### Issue: Survey not showing at all
- Check if flag incorrectly exists
- User can reset with: `customer-survey.exe -reset`
- Manually delete %APPDATA%\CustomerSurvey folder

### Issue: Webhook failing
- Check config.json has correct URL
- Verify network/firewall allows HTTPS to Zoho
- Review webhook.log in %APPDATA%\.customer-survey\

### Issue: Antivirus blocking
- Submit exe to SOC for whitelisting
- Code sign the exe
- Add exception in antivirus policy

## Rollback Procedure

If issues occur, uninstall via SCCM:

1. Change deployment from Required to Available
2. Or deploy Uninstall application
3. Or manually run:
```powershell
powershell.exe -ExecutionPolicy Bypass -File ".\Uninstall.ps1"
```

User data in %APPDATA%\CustomerSurvey is preserved during uninstall.

## Support

### Logs
- Install log: `%TEMP%\CustomerSurvey_Install.log`
- Uninstall log: `%TEMP%\CustomerSurvey_Uninstall.log`
- Webhook log: `%APPDATA%\.customer-survey\webhook.log`
- SCCM log: `C:\Windows\CCM\Logs\AppEnforce.log`

### Common User Questions

**Q: Survey showed once, now it's not appearing?**
A: That's correct behavior. Survey shows only once per user (unless they clicked "Remind Me Later").

**Q: I clicked "Remind Me Later" but want to complete now?**
A: Run: `customer-survey.exe -reset` from command line.

**Q: How to disable survey for a specific user?**
A: Create empty file: `%APPDATA%\CustomerSurvey\nothanks.flag`

## Package Information

- **Version**: 2.0.0
- **Build Date**: 2025-11-07
- **Platform**: Windows (x64)
- **Framework**: Wails v2 + Go
- **Dependencies**: None (standalone exe)
- **Minimum OS**: Windows 10 / Server 2016+

## Security & Compliance

- Exe contains embedded config (fallback)
- HTTPS webhook communication only
- Local backup in %LOCALAPPDATA%\Acesurvey.txt
- No privileged operations required
- Per-user data isolation
- No data collection beyond survey responses

## Contact

For deployment issues or questions:
- Check logs first
- Review troubleshooting section
- Contact IT Helpdesk with log files

---

**Package prepared by**: Aryan Gupta  
**Date**: November 7, 2025  
**Repository**: CustomerSurveyFeedback (branch: v2)
