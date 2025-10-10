package main

import (
    "customer-survey/internal/ui"
    "log"
    "net/http"
)

func main() {
    // Initialize the application
    ui.ShowPopup()

    // Set up the HTTP server
    http.HandleFunc("/submit", ui.HandleSurveySubmission)
    log.Println("Starting server on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("Could not start server: %s\n", err)
    }
}