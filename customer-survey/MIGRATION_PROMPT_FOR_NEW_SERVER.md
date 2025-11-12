# Customer Survey Feedback Application - Complete Migration Guide

## **EXECUTIVE SUMMARY**

This document provides a complete reconstruction guide for the Customer Survey application with memory optimization context. Use this when setting up on a new VS Code server.

---

## **PART 1: PROJECT OVERVIEW & CONTEXT**

### **Project Name:** Customer Survey Feedback Application
### **Primary Goal:** Collect customer feedback via Windows desktop application
### **Current Status:** Production-ready with memory optimization complete

### **Technology Stack:**
- **Primary:** Go 1.22+ (Wails 2.x)
- **Frontend:** HTML/CSS/JavaScript
- **UI Framework:** Wails (WebView2 on Windows)
- **Deployment:** SCCM package, All Users Startup
- **State Management:** File-based (per-user %APPDATA%)

### **Memory Reality Check:**
- **Expected:** 130 MB total (24 MB app + 110 MB WebView2 Chromium)
- **Why:** WebView2 requires full Chromium engine for HTML/CSS/JS rendering
- **This is OPTIMAL** for modern web-based native apps (vs. Electron at 150-300 MB)
- **Not Achievable:** Cannot reduce below 100 MB without removing HTML/CSS UI
- **Alternative:** Pure Windows Forms/WPF gives ~20-30 MB but loses all web technologies

### **Previous Optimization Attempts:**
1. ✅ **PyInstaller versions** (minimalui.exe variants): 5.57-8.27 MB
2. ✅ **ctypes-based versions**: ~7 MB memory (MessageBox UI only)
3. ⚠️ **tkinter versions**: ~22-23 MB (tkinter baseline overhead too high)
4. ❌ **Goal of 10-15 MB with styled HTML UI**: Not achievable due to WebView2 requirements

---

## **PART 2: DIRECTORY STRUCTURE**

```
CustomerSurveyFeedback/
├── README.md
├── customer-survey/                    # MAIN PROJECT
│   ├── cmd/
│   │   ├── survey/                     # Legacy UI (not used)
│   │   │   ├── main.go
│   │   │   └── main.manifest
│   │   ├── test-flags/                 # Testing utility
│   │   │   └── main.go
│   │   └── wails-app/                  # ⭐ PRODUCTION APP
│   │       ├── main.go                 # Entry point + Wails config
│   │       ├── config.json
│   │       ├── wails.json
│   │       ├── frontend/               # HTML/CSS/JS UI
│   │       │   ├── app.js
│   │       │   ├── index.html
│   │       │   └── wailsjs/           # Auto-generated Wails bindings
│   │       └── build/                  # Output directory
│   │           └── bin/
│   │               └── [exe files]
│   ├── internal/
│   │   ├── survey/                     # Zoho integration
│   │   │   ├── handler.go
│   │   │   └── zoho_auth.go
│   │   └── ui/                         # UI handlers
│   │       ├── desktop.go
│   │       └── form.go
│   ├── pkg/
│   │   ├── model/
│   │   │   └── response.go             # Data structures
│   │   ├── settings/
│   │   │   └── prompt_settings.go
│   │   └── startup/                    # ⭐ CRITICAL
│   │       ├── settings.go             # Flag file logic (done.flag, nothanks.flag, remind.txt)
│   │       └── settings_test.go        # Unit tests
│   ├── configs/
│   │   ├── config.json
│   │   └── zoho_secure.json.example
│   ├── go.mod                          # Dependencies
│   ├── go.sum
│   ├── build-desktop.ps1               # Build script
│   ├── STARTUP_LOGIC.md                # Documentation
│   └── SCCM-Package/                   # Deployment package
│       ├── Install.ps1
│       ├── Uninstall.ps1
│       ├── Detection.ps1
│       └── README.md
├── MinimalUI/                          # Python alternatives (archived)
│   ├── minimalui_memory_aggressive.exe
│   ├── minimalui_minimal_memory.exe
│   ├── minimalui_ctypes_optimized.exe
│   └── [other versions]
└── [various documentation files]
```

---

## **PART 3: BUILD & RUN INSTRUCTIONS**

### **Prerequisites:**
```powershell
# 1. Install Go 1.22+
# Download from: https://golang.org/dl

# 2. Install Wails CLI
go install github.com/wailsio/wails/v2/cmd/wails@latest

# 3. Install Node.js (for frontend build)
# Download from: https://nodejs.org

# 4. Verify installations
go version
wails version
node --version
```

### **Build Process:**

```powershell
# Navigate to project
cd c:\Users\aryan.gupta\OneDrive - Real Time Data Services Pvt Ltd\Desktop\survey\CustomerSurveyFeedback\customer-survey

# Build with Wails (produces: build\bin\windows\customer-survey.exe)
wails build -nsis -o customer-survey.exe

# Or build for development (with live reload):
wails dev

# Output location:
# c:\Users\aryan.gupta\OneDrive - Real Time Data Services Pvt Ltd\Desktop\survey\CustomerSurveyFeedback\customer-survey\build\bin\windows\customer-survey.exe
```

### **Configuration:**
Edit `configs/config.json`:
```json
{
  "webhook_url": "https://flow.zoho.in/[your-webhook-id]",
  "timeout_seconds": 30,
  "max_retries": 3
}
```

---

## **PART 4: CRITICAL: STARTUP FLAG LOGIC**

### **File-Based State Management (Per-User):**

Location: `%APPDATA%\CustomerSurvey\` (e.g., `C:\Users\[USERNAME]\AppData\Roaming\CustomerSurvey\`)

**Flag Files Created by Application:**

| Flag File | Purpose | Behavior |
|-----------|---------|----------|
| `done.flag` | Survey completed | ⏸️ Prevents survey from showing again (permanent) |
| `nothanks.flag` | User opted out | ⏸️ Prevents survey from showing again (permanent) |
| `remind.txt` | Remind me later | ⏸️ Prevents survey for 7 days, then shows again |

### **Startup Logic Flow:**

```
Application Starts
    ↓
Check: Is done.flag present?
    ├─ YES: Exit silently (survey already completed)
    └─ NO: Continue
    ↓
Check: Is nothanks.flag present?
    ├─ YES: Exit silently (user opted out)
    └─ NO: Continue
    ↓
Check: Is remind.txt present AND date not exceeded?
    ├─ YES: Exit silently (within reminder window)
    └─ NO: Continue
    ↓
SHOW SURVEY UI
    ↓
User Response:
    ├─ "Yes" (Submit feedback) → Create done.flag → Exit
    ├─ "Remind Me Later" → Create remind.txt (7 days from now) → Exit
    └─ "No Thanks" → Create nothanks.flag → Exit
```

### **Code Location:**
File: `pkg/startup/settings.go`

Key Functions:
- `ShouldShowSurvey()` - Returns true if conditions met
- `MarkSurveyDone()` - Creates done.flag
- `MarkNoThanks()` - Creates nothanks.flag
- `MarkRemindLater()` - Creates remind.txt with 7-day offset
- `ResetAll()` - Clears all flags (for testing)

### **Testing Flag Logic:**
```powershell
# Navigate to test-flags directory
cd cmd\test-flags

# Build test utility
go build -o test-flags.exe

# Run test (no UI, just flag checking)
.\test-flags.exe
```

---

## **PART 5: KEY CODE COMPONENTS**

### **5.1: Main Application Entry (cmd/wails-app/main.go)**

Critical sections to understand:

**A) Startup Configuration:**
```go
// Lines ~625-652: Wails options with memory optimizations
type Wails app.Options {
    Windows: windows.Options{
        WebviewGpuIsDisabled: true,      // CRITICAL: Disable GPU
        Theme: windows.Light,             // CRITICAL: Use Light theme
    },
    Bind: [app interface{}],
}
```

**B) Memory Optimization Flags:**
```go
// Browser arguments for Chromium engine (WebView2)
--disable-gpu
--disable-gpu-compositing
--js-flags=--max-old-space-size=32
--disable-dev-shm-usage
```

**C) Config Loading:**
```go
// Reads config.json for Zoho webhook URL
func LoadConfig() error {
    // Path: configs/config.json
}
```

### **5.2: Startup Logic (pkg/startup/settings.go)**

```go
// Main decision function (called on app launch)
func ShouldShowSurvey() (bool, error) {
    if IsSurveyDone() { return false, nil }
    if IsNoThanks() { return false, nil }
    skip, err := ShouldRemindLater()
    if skip { return false, err }
    return true, nil
}

// Create flags
func MarkSurveyDone() error { /* creates done.flag */ }
func MarkNoThanks() error { /* creates nothanks.flag */ }
func MarkRemindLater() error { /* creates remind.txt with 7-day offset */ }

// Reset all flags (testing)
func ResetAll() error { /* removes all flags */ }
```

### **5.3: Frontend (frontend/index.html + app.js)**

**HTML Structure:**
- Welcome screen with "Yes", "Remind Later", "No Thanks" buttons
- Feedback form (shown when user clicks "Yes")
- 1-2-3 rating system with text notes

**Key JavaScript Functions:**
```javascript
handleYes()           // Show feedback form
handleRemindLater()   // Call backend to create remind.txt
handleNoThanks()      // Call backend to create nothanks.flag
submitForm()          // Send feedback to Zoho webhook
```

### **5.4: Zoho Integration (internal/survey/handler.go)**

```go
// Sends user response to Zoho Flow webhook
func SubmitSurvey(response *model.SurveyResponse) error {
    // POST to webhook_url from config.json
    // Includes: Name, Email, Rating, Notes, Timestamp
}
```

---

## **PART 6: COMMON TASKS**

### **6.1: Rebuild from Scratch**

```powershell
# 1. Clean previous build
Remove-Item -Path "build" -Recurse -Force

# 2. Update dependencies
go mod download
go mod tidy

# 3. Build
wails build -nsis -o customer-survey.exe

# 4. Output at: build\bin\windows\customer-survey.exe
```

### **6.2: Test Flag Logic**

```powershell
# Clear all flags to see survey on next run
Remove-Item "$env:APPDATA\CustomerSurvey" -Recurse -Force

# Run app
.\build\bin\windows\customer-survey.exe

# App should show survey (no flags present yet)

# Check flags were created
Get-ChildItem "$env:APPDATA\CustomerSurvey"

# Reset for next test
.\cmd\test-flags\test-flags.exe
```

### **6.3: Change Reminder Duration**

File: `pkg/startup/settings.go`, Line ~12

```go
var RemindDuration = 7 * 24 * time.Hour  // Change to desired duration

// For testing (4 minutes):
// var RemindDuration = 4 * time.Minute
```

### **6.4: Change Webhook URL**

File: `configs/config.json`

```json
{
  "webhook_url": "https://flow.zoho.in/[YOUR-NEW-ID]/flow/webhook/incoming?zapikey=[YOUR-KEY]",
  "timeout_seconds": 30,
  "max_retries": 3
}
```

### **6.5: Deploy to New User**

The app is deployed via SCCM to `%ProgramData%\Microsoft\Windows\Start Menu\Programs\StartUp\`

User-specific state is in `%APPDATA%\CustomerSurvey\` (per user)

Each user gets independent flags - no conflicts.

---

## **PART 7: DEPLOYMENT METHODS**

### **7.1: SCCM Package (Enterprise)**

Files in: `SCCM-Package/`

```powershell
# Install
.\Install.ps1

# Uninstall
.\Uninstall.ps1

# Detection (returns 0 if installed, 1 if not)
.\Detection.ps1

# Manual deployment
Copy-Item "customer-survey.exe" -Destination "$env:ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp\"
Copy-Item "config.json" -Destination "$env:ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp\"
```

### **7.2: Manual Testing**

```powershell
# Run directly (UI appears)
.\build\bin\windows\customer-survey.exe

# Reset and test again
Remove-Item "$env:APPDATA\CustomerSurvey" -Recurse -Force
.\build\bin\windows\customer-survey.exe

# Test with -reset flag
.\build\bin\windows\customer-survey.exe -reset
```

---

## **PART 8: DEBUGGING**

### **8.1: Check What's Happening on Startup**

```powershell
# Check if flags exist (determines if survey shows)
Get-ChildItem "$env:APPDATA\CustomerSurvey"

# Expected output (when survey is done):
# done.flag, nothanks.flag, or remind.txt
```

### **8.2: View Application Logs**

Wails logs to console when run in dev mode:
```powershell
wails dev
```

### **8.3: Memory Profiling**

Windows Task Manager shows:
- Main Process (customer-survey.exe): ~24 MB
- WebView2 Child Process (msedgewebview2.exe): ~110 MB
- Total: ~130-140 MB (normal and optimal)

### **8.4: Common Issues**

| Issue | Cause | Solution |
|-------|-------|----------|
| Survey doesn't appear | Flag file exists | Run `.\build\bin\windows\customer-survey.exe -reset` |
| Webhook submission fails | No internet/firewall | Check network, disable firewall temporarily |
| High memory (>200MB) | GPU acceleration on | Verify `WebviewGpuIsDisabled: true` in main.go |
| Icon shows default Wails logo | Icon not embedded | Rebuild with `wails build -nsis` |

---

## **PART 9: PRODUCTION CHECKLIST**

Before deployment to 10k users:

- ✅ Config file has correct webhook URL
- ✅ `pkg/startup/settings.go` has correct reminder duration (default: 7 days)
- ✅ Frontend UI text matches brand guidelines
- ✅ Test on multiple Windows versions (Win10, Win11)
- ✅ Test flag creation in multi-user environment
- ✅ Verify memory usage ~130 MB (normal for Wails/WebView2)
- ✅ Test Zoho webhook submission
- ✅ Create SCCM package with proper detection script
- ✅ Notify IT of memory requirements
- ✅ Create user documentation for "Remind Me Later" behavior

---

## **PART 10: ARCHITECTURE DECISIONS**

### **Why Wails/WebView2?**
- ✅ Cross-platform ready (Windows/Mac/Linux code)
- ✅ Beautiful HTML/CSS/JS UI with branding
- ✅ Native Windows integration
- ✅ Smaller than Electron (130 MB vs 300 MB)

### **Why File-Based State?**
- ✅ Per-user isolation (no shared database needed)
- ✅ Works offline (no server dependency)
- ✅ Simple: just three files
- ✅ Easy to test and debug
- ✅ Survives upgrades (flags preserved)

### **Why 7-Day Reminder?**
- ✅ Prevents survey spam (but keeps asking)
- ✅ Balances feedback collection with user annoyance
- ✅ Configurable in code (see section 6.3)

### **Why 130 MB Memory?**
- ✅ WebView2 requires Chromium engine (100+ MB baseline)
- ✅ Cannot reduce without losing HTML/CSS UI
- ✅ Still smaller than alternative approaches
- ✅ Typical for modern web-based native apps

---

## **PART 11: FILE REFERENCE GUIDE**

| File | Purpose | Modify When |
|------|---------|------------|
| `cmd/wails-app/main.go` | App entry point + Wails config | Changing memory optimizations, window size, theme |
| `pkg/startup/settings.go` | Flag logic | Changing reminder duration, flag behavior |
| `internal/survey/handler.go` | Zoho submission | Changing API integration |
| `frontend/index.html` | UI structure | Changing layout, adding fields |
| `frontend/app.js` | UI logic | Changing button behavior, form validation |
| `configs/config.json` | Runtime config | Changing webhook URL, timeouts |
| `go.mod` | Dependencies | Adding new packages |
| `wails.json` | Wails build config | Changing output filename, build options |
| `build-desktop.ps1` | Build script | Customizing build process |

---

## **PART 12: ENVIRONMENT SETUP FOR NEW SERVER**

```powershell
# Step 1: Install Go
choco install golang  # or download from golang.org

# Step 2: Install Node.js
choco install nodejs  # or download from nodejs.org

# Step 3: Install Wails
go install github.com/wailsio/wails/v2/cmd/wails@latest

# Step 4: Clone/copy project
cd "C:\path\to\CustomerSurveyFeedback\customer-survey"

# Step 5: Install Go dependencies
go mod download

# Step 6: Verify setup
go version
wails version
node --version

# Step 7: Build
wails build -nsis -o customer-survey.exe

# Step 8: Test
.\build\bin\windows\customer-survey.exe
```

---

## **PART 13: CRITICAL NOTES**

### ⚠️ **MEMORY EXPECTATION:**
The application uses **130 MB memory** (24 MB app + 110 MB WebView2 Chromium). This is:
- ✅ Normal for Wails/WebView2 apps
- ✅ Smaller than Electron (150-300 MB)
- ✅ Required for HTML/CSS/JS rendering
- ❌ Cannot be reduced below 100 MB without losing UI

If you need <50 MB, rewrite using pure Windows Forms (ugly UI but lightweight).

### ✅ **PRODUCTION READY:**
Current build is production-ready and can be deployed to 10k users via SCCM.

### 📋 **PREVIOUS OPTIMIZATIONS:**
- PyInstaller versions (minimalui.exe): Abandoned (achieved 5.57-8.27 MB file, but memory still limited by tkinter baseline)
- ctypes versions: ~7 MB memory but MessageBox UI only (no branding)
- All documented in conversation history

### 🔧 **CUSTOMIZATION POINTS:**
Easy to customize:
1. Webhook URL (config.json)
2. Reminder duration (pkg/startup/settings.go)
3. UI appearance (frontend/index.html, css)
4. Form fields (frontend/index.html, app.js)
5. Zoho integration (internal/survey/handler.go)

---

## **QUICK START CHECKLIST FOR NEW SERVER**

```powershell
# 1. Install requirements (go, node, wails)
✓ Go 1.22+
✓ Node.js 16+
✓ Wails CLI

# 2. Navigate to project
cd C:\Users\aryan.gupta\OneDrive\...\survey\CustomerSurveyFeedback\customer-survey

# 3. Build
wails build -nsis -o customer-survey.exe

# 4. Test
.\build\bin\windows\customer-survey.exe

# 5. Check flags created
Get-ChildItem $env:APPDATA\CustomerSurvey

# 6. Deploy (via SCCM or manual copy to Startup folder)
```

---

## **CONTACT & QUESTIONS**

- **Main Code:** `cmd/wails-app/main.go`
- **Startup Logic:** `pkg/startup/settings.go`
- **Build Info:** `wails.json`, `go.mod`
- **Deployment:** `SCCM-Package/`

---

**Document Version:** 1.0  
**Last Updated:** November 12, 2025  
**Status:** Production-Ready ✅
