package survey

import (
    "bytes"
    "context"
    "encoding/json"
    "io"
    "net/http"
    "os"
    "time"

    "customer-survey/pkg/model"
    "log"
)

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
    payload, _ := json.Marshal(resp)
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
    log.Printf("Webhook response status: %s", res.Status)
    log.Printf("Webhook response body: %s", string(body))

    if res.StatusCode >= 400 {
        return &httpError{StatusCode: res.StatusCode}
    }
    return nil
}

type httpError struct{ StatusCode int }

func (h *httpError) Error() string { return http.StatusText(h.StatusCode) }