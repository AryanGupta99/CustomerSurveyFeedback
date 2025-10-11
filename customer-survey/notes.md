# Customer Survey App — Meeting Notes

## Executive summary
- Lightweight Go-based customer survey app with a clean, fast UI.
- Captures three ratings (Server Performance, Technical Support, Overall Support) plus an optional note.
- Automatically adds server hostname and logged-in username to each submission.
- Forwards submissions to a Google Apps Script Web App, which writes to a Google Sheet (Customer Survey Data).
- Minimal footprint: single Go binary, embedded static assets, standard library only.

## Architecture overview
- Client/UI: Single HTML/CSS/JS page served locally (Roboto font, gradient background, custom sliders, compact layout).
- Server (Go):
  - `GET /` — serves UI.
  - `POST /submit` — receives ratings + note, enriches with server_name and user_name.
  - Forwards payload to a configured webhook (Google Apps Script Web App).
- Data sink: Google Apps Script accepts JSON or form-encoded posts and appends a row to the Google Sheet.
- Config: Webhook URL via env var `ZOHO_WEBHOOK_URL`. If unset, app can log locally to `submissions.log` for testing.

## Data flow
1. User opens `http://localhost:8080` and submits.
2. Go server enriches payload with:
   - `server_name`: OS hostname
   - `user_name`: Windows USERNAME
3. Payload fields sent to webhook:
   - `server_name`, `user_name`, `server_performance`, `technical_support`, `overall_support`, `note`
4. Apps Script writes to sheet columns: Server Name, User Name, Server Performance, Technical Support, Overall Support, Note, Timestamp.

## Why it’s lightweight
- Go standard library only; no external web frameworks.
- UI assets embedded; no extra runtime dependencies.
- Low memory/CPU, suitable for broad desktop rollout.

## Security and privacy
- Apps Script Web App currently public to accept posts from clients.
- Data captured: hostname and username for follow-ups; no sensitive PII beyond that.
- Recommendation: add a shared secret token (query/header) to prevent unsolicited posts; document data retention and access to the sheet.

## Operations
- Run: `survey.exe` (Windows). Set webhook per session with `ZOHO_WEBHOOK_URL`.
- Logs:
  - `webhook.log`: outgoing payloads and responses (for debugging).
  - `server.log`, `server.err`: stdout/stderr redirection when started from scripts.
  - `submissions.log`: local fallback when webhook not configured.
- Monitoring: Apps Script “Executions” view for request logs and errors.
- Backup: Google Sheet version history; export CSV if needed.

## Current status
- UI finalized and compact (no scrolling; modern look & feel).
- End-to-end workflow verified: local UI → Go server → Apps Script → Google Sheet.
- Apps Script now robust to JSON and form-encoded posts.

## Known limitations
- Google Apps Script quotas apply (suitable for light/medium volume).
- Public webhook without token is open to unsolicited posts.
- Requires internet connectivity to write to Google Sheets.

## Recommended improvements
### Near term (low effort)
- Add simple auth token:
  - Server: include header `X-Webhook-Token: <secret>` (or `?token=` query).
  - Apps Script: validate token and reject invalid requests.
- Implement retry with exponential backoff for 429/5xx responses.
- Client-side validation: ensure sliders are within 1–10 before submit; real-time feedback.
- Basic analytics: average scores, daily counts; optional chart on a separate sheet or Looker Studio.

### Mid term
- Offline queue: cache posts locally and flush when online.
- Log rotation and levels (info/error) to keep logs small.
- Accessibility improvements (keyboard navigation, ARIA labels, contrast).
- Internationalization: externalize UI strings.
- Packaging: Windows service or tray app for auto-start.

### Long term
- Replace public webhook with a small authenticated API (self-hosted) if security/scale needs grow.
- Dashboard/reporting: automated charts and trends for leadership; NPS-style rollups.
- Admin view (SSO) to filter/export submissions.
- A/B testing of UI copy and layout to increase completion rates.

## Demo script (for the meeting)
1. Open `http://localhost:8080`.
2. Adjust the three sliders; add a brief note; click Submit.
3. Show new row in “Customer Survey Data” sheet with seven columns populated.
4. Open Apps Script → Executions to show a successful run (auditability).
5. Mention robustness: JSON and form-encoded support; minimal footprint.

## Q&A prep
- Why Google Sheets? Free, quick adoption by stakeholders, instant export and charts.
- Scale/security? Start with token auth; migrate to a private API if volume or security requirements increase.
- Can we change questions? Yes—update UI labels, keep JSON keys stable; or add fields and columns as needed.

## Appendix — Key technical notes
- Backend: Go 1.18+, standard `net/http` server, embedded UI assets via `//go:embed`.
- Endpoints: `/` (UI), `/submit` (POST handler), optional `/submissions` (local log viewer).
- Webhook sender: JSON first; if HTML/error/empty response detected, retries with `application/x-www-form-urlencoded`.
- Env var: `ZOHO_WEBHOOK_URL` (Apps Script Web App URL).
- Files of interest:
  - `cmd/survey/main.go` — entry point and route wiring
  - `internal/ui/static/index.html` & `script.js` — UI and client logic
  - `internal/ui/form.go` — request handler and enrichment
  - `pkg/model/response.go` — payload schema
  - `internal/survey/handler.go` — webhook sender with logging and fallback
