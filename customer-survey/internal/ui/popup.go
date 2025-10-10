package ui

import (
    "fmt"
    "time"
    "github.com/yourusername/yourproject/internal/survey"
)

func ShowSurveyPopup() {
    // Simulate a popup asking the user if they want to fill out the survey
    fmt.Println("Would you like to fill out a survey form? It will take approximately 10 seconds (yes/no):")
    
    var response string
    fmt.Scanln(&response)

    if response == "yes" {
        time.Sleep(10 * time.Second) // Simulate time taken to fill out the form
        survey.DisplaySurveyForm()
    } else {
        fmt.Println("Thank you for your time!")
    }
}