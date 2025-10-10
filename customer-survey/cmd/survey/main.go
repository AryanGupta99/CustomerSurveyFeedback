package main

import (
    "customer-survey/internal/ui"
    "log"
    "net/http"
)

func main() {
    // Serve UI static files and API
    http.HandleFunc("/", ui.HandleIndex)
    http.HandleFunc("/submit", ui.HandleSurveySubmission)
    http.HandleFunc("/submissions", ui.ListSubmissions)

    log.Println("Starting server on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("Could not start server: %s\n", err)
    }
}