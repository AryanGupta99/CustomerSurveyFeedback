package survey

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"customer-survey/pkg/model"
)

// getLogPath returns a hidden log path in AppData to keep desktop clean
func getLogPath(filename string) string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = os.Getenv("USERPROFILE")
	}
	logDir := filepath.Join(appData, ".customer-survey")
	os.MkdirAll(logDir, 0755) // Create hidden directory
	return filepath.Join(logDir, filename)
}

// appendFile appends data to a file, creating it if necessary.
func appendFile(path, data string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(data)
	return err
}

// DefaultWebhookURL can be set at build time via:
//
//	go build -ldflags "-X 'customer-survey/internal/survey.DefaultWebhookURL=https://.../exec'"
//
// If empty, env var or config.json will be used.
// Default to Zoho Flow webhook for direct sheet integration
var DefaultWebhookURL = "https://flow.zoho.in/60006321785/flow/webhook/incoming?zapikey=1001.754e60b74ab20d6a1f255f55358ee47d.815d8c8feab82ae7a18f99777d41a05f&isdebug=false"

type appConfig struct {
	WebhookURL string `json:"webhook_url"`
}

// getWebhookURL resolves the webhook URL from (in priority order):
// 1) Env var ZOHO_WEBHOOK_URL
// 2) config.json next to the executable (or current working dir)
// 3) Build-time default (DefaultWebhookURL)
func getWebhookURL() string {
	if v := os.Getenv("ZOHO_WEBHOOK_URL"); strings.TrimSpace(v) != "" {
		return v
	}
	// Try config.json next to the executable
	if exe, err := os.Executable(); err == nil {
		cfgPath := filepath.Join(filepath.Dir(exe), "config.json")
		if b, err := os.ReadFile(cfgPath); err == nil {
			var cfg appConfig
			if json.Unmarshal(b, &cfg) == nil && strings.TrimSpace(cfg.WebhookURL) != "" {
				return cfg.WebhookURL
			}
		}
	}
	// Try config.json in current working directory
	if b, err := os.ReadFile("config.json"); err == nil {
		var cfg appConfig
		if json.Unmarshal(b, &cfg) == nil && strings.TrimSpace(cfg.WebhookURL) != "" {
			return cfg.WebhookURL
		}
	}
	// Fallback to compile-time default
	if strings.TrimSpace(DefaultWebhookURL) != "" {
		return DefaultWebhookURL
	}
	// Nothing found — write a helpful diagnostic to webhook.log so packaged EXEs report clearly
	// This helps identify when the exe is run from a different folder or the config file wasn't packaged next to the exe.
	diag := []string{"webhook resolution failed; attempted sources:"}
	diag = append(diag, " - env ZOHO_WEBHOOK_URL (empty)")
	if exe, err := os.Executable(); err == nil {
		cfgPath := filepath.Join(filepath.Dir(exe), "config.json")
		diag = append(diag, fmt.Sprintf(" - config next to exe: %s (exists=%v)", cfgPath, fileExists(cfgPath)))
	}
	// cwd config
	cwdCfg := "config.json"
	diag = append(diag, fmt.Sprintf(" - cwd config: %s (exists=%v)", cwdCfg, fileExists(cwdCfg)))
	if strings.TrimSpace(DefaultWebhookURL) != "" {
		diag = append(diag, fmt.Sprintf(" - build-time DefaultWebhookURL present"))
	}
	_ = appendFile(getLogPath("webhook.log"), fmt.Sprintf("%s | %s\n", time.Now().UTC().Format(time.RFC3339), strings.Join(diag, "; ")))
	return ""
}

// fileExists returns true if the given path exists and is a file
func fileExists(path string) bool {
	if path == "" {
		return false
	}
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		return true
	}
	return false
}

// saveToLocalFile saves survey response to AppData when webhook fails
func saveToLocalFile(resp model.SurveyResponse) error {
	// Get the correct AppData\Local path
	var filePath string

	// Try LOCALAPPDATA first (points to AppData\Local)
	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		filePath = filepath.Join(localAppData, "Acesurvey.txt")
	} else {
		// Fallback to USERPROFILE\AppData\Local
		userProfile := os.Getenv("USERPROFILE")
		if userProfile == "" {
			return fmt.Errorf("cannot determine user profile path")
		}
		filePath = filepath.Join(userProfile, "AppData", "Local", "Acesurvey.txt")
	}

	// Prepare the data as JSON with timestamp
	entry := map[string]interface{}{
		"timestamp":          time.Now().Format(time.RFC3339),
		"server_name":        resp.ServerName,
		"username":           resp.UserName,
		"survey_response":    resp.SurveyResponse,
		"server_performance": resp.ServerPerformance,
		"technical_support":  resp.TechnicalSupport,
		"overall_rating":     resp.OverallSupport,
		"feedback":           resp.Note,
	}

	entryJSON, _ := json.MarshalIndent(entry, "", "  ")

	// OVERWRITE file with latest entry (do not append)
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open local backup file: %w", err)
	}
	defer f.Close()

	// Write JSON entry with separator
	_, err = f.WriteString(string(entryJSON) + "\n")
	if err != nil {
		return fmt.Errorf("failed to write to local backup file: %w", err)
	}

	return nil
}

// SubmitSurvey sends the survey response to Zoho Flow webhook which saves to Zoho Sheet
func SubmitSurvey(ctx context.Context, resp model.SurveyResponse) error {
	webhookLogPath := getLogPath("webhook.log")

	// Log the resolved webhook target
	wb := getWebhookURL()
	_ = appendFile(webhookLogPath, fmt.Sprintf("%s | resolved-webhook: %s\n", time.Now().UTC().Format(time.RFC3339), wb))

	if wb == "" {
		return fmt.Errorf("no webhook URL configured")
	}

	// Prepare the payload for Zoho Flow (keys aligned to Sheet/Flow mappings)
	// Ensure all values are strings to avoid type mismatch issues in Zoho Flow/Sheets.
	// Keys used by Flow action:
	// - server_name
	// - username
	// - server_performance
	// - technical_support
	// - overall_rating
	// - feedback
	// - survey_response (optional)
	// - timestamp (optional)
	// Map rating numbers to words
	ratingWord := func(val int) string {
		switch val {
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
		"server_name":        fmt.Sprintf("%v", resp.ServerName),
		"username":           fmt.Sprintf("%v", resp.UserName),
		"survey_response":    fmt.Sprintf("%v", resp.SurveyResponse),
		"server_performance": ratingWord(resp.ServerPerformance),
		"technical_support":  ratingWord(resp.TechnicalSupport),
		"overall_rating":     ratingWord(resp.OverallSupport),
		"feedback":           fmt.Sprintf("%v", resp.Note),
		"timestamp":          time.Now().Format(time.RFC3339),
	}

	payloadJSON, _ := json.Marshal(payload)
	_ = appendFile(webhookLogPath, fmt.Sprintf("%s | Zoho Flow payload: %s\n", time.Now().UTC().Format(time.RFC3339), string(payloadJSON)))

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, wb, bytes.NewBuffer(payloadJSON))
	if err != nil {
		_ = appendFile(webhookLogPath, fmt.Sprintf("%s | ERROR creating request: %v\n", time.Now().UTC().Format(time.RFC3339), err))
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		errMsg := fmt.Sprintf("%s | ERROR sending to Zoho Flow: %v\n", time.Now().UTC().Format(time.RFC3339), err)
		_ = appendFile(webhookLogPath, errMsg)
		log.Printf("[zoho-flow] Error: %v. Saving to local backup...", err)

		// Save to local backup on network/webhook failure
		if backupErr := saveToLocalFile(resp); backupErr != nil {
			_ = appendFile(webhookLogPath, fmt.Sprintf("%s | ERROR saving backup: %v\n", time.Now().UTC().Format(time.RFC3339), backupErr))
			log.Printf("[backup] Failed to save backup: %v", backupErr)
		} else {
			_ = appendFile(webhookLogPath, fmt.Sprintf("%s | SUCCESS: Backed up to local file\n", time.Now().UTC().Format(time.RFC3339)))
			log.Printf("[backup] Saved to local backup file")
		}

		return err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		_ = appendFile(webhookLogPath, fmt.Sprintf("%s | SUCCESS: Zoho Flow response status: %d\n", time.Now().UTC().Format(time.RFC3339), res.StatusCode))
		log.Printf("[zoho-flow] Successfully submitted to Zoho Sheet via Flow")

		// Also save to local backup for redundancy (even on success)
		if backupErr := saveToLocalFile(resp); backupErr != nil {
			log.Printf("[backup] Note: could not save backup copy: %v", backupErr)
		}

		return nil
	}

	// Log error response
	_ = appendFile(webhookLogPath, fmt.Sprintf("%s | ERROR: Zoho Flow response status: %d body: %s\n", time.Now().UTC().Format(time.RFC3339), res.StatusCode, string(body)))
	log.Printf("[zoho-flow] Error response: status=%d body=%s", res.StatusCode, string(body))

	if res.StatusCode >= 400 {
		// Save to local backup on HTTP error
		if backupErr := saveToLocalFile(resp); backupErr != nil {
			_ = appendFile(webhookLogPath, fmt.Sprintf("%s | ERROR saving backup: %v\n", time.Now().UTC().Format(time.RFC3339), backupErr))
			log.Printf("[backup] Failed to save backup: %v", backupErr)
		} else {
			_ = appendFile(webhookLogPath, fmt.Sprintf("%s | SUCCESS: Backed up to local file (HTTP %d)\n", time.Now().UTC().Format(time.RFC3339), res.StatusCode))
			log.Printf("[backup] Saved to local backup file (HTTP %d)", res.StatusCode)
		}

		return fmt.Errorf("zoho flow returned %d: %s", res.StatusCode, string(body))
	}

	return nil
}

// Removed legacy chromedp-based Zoho Survey automation to minimize binary size and dependencies.

// submitViaZohoOAuth submits survey data using secure OAuth authentication to Zoho Creator
func submitViaZohoOAuth(ctx context.Context, resp model.SurveyResponse, config *ZohoConfig) error {
	// Create Zoho Auth manager
	zohoAuth := NewZohoAuth(config)

	// Prepare data for Zoho Creator
	data := map[string]interface{}{
		"Server_Name":         resp.ServerName,
		"User_Name":           resp.UserName,
		"Survey_Response":     resp.SurveyResponse,
		"Server_Performance":  resp.ServerPerformance,
		"Technical_Support":   resp.TechnicalSupport,
		"Overall_Support":     resp.OverallSupport,
		"Additional_Comments": resp.Note,
	}

	// Log attempt
	webhookLogPath := getLogPath("webhook.log")
	_ = appendFile(webhookLogPath, fmt.Sprintf("%s | Using Zoho OAuth to %s\n",
		time.Now().UTC().Format(time.RFC3339), zohoAuth.GetAPIEndpoint()))

	// Submit to Zoho Creator
	err := zohoAuth.SubmitToZohoCreator(data)
	if err != nil {
		_ = appendFile(webhookLogPath, fmt.Sprintf("%s | Zoho OAuth ERROR: %v\n",
			time.Now().UTC().Format(time.RFC3339), err))
		log.Printf("[zoho-oauth] Error: %v", err)
		return err
	}

	_ = appendFile(webhookLogPath, fmt.Sprintf("%s | Zoho OAuth SUCCESS\n",
		time.Now().UTC().Format(time.RFC3339)))
	log.Printf("[zoho-oauth] Successfully submitted to Zoho Creator")

	return nil
}
