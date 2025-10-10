package model

type SurveyResponse struct {
	ServerName   string `json:"server_name"`
	UserName     string `json:"user_name"`
	RatingQ1     int    `json:"rating_q1"`
	RatingQ2     int    `json:"rating_q2"`
	RatingQ3     int    `json:"rating_q3"`
}