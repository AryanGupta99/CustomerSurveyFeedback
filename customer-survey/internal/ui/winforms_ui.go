package ui

import (
	"customer-survey/pkg/model"
	"fmt"
)

// ShowProfessionalSurveyForm launches embedded web UI
func ShowProfessionalSurveyForm(handler func(model.SurveyResponse) error) error {
	// We'll use the embedded HTTP server with browser in app mode
	// This provides the beautiful v2 UI design
	return fmt.Errorf("should use browser mode")
}
