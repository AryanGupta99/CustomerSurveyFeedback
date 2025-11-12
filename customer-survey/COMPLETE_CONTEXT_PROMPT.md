# COPY-PASTE PROMPT FOR NEW VS CODE SERVER

Use this as your main prompt context when setting up on a new server. Copy the entire content and paste when initializing your session.

---

## **FULL CONTEXT PROMPT**

I am working on the **Customer Survey Feedback Application** - a Windows desktop application built with **Go + Wails + WebView2** that collects customer feedback and submits it to Zoho Flow.

### **PROJECT LOCATION:**
```
c:\Users\aryan.gupta\OneDrive - Real Time Data Services Pvt Ltd\Desktop\survey\CustomerSurveyFeedback\customer-survey
```

### **TECHNOLOGY STACK:**
- **Language:** Go 1.22+
- **UI Framework:** Wails 2.x (WebView2 on Windows)
- **Frontend:** HTML/CSS/JavaScript
- **Deployment:** SCCM to All Users Startup folder
- **State Management:** File-based flags in %APPDATA%\CustomerSurvey\
- **Backend:** Zoho Flow webhook integration

### **CORE ARCHITECTURE:**

#### **1. FLAG-BASED STATE MANAGEMENT (CRITICAL)**
The application uses three flag files stored per-user in `%APPDATA%\CustomerSurvey\`:
- `done.flag` - Survey completed (prevents showing again)
- `nothanks.flag` - User opted out (prevents showing again)
- `remind.txt` - Remind me later (prevents showing for 7 days)

**Startup Logic:**
```
App starts → Check flags → If any flag found, exit silently
            → If no flags, SHOW SURVEY UI
User responds → Create appropriate flag → Exit
```

**Code Location:** `pkg/startup/settings.go`

#### **2. PRODUCTION APPLICATION (cmd/wails-app/)**
```
main.go               - Entry point + Wails config (CRITICAL: GPU disabled, memory optimized)
frontend/
  ├── index.html     - UI structure
  ├── app.js         - Button handlers, form submission
  └── wailsjs/       - Auto-generated Wails bindings (don't edit)
```

**Key Configurations in main.go:**
- `WebviewGpuIsDisabled: true` (memory optimization)
- `Theme: windows.Light` (lighter footprint)
- Browser args: `--disable-gpu`, `--js-flags=--max-old-space-size=32`, etc.

#### **3. ZOHO INTEGRATION (internal/survey/)**
- `handler.go` - Submits survey data to webhook URL
- Config file: `configs/config.json` (contains webhook URL)

#### **4. DEPENDENCIES (go.mod)**
Key packages:
- `github.com/wailsio/wails/v2` - Desktop app framework
- Standard library only (minimal dependencies)

### **MEMORY EXPECTATIONS:**
- **Total:** ~130 MB (24 MB app + 110 MB WebView2 Chromium)
- **Why:** WebView2 requires Chromium engine to render HTML/CSS/JavaScript
- **This is OPTIMAL** for Wails/web-based native apps
- **Cannot reduce below 100 MB** without removing HTML/CSS UI
- **Alternative:** Pure Windows Forms gives 20-30 MB but ugly UI

### **BUILD PROCESS:**
```powershell
cd c:\Users\aryan.gupta\OneDrive - Real Time Data Services Pvt Ltd\Desktop\survey\CustomerSurveyFeedback\customer-survey
wails build -nsis -o customer-survey.exe
# Output: build\bin\windows\customer-survey.exe (~4-5 MB file)
```

### **DEVELOPMENT MODE:**
```powershell
wails dev
# Live reload, browser preview at http://localhost:34115
```

### **KEY FILES & PURPOSES:**

| File | Purpose | Modify When |
|------|---------|------------|
| `cmd/wails-app/main.go` | App entry, Wails config | Window size, memory settings, theme |
| `pkg/startup/settings.go` | Flag logic | Reminder duration (7 days default), flag behavior |
| `internal/survey/handler.go` | Zoho submission | API integration changes |
| `frontend/index.html` | UI structure | Layout, fields, styling |
| `frontend/app.js` | UI behavior | Button handlers, form submission |
| `configs/config.json` | Configuration | **Webhook URL** (CRITICAL) |
| `go.mod` | Dependencies | Adding new packages |
| `wails.json` | Wails config | Build options |
| `SCCM-Package/Install.ps1` | Deployment script | Installation logic |

### **TESTING FLAG LOGIC (Without UI):**
```powershell
cd cmd/test-flags
go build -o test-flags.exe
.\test-flags.exe
```

### **RESET APPLICATION:**
```powershell
# Clear all flags to show survey again
Remove-Item "$env:APPDATA\CustomerSurvey" -Recurse -Force

# Or via command:
.\build\bin\windows\customer-survey.exe -reset
```

### **COMMON TASKS:**

**Change Webhook URL:**
1. Edit: `configs/config.json`
2. Update webhook_url field
3. Rebuild: `wails build -nsis -o customer-survey.exe`

**Change Reminder Duration:**
1. Edit: `pkg/startup/settings.go`
2. Change line ~12: `var RemindDuration = 7 * 24 * time.Hour`
3. Rebuild: `wails build -nsis -o customer-survey.exe`
4. Test: `Remove-Item "$env:APPDATA\CustomerSurvey" -Recurse -Force` then run app

**Update UI:**
1. Edit: `frontend/index.html` or `frontend/app.js`
2. Run: `wails dev` (live reload)
3. When satisfied, rebuild: `wails build -nsis -o customer-survey.exe`

**Add Form Fields:**
1. Add HTML input in `frontend/index.html`
2. Add JavaScript handler in `frontend/app.js`
3. Add Go struct field in `pkg/model/response.go`
4. Rebuild

**Deploy to SCCM:**
1. Update SCCM-Package files
2. Create SCCM package with latest .exe
3. Deploy to pilot group
4. Monitor success, then full rollout

### **DEPLOYMENT NOTES:**

- **Target:** ~10,000 Windows users
- **Deployment Method:** SCCM to All Users Startup folder
- **Per-User State:** Each user gets independent flags in their %APPDATA%
- **No Shared Database Needed:** File-based state works offline
- **Survives Upgrades:** Flags preserved when app updates
- **Remind Logic:** Users see survey every 7 days until they click "Done" or "No Thanks"

### **PREVIOUS CONTEXT (FOR REFERENCE):**

I initially attempted Python-based optimizations:
- PyInstaller versions (minimalui.exe): 5.57-8.27 MB file, but limited by Python runtime
- ctypes versions: ~7 MB memory (MessageBox only, no HTML UI)
- tkinter versions: ~22-23 MB (tkinter overhead too high)

**Conclusion:** Switched to Go + Wails because:
- ✅ Smaller file size (4-5 MB binary)
- ✅ Faster startup
- ✅ Modern web UI with HTML/CSS/JS
- ✅ WebView2 integration with branding
- ✅ Production-ready for enterprise SCCM deployment

The 130 MB memory is inherent to WebView2 (requires Chromium), not a code issue.

### **QUICK START ON NEW SERVER:**

```powershell
# 1. Install requirements
choco install golang nodejs  # or download from official sites
go install github.com/wailsio/wails/v2/cmd/wails@latest

# 2. Navigate to project
cd "C:\Users\aryan.gupta\OneDrive - Real Time Data Services Pvt Ltd\Desktop\survey\CustomerSurveyFeedback\customer-survey"

# 3. Install dependencies
go mod download

# 4. Build
wails build -nsis -o customer-survey.exe

# 5. Test
.\build\bin\windows\customer-survey.exe

# 6. Verify flags created
Get-ChildItem "$env:APPDATA\CustomerSurvey"

# 7. Reset for next test
Remove-Item "$env:APPDATA\CustomerSurvey" -Recurse -Force
```

### **CRITICAL CHECKLIST BEFORE PRODUCTION DEPLOYMENT:**

- [ ] Go 1.22+ installed and working
- [ ] Node.js 16+ installed and working
- [ ] Wails CLI installed: `go install github.com/wailsio/wails/v2/cmd/wails@latest`
- [ ] Project dependencies: `go mod download`
- [ ] Config.json has correct webhook URL
- [ ] Reminder duration set (default: 7 days)
- [ ] Build successful: `wails build -nsis -o customer-survey.exe`
- [ ] App launches and shows survey
- [ ] Flags created properly in %APPDATA%\CustomerSurvey\
- [ ] Zoho webhook receives submissions
- [ ] Reset command working: `customer-survey.exe -reset`
- [ ] Tested on multiple Windows versions
- [ ] Multi-user testing done (separate flags per user)
- [ ] SCCM package created
- [ ] Pilot deployment completed
- [ ] Memory footprint verified (~130 MB - this is NORMAL)
- [ ] Support team briefed

### **DOCUMENTATION AVAILABLE:**

1. **MIGRATION_PROMPT_FOR_NEW_SERVER.md** - Complete rebuild guide
2. **SCENARIO_BASED_REBUILDING.md** - Step-by-step scenarios for common tasks
3. **STARTUP_LOGIC.md** - Detailed flag behavior documentation
4. **BUILD_TEST_RESULTS.md** - Verification results
5. **SCCM-Package/README.md** - Deployment guide

### **WHEN TO USE THIS PROMPT:**

When I ask: **"I'm setting up on a new server, what should I do?"**

Provide this context, then I can help with:
- Setting up environment
- Building from scratch
- Modifying config
- Changing UI
- Deploying updates
- Troubleshooting issues
- Adding features

### **IMPORTANT GOTCHAS:**

1. ⚠️ **Memory is 130 MB, not 10 MB** - This is NORMAL for Wails/WebView2 apps. Don't try to reduce further.
2. ⚠️ **Webhook URL must be correct** in `configs/config.json` - Build won't change it if wrong
3. ⚠️ **Flags persist across updates** - Users won't see survey again unless reset or 7-day reminder expires
4. ⚠️ **Per-user isolation** - Each Windows user has separate flags (by design for SCCM deployment)
5. ⚠️ **Live reload only in dev mode** - Use `wails dev` for frontend changes, rebuild for production

---

**CONVERSATION SUMMARY:**

Our journey went from:
1. Python optimization attempts (5.57-23 MB memory)
2. Realization that WebView2/Wails is better suited for enterprise
3. Current production-ready Go + Wails application (4.5 MB file, 130 MB runtime with WebView2)
4. Full SCCM deployment ready for ~10k users

The application is **production-ready** and can be deployed immediately.

---

**End of Context Prompt**

Copy everything above and paste when starting work on your new VS Code server. This gives me full context about:
- Project structure
- Technology choices
- Memory expectations
- Common tasks
- Deployment strategy
- Previous optimization attempts

---

## **HOW TO USE THIS PROMPT:**

1. Copy all text above
2. Start new session on your new VS Code server
3. Paste into first message
4. Then ask: "I'm ready to set up the Customer Survey app. What's my first step?"
5. I'll have full context from this prompt and can help immediately

---

**Version:** 1.0  
**Last Updated:** November 12, 2025  
**Status:** Ready for New Server Migration
