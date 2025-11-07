# Production Deployment Test Plan
# Customer Survey Application - v2

## Pre-Deployment Checklist

### 1. Build Production Executable
- [ ] Clean build environment
- [ ] Run `wails build` with production flags
- [ ] Verify exe is not flagged by antivirus
- [ ] Code sign the executable (recommended)
- [ ] Test exe on clean Windows machine

### 2. Prepare Deployment Package
- [ ] Copy exe to deployment folder
- [ ] Include config.json with webhook URL
- [ ] Create installer script
- [ ] Test on pilot machine

### 3. SCCM/Deployment Preparation
- [ ] Create SCCM application package
- [ ] Set detection rules
- [ ] Configure deployment settings
- [ ] Target pilot collection first

---

## Phase 1: Local Machine Testing (Your Workstation)

### Test 1.1: Clean Install Simulation
```powershell
# Simulate fresh install
$installPath = "C:\Program Files\CustomerSurvey"
$exeSource = ".\build\bin\customer-survey.exe"
$configSource = ".\config.json"

# Create install directory (requires admin)
New-Item -Path $installPath -ItemType Directory -Force
Copy-Item $exeSource -Destination "$installPath\customer-survey.exe"
Copy-Item $configSource -Destination "$installPath\config.json"

# Verify files
Get-ChildItem $installPath
```

### Test 1.2: First User Logon Simulation
```powershell
# Clean user state
Remove-Item "$env:APPDATA\CustomerSurvey" -Recurse -Force -ErrorAction SilentlyContinue

# Run as if from Startup folder
& "C:\Program Files\CustomerSurvey\customer-survey.exe"

# Expected: Survey UI shows
# After interaction: Check flag creation
Get-ChildItem "$env:APPDATA\CustomerSurvey"
```

### Test 1.3: Second Logon (Should Not Show)
```powershell
# Simulate second login (flag exists)
& "C:\Program Files\CustomerSurvey\customer-survey.exe"

# Expected: Exe exits silently, no UI
```

### Test 1.4: Reset and Re-test
```powershell
& "C:\Program Files\CustomerSurvey\customer-survey.exe" -reset

# Expected: Survey shows again
```

---

## Phase 2: Multi-User Testing (RDS/Shared Server)

### Test 2.1: Simulate Different Users
```powershell
# User 1: Complete survey
Remove-Item "$env:APPDATA\CustomerSurvey" -Recurse -Force -ErrorAction SilentlyContinue
New-Item -Path "$env:APPDATA\CustomerSurvey" -ItemType Directory -Force
Set-Content "$env:APPDATA\CustomerSurvey\done.flag" -Value (Get-Date -Format "o")

Write-Host "User 1 state:"
Get-ChildItem "$env:APPDATA\CustomerSurvey"

# Note: On actual RDS, each user would have separate %APPDATA%
# Test with different user accounts if available
```

### Test 2.2: Concurrent Session Test
- [ ] Log in as User A - complete survey
- [ ] Log in as User B (different session) - should show survey
- [ ] Log in as User A again - should NOT show
- [ ] Verify %APPDATA% isolation

---

## Phase 3: Startup Folder Deployment Test

### Test 3.1: User Startup Folder (Per-User)
```powershell
# Copy to user startup (no admin needed)
$startupFolder = [Environment]::GetFolderPath('Startup')
$shortcutPath = Join-Path $startupFolder "CustomerSurvey.lnk"

$WshShell = New-Object -ComObject WScript.Shell
$Shortcut = $WshShell.CreateShortcut($shortcutPath)
$Shortcut.TargetPath = "C:\Program Files\CustomerSurvey\customer-survey.exe"
$Shortcut.WorkingDirectory = "C:\Program Files\CustomerSurvey"
$Shortcut.Save()

Write-Host "Shortcut created in: $startupFolder"

# Test: Log out and log back in
# Expected: Survey shows on login
```

### Test 3.2: All Users Startup (Requires Admin)
```powershell
# Copy to all users startup (admin required)
$allUsersStartup = "$env:ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp"
$shortcutPath = Join-Path $allUsersStartup "CustomerSurvey.lnk"

$WshShell = New-Object -ComObject WScript.Shell
$Shortcut = $WshShell.CreateShortcut($shortcutPath)
$Shortcut.TargetPath = "C:\Program Files\CustomerSurvey\customer-survey.exe"
$Shortcut.WorkingDirectory = "C:\Program Files\CustomerSurvey"
$Shortcut.Save()

Write-Host "All users shortcut created"
```

---

## Phase 4: Pilot Deployment (5-10 Servers)

### Test 4.1: Deploy to Test Servers
- [ ] Select 5-10 pilot servers (mix of dedicated and RDS)
- [ ] Deploy via SCCM or manual script
- [ ] Monitor first user logins
- [ ] Collect feedback

### Test 4.2: Monitor and Validate
```powershell
# On each pilot server, check deployment status
$installPath = "C:\Program Files\CustomerSurvey"
if (Test-Path "$installPath\customer-survey.exe") {
    Write-Host "✓ Deployed" -ForegroundColor Green
    Get-Item "$installPath\customer-survey.exe" | Select-Object Name, Length, LastWriteTime
} else {
    Write-Host "✗ Not deployed" -ForegroundColor Red
}

# Check if any users have interacted
$users = Get-ChildItem "C:\Users" -Directory | Select-Object -ExpandProperty Name
foreach ($user in $users) {
    $userAppData = "C:\Users\$user\AppData\Roaming\CustomerSurvey"
    if (Test-Path $userAppData) {
        Write-Host "User: $user"
        Get-ChildItem $userAppData | Format-Table Name, LastWriteTime
    }
}
```

### Test 4.3: Webhook Validation
- [ ] Verify webhook submissions in Zoho Sheets
- [ ] Check for any failed submissions
- [ ] Validate data format and completeness

---

## Phase 5: Production Rollout (All Servers)

### Test 5.1: Staged Rollout
- Week 1: Deploy to 100 servers
- Week 2: Deploy to 500 servers
- Week 3: Deploy to remaining servers

### Test 5.2: Monitoring Checklist
- [ ] SCCM deployment success rate
- [ ] Antivirus/SOC alerts
- [ ] User support tickets
- [ ] Webhook submission rate
- [ ] Error logs

---

## Testing Scenarios (Critical Paths)

### Scenario A: New User First Login
1. User logs in (never seen survey)
2. Survey shows
3. User completes survey
4. done.flag created
5. Next login: no survey

**Validation:** Check done.flag exists with timestamp

### Scenario B: Remind Me Later
1. User clicks "Remind Me Later"
2. remind.txt created with +7 days
3. User logs in within 7 days: no survey
4. User logs in after 7 days: survey shows

**Validation:** Check remind.txt date math

### Scenario C: No Thanks
1. User clicks "No Thanks"
2. nothanks.flag created
3. All future logins: no survey

**Validation:** Flag persists forever

### Scenario D: Multiple Users Same Machine
1. User A completes survey
2. User B logs in: sees survey
3. User A logs in again: no survey
4. User B completes survey
5. Both users: no survey

**Validation:** Separate %APPDATA% folders

### Scenario E: Reset for Testing
1. Admin runs: customer-survey.exe -reset
2. Flags removed
3. Next run: survey shows

**Validation:** All flags deleted

---

## Rollback Plan

### If Issues Occur
```powershell
# Emergency uninstall script
$installPath = "C:\Program Files\CustomerSurvey"
$startupShortcut = "$env:ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp\CustomerSurvey.lnk"

# Remove startup shortcut
Remove-Item $startupShortcut -Force -ErrorAction SilentlyContinue

# Remove installation
Remove-Item $installPath -Recurse -Force -ErrorAction SilentlyContinue

# Optionally: Remove user data
# Get-ChildItem "C:\Users\*\AppData\Roaming\CustomerSurvey" -Recurse | Remove-Item -Recurse -Force
```

---

## Success Criteria

### Technical Success
- [ ] 95%+ deployment success rate
- [ ] No critical antivirus alerts
- [ ] Webhook submissions working
- [ ] No performance impact on servers
- [ ] Multi-user sessions working correctly

### User Success
- [ ] Survey shows once per user
- [ ] "Remind Me Later" works correctly
- [ ] "No Thanks" honored
- [ ] Less than 5% support tickets
- [ ] User data isolated per account

### Data Success
- [ ] All responses in Zoho Sheets
- [ ] Backup files created locally
- [ ] No data loss
- [ ] Proper timestamps
- [ ] User identification correct

---

## Post-Deployment Monitoring (First Week)

### Daily Checks
```powershell
# Check webhook logs
Get-Content "C:\Users\*\AppData\Roaming\.customer-survey\webhook.log" -Tail 50

# Count flag files (how many users responded)
$doneCount = (Get-ChildItem "C:\Users\*\AppData\Roaming\CustomerSurvey\done.flag" -ErrorAction SilentlyContinue).Count
$noThanksCount = (Get-ChildItem "C:\Users\*\AppData\Roaming\CustomerSurvey\nothanks.flag" -ErrorAction SilentlyContinue).Count
$remindCount = (Get-ChildItem "C:\Users\*\AppData\Roaming\CustomerSurvey\remind.txt" -ErrorAction SilentlyContinue).Count

Write-Host "Survey Completed: $doneCount"
Write-Host "No Thanks: $noThanksCount"
Write-Host "Remind Later: $remindCount"
```

---

## Timeline

### Week 0: Pre-Production
- Build and test locally
- Deploy to your workstation
- Test all scenarios manually

### Week 1: Pilot (5-10 servers)
- Deploy to pilot group
- Monitor closely
- Gather feedback
- Fix any issues

### Week 2: Limited Rollout (100 servers)
- Deploy to 100 servers
- Monitor webhook submissions
- Check for errors

### Week 3: Expanded Rollout (500 servers)
- Deploy to 500 servers
- Continue monitoring

### Week 4: Full Rollout (Remaining servers)
- Deploy to all remaining servers
- Monitor for 1 week
- Mark as production stable

---

## Contact and Escalation

### Issues to Watch For
1. Antivirus blocking exe
2. Webhook timeouts/failures
3. File permission errors
4. Survey showing repeatedly (flag not created)
5. Survey not showing (flag incorrectly created)

### Escalation Path
- L1: User reports issue → Check local logs
- L2: Check %APPDATA% flags and webhook.log
- L3: Reset user flags with -reset
- L4: Redeploy exe if corrupted
