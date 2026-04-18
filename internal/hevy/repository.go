package hevy

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

func (r *Repository) List(ctx context.Context, from, to string) ([]Workout, error) {
	query := `SELECT id, date, session_type, session_num, bodyweight_kg, notes, source
		FROM workouts WHERE 1=1`
	args := []any{}

	if from != "" {
		query += " AND date >= ?"
		args = append(args, from)
	}
	if to != "" {
		query += " AND date <= ?"
		args = append(args, to)
	}
	query += " ORDER BY date DESC, session_num"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workouts []Workout
	for rows.Next() {
		var w Workout
		if err := rows.Scan(&w.ID, &w.Date, &w.SessionType, &w.SessionNum, &w.BodyweightKg, &w.Notes, &w.Source); err != nil {
			return nil, err
		}
		workouts = append(workouts, w)
	}
	return workouts, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*Workout, error) {
	var w Workout
	err := r.db.QueryRowContext(ctx, `SELECT id, date, session_type, session_num, bodyweight_kg, notes, source
		FROM workouts WHERE id = ?`, id).Scan(
		&w.ID, &w.Date, &w.SessionType, &w.SessionNum, &w.BodyweightKg, &w.Notes, &w.Source)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Load sets
	sets, err := r.ListSets(ctx, w.ID)
	if err != nil {
		return nil, err
	}
	w.Sets = sets
	return &w, nil
}

func (r *Repository) ListSets(ctx context.Context, workoutID int64) ([]WorkoutSet, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, workout_id, exercise_name, set_num, weight_kg, reps, rpe, notes
		FROM workout_sets WHERE workout_id = ? ORDER BY set_num`, workoutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sets []WorkoutSet
	for rows.Next() {
		var s WorkoutSet
		if err := rows.Scan(&s.ID, &s.WorkoutID, &s.ExerciseName, &s.SetNum, &s.WeightKg, &s.Reps, &s.RPE, &s.Notes); err != nil {
			return nil, err
		}
		sets = append(sets, s)
	}
	return sets, rows.Err()
}

func (r *Repository) UpsertWorkout(ctx context.Context, w *Workout) (int64, error) {
	res, err := r.db.ExecContext(ctx, `INSERT OR REPLACE INTO workouts
		(date, session_type, session_num, bodyweight_kg, notes, source)
		VALUES (?,?,?,?,?,?)`,
		w.Date, w.SessionType, w.SessionNum, w.BodyweightKg, w.Notes, w.Source)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) DeleteSetsByWorkout(ctx context.Context, workoutID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM workout_sets WHERE workout_id = ?", workoutID)
	return err
}

func (r *Repository) InsertSet(ctx context.Context, s *WorkoutSet) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO workout_sets
		(workout_id, exercise_name, set_num, weight_kg, reps, rpe, notes)
		VALUES (?,?,?,?,?,?,?)`,
		s.WorkoutID, s.ExerciseName, s.SetNum, s.WeightKg, s.Reps, s.RPE, s.Notes)
	return err
}
