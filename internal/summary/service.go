package summary

import (
	"context"
	"database/sql"
	"math"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) DailySummary(ctx context.Context, from, to, date string) ([]DailySummary, error) {
	if date != "" {
		from = date
		to = date
	}

	query := `
		WITH dates AS (
			SELECT DISTINCT date FROM (
				SELECT date FROM weight_readings
				UNION SELECT day AS date FROM sleep_sessions
				UNION SELECT day AS date FROM daily_scores
				UNION SELECT date FROM workouts
				UNION SELECT date FROM activities
				UNION SELECT date FROM meals
			) WHERE date IS NOT NULL AND date != ''`

	args := []any{}
	if from != "" {
		query += " AND date >= ?"
		args = append(args, from)
	}
	if to != "" {
		query += " AND date <= ?"
		args = append(args, to)
	}

	query += `
		)
		SELECT
			d.date,
			(SELECT AVG(w.weight_kg) FROM weight_readings w WHERE w.date = d.date AND w.source != 'inbody'),
			(SELECT AVG(w.body_fat_pct) FROM weight_readings w WHERE w.date = d.date AND w.body_fat_pct IS NOT NULL),
			(SELECT ROUND(s.total_sleep_sec / 3600.0, 1) FROM sleep_sessions s
			 WHERE s.day = d.date AND s.type = 'long_sleep' ORDER BY s.total_sleep_sec DESC LIMIT 1),
			(SELECT ROUND(s.deep_sleep_sec * 100.0 / NULLIF(s.total_sleep_sec, 0), 1)
			 FROM sleep_sessions s WHERE s.day = d.date AND s.type = 'long_sleep'
			 ORDER BY s.total_sleep_sec DESC LIMIT 1),
			(SELECT s.efficiency FROM sleep_sessions s WHERE s.day = d.date AND s.type = 'long_sleep'
			 ORDER BY s.total_sleep_sec DESC LIMIT 1),
			(SELECT s.avg_hrv FROM sleep_sessions s WHERE s.day = d.date AND s.type = 'long_sleep'
			 ORDER BY s.total_sleep_sec DESC LIMIT 1),
			(SELECT s.avg_hr FROM sleep_sessions s WHERE s.day = d.date AND s.type = 'long_sleep'
			 ORDER BY s.total_sleep_sec DESC LIMIT 1),
			ds.sleep_score,
			ds.readiness_score,
			ds.activity_score,
			ds.steps,
			ds.active_cal,
			ds.total_cal,
			ds.spo2_pct,
			(SELECT COUNT(*) FROM workouts wk WHERE wk.date = d.date),
			(SELECT ROUND(SUM(a.distance_m) / 1000.0, 2) FROM activities a WHERE a.date = d.date AND a.type = 'Run'),
			(SELECT SUM(m.calories) FROM meals m WHERE m.date = d.date),
			(SELECT ROUND(SUM(m.protein_g), 1) FROM meals m WHERE m.date = d.date)
		FROM dates d
		LEFT JOIN daily_scores ds ON ds.day = d.date
		ORDER BY d.date DESC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []DailySummary
	for rows.Next() {
		var ds DailySummary
		if err := rows.Scan(
			&ds.Date, &ds.WeightKg, &ds.BodyFatPct,
			&ds.SleepHrs, &ds.DeepSleepPct, &ds.SleepEfficiency, &ds.SleepHRV, &ds.SleepHR,
			&ds.SleepScore, &ds.ReadinessScore, &ds.ActivityScore,
			&ds.Steps, &ds.ActiveCal, &ds.TotalCal, &ds.SpO2Pct,
			&ds.WorkoutCount, &ds.RunDistanceKm,
			&ds.CaloriesIn, &ds.ProteinG,
		); err != nil {
			return nil, err
		}
		// Round weight
		if ds.WeightKg != nil {
			v := math.Round(*ds.WeightKg*10) / 10
			ds.WeightKg = &v
		}
		summaries = append(summaries, ds)
	}
	return summaries, rows.Err()
}

func (s *Service) WeeklyTrends(ctx context.Context, from, to string) ([]WeeklyTrend, error) {
	query := `
		WITH daily AS (
			SELECT
				d.date,
				(SELECT AVG(w.weight_kg) FROM weight_readings w WHERE w.date = d.date AND w.source != 'inbody') AS weight_kg,
				(SELECT ROUND(s.total_sleep_sec / 3600.0, 1) FROM sleep_sessions s
				 WHERE s.day = d.date AND s.type = 'long_sleep' ORDER BY s.total_sleep_sec DESC LIMIT 1) AS sleep_hrs,
				ds.steps,
				ds.readiness_score,
				(SELECT s.efficiency FROM sleep_sessions s WHERE s.day = d.date AND s.type = 'long_sleep'
				 ORDER BY s.total_sleep_sec DESC LIMIT 1) AS sleep_efficiency,
				(SELECT SUM(m.calories) FROM meals m WHERE m.date = d.date) AS calories_in
			FROM (SELECT DISTINCT day AS date FROM daily_scores) d
			LEFT JOIN daily_scores ds ON ds.day = d.date
			WHERE 1=1`

	args := []any{}
	if from != "" {
		query += " AND d.date >= ?"
		args = append(args, from)
	}
	if to != "" {
		query += " AND d.date <= ?"
		args = append(args, to)
	}

	query += `
		)
		SELECT
			date, weight_kg, sleep_hrs, steps, readiness_score, sleep_efficiency, calories_in,
			ROUND(AVG(weight_kg) OVER w7, 1),
			ROUND(AVG(sleep_hrs) OVER w7, 1),
			CAST(AVG(steps) OVER w7 AS INTEGER),
			CAST(ROUND(AVG(readiness_score) OVER w7) AS INTEGER),
			CAST(ROUND(AVG(sleep_efficiency) OVER w7) AS INTEGER),
			CAST(AVG(calories_in) OVER w7 AS INTEGER)
		FROM daily
		WINDOW w7 AS (ORDER BY date ROWS BETWEEN 6 PRECEDING AND CURRENT ROW)
		ORDER BY date DESC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trends []WeeklyTrend
	for rows.Next() {
		var t WeeklyTrend
		if err := rows.Scan(
			&t.Date, &t.WeightKg, &t.SleepHrs, &t.Steps, &t.ReadinessScore,
			&t.SleepEfficiency, &t.CaloriesIn,
			&t.Weight7d, &t.Sleep7d, &t.Steps7d, &t.Readiness7d, &t.Efficiency7d, &t.Calories7d,
		); err != nil {
			return nil, err
		}
		trends = append(trends, t)
	}
	return trends, rows.Err()
}

func (s *Service) Report(ctx context.Context, date string, lookback int) (*ReportData, error) {
	if lookback <= 0 {
		lookback = 30
	}

	// Get daily summaries for the date range
	summaries, err := s.DailySummary(ctx, "", "", date)
	if err != nil {
		return nil, err
	}

	today := map[string]interface{}{}
	if len(summaries) > 0 {
		ds := summaries[0]
		today["sleep_score"] = ds.SleepScore
		today["sleep_hours"] = ds.SleepHrs
		today["readiness"] = ds.ReadinessScore
		today["steps"] = ds.Steps
		today["cal_burned"] = ds.TotalCal
		today["active_cal"] = ds.ActiveCal
		today["cal_intake"] = ds.CaloriesIn
		today["protein"] = ds.ProteinG
		today["weight_kg"] = ds.WeightKg
		today["hrv"] = ds.SleepHRV
	}

	// InBody latest scans
	var inbody []map[string]interface{}
	rows, err := s.db.QueryContext(ctx, `SELECT date, weight_kg, smm_kg, bfm_kg, pbf_pct, bmi, bmr_kcal, ffm_kg
		FROM body_comp_scans ORDER BY scan_datetime DESC LIMIT 20`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var date string
			var weight, smm, bfm, pbf, bmi, ffm *float64
			var bmr *int
			if rows.Scan(&date, &weight, &smm, &bfm, &pbf, &bmi, &bmr, &ffm) == nil {
				inbody = append(inbody, map[string]interface{}{
					"date": date, "weight": weight, "smm": smm, "bfm": bfm,
					"pbf": pbf, "bmi": bmi, "bmr": bmr, "ffm": ffm,
				})
			}
		}
	}

	return &ReportData{
		TargetDate: date,
		Lookback:   lookback,
		Today:      today,
		Avg7d:      map[string]interface{}{},
		Series:     map[string]interface{}{},
		InBody:     inbody,
	}, nil
}
