package ui

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	userpkg "os/user"

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
		SurveyResponse    string `json:"survey_response"`
		ServerPerformance int    `json:"server_performance"`
		TechnicalSupport  int    `json:"technical_support"`
		OverallSupport    int    `json:"overall_support"`
		Note              string `json:"note"`
	}
	if err := json.Unmarshal(body, &incoming); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// enrich with local data
	hostname, _ := os.Hostname()
	// Robust username resolution (Windows and general fallback)
	user := os.Getenv("USERNAME")
	if user == "" {
		user = os.Getenv("USER")
	}
	if user == "" {
		if cu, err := userpkg.Current(); err == nil && cu != nil {
			user = cu.Username
		}
	}
	if user == "" {
		user = "unknown"
	}

	// Set survey_response to "completed" if not provided (for backward compatibility)
	surveyResponse := incoming.SurveyResponse
	if surveyResponse == "" {
		surveyResponse = "completed"
	}

	resp := model.SurveyResponse{
		ServerName:        hostname,
		UserName:          user,
		SurveyResponse:    surveyResponse,
		ServerPerformance: incoming.ServerPerformance,
		TechnicalSupport:  incoming.TechnicalSupport,
		OverallSupport:    incoming.OverallSupport,
		Note:              incoming.Note,
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
