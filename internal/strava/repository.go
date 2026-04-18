package strava

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

func (r *Repository) List(ctx context.Context, from, to, typ string) ([]Activity, error) {
	query := `SELECT strava_id, start_time, date, name, type, sport_type,
		distance_m, moving_time_sec, elapsed_time_sec, elevation_gain_m,
		avg_speed_mps, max_speed_mps, has_heartrate
		FROM activities WHERE 1=1`
	args := []any{}

	if from != "" {
		query += " AND date >= ?"
		args = append(args, from)
	}
	if to != "" {
		query += " AND date <= ?"
		args = append(args, to)
	}
	if typ != "" {
		query += " AND type = ?"
		args = append(args, typ)
	}
	query += " ORDER BY date DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []Activity
	for rows.Next() {
		var a Activity
		var hasHR int
		if err := rows.Scan(
			&a.StravaID, &a.StartTime, &a.Date, &a.Name, &a.Type, &a.SportType,
			&a.DistanceM, &a.MovingTimeSec, &a.ElapsedTimeSec, &a.ElevationGainM,
			&a.AvgSpeedMPS, &a.MaxSpeedMPS, &hasHR,
		); err != nil {
			return nil, err
		}
		a.HasHeartrate = hasHR == 1
		activities = append(activities, a)
	}
	return activities, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*Activity, error) {
	var a Activity
	var hasHR int
	err := r.db.QueryRowContext(ctx, `SELECT strava_id, start_time, date, name, type, sport_type,
		distance_m, moving_time_sec, elapsed_time_sec, elevation_gain_m,
		avg_speed_mps, max_speed_mps, has_heartrate
		FROM activities WHERE strava_id = ?`, id).Scan(
		&a.StravaID, &a.StartTime, &a.Date, &a.Name, &a.Type, &a.SportType,
		&a.DistanceM, &a.MovingTimeSec, &a.ElapsedTimeSec, &a.ElevationGainM,
		&a.AvgSpeedMPS, &a.MaxSpeedMPS, &hasHR,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	a.HasHeartrate = hasHR == 1
	return &a, nil
}

func (r *Repository) Upsert(ctx context.Context, a *Activity) error {
	hasHR := 0
	if a.HasHeartrate {
		hasHR = 1
	}
	_, err := r.db.ExecContext(ctx, `INSERT OR REPLACE INTO activities
		(strava_id, start_time, date, name, type, sport_type,
		 distance_m, moving_time_sec, elapsed_time_sec, elevation_gain_m,
		 avg_speed_mps, max_speed_mps, has_heartrate)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		a.StravaID, a.StartTime, a.Date, a.Name, a.Type, a.SportType,
		a.DistanceM, a.MovingTimeSec, a.ElapsedTimeSec, a.ElevationGainM,
		a.AvgSpeedMPS, a.MaxSpeedMPS, hasHR)
	return err
}
