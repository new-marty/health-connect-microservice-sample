package oura

type SleepSession struct {
	OuraID        string   `json:"oura_id"`
	Day           string   `json:"day"`
	Type          string   `json:"type"`
	BedtimeStart  string   `json:"bedtime_start"`
	BedtimeEnd    string   `json:"bedtime_end"`
	TotalSleepSec int      `json:"total_sleep_sec"`
	DeepSleepSec  int      `json:"deep_sleep_sec"`
	REMSleepSec   int      `json:"rem_sleep_sec"`
	LightSleepSec int      `json:"light_sleep_sec"`
	AwakeSec      int      `json:"awake_sec"`
	TimeInBedSec  int      `json:"time_in_bed_sec"`
	Efficiency    *int     `json:"efficiency"`
	LatencySec    *int     `json:"latency_sec"`
	AvgHR         *float64 `json:"avg_hr"`
	LowestHR      *int     `json:"lowest_hr"`
	AvgHRV        *int     `json:"avg_hrv"`
	AvgBreath     *float64 `json:"avg_breath"`
}

type DailyScore struct {
	Day string `json:"day"`

	// Sleep scores
	SleepScore        *int `json:"sleep_score"`
	SleepDeep         *int `json:"sleep_deep"`
	SleepEfficiency   *int `json:"sleep_efficiency"`
	SleepLatency      *int `json:"sleep_latency"`
	SleepREM          *int `json:"sleep_rem"`
	SleepRestfulness  *int `json:"sleep_restfulness"`
	SleepTiming       *int `json:"sleep_timing"`
	SleepTotal        *int `json:"sleep_total"`

	// Readiness
	ReadinessScore           *int     `json:"readiness_score"`
	ReadinessTempDev         *float64 `json:"readiness_temp_dev"`
	ReadinessTempTrend       *float64 `json:"readiness_temp_trend"`
	ReadinessActivityBalance *int     `json:"readiness_activity_balance"`
	ReadinessBodyTemp        *int     `json:"readiness_body_temp"`
	ReadinessHRVBalance      *int     `json:"readiness_hrv_balance"`
	ReadinessPrevDayActivity *int     `json:"readiness_prev_day_activity"`
	ReadinessPrevNight       *int     `json:"readiness_prev_night"`
	ReadinessRecoveryIndex   *int     `json:"readiness_recovery_index"`
	ReadinessRestingHR       *int     `json:"readiness_resting_hr"`
	ReadinessSleepBalance    *int     `json:"readiness_sleep_balance"`

	// Activity
	ActivityScore    *int `json:"activity_score"`
	Steps            int  `json:"steps"`
	ActiveCal        int  `json:"active_cal"`
	TotalCal         int  `json:"total_cal"`
	HighActivitySec  int  `json:"high_activity_sec"`
	MediumActivitySec int `json:"medium_activity_sec"`
	LowActivitySec   int  `json:"low_activity_sec"`
	SedentarySec     int  `json:"sedentary_sec"`
	EquivWalkingM    int  `json:"equivalent_walking_m"`

	// SpO2
	SpO2Pct               *float64 `json:"spo2_pct"`
	BreathingDisturbanceIdx *float64 `json:"breathing_disturbance_idx"`
}
