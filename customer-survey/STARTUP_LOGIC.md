# Customer Survey - Startup Logic

## Overview
This survey application uses a per-user file-based approach to manage survey display logic. The app checks specific files in `%AppData%\CustomerSurvey` to determine whether to show the survey to a user.

## How It Works

### File Locations
All per-user state files are stored in: `%APPDATA%\CustomerSurvey\`

Example path: `C:\Users\username\AppData\Roaming\CustomerSurvey\`

### State Files

1. **done.flag**
   - Created when user completes the survey
   - If exists: survey will NEVER show again for that user
   - Contains: timestamp of completion

2. **nothanks.flag**
   - Created when user clicks "No Thanks"
   - If exists: survey will NEVER show again for that user
   - Contains: timestamp of opt-out

3. **remind.txt**
   - Created/updated when user clicks "Remind Me Later"
   - Contains: RFC3339 formatted date (current time + 7 days)
   - If exists and current time < reminder date: survey will NOT show
   - If exists and current time >= reminder date: survey WILL show

## Startup Flow

```
User logs in → Windows launches exe from Startup folder
                     ↓
            Check done.flag exists?
                     ↓ Yes → Exit silently
                     ↓ No
            Check nothanks.flag exists?
                     ↓ Yes → Exit silently
                     ↓ No
            Check remind.txt exists?
                     ↓ No → Show survey
                     ↓ Yes
            Is current time < reminder date?
                     ↓ Yes → Exit silently
                     ↓ No → Show survey
```

## User Actions

### "Yes, I will give feedback"
- Shows full survey form
- On completion: creates `done.flag`
- Future logins: survey never shows again

### "Remind me later"
- Creates/updates `remind.txt` with date = now + 7 days
- Submits "Remind Me Later" event to webhook
- Future logins: survey hidden for 7 days, then shows again

### "No thanks"
- Creates `nothanks.flag`
- Submits "No Thanks" event to webhook
- Future logins: survey never shows again

## Multi-User Support

- Each Windows user has their own `%APPDATA%` folder
- State files are completely isolated per user
- Multiple users on the same machine can have different survey states
- Works correctly in RDS/Citrix multi-session environments

## Testing/Debugging

### Reset All Settings
```powershell
customer-survey.exe -reset
```
This removes all state files and shows the survey again.

### Check Current Status
The app logs the current status on startup:
```
Survey prompt suppressed: Survey completed
```

### Manual File Inspection
Check files in: `%APPDATA%\CustomerSurvey\`
- Open folder: Run → `%APPDATA%\CustomerSurvey`

## Backend Functions (Wails Bindings)

These functions are available from the frontend:

- `HandleRemindMeLater()` - Marks remind me later
- `HandleNoThanks()` - Marks no thanks
- `SubmitSurvey(...)` - Submits survey and marks done
- `GetStartupStatus()` - Returns current status string
- `ResetStartupSettings()` - Resets all flags (testing only)

## Deployment

### For All Users via SCCM/GPO
1. Deploy exe to: `C:\Program Files\CustomerSurvey\customer-survey.exe`
2. Create shortcut in All Users Startup folder:
   ```
   %ProgramData%\Microsoft\Windows\Start Menu\Programs\StartUp\
   ```
3. No admin rights needed for end users
4. Each user's first login will show the survey
5. Subsequent logins respect per-user state

### Command-Line Options
- `--help` - Show help
- `--reset` - Reset settings and show survey

## File Structure
```
customer-survey/
├── cmd/wails-app/main.go        # Main application entry
├── pkg/startup/
│   ├── settings.go              # Startup logic functions
│   └── settings_test.go         # Unit tests
└── go.mod                       # Module definition
```

## API Reference

### startup.ShouldShowSurvey() bool, error
Checks all conditions and returns true if survey should be shown.

### startup.MarkSurveyDone() error
Creates done.flag to mark survey as completed.

### startup.MarkNoThanks() error
Creates nothanks.flag to mark user opted out.

### startup.MarkRemindLater() error
Creates/updates remind.txt with date 7 days in future.

### startup.GetStatus() string
Returns human-readable status for debugging.

### startup.ResetAll() error
Removes all state files (for testing/reset).

## Notes

- Startup checks are performed before Wails UI initialization
- App exits silently (no window) if survey shouldn't be shown
- All file operations use Windows-safe paths
- Error handling allows survey to show if file checks fail
- Logs written to console/debug output for troubleshooting
