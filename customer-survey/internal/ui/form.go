package ui

import (
    "encoding/json"
    "io"
    "log"
    "net/http"
    "os"

    "customer-survey/internal/survey"
    "customer-survey/pkg/model"
)

// HandleSurveySubmission accepts JSON payload from the UI and forwards it to survey handler
func HandleSurveySubmission(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "could not read body", http.StatusBadRequest)
        return
    }

    var incoming struct {
        RatingQ1 int `json:"rating_q1"`
        RatingQ2 int `json:"rating_q2"`
        RatingQ3 int `json:"rating_q3"`
        Note     string `json:"note"`
    }
    if err := json.Unmarshal(body, &incoming); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }

    // enrich with local data
    hostname, _ := os.Hostname()
    user := os.Getenv("USERNAME")

    resp := model.SurveyResponse{
        ServerName: hostname,
        UserName:   user,
        RatingQ1:   incoming.RatingQ1,
        RatingQ2:   incoming.RatingQ2,
        RatingQ3:   incoming.RatingQ3,
        Note:       incoming.Note,
    }

    if err := survey.SubmitSurvey(r.Context(), resp); err != nil {
        log.Printf("error submitting survey: %v", err)
        http.Error(w, "could not submit survey", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"message":"submitted"}`))
}