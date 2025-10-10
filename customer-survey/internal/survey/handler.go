package survey

import (
    "net/http"
    "encoding/json"
    "github.com/yourusername/customer-survey/pkg/model"
)

func SubmitSurveyHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        var response model.SurveyResponse
        err := json.NewDecoder(r.Body).Decode(&response)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // Here you would typically send the response to Zoho Forms or another endpoint

        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"message": "Survey submitted successfully!"})
        return
    }

    http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
}