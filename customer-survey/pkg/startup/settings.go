package startup

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// GetAppDataDir returns the per-user AppData folder for the survey app
func GetAppDataDir() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = os.Getenv("USERPROFILE")
	}
	return filepath.Join(appData, "CustomerSurvey")
}

// ensureAppDataDir creates the AppData directory if it doesn't exist
func ensureAppDataDir() error {
	dir := GetAppDataDir()
	return os.MkdirAll(dir, 0755)
}

// IsSurveyDone checks if done.flag exists
func IsSurveyDone() bool {
	flagPath := filepath.Join(GetAppDataDir(), "done.flag")
	_, err := os.Stat(flagPath)
	return err == nil
}

// IsNoThanks checks if nothanks.flag exists
func IsNoThanks() bool {
	flagPath := filepath.Join(GetAppDataDir(), "nothanks.flag")
	_, err := os.Stat(flagPath)
	return err == nil
}

// ShouldRemindLater checks if remind.txt exists and if current time is before the reminder date
// Returns true if we should skip showing the survey (still within reminder window)
func ShouldRemindLater() (bool, error) {
	remindPath := filepath.Join(GetAppDataDir(), "remind.txt")

	data, err := os.ReadFile(remindPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // file doesn't exist, show survey
		}
		return false, err
	}

	// Parse the stored date
	remindDate, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		// Invalid date format, ignore and show survey
		return false, nil
	}

	// If current time is before remind date, skip survey
	if time.Now().Before(remindDate) {
		return true, nil
	}

	// Reminder period has passed, show survey
	return false, nil
}

// ShouldShowSurvey checks all conditions and returns true if survey should be shown
func ShouldShowSurvey() (bool, error) {
	// Check done.flag
	if IsSurveyDone() {
		return false, nil
	}

	// Check nothanks.flag
	if IsNoThanks() {
		return false, nil
	}

	// Check remind.txt
	shouldSkip, err := ShouldRemindLater()
	if err != nil {
		return false, err
	}
	if shouldSkip {
		return false, nil
	}

	// All checks passed, show survey
	return true, nil
}

// MarkSurveyDone creates done.flag to indicate survey completion
func MarkSurveyDone() error {
	if err := ensureAppDataDir(); err != nil {
		return err
	}

	flagPath := filepath.Join(GetAppDataDir(), "done.flag")
	timestamp := time.Now().Format(time.RFC3339)
	return os.WriteFile(flagPath, []byte(timestamp), 0644)
}

// MarkNoThanks creates nothanks.flag to indicate user opted out
func MarkNoThanks() error {
	if err := ensureAppDataDir(); err != nil {
		return err
	}

	flagPath := filepath.Join(GetAppDataDir(), "nothanks.flag")
	timestamp := time.Now().Format(time.RFC3339)
	return os.WriteFile(flagPath, []byte(timestamp), 0644)
}

// MarkRemindLater creates/updates remind.txt with current time + 7 days
func MarkRemindLater() error {
	if err := ensureAppDataDir(); err != nil {
		return err
	}

	remindPath := filepath.Join(GetAppDataDir(), "remind.txt")
	remindDate := time.Now().Add(7 * 24 * time.Hour)
	remindDateStr := remindDate.Format(time.RFC3339)

	return os.WriteFile(remindPath, []byte(remindDateStr), 0644)
}

// ResetAll removes all flags/settings (useful for testing or reset functionality)
func ResetAll() error {
	dir := GetAppDataDir()

	files := []string{
		filepath.Join(dir, "done.flag"),
		filepath.Join(dir, "nothanks.flag"),
		filepath.Join(dir, "remind.txt"),
	}

	var lastErr error
	for _, f := range files {
		if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
			lastErr = err
		}
	}

	return lastErr
}

// GetStatus returns a human-readable status for debugging
func GetStatus() string {
	if IsSurveyDone() {
		return "Survey completed"
	}
	if IsNoThanks() {
		return "User opted out (No Thanks)"
	}

	shouldSkip, err := ShouldRemindLater()
	if err != nil {
		return fmt.Sprintf("Error checking reminder: %v", err)
	}
	if shouldSkip {
		remindPath := filepath.Join(GetAppDataDir(), "remind.txt")
		data, _ := os.ReadFile(remindPath)
		return fmt.Sprintf("Remind me later (until %s)", string(data))
	}

	return "Survey should be shown"
}
