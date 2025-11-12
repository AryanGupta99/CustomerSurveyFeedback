# SCENARIO-SPECIFIC REBUILD INSTRUCTIONS

This document provides step-by-step instructions for common rebuild scenarios on your new VS Code server.

---

## **SCENARIO 1: Fresh Build from Source**

**When to use:** First time building on new server, or clean rebuild

```powershell
# ========== STEP 1: Environment Setup ==========
# Install Go 1.22+ from https://golang.org/dl
# Install Node.js 16+ from https://nodejs.org
# Install VS Code (or use existing)

# Verify installations
go version          # Expected: go version go1.22.x windows/amd64
wails version       # Expected: v2.8.x or higher
node --version      # Expected: v16+ or v18+

# ========== STEP 2: Navigate to Project ==========
cd "C:\Users\aryan.gupta\OneDrive - Real Time Data Services Pvt Ltd\Desktop\survey\CustomerSurveyFeedback\customer-survey"

# Verify you're in correct directory
ls -la cmd/wails-app/main.go       # Should exist
ls -la pkg/startup/settings.go     # Should exist
ls -la frontend/index.html         # Should exist

# ========== STEP 3: Install Dependencies ==========
go mod download
go mod tidy

# Verify dependencies
go mod graph        # Shows dependency tree

# ========== STEP 4: Build Application ==========
# Option A: Production build (recommended)
wails build -nsis -o customer-survey.exe

# Option B: Development build with live reload (testing)
wails dev

# ========== STEP 5: Verify Build Output ==========
# Check file exists
ls -la build\bin\windows\customer-survey.exe

# Check file size (should be ~4-5 MB)
(Get-Item build\bin\windows\customer-survey.exe).Length / 1MB

# ========== STEP 6: Test Application ==========
# Run the app
.\build\bin\windows\customer-survey.exe

# Expected behavior:
# - Survey UI appears
# - Click "Remind Me Later" or complete survey
# - Flags should be created in %APPDATA%\CustomerSurvey\

# Verify flags created
Get-ChildItem "$env:APPDATA\CustomerSurvey"
# Expected files: done.flag, nothanks.flag, or remind.txt

# ========== STEP 7: Clean up and prepare for distribution ==========
# Copy to output folder
Copy-Item "build\bin\windows\customer-survey.exe" -Destination ".\customer-survey.exe"
Copy-Item "configs\config.json" -Destination ".\config.json"

# Create distribution package
Write-Host "Build complete! Files ready at:" -ForegroundColor Green
ls -la customer-survey.exe
ls -la config.json
```

---

## **SCENARIO 2: Change Webhook URL**

**When to use:** Deploying to different Zoho Flow webhook

```powershell
# ========== STEP 1: Edit Config File ==========
# Edit this file: configs/config.json
# Using VS Code:
code configs/config.json

# Update the webhook_url:
{
  "webhook_url": "https://flow.zoho.in/[NEW-ID]/flow/webhook/incoming?zapikey=[NEW-KEY]",
  "timeout_seconds": 30,
  "max_retries": 3
}

# Save the file (Ctrl+S)

# ========== STEP 2: Rebuild ==========
wails build -nsis -o customer-survey.exe

# ========== STEP 3: Test New Webhook ==========
.\build\bin\windows\customer-survey.exe

# Click "Yes" and submit feedback
# Verify submission reaches Zoho Flow webhook

# ========== STEP 4: Verify Webhook Response ==========
# Check Zoho Flow execution logs at:
# https://flow.zoho.in/[workspace]/execution/[app-id]
```

---

## **SCENARIO 3: Change Reminder Duration**

**When to use:** Testing different "Remind Me Later" periods or changing production duration

```powershell
# ========== STEP 1: Edit Reminder Duration ==========
# Edit: pkg/startup/settings.go
# Using VS Code:
code pkg/startup/settings.go

# Find line ~12:
var RemindDuration = 7 * 24 * time.Hour

# Change to desired duration:
var RemindDuration = 4 * 24 * time.Hour  # 4 days
# OR
var RemindDuration = 4 * time.Minute     # 4 minutes (for testing)

# ========== STEP 2: Rebuild ==========
wails build -nsis -o customer-survey.exe

# ========== STEP 3: Test New Duration ==========
# Clear flags first
Remove-Item "$env:APPDATA\CustomerSurvey" -Recurse -Force

# Run app
.\build\bin\windows\customer-survey.exe

# Click "Remind Me Later"

# Check remind.txt
$remindText = Get-Content "$env:APPDATA\CustomerSurvey\remind.txt"
Write-Host "Reminder set until: $remindText"

# Verify duration is correct
# (Should be current time + your new duration)

# ========== STEP 4: Revert to Production Duration ==========
# After testing, change back to 7 days:
code pkg/startup/settings.go
# Change: var RemindDuration = 7 * 24 * time.Hour
wails build -nsis -o customer-survey.exe
```

---

## **SCENARIO 4: Customize UI (Change Text, Colors, Layout)**

**When to use:** Updating branding, changing questions, or modifying appearance

```powershell
# ========== STEP 1: Edit HTML/CSS ==========
# Edit: frontend/index.html
code frontend/index.html

# Update text, layout, colors, fonts
# Changes made will appear in live reload if using 'wails dev'

# ========== STEP 2: Edit JavaScript (Button Behavior) ==========
# Edit: frontend/app.js
code frontend/app.js

# Modify button handlers, form validation, etc.

# ========== STEP 3: Test Changes (Live Reload) ==========
# Use development mode for instant feedback:
wails dev

# This starts a dev server with live reload
# Open browser at: http://localhost:34115 (or port shown)
# Make changes to frontend files - they reload automatically

# ========== STEP 4: Rebuild for Production ==========
# When satisfied with changes:
wails build -nsis -o customer-survey.exe

# ========== STEP 5: Test Production Build ==========
.\build\bin\windows\customer-survey.exe

# Verify all changes appear correctly
```

---

## **SCENARIO 5: Add New Form Fields**

**When to use:** Collecting additional data (phone, company, etc.)

```powershell
# ========== STEP 1: Add HTML Input ==========
code frontend/index.html

# Add new input field in the form section, e.g.:
<label>Phone Number:</label>
<input type="tel" id="phone" placeholder="Your phone" required />

# ========== STEP 2: Update JavaScript Handler ==========
code frontend/app.js

# Find submitForm() function
# Add new field to data object:
let data = {
    name: document.getElementById('name').value,
    email: document.getElementById('email').value,
    rating: selectedRating,
    notes: document.getElementById('note').value,
    phone: document.getElementById('phone').value,  // NEW
};

# ========== STEP 3: Update Go Struct ==========
code pkg/model/response.go

# Add new field to SurveyResponse struct:
type SurveyResponse struct {
    Name   string `json:"name"`
    Email  string `json:"email"`
    Rating int    `json:"rating"`
    Notes  string `json:"notes"`
    Phone  string `json:"phone"`  // NEW
}

# ========== STEP 4: Test ==========
wails dev
# Make form changes
# Enter test data including new field
# Verify submission includes new field

# ========== STEP 5: Rebuild ==========
wails build -nsis -o customer-survey.exe
```

---

## **SCENARIO 6: Test Flag Logic Without UI**

**When to use:** Debugging startup behavior, testing flag creation logic

```powershell
# ========== STEP 1: Build Test Utility ==========
cd cmd/test-flags
go build -o test-flags.exe

# ========== STEP 2: Run Test (No UI) ==========
.\test-flags.exe

# Expected output:
# === Testing Startup Package Functions (No UI) ===
# [Test 1] Initial State Check
#   IsSurveyDone: false
#   IsNoThanks: false
#   ShouldRemindLater: false
#   ShouldShowSurvey: true
#   Status: Survey should be shown
# 
# [Test 2] Marking Survey as Done
# [Test 3] Marking No Thanks
# ... etc

# ========== STEP 3: Verify Flag Creation ==========
Get-ChildItem "$env:APPDATA\CustomerSurvey"

# Should show files created by test:
# done.flag
# nothanks.flag
# remind.txt

# ========== STEP 4: Check Flag Contents ==========
Get-Content "$env:APPDATA\CustomerSurvey\done.flag"
Get-Content "$env:APPDATA\CustomerSurvey\remind.txt"

# ========== STEP 5: Clean Up ==========
Remove-Item "$env:APPDATA\CustomerSurvey" -Recurse -Force
```

---

## **SCENARIO 7: Reset Application (Clear All Flags)**

**When to use:** After testing, need to show survey again

```powershell
# ========== METHOD 1: Command Line Flag ==========
.\build\bin\windows\customer-survey.exe -reset

# ========== METHOD 2: Manual Deletion ==========
Remove-Item "$env:APPDATA\CustomerSurvey" -Recurse -Force

# Verify deletion
Get-ChildItem "$env:APPDATA\CustomerSurvey" -ErrorAction SilentlyContinue
# Should show: Cannot find path (error) - that's correct

# ========== STEP 3: Next Run Will Show Survey ==========
.\build\bin\windows\customer-survey.exe

# Survey should appear (no flags to prevent it)
```

---

## **SCENARIO 8: Deploy to Production (SCCM)**

**When to use:** Rolling out to enterprise users via SCCM

```powershell
# ========== STEP 1: Prepare Deployment Package ==========
cd SCCM-Package

# Update Detection.ps1 to match your installation path
code Detection.ps1

# Update Install.ps1 with correct paths
code Install.ps1

# Verify config.json has correct webhook URL
code config.json

# ========== STEP 2: Create SCCM Package ==========
# In SCCM Console:
# 1. Create new Package
# 2. Add sources from SCCM-Package folder
# 3. Create program:
#    - Command line: powershell.exe -ExecutionPolicy Bypass -File Install.ps1
#    - Run with admin rights
#    - Allow users to interact: No (runs silently)
# 4. Detection method: Detection.ps1
# 5. Uninstall: Uninstall.ps1

# ========== STEP 3: Test on Pilot Group ==========
# Deploy to small test group first
# Verify:
# - Exe installed to Startup folder
# - Config.json present
# - Flags created per user
# - Webhook submissions working
# - No SCCM errors

# ========== STEP 4: Monitor Execution ==========
# Check SCCM logs on client:
Get-Item "C:\Program Files\...\customer-survey.exe"
Get-Item "$env:APPDATA\CustomerSurvey\done.flag"

# ========== STEP 5: Full Rollout ==========
# After pilot testing successful:
# Deploy to all users via SCCM
# Monitor logs and support tickets
```

---

## **SCENARIO 9: Update Application (New Version)**

**When to use:** Pushing app update to users

```powershell
# ========== STEP 1: Make Code Changes ==========
# Update any of:
# - frontend/index.html (UI)
# - frontend/app.js (logic)
# - pkg/startup/settings.go (behavior)
# - internal/survey/handler.go (Zoho integration)
# etc.

# ========== STEP 2: Test Changes ==========
wails dev

# Test in browser preview
# Verify functionality

# ========== STEP 3: Rebuild ==========
wails build -nsis -o customer-survey.exe

# ========== STEP 4: Test New Build ==========
# Reset flags first
Remove-Item "$env:APPDATA\CustomerSurvey" -Recurse -Force

# Run new build
.\build\bin\windows\customer-survey.exe

# Verify new behavior

# ========== STEP 5: Create New SCCM Package ==========
# Bump version in SCCM package
# Create new package with updated .exe
# Deploy to pilot group first
# Then full rollout

# ========== STEP 6: IMPORTANT: Preserve User Flags ==========
# When users update, their flags are preserved in %APPDATA%
# They won't see survey again unless:
# - They click "Remind Me Later" (7 day reset)
# - Admin runs reset command
# - They delete %APPDATA%\CustomerSurvey manually
# This is by design (don't spam users with repeated surveys)
```

---

## **SCENARIO 10: Debug Memory Usage**

**When to use:** Verifying memory footprint is within expectations

```powershell
# ========== STEP 1: Launch Application ==========
.\build\bin\windows\customer-survey.exe

# ========== STEP 2: Open Task Manager ==========
# Method 1: Ctrl+Shift+Esc
# Method 2: tasklist /v /fi "imagename eq customer-survey.exe"

# ========== STEP 3: Check Memory ==========
# Look for two processes:
#
# customer-survey.exe (main process):  ~24 MB
# msedgewebview2.exe (WebView2 child): ~110 MB
# ────────────────────────────────────────────
# TOTAL:                               ~130-140 MB
#
# This is NORMAL and OPTIMAL for Wails/WebView2

# ========== STEP 4: Verify via PowerShell ==========
# Real-time monitoring:
$process = Get-Process customer-survey -ErrorAction SilentlyContinue
$totalMemory = 0
foreach ($p in $process) {
    $totalMemory += $p.WorkingSet64 / 1MB
    Write-Host "$($p.Name): $($p.WorkingSet64 / 1MB)MB"
}
Write-Host "Total: $($totalMemory)MB"

# ========== STEP 5: Verify GPU Optimization ==========
# Check that GPU is disabled in main.go
# File: cmd/wails-app/main.go
# Line ~677: WebviewGpuIsDisabled: true,

# If memory is > 200MB, check:
# 1. GPU acceleration is disabled
# 2. Light theme is used
# 3. Browser flags are present
# 4. System isn't low on RAM (triggers compression)

# ========== STEP 6: Memory is EXPECTED to be ~130MB ==========
# This is GOOD - means Chromium is fully loaded
# DO NOT expect <50MB with HTML/CSS UI
# Alternative: Use pure Windows Forms for <30MB but ugly UI
```

---

## **SCENARIO 11: Troubleshoot Common Issues**

### **Issue: Survey doesn't appear**

```powershell
# Check 1: Are flags blocking it?
Get-ChildItem "$env:APPDATA\CustomerSurvey"

# Check 2: Is reminder period expired?
$remind = Get-Content "$env:APPDATA\CustomerSurvey\remind.txt" -ErrorAction SilentlyContinue
Get-Date -Date $remind

# Fix: Clear flags
Remove-Item "$env:APPDATA\CustomerSurvey" -Recurse -Force

# Or use reset command:
.\build\bin\windows\customer-survey.exe -reset

# Then run app again
.\build\bin\windows\customer-survey.exe
```

### **Issue: Webhook submission fails**

```powershell
# Check 1: Correct webhook URL in config?
Get-Content configs\config.json

# Check 2: Internet connectivity
ping flow.zoho.in

# Check 3: Firewall blocking?
# Temporarily disable firewall and retry

# Check 4: Webhook URL valid?
# Test with curl/PowerShell:
$webhook = "https://flow.zoho.in/[your-webhook-url]"
$body = @{ test = "value" } | ConvertTo-Json
Invoke-WebRequest -Uri $webhook -Method POST -Body $body

# Fix: Update webhook URL
code configs/config.json
# Update URL, save, rebuild:
wails build -nsis -o customer-survey.exe
```

### **Issue: Wrong reminder duration**

```powershell
# Check current setting
code pkg/startup/settings.go
# Line ~12: var RemindDuration = ...

# Update to correct duration
# Default production: 7 * 24 * time.Hour
# For testing: 4 * time.Minute

# Rebuild:
wails build -nsis -o customer-survey.exe
```

### **Issue: UI changes not appearing**

```powershell
# Check 1: Are you in dev mode?
wails dev
# Live reload works in dev mode

# Check 2: Is production build stale?
Remove-Item "build" -Recurse -Force

# Check 3: Rebuild production version
wails build -nsis -o customer-survey.exe

# Check 4: Clear old cached version
Remove-Item "$env:APPDATA\CustomerSurvey" -Recurse -Force

# Run new build:
.\build\bin\windows\customer-survey.exe
```

---

## **QUICK REFERENCE COMMANDS**

```powershell
# ===== BUILD COMMANDS =====
wails build -nsis -o customer-survey.exe    # Production build
wails dev                                    # Development with live reload
wails version                                # Check Wails version

# ===== TEST COMMANDS =====
.\build\bin\windows\customer-survey.exe      # Run built app
.\build\bin\windows\customer-survey.exe -reset # Reset all flags

# ===== FLAG MANAGEMENT =====
Get-ChildItem "$env:APPDATA\CustomerSurvey" # List all flags
Remove-Item "$env:APPDATA\CustomerSurvey" -Recurse -Force # Clear all flags
Get-Content "$env:APPDATA\CustomerSurvey\remind.txt" # Check reminder date

# ===== EDIT COMMANDS =====
code cmd/wails-app/main.go           # Main app entry
code pkg/startup/settings.go         # Flag logic
code frontend/index.html             # UI layout
code frontend/app.js                 # UI behavior
code configs/config.json             # Configuration
code pkg/model/response.go           # Data structure

# ===== DEPENDENCY COMMANDS =====
go mod download                      # Install dependencies
go mod tidy                         # Clean up unused dependencies
go mod graph                        # Show dependency tree

# ===== BUILD OUTPUT =====
ls -la build\bin\windows\customer-survey.exe # Check exe location
(Get-Item build\bin\windows\customer-survey.exe).Length / 1MB # Check size
```

---

## **CHECKLIST: Before Deploying to Production**

- [ ] Config.json has correct webhook URL
- [ ] Reminder duration set correctly (7 days default)
- [ ] UI text and branding correct
- [ ] All form fields working
- [ ] Flags created properly (done.flag, nothanks.flag, remind.txt)
- [ ] Memory usage ~130 MB (normal for Wails)
- [ ] Tested on multiple Windows versions
- [ ] Multi-user testing completed (different users get separate flags)
- [ ] Zoho webhook submissions working
- [ ] Reset command working (`-reset` flag)
- [ ] SCCM package created with Detection.ps1
- [ ] Pilot deployment to test group successful
- [ ] Support team trained on "Remind Me Later" behavior
- [ ] Backup of current build before full rollout

---

**Version:** 1.0  
**Last Updated:** November 12, 2025  
**Document Type:** Scenario-Based Rebuild Guide
