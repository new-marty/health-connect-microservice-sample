package analysis

type AnalysisRequest struct {
	Date string `json:"date"`
}

type AnalysisResponse struct {
	Date     string `json:"date"`
	Analysis string `json:"analysis"`
}
