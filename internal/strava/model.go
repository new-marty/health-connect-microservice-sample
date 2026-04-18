package strava

type Activity struct {
	StravaID       int64    `json:"strava_id"`
	StartTime      string   `json:"start_time"`
	Date           string   `json:"date"`
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	SportType      string   `json:"sport_type"`
	DistanceM      float64  `json:"distance_m"`
	MovingTimeSec  int      `json:"moving_time_sec"`
	ElapsedTimeSec int      `json:"elapsed_time_sec"`
	ElevationGainM float64  `json:"elevation_gain_m"`
	AvgSpeedMPS    *float64 `json:"avg_speed_mps"`
	MaxSpeedMPS    *float64 `json:"max_speed_mps"`
	HasHeartrate   bool     `json:"has_heartrate"`
}
