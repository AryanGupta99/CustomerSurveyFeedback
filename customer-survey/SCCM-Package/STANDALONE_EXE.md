# ✅ STANDALONE EXE - No Config File Needed!

## Important: Config is Embedded in the Exe

The `customer-survey.exe` is **completely standalone** and **does NOT require config.json** to be deployed alongside it.

### How It Works

The config.json is **embedded into the exe** during the `wails build` process using Go's `//go:embed` directive:

```go
//go:embed config.json
var defaultConfigData []byte
```

### Priority Order (How Exe Finds Config)

1. **First**: Looks for `config.json` next to the exe (in Startup folder)
2. **Fallback**: Uses **embedded config** (built into exe) ✅
3. **Auto-creates**: Creates copy in `%LOCALAPPDATA%\.ace-survey\config.json`

### What This Means for Deployment

**ONLY deploy the EXE:**
- ✅ `customer-survey.exe` (10.1 MB) - **STANDALONE**
- ❌ `config.json` - NOT needed (already embedded)

The exe contains:
- Embedded config.json with webhook URL
- Frontend HTML/JS/CSS
- All application logic
- Startup flag checking

### SCCM Deployment - Simplified

**Option 1: Manual Copy (Simplest)**
```powershell
Copy-Item customer-survey.exe "$env:ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp\"
```

**Option 2: SCCM Package (Recommended)**
- Use Install.ps1 script (already configured)
- It copies ONLY the exe to Startup folder
- Creates registry detection key
- Logs installation

### Files Actually Deployed to Each Server

```
%ProgramData%\Microsoft\Windows\Start Menu\Programs\StartUp\
└── customer-survey.exe  (ONLY THIS FILE)
```

### User Data Files (Auto-Created Per User)

```
%APPDATA%\CustomerSurvey\
├── done.flag         (created when user completes survey)
├── nothanks.flag     (created when user clicks "No Thanks")
└── remind.txt        (created when user clicks "Remind Me Later")
```

### Why config.json is in SCCM-Package

The config.json in the SCCM-Package folder is used:
1. **During build** - Embedded into exe with `wails build`
2. **As reference** - Shows what's embedded in the exe
3. **Optional override** - Can be placed next to exe if webhook URL needs changing

### Updating Webhook URL After Deployment

If you need to change the webhook URL after deployment, you have 2 options:

**Option 1: Rebuild exe** (Recommended)
```powershell
# Edit config.json with new webhook URL
# Then rebuild:
wails build
```

**Option 2: Deploy config.json alongside exe** (Override)
```powershell
# Copy updated config.json to Startup folder
# Exe will use this instead of embedded config
Copy-Item config.json "$env:ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp\"
```

### Verification

To verify config is embedded in exe:

1. **Test without config.json:**
   ```powershell
   # Delete any config.json files
   Remove-Item config.json -ErrorAction SilentlyContinue
   
   # Run exe - should still work!
   .\customer-survey.exe
   ```

2. **Check logs:**
   ```
   Looking for config.json in:
     1. C:\Path\To\Exe\config.json
     ...
   ⚠ config.json not found on disk - using embedded default
   ✓ Using embedded config (built into exe)
   ✓ Webhook URL configured and validated
   ```

### Final Answer

**For SCCM deployment, you need:**
- ✅ **customer-survey.exe ONLY** (standalone, config embedded)
- ❌ config.json (NOT required for deployment)
- ✅ Install.ps1 (SCCM script to copy exe + create registry key)
- ✅ Uninstall.ps1 (SCCM script to remove exe)
- ✅ Detection.ps1 (SCCM detection method)

**The exe is 100% standalone!** 🎉

---

**Build Date:** November 7, 2025  
**Exe Size:** 10.1 MB  
**Config:** Embedded (no external file needed)  
**Dependencies:** None
