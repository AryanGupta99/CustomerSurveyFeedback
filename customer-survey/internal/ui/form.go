package ui

import (
    "fmt"
    "github.com/yourusername/yourproject/internal/survey"
)

type SurveyForm struct {
    Question1 int
    Question2 int
    Question3 int
}

func NewSurveyForm() *SurveyForm {
    return &SurveyForm{}
}

func (f *SurveyForm) CollectResponses() {
    fmt.Println("Please rate the following questions on a scale of 1 to 10:")

    fmt.Print("Question 1: How satisfied are you with our service? ")
    fmt.Scan(&f.Question1)

    fmt.Print("Question 2: How likely are you to recommend us to a friend? ")
    fmt.Scan(&f.Question2)

    fmt.Print("Question 3: How would you rate the quality of our products? ")
    fmt.Scan(&f.Question3)
}

func (f *SurveyForm) GetResponses() survey.Response {
    return survey.Response{
        Rating1: f.Question1,
        Rating2: f.Question2,
        Rating3: f.Question3,
    }
}