# SCCM Deployment Package - READY FOR DEPLOYMENT

## Package Verification Results

**Status:** ✅ PASSED  
**Date:** November 7, 2025  
**Package Version:** 2.0.0

## Package Contents Summary

| File | Size | Status |
|------|------|--------|
| customer-survey.exe | 10.09 MB | ✅ Ready |
| config.json | 0.18 KB | ✅ Configured |
| Install.ps1 | 3.61 KB | ✅ Verified |
| Uninstall.ps1 | 1.93 KB | ✅ Verified |
| Detection.ps1 | 1.73 KB | ✅ Working |
| README.md | 8.86 KB | ✅ Complete |
| DEPLOYMENT.md | 4.39 KB | ✅ Complete |
| PACKAGE_MANIFEST.json | 6.25 KB | ✅ Complete |

**Total Package Size:** ~10.3 MB

## Configuration Verified

- **Webhook URL:** Configured and valid (Zoho Flow endpoint)
- **Exe Status:** Not blocked by Windows
- **Detection Script:** Working correctly
- **Build Date:** November 7, 2025 19:03:01

## SCCM Application Quick Setup

### Installation Command
```
powershell.exe -ExecutionPolicy Bypass -File ".\Install.ps1" -Silent
```

### Uninstallation Command
```
powershell.exe -ExecutionPolicy Bypass -File ".\Uninstall.ps1" -Silent
```

### Detection Method
**Registry Detection:**
- Path: `HKLM:\SOFTWARE\CustomerSurvey`
- Value Name: `Installed`
- Value Type: REG_DWORD
- Value Data: `1`

## Deployment Target

- **Target Devices:** ~10,000 users on Windows servers (dedicated + RDS/Citrix)
- **Install Location:** `%ProgramData%\Microsoft\Windows\Start Menu\Programs\StartUp`
- **User Data Location:** `%APPDATA%\CustomerSurvey` (per-user flags)
- **Deployment Method:** SCCM Required deployment

## How It Works (User Experience)

1. **SCCM installs exe** → Copied to All Users Startup folder
2. **User logs in** → Windows automatically launches exe
3. **Exe checks flags** in `%APPDATA%\CustomerSurvey`:
   - If `done.flag` exists → Exit silently (survey completed)
   - If `nothanks.flag` exists → Exit silently (user opted out)
   - If `remind.txt` exists and date not passed → Exit silently
   - **Otherwise** → Show survey UI
4. **User interacts** → Creates appropriate flag
5. **Next login** → Exe checks flag → Exits silently

## Recommended Deployment Phases

| Phase | Week | Servers | Purpose |
|-------|------|---------|---------|
| **Pilot** | 1 | 5-10 | Initial testing with close monitoring |
| **Limited** | 2 | 100 | Verify stability and user experience |
| **Expanded** | 3 | 500 | Broader rollout with metrics tracking |
| **Full** | 4 | ~9,400 | Complete deployment to all users |

## Pre-Deployment Checklist

- [x] Built exe with production settings
- [x] Configured webhook URL in config.json
- [x] Created installation script (Install.ps1)
- [x] Created uninstallation script (Uninstall.ps1)
- [x] Created detection script (Detection.ps1)
- [x] Verified package integrity (all checks passed)
- [x] Documented deployment process (README.md, DEPLOYMENT.md)
- [x] Created package manifest (PACKAGE_MANIFEST.json)
- [ ] Code sign the exe (RECOMMENDED - submit to security team)
- [ ] Submit to SOC for whitelisting (CrowdStrike/AV exclusion)
- [ ] Copy package to SCCM content library location
- [ ] Create SCCM application
- [ ] Configure detection method
- [ ] Create pilot device collection
- [ ] Deploy to pilot servers
- [ ] Monitor and verify
- [ ] Proceed with phased rollout

## Support & Documentation

- **Full Documentation:** See `README.md` in this package
- **Quick Reference:** See `DEPLOYMENT.md` for SCCM admin commands
- **Package Manifest:** See `PACKAGE_MANIFEST.json` for complete specs
- **Verification:** Run `.\Verify-Package.ps1` to re-verify package

## Installation Verification Commands

After SCCM deploys, verify on a target machine:

```powershell
# Check if installed via detection method
Get-ItemProperty "HKLM:\SOFTWARE\CustomerSurvey"

# Check if exe exists in Startup folder
Get-Item "$env:ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp\customer-survey.exe"

# Test as a user
# 1. Log in to the server
# 2. Survey should appear (if no flags exist)
# 3. Complete survey or choose an option
# 4. Log out and log back in
# 5. Survey should NOT appear (flag created)

# Check user flags (as admin)
Get-ChildItem "C:\Users\*\AppData\Roaming\CustomerSurvey"
```

## Monitoring & Metrics

### SCCM Deployment Status
Monitor in SCCM Console:
```
Software Library → Applications → Customer Survey 
→ Right-click → View Status → Deployment Status
```

### Per-Server Statistics
```powershell
$appData = "C:\Users\*\AppData\Roaming\CustomerSurvey"
$completed = @(Get-ChildItem "$appData\done.flag" -EA SilentlyContinue).Count
$declined = @(Get-ChildItem "$appData\nothanks.flag" -EA SilentlyContinue).Count
$reminded = @(Get-ChildItem "$appData\remind.txt" -EA SilentlyContinue).Count

Write-Host "Completed: $completed | Declined: $declined | Remind Later: $reminded"
```

### Webhook Monitoring
- Check Zoho Sheets for incoming survey responses
- Monitor submission rate vs expected user count
- Verify data completeness and format

## Log Files

| Purpose | Location |
|---------|----------|
| Installation | `%TEMP%\CustomerSurvey_Install.log` |
| Uninstallation | `%TEMP%\CustomerSurvey_Uninstall.log` |
| SCCM Enforcement | `C:\Windows\CCM\Logs\AppEnforce.log` |
| Application Runtime | `%APPDATA%\.customer-survey\webhook.log` |

## Troubleshooting Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| Exe not launching on login | Not in Startup folder | Check SCCM deployment status, re-deploy if needed |
| Survey showing every login | Flag not being created | Check %APPDATA% permissions, verify flag write access |
| Survey never shows | Flag incorrectly exists | User can delete `%APPDATA%\CustomerSurvey` folder |
| Webhook failing | Network/firewall blocking | Verify HTTPS access to flow.zoho.in |
| Antivirus blocking | Not whitelisted | Submit to SOC for AV exception |

## Security Considerations

- **Code Signing:** Highly recommended to avoid AV false positives
- **SOC Coordination:** Submit exe to security team for review
- **CrowdStrike:** Request whitelisting (previous ticket: #796630)
- **Network:** Requires HTTPS outbound to flow.zoho.in
- **Permissions:** Runs as user (no admin rights required)
- **Data Privacy:** Survey responses sent to Zoho webhook only

## Contact Information

**Package Author:** Aryan Gupta  
**Organization:** Real Time Data Services Pvt Ltd  
**Build Date:** November 7, 2025  
**Package Location:** `customer-survey\SCCM-Package\`

## Next Immediate Steps

1. **Code sign the exe** (submit to security team for signing)
2. **Submit to SOC** for CrowdStrike/AV whitelisting approval
3. **Copy to SCCM** content library (e.g., `\\sccm-server\ContentLib\Customer-Survey\`)
4. **Create SCCM application** using instructions in DEPLOYMENT.md
5. **Create pilot collection** with 5-10 test servers
6. **Deploy to pilot** and monitor for 3-5 days
7. **Proceed with phased rollout** per deployment phases above

## Package Ready ✅

This package has been verified and is ready for SCCM deployment. All files are present, configuration is valid, and scripts are working correctly.

**Run `.\Verify-Package.ps1 -Detailed` at any time to re-verify the package.**

---

**Package prepared on:** November 7, 2025  
**Verification status:** PASSED  
**Ready for deployment:** YES
