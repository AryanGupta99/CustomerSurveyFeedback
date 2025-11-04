package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PromptSettings stores user's choice for survey prompt (Remind Me Later or No Thanks)
type PromptSettings struct {
	NoThanks                 bool      `json:"no_thanks"`
	RemindMeLaterTimestamp   time.Time `json:"remind_me_later_timestamp"`
	LastShownTimestamp       time.Time `json:"last_shown_timestamp"`
	RemindMeLaterDays        int       `json:"remind_me_later_days"` // Number of days to wait (default: 7)
}

// getPromptSettingsPath returns the path to the prompt settings file in AppData
func getPromptSettingsPath() (string, error) {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = os.Getenv("APPDATA")
	}
	if localAppData == "" {
		return "", fmt.Errorf("cannot determine AppData path")
	}

	// Create .ace-survey directory if it doesn't exist
	settingsDir := filepath.Join(localAppData, ".ace-survey")
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create settings directory: %w", err)
	}

	return filepath.Join(settingsDir, "prompt_settings.json"), nil
}

// LoadPromptSettings reads the prompt settings from file
func LoadPromptSettings() (*PromptSettings, error) {
	path, err := getPromptSettingsPath()
	if err != nil {
		return &PromptSettings{RemindMeLaterDays: 7}, nil // Return default on error
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return default settings
			return &PromptSettings{RemindMeLaterDays: 7}, nil
		}
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings PromptSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings file: %w", err)
	}

	if settings.RemindMeLaterDays == 0 {
		settings.RemindMeLaterDays = 7
	}

	return &settings, nil
}

// SavePromptSettings writes the prompt settings to file
func SavePromptSettings(settings *PromptSettings) error {
	path, err := getPromptSettingsPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// ShouldShowPrompt determines if the survey prompt should be shown
func ShouldShowPrompt() (bool, error) {
	settings, err := LoadPromptSettings()
	if err != nil {
		// If we can't read settings, show the prompt
		return true, nil
	}

	// If user clicked "No Thanks", never show again
	if settings.NoThanks {
		return false, nil
	}

	// If user clicked "Remind Me Later", check if the time window has passed
	if !settings.RemindMeLaterTimestamp.IsZero() {
		remindAfter := settings.RemindMeLaterTimestamp.AddDate(0, 0, settings.RemindMeLaterDays)
		if time.Now().Before(remindAfter) {
			// Still within the "Remind Me Later" window
			return false, nil
		}
	}

	// Otherwise, show the prompt
	return true, nil
}

// SetNoThanks marks that user clicked "No Thanks"
func SetNoThanks() error {
	settings, _ := LoadPromptSettings()
	if settings == nil {
		settings = &PromptSettings{RemindMeLaterDays: 7}
	}

	settings.NoThanks = true
	settings.LastShownTimestamp = time.Now()

	return SavePromptSettings(settings)
}

// SetRemindMeLater marks that user clicked "Remind Me Later"
func SetRemindMeLater(days int) error {
	settings, _ := LoadPromptSettings()
	if settings == nil {
		settings = &PromptSettings{RemindMeLaterDays: 7}
	}

	if days == 0 {
		days = 7 // Default to 7 days
	}

	settings.RemindMeLaterTimestamp = time.Now()
	settings.RemindMeLaterDays = days
	settings.LastShownTimestamp = time.Now()
	settings.NoThanks = false // Reset No Thanks flag if user clicks Remind Me Later

	return SavePromptSettings(settings)
}

// ResetPromptSettings clears all settings (for testing or manual reset)
func ResetPromptSettings() error {
	settings := &PromptSettings{RemindMeLaterDays: 7}
	return SavePromptSettings(settings)
}
