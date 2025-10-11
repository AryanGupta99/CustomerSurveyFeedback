package survey

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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
var DefaultWebhookURL string

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
	return ""
}

// SubmitSurvey sends the survey to an external endpoint (Zoho Forms) if configured.
// It returns an error if the forward fails.
func SubmitSurvey(ctx context.Context, resp model.SurveyResponse) error {
	// Resolve webhook URL from env/config/build-time default
	webhook := getWebhookURL()
	if webhook == "" {
		// If not configured, append to a hidden submissions.log for local testing.
		b, _ := json.Marshal(resp)
		submissionsLogPath := getLogPath("submissions.log")
		f, err := os.OpenFile(submissionsLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			// fallback to stdout
			_, _ = os.Stdout.Write(append(b, '\n'))
			return nil
		}
		defer f.Close()
		f.Write(append(b, '\n'))
		return nil
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}

	// First attempt: JSON
	payload, _ := json.Marshal(resp)
	// append outgoing JSON to hidden webhook.log for debugging
	webhookLogPath := getLogPath("webhook.log")
	_ = appendFile(webhookLogPath, fmt.Sprintf("%s | JSON payload: %s\n", time.Now().UTC().Format(time.RFC3339), string(payload)))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Read and log the response body for debugging
	body, _ := io.ReadAll(res.Body)
	log.Printf("[webhook] JSON request sent to %s", webhook)
	log.Printf("[webhook] response status: %s", res.Status)
	log.Printf("[webhook] response body: %s", string(body))
	_ = appendFile(webhookLogPath, fmt.Sprintf("%s | JSON response status: %s body: %s\n", time.Now().UTC().Format(time.RFC3339), res.Status, string(body)))

	// Retry only on true errors. Apps Script often returns 302 redirects on success; do not retry in that case.
	lb := strings.ToLower(string(body))
	hasScriptError := strings.Contains(lb, "referenceerror") || strings.Contains(lb, "exception") || strings.Contains(lb, "typeerror")
	if res.StatusCode >= 400 || hasScriptError {
		log.Printf("[webhook] Error detected (status=%d, markers=%v); retrying as form-encoded POST", res.StatusCode, hasScriptError)

		vals := url.Values{}
		vals.Set("server_name", resp.ServerName)
		vals.Set("user_name", resp.UserName)
		vals.Set("server_performance", fmt.Sprintf("%d", resp.ServerPerformance))
		vals.Set("technical_support", fmt.Sprintf("%d", resp.TechnicalSupport))
		vals.Set("overall_support", fmt.Sprintf("%d", resp.OverallSupport))
		vals.Set("note", resp.Note)

		// append outgoing form payload to hidden webhook.log
		_ = appendFile(webhookLogPath, fmt.Sprintf("%s | Form payload: %s\n", time.Now().UTC().Format(time.RFC3339), vals.Encode()))

		req2, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook, strings.NewReader(vals.Encode()))
		if err != nil {
			return err
		}
		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		res2, err := client.Do(req2)
		if err != nil {
			return err
		}
		defer res2.Body.Close()
		body2, _ := io.ReadAll(res2.Body)
		log.Printf("[webhook] Form request response status: %s", res2.Status)
		log.Printf("[webhook] Form request response body: %s", string(body2))
		_ = appendFile(webhookLogPath, fmt.Sprintf("%s | Form response status: %s body: %s\n", time.Now().UTC().Format(time.RFC3339), res2.Status, string(body2)))
		if res2.StatusCode >= 400 {
			return &httpError{StatusCode: res2.StatusCode}
		}
	} else {
		// Status is 2xx/3xx and no error markers -> treat as success, do not retry to avoid duplicates
		log.Printf("[webhook] Success without retry (status=%d)", res.StatusCode)
	}

	return nil
}

type httpError struct{ StatusCode int }

func (h *httpError) Error() string { return http.StatusText(h.StatusCode) }
