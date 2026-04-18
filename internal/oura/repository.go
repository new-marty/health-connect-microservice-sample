package oura

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListSleep(ctx context.Context, from, to, typ string) ([]SleepSession, error) {
	query := `SELECT oura_id, day, type, bedtime_start, bedtime_end,
		total_sleep_sec, deep_sleep_sec, rem_sleep_sec, light_sleep_sec,
		awake_sec, time_in_bed_sec, efficiency, latency_sec,
		avg_hr, lowest_hr, avg_hrv, avg_breath
		FROM sleep_sessions WHERE 1=1`
	args := []any{}

	if from != "" {
		query += " AND day >= ?"
		args = append(args, from)
	}
	if to != "" {
		query += " AND day <= ?"
		args = append(args, to)
	}
	if typ != "" {
		query += " AND type = ?"
		args = append(args, typ)
	}
	query += " ORDER BY day DESC, bedtime_start DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []SleepSession
	for rows.Next() {
		var s SleepSession
		if err := rows.Scan(
			&s.OuraID, &s.Day, &s.Type, &s.BedtimeStart, &s.BedtimeEnd,
			&s.TotalSleepSec, &s.DeepSleepSec, &s.REMSleepSec, &s.LightSleepSec,
			&s.AwakeSec, &s.TimeInBedSec, &s.Efficiency, &s.LatencySec,
			&s.AvgHR, &s.LowestHR, &s.AvgHRV, &s.AvgBreath,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func (r *Repository) GetSleepByDay(ctx context.Context, day string) ([]SleepSession, error) {
	return r.ListSleep(ctx, day, day, "")
}

func (r *Repository) UpsertSleep(ctx context.Context, s *SleepSession) error {
	_, err := r.db.ExecContext(ctx, `INSERT OR REPLACE INTO sleep_sessions
		(oura_id, day, type, bedtime_start, bedtime_end,
		 total_sleep_sec, deep_sleep_sec, rem_sleep_sec, light_sleep_sec,
		 awake_sec, time_in_bed_sec, efficiency, latency_sec,
		 avg_hr, lowest_hr, avg_hrv, avg_breath)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		s.OuraID, s.Day, s.Type, s.BedtimeStart, s.BedtimeEnd,
		s.TotalSleepSec, s.DeepSleepSec, s.REMSleepSec, s.LightSleepSec,
		s.AwakeSec, s.TimeInBedSec, s.Efficiency, s.LatencySec,
		s.AvgHR, s.LowestHR, s.AvgHRV, s.AvgBreath)
	return err
}

func (r *Repository) ListScores(ctx context.Context, from, to string) ([]DailyScore, error) {
	query := `SELECT day, sleep_score, sleep_deep, sleep_efficiency, sleep_latency,
		sleep_rem, sleep_restfulness, sleep_timing, sleep_total,
		readiness_score, readiness_temp_dev, readiness_temp_trend,
		readiness_activity_balance, readiness_body_temp, readiness_hrv_balance,
		readiness_prev_day_activity, readiness_prev_night, readiness_recovery_index,
		readiness_resting_hr, readiness_sleep_balance,
		activity_score, steps, active_cal, total_cal,
		high_activity_sec, medium_activity_sec, low_activity_sec, sedentary_sec,
		equivalent_walking_m, spo2_pct, breathing_disturbance_idx
		FROM daily_scores WHERE 1=1`
	args := []any{}

	if from != "" {
		query += " AND day >= ?"
		args = append(args, from)
	}
	if to != "" {
		query += " AND day <= ?"
		args = append(args, to)
	}
	query += " ORDER BY day DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []DailyScore
	for rows.Next() {
		var s DailyScore
		if err := rows.Scan(
			&s.Day, &s.SleepScore, &s.SleepDeep, &s.SleepEfficiency, &s.SleepLatency,
			&s.SleepREM, &s.SleepRestfulness, &s.SleepTiming, &s.SleepTotal,
			&s.ReadinessScore, &s.ReadinessTempDev, &s.ReadinessTempTrend,
			&s.ReadinessActivityBalance, &s.ReadinessBodyTemp, &s.ReadinessHRVBalance,
			&s.ReadinessPrevDayActivity, &s.ReadinessPrevNight, &s.ReadinessRecoveryIndex,
			&s.ReadinessRestingHR, &s.ReadinessSleepBalance,
			&s.ActivityScore, &s.Steps, &s.ActiveCal, &s.TotalCal,
			&s.HighActivitySec, &s.MediumActivitySec, &s.LowActivitySec, &s.SedentarySec,
			&s.EquivWalkingM, &s.SpO2Pct, &s.BreathingDisturbanceIdx,
		); err != nil {
			return nil, err
		}
		scores = append(scores, s)
	}
	return scores, rows.Err()
}

func (r *Repository) GetScoreByDay(ctx context.Context, day string) (*DailyScore, error) {
	scores, err := r.ListScores(ctx, day, day)
	if err != nil {
		return nil, err
	}
	if len(scores) == 0 {
		return nil, nil
	}
	return &scores[0], nil
}

func (r *Repository) UpsertScore(ctx context.Context, s *DailyScore) error {
	_, err := r.db.ExecContext(ctx, `INSERT OR REPLACE INTO daily_scores
		(day, sleep_score, sleep_deep, sleep_efficiency, sleep_latency,
		 sleep_rem, sleep_restfulness, sleep_timing, sleep_total,
		 readiness_score, readiness_temp_dev, readiness_temp_trend,
		 readiness_activity_balance, readiness_body_temp, readiness_hrv_balance,
		 readiness_prev_day_activity, readiness_prev_night, readiness_recovery_index,
		 readiness_resting_hr, readiness_sleep_balance,
		 activity_score, steps, active_cal, total_cal,
		 high_activity_sec, medium_activity_sec, low_activity_sec, sedentary_sec,
		 equivalent_walking_m, spo2_pct, breathing_disturbance_idx)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		s.Day, s.SleepScore, s.SleepDeep, s.SleepEfficiency, s.SleepLatency,
		s.SleepREM, s.SleepRestfulness, s.SleepTiming, s.SleepTotal,
		s.ReadinessScore, s.ReadinessTempDev, s.ReadinessTempTrend,
		s.ReadinessActivityBalance, s.ReadinessBodyTemp, s.ReadinessHRVBalance,
		s.ReadinessPrevDayActivity, s.ReadinessPrevNight, s.ReadinessRecoveryIndex,
		s.ReadinessRestingHR, s.ReadinessSleepBalance,
		s.ActivityScore, s.Steps, s.ActiveCal, s.TotalCal,
		s.HighActivitySec, s.MediumActivitySec, s.LowActivitySec, s.SedentarySec,
		s.EquivWalkingM, s.SpO2Pct, s.BreathingDisturbanceIdx)
	return err
}
