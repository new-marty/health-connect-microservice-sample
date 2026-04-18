package inbody

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

func (r *Repository) List(ctx context.Context, from, to string, limit int) ([]BodyCompScan, error) {
	query := `SELECT scan_datetime, date, weight_kg, smm_kg, bfm_kg, pbf_pct, bmi,
		bmr_kcal, ffm_kg, protein_kg, mineral_kg, icw_kg, ecw_kg, vfl
		FROM body_comp_scans WHERE 1=1`
	args := []any{}

	if from != "" {
		query += " AND date >= ?"
		args = append(args, from)
	}
	if to != "" {
		query += " AND date <= ?"
		args = append(args, to)
	}
	query += " ORDER BY scan_datetime DESC"
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scans []BodyCompScan
	for rows.Next() {
		var s BodyCompScan
		if err := rows.Scan(
			&s.ScanDatetime, &s.Date, &s.WeightKg, &s.SMMKg, &s.BFMKg, &s.PBFPct, &s.BMI,
			&s.BMRKcal, &s.FFMKg, &s.ProteinKg, &s.MineralKg, &s.ICWKg, &s.ECWKg, &s.VFL,
		); err != nil {
			return nil, err
		}
		scans = append(scans, s)
	}
	return scans, rows.Err()
}

func (r *Repository) Latest(ctx context.Context) (*BodyCompScan, error) {
	scans, err := r.List(ctx, "", "", 1)
	if err != nil {
		return nil, err
	}
	if len(scans) == 0 {
		return nil, nil
	}
	return &scans[0], nil
}

func (r *Repository) Upsert(ctx context.Context, s *BodyCompScan) error {
	_, err := r.db.ExecContext(ctx, `INSERT OR REPLACE INTO body_comp_scans
		(scan_datetime, date, weight_kg, smm_kg, bfm_kg, pbf_pct, bmi,
		 bmr_kcal, ffm_kg, protein_kg, mineral_kg, icw_kg, ecw_kg, vfl)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		s.ScanDatetime, s.Date, s.WeightKg, s.SMMKg, s.BFMKg, s.PBFPct, s.BMI,
		s.BMRKcal, s.FFMKg, s.ProteinKg, s.MineralKg, s.ICWKg, s.ECWKg, s.VFL)
	return err
}
