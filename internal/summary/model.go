package summary

type DailySummary struct {
	Date            string   `json:"date"`
	WeightKg        *float64 `json:"weight_kg"`
	BodyFatPct      *float64 `json:"body_fat_pct"`
	SleepHrs        *float64 `json:"sleep_hrs"`
	DeepSleepPct    *float64 `json:"deep_sleep_pct"`
	SleepEfficiency *int     `json:"sleep_efficiency"`
	SleepHRV        *int     `json:"sleep_hrv"`
	SleepHR         *float64 `json:"sleep_hr"`
	SleepScore      *int     `json:"sleep_score"`
	ReadinessScore  *int     `json:"readiness_score"`
	ActivityScore   *int     `json:"activity_score"`
	Steps           *int     `json:"steps"`
	ActiveCal       *int     `json:"active_cal"`
	TotalCal        *int     `json:"total_cal"`
	SpO2Pct         *float64 `json:"spo2_pct"`
	WorkoutCount    int      `json:"workout_count"`
	RunDistanceKm   *float64 `json:"run_distance_km"`
	CaloriesIn      *int     `json:"calories_in"`
	ProteinG        *float64 `json:"protein_g"`
}

type WeeklyTrend struct {
	Date            string   `json:"date"`
	WeightKg        *float64 `json:"weight_kg"`
	SleepHrs        *float64 `json:"sleep_hrs"`
	Steps           *int     `json:"steps"`
	ReadinessScore  *int     `json:"readiness_score"`
	SleepEfficiency *int     `json:"sleep_efficiency"`
	CaloriesIn      *int     `json:"calories_in"`
	Weight7d        *float64 `json:"weight_7d"`
	Sleep7d         *float64 `json:"sleep_7d"`
	Steps7d         *int     `json:"steps_7d"`
	Readiness7d     *int     `json:"readiness_7d"`
	Efficiency7d    *int     `json:"efficiency_7d"`
	Calories7d      *int     `json:"calories_7d"`
}

type ReportData struct {
	TargetDate string                    `json:"target_date"`
	Lookback   int                       `json:"lookback"`
	Today      map[string]interface{}    `json:"today"`
	Avg7d      map[string]interface{}    `json:"avg_7d"`
	Series     map[string]interface{}    `json:"series"`
	InBody     []map[string]interface{}  `json:"inbody"`
}
