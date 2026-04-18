package meals

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

func (r *Repository) List(ctx context.Context, date, from, to string) ([]Meal, error) {
	query := `SELECT id, date, meal, description, calories, protein_g, fat_g, carbs_g, source
		FROM meals WHERE 1=1`
	args := []any{}

	if date != "" {
		query += " AND date = ?"
		args = append(args, date)
	}
	if from != "" {
		query += " AND date >= ?"
		args = append(args, from)
	}
	if to != "" {
		query += " AND date <= ?"
		args = append(args, to)
	}
	query += " ORDER BY date DESC, id"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var meals []Meal
	for rows.Next() {
		var m Meal
		if err := rows.Scan(&m.ID, &m.Date, &m.Meal, &m.Description, &m.Calories, &m.ProteinG, &m.FatG, &m.CarbsG, &m.Source); err != nil {
			return nil, err
		}
		meals = append(meals, m)
	}
	return meals, rows.Err()
}

func (r *Repository) Create(ctx context.Context, m *Meal) (int64, error) {
	res, err := r.db.ExecContext(ctx, `INSERT INTO meals
		(date, meal, description, calories, protein_g, fat_g, carbs_g, source)
		VALUES (?,?,?,?,?,?,?,?)
		ON CONFLICT(date, meal, description) DO UPDATE SET
			calories=excluded.calories,
			protein_g=excluded.protein_g,
			fat_g=excluded.fat_g,
			carbs_g=excluded.carbs_g,
			source=excluded.source`,
		m.Date, m.Meal, m.Description, m.Calories, m.ProteinG, m.FatG, m.CarbsG, m.Source)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM meals WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
