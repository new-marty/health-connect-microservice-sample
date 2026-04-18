package applehealth

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

// Weight

func (r *Repository) ListWeight(ctx context.Context, from, to string) ([]WeightReading, error) {
	query := `SELECT timestamp, date, weight_kg, body_fat_pct, bmi, lean_mass_kg, source
		FROM weight_readings WHERE 1=1`
	args := []any{}

	if from != "" {
		query += " AND date >= ?"
		args = append(args, from)
	}
	if to != "" {
		query += " AND date <= ?"
		args = append(args, to)
	}
	query += " ORDER BY date DESC, timestamp DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []WeightReading
	for rows.Next() {
		var w WeightReading
		if err := rows.Scan(&w.Timestamp, &w.Date, &w.WeightKg, &w.BodyFatPct, &w.BMI, &w.LeanMassKg, &w.Source); err != nil {
			return nil, err
		}
		readings = append(readings, w)
	}
	return readings, rows.Err()
}

func (r *Repository) InsertWeight(ctx context.Context, w *WeightReading) error {
	_, err := r.db.ExecContext(ctx, `INSERT OR IGNORE INTO weight_readings
		(timestamp, date, weight_kg, body_fat_pct, bmi, lean_mass_kg, source)
		VALUES (?,?,?,?,?,?,?)`,
		w.Timestamp, w.Date, w.WeightKg, w.BodyFatPct, w.BMI, w.LeanMassKg, w.Source)
	return err
}

// Vitals

func (r *Repository) ListVitals(ctx context.Context, metric, from, to string) ([]Vital, error) {
	query := `SELECT id, timestamp, date, metric, value, unit, source
		FROM vitals WHERE 1=1`
	args := []any{}

	if metric != "" {
		query += " AND metric = ?"
		args = append(args, metric)
	}
	if from != "" {
		query += " AND date >= ?"
		args = append(args, from)
	}
	if to != "" {
		query += " AND date <= ?"
		args = append(args, to)
	}
	query += " ORDER BY date DESC, timestamp DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vitals []Vital
	for rows.Next() {
		var v Vital
		if err := rows.Scan(&v.ID, &v.Timestamp, &v.Date, &v.Metric, &v.Value, &v.Unit, &v.Source); err != nil {
			return nil, err
		}
		vitals = append(vitals, v)
	}
	return vitals, rows.Err()
}

func (r *Repository) InsertVital(ctx context.Context, v *Vital) error {
	_, err := r.db.ExecContext(ctx, `INSERT OR IGNORE INTO vitals
		(timestamp, date, metric, value, unit, source)
		VALUES (?,?,?,?,?,?)`,
		v.Timestamp, v.Date, v.Metric, v.Value, v.Unit, v.Source)
	return err
}
