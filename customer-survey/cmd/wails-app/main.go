package main

import (
	"bytes"
	"context"
	"customer-survey/pkg/startup"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend
var assets embed.FS

// Embedded default config - THIS MUST BE UPDATED WITH YOUR WEBHOOK URL
//
//go:embed config.json
var defaultConfigData []byte

// Survey represents the survey data structure
type Survey struct {
	SurveyResponse    string `json:"survey_response"`
	ServerPerformance int    `json:"server_performance"`
	TechnicalSupport  int    `json:"technical_support"`
	OverallSupport    int    `json:"overall_support"`
	Note              string `json:"note"`
	Timestamp         string `json:"timestamp"`
	Username          string `json:"username"`
	MachineName       string `json:"machine_name"`
}

// Config represents the Zoho configuration
type Config struct {
	ZohoWebhookURL string `json:"zoho_webhook_url"`
}

// App struct
type App struct {
	ctx    context.Context
	config *Config
}

// NewApp creates a new App application struct
func NewApp() *App {
	// Load config
	config := loadConfig()
	return &App{
		config: config,
	}
}

// loadConfig loads the configuration from config.json
func loadConfig() *Config {
	// Get executable directory FIRST - this is critical for deployed apps
	exePath, err := os.Executable()
	exeDir := ""
	if err == nil {
		exeDir = filepath.Dir(exePath)
	}

	// PRIORITY ORDER: Search executable directory FIRST, then fallback to other locations
	possiblePaths := []string{}

	// 1. Executable directory paths (HIGHEST PRIORITY - for deployed apps)
	if exeDir != "" {
		possiblePaths = append(possiblePaths,
			filepath.Join(exeDir, "config.json"),
			filepath.Join(exeDir, "configs", "config.json"),
		)
	}

	// 2. Current working directory (fallback for development)
	possiblePaths = append(possiblePaths,
		"config.json",
		filepath.Join("configs", "config.json"),
		filepath.Join("..", "..", "config.json"),
		filepath.Join("..", "..", "configs", "config.json"),
	)

	log.Printf("Looking for config.json in:")
	for i, path := range possiblePaths {
		absPath, _ := filepath.Abs(path)
		log.Printf("  %d. %s", i+1, absPath)
	}

	var configData []byte
	var foundPath string

	for _, path := range possiblePaths {
		configData, err = ioutil.ReadFile(path)
		if err == nil {
			foundPath = path
			absFoundPath, _ := filepath.Abs(path)
			log.Printf("✓ Config loaded successfully from: %s", absFoundPath)
			break
		}
	}

	// If not found on disk, use embedded config and create file in hidden location
	if foundPath == "" {
		log.Printf("⚠ config.json not found on disk - using embedded default")

		// Use embedded config data
		if len(defaultConfigData) > 0 {
			configData = defaultConfigData
			log.Printf("✓ Using embedded config (built into exe)")

			// Create config.json in hidden AppData location
			localAppData := os.Getenv("LOCALAPPDATA")
			if localAppData == "" {
				localAppData = os.Getenv("APPDATA")
			}

			if localAppData != "" {
				// Create .ace-survey directory if it doesn't exist
				configDir := filepath.Join(localAppData, ".ace-survey")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					log.Printf("⚠ Could not create config directory: %v", err)
				} else {
					configPath := filepath.Join(configDir, "config.json")
					if err := ioutil.WriteFile(configPath, defaultConfigData, 0644); err != nil {
						log.Printf("⚠ Could not save config to AppData: %v", err)
					} else {
						log.Printf("✓ Saved config to: %s", configPath)
						log.Printf("  (Hidden location in AppData - embedded config will be used)")
					}
				}
			}
		} else {
			log.Printf("╔════════════════════════════════════════════════════════╗")
			log.Printf("║ ERROR: No config.json found and no embedded config!   ║")
			log.Printf("╚════════════════════════════════════════════════════════╝")
			log.Printf("CRITICAL: Without config.json, webhook will NOT work!")
			log.Printf("")
			log.Printf("To fix this issue:")
			log.Printf("  1. Create config.json in the SAME folder as the .exe")
			if exeDir != "" {
				log.Printf("     Location: %s", filepath.Join(exeDir, "config.json"))
			}
			log.Printf("  2. Add this content:")
			log.Printf("     {")
			log.Printf("       \"zoho_webhook_url\": \"your-webhook-url-here\"")
			log.Printf("     }")
			log.Printf("")
			log.Printf("Data will be saved LOCALLY ONLY until config.json is added!")
			log.Printf("")
			return &Config{
				ZohoWebhookURL: "",
			}
		}
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		log.Printf("ERROR: Could not parse config.json: %v", err)
		log.Printf("File may be corrupted or have invalid JSON syntax")
		return &Config{
			ZohoWebhookURL: "",
		}
	}

	// Validate webhook URL
	if config.ZohoWebhookURL == "" {
		log.Printf("╔════════════════════════════════════════════════════════╗")
		log.Printf("║ WARNING: zoho_webhook_url is EMPTY in config.json!    ║")
		log.Printf("╚════════════════════════════════════════════════════════╝")
		log.Printf("Data will be saved LOCALLY ONLY - no data sent to Zoho!")
	} else if !isValidURL(config.ZohoWebhookURL) {
		log.Printf("╔════════════════════════════════════════════════════════╗")
		log.Printf("║ ERROR: Invalid webhook URL in config.json!            ║")
		log.Printf("╚════════════════════════════════════════════════════════╝")
		log.Printf("URL must start with http:// or https://")
		log.Printf("Current value: %s", config.ZohoWebhookURL)
		config.ZohoWebhookURL = "" // Clear invalid URL
	} else {
		log.Printf("✓ Webhook URL configured and validated")
		log.Printf("  URL: %s", config.ZohoWebhookURL)
	}

	return &config
}

// GetStartupStatus returns the current startup status for debugging
func (a *App) GetStartupStatus() string {
	return startup.GetStatus()
}

// ResetStartupSettings resets all startup flags (for testing/debugging)
func (a *App) ResetStartupSettings() map[string]interface{} {
	if err := startup.ResetAll(); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	return map[string]interface{}{"success": true, "message": "All startup settings reset"}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// Force window to foreground on startup. This tries multiple strategies with small delays
	// because Windows focus rules may block immediate foreground in some situations.
	go func() {
		// Short initial delay to allow window creation
		time.Sleep(150 * time.Millisecond)
		runtime.WindowShow(ctx)
		runtime.WindowSetAlwaysOnTop(ctx, true)
		// After a moment, remove always-on-top so normal window stacking resumes
		time.Sleep(600 * time.Millisecond)
		runtime.WindowSetAlwaysOnTop(ctx, false)
	}()
}

// HandleRemindMeLater saves reminder settings and closes the app
func (a *App) HandleRemindMeLater() map[string]interface{} {
	log.Printf("\n========== REMIND ME LATER ==========")

	if err := startup.MarkRemindLater(); err != nil {
		log.Printf("Error saving Remind Me Later: %v", err)
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	log.Printf("✓ Reminder set for 7 days")

	// Submit to Zoho Sheets
	username := os.Getenv("USERNAME")
	if username == "" {
		username = os.Getenv("USER")
	}
	machineName := os.Getenv("COMPUTERNAME")
	if machineName == "" {
		machineName, _ = os.Hostname()
	}

	reminderData := &Survey{
		SurveyResponse:    "Remind Me Later",
		ServerPerformance: 0,
		TechnicalSupport:  0,
		OverallSupport:    0,
		Note:              "User clicked 'Remind Me Later' - will be shown again in 7 days",
		Timestamp:         time.Now().Format(time.RFC3339),
		Username:          username,
		MachineName:       machineName,
	}

	log.Printf("Survey Response: %s", reminderData.SurveyResponse)
	log.Printf("Username: %s", username)
	log.Printf("Machine Name: %s", machineName)
	log.Printf("Note: %s", reminderData.Note)

	// Save backup locally
	if err := saveBackup(reminderData); err != nil {
		log.Printf("Error saving reminder backup: %v", err)
	}

	// Submit to Zoho webhook if configured
	webhookURL := a.config.ZohoWebhookURL
	if webhookURL != "" {
		log.Printf("Attempting to submit Remind Me Later to webhook...")
		if err := submitToZoho(webhookURL, reminderData); err != nil {
			log.Printf("ERROR: Failed to submit reminder to Zoho: %v", err)
			log.Printf("Data is saved locally. Check config.json webhook URL.")
		} else {
			log.Printf("✓ Reminder event submitted to Zoho Sheets")
		}
	} else {
		log.Printf("ERROR: No webhook URL configured in config.json")
		log.Printf("Data is saved locally only - webhook will not be called")
	}

	log.Printf("========================================\n")
	return map[string]interface{}{"success": true}
}

// HandleNoThanks saves no thanks settings and closes the app
func (a *App) HandleNoThanks() map[string]interface{} {
	log.Printf("\n========== NO THANKS ==========")

	if err := startup.MarkNoThanks(); err != nil {
		log.Printf("Error saving No Thanks: %v", err)
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	log.Printf("✓ Survey disabled (No Thanks)")

	// Submit to Zoho Sheets
	username := os.Getenv("USERNAME")
	if username == "" {
		username = os.Getenv("USER")
	}
	machineName := os.Getenv("COMPUTERNAME")
	if machineName == "" {
		machineName, _ = os.Hostname()
	}

	noThanksData := &Survey{
		SurveyResponse:    "No Thanks",
		ServerPerformance: 0,
		TechnicalSupport:  0,
		OverallSupport:    0,
		Note:              "User clicked 'No Thanks' - survey will not be shown again",
		Timestamp:         time.Now().Format(time.RFC3339),
		Username:          username,
		MachineName:       machineName,
	}

	log.Printf("Survey Response: %s", noThanksData.SurveyResponse)
	log.Printf("Username: %s", username)
	log.Printf("Machine Name: %s", machineName)
	log.Printf("Note: %s", noThanksData.Note)

	// Save backup locally
	if err := saveBackup(noThanksData); err != nil {
		log.Printf("Error saving no thanks backup: %v", err)
	}

	// Submit to Zoho webhook if configured
	webhookURL := a.config.ZohoWebhookURL
	if webhookURL != "" {
		log.Printf("Attempting to submit No Thanks to webhook...")
		if err := submitToZoho(webhookURL, noThanksData); err != nil {
			log.Printf("ERROR: Failed to submit no thanks to Zoho: %v", err)
			log.Printf("Data is saved locally. Check config.json webhook URL.")
		} else {
			log.Printf("✓ No Thanks event submitted to Zoho Sheets")
		}
	} else {
		log.Printf("ERROR: No webhook URL configured in config.json")
		log.Printf("Data is saved locally only - webhook will not be called")
	}

	log.Printf("========================================\n")
	return map[string]interface{}{"success": true}
}

// SubmitSurvey submits the survey data
func (a *App) SubmitSurvey(surveyResponse string, serverPerformance, technicalSupport, overallSupport int, note string) map[string]interface{} {
	// Get user and machine info
	username := os.Getenv("USERNAME")
	if username == "" {
		username = os.Getenv("USER")
	}
	machineName := os.Getenv("COMPUTERNAME")
	if machineName == "" {
		machineName, _ = os.Hostname()
	}

	surveyData := &Survey{
		SurveyResponse:    "Complete",
		ServerPerformance: serverPerformance,
		TechnicalSupport:  technicalSupport,
		OverallSupport:    overallSupport,
		Note:              note,
		Timestamp:         time.Now().Format(time.RFC3339),
		Username:          username,
		MachineName:       machineName,
	}

	// Log the submission with clear formatting
	log.Printf("\n========== SURVEY SUBMISSION ==========")
	log.Printf("Survey Response: %s", surveyResponse)
	log.Printf("Server Performance Rating: %d", serverPerformance)
	log.Printf("Technical Support Rating: %d", technicalSupport)
	log.Printf("Overall Support Rating: %d", overallSupport)
	log.Printf("Feedback/Note: %s", note)
	log.Printf("Timestamp: %s", surveyData.Timestamp)
	log.Printf("Username: %s", username)
	log.Printf("Machine Name: %s", machineName)
	log.Printf("========================================\n")

	// Save backup locally
	if err := saveBackup(surveyData); err != nil {
		log.Printf("Error saving backup: %v", err)
	}

	// Mark survey as done so it won't show again for this user
	if err := startup.MarkSurveyDone(); err != nil {
		log.Printf("Error marking survey as done: %v", err)
	} else {
		log.Printf("✓ Survey marked as completed for this user")
	}

	// Submit to Zoho webhook if configured
	webhookURL := a.config.ZohoWebhookURL
	if webhookURL == "" {
		log.Printf("╔════════════════════════════════════════════════════════╗")
		log.Printf("║ WEBHOOK DISABLED: No URL configured                   ║")
		log.Printf("╚════════════════════════════════════════════════════════╝")
		log.Printf("✓ Data saved to local backup: %s\\Acesurvey.txt", os.Getenv("LOCALAPPDATA"))
		log.Printf("")
		log.Printf("To enable Zoho integration:")
		log.Printf("  1. Create config.json in same directory as exe")
		log.Printf("  2. Add: {\"zoho_webhook_url\": \"your-webhook-url\"}")
	} else {
		log.Printf("🌐 Attempting webhook submission...")
		log.Printf("   Target: %s", webhookURL)
		if err := submitToZoho(webhookURL, surveyData); err != nil {
			log.Printf("╔════════════════════════════════════════════════════════╗")
			log.Printf("║ ❌ WEBHOOK SUBMISSION FAILED                          ║")
			log.Printf("╚════════════════════════════════════════════════════════╝")
			log.Printf("Error: %v", err)
			log.Printf("")
			log.Printf("✓ Data IS saved locally to: %s\\Acesurvey.txt", os.Getenv("LOCALAPPDATA"))
			log.Printf("")
			log.Printf("Troubleshooting:")
			log.Printf("  1. Verify config.json exists next to exe")
			log.Printf("  2. Check zoho_webhook_url is not empty")
			log.Printf("  3. Test webhook URL in browser or Postman")
			log.Printf("  4. Verify Zoho Flow is active")
			log.Printf("  5. Check internet connection and firewall")
		} else {
			log.Printf("╔════════════════════════════════════════════════════════╗")
			log.Printf("║ ✅ WEBHOOK SUBMISSION SUCCESSFUL                      ║")
			log.Printf("╚════════════════════════════════════════════════════════╝")
			log.Printf("Data sent to Zoho Sheets successfully!")
		}
	}

	return map[string]interface{}{
		"success": true,
	}
}

// saveBackup saves survey data to local file
func saveBackup(data *Survey) error {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = os.Getenv("TEMP")
	}

	backupFile := filepath.Join(localAppData, "Acesurvey.txt")

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal survey data: %w", err)
	}

	// Append to file
	f, err := os.OpenFile(backupFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(jsonData); err != nil {
		return fmt.Errorf("failed to write to backup file: %w", err)
	}
	if _, err := f.WriteString("\n---\n"); err != nil {
		return fmt.Errorf("failed to write separator: %w", err)
	}

	log.Printf("Backup saved to: %s", backupFile)
	return nil
}

// submitToZoho submits data to Zoho webhook
func isValidURL(urlStr string) bool {
	if urlStr == "" {
		return false
	}
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		return false
	}
	return true
}

func submitToZoho(webhookURL string, data *Survey) error {
	// Create payload with field names that match Zoho Flow expectations
	// These field names match exactly what the Zoho Flow is configured to receive
	// Convert ratings to strings for Zoho Flow
	ratingToString := func(r int) string {
		switch r {
		case 3:
			return "Good"
		case 2:
			return "Okay"
		case 1:
			return "Bad"
		default:
			return "Unknown"
		}
	}
	payload := map[string]interface{}{
		"timestamp":          data.Timestamp,
		"machine_name":       data.MachineName, // Zoho Flow expects: machine_name
		"username":           data.Username,
		"server_performance": ratingToString(data.ServerPerformance),
		"technical_support":  ratingToString(data.TechnicalSupport),
		"overall_support":    ratingToString(data.OverallSupport),
		"note":               data.Note, // Zoho Flow expects: note
		"survey_response":    data.SurveyResponse,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	log.Printf("Submitting to webhook: %s", webhookURL)
	log.Printf("Payload JSON: %s", string(jsonData))
	log.Printf("Payload size: %d bytes", len(jsonData))

	// Validate webhook URL
	if webhookURL == "" {
		return fmt.Errorf("webhook URL is empty")
	}
	if !isValidURL(webhookURL) {
		return fmt.Errorf("webhook URL is invalid: %s", webhookURL)
	}

	// Create request with proper headers for Zoho
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("ERROR: Failed to create request: %v", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set critical headers - Zoho is sensitive to these
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "CustomerSurvey/2.0")
	req.Header.Set("Accept", "application/json")

	log.Printf("Request headers: %v", req.Header)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("ERROR: HTTP request failed: %v", err)
		log.Printf("Network troubleshooting:")
		log.Printf("  - Check internet connection")
		log.Printf("  - Check firewall/proxy settings")
		log.Printf("  - Try webhook URL in browser")
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("Webhook response status: %d", resp.StatusCode)
	if len(body) > 0 {
		log.Printf("Webhook response body: %s", string(body))
	}

	// Success if we get 200-299 status code
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("✓ Successfully submitted to Zoho webhook (HTTP %d)", resp.StatusCode)
		return nil
	}

	// Any other status is an error
	log.Printf("ERROR: Webhook rejected the request with status %d", resp.StatusCode)
	log.Printf("Response: %s", string(body))
	return fmt.Errorf("webhook returned HTTP %d", resp.StatusCode)
}

func main() {
	// Parse command-line flags
	resetFlag := flag.Bool("reset", false, "Reset survey settings and show prompt")
	helpFlag := flag.Bool("help", false, "Show help message")
	flag.Parse()

	// Show help if requested
	if *helpFlag {
		fmt.Println("ACE Customer Survey")
		fmt.Println("Usage: survey.exe [options]")
		fmt.Println("Options:")
		fmt.Println("  -reset   Reset survey settings and show prompt again")
		fmt.Println("  -help    Show this help message")
		return
	}

	// Reset settings if requested
	if *resetFlag {
		err := startup.ResetAll()
		if err != nil {
			log.Printf("Error resetting settings: %v", err)
		} else {
			log.Printf("✓ Survey settings reset successfully")
		}
		// Continue to show survey after reset
	}

	// Check if survey prompt should be shown (unless we just reset)
	shouldShow := *resetFlag // Always show if reset was used
	if !*resetFlag {
		var err error
		shouldShow, err = startup.ShouldShowSurvey()
		if err != nil {
			log.Printf("Error checking startup settings: %v", err)
			shouldShow = true // Show by default if error
		}
	}

	// If user said "No Thanks" or within "Remind Me Later" window or already completed, exit silently
	if !shouldShow {
		status := startup.GetStatus()
		log.Printf("Survey prompt suppressed: %s", status)
		log.Printf("To reset and show the survey again, run: survey.exe -reset")
		return
	}

	log.Printf("✓ Showing survey prompt")
	log.Printf("Startup status: %s", startup.GetStatus())

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "ACE Customer Survey",
		Width:  420,
		Height: 720,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
