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
	"strings"
	"time"

	"customer-survey/pkg/model"
)

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

// SubmitSurvey sends the survey to an external endpoint (Zoho Forms) if configured.
// It returns an error if the forward fails.
func SubmitSurvey(ctx context.Context, resp model.SurveyResponse) error {
	// If a ZOHO_WEBHOOK_URL env var is set, POST to it.
	webhook := os.Getenv("ZOHO_WEBHOOK_URL")
	if webhook == "" {
		// If not configured, append to a local submissions.log for local testing.
		b, _ := json.Marshal(resp)
		f, err := os.OpenFile("submissions.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
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
	// append outgoing JSON to webhook.log for debugging
	_ = appendFile("webhook.log", fmt.Sprintf("%s | JSON payload: %s\n", time.Now().UTC().Format(time.RFC3339), string(payload)))
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
	_ = appendFile("webhook.log", fmt.Sprintf("%s | JSON response status: %s body: %s\n", time.Now().UTC().Format(time.RFC3339), res.Status, string(body)))

	// If response is empty, an error status, or contains an HTML error page (Apps Script can return an HTML error),
	// retry with form-encoded body which Apps Script reliably parses into e.parameters.
	lb := strings.ToLower(string(body))
	isHTML := strings.Contains(lb, "<!doctype html") || strings.Contains(lb, "<html") || strings.Contains(lb, "referenceerror")
	if res.StatusCode >= 400 || len(strings.TrimSpace(string(body))) == 0 || isHTML {
		log.Printf("[webhook] Empty or error response detected; retrying as form-encoded POST")

		vals := url.Values{}
		vals.Set("server_name", resp.ServerName)
		vals.Set("user_name", resp.UserName)
		vals.Set("server_performance", fmt.Sprintf("%d", resp.ServerPerformance))
		vals.Set("technical_support", fmt.Sprintf("%d", resp.TechnicalSupport))
		vals.Set("overall_support", fmt.Sprintf("%d", resp.OverallSupport))
		vals.Set("note", resp.Note)

		// append outgoing form payload to webhook.log
		_ = appendFile("webhook.log", fmt.Sprintf("%s | Form payload: %s\n", time.Now().UTC().Format(time.RFC3339), vals.Encode()))

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
		_ = appendFile("webhook.log", fmt.Sprintf("%s | Form response status: %s body: %s\n", time.Now().UTC().Format(time.RFC3339), res2.Status, string(body2)))
		if res2.StatusCode >= 400 {
			return &httpError{StatusCode: res2.StatusCode}
		}
	}

	return nil
}

type httpError struct{ StatusCode int }

func (h *httpError) Error() string { return http.StatusText(h.StatusCode) }
