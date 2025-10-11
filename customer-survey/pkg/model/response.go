package model

type SurveyResponse struct {
	ServerName        string `json:"server_name"`
	UserName          string `json:"user_name"`
	ServerPerformance int    `json:"server_performance"`
	TechnicalSupport  int    `json:"technical_support"`
	OverallSupport    int    `json:"overall_support"`
	Note              string `json:"note,omitempty"`
}
